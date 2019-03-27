package Router

import "net/http"
import "../core"

type Router struct {
	routes []*Route
}

type HandleFunc func(http.ResponseWriter,*http.ReadRequest)

// Match matches registered routes against the request.
func (r *Router) Match(req *http.Request, h *http.Handler) bool {
	for _, route := range r.routes {
		if route.Match(req) {
			h = &route.Handler
			return true
		}
	}
	return false
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler
	if r.Match(req, &handler) {
		handler = http.NotFoundHandler()
	}
	handler.ServeHTTP(w, req)
}

func (r *Router) Mid(w core.ResponseWriteBody, req *http.Request, next func()) {
	r.ServeHTTP(w.ResponseWriter, req)
	next()
}

func (r *Router) HandleFunc(pathTpl string,fn HandleFunc) *Route{
	route := &Route{
		Handler: http.HandlerFunc(fn)
	}.Path(pathTpl)
	r.routes = append(r.routes,route)
	return route
}

func (r *Router) Handler(pathTpl string,handler http.Handler) *route{
	route := &Route{
		Handler: handler
	}.Path(pathTpl)
	r.routes = append(r.routes,route)
	return route
}

func (r *Router)Registe(conf map[Route]HandleFunc){
	for route,fn := range conf{
		route.Handler = http.HandlerFunc(fn)
		r.routes = append(r.routes, route)
	}
}
