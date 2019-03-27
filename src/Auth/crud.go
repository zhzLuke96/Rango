package Auth

import "strings"

// 0b1111 CRUD
// 0b1000 post
// 0b0100 get
// 0b0010 updata
// 0b0001 delete
func canDo(auth int, method string) bool {
	if auth > 15 {
		return true
	}
	switch strings.ToLower(method) {
	case "post":
		return (auth & 1 << 3) != 0
	case "get":
		return (auth & 1 << 2) != 0
	case "updata":
		return (auth & 1 << 1) != 0
	case "delete":
		return (auth & 1) != 0
	default:
		return false
	}
}

func newAuth(C, R, U, D bool) (auth int) {
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

func mergeAuth(authA, authB int) int {
	return authA | authB
}

func toggleAuth(auth *int, C, R, U, D bool) {
	if C {
		*auth ^= 1 << 3
	}
	if R {
		*auth ^= 1 << 2
	}
	if U {
		*auth ^= 1 << 1
	}
	if D {
		*auth ^= 1
	}
}

func cancelAuth(auth *int, C, R, U, D bool) {
	if C {
		*auth &= 1 << 3
	}
	if R {
		*auth &= 1 << 2
	}
	if U {
		*auth &= 1 << 1
	}
	if D {
		*auth &= 1
	}
}
