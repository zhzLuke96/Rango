package rango

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"time"
)

const strs = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var randsrc = rand.NewSource(time.Now().UnixNano())

// randStr rand string
func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = strs[randsrc.Int63()%int64(len(strs))]
	}
	return string(b)
}

var stackRe = regexp.MustCompile(`([\w.()*/]+)\(.*?\)\n\t(\S+?):(\d+?) .+`)

func getDebugStackArr() []interface{} {
	var ret []interface{}
	if !isDebugOn() {
		return ret
	}
	stackText := debug.Stack()
	matchs := stackRe.FindAllSubmatch(stackText, -1)
	for _, v := range matchs {
		ret = append(ret, map[string]string{
			"func":    string(v[1]),
			"file":    string(v[2]),
			"lineNum": string(v[3]),
		})
	}
	return ret
}

func httpCodeText(code int) string {
	if http.StatusText(code) == "" {
		return "unknow"
	}
	return http.StatusText(code)
}

func loadFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func loadJSONFile(filename string) (map[string]interface{}, error) {
	data, err := loadFile(filename)
	if err != nil {
		return nil, err
	}
	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func mustReadJSONFile(filenames ...string) (map[string]interface{}, string) {
	for _, name := range filenames {
		if ret, err := loadJSONFile(name); err == nil {
			return ret, name
		}
	}
	return nil, ""
}

func contentType(filePth string) string {
	return mime.TypeByExtension(filepath.Ext(filePth))
}
