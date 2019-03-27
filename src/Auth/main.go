package Auth

import (
	"net/http"

	"../core"
)

const (
	CookieKey = ""
	HeaderKey = ""
)

type User struct {
	Name    string
	UID     string
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
	Passer       Passer
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
	Tickets map[string]Ticket
}

func (a *authManager) Query(u *User) *Passer {
	p := NewPasser()
	for _, t := range a.Tickets {
		if t.IsSigned(u) {
			p.MergePasser(&t.Passer)
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
