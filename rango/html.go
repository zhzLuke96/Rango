package rango

import (
	"fmt"
	"regexp"
	"strings"
)

type html string

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
	return html(h)
}

func NewEmptyHTML() html {
	return html(emptyHTML)
}

func NewHTMLLoadFile(pth string) (html, error) {
	d, err := loadFile(pth)
	ht := html(string(d))
	ht.Check()
	return ht, err
}

func (h *html) Check() {
	if !h.Has("html") && !h.Has("body") && !h.Has("head") {
		nh := NewEmptyHTML()
		nh.AppendBody(string(*h))
		*h = nh
	}
}

func (h *html) Has(tagName string) bool {
	headIdx := strings.Index(string(*h), "<"+tagName+">")
	tailIdx := strings.Index(string(*h), "</"+tagName+">")
	return headIdx < tailIdx && headIdx != -1
}

func (h *html) appendTagChild(s string, tName string) {
	tail := "</" + tName + ">"
	newHTML := strings.Replace(string(*h), tail, s+tail, 1)
	(*h) = html(newHTML)
}

func (h *html) AppendHead(s string) {
	h.appendTagChild(s, "head")
}

func (h *html) AppendBody(s string) {
	h.appendTagChild(s, "body")
}

func (h *html) AppendStyle(s string) {
	h.AppendHead("<style>" + s + "</style>")
}

func (h *html) AppendScript(s string) {
	h.AppendBody("<script>;" + s + ";</script>")
}

func (h *html) modifyTag(tagName, inner string) {
	reg := regexp.MustCompile("<" + tagName + " ?[^>]*?>[\\s\\S]+?<\\/" + tagName + ">")
	reg.ReplaceAll([]byte(*h), []byte(fmt.Sprintf("<%s>%s</%s>", tagName, inner, tagName)))
}

func (h *html) Inner(tagName string) string {
	match := h.InnerAll(tagName, 1)
	if match == nil || len(match) == 0 {
		return ""
	}
	return match[0]
}

func (h *html) InnerAll(tagName string, n int) []string {
	reg := regexp.MustCompile("<" + tagName + " ?[^>]*?>([\\s\\S]+?)<\\/" + tagName + ">")
	match := reg.FindAllStringSubmatch(string(*h), n)
	var res []string
	for _, m := range match {
		res = append(res, m[1])
	}
	if len(match) == 0 {
		return nil
	}
	return res
}

func (h *html) Title(t string) {
	h.modifyTag("title", t)
}

func (h *html) Body(b string) {
	h.modifyTag("body", b)
}

func (h *html) Style(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendStyle(string(d))
	}
}

func (h *html) Script(cssPth string) {
	if d, err := loadFile(cssPth); err == nil {
		h.AppendScript(string(d))
	}
}
