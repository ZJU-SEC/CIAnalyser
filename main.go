package main

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/web"
	//"os"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	models.Init()

	// crawl gitstar-ranking.com
	switch config.STAGE {
	case 1:
		web.CrawlGHAPI()
	case 2:
		web.CrawlActions()
	}
}
