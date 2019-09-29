package rango

import (
	"context"
	"net/http"
	"time"
)

// ctx
func contextGet(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func contextSet(r *http.Request, key, val interface{}) *http.Request {
	if val == nil {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

// VARS
var varsKey = time.Now().Unix()

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	if rv := contextGet(r, varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

func SetVars(r *http.Request, val interface{}) *http.Request {
	return contextSet(r, varsKey, val)
}
