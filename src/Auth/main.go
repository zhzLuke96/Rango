package Auth

import (
	"fmt"
	"net/http"

	core "../Core"
	"../utils"
)

const (
	CookieKey = ""
	HeaderKey = ""
)

var GlobalAuthManager = NewAuthManager()

type User struct {
	Name string
	UID  string

	password string

	Tickets map[string]string // liteToken:ciphertext
}

func SignDoc(u *User, t *Ticket) string {
	return u.UID + t.salt + u.Name + t.CreateUserID
}

type Ticket struct {
	salt         string // random string
	Token        string // publicKey
	key          string // privateKey
	CreateUserID string
	Passer       *Passer
}

func NewTicket(UserID string) *Ticket {
	pu, pr := NewKey()
	return &Ticket{
		salt:         utils.RandStr(20),
		Token:        pu,
		key:          pr,
		CreateUserID: UserID,
		Passer:       NewPasser(),
	}
}

func NewTopTicket(UserID string) *Ticket {
	t := NewTicket(UserID)
	t.Passer.AllowMap["/"] = newAuth(true, true, true, true)
	return t
}

func NewBanTicket(UserID string) *Ticket {
	t := NewTicket(UserID)
	t.Passer.BlackMap["/"] = newAuth(true, true, true, true)
	return t
}

func (t *Ticket) LiteToken() string {
	return t.Token[:16]
}

func (t *Ticket) IsSigned(u *User) bool {
	ciphertext, ok := u.Tickets[t.LiteToken()]
	singDoc := SignDoc(u, t)
	if ok {
		orig := Decrypt(ciphertext, t.key)
		if orig == singDoc {
			return true
		}
	}
	return false
}

type Passer struct {
	AllowMap map[string]int
	BlackMap map[string]int
}

func NewPasser() *Passer {
	return &Passer{
		AllowMap: make(map[string]int),
		BlackMap: make(map[string]int),
	}
}

func (p *Passer) MergePasser(inpass *Passer) {
	for path, auth := range inpass.AllowMap {
		if _, ok := p.AllowMap[path]; ok {
			p.AllowMap[path] = mergeAuth(p.AllowMap[path], auth)
		}
	}
	for path, auth := range inpass.BlackMap {
		if _, ok := p.BlackMap[path]; ok {
			p.BlackMap[path] = mergeAuth(p.BlackMap[path], auth)
		}
	}
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

func (a *authManager) Push(flag string, t *Ticket) {
	count := 1
	flagtemp := flag
	for {
		if _, ok := a.Tickets[flag]; ok {
			flag = fmt.Sprintf("%s%d", flagtemp, count)
		} else {
			a.Tickets[flag] = t
		}
	}
}

func (a *authManager) Query(u *User) *Passer {
	p := NewPasser()
	for _, t := range a.Tickets {
		if t.IsSigned(u) {
			p.MergePasser(t.Passer)
		}
	}
	return p
}

func (a *authManager) IsPassed(r *http.Request) bool {
	// [TODO]
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

func init() {
	GlobalUsers.AddAdmin("admin", "admin")
}
