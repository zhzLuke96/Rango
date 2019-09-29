package rango

import (
	"encoding/json"
	"net/http"
)

func getRespCode(offset, statusCode int, msg string) int {
	code := strOffset(msg, responseIdxMAX)%responseIdxMAX + offset*responseIdxMAX
	return code + statusCode*responseIdxMAX*responseIdxMAX
}

type rResponse struct {
	offsetCode int

	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Token   string `json:"token"`

	Data interface{} `json:"data"` // 数据
	Meta interface{} `json:"meta"` // pagebar navbar
}

func (r *rResponse) Reset() {
	r.Message = "null"
	r.Code = -1
	r.Token = "null"
	r.Status = "null"
	r.Data = nil
	r.Meta = nil
}

func (r *rResponse) Push(w http.ResponseWriter, statusCode int, msg string, data interface{}) {
	r.PushWithMeta(w, statusCode, msg, data, nil)
}

func (r *rResponse) PushWithMeta(w http.ResponseWriter, statusCode int, msg string, data, meta interface{}) {
	r.Reset()

	r.Code = getRespCode(r.offsetCode, statusCode, msg)
	r.Message = msg
	r.Status = httpCodeText(statusCode)
	r.Data = data
	r.Meta = meta

	w.Header().Set("Content-Type", "application/json")
	if statusCode != 200 {
		w.WriteHeader(statusCode)
	}
	w.Write(r.JSON())
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

	Code    int           `json:"error_code"`
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Status  string        `json:"status"`
	Debug   []interface{} `json:"debug"`
}

func (e *errResponse) Reset() {
	e.Code = -1
	e.Error = "null"
	e.Message = "null"
	e.Debug = nil
}

func (e *errResponse) Push(w http.ResponseWriter, statusCode int, msg string, err error) {
	e.Reset()

	e.Code = getRespCode(e.offsetCode, statusCode, msg)
	e.Message = msg
	if isDebugOn() && err != nil {
		e.Error = err.Error()
	}
	e.Status = httpCodeText(statusCode)

	w.Header().Set("Content-Type", "application/json")
	if statusCode != 200 {
		w.WriteHeader(statusCode)
	}
	w.Write(e.JSON())
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
