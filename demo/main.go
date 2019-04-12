package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	Rango "../src"
	"../src/auth"
	rangoKit "../src/core"
	"../src/mid"
	RangoRouter "../src/router"
)

func main() {
	var demoNum int
	var port string
	flag.IntVar(&demoNum, "demo", 2, "choose demo.")
	flag.IntVar(&demoNum, "d", 2, "choose demo.")

	flag.StringVar(&port, "port", "8080", "server port.")
	flag.StringVar(&port, "p", "8080", "server port.")

	flag.Parse()

	switch demoNum {
	case 0:
		DefaultSev(port)
	case 1:
		CustomSev(port)
	case 2:
		TokenAuthSev(port)
	default:
		fmt.Println("Wrong Demo Num.")
		flag.Usage()
	}
	return
}

func clearCookie(w rangoKit.ResponseWriteBody, r *http.Request) {
	c := new(http.Cookie)
	c.Name = "RangoToken"
	c.Expires = time.Now().AddDate(-1, 0, 0)
	http.SetCookie(w, c)
	oc, _ := r.Cookie("RangoToken")
	oc.Expires = time.Now().AddDate(-1, 0, 0)
}

func DefaultSev(port string) {
	Rango.SimpleGo(port)
}

func CustomSev(port string) {
	sev := Rango.NewSev()
	router := Rango.NewRouter()
	// session := Rango.Session()
	sev.Use(mid.LogRequest)
	sev.Use(mid.ErrCatch)
	// sev.Use(session.mid)
	sev.Use(router.Mid)
	router.Handler("/home", http.StripPrefix("/home", http.FileServer(http.Dir("./www"))))
	// sev.Use(mid.Log, mid.ErrCAtch, session.mid, router.mid)

	sev.Go(port)
}

func TokenAuthSev(port string) {
	sev := Rango.NewSev()
	router := Rango.NewRouter()
	CMSHandler := Rango.NewSev()

	sev.Use(mid.LogRequest)
	sev.Use(mid.ErrCatch)
	sev.Use(router.Mid)

	router.Registe(map[string]http.Handler{
		"/CMS/":     CMSHandler,
		"/login/":   rangoKit.HandlerFunc(auth.GlobalManager.LoginHandler),
		"/registr/": rangoKit.HandlerFunc(auth.GlobalManager.RegisteHandler),
		"/ticket/":  rangoKit.HandlerFunc(auth.GlobalManager.TicketHandler),
		"/clear/":   rangoKit.HandlerFunc(clearCookie),
		"/sysuser/": rangoKit.HandlerFunc(func(w rangoKit.ResponseWriteBody, r *http.Request) {
			body, _ := json.Marshal(auth.GlobalManager.SystemUser())
			w.Write(body)
		}),
		"/router/{id:[0-9]+}/{name}": rangoKit.HandlerFunc(func(w rangoKit.ResponseWriteBody, r *http.Request) {
			body, _ := json.Marshal(RangoRouter.Vars(r))
			w.Write(body)
		}),
	})
	router.Handler("/", http.FileServer(http.Dir("./www")))

	subRouter := Rango.NewRouter()
	CMSHandler.Use(mid.StripPrefix("/CMS"))
	CMSHandler.Use(auth.GlobalManager.Mid)
	CMSHandler.Use(subRouter.Mid)

	subRouter.Registe(map[string]http.Handler{
		"/": http.FileServer(http.Dir("./zone")),
	})

	auth.DefaultAuthInit()
	auth.GlobalManager.DB.RegisteUser(auth.NewUser("user1", "", "123456"))

	sev.Go(port)
}
