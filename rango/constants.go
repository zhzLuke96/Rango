package rango

const (
	// 版本号
	Version = "0.08h"

	debugKey         = "**DEBUG**"
	forceLoadKey     = "forceLoad"
	simpleServerName = "simple file server"
	systemError      = "Error occurred in the system. Please repeat it later."

	responseIdxMAX = 1000

	// Deprecated
	// ----------
	// SissionMid used
	sessionCookieName = "_sid_"
)

var (
	configFoundList = []string{"./config.json", "~/config.json", "/config.json"}

	uploadResponser    = NewResponser("uploadServer")
	rHFuncResponser    = NewResponser("rangoHandlerFunction")
	errCatchResponser  = NewResponser("ErrCatchMid")
	notFoundResponser  = NewResponser("notFound")
	mainFileResponser  = NewResponser("mainFile")
	mainBytesResponser = NewResponser("mainBytes")
	// hateoasResponser = NewResponser("HATEOAS")
)
