package middleware

import (
	"net/http"
)

// 这里放置和编写各种中间件

type ResponseProxyWriter struct {
	writer http.ResponseWriter
	Body   []byte
}

func (r *ResponseProxyWriter) Header() http.Header {
	return r.writer.Header()
}
func (r *ResponseProxyWriter) Write(bytes []byte) (int, error) {
	r.Body = append(r.Body, bytes[0:len(bytes)]...)
	return r.writer.Write(bytes)
}
func (r *ResponseProxyWriter) WriteHeader(i int) {
	r.writer.WriteHeader(i)
}

func newRespProxyWriter(w http.ResponseWriter) *ResponseProxyWriter {
	return &ResponseProxyWriter{
		writer: w,
		Body:   []byte{},
	}
}
