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

// RangoSev server struct
type RangoSev struct {
	middlewares []MiddlewareFunc
	Handler     http.Handler
}

// ServeHTTP for http.Handler interface
func (p *RangoSev) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i := 0
	wr := NewWrapResponseWriter(w)
	var next func()
	next = func() {
		if i < len(p.middlewares) {
			i++
			p.middlewares[i-1](wr, r, next)
		} else if p.Handler != nil {
			p.Handler.ServeHTTP(wr, r)
		}
	}
	next()
}

// Use push MiddlewareFunc
func (p *RangoSev) Use(funcs ...MiddlewareFunc) {
	for _, f := range funcs {
		p.middlewares = append(p.middlewares, f)
	}
}

func (p *RangoSev) Go(port string) {
	sev := &http.Server{
		Addr:        ":" + port,
		Handler:     p,
		ReadTimeout: 5 * time.Second,
	}
	sev.ListenAndServe()
}
