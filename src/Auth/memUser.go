package Auth

import "fmt"

var (
	GlobalUsers = NewMemUsersDB()
	SystemUser  = GlobalUsers.AddAdmin("system", "admin")
)

type memUsersDB struct {
	users   map[string]*User
	usersID map[string]*User
}

func NewMemUsersDB() *memUsersDB {
	return &memUsersDB{
		users:   make(map[string]*User),
		usersID: make(map[string]*User),
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

func (m *memUsersDB) NameCanUse(name string) bool {
	if _, ok := m.users[name]; ok {
		return false
	}
	return true
}

func (m *memUsersDB) GetUser(name string) *User {
	return m.users[name]
}

func (m *memUsersDB) GetUserID(id string) *User {
	return m.usersID[id]
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
	m.usersID[u.UID] = u
	return u
}

func (m *memUsersDB) AddAdmin(Name, Password string) *User {
	admin := &User{
		Name:         Name,
		UID:          fmt.Sprint(m.length()),
		Password:     Password,
		TicketProofs: make(map[string]string),
	}
	t := NewTopTicket(admin.UID)
	GlobalAuthManager.Push(t)
	admin.TakeTicket(admin, t)

	m.users[Name] = admin
	m.usersID[admin.UID] = admin
	return admin
}
