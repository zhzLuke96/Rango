package main

import (
	"flag"
	"net/http"

	Rango "../src"
	"../src/Auth"
	mid "../src/Mid"
)

func main() {
	var demoNum int
	flag.IntVar(&demoNum, "demo", 0, "choose demo.")
	flag.IntVar(&demoNum, "d", 0, "choose demo.")
	flag.Parse()

	switch demoNum {
	case 0:
		useDefaultSev("8080")
	case 1:
		useCustomSev("8080")
	case 2:
		useTokenAuthSev("8080")
	}
	return
}

func useDefaultSev(port string) {
	Rango.SimpleGo(port)
}

func useCustomSev(port string) {
	sev := Rango.NewSev()
	router := Rango.NewRouter()
	// session := Rango.Session()
	sev.Use(mid.LogRequest)
	sev.Use(mid.ErrCatch)
	// sev.Use(session.mid)
	sev.Use(router.Mid)
	router.Handler("/home/", http.StripPrefix("/home/", http.FileServer(http.Dir("./www"))))
	// sev.Use(mid.Log, mid.ErrCAtch, session.mid, router.mid)

	sev.Go(port)
}

func useTokenAuthSev(port string) {
	sev := Rango.NewSev()
	router := Rango.NewRouter()

	sev.Use(mid.LogRequest)
	sev.Use(mid.ErrCatch)
	sev.Use(Auth.GlobalAuthManager.Mid)
	sev.Use(router.Mid)

	Auth.GlobalUsers.AddOne("user1", "123456")

	sev.Go(port)
}
