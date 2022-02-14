package config

import (
	"gopkg.in/ini.v1"
	"time"
)

// Config ini file of the whole application
var Config *ini.File

// APP
var UPDATE_DIFF time.Duration
var NOW = time.Now()
var DEV_SHM string
var THREAD_SIZE int

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
	updateDays := APPSection.Key("UPDATE_DIFF").MustInt(7)
	UPDATE_DIFF = time.Duration(24*updateDays) * time.Hour
	DEV_SHM = APPSection.Key("DEV_SHM").String()
	THREAD_SIZE = APPSection.Key("THREAD_SIZE").MustInt(16)

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
