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

type Response struct {
	offsetCode int

	statusCode int

	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Token   string `json:"token"`

	Data interface{} `json:"data"` // 数据
	Meta interface{} `json:"meta"` // pagebar navbar
}

func (r *Response) Reset() *Response {
	r.statusCode = 200
	r.Message = "null"
	r.Code = -1
	r.Token = "null"
	r.Status = "null"
	r.Data = nil
	r.Meta = nil
	return r
}

func (r *Response) Set(statusCode int, msg string, data, meta interface{}) *Response {
	r.statusCode = statusCode
	r.Code = getRespCode(r.offsetCode, statusCode, msg)
	r.Message = msg
	r.Status = httpCodeText(statusCode)
	r.Data = data
	r.Meta = meta
	return r
}

func (r Response) Push(w http.ResponseWriter) {
	if r.statusCode < 100 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if r.statusCode != 200 {
		w.WriteHeader(r.statusCode)
	}
	w.Write(r.JSON())
}

func (r *Response) PushReset(w http.ResponseWriter, statusCode int, msg string, data, meta interface{}) {
	r.Reset()
	r.Set(statusCode, msg, data, meta)
	r.Push(w)
}

func (r *Response) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

type ErrResponse struct {
	offsetCode int

	statusCode int

	Code    int           `json:"error_code"`
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Status  string        `json:"status"`
	Debug   []interface{} `json:"debug"`
}

func (e *ErrResponse) Reset() *ErrResponse {
	e.Code = -1
	e.Error = "null"
	e.Message = "null"
	e.Debug = nil
	e.statusCode = 500
	return e
}

func (e *ErrResponse) Set(statusCode int, msg string, err error) *ErrResponse {
	e.statusCode = statusCode
	e.Code = getRespCode(e.offsetCode, statusCode, msg)
	e.Message = msg
	if isDebugOn() && err != nil {
		e.Error = err.Error()
	}
	e.Status = httpCodeText(statusCode)
	return e
}

func (e ErrResponse) Push(w http.ResponseWriter) {
	if e.statusCode < 100 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if e.statusCode != 200 {
		w.WriteHeader(e.statusCode)
	}
	w.Write(e.JSON())
}

func (e *ErrResponse) PushReset(w http.ResponseWriter, statusCode int, msg string, err error) {
	e.Reset()
	e.Set(statusCode, msg, err)
	e.Push(w)
}

func (e *ErrResponse) JSON() []byte {
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

func (r *rResponser) NewResponse() *Response {
	return &Response{
		offsetCode: r.Code,
	}
}

func (r *rResponser) NewErrResponse() *ErrResponse {
	return &ErrResponse{
		offsetCode: r.Code,
	}
}
