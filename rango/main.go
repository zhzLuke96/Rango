package rango

import (
	"net/http"
)

func init() {
	// load config.json file
	readConfigFile()
}

func SimpleGo(port string) error {
	sev := New(simpleServerName)
	sev.Static("/", "./www")
	return sev.Start(port)
}

type rangoSev struct {
	Name    string
	Router  *Router
	Handler *SevHandler
}

func New(name string) *rangoSev {
	if name == "" {
		name = randStr(10)
	}
	sev := &rangoSev{
		Name:    name,
		Router:  NewRouter(),
		Handler: NewSevHandler(),
	}
	sev.Handler.Use(LogRequestMid)
	sev.Handler.Use(ErrCatchMid)
	sev.Handler.Use(SignHeader("server", "rango/"+Version))
	sev.Handler.Use(sev.Router.Mid)
	return sev
}

func NewSev(name string) *rangoSev {
	if name == "" {
		name = randStr(10)
	}
	return &rangoSev{
		Name:    name,
		Router:  NewRouter(),
		Handler: NewSevHandler(),
	}
}

func (r *rangoSev) Func(routerPthTpl string, fn rHFunc) *Route {
	return r.Router.Func(fn).Path(routerPthTpl)
}

func (r *rangoSev) Handle(routerPthTpl string, handler http.Handler) *Route {
	return r.Router.Handle(handler).Path(routerPthTpl)
}

func (r *rangoSev) Static(routerPath, dirPth string) *Route {
	dir := http.Dir(dirPth)
	fs := http.FileServer(dir)
	return r.Router.Handle(fs).PathMatch(routerPath, true)
}

func (r *rangoSev) File(routerPath, filePth string) *Route {
	return r.Router.Handle(fileServer(filePth)).Path(routerPath)
}

func (r *rangoSev) GET(routerPthTpl string, fn rHFunc) *Route {
	return r.Func(routerPthTpl, fn).Methods("GET")
}

func (r *rangoSev) POST(routerPthTpl string, fn rHFunc) *Route {
	return r.Func(routerPthTpl, fn).Methods("POST")
}

func (r *rangoSev) Group(routerPthTpl string) *rangoSev {
	// [TODO] 应该支持带有参数的分组
	// 对于匹配到的参数，将作为vars设置到request之上
	// eg.
	// /user/{id:d+}

	// 支持对router和handler分组
	// 也表示可以向下层插入中间件以及matcher
	subSev := NewSev(r.Name + "_" + routerPthTpl)

	subSev.Use(StripPrefixMid(routerPthTpl))
	subSev.Use(subSev.Router.Mid)

	r.Router.Handle(subSev.Handler).PathMatch(routerPthTpl, false)
	return subSev
}

func (r *rangoSev) StartServer(sev *http.Server) error {
	return r.Handler.StartServer(sev)
}

func (r *rangoSev) StartServerTLS(sev *http.Server, certFile, keyFile string) error {
	return r.Handler.StartServerTLS(sev, certFile, keyFile)
}

func (r *rangoSev) Start(port string) error {
	return r.Handler.Start(port)
}

func (r *rangoSev) Use(funcs ...MiddlewareFunc) *rangoSev {
	r.Handler.Use(funcs...)
	return r
}

func (r *rangoSev) UseBefore(funcs ...MiddlewareFunc) *rangoSev {
	r.Handler.UseBefore(funcs...)
	return r
}
