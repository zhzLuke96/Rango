package rango

import (
	"bytes"
	"fmt"
	"regexp"
)

type Rhtml []byte

var emptyHTML = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body></body>
</html>`)

var wasmEnv = []byte(`<script src="wasm_exec.js"></script>
<script>
// Check for wasm support.
if (!('WebAssembly' in window)) {
	alert('you need a browser with wasm support enabled :(');
}
// webassembly polyfill
if (!WebAssembly.instantiateStreaming) {
	WebAssembly.instantiateStreaming = async (resp, importObject) => {
		const source = await (await resp).arrayBuffer();
		return await WebAssembly.instantiate(source, importObject);
	};
}
if(!window.rango)rango = {};
if(!window.rango.go)window.rango.go = new Go();
function fetchGolangWasmRun(URL) {
	URL = URL || "main.wasm"
	return WebAssembly
		.instantiateStreaming(fetch(URL), rango.go.importObject)
		.then(result => rango.go.run(result.instance))
		.catch((err) => console.error(err));
}
</script>`)

func NewHTML(h string) Rhtml {
	return Rhtml([]byte(h))
}

func NewEmptyHTML() Rhtml {
	return Rhtml(emptyHTML)
}

func NewHTMLLoadFile(pth string) (Rhtml, error) {
	d, err := loadFile(pth)
	ht := Rhtml(d)
	ht.Check()
	return ht, err
}

func (h *Rhtml) Check() {
	if !h.Has("html") && !h.Has("body") && !h.Has("head") {
		nh := NewEmptyHTML()
		nh.AppendBody(*h)
		*h = nh
	}
}

func (h *Rhtml) Has(tagName string) bool {
	headIdx := bytes.Index(*h, []byte("<"+tagName+">"))
	tailIdx := bytes.Index(*h, []byte("</"+tagName+">"))
	return headIdx < tailIdx && headIdx != -1
}

func (h *Rhtml) appendTagChild(s []byte, tName string) {
	tail := "</" + tName + ">"
	newHTML := bytes.Replace(*h, []byte(tail), append(s, []byte(tail)...), 1)
	(*h) = Rhtml(newHTML)
}

func (h *Rhtml) AppendHead(s []byte) {
	h.appendTagChild(s, "head")
}

func (h *Rhtml) AppendBody(s []byte) {
	h.appendTagChild(s, "body")
}

func (h *Rhtml) AppendStyle(s []byte) {
	h.AppendHead(append(append([]byte("<style>"), s...), []byte("</style>")...))
}

func (h *Rhtml) AppendScript(s []byte) {
	h.AppendBody(append(append([]byte("<script>"), s...), []byte("</script>")...))
}

func (h *Rhtml) modifyTag(tagName string, inner []byte) {
	reg := regexp.MustCompile("<" + tagName + " ?[^>]*?>[\\s\\S]+?<\\/" + tagName + ">")
	*h = reg.ReplaceAll(*h, []byte(fmt.Sprintf("<%s>%s</%s>", tagName, inner, tagName)))
}

func (h *Rhtml) Inner(tagName string) []byte {
	match := h.InnerAll(tagName, 1)
	if match == nil || len(match) == 0 {
		return nil
	}
	return match[0]
}

func (h *Rhtml) InnerAll(tagName string, n int) [][]byte {
	reg := regexp.MustCompile("<" + tagName + " ?[^>]*?>([\\s\\S]+?)<\\/" + tagName + ">")
	match := reg.FindAllSubmatch(*h, n)
	var res [][]byte
	for _, m := range match {
		res = append(res, m[1])
	}
	if len(match) == 0 {
		return nil
	}
	return res
}

func (h *Rhtml) Title(t []byte) {
	h.modifyTag("title", t)
}

func (h *Rhtml) Body(b []byte) {
	h.modifyTag("body", b)
}

func (h *Rhtml) Style(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendStyle(d)
	}
}

func (h *Rhtml) Script(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendScript(d)
	}
}

func (h *Rhtml) Wasm(wasmURL string) {
	wasmExecRequired := bytes.Index(*h, []byte("src=\"wasm_exec.js\"")) != -1
	if !wasmExecRequired {
		h.AppendBody(wasmEnv)
	}
	h.AppendBody([]byte(`<script>fetchGolangWasmRun("` + wasmURL + `");</script>`))
}
