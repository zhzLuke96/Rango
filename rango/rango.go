package rango

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

// ResponseWriter implement ResponseWriteBody interface
type ResponseWriter struct {
	status int
	length int
	http.ResponseWriter
}

// NewResponseWriter create ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	wr := new(ResponseWriter)
	wr.ResponseWriter = w
	wr.status = 200
	return wr
}

// WriteHeader write status code
func (r *ResponseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.length += n
	return n, err
}

// StatusCode return status code
func (r *ResponseWriter) StatusCode() int {
	return r.status
}

// ContentLength return content length
func (r *ResponseWriter) ContentLength() int {
	return r.length
}

// MiddlewareFunc filter type
type MiddlewareFunc func(ResponseWriteBody, *http.Request, func())

// SevHandler server struct
type SevHandler struct {
	middlewares []MiddlewareFunc
	Handler     http.Handler
}

func NewSevHandler() *SevHandler {
	return &SevHandler{}
}

func (s *SevHandler) HandleFunc(fn func(http.ResponseWriter, *http.Request)) *SevHandler {
	s.Handler = http.HandlerFunc(fn)
	return s
}

// ServeHTTP for http.Handler interface
func (s *SevHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	i := 0
	wr := NewResponseWriter(w)
	var next func()
	next = func() {
		if i < len(s.middlewares) {
			i++
			s.middlewares[i-1](wr, req, next)
		} else if s.Handler != nil {
			s.Handler.ServeHTTP(wr, req)
		}
	}
	next()
}

// Use push MiddlewareFunc
func (r *SevHandler) Use(funcs ...MiddlewareFunc) *SevHandler {
	r.middlewares = append(r.middlewares, funcs...)
	return r
}
func (r *SevHandler) UseBefore(funcs ...MiddlewareFunc) *SevHandler {
	r.middlewares = append(funcs, r.middlewares...)
	return r
}

func (r *SevHandler) Start(port string) error {
	sev := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	return sev.ListenAndServe()
}

func (r *SevHandler) StartServer(sev *http.Server) error {
	sev.Handler = r
	return sev.ListenAndServe()
}

func (r *SevHandler) StartServerTLS(sev *http.Server, certFile, keyFile string) error {
	sev.Handler = r
	return sev.ListenAndServeTLS(certFile, keyFile)
}

type HandlerFunc func(ResponseWriteBody, *http.Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(NewResponseWriter(w), r)
}

type fileServer string

func (f fileServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	data, err := loadFile(string(f))
	if err != nil {
		w.WriteHeader(404)
		return
	}
	ctype := contentType(string(f))
	w.Header().Set("Content-Type", ctype)
	w.Write(data)
	// w.WriteHeader(200)
}
