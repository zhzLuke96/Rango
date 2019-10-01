package rango

import "fmt"

var rSYSCONF = make(map[string]interface{})

func SetConf(key string, value interface{}) {
	rSYSCONF[key] = value
}

func GetConf(key string, defaultValue interface{}) interface{} {
	if v, ok := rSYSCONF[key]; ok {
		return v
	}
	return defaultValue
}

func isDebugOn() bool {
	return GetConf(debugKey, false).(bool)
}

func DebugOn() {
	SetConf(debugKey, true)
}

func DebugOff() {
	SetConf(debugKey, false)
}

func loadConfig(config map[string]interface{}) {
	for k, v := range config {
		if _, ok := rSYSCONF[k]; !ok {
			rSYSCONF[k] = v
		}
	}
}

func loadConfigForce(config map[string]interface{}) {
	for k, v := range config {
		rSYSCONF[k] = v
	}
}

// read config.json
func ReadConfig() {
	ReadConfigFile(configFoundList...)
}

// read config.json
func ReadConfigFile(pths ...string) {
	config, pth := mustReadJSONFile(pths...)
	if config == nil {
		fmt.Println("Cant Found config json file.")
		return
	}
	if _, ok := config[forceLoadKey].(bool); ok {
		fmt.Printf("force load config file in [%s]\n", pth)
		loadConfigForce(config)
	} else {
		fmt.Printf("load config file in [%s]\n", pth)
		loadConfig(config)
	}
}
