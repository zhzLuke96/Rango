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

type rResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Token   string `json:"token"`

	Data interface{} `json:"data"` // 数据
	Meta interface{} `json:"meta"` // pagebar navbar
}

func NewEmptyResponse() *rResponse {
	return &rResponse{
		Code:    1,
		Status:  httpCodeText(200),
		Message: "success",
		Token:   "null",
	}
}

func NewResponse(data interface{}) *rResponse {
	resp := NewEmptyResponse()
	resp.Data = data
	return resp
}

func (r *rResponse) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

type errResponse struct {
	httpStatusCode int

	Code    int           `json:"error_code"`
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Debug   []interface{} `json:"debug"`
}

func NewErrResp(code int, err, msg string) *errResponse {
	return &errResponse{
		Code:    code,
		Error:   err,
		Message: msg,
		Debug:   getDebugStackArr(),
	}
}

func (e *errResponse) JSON() []byte {
	ret, err := json.MarshalIndent(e, "", "  ")
	if err == nil {
		return ret
	}
	return nil
}

type rHFunc func(ReqVars) []byte

func (r rHFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var respBytes []byte
	stateCode := 200
	vars, err := newReqVars(req)
	if err != nil {
		stateCode = 400
		respBytes = NewErrResp(1000, err.Error(), systemError).JSON()
	} else {
		respBytes = r(*vars)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
	if stateCode != 200 {
		w.WriteHeader(stateCode)
	}
}
