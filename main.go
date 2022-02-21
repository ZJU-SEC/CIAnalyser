package main

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"CIHunter/src/web"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	models.Init()

	utils.Init()

	// crawl gitstar-ranking.com
	//res := web.CrawlActions()
	//os.Exit(res)

	web.CrawlGHAPI()
}
