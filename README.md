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

# todo
- router VarN
- auth DB factory
- sessionVar

# LICENSE
GPL-3.0