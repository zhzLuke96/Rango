package Router

import (
	"net/http"
	"strings"

	"../core"
)

type Route struct {
	Handler core.RangoSevHandler
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

func (r *Route) addMatcher(m matcher) *Route {
	if r.err != nil {
		return r
	}
	r.matchers = append(r.matchers, m)
	return r
}

func (r *Route) Headers(pairs ...string) *Route {
	var headers map[string]string
	headers, r.err = mapFromPairs(pairs...)
	return r.addMatcher(headerMatcher(headers))
}

func (r *Route) Methods(methods ...string) *Route {
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(methodMatcher(methods))
}

func (r *Route) Path(tpl string) *Route {
	return r.PathMatch(tpl, false)
}
func (r *Route) PathMatch(tpl string, strictSlash bool) *Route {
	return r.addMatcher(NewPathMatcher(tpl, strictSlash))
}
