package Auth

import "strings"

type CRUD int8

// 0b1111 CRUD
// 0b1000 post
// 0b0100 get
// 0b0010 updata
// 0b0001 delete
func (c *CRUD) canDo(method string) bool {
	if *c > 15 {
		return true
	}
	switch strings.ToLower(method) {
	case "post":
		return (*c & 1 << 3) != 0
	case "get":
		return (*c & 1 << 2) != 0
	case "updata":
		return (*c & 1 << 1) != 0
	case "delete":
		return (*c & 1) != 0
	default:
		return false
	}
}

func newCRUD(C, R, U, D bool) (auth CRUD) {
	if C {
		auth |= 1 << 3
	}
	if R {
		auth |= 1 << 2
	}
	if U {
		auth |= 1 << 1
	}
	if D {
		auth |= 1
	}
	return auth
}

func mergeCRUD(authA, authB CRUD) CRUD {
	return authA | authB
}

func (c *CRUD) Toggle(C, R, U, D bool) {
	if C {
		*c ^= 1 << 3
	}
	if R {
		*c ^= 1 << 2
	}
	if U {
		*c ^= 1 << 1
	}
	if D {
		*c ^= 1
	}
}

func (c *CRUD) Cancel(C, R, U, D bool) {
	if C {
		*c &= 1 << 3
	}
	if R {
		*c &= 1 << 2
	}
	if U {
		*c &= 1 << 1
	}
	if D {
		*c &= 1
	}
}
