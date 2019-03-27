package mid

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"

	"../core"
)

// LogRequest print a request status
func LogRequest(w core.ResponseWriteBody, r *http.Request, next func()) {
	t := time.Now()
	next()
	log.Printf("%v %v %v use time %v content-length %v",
		r.Method,
		w.StatusCode(),
		r.URL.String(),
		time.Now().Sub(t).String(),
		w.ContentLength())
}

// ErrCatch catch and recover
func ErrCatch(w core.ResponseWriteBody, r *http.Request, next func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
	}()
	next()
	if w.StatusCode() == 404 {
		w.Write([]byte("404!"))
	}
}

const sessionCookieName = "__sid__"

// Sission if request cookies.length == 0 then add a cookie
func Sission(w core.ResponseWriteBody, r *http.Request, next func()) {

	if _, err := r.Cookie(sessionCookieName); err != nil {
		c := new(http.Cookie)
		c.HttpOnly = true
		c.Expires = time.Now().Add(time.Hour)
		c.Name = sessionCookieName
		c.Value = randStr(40)
		http.SetCookie(w, c)
	}
	next()
}

const strs = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var randsrc = rand.NewSource(time.Now().UnixNano())

// randStr rand string
func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = strs[randsrc.Int63()%int64(len(strs))]
	}
	return string(b)
}
