# Rango
http web kit

# exmple
```golang
func CustomSev(port string) {
	sev := Rango.NewSev()
    router := Rango.NewRouter()
    
	sev.Use(mid.LogRequest)
    sev.Use(mid.ErrCatch)
    sev.Use(router.Mid)
    
    router.Handler("/home/",
        http.StripPrefix("/home/",
            http.FileServer(
                    http.Dir("./www")
                    )
                )
            ).Methods("get")

	sev.Go(port)
}
```

Router Interceptor or filter
```golang
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
```

# todo
- [ ] router VarN
- [x] auth DB factory
- [ ] sessionVar
- [ ] more demo
- [ ] CLI
- [ ] ...

# LICENSE
GPL-3.0