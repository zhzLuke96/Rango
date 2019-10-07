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
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
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

func anyLess(a, b interface{}) bool {
	switch a.(type) {
	case nil:
		return true
	case string:
		return strings.Compare(a.(string), b.(string)) == -1
	case int:
		return a.(int) < b.(int)
	case float64:
		return a.(float64) < b.(float64)
	case bool:
		return a.(bool)
	default:
		return true
	}
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

func excluding(in map[string]string, ex ...string) map[string]string {
	if len(ex) == 0 {
		return in
	}
	ret := make(map[string]string)
	for k, v := range in {
		isPassKey := false
		for _, v := range ex {
			if k == v {
				isPassKey = true
				break
			}
		}
		if isPassKey {
			continue
		}
		ret[k] = v
	}
	return ret
}

func queryEqule(a, b interface{}) bool {
	switch at := a.(type) {
	case nil:
		return b == nil
	case bool:
		if bt, ok := b.(bool); ok {
			return at == bt
		}
		return false
	case string, int, float64, float32, uint, complex64, complex128:
		return fmt.Sprint(a) == fmt.Sprint(b)
	default:
		return false
	}
}

func toFloat(v interface{}) (float64, error) {
	switch vt := v.(type) {
	case string:
		return strconv.ParseFloat(vt, 64)
	case int:
		return float64(vt), nil
	case float64:
		return vt, nil
	case float32:
		return float64(vt), nil
	default:
		return 0, fmt.Errorf("cant parsing")
	}
}
