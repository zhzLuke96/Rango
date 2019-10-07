package rango

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// LogRequest print a request status
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

// ErrCatch catch and recover
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

// Sission if request cookies.length == 0 then add a cookie
func SissionMid(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
	if _, err := r.Cookie(sessionCookieName); err != nil {
		c := new(http.Cookie)
		c.HttpOnly = true
		c.Expires = time.Now().Add(time.Hour)
		c.Name = sessionCookieName
		c.Value = randStr(40)
		c.Path = "/"
		http.SetCookie(w, c)
	}
	next(w, r)
}

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

func SignHeader(key, value string) MiddlewareFunc {
	return func(w ResponseWriteBody, r *http.Request, next MiddleNextFunc) {
		w.Header().Set(key, value)
		next(w, r)
	}
}
