package Auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	core "../Core"
)

type ResponesAuth struct {
	Status string             `json:"status"`
	Msg    string             `json:"msg"`
	Data   *map[string]string `json:"D"`
}

func (r *ResponesAuth) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

type authManager struct {
	Tickets map[string]*Ticket
}

func NewAuthManager() *authManager {
	a := &authManager{
		Tickets: make(map[string]*Ticket),
	}
	return a
}

func (a *authManager) Push(t *Ticket) {
	a.Tickets[t.salt] = t
}

func (a *authManager) Query(u *User) *Passer {
	p := NewPasser()
	for salt := range u.TicketProofs {
		t := a.Tickets[salt]
		if t.IsSigned(u) {
			p.MergePasser(t.Passer)
		}
	}
	return p
}

func GetTokenFromReq(r *http.Request) (t string, ok bool) {
	if c, err := r.Cookie(CookieKey); err != nil {
		if h := r.Header.Get(HeaderKey); h != "" {
			return h, true
		}
	} else {
		return c.String(), true
	}
	return "", false
}

func SetTokenInRes(w core.ResponseWriteBody, t *Token) {
	c := new(http.Cookie)
	c.Expires = time.Now().Add(time.Hour)
	c.Name = CookieKey
	c.Value = base64.StdEncoding.EncodeToString(t.JSON())
	http.SetCookie(w, c)
}

func (a *authManager) IsPassed(r *http.Request) bool {
	if t, ok := GetTokenFromReq(r); ok {
		cipher, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			log.Printf("[Auth ERROR]:\n%s\n", err.Error())
			return false
		}
		json, err := Decrypt(string(cipher), SystemUser.PublicKey)
		if err != nil {
			log.Printf("[Auth ERROR]:\n%s\n", err.Error())
			return false
		}
		Token, err := NewTokenFromJSON(json)
		if err != nil {
			log.Printf("[Passed check ERROR]:\n%s\n", err.Error())
			return false
		}
		return Token.IsAllowReq(r)
	}
	return false
}

func (a *authManager) Mid(w core.ResponseWriteBody, r *http.Request, next func()) {
	if a.IsPassed(r) {
		next()
	} else {
		w.WriteHeader(403)
		w.Write([]byte("Unauthorized."))
	}
}

func (a *authManager) LoginHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var user struct {
		Name      string `json:"name"`
		Pass      string `json:"pass"`
		AutoLogin bool   `json:"auto"`
	}
	if err := json.Unmarshal(body, &user); err == nil {
		if ok, err := GlobalUsers.Login(user.Name, user.Pass); ok {
			dur, _ := time.ParseDuration(fmt.Sprintf("%dh", 24*7))
			token := NewToken(dur, user.Name)
			SetTokenInRes(w, token)
			resp.Msg = "Login Success."
			resp.Status = "SUCCESS"
			*resp.Data = make(map[string]string)
			(*resp.Data)["TOKEN"] = base64.StdEncoding.EncodeToString(token.JSON())
		} else {
			resp.Status = "ERROR"
			resp.Msg = err.Error()
		}
	} else {
		resp.Status = "ERROR"
		resp.Msg = err.Error()
	}
	w.Write(resp.JSON())
}

func (a *authManager) RegisteHandler(w core.ResponseWriteBody, r *http.Request) {

}

func (a *authManager) TicketHandler(w core.ResponseWriteBody, r *http.Request) {

}

func (a *authManager) GetAllPubKeyHandler(w core.ResponseWriteBody, r *http.Request) {

}
