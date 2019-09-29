package rango

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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

func DefaultFailed(statusCode int, err error, msg string, w http.ResponseWriter) {
	resp := uploadResponser.NewErrResponse()
	resp.Push(w, statusCode, msg, err)
}

type afterFunc func([]byte, string) (error, interface{})
type faildFunc func(int, error, string, http.ResponseWriter)

type uploadServer struct {
	storageDir    string
	maxUploadSize int64
	acceptType    []string
	allowAll      bool

	after  afterFunc
	failed faildFunc
}

func newUploadServer(dir string, maxsize int64, accept []string) *uploadServer {
	if dirExist, _ := pathExists(dir); !dirExist {
		os.Mkdir(dir, os.ModePerm)
	}
	return &uploadServer{
		storageDir:    dir,
		maxUploadSize: maxsize * 1024,
		acceptType:    accept,
		allowAll:      sliceIndexPrefix(accept, "*") != -1,
	}
}

func (u *uploadServer) After(fn afterFunc) *uploadServer {
	u.after = fn
	return u
}

func (u *uploadServer) Failed(fn faildFunc) *uploadServer {
	u.failed = fn
	return u
}

func (u *uploadServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var failedCallback faildFunc
	if u.failed != nil {
		failedCallback = u.failed
	} else {
		failedCallback = DefaultFailed
	}
	req.Body = http.MaxBytesReader(w, req.Body, u.maxUploadSize)
	if err := req.ParseMultipartForm(u.maxUploadSize); err != nil {
		failedCallback(413, err, "Request Entity Too Large", w)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		failedCallback(400, err, "Cant Load [file] Filed In Body.", w)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		failedCallback(400, err, "ReadAll Error", w)
		return
	}
	filetype := http.DetectContentType(fileBytes)
	if !u.allowAll && sliceIndexPrefix(u.acceptType, filetype) == -1 {
		failedCallback(405, nil, "File Type Not Allowed", w)
		return
	}

	fileName := fileMD5(fileBytes)
	fileEndings, err := mime.ExtensionsByType(filetype)
	if err != nil || len(fileEndings) == 0 {
		failedCallback(400, err, "Extensions Error", w)
		return
	}
	newPath := filepath.Join(u.storageDir, fileName+fileEndings[0])
	// fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

	var afterErr error
	var data interface{}
	if u.after != nil {
		afterErr, data = u.after(fileBytes, newPath)
	} else {
		afterErr = SaveFile(fileBytes, newPath)
	}
	if afterErr != nil {
		failedCallback(400, afterErr, "Save File Error", w)
		return
	}
	resp := uploadResponser.NewResponse()
	resp.Push(w, 200, "UPLOAD SUCCESS", data)
}
