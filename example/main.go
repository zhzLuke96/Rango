package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"../middleware"
	"../rango"
)

var memCacher = middleware.NewMemCacher(10)

func main() {
	rango.DebugOn()
	rango.ReadConfig()

	sev := initServer()
	// Runing Server
	sev.Start(":8080")
}

func newSev(name string) *rango.Server {
	sev := &rango.Server{
		Name:    name,
		Router:  rango.NewRouter(),
		Handler: rango.NewSevHandler(),
	}
	sev.Handler.Use(rango.LogRequestMid)
	sev.Handler.Use(rango.ErrCatchMid)
	sev.Handler.Use(rango.SignHeaderMid("server", "rango/"+rango.Version))
	sev.Handler.Use(memCacher.Mid)
	sev.Handler.Use(sev.Router.Mid)
	return sev
}

func initServer() *rango.Server {
	sev := newSev("demo")

	// rango Func example
	// use custom matcher
	// set before-route.middleware
	sev.Func("/add", func(vars *rango.ReqVars) interface{} {
		numa := vars.GetDefault("a", 1)
		numb := vars.GetDefault("b", 1)
		res := toInt(numb) + toInt(numa)
		// return []byte(fmt.Sprintf("<h1>%v+%v=%v</h1>", numa, numb, res))
		return res
	}).AddMatcher(newThrottleMatcher(5000))

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
	apiGroup.Func("/user/{id:\\d+}", func(vars *rango.ReqVars) interface{} {
		userID := rango.GetConf("userPrefix", "").(string) + "_" + vars.GetDefault("id", "null").(string)
		return userID
	})

	// map url [/api/] to hateoas serveHTTP
	// apiGroup.Handle("/", hateoas)

	// HTML /home form inde.html index.js style.css
	apiGroup.HTML("/", "./www/index.html", "./www/index.js", "./www/style.css")

	// map [/api] to static file
	apiGroup.File("", "./api_README.md")

	// set Static folder
	sev.Static("/", "./www")
	sev.Static("/image", "./imgs")

	// WebAssembly
	_, wasmHTML := sev.HTML("/wasm", "./www/index.wasm", "./www/rainbow.wasm")
	wasmHTML(func(h *rango.Rhtml) {
		h.Title([]byte("WebAssembly"))
		h.AppendStyle([]byte(`body,html { margin:0;padding:0;width:100%;height:100%;background:#282c34;color:honeydew; }
#app{text-align: center;}
#bgcanvas { position:fixed;opacity:0.5;top:0;right:0;bottom:0;left:0;}`))
		h.AppendBody([]byte("<div id='app'></div>"))
		h.AppendBody([]byte("<canvas id='bgcanvas'></canvas>"))
	})
	sev.File("/wasm_exec.js", "./wasm_exec.js")

	// sort router table
	sev.Sort()

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
