package rango

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

// Server 提供简单的操作
// 需要给server命名，可以方便调试
// 包含了router handler 以及 groups
type Server struct {
	Name    string
	Router  *Router
	Handler *SevHandler

	groups []*Server
}

// New 创建一个带有默认中间件的 Server 并命名
// 如果name为空，则会命名为长度为10的随机字符串
func New(name string) *Server {
	if name == "" {
		name = randStr(10)
	}
	sev := &Server{
		Name:    name,
		Router:  NewRouter(),
		Handler: NewSevHandler(),
	}
	sev.Handler.Use(LogRequestMid)
	sev.Handler.Use(ErrCatchMid)
	sev.Handler.Use(SignHeaderMid("server", "rango/"+Version))
	sev.Handler.Use(sev.Router.Mid)
	return sev
}

// NewSev 创建一个空的 Server
// 这个 Server 中没有任何的预操作
// 如果要正常使用的话，第一步是挂载 sev.Router 中间件
func NewSev(name string) *Server {
	if name == "" {
		name = randStr(10)
	}
	return &Server{
		Name:    name,
		Router:  NewRouter(),
		Handler: NewSevHandler(),
	}
}

// Func 挂起一个以 rHFunc 作为处理函数的路由映射
func (r *Server) Func(routerPthTpl string, fn rHFunc) *Route {
	return r.Router.Handle(fn).Path(routerPthTpl)
}

// Handle 挂起一个以 http.Handler 为处理函数的路由映射
func (r *Server) Handle(routerPthTpl string, handler http.Handler) *Route {
	return r.Router.Handle(handler).Path(routerPthTpl)
}

// Static 挂起一个不打印目录列表的静态资源目录
func (r *Server) Static(routerPath, dirPth string) *Route {
	return r.StaticDir(routerPath, dirPth, true)
}

// StaticDir 挂起一个静态资源目录
// justFiles 为true时，系统将不回复目录请求
func (r *Server) StaticDir(routerPath, dirPth string, justFiles bool) *Route {
	var fs http.Handler
	if justFiles {
		fs = newJustFilesFS(dirPth, 2)
	} else {
		dir := http.Dir(dirPth)
		fs = http.FileServer(dir)
	}
	fs = http.StripPrefix(routerPath, fs)
	return r.Router.Handle(fs).PathMatch(routerPath, true, true)
}

var autoFiled = []string{}

func (r *Server) autoFile(filePth string) string {
	routerPth := filePth
	if routerPth[:1] == "." {
		routerPth = routerPth[1:]
	}
	for _, v := range autoFiled {
		if v == routerPth {
			return v
		}
	}
	r.File(routerPth, filePth)
	autoFiled = append(autoFiled, routerPth)
	r.Sort()
	return routerPth
}

// File 将文件路径映射到路由上
// 如果文件不存在或者文件路径错误，将回复404错误
//
// 需要注意的 这个处理函数每次调用都会读取文件，
// 正常使用需要考虑缓存问题
func (r *Server) File(routerPath, filePth string) *Route {
	// return r.Router.Handle(fileServer(filePth)).PathMatch(routerPath, true, false)
	return r.Router.Handle(rHFunc(func(vars *ReqVars) interface{} {
		data, err := loadFile(string(filePth))
		if err != nil {
			return mainFileResponser.NewErrResponse().Set(404, "", err)
		}
		return data
	})).PathMapping(routerPath)
}

// Bytes 创建一个以 []Byte 作为返回值的映射
func (r *Server) Bytes(routerPath string, genFn func() ([]byte, error)) *Route {
	// return r.Router.Handle(bytesServer(genFn)).PathMatch(routerPath, true, false)
	return r.Router.Handle(rHFunc(func(vars *ReqVars) interface{} {
		data, err := genFn()
		if err != nil {
			return mainBytesResponser.NewErrResponse().Set(500, "", err)
		}
		return data
	})).PathMapping(routerPath)
}

// String 创建一个以字符串作为返回值的映射
func (r *Server) String(routerPath string, genFn func() string) *Route {
	return r.Bytes(routerPath, func() ([]byte, error) {
		return []byte(genFn()), nil
	})
}

