package rango

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type matcher interface {
	Match(r *http.Request) bool
}

type headerMatcher map[string]string

func (m headerMatcher) Match(r *http.Request) bool {
	for k, v := range m {
		if r.Header.Get(k) != v {
			return false
		}
	}
	return true
}

type methodMatcher []string

func (m methodMatcher) Match(r *http.Request) bool {
	for _, method := range m {
		if strings.EqualFold(method, r.Method) {
			return true
		}
	}
	return false
}

type pathMappingMatcher string

func (p pathMappingMatcher) Match(r *http.Request) bool {
	return string(p) == r.URL.Path
}

type pathMatcher struct {
	Template string
	Regexp   regexp.Regexp
}

func (p pathMatcher) Match(r *http.Request) bool {
	pth := r.URL.Path
	matchArr := p.Regexp.FindStringSubmatch(pth)
	if len(matchArr) != 0 {
		vars := make(map[string]string)
		groupNames := p.Regexp.SubexpNames()
		for i, k := range groupNames {
			if i != 0 && k != "" {
				vars[k] = matchArr[i]
			}
		}
		*r = *setVars(r, vars)
		return true
	}
	return false
}

func newPathMatcher(tpl string, strictSlash, weakMatch bool) *pathMatcher {
	var idxs []int
	idxs, _ = braceIndices(tpl)
	template := tpl
	defaultPattern := "[^/]+"

	var end int
	pattern := bytes.NewBufferString("^")
	for i := 0; i < len(idxs); i += 2 {
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(tpl[idxs[i]+1:end-1], ":", 2)
		name := parts[0]
		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}

		fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), name, patt)
	}

	raw := tpl[end:]
	if strictSlash {
		if len(raw) != 0 && raw[len(raw)-1:] == "/" {
			raw = raw[:len(raw)-1]
		}
		pattern.WriteString(regexp.QuoteMeta(raw))
		pattern.WriteString("[/]?")
	} else {
		pattern.WriteString(regexp.QuoteMeta(raw))
	}
	if !weakMatch {
		pattern.WriteString("$")
	}

	reg, _ := regexp.Compile(pattern.String())
	return &pathMatcher{
		Template: template,
		Regexp:   *reg,
	}
}

// braceIndices returns the first level curly brace indices from a string.
// It returns an error in case of unbalanced braces.
func braceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("rango: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("rango: unbalanced braces in %q", s)
	}
	return idxs, nil
}
