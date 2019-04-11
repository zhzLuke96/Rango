package Rango

import (
	"net/http"
	"time"

	"./core"
	"./mid"
	"./router"
)

func newDefaultSev(port string) (*http.Server, *core.RangoSevHandler) {
	handler := &core.RangoSevHandler{}
	return &http.Server{
		Addr:        ":" + port,
		Handler:     handler,
		ReadTimeout: 5 * time.Second,
	}, handler
}

func SimpleGo(port string) error {
	sev, handler := newDefaultSev(port)
	handler.Handler = http.FileServer(http.Dir("./www"))
	handler.Use(mid.LogRequest, mid.ErrCatch, mid.Sission)
	return sev.ListenAndServe()
}

func NewRouter() *router.Router {
	return &router.Router{}
}

func NewSev() *core.RangoSevHandler {
	return &core.RangoSevHandler{}
}
