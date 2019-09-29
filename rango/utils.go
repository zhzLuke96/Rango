package rango

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"mime"
	"net/http"
	"os"
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
