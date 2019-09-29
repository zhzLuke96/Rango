package rango

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

// LogRequest print a request status
func LogRequestMid(w ResponseWriteBody, r *http.Request, next func()) {
	t := time.Now()
	method := r.Method
	url := r.URL.String()
	next()
	log.Printf("%v %v %.1fms\t%v byte\t%v",
		w.StatusCode(),
		method,
		time.Now().Sub(t).Seconds()*1000,
		w.ContentLength(),
		url,
	)
}

// ErrCatch catch and recover
func ErrCatchMid(w ResponseWriteBody, r *http.Request, next func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
	}()
	next()
	if w.StatusCode() >= 400 {
		if w.ContentLength() == 0 {
			errCatchResponser.NewErrResponse().Push(w, w.StatusCode(), "Unknown Error Catched", nil)
		}
	}
}

// Sission if request cookies.length == 0 then add a cookie
func SissionMid(w ResponseWriteBody, r *http.Request, next func()) {
	if _, err := r.Cookie(sessionCookieName); err != nil {
		c := new(http.Cookie)
		c.HttpOnly = true
		c.Expires = time.Now().Add(time.Hour)
		c.Name = sessionCookieName
		c.Value = randStr(40)
		c.Path = "/"
		http.SetCookie(w, c)
	}
	next()
}

func StripPrefixMid(prefix string) MiddlewareFunc {
	return func(w ResponseWriteBody, r *http.Request, next func()) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			newURL := new(url.URL)
			newURL.Path = p
			*r.URL = *newURL
		}
		next()
	}
}

func SignHeader(key, value string) MiddlewareFunc {
	return func(w ResponseWriteBody, r *http.Request, next func()) {
		w.Header().Set(key, value)
		next()
	}
}
