package Router

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
	for _, me := range m {
		if me == r.Method {
			return true
		}
	}
	return false
}

type pathMatcher struct {
	Template string
	Regexp   regexp.Regexp
	VarsN    []string
}

func (p *pathMatcher) Match(r *http.Request) bool {
	return p.Regexp.MatchString(r.URL.Path)
}

func NewPathMatcher(tpl string, strictSlash bool) *pathMatcher {
	var idxs []int
	idxs, _ = braceIndices(tpl)
	template := tpl
	defaultPattern := "[^/]+"

	varsN := make([]string, len(idxs)/2)
	var end int
	pattern := bytes.NewBufferString("")
	for i := 0; i < len(idxs); i += 2 {
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(tpl[idxs[i]+1:end-1], ":", 2)
		name := parts[0]
		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}

		fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(raw), patt)
		varsN[i/2] = name
	}

	raw := tpl[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	if strictSlash {
		pattern.WriteString("[/]?")
	}

	reg, _ := regexp.Compile(pattern.String())
	return &pathMatcher{
		Template: template,
		Regexp:   *reg,
		VarsN:    varsN,
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
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}

// checkPairs returns the count of strings passed in, and an error if
// the count is not an even number.
func checkPairs(pairs ...string) (int, error) {
	length := len(pairs)
	if length%2 != 0 {
		return length, fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
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
