package auth

import (
	"fmt"
	"strconv"
)

type AuthDB interface {
	FindUser(username string) (*User, error)
	FindUserByUID(uid string) (*User, error)
	LoginUser(username string, password string) (*User, error)
	RegisteUser(u *User) error

	FindTicket(salt string) (*Ticket, error)
	FindTicketByCreateUID(uid string) ([]*Ticket, error)
	CreateTicket(uid string, pass *Passer) (*Ticket, error)
	UpdataTicket(salt string, pass *Passer) error

	QueryUserPasser(uname string) (*Passer, error)
}

type MemSimpleDB struct {
	users      map[string]*User
	usersByUID map[string]*User

	tickets      map[string]*Ticket
	ticketsByUID map[string][]*Ticket
}

func newMemSimpleDB() *MemSimpleDB {
	return &MemSimpleDB{
		users:        make(map[string]*User),
		usersByUID:   make(map[string]*User),
		tickets:      make(map[string]*Ticket),
		ticketsByUID: make(map[string][]*Ticket),
	}
}

func (m MemSimpleDB) nextUID() string {
	userlistlen := len(m.users)
	if userlistlen == 0 {
		return "0"
	}
	// check uid vaild
	for {
		if _, err := m.FindUserByUID(strconv.Itoa(userlistlen)); err == nil {
			userlistlen += 1
		} else {
			return strconv.Itoa(userlistlen)
		}
	}
}

func (m MemSimpleDB) UIDTouname(uid string) string {
	if u, ok := m.usersByUID[uid]; ok {
		return u.Name
	}
	return ""
}

func (m MemSimpleDB) unameToUID(uname string) string {
	if u, ok := m.users[uname]; ok {
		return u.UID
	}
	return ""
}

// User function
func (m MemSimpleDB) FindUser(username string) (*User, error) {
	if v, ok := m.users[username]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cant found [name: %s]user", username)
}
func (m MemSimpleDB) FindUserByUID(uid string) (*User, error) {
	if v, ok := m.usersByUID[uid]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cant found [uid: %s]user", uid)
}
func (m MemSimpleDB) LoginUser(username string, password string) (*User, error) {
	if u, err := m.FindUser(username); err != nil {
		if u.Password == password {
			return u, nil
		}
		return nil, fmt.Errorf("Wrong Password")
	} else {
		return nil, err
	}
}
func (m MemSimpleDB) RegisteUser(u *User) error {
	if _, err := m.FindUser(u.Name); err != nil {
		m.users[u.Name] = u
		if u.UID == "" {
			u.UID = m.nextUID()
		}
		m.usersByUID[u.UID] = u
	} else {
		return fmt.Errorf("Username Is Exist")
	}
	return nil
}

// Ticket Function
func (m MemSimpleDB) FindTicket(salt string) (*Ticket, error) {
	if v, ok := m.tickets[salt]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cant found [salt: %s]ticket", salt)
}
func (m MemSimpleDB) FindTicketByCreateUID(uid string) ([]*Ticket, error) {
	if v, ok := m.ticketsByUID[uid]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cant found [cuid: %s]ticket", uid)
}
func (m MemSimpleDB) CreateTicket(uid string, pass *Passer) (*Ticket, error) {
	if uid == "0" {
		// system call
		t := NewTicket(uid)
		t.Salt = "__ROOT__"
		t.Passer.MergePasser(pass)

		m.tickets[t.Salt] = t
		// m.ticketsByUID[t.CreateUserID] = t

		return t, nil
	}
	if u, err := m.FindUserByUID(uid); err == nil {
		if p, err := m.QueryUserPasser(u.Name); err == nil {
			if p.IsCover(pass) {
				t := NewTicket(uid)
				t.Passer.MergePasser(pass)

				m.tickets[t.Salt] = t
				// m.ticketsByUID[t.CreateUserID] = t

				return t, nil
			}
			return nil, fmt.Errorf("Passer to low too sign new one")
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
func (m MemSimpleDB) UpdataTicket(salt string, pass *Passer) error {
	if t, err := m.FindTicket(salt); err != nil {
		return err
	} else {
		t.Passer = pass
	}
	return nil
}

// Passer
func (m MemSimpleDB) QueryUserPasser(uname string) (*Passer, error) {
	u, err := m.FindUser(uname)
	if err != nil {
		return nil, err
	}
	p := NewPasser()
	for salt := range u.TicketProofs {
		if t, err := m.FindTicket(salt); err == nil {
			if t.IsSigned(u) {
				p.MergePasser(t.Passer)
			}
		}
	}
	return p, nil
}
