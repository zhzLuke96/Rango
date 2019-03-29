package Auth

import "fmt"

var (
	GlobalUsers = NewMemUsersDB()
	SystemUser  = GlobalUsers.AddAdmin("system", "admin")
)

type memUsersDB struct {
	users map[string]*User
}

func NewMemUsersDB() *memUsersDB {
	return &memUsersDB{
		users: make(map[string]*User),
	}
}

func (m *memUsersDB) Login(name, pass string) (bool, error) {
	if u, ok := m.users[name]; ok {
		if u.Password == pass {
			return true, nil
		} else {
			return false, fmt.Errorf("Wrong Password")
		}
	} else {
		return false, fmt.Errorf("User Name Not Exist")
	}
}

func (m *memUsersDB) GetUser(name string) *User {
	return m.users[name]
}

func (m *memUsersDB) length() int {
	return len(m.users)
}

func (m *memUsersDB) AddOne(Name, Password string) *User {
	u := &User{
		Name:         Name,
		UID:          fmt.Sprint(m.length()),
		Password:     Password,
		TicketProofs: make(map[string]string),
	}
	m.users[Name] = u
	return u
}

func (m *memUsersDB) AddAdmin(Name, Password string) *User {
	admin := &User{
		Name:         Name,
		UID:          fmt.Sprint(m.length()),
		Password:     Password,
		TicketProofs: make(map[string]string),
	}
	GlobalAuthManager.Push(NewTopTicket(admin.UID))

	m.users[Name] = admin
	return admin
}
