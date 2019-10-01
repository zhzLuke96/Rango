package rango

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type hookFunc func(http.ResponseWriter, *http.Request) bool

type Route struct {
	BeforeH     hookFunc
	AfterH      hookFunc
	Handler     http.Handler
	PathMatcher interface{}

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
	if isRouterRegexp(tpl) {
		return r.PathMatch(tpl, false)
	}
	r.PathMatcher = pathMappingMatcher(tpl)
	return r.AddMatcher(r.PathMatcher.(pathMappingMatcher))
}

// 设置path路由
// 设置strictSlash将会匹配尾部的 "/"
// 比如 PathMatch("/user", true)
// 将会匹配 host/user host/user/
//
// PathMatch("/user", false)
// 则只匹配 host/user

func (r *Route) PathMatch(tpl string, strictSlash bool) *Route {
	r.PathMatcher = newPathMatcher(tpl, strictSlash)
	return r.AddMatcher(r.PathMatcher.(*pathMatcher))
}

// util

// checkPairs returns the count of strings passed in, and an error if
// the count is not an even number.
func checkPairs(pairs ...string) (int, error) {
	length := len(pairs)
	if length%2 != 0 {
		return length, fmt.Errorf(
			"rango: number of parameters must be multiple of 2, got %v", pairs)
	}
	return length, nil
}

// mapFromPairsToString converts variadic string parameters to a
// string to string map.
func mapFromPairs(pairs ...string) (map[string]string, error) {
	length, err := checkPairs(pairs...)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, length/2)
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m, nil
}
