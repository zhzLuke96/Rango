package rango

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// LogRequestMid print a request status
func LogRequestMid(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
	t := time.Now()
	method := r.Method
	url := r.URL.String()
	next(w, r)
	log.Printf("%v %v %.1fms\t%v byte\t%v",
		w.StatusCode(),
		method,
		time.Now().Sub(t).Seconds()*1000,
		w.ContentLength(),
		url,
	)
}

// ErrCatchMid catch and recover
func ErrCatchMid(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
	}()
	next(w, r)
	if w.StatusCode() >= 400 {
		if w.ContentLength() == 0 {
			errCatchResponser.NewErrResponse().PushReset(w, w.StatusCode(), "Unknown Error Catched", nil)
		}
	}
}

// StripPrefixMid 行为和http.StripPrefix一样
func StripPrefixMid(prefix string) MiddlewareFunc {
	return func(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			newURL := cloneURL(r.URL)
			newURL.Path = p
			*r.URL = *newURL
		}
		next(w, r)
	}
}

// SignHeaderMid 修改http返回头中的一个值
func SignHeaderMid(key, value string) MiddlewareFunc {
	return func(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
		w.Header().Set(key, value)
		next(w, r)
	}
}
