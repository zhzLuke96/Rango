package rango

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type DictArr struct {
	arr     []map[string]interface{}
	lessKey string
}

func NewDictArr(a []map[string]interface{}) *DictArr {
	return &DictArr{
		arr:     a,
		lessKey: "idx",
	}
}

func (a *DictArr) Len() int {
	return len(a.arr)
}

func (a *DictArr) Less(i, j int) bool {
	if a.lessKey == "" {
		return true
	}
	iV := a.arr[i][a.lessKey]
	jV := a.arr[j][a.lessKey]
	return anyLess(iV, jV)
}

func (a *DictArr) Swap(i, j int) {
	a.arr[i], a.arr[j] = a.arr[j], a.arr[i]
}

func (a *DictArr) Sort(key string) *DictArr {
	a.lessKey = key
	sort.Sort(a)
	return a
}

type Param map[string]string

func (p Param) Pass(check map[string]interface{}) bool {
	for pk, pv := range p {
		keyFound := false
		for k, v := range check {
			if pk == k {
				if !queryEqule(pv, v) {
					return false
				} else {
					continue
				}
			}
			if len(pk)-len(k) == 1 && strings.HasPrefix(pk, k) {
				keyFound = true
				if !conditionPass(pk[len(pk)-1:], pv, v) {
					return false
				}
			}
		}
		if !keyFound {
			return false
		}
	}
	return true
}

func conditionPass(method string, pv string, v interface{}) bool {
	pvnum, pverr := strconv.ParseFloat(pv, 64)

	vstr, vstrok := v.(string)
	vnum, verr := toFloat(v)
	if !vstrok && verr != nil {
		return false
	}
	switch method {
	case "<":
		if pverr != nil || verr != nil {
			return false
		}
		return vnum <= pvnum
	case ">":
		if pverr != nil || verr != nil {
			return false
		}
		return vnum >= pvnum
	case "!":
		if pverr != nil || verr != nil {
			return vstr != pv
		}
		return vnum != pvnum
	case "^":
		return strings.HasPrefix(vstr, pv)
	case "$":
		return strings.HasSuffix(vstr, pv)
	case "*":
		return strings.Index(vstr, pv) != -1
	case "@", "%", ":", ";", "|", ".", ",":
		return false
	default:
		return false
	}
}

type crudify interface {
	// pager
	// Mate(cname string) (mate map[string]interface{},err error)
	Mate(string) (map[string]interface{}, error)

	// Create
	// Insert(cname string, e map[string]interface{}) (idx int, err error)
	Insert(string, map[string]interface{}) (int, error)

	// Updata
	// Update(canme string, idx int, e map[string]interface{}) error
	Update(string, int, map[string]interface{}) error

	// Retrieve
	// Query(cname string, idx int) (e map[string]interface{}, err error)
	Query(string, int) (map[string]interface{}, error)
	// QueryMap(cname string, param map[string]interface{}, page, n int) (es []map[string]interface{}, err error)
	QueryMap(string, *Param, int, int) ([]map[string]interface{}, error)

	// Delete
	// Remove(cname string, idx int) error
	Remove(string, int) error
	// RemoveMap(cname string, param map[string]interface{}, n int) error
	RemoveMap(string, *Param, int) error

	// Auth
	AuthHandler(*http.Request, string, string) (string, error)
}

func curdErrResp(code int, msg string, err error) *errResponse {
	resp := curdResponser.NewErrResponse()
	return resp.Set(code, msg, err)
}

type rCRUD struct {
	crudEntry crudify
	sev       *Server
}

func (c rCRUD) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.sev.Handler.ServeHTTP(w, r)
}

func newCRUD(c crudify, preFix string) *rCRUD {
	sev := NewSev("__crud__")
	// query list
	sev.GET("/{entity:\\w+}s/", func(vars *ReqVars) interface{} {
		query := Param(excluding(vars.Query(), "page", "n"))
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), readRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		count := vars.GetDefault("n", "-1").(string)
		countNum, err := strconv.Atoi(count)
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		page := vars.GetDefault("page", "0").(string)
		pageNum, err := strconv.Atoi(page)
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		res, err := c.QueryMap(cname.(string), &query, pageNum, countNum)
		if err != nil {
			return curdErrResp(404, "cant find entity.", err)
		}
		if res == nil {
			return []string{}
		}
		return res
	})
	// query one index
	sev.GET("/{entity:\\w+}/{idx:\\d+}/", func(vars *ReqVars) interface{} {
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), readRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		idx, err := vars.Get("idx")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		idxNum, err := strconv.Atoi(idx.(string))
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		res, err := c.Query(cname.(string), idxNum)
		if err != nil {
			return curdErrResp(404, "cant find entity.", err)
		}
		return res
	})
	// query one map
	sev.GET("/{entity:\\w+}/", func(vars *ReqVars) interface{} {
		query := Param(vars.Query())
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), readRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		if len(query) == 0 {
			return nil
		}
		res, err := c.QueryMap(cname.(string), &query, 0, 1)
		if err != nil {
			return curdErrResp(404, "cant find entity.", err)
		}
		if len(res) == 0 {
			return nil
		}
		return res[0]
	})
	// insert
	sev.POST("/{entity:\\w+}s/", func(vars *ReqVars) interface{} {
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), createRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		idx, err := c.Insert(cname.(string), vars.JSON())
		if err != nil {
			return curdErrResp(404, "cant insert entity.", err)
		}
		return map[string]interface{}{
			"idx": idx,
		}
	})
	// updata
	sev.Func("/{entity:\\w+}/{idx:\\d+}/", func(vars *ReqVars) interface{} {
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), updateRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		idx, err := vars.Get("idx")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		idxNum, err := strconv.Atoi(idx.(string))
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		err = c.Update(cname.(string), idxNum, vars.JSON())
		if err != nil {
			return curdErrResp(404, "cant updata entity.", err)
		}
		return "success"
	}).Methods("PUT")
	// delete
	sev.Func("/{entity:\\w+}/{idx:\\d+}/", func(vars *ReqVars) interface{} {
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), deleteRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		idx, err := vars.Get("idx")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		idxNum, err := strconv.Atoi(idx.(string))
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		err = c.Remove(cname.(string), idxNum)
		if err != nil {
			return curdErrResp(404, "cant updata entity.", err)
		}
		return "success"
	}).Methods("DELETE")
	// delete map
	sev.Func("/{entity:\\w+}s/", func(vars *ReqVars) interface{} {
		query := Param(excluding(vars.Query(), "n"))
		cname, err := vars.Get("entity")
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		if msg, err := c.AuthHandler(vars.Request(), deleteRequestKey, cname.(string)); err != nil {
			return curdErrResp(401, msg, err)
		}
		count := vars.GetDefault("n", "-1").(string)
		countNum, err := strconv.Atoi(count)
		if err != nil {
			return curdErrResp(500, "server is wrong.", err)
		}
		err = c.RemoveMap(cname.(string), &query, countNum)
		if err != nil {
			return curdErrResp(404, "cant remove entity.", err)
		}
		return "success"
	}).Methods("DELETE")

	sev.Handler.Use(StripPrefixMid(preFix))
	sev.Handler.Use(sev.Router.Mid)
	return &rCRUD{
		crudEntry: c,
		sev:       sev,
	}
}
