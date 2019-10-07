package rango

import (
	"encoding/json"
	"net/http"
)

type responseify interface {
	Push(http.ResponseWriter)
}

func getRespCode(offset, statusCode int, msg string) int {
	code := strOffset(msg, responseIdxMAX)%responseIdxMAX + offset*responseIdxMAX
	return code + statusCode*responseIdxMAX*responseIdxMAX
}

type rResponse struct {
	offsetCode int

	statusCode int

	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Token   string `json:"token"`

	Data interface{} `json:"data"` // 数据
	Meta interface{} `json:"meta"` // pagebar navbar
}

func (r *rResponse) Reset() *rResponse {
	r.statusCode = 200
	r.Message = "null"
	r.Code = -1
	r.Token = "null"
	r.Status = "null"
	r.Data = nil
	r.Meta = nil
	return r
}

func (r *rResponse) Set(statusCode int, msg string, data, meta interface{}) *rResponse {
	r.statusCode = statusCode
	r.Code = getRespCode(r.offsetCode, statusCode, msg)
	r.Message = msg
	r.Status = httpCodeText(statusCode)
	r.Data = data
	r.Meta = meta
	return r
}

func (r rResponse) Push(w http.ResponseWriter) {
	if r.statusCode < 100 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if r.statusCode != 200 {
		w.WriteHeader(r.statusCode)
	}
	w.Write(r.JSON())
}

func (r *rResponse) PushReset(w http.ResponseWriter, statusCode int, msg string, data, meta interface{}) {
	r.Reset()
	r.Set(statusCode, msg, data, meta)
	r.Push(w)
}

func (r *rResponse) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

type errResponse struct {
	offsetCode int

	statusCode int

	Code    int           `json:"error_code"`
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Status  string        `json:"status"`
	Debug   []interface{} `json:"debug"`
}

func (e *errResponse) Reset() *errResponse {
	e.Code = -1
	e.Error = "null"
	e.Message = "null"
	e.Debug = nil
	e.statusCode = 500
	return e
}

func (e *errResponse) Set(statusCode int, msg string, err error) *errResponse {
	e.statusCode = statusCode
	e.Code = getRespCode(e.offsetCode, statusCode, msg)
	e.Message = msg
	if isDebugOn() && err != nil {
		e.Error = err.Error()
	}
	e.Status = httpCodeText(statusCode)
	return e
}

func (e errResponse) Push(w http.ResponseWriter) {
	if e.statusCode < 100 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if e.statusCode != 200 {
		w.WriteHeader(e.statusCode)
	}
	w.Write(e.JSON())
}

func (e *errResponse) PushReset(w http.ResponseWriter, statusCode int, msg string, err error) {
	e.Reset()
	e.Set(statusCode, msg, err)
	e.Push(w)
}

func (e *errResponse) JSON() []byte {
	ret, err := json.MarshalIndent(e, "", "  ")
	if err == nil {
		return ret
	}
	return nil
}

type rResponser struct {
	Name string
	Code int
}

func NewResponser(name string) *rResponser {
	return &rResponser{
		Name: name,
		Code: strOffset(name, responseIdxMAX),
	}
}

func (r *rResponser) NewResponse() *rResponse {
	return &rResponse{
		offsetCode: r.Code,
	}
}

func (r *rResponser) NewErrResponse() *errResponse {
	return &errResponse{
		offsetCode: r.Code,
	}
}
