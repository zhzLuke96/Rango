package Rango

import (
	"net/http"
	"time"

	mid "./Mid"
	"./Router"
	"./core"
)

func newDefaultSev(port string) (*http.Server, *core.MidHandler) {
	handler := &core.RangoSev{}
	return &http.Server{
		Addr:        ":" + port,
		Handler:     handler,
		ReadTimeout: 5 * time.Second,
	}, handler
}

func SimpleGo(port string) error {
	sev, handler := newDefaultSev(port)
	handler.Use(mid.LogRequest, mid.ErrCatch, mid.Sission)
	return sev.ListenAndServe()
}

func NewRouter() *Router.Router {
	return &Router.Router{}
}

func NewSev() *core.RangoSev {
	return &core.RangoSev{}
}
