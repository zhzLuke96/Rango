package main

import (
	"net/http"
	"strconv"
	"time"

	"../rango"
	// "../rango/auth"
)

func main() {
	rango.DebugOn()
	run("8080")
}

func toInt(i interface{}) int {
	if v, ok := i.(int); ok {
		return v
	}
	if v, ok := i.(string); ok {
		if ret, err := strconv.Atoi(v); err == nil {
			return ret
		}
	}
	if v, ok := i.(float32); ok {
		return int(v)
	}
	return 0
}

type throttleMatcher struct {
	lastCall time.Time
	timeout  float64
}

func newThrottleMatcher(timeout float64) *throttleMatcher {
	return &throttleMatcher{
		// lastCall: time.Now(),
		timeout: timeout,
	}
}

func (t *throttleMatcher) Match(r *http.Request) bool {
	now := time.Now()
	if now.Sub(t.lastCall).Seconds()*1000 > t.timeout {
		t.lastCall = now
		return true
	}
	return false
}

func run(port string) {
	sev := rango.New("demo")
	sev.Func("/add", func(vars rango.ReqVars) []byte {
		numa := vars.GetDefault("a", 1)
		numb := vars.GetDefault("b", 1)
		res := toInt(numb) + toInt(numa)
		// return []byte(fmt.Sprintf("<h1>%v+%v=%v</h1>", numa, numb, res))
		return rango.NewResponse(res).JSON()
	}).AddMatcher(newThrottleMatcher(5000)).Before(func(w http.ResponseWriter, r *http.Request) bool {
		w.Write([]byte("this tool closed."))
		w.WriteHeader(500)
		return false
	})

	apiGroup := sev.Group("/api")
	apiGroup.Func("/user/{id:\\d+}", func(vars rango.ReqVars) []byte {
		userID := rango.GetConf("userPrefix", "").(string) + "_" + vars.GetDefault("id", "null").(string)
		return rango.NewResponse(userID).JSON()
	})

	// HATEOAS
	hateoas := rango.NewRHateoas("**HATEOAS**", "In progress")
	hateoas.Add("add tool", "/add?a={a}&b={b}", "add number a and number b", []string{"GET", "POST"})
	apiGroup.Handle("/", hateoas)
	apiGroup.File("", "./api_README.md")

	sev.Static("/", "./www")
	sev.Start(port)
}

// func run(port string) {
// 	sev := rango.NewSev()
// 	router := rango.NewRouter()
// 	hateoas := rango.NewRHateoas("**HATEOAS**", "In progress")

// 	sev.Use(rango.LogRequestMid)
// 	sev.Use(rango.ErrCatchMid)
// 	sev.Use(router.Mid)

// 	router.Handler("/home", http.StripPrefix("/home", http.FileServer(http.Dir("./www"))))

// router.RangoFunc("/add/{a:\\d+}:{b:\\d+}", rango.Handler(func(vars rango.ReqVars) []byte {
// 	numa := vars.GetDefault("a", 1)
// 	numb := vars.GetDefault("b", 1)
// 	res := toInt(numb) + toInt(numa)
// 	// return []byte(fmt.Sprintf("<h1>%v+%v=%v</h1>", numa, numb, res))
// 	return rango.NewResponse(res).JSON()
// })).AddMatcher(newThrottleMatcher(5000))

// 	hateoas.AppendURL("/add/{a}:{b}", "ALL", "add number a and number b")
// 	router.RangoFunc("/apis", hateoas.HandleFunc)

// 	router.RangoFunc("/error", rango.Handler(func(_ rango.ReqVars) []byte {
// 		return rango.NewErrResp(404001, "error some", "worng.").JSON()
// 	}))

// 	sev.Go(port)
// }
