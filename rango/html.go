package rango

import (
	"bytes"
	"fmt"
	"regexp"
)

type html []byte

const emptyHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body></body>
</html>`

func NewHTML(h string) html {
	return html([]byte(h))
}

func NewEmptyHTML() html {
	return html(emptyHTML)
}

func NewHTMLLoadFile(pth string) (html, error) {
	d, err := loadFile(pth)
	ht := html(d)
	ht.Check()
	return ht, err
}

func (h *html) Check() {
	if !h.Has("html") && !h.Has("body") && !h.Has("head") {
		nh := NewEmptyHTML()
		nh.AppendBody(*h)
		*h = nh
	}
}

func (h *html) Has(tagName string) bool {
	headIdx := bytes.Index(*h, []byte("<"+tagName+">"))
	tailIdx := bytes.Index(*h, []byte("</"+tagName+">"))
	return headIdx < tailIdx && headIdx != -1
}

func (h *html) appendTagChild(s []byte, tName string) {
	tail := "</" + tName + ">"
	newHTML := bytes.Replace(*h, []byte(tail), append(s, []byte(tail)...), 1)
	(*h) = html(newHTML)
}

func (h *html) AppendHead(s []byte) {
	h.appendTagChild(s, "head")
}

func (h *html) AppendBody(s []byte) {
	h.appendTagChild(s, "body")
}

func (h *html) AppendStyle(s []byte) {
	h.AppendHead(append(append([]byte("<style>"), s...), []byte("</style>")...))
}

func (h *html) AppendScript(s []byte) {
	h.AppendBody(append(append([]byte("<script>"), s...), []byte("</script>")...))
}

func (h *html) modifyTag(tagName string, inner []byte) {
	reg := regexp.MustCompile("<" + tagName + " ?[^>]*?>[\\s\\S]+?<\\/" + tagName + ">")
	reg.ReplaceAll(*h, []byte(fmt.Sprintf("<%s>%s</%s>", tagName, inner, tagName)))
}

func (h *html) Inner(tagName string) []byte {
	match := h.InnerAll(tagName, 1)
	if match == nil || len(match) == 0 {
		return nil
	}
	return match[0]
}

func (h *html) InnerAll(tagName string, n int) [][]byte {
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

func (h *html) Title(t []byte) {
	h.modifyTag("title", t)
}

func (h *html) Body(b []byte) {
	h.modifyTag("body", b)
}

func (h *html) Style(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendStyle(d)
	}
}

func (h *html) Script(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendScript(d)
	}
}
