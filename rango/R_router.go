package rango

import (
	"net/http"
	"sort"
)

type Router struct {
	routes []*Route
}

func NewRouter() *Router {
	return &Router{}
}

// Match matches registered routes against the request.
func (r *Router) Match(req *http.Request, rte *Route) bool {
	for _, route := range r.routes {
		if route.Match(req) {
			*rte = *route
			return true
		}
	}
	return false
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var route Route
	if !r.Match(req, &route) {
		// http.NotFoundHandler().ServeHTTP(w, req)
		notFoundResponser.NewErrResponse().Push(w, 404, "Not Found", nil)
		return
	}
	if route.BeforeH != nil && !route.BeforeH(w, req) {
		return
	}
	route.Handler.ServeHTTP(w, req)
	if route.BeforeH != nil {
		route.AfterH(w, req)
	}
}

func (r *Router) Mid(w ResponseWriteBody, req *http.Request, next func()) {
	r.ServeHTTP(w, req)
	next()
}

func (r *Router) Func(fn rHFunc) *Route {
	return r.Handle(fn)
}

func (r *Router) Handle(handler http.Handler) *Route {
	route := &Route{Handler: handler}
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Registe(conf map[string]http.Handler) {
	for pathTpl, h := range conf {
		route := Route{Handler: h}
		route.Path(pathTpl)
		r.routes = append(r.routes, &route)
	}
}

// for sort

func routePathLen(r *Route) int {
	return len(r.PathTpl)
}

func (r *Router) Sort() {
	sort.Sort(r)
}

func (r *Router) IsSorted() bool {
	return sort.IsSorted(r)
}

func (r *Router) Len() int {
	return len(r.routes)
}

func (r *Router) Less(i, j int) bool {
	iLen := routePathLen(r.routes[i])
	jLen := routePathLen(r.routes[j])
	return iLen > jLen
}

func (r *Router) Swap(i, j int) {
	r.routes[i], r.routes[j] = r.routes[j], r.routes[i]
}
