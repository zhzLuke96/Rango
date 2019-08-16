package auth

import (
	"net/http"
	"regexp"
)

type Passer struct {
	AllowMap map[string]CRUD `json:"allow"`
	BlackMap map[string]CRUD `json:"black"`
}

func NewPasser() *Passer {
	return &Passer{
		AllowMap: make(map[string]CRUD),
		BlackMap: make(map[string]CRUD),
	}
}

func (p *Passer) IsCover(pass *Passer) bool {
	for path, auth := range pass.AllowMap {
		if !p.IsPassed(path, auth) {
			return false
		}
	}
	for path, auth := range pass.BlackMap {
		if !p.IsPassed(path, auth) {
			return false
		}
	}
	return false
}

func (p *Passer) MergePasser(inpass *Passer) {
	for path, auth := range inpass.AllowMap {
		if _, ok := p.AllowMap[path]; ok {
			p.AllowMap[path] = mergeCRUD(p.AllowMap[path], auth)
		} else {
			p.AllowMap[path] = auth
		}
	}
	for path, auth := range inpass.BlackMap {
		if _, ok := p.BlackMap[path]; ok {
			p.BlackMap[path] = mergeCRUD(p.BlackMap[path], auth)
		} else {
			p.BlackMap[path] = auth
		}
	}
}

func (p *Passer) IsPassed(path string, c CRUD) bool {
	for reg, auth := range p.BlackMap {
		re, err := regexp.Compile(reg)
		if err != nil {
			continue
		}
		if re.MatchString(path) && auth.CanCover(c) {
			return false
		}
	}
	for reg, auth := range p.AllowMap {
		re, err := regexp.Compile(reg)
		if err != nil {
			continue
		}
		if re.MatchString(path) && auth.CanCover(c) {
			return true
		}
	}
	return false
}

func (p *Passer) IsPassedMethod(path string, method string) bool {
	return p.IsPassed(path, methodToCRUD(method))
}

func (p *Passer) IsPassedReq(r *http.Request) bool {
	return p.IsPassedMethod(r.URL.Path, r.Method)
}
