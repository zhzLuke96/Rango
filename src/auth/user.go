package auth

import (
	"log"
	"strings"
)

type User struct {
	Name     string `json:"name"`
	UID      string `json:"uid"`
	Password string `json:"pass"`

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

func (u *User) TakeTicket(subU *User, t *Ticket) {
	u.TicketProofs[t.Salt] = u.SignTicket(subU, t)
}

func (u *User) SignTicket(subU *User, t *Ticket) string {
	return u.SignDoc(subU.UID + t.Salt + u.UID)
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
