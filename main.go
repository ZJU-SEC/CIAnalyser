package main

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/web"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	models.Init()

	// crawl gitstar-ranking.com
	web.CrawlActions()
}
