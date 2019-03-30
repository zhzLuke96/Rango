package Auth

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"../utils"
)

const (
	CookieKey = "RangoToken"
	HeaderKey = "RANGO_TOKEN"
)

var GlobalAuthManager = NewAuthManager()

type User struct {
	Name     string
	UID      string
	Password string

	PublicKey  string
	privateKey string

	TicketProofs map[string]string // salt:proof
}

func NewUser(Name, Uid, Pass string) *User {
	return &User{
		Name:         Name,
		UID:          Uid,
		Password:     Pass,
		TicketProofs: make(map[string]string),
	}
}

func (u *User) IsMySigned(proof, salt string, subU *User) bool {
	decode, err := Decrypt(proof, u.PublicKey)
	if err != nil {
		log.Printf("[Signed Check ERROR]:\n%s\n", err.Error())
		return false
	}
	temp := strings.Replace(decode, salt, "", 1)
	temp = strings.Replace(temp, subU.UID, "", 1)
	return temp == u.UID
}

func (u *User) SignTicket(subU *User, t *Ticket) string {
	return u.SignDoc(subU.UID + t.Salt + u.UID)
}

func (u *User) TakeTicket(subU *User, t *Ticket) {
	u.TicketProofs[t.Salt] = u.SignTicket(subU, t)
}

func (u *User) SignDoc(doc string) string {
	if u.privateKey == "" {
		u.GenNewKey()
	}
	enc, err := Encrypt(doc, u.privateKey)
	if err != nil {
		log.Printf("[SignDoc ERROR]:\n%s\n", err.Error())
		return ""
	}
	return enc
}

func (u *User) GetPublicKey() string {
	return u.PublicKey
}

func (u *User) GenNewKey() string {
	if u.privateKey+u.PublicKey != "" {
		return u.PublicKey
	}
	decodekey, encodekey, err := NewKey()
	if err != nil {
		log.Printf("[New RSA KEY ERROR]:\n%s\n", err.Error())
		return ""
	}
	u.PublicKey = decodekey
	u.privateKey = encodekey
	return decodekey
}

type Ticket struct {
	Salt         string // random string
	CreateUserID string
	Passer       *Passer
}

func (t *Ticket) IsSigned(u *User) bool {
	if proof, ok := u.TicketProofs[t.Salt]; ok {
		supU := GlobalUsers.GetUserID(t.CreateUserID)
		return supU.IsMySigned(proof, t.Salt, u)
	}
	return false
}

func NewTicket(UserID string) *Ticket {
	return &Ticket{
		Salt:         utils.RandStr(20),
		CreateUserID: UserID,
		Passer:       NewPasser(),
	}
}

func NewTopTicket(UserID string) *Ticket {
	t := NewTicket(UserID)
	t.Passer.AllowMap["/"] = newCRUD(true, true, true, true)
	return t
}

func NewBanTicket(UserID string) *Ticket {
	t := NewTicket(UserID)
	t.Passer.BlackMap["/"] = newCRUD(true, true, true, true)
	return t
}

type Passer struct {
	AllowMap map[string]CRUD `json:"allow"`
	BlackMap map[string]CRUD `json:"black"`
}

func NewPasser() *Passer {
	return &Passer{
		AllowMap: make(map[string]CRUD),
		BlackMap: make(map[string]CRUD),
	}
}

func (p *Passer) MergePasser(inpass *Passer) {
	for path, auth := range inpass.AllowMap {
		if _, ok := p.AllowMap[path]; ok {
			p.AllowMap[path] = mergeCRUD(p.AllowMap[path], auth)
		}else{
			p.AllowMap[path] = auth
		}
	}
	for path, auth := range inpass.BlackMap {
		if _, ok := p.BlackMap[path]; ok {
			p.BlackMap[path] = mergeCRUD(p.BlackMap[path], auth)
		}else{
			p.BlackMap[path] = auth
		}
	}
}

func (p *Passer) IsPassedReq(r *http.Request) bool {
	UrlPth := r.URL.Path
	method := r.Method
	for reg, auth := range p.BlackMap {
		re, err := regexp.Compile(reg)
		if err != nil {
			continue
		}
		if re.MatchString(UrlPth) && auth.canDo(method) {
			return false
		}
	}
	for reg, auth := range p.AllowMap {
		re, err := regexp.Compile(reg)
		if err != nil {
			continue
		}
		if re.MatchString(UrlPth) && auth.canDo(method) {
			return true
		}
	}
	return false
}

func DefaultAuthInit() {
	// randPassword := utils.RandStr(8)
	// SystemUser.Password = randPassword
	SystemUser.GenNewKey()
	fmt.Printf("=== [Auth] middleware inited ===\n")
	// fmt.Printf("System password: %s\n", randPassword)
	fmt.Printf("System Public Key: %s\n\n", SystemUser.GetPublicKey())
}
