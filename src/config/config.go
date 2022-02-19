package config

import (
	"gopkg.in/ini.v1"
)

// Config ini file of the whole application
var Config *ini.File

// APP
var WORKER int
var QUEUE_SIZE int
var TRYOUT int
var TIMEOUT int
var TIMEOUT_THRESHOLD int
var DEBUG bool
var BATCH_SIZE int

// STORAGE
var REPOS_PATH string
var WORKFLOWS_PATH string

// Web
var GITHUB_TOKEN string

// Init configurations
func Init() {
	var err error

	// load config.ini file
	Config, err = ini.Load("config.ini")
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
	TRYOUT = APPSection.Key("TRYOUT").MustInt(5)
	TIMEOUT = APPSection.Key("TIMEOUT").MustInt(3)
	TIMEOUT_THRESHOLD = APPSection.Key("TIMEOUT_THRESHOLD").MustInt(10)
	DEBUG = APPSection.Key("DEBUG").MustBool(false)
	BATCH_SIZE = APPSection.Key("BATCH_SIZE").MustInt(1024)

	STORAGESection, err := Config.GetSection("STORAGE")
	if err != nil {
		panic(err)
	}
	REPOS_PATH = STORAGESection.Key("REPOS_PATH").String()
	WORKFLOWS_PATH = STORAGESection.Key("WORKFLOWS_PATH").String()

	// load WEB section
	WEBSection, err := Config.GetSection("WEB")
	if err != nil {
		panic(err)
	}
	GITHUB_TOKEN = ParseKey(WEBSection, "GITHUB_TOKEN")
}

func ParseKey(section *ini.Section, key string) string {
	return section.Key(key).String()
}
