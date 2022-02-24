package main

import (
	"CIHunter/src/analyzer"
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
	switch config.STAGE {
	case 1:
		web.CrawlGHAPI()
	case 2:
		web.CrawlActions()
	case 3:
		analyzer.Analyze()
	}
}
