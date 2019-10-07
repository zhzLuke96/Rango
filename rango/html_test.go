package rango

import (
	"testing"
)

const (
	failFlag = "×"
	passFlag = "✔"
)

var (
	testHTML  = "<h1>Rango.HTML</h1>"
	testCSS   = []byte("html,body{padding:0;margin:0;}")
	testJS    = []byte("console.log('hello world!')")
	testTitle = []byte("rango.HTML")
)

func TestNewEmptyHTML(t *testing.T) {
	h := NewEmptyHTML()

	if !h.Has("html") && !h.Has("body") && !h.Has("head") {
		t.Fatalf("%s NewEmptyHTML() is fatal, cant resolving.", failFlag)
	}

	title := h.Inner("title")
	if string(title) != "Document" {
		t.Fatalf("%s NewEmptyHTML() is fatal, need %v but %v", failFlag, "Document", title)
	}
	t.Logf("%s NewEmptyHTML() passed", passFlag)
}

func TestNewHTML(t *testing.T) {
	h := NewHTML(testHTML)

	if h.Has("html") || h.Has("body") || h.Has("head") {
		t.Fatalf("%s NewHTML() is fatal, cant resolving.", failFlag)
	}

	t.Logf("%s NewHTML() passed", passFlag)
}

func TestCheck(t *testing.T) {
	h := NewHTML(testHTML)

	h.Check()

	if !h.Has("html") && !h.Has("body") && !h.Has("head") {
		t.Fatalf("%s Check() is fatal, cant resolving.", failFlag)
	}

	t.Logf("%s NewHTML() passed", passFlag)
}

func TestAppendScript(t *testing.T) {
	h := NewEmptyHTML()

	h.AppendScript(testJS)

	if !includeBytes(h.Inner("script"), testJS) {
		t.Fatalf("%s AppendScript() is fatal, cant resolving.", failFlag)
	}

	t.Logf("%s AppendScript() passed", passFlag)
}

func TestAppendStyle(t *testing.T) {
	h := NewEmptyHTML()

	h.AppendStyle(testCSS)

	if !h.Has("style") {
		t.Fatalf("%s AppendStyle() is fatal, cant resolving.", failFlag)
	}

	if !includeBytes(h.Inner("style"), testCSS) {
		t.Fatalf("%s AppendStyle() is fatal, cant resolving.", failFlag)
	}

	t.Logf("%s AppendStyle() passed", passFlag)
}

func TestTitle(t *testing.T) {
	h := NewEmptyHTML()

	h.Title(testTitle)

	title := h.Inner("title")
	if string(title) == string(testTitle) {
		t.Fatalf("%s Title() is fatal, need %v, but %v.", failFlag, title, testTitle)
	}

	t.Logf("%s Title() passed", passFlag)
}

func TestBody(t *testing.T) {
	h := NewEmptyHTML()

	h.Body([]byte(testHTML))

	body := h.Inner("body")
	if string(body) == string(testHTML) {
		t.Fatalf("%s Title() is fatal, need %v, but %v.", failFlag, body, testHTML)
	}

	t.Logf("%s Title() passed", passFlag)
}
