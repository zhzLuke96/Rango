package rango

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ReqVars map[string]interface{}

func newReqVarsBase() *ReqVars {
	inner := make(ReqVars)
	return &inner
}

// func newReqVarsFromJSONStr(jsonText string) (*ReqVars, error) {
// 	return newReqVarsFromJSON([]byte(jsonText))
// }

// func newReqVarsFromJSON(jsonText []byte) (*ReqVars, error) {
// 	vars := newReqVarsBase()

// 	if err := json.Unmarshal(jsonText, vars); err != nil {
// 		return nil, err
// 	}
// 	return vars, nil
// }

func newReqVars(req *http.Request) (*ReqVars, error) {
	// 优先级
	// path < query < body
	vars := newReqVarsBase()
	// load path vars
	pathVars := Vars(req)
	for k, v := range pathVars {
		(*vars)[k] = v
	}
	// load query string
	values := req.URL.Query()
	for k := range values {
		(*vars)[k] = values.Get(k)
	}
	// load body json
	body, _ := ioutil.ReadAll(req.Body)
	if len(body) != 0 {
		if body != nil {
			if err := json.Unmarshal(body, vars); err != nil {
				return nil, err
			}
		}
	}
	// success parse
	return vars, nil
}

func (r ReqVars) Has(key string) bool {
	_, ok := r[key]
	return ok
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
	if v, ok := r[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("cannot get value of key: %v", key)
}

func (r ReqVars) GetDefault(key string, defaultValue interface{}) interface{} {
	if v, ok := r[key]; ok {
		return v
	}
	return defaultValue
}

type rHFunc func(ReqVars) interface{}

func (r rHFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars, err := newReqVars(req)
	if err != nil {
		errResp := rHFuncResponser.NewErrResponse()
		errResp.Push(w, 400, "Error parsing request body", err)
		return
	}
	resp := rHFuncResponser.NewResponse()
	resp.Push(w, 200, "success", r(*vars))
}
