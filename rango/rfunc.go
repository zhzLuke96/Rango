package rango

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ReqVars struct {
	path  map[string]interface{}
	query map[string]string
	json  map[string]interface{}

	req *http.Request
}

func newReqVarsBase() *ReqVars {
	return &ReqVars{
		path:  make(map[string]interface{}),
		query: make(map[string]string),
		json:  make(map[string]interface{}),
	}
}

func newReqVars(req *http.Request) (*ReqVars, error) {
	// 优先级
	// path < query < body
	vars := newReqVarsBase()
	vars.req = req

	// load path vars
	path := Vars(req)
	for k, v := range path {
		vars.path[k] = v
	}
	// load query string
	values := req.URL.Query()
	for k := range values {
		vars.query[k] = values.Get(k)
	}
	// load body json
	body, _ := ioutil.ReadAll(req.Body)
	if len(body) != 0 {
		if body != nil {
			if err := json.Unmarshal(body, &vars.json); err != nil {
				return nil, err
			}
		}
	}
	// success parse
	return vars, nil
}

func (r ReqVars) Request() *http.Request {
	return r.req
}

func (r ReqVars) Query() map[string]string {
	return r.query
}

func (r ReqVars) Path() map[string]interface{} {
	return r.path
}

func (r ReqVars) JSON() map[string]interface{} {
	return r.json
}

func (r ReqVars) Has(key string) bool {
	if _, ok := r.json[key]; ok {
		return true
	}
	if _, ok := r.path[key]; ok {
		return true
	}
	if _, ok := r.query[key]; ok {
		return true
	}
	return false
}

func (r ReqVars) HasAll(keys []string) bool {
	for _, k := range keys {
		if !r.Has(k) {
			return false
		}
	}
	return true
}

func (r ReqVars) Get(key string) (interface{}, error) {
	if v, ok := r.json[key]; ok {
		return v, nil
	}
	if v, ok := r.path[key]; ok {
		return v, nil
	}
	if v, ok := r.query[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("cannot get value of key: %v", key)
}

func (r ReqVars) GetDefault(key string, defaultValue interface{}) interface{} {
	if v, ok := r.json[key]; ok {
		return v
	}
	if v, ok := r.path[key]; ok {
		return v
	}
	if v, ok := r.query[key]; ok {
		return v
	}
	return defaultValue
}

type rHFunc func(*ReqVars) interface{}

func (r rHFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars, err := newReqVars(req)
	if err != nil {
		errResp := rHFuncResponser.NewErrResponse()
		errResp.PushReset(w, 400, "Error parsing request body", err)
		return
	}
	respValue := r(vars)
	switch t := respValue.(type) {
	case responseify:
		t.Push(w)
		return
	case ErrResponse, Response:
		respValue.(responseify).Push(w)
		return
	case *ErrResponse:
		t.Push(w)
		return
	case *Response:
		t.Push(w)
		return
	case []byte:
		w.Write(t)
		return
	case *[]byte:
		w.Write(*t)
		return
	default:
		resp := rHFuncResponser.NewResponse()
		resp.PushReset(w, 200, "success", respValue, nil)
	}
}
