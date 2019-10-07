package rango

import (
	"bytes"
	"strings"
)

func includeString(s, t string) bool {
	return strings.Index(s, t) != -1
}

func includeBytes(s, t []byte) bool {
	return bytes.Index(s, t) != -1
}
