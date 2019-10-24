// https://github.com/golang/go/wiki/WebAssembly
// https://godoc.org/syscall/js

package main

import (
	"fmt"
	"runtime"
	"syscall/js"
)

var (
	consoleLog = js.Global().Get("console").Get("log")
	document   = js.Global().Get("document")
)

func evalJS(code string) {
	js.Global().Call("eval", code)
}

func queryDOM(sele string) js.Value {
	return document.Call("querySelector", sele)
}

func main() {
	done := make(chan struct{}, 0)
	consoleLog.Invoke("Rango.Rhtml.Wasm")
	js.Global().Call("eval", "console.log('hello, wasm!')")

	contentDiv := queryDOM("#app")

	contentDiv.Set("innerHTML", fmt.Sprintf(`<h1>%s</h1>
	runtime.GOOS => <span> %v </span><br>
	runtime.GOARCH => <span> %v </span><br>`, "Rango WebAssembly", runtime.GOOS, runtime.GOARCH))
	<-done
}
