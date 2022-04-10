package config

import (
	"gopkg.in/ini.v1"
)

// Config ini file of the whole application
var Config *ini.File

// APP
var WORKER int
var QUEUE_SIZE int
var DEBUG bool
var REPORT string
var STAGE string

// STORAGE
var REPOS_PATH string
var SCRIPTS_PATH string
var WORKFLOWS_PATH string
var BATCH_SIZE int

// Web
var GITHUB_TOKEN []string
var TRYOUT int
var TIMEOUT int
var TIMEOUT_THRESHOLD int
var SINCE_INTERVAL int
var MAX_SINCE int

// Init configurations
func Init() {
	var err error

	// load config.ini file
	Config, err = ini.ShadowLoad("config.ini")
	if err != nil {
		panic(err)
	}

	// load APP section
	APPSection, err := Config.GetSection("APP")
	if err != nil {
		panic(err)
	}
	WORKER = APPSection.Key("WORKER").MustInt(16)
	QUEUE_SIZE = APPSection.Key("QUEUE_SIZE").MustInt(128)
	DEBUG = APPSection.Key("DEBUG").MustBool(false)
	REPORT = APPSection.Key("REPORT").String()
	STAGE = APPSection.Key("STAGE").String()

	// load STORAGE section
	STORAGESection, err := Config.GetSection("STORAGE")
	if err != nil {
		panic(err)
	}
	REPOS_PATH = STORAGESection.Key("REPOS_PATH").String()
	WORKFLOWS_PATH = STORAGESection.Key("WORKFLOWS_PATH").String()
	SCRIPTS_PATH = STORAGESection.Key("SCRIPTS_PATH").String()
	BATCH_SIZE = STORAGESection.Key("BATCH_SIZE").MustInt(1024)

	// load WEB section
	WEBSection, err := Config.GetSection("WEB")
	if err != nil {
		panic(err)
	}
	GITHUB_TOKEN = WEBSection.Key("GITHUB_TOKEN").ValueWithShadows() // parse token list from config
	TRYOUT = WEBSection.Key("TRYOUT").MustInt(5)
	TIMEOUT = WEBSection.Key("TIMEOUT").MustInt(3)
	TIMEOUT_THRESHOLD = WEBSection.Key("TIMEOUT_THRESHOLD").MustInt(10)
	SINCE_INTERVAL = WEBSection.Key("SINCE_INTERVAL").MustInt(5000)
	MAX_SINCE = WEBSection.Key("MAX_SINCE").MustInt(450000000)
}

func ParseKey(section *ini.Section, key string) string {
	return section.Key(key).String()
}
