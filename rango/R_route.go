package rango

import (
	"net/http"
	"regexp"
	"strings"
)

type hookFunc func(http.ResponseWriter, *http.Request) bool

type Route struct {
	BeforeH hookFunc
	AfterH  hookFunc
	Handler http.Handler
	PathTpl string
	// List of matchers.
	matchers []matcher
	err      error
}

func (r *Route) Match(req *http.Request) bool {
	if r.err != nil {
		return false
	}
	for _, m := range r.matchers {
		if m.Match(req) == false {
			return false
		}
	}
	return true
}

func (r *Route) Before(h hookFunc) {
	r.BeforeH = h
}

func (r *Route) After(h hookFunc) {
	r.AfterH = h
}

func (r *Route) AddMatcher(m matcher) *Route {
	if r.err != nil {
		return r
	}
	r.matchers = append(r.matchers, m)
	return r
}

// 设置http header路由
func (r *Route) Headers(pairs ...string) *Route {
	var headers map[string]string
	headers, r.err = mapFromPairs(pairs...)
	return r.AddMatcher(headerMatcher(headers))
}

// 设置methods路由
func (r *Route) Methods(methods ...string) *Route {
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.AddMatcher(methodMatcher(methods))
}

var routerRegexp = regexp.MustCompile("\\{.+?:.+?\\}")

func isRouterRegexp(tpl string) bool {
	return routerRegexp.FindString(tpl) != ""
}

// 设置path路由
func (r *Route) Path(tpl string) *Route {
	r.PathTpl = tpl
	if isRouterRegexp(tpl) {
		return r.PathMatch(tpl, false)
	}
	return r.AddMatcher(pathMappingMatcher(tpl))
}

// 设置path路由
// 设置strictSlash将会匹配尾部的 "/"
// 比如 PathMatch("/user", true)
// 将会匹配 host/user host/user/
//
// PathMatch("/user", false)
// 则只匹配 host/user

func (r *Route) PathMatch(tpl string, strictSlash bool) *Route {
	r.PathTpl = tpl
	return r.AddMatcher(newPathMatcher(tpl, strictSlash))
}
