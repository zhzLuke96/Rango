package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"../core"
)

func (a *manager) LoginHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var user struct {
		Name      string `json:"name"`
		Pass      string `json:"pass"`
		AutoLogin bool   `json:"auto"`
	}
	if err := json.Unmarshal(body, &user); err == nil {
		if _, err := a.DB.LoginUser(user.Name, user.Pass); err == nil {
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

func (a *manager) RegisteHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	user := new(User)
	if err := json.Unmarshal(body, user); err == nil {
		if err := a.DB.RegisteUser(user); err == nil {
			resp.Status = "SUCCESS"
			resp.Msg = "Registe Success"
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

func (a *manager) TicketHandler(w core.ResponseWriteBody, r *http.Request) {
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

func (a *manager) NewTicketHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var ticketRequest struct {
		Passer Passer `json:"pass"`
	}
	if err := json.Unmarshal(body, &ticketRequest); err == nil {
		if callerUser := GetUserFromReq(r); callerUser != nil {
			t, err := a.DB.CreateTicket(callerUser.UID, &ticketRequest.Passer)
			if err != nil {
				resp.Status = "ERROR"
				resp.Msg = "Unauthorized"
			} else {
				resp.Status = "SUCCESS"
				resp.Msg = "Create Success"
				resp.Data = make(map[string]string)
				resp.Data["SALT"] = t.Salt
			}
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

func (a *manager) SignTicketHandler(w core.ResponseWriteBody, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	resp := new(ResponesAuth)
	var ticketRequest struct {
		TicketName  string `json:"name"`
		TargetUName string `json:"target"`
	}
	for {
		if err := json.Unmarshal(body, &ticketRequest); err == nil {
			if callerUser := GetUserFromReq(r); callerUser != nil {
				setterUser, err := a.DB.FindUser(ticketRequest.TargetUName)
				if err != nil {
					resp.Status = "ERROR"
					resp.Msg = err.Error()
					break
				}
				tik, err := a.DB.FindTicket(ticketRequest.TicketName)
				if err != nil {
					resp.Status = "ERROR"
					resp.Msg = err.Error()
					break
				}
				doc := callerUser.SignTicket(setterUser, tik)
				setterUser.TicketProofs[ticketRequest.TicketName] = doc
			} else {
				resp.Status = "ERROR"
				resp.Msg = "Unauthorized"
				break
			}
		} else {
			resp.Status = "ERROR"
			resp.Msg = err.Error()
			break
		}
	}
	w.Write(resp.JSON())
}
