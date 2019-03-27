package main

import Rango "../src"

func main(){
	return
}


func useDefaultSev(port string){
	Rango.SimpleGo(port)
	Rango.Router(&{
		//...
	})
}

func useCustomSev(port string){
	sev := Rango.NewSev()
	router := Rango.NewRouter()
	session := Rango.Session()
	// sev.Use(Rango.Mid.Log)
	// sev.Use(Rango.Mid.ErrCAtch)
	// sev.Use(session.Mid)
	// sev.Use(router.Mid)
	sev.Use(Rango.Mid.Log,Rango.Mid.ErrCAtch,session.Mid,router.Mid)

	sev.Go(port)
}

func useTokenAuthSev(port string){
	sev := Rango.NewSev()
	auth := Rango.NewAuthManager()
	router := Rango.NewRouter()

	// auth.AddUser(0,"admin","admin")
	// auth.AddUser(1,"user1","123456")
	// auth.AddAdmin("admin")

	sev.Use(Rango.Mid.Log)
	sev.Use(Rango.Mid.ErrCAtch)
	sev.Use(router.Mid)

	Rango.Go(sev,port)
}