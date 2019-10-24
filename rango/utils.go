package rango

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
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

func getDebugStackArr() []map[string]string {
	var ret []map[string]string
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

func sliceIndexPrefix(s []string, value string) int {
	if len(s) == 0 {
		return -1
	}
	value = strings.ToLower(value)
	for i, v := range s {
		if strings.HasPrefix(value, v) {
			return i
		}
	}
	return -1
}

func sliceHasPrefix(s []string, value string) bool {
	for _, v := range s {
		if v == "*" {
			return true
		}
	}
	return sliceIndexPrefix(s, value) != -1
}

// SaveFile 保存文件
func SaveFile(fb []byte, pth string) error {
	if exist, _ := pathExists(pth); exist {
		return nil
	}
	newFile, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := newFile.Write(fb); err != nil {
		return err
	}
	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fileMD5(file []byte) string {
	md5 := md5.New()
	md5.Write(file)
	return hex.EncodeToString(md5.Sum(nil))
}

func strOffset(t string, max int) int {
	sum := 0
	for _, c := range t {
		sum += int(c)
	}
	return sum % max
}

func cloneURL(u *url.URL) *url.URL {
	return &url.URL{
		Path:       u.Path,
		Scheme:     u.Scheme,
		Opaque:     u.Opaque,
		User:       u.User,
		RawPath:    u.RawPath,
		RawQuery:   u.RawQuery,
		Host:       u.Host,
		ForceQuery: u.ForceQuery,
		Fragment:   u.Fragment,
	}
}

var buildWasms = make(map[string]string)

func buildGoWasm(filePth string) (string, error) {
	// ----
	// WARING 部分系统中 无法正常使用
	// ----

	if v, ok := buildWasms[filePth]; ok {
		return v, nil
	}
	if ok, err := pathExists(filePth); !ok || err != nil {
		return "", err
	}
	hash := randStr(10)
	wasmfile := "./wasm/" + hash + ".wasm"
	// if pwd, err := os.Getwd(); err == nil {
	// 	if wasmfile[:1] == "." {
	// 		wasmfile = pwd + wasmfile[1:]
	// 	}
	// 	if filePth[:1] == "." {
	// 		filePth = pwd + filePth[1:]
	// 	}
	// }
	cmd := exec.Command("go", "build", "-o", wasmfile, filePth)
	cmd.Env = []string{"GOOS=js", "GOARCH=wasm"}
	buf, err := cmd.Output()
	if len(buf) == 0 || err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf(string(e.Stderr))
		}
		return "", fmt.Errorf("%s\n%s", string(buf), err.Error())
	}
	buildWasms[filePth] = wasmfile
	return wasmfile, nil
}
