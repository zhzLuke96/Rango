package core

import (
	"net/http"
	"time"
)

// ResponseWriteBody for middleware
type ResponseWriteBody interface {
	StatusCode() int
	ContentLength() int
	http.ResponseWriter
}

// WrapResponseWriter implement ResponseWriteBody interface
type WrapResponseWriter struct {
	status int
	length int
	http.ResponseWriter
}

// NewWrapResponseWriter create wrapResponseWriter
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	wr := new(WrapResponseWriter)
	wr.ResponseWriter = w
	wr.status = 200
	return wr
}

// WriteHeader write status code
func (p *WrapResponseWriter) WriteHeader(status int) {
	p.status = status
	p.ResponseWriter.WriteHeader(status)
}

func (p *WrapResponseWriter) Write(b []byte) (int, error) {
	n, err := p.ResponseWriter.Write(b)
	p.length += n
	return n, err
}

// StatusCode return status code
func (p *WrapResponseWriter) StatusCode() int {
	return p.status
}

// ContentLength return content length
func (p *WrapResponseWriter) ContentLength() int {
	return p.length
}

// MiddlewareFunc filter type
type MiddlewareFunc func(ResponseWriteBody, *http.Request, func())

// RangoSevHandler server struct
type RangoSevHandler struct {
	middlewares []MiddlewareFunc
	Handler     http.Handler
}

func (r *RangoSevHandler) HandleFunc(fn func(http.ResponseWriter, *http.Request)) *RangoSevHandler {
	r.Handler = http.HandlerFunc(fn)
	return r
}

// ServeHTTP for http.Handler interface
func (r *RangoSevHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	i := 0
	wr := NewWrapResponseWriter(w)
	var next func()
	next = func() {
		if i < len(r.middlewares) {
			i++
			r.middlewares[i-1](wr, req, next)
		} else if r.Handler != nil {
			r.Handler.ServeHTTP(wr, req)
		}
	}
	next()
}

// Use push MiddlewareFunc
func (r *RangoSevHandler) Use(funcs ...MiddlewareFunc) *RangoSevHandler {
	for _, f := range funcs {
		r.middlewares = append(r.middlewares, f)
	}
	return r
}

func (r *RangoSevHandler) Go(port string) {
	sev := &http.Server{
		Addr:        ":" + port,
		Handler:     r,
		ReadTimeout: 5 * time.Second,
	}
	sev.ListenAndServe()
}

type HandlerFunc func(ResponseWriteBody, *http.Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(NewWrapResponseWriter(w), r)
}
