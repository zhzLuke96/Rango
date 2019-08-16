package auth

import (
	"encoding/base64"
	"log"
	"net/http"

	"../core"
	"../utils"
)

type manager struct {
	DB AuthDB
}

func NewManager(DB AuthDB) *manager {
	return &manager{
		DB: DB,
	}
}

func (m *manager) InitDB() string {
	// push system user
	randomPassword := utils.RandStr(8)
	sysUser := NewUser("system", "0", randomPassword)
	// push system ticket
	p := NewPasser()
	p.AllowMap["/"] = CRUD(15) // newCRUD(true, true, true, true)
	t, _ := m.DB.CreateTicket("0", p)
	sysUser.TakeTicket(sysUser, t)
	// RSA KEY
	sysUser.GenNewKey()

	m.DB.RegisteUser(sysUser)
	return randomPassword
}

func (m *manager) SoftInitDB() string {
	if u, err := m.DB.FindUserByUID("0"); err == nil {
		if p, err := m.DB.QueryUserPasser(u.Name); err == nil {
			if p.AllowMap["/"] >= 15 {
				return u.Password
			}
		}
	}
	return m.InitDB()
}

func (m *manager) SystemUser() *User {
	if u, err := m.DB.FindUser("system"); err == nil {
		return u
	}
	if u, err := m.DB.FindUserByUID("0"); err == nil {
		return u
	}
	return nil
}

func (m *manager) IsPassed(r *http.Request) bool {
	if t, ok := GetTokenFromReq(r); ok {
		cipher, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			log.Printf("[Auth ERROR]:%s\n", err.Error())
			return false
		}
		json, err := Decrypt(string(cipher), m.SystemUser().PublicKey)
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

func (a *manager) Mid(w core.ResponseWriteBody, r *http.Request, next func()) {
	if a.IsPassed(r) {
		next()
	} else {
		w.WriteHeader(403)
		w.Write([]byte("Unauthorized."))
	}
}
