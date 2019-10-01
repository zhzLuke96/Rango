package rango

import (
	"net/http"
	"strings"
)

func SimpleGo(port string) error {
	sev := New(simpleServerName)
	sev.Static("/", "./www")
	return sev.Start(port)
}

type RangoSev struct {
	Name    string
	Router  *Router
	Handler *SevHandler

	groups []*RangoSev
}

func New(name string) *RangoSev {
	if name == "" {
		name = randStr(10)
	}
	sev := &RangoSev{
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

func NewSev(name string) *RangoSev {
	if name == "" {
		name = randStr(10)
	}
	return &RangoSev{
		Name:    name,
		Router:  NewRouter(),
		Handler: NewSevHandler(),
	}
}

func (r *RangoSev) Func(routerPthTpl string, fn rHFunc) *Route {
	return r.Router.Func(fn).Path(routerPthTpl)
}

func (r *RangoSev) Handle(routerPthTpl string, handler http.Handler) *Route {
	return r.Router.Handle(handler).Path(routerPthTpl)
}

func (r *RangoSev) Static(routerPath, dirPth string) *Route {
	return r.StaticDir(routerPath, dirPth, true)
}

func (r *RangoSev) StaticDir(routerPath, dirPth string, justFiles bool) *Route {
	var fs http.Handler
	if justFiles {
		fs = newJustFilesFS(dirPth, 2)
	} else {
		dir := http.Dir(dirPth)
		fs = http.FileServer(dir)
	}
	fs = http.StripPrefix(routerPath, fs)
	return r.Router.Handle(fs).PathMatch(routerPath, true)
}

func (r *RangoSev) File(routerPath, filePth string) *Route {
	return r.Router.Handle(fileServer(filePth)).Path(routerPath)
}

func (r *RangoSev) Bytes(routerPath string, genFn func() []byte) *Route {
	return r.Router.Handle(bytesServer(genFn)).Path(routerPath)
}

func (r *RangoSev) String(routerPath string, genFn func() string) *Route {
	return r.Bytes(routerPath, func() []byte {
		return []byte(genFn())
	})
}

func (r *RangoSev) HTML(routerPath string, filenames ...string) *Route {
	nonHTMLfile := true
	htmlFilename := ""
	for _, v := range filenames {
		if strings.HasSuffix(v, ".html") {
			nonHTMLfile = false
			htmlFilename = v
		}
	}
	return r.String(routerPath, func() string {
		if nonHTMLfile {
			return ""
		}
		htmlContent, err := loadFile(htmlFilename)
		if err != nil {
			return ""
		}
		HTML := html(htmlContent)
		for _, v := range filenames {
			if v == htmlFilename {
				continue
			}
			if strings.HasSuffix(v, ".css") {
				cssFile, err := loadFile(htmlFilename)
				if err != nil {
					continue
				}
				HTML.AppendStyle(string(cssFile))
			}
			if strings.HasSuffix(v, ".js") {
				cssFile, err := loadFile(htmlFilename)
				if err != nil {
					continue
				}
				HTML.AppendScript(string(cssFile))
			}
			if strings.HasSuffix(v, ".html") {
				cssFile, err := loadFile(htmlFilename)
				if err != nil {
					continue
				}
				HTML.AppendBody(string(cssFile))
			}
		}
		return string(HTML)
	})
}

func (r *RangoSev) Upload(routerPath, dir string, maxsize int64, accept []string) (*Route, *uploadServer) {
	sev := newUploadServer(dir, maxsize, accept)
	return r.Router.Handle(sev).Path(routerPath), sev
}

func (r *RangoSev) GET(routerPthTpl string, fn rHFunc) *Route {
	return r.Func(routerPthTpl, fn).Methods("GET")
}

func (r *RangoSev) POST(routerPthTpl string, fn rHFunc) *Route {
	return r.Func(routerPthTpl, fn).Methods("POST")
}

func (r *RangoSev) Group(routerPthTpl string) *RangoSev {
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

	r.groups = append(r.groups, subSev)
	return subSev
}

func (r *RangoSev) StartServer(sev *http.Server) error {
	return r.Handler.StartServer(sev)
}

func (r *RangoSev) StartServerTLS(sev *http.Server, certFile, keyFile string) error {
	return r.Handler.StartServerTLS(sev, certFile, keyFile)
}

func (r *RangoSev) Start(port string) error {
	return r.Handler.Start(port)
}

func (r *RangoSev) Use(funcs ...MiddlewareFunc) *RangoSev {
	r.Handler.Use(funcs...)
	return r
}

func (r *RangoSev) UseBefore(funcs ...MiddlewareFunc) *RangoSev {
	r.Handler.UseBefore(funcs...)
	return r
}

func (r *RangoSev) Sort() {
	r.Router.Sort()
	for _, g := range r.groups {
		g.Router.Sort()
	}
}

func (r *RangoSev) IsSorted() bool {
	if !r.Router.IsSorted() {
		return false
	}
	for _, g := range r.groups {
		if !g.Router.IsSorted() {
			return false
		}
	}
	return true
}
