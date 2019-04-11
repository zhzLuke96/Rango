package auth

import "../utils"

type Ticket struct {
	Salt         string // random string
	CreateUserID string
	Passer       *Passer
}

func (t *Ticket) IsSigned(u *User) bool {
	if proof, ok := u.TicketProofs[t.Salt]; ok {
		supU, err := GlobalManager.DB.FindUserByUID(t.CreateUserID)
		if err != nil {
			return false
		}
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
	t.Passer.AllowMap["/"] = CRUD(15) // newCRUD(true, true, true, true)
	return t
}

func NewBanTicket(UserID string) *Ticket {
	t := NewTicket(UserID)
	t.Passer.BlackMap["/"] = CRUD(15) // newCRUD(true, true, true, true)
	return t
}
