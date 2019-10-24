cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

cd ./www
GOOS=js GOARCH=wasm go build -o index.wasm ./index.go
GOOS=js GOARCH=wasm go build -o rainbow.wasm ./mouse_rainbow.go
