package Auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	core "../Core"
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

type authManager struct {
	Tickets map[string]*Ticket
}

func NewAuthManager() *authManager {
	a := &authManager{
		Tickets: make(map[string]*Ticket),
	}
	return a
}

func (a *authManager) Get(name string) *Ticket {
	return a.Tickets[name]
}

func (a *authManager) Push(t *Ticket) {
	a.Tickets[t.Salt] = t
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
		return strings.Replace(c.String(), CookieKey+"=", "", 1), true
	}
	return "", false
}

func GetUserFromReq(r *http.Request) *User {
	t, ok := GetTokenFromReq(r)
	if !ok {
		return nil
	}
	cipher, err := base64.StdEncoding.DecodeString(t)
	if err != nil {
		log.Printf("[Auth ERROR]:%s\n", err.Error())
		return nil
	}
	json, err := Decrypt(string(cipher), SystemUser.PublicKey)
	if err != nil {
		log.Printf("[Auth ERROR]:%s\n", err.Error())
		return nil
	}
	token, err := NewTokenFromJSON(json)
	if err != nil {
		log.Printf("[Passed check ERROR]:%s\n", err.Error())
		return nil
	}
	return GlobalUsers.GetUser(token.TargetUserName)
}

func SetTokenInRes(w core.ResponseWriteBody, t *Token) string {
	c := new(http.Cookie)
	c.Expires = time.Now().Add(7 * 24 * time.Hour)
	c.Name = CookieKey
	criptext := SystemUser.SignDoc(string(t.JSON()))
	c.Value = base64.StdEncoding.EncodeToString([]byte(criptext))
	c.Path = "/"
	http.SetCookie(w, c)
	return c.Value
}

func (a *authManager) IsPassed(r *http.Request) bool {
	if t, ok := GetTokenFromReq(r); ok {
		cipher, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			log.Printf("[Auth ERROR]:%s\n", err.Error())
			return false
		}
		json, err := Decrypt(string(cipher), SystemUser.PublicKey)
		if err != nil {
			log.Printf("[Auth ERROR]:%s\n", err.Error())
			return false
		}
		Token, err := NewTokenFromJSON(json)
		if err != nil {
			log.Printf("[Passed check ERROR]:%s\n", err.Error())
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
			ciphertext := SetTokenInRes(w, token)
			resp.Msg = "Login Success."
			resp.Status = "SUCCESS"
			resp.Data = make(map[string]string)
			resp.Data["TOKEN"] = ciphertext
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
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var user struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
	}
	if err := json.Unmarshal(body, &user); err == nil {
		if GlobalUsers.NameCanUse(user.Name) {
			GlobalUsers.AddOne(user.Name, user.Pass)
			resp.Status = "SUCCESS"
			resp.Msg = "Registe Success"
		} else {
			resp.Status = "ERROR"
			resp.Msg = "Name Exist"
		}
	} else {
		resp.Status = "ERROR"
		resp.Msg = err.Error()
	}
	w.Write(resp.JSON())
}

func (a *authManager) TicketHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var ticketRequest struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(body, &ticketRequest); err == nil {
		switch ticketRequest.Method {
		case "new":
			a.NewTicketHandler(w, r)
		case "sign":
			a.SignTicketHandler(w, r)
		default:
			resp.Status = "ERROR"
			resp.Msg = "Wrong Request Method"
			w.Write(resp.JSON())
		}
		return
	} else {
		resp.Status = "ERROR"
		resp.Msg = err.Error()
	}
	w.Write(resp.JSON())
}

func (a *authManager) NewTicketHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var ticketRequest struct {
		Passer Passer `json:"pass"`
	}
	if err := json.Unmarshal(body, &ticketRequest); err == nil {
		if callerUser := GetUserFromReq(r); callerUser != nil {
			t := NewTicket(callerUser.UID)
			t.Passer.MergePasser(&ticketRequest.Passer)
			GlobalAuthManager.Push(t)

			resp.Status = "SUCCESS"
			resp.Msg = "Create Success"
			resp.Data = make(map[string]string)
			resp.Data["SALT"] = t.Salt
		} else {
			resp.Status = "ERROR"
			resp.Msg = "Unauthorized"
		}
	} else {
		resp.Status = "ERROR"
		resp.Msg = err.Error()
	}
	w.Write(resp.JSON())
}

func (a *authManager) SignTicketHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var ticketRequest struct {
		TicketName  string `json:"name"`
		TargetUName string `json:"target"`
	}
	if err := json.Unmarshal(body, &ticketRequest); err == nil {
		if callerUser := GetUserFromReq(r); callerUser != nil {
			setterUser := GlobalUsers.GetUser(ticketRequest.TargetUName)
			tik := GlobalAuthManager.Get(ticketRequest.TicketName)
			doc := callerUser.SignTicket(setterUser, tik)
			setterUser.TicketProofs[ticketRequest.TicketName] = doc
		} else {
			resp.Status = "ERROR"
			resp.Msg = "Unauthorized"
		}
	} else {
		resp.Status = "ERROR"
		resp.Msg = err.Error()
	}
	w.Write(resp.JSON())
}

func (a *authManager) GetAllPubKeyHandler(w core.ResponseWriteBody, r *http.Request) {
	return
}
