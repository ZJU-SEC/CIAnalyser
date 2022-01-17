package config

import "gopkg.in/ini.v1"

// Config ini file of the whole application
var Config *ini.File

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
