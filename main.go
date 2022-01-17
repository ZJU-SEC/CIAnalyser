package main

import (
	"CIHunter/src/config"
	"CIHunter/src/database"
	"CIHunter/src/web"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize database
	database.Init()

	crawler.Crawl()
}
