package rango

import (
	"encoding/json"
	"net/http"
)

type rLink struct {
	Describe string   `json:"desc"`
	Methods  []string `json:"methods"`
	URL      string   `json:"url"`
}

type rHateoas struct {
	links map[string]rLink

	Message string
	Status  string
}

func NewRHateoas(msg, status string) *rHateoas {
	return &rHateoas{Message: msg, Status: status, links: make(map[string]rLink)}
}

func (r *rHateoas) Add(name, url, desc string, methods []string) {
	r.links[name] = rLink{
		Describe: desc,
		URL:      url,
		Methods:  methods,
	}
}

func (r *rHateoas) JSON() ([]byte, error) {
	return json.MarshalIndent(r.links, "", "    ")
}

func (r *rHateoas) MustJSON() []byte {
	if data, err := r.JSON(); err == nil {
		return data
	}
	return []byte("{}")
}

func (r *rHateoas) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// resp := NewResponse(r.links)
	// resp.Code = 1
	// resp.Message = r.Message
	// resp.Status = r.Status
	w.Header().Set("Content-Type", "application/json")
	w.Write(r.MustJSON())
	// w.WriteHeader(200)
}
