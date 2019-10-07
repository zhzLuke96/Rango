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
		notFoundResponser.NewErrResponse().PushReset(w, 404, "Not Found", nil)
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

func (r *Router) Mid(w ResponseWriteBody, req *http.Request, next MiddleNextFunc) {
	r.ServeHTTP(w, req)
	next(w, req)
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
	iR := r.routes[i].PathMatcher
	jR := r.routes[j].PathMatcher
	if iR == nil || jR == nil {
		return i > j
	}

	irmapping := true
	if _, ok := iR.(*pathMatcher); ok {
		irmapping = false
	}
	jrmapping := true
	if _, ok := jR.(*pathMatcher); ok {
		jrmapping = false
	}
	if irmapping && jrmapping {
		return len(iR.(pathMappingMatcher)) > len(jR.(pathMappingMatcher))
	}
	if !irmapping && jrmapping {
		return !iR.(*pathMatcher).Regexp.MatchString(string(jR.(pathMappingMatcher)))
	}
	return i > j
}

func (r *Router) Swap(i, j int) {
	r.routes[i], r.routes[j] = r.routes[j], r.routes[i]
}
