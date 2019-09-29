package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"../rango"
	// "../rango/auth"
)

func main() {
	rango.DebugOn()
	sev := initServer()
	// Runing Server
	sev.Start(":8080")
}

func initServer() *rango.RangoSev {
	sev := rango.New("demo")

	// rango Func example
	// use custom matcher
	// set before-route.middleware
	sev.Func("/add", func(vars rango.ReqVars) interface{} {
		numa := vars.GetDefault("a", 1)
		numb := vars.GetDefault("b", 1)
		res := toInt(numb) + toInt(numa)
		// return []byte(fmt.Sprintf("<h1>%v+%v=%v</h1>", numa, numb, res))
		return res
	}).AddMatcher(newThrottleMatcher(5000)).Before(func(w http.ResponseWriter, r *http.Request) bool {
		w.Write([]byte("Add tool is closed now."))
		w.WriteHeader(500)
		return false
	})

	// upload handler
	// route, handler := sev.Upload(...)
	_, uploadHandler := sev.Upload("/upload", "./imgs", 10*1024, []string{"image"})
	uploadHandler.Failed(func(code int, err error, msg string, w http.ResponseWriter) {
		rango.DefaultFailed(code, err, msg, w)
		fmt.Printf("[LOG] code, msg = %v, %v\n", code, msg)
	}).After(func(fileBytes []byte, pth string) (error, interface{}) {
		err := rango.SaveFile(fileBytes, pth)
		_, filename := filepath.Split(pth)
		return err, map[string]string{
			"url": "/image/" + filename,
		}
	})

	// Group routing
	apiGroup := sev.Group("/api")
	apiGroup.Func("/user/{id:\\d+}", func(vars rango.ReqVars) interface{} {
		userID := rango.GetConf("userPrefix", "").(string) + "_" + vars.GetDefault("id", "null").(string)
		return userID
	})

	// HATEOAS
	hateoas := rango.NewRHateoas("**HATEOAS**", "In progress")
	hateoas.Add("add tool", "/add?a={a}&b={b}", "add number a and number b", []string{"GET", "POST"})
	// map url [/api/] to hateoas serveHTTP
	apiGroup.Handle("/", hateoas)

	// map [/api] to static file
	apiGroup.File("", "./api_README.md")

	// set Static folder
	sev.Static("/image", "./imgs")
	sev.Static("/", "./www")

	return sev
}

func toInt(i interface{}) int {
	if v, ok := i.(int); ok {
		return v
	}
	if v, ok := i.(string); ok {
		if ret, err := strconv.Atoi(v); err == nil {
			return ret
		}
	}
	if v, ok := i.(float32); ok {
		return int(v)
	}
	return 0
}

type throttleMatcher struct {
	lastCall time.Time
	timeout  float64
}

func newThrottleMatcher(timeout float64) *throttleMatcher {
	return &throttleMatcher{
		// lastCall: time.Now(),
		timeout: timeout,
	}
}

func (t *throttleMatcher) Match(r *http.Request) bool {
	now := time.Now()
	if now.Sub(t.lastCall).Seconds()*1000 > t.timeout {
		t.lastCall = now
		return true
	}
	return false
}
