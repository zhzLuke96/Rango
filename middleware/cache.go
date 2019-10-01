package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"time"

	"../rango"
)

func bsMD5(bs []byte) string {
	md5 := md5.New()
	md5.Write(bs)
	return hex.EncodeToString(md5.Sum(nil))
}

func reqKey(r *http.Request) string {
	method := []byte(r.Method)
	url := []byte(r.URL.String())
	body, _ := ioutil.ReadAll(r.Body)
	return bsMD5(append(append(method, url...), body...))
}

type cacherCallback func() (interface{}, int)
type cacher interface {
	Cached(string) bool
	Cache(string, interface{})
	Get(string, cacherCallback) interface{}
	Mid(rango.ResponseWriteBody, *http.Request, func())
}

type memCacherItem struct {
	timeStamp time.Time
	value     interface{}
}

type memCacher struct {
	mems    map[string]memCacherItem
	timeout float64
}

func NewMemCacher(timeout float64) *memCacher {
	return &memCacher{
		mems:    make(map[string]memCacherItem),
		timeout: timeout,
	}
}

func (m *memCacher) Mid(w rango.ResponseWriteBody, r *http.Request, next rango.MiddleNextFunc) {
	reqK := reqKey(r)

	resp := m.Get(reqK, func() (interface{}, int) {
		// proxy
		wr := newRespProxyWriter(w.Writer())
		w.Target(wr)
		next(w, r)
		return wr.Body, w.StatusCode()
	})
	w.Write(resp.([]byte))
}

func (m *memCacher) Cached(key string) bool {
	if v, ok := m.mems[key]; ok {
		cached := time.Now().Sub(v.timeStamp).Seconds() < m.timeout
		if !cached {
			delete(m.mems, key)
		}
		return cached
	}
	return false
}
func (m *memCacher) Cache(key string, value interface{}) {
	m.mems[key] = memCacherItem{
		timeStamp: time.Now(),
		value:     value,
	}
}
func (m *memCacher) Get(key string, fn cacherCallback) interface{} {
	if m.Cached(key) {
		return m.mems[key].value
	}
	newValue, code := fn()
	if code < 400 {
		m.Cache(key, newValue)
	}
	return []byte("")
}
