package Router

import "net/http"
import core "../Core"

type Router struct {
	routes []*Route
}

type HandleFunc func(http.ResponseWriter, *http.Request)

// Match matches registered routes against the request.
func (r *Router) Match(req *http.Request, h *core.RangoSevHandler) bool {
	for _, route := range r.routes {
		if route.Match(req) {
			*h = route.Handler
			return true
		}
	}
	return false
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler core.RangoSevHandler
	if !r.Match(req, &handler) {
		handler = core.RangoSevHandler{
			Handler: http.NotFoundHandler(),
		}
	}
	handler.ServeHTTP(w, req)
}

func (r *Router) Mid(w core.ResponseWriteBody, req *http.Request, next func()) {
	r.ServeHTTP(w, req)
	next()
}

func (r *Router) HandleFunc(pathTpl string, fn HandleFunc) *Route {
	h := core.RangoSevHandler{}
	h.HandleFunc(fn)
	route := &Route{
		Handler: h,
	}
	route.Path(pathTpl)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Handler(pathTpl string, handler http.Handler) *Route {
	route := &Route{
		Handler: core.RangoSevHandler{
			Handler: handler,
		},
	}
	route.Path(pathTpl)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Registe(conf map[string]http.Handler) {
	for pathTpl, h := range conf {
		route := Route{}
		route.Path(pathTpl)
		route.Handler = core.RangoSevHandler{
			Handler: h,
		}
		r.routes = append(r.routes, &route)
	}
}
