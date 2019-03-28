package Auth

import "fmt"

var (
	GlobalUsers = NewMemUsersDB()
)

type memUsersDB struct {
	users map[string]*User
}

func NewMemUsersDB() *memUsersDB {
	return &memUsersDB{
		users: make(map[string]*User),
	}
}

func (m *memUsersDB) length() int {
	return len(m.users)
}

func (m *memUsersDB) AddOne(Name, Password string) {
	m.users[Name] = &User{
		Name:     Name,
		UID:      fmt.Sprint(m.length()),
		password: Password,
		Tickets:  make(map[string]string),
	}
}

func (m *memUsersDB) AddAdmin(Name, Password string) {
	admin := &User{
		Name:     Name,
		UID:      fmt.Sprint(m.length()),
		password: Password,
		Tickets:  make(map[string]string),
	}
	GlobalAuthManager.Push("admin", NewTopTicket(admin.UID))

	m.users[Name] = admin
}
