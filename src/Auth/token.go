package Auth

import (
	"encoding/json"
	"net/http"
	"time"

	"../utils"
)

type Token struct {
	Expires        time.Time `json:"exp"`
	IssuedAt       time.Time `json:"iss"`
	TargetUserName string    `json:"U"`
	Salt           string    `json:"salt"`
	Passd          Passer    `json:"pass"`
}

func NewToken(dur time.Duration, uname string) *Token {
	return &Token{
		Expires:        time.Now().Add(dur),
		IssuedAt:       time.Now(),
		TargetUserName: uname,
		Salt:           utils.RandStr(10),
	}
}

func NewTokenFromJSON(text string) (*Token, error) {
	var T Token
	err := json.Unmarshal([]byte(text), T)
	if err != nil {
		return nil, err
	}
	return &T, nil
}

func (t *Token) IsAllowReq(r *http.Request) bool {
	return t.Passd.IsPassedReq(r)
}

func (t *Token) JSON() []byte {
	ret, _ := json.Marshal(t)
	return ret
}
