package router

import (
	"net/http"
	"time"
)

var varsKey = time.Now().Unix()

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	if rv := contextGet(r, varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

func setVars(r *http.Request, val interface{}) *http.Request {
	return contextSet(r, varsKey, val)
}
