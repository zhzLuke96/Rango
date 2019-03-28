package mid

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	core "../Core"
	"../utils"
)

// LogRequest print a request status
func LogRequest(w core.ResponseWriteBody, r *http.Request, next func()) {
	t := time.Now()
	next()
	log.Printf("%v %v %.1fms\t%v byte\t%v",
		r.Method,
		w.StatusCode(),
		time.Now().Sub(t).Seconds()*1000,
		w.ContentLength(),
		r.URL.String(),
	)
}

// ErrCatch catch and recover
func ErrCatch(w core.ResponseWriteBody, r *http.Request, next func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
	}()
	next()
	if w.StatusCode() == 404 {
		if w.ContentLength() == 0 {
			w.Write([]byte("404!"))
		}
	}
}

const sessionCookieName = "__sid__"

// Sission if request cookies.length == 0 then add a cookie
func Sission(w core.ResponseWriteBody, r *http.Request, next func()) {

	if _, err := r.Cookie(sessionCookieName); err != nil {
		c := new(http.Cookie)
		c.HttpOnly = true
		c.Expires = time.Now().Add(time.Hour)
		c.Name = sessionCookieName
		c.Value = utils.RandStr(40)
		http.SetCookie(w, c)
	}
	next()
}
