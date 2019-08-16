package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"../core"
)

type ResponesAuth struct {
	Status string            `json:"status"`
	Msg    string            `json:"msg"`
	Data   map[string]string `json:"D"`
}

func (r *ResponesAuth) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

const (
	CookieKey = "RangoToken"
	HeaderKey = "RANGO_TOKEN"
)

var (
	defaultMemDB  = newMemSimpleDB()
	GlobalManager = NewManager(*defaultMemDB)
)

func DefaultAuthInit() {
	GlobalManager.SoftInitDB()
	sysUser := GlobalManager.SystemUser()
	// randPassword := utils.RandStr(8)
	// SystemUser.Password = randPassword
	fmt.Printf("=== [Auth] middleware inited ===\n")
	if sysUser != nil {
		sysUser.GenNewKey()
		// fmt.Printf("System password: %s\n", randPassword)
		fmt.Printf("System Public Key: %s\n\n", sysUser.GetPublicKey())
	}
}

func GetTokenFromReq(r *http.Request) (t string, ok bool) {
	if c, err := r.Cookie(CookieKey); err != nil {
		if h := r.Header.Get(HeaderKey); h != "" {
			return h, true
		}
	} else {
		return strings.Replace(c.String(), CookieKey+"=", "", 1), true
	}
	return "", false
}

func GetUserFromReq(r *http.Request) *User {
	sysUser := GlobalManager.SystemUser()
	t, ok := GetTokenFromReq(r)
	if !ok {
		return nil
	}
	cipher, err := base64.StdEncoding.DecodeString(t)
	if err != nil {
		log.Printf("[Auth ERROR]:%s\n", err.Error())
		return nil
	}
	json, err := Decrypt(string(cipher), sysUser.PublicKey)
	if err != nil {
		log.Printf("[Auth ERROR]:%s\n", err.Error())
		return nil
	}
	token, err := NewTokenFromJSON(json)
	if err != nil {
		log.Printf("[Passed check ERROR]:%s\n", err.Error())
		return nil
	}
	u, err := GlobalManager.DB.FindUser(token.TargetUserName)
	if err != nil {
		log.Printf("[DB ERROR]:%s\n", err.Error())
		return nil
	}
	return u
}

func SetTokenInRes(w core.ResponseWriteBody, t *Token) string {
	sysUser := GlobalManager.SystemUser()
	c := new(http.Cookie)
	c.Expires = time.Now().Add(7 * 24 * time.Hour)
	c.Name = CookieKey
	criptext := sysUser.SignDoc(string(t.JSON()))
	c.Value = base64.StdEncoding.EncodeToString([]byte(criptext))
	c.Path = "/"
	http.SetCookie(w, c)
	return c.Value
}