// HTML 创建一个 html 文件映射
//
// 文件列表可以接受 css js html
// 且最后会将各种 html 文件拼接之后返回
func (r *Server) HTML(routerPath string, filenames ...string) (*Route, func(func(h *Rhtml))) {
	handlers := []func(h *Rhtml){}
	nonHTMLfile := true
	htmlFilename := ""
	for _, v := range filenames {
		if strings.HasSuffix(v, ".html") {
			nonHTMLfile = false
			htmlFilename = v
		}
	}
	return r.Bytes(routerPath, func() ([]byte, error) {
			var HTML Rhtml
			if nonHTMLfile {
				HTML = NewEmptyHTML()
			} else {
				htmlContent, err := loadFile(htmlFilename)
				if err != nil {
					HTML = NewEmptyHTML()
				}
				HTML = Rhtml(htmlContent)
			}
			for _, handle := range handlers {
				handle(&HTML)
			}
			for _, filePth := range filenames {
				if !nonHTMLfile && filePth == htmlFilename {
					continue
				}
				switch path.Ext(filePth) {
				case ".css":
					cssFile, err := loadFile(filePth)
					if err != nil {
						continue
					}
					HTML.AppendStyle(cssFile)
				case ".js":
					jsFile, err := loadFile(filePth)
					if err != nil {
						continue
					}
					HTML.AppendScript(jsFile)
				case ".html":
					htmlFile, err := loadFile(filePth)
					if err != nil {
						continue
					}
					HTML.AppendBody(htmlFile)
				case ".wasm":
					if ok, err := pathExists(filePth); !ok || err != nil {
						continue
					}
					HTML.Wasm(r.autoFile(filePth))
				case ".go":
					// 不太好使
					if wPth, err := buildGoWasm(filePth); err == nil {
						HTML.Wasm(r.autoFile(wPth))
					} else {
						fmt.Println(err)
					}
				default:
					continue
				}
			}
			return HTML, nil
		}), func(fn func(h *Rhtml)) {
			handlers = append(handlers, fn)
		}
}

// Upload 创建一个 upload 服务
//
// 默认行为，设置 dir 的时候，如果文件夹不存在，会自动创建
//
// maxsize 时以 kb 为单位
// 例如最大限制 10MB 应该设置为 maxsize = 10 * 1024
func (r *Server) Upload(routerPath, dir string, maxsize int64, accept []string) (*Route, *uploadServer) {
	sev := newUploadServer(dir, maxsize, accept)
	return r.Router.Handle(sev).Path(routerPath), sev
}

// GET 创建一个只接受 GET 请求的映射
func (r *Server) GET(routerPthTpl string, fn rHFunc) *Route {
	return r.Router.Handle(fn).PathMatch(routerPthTpl, true, false).Methods("GET")
}

// POST 创建一个只接受 POST 请求的映射
func (r *Server) POST(routerPthTpl string, fn rHFunc) *Route {
	return r.Router.Handle(fn).PathMatch(routerPthTpl, true, false).Methods("POST")
}

// Group 路径分组功能
//
// 默认将匹配路径下的所有子路径
// 例如 Group('/api/')
// 将会包含 /api/1 /api/v1/a /api/ /api/a/b/c/d/e/f
func (r *Server) Group(routerPthTpl string) *Server {
	// [TODO] 应该支持带有参数的分组
	// 对于匹配到的参数，将作为vars设置到request之上
	// eg.
	// /user/{id:d+}

	subSev := NewSev(r.Name + "_" + routerPthTpl)

	subSev.Use(StripPrefixMid(routerPthTpl))
	subSev.Use(subSev.Router.Mid)

	r.Router.Handle(subSev.Handler).PathMatch(routerPthTpl, false, true)

	r.groups = append(r.groups, subSev)
	return subSev
}

// StartServer 使用自定义的 http.Server 进行服务
func (r *Server) StartServer(sev *http.Server) error {
	return r.Handler.StartServer(sev)
}

// StartServerTLS 使用自定义的 http.Server 进行HTTPS服务
func (r *Server) StartServerTLS(sev *http.Server, certFile, keyFile string) error {
	return r.Handler.StartServerTLS(sev, certFile, keyFile)
}

// Start 在指定端口打开HTTP服务
func (r *Server) Start(port string) error {
	return r.Handler.Start(port)
}

// Use 使用中间件
//
// 需要注意的中间件是有调用顺序的
// Use 函数会把中间件根据传输顺序添加到最后
func (r *Server) Use(funcs ...MiddlewareFunc) *Server {
	r.Handler.Use(funcs...)
	return r
}

// UseBefore 使用中间件，并加到前面
// 效果与 Use 相同，但是将会把中间件插入到队列的开始
func (r *Server) UseBefore(funcs ...MiddlewareFunc) *Server {
	r.Handler.UseBefore(funcs...)
	return r
}

// Sort 排序路由表
//
// 用于解决路由表被包含问题
// 例如 / 路由如果被注册，之后的同级路由都无法匹配
// 调用 Sort 之后，路由表将会把被包含的路由放到之前
//
// 注意，对于两个带有参数的映射时无法正常排序的
// 需要手动调整
func (r *Server) Sort() {
	r.Router.Sort()
	for _, g := range r.groups {
		g.Router.Sort()
	}
}

// IsSorted 当前路由表是否以排序
func (r *Server) IsSorted() bool {
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
