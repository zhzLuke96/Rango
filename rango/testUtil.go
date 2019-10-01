package rango

import "strings"

func includeString(s, t string) bool {
	return strings.Index(s, t) != -1
}
