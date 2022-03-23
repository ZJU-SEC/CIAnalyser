package main

import (
	"CIHunter/src/analyzer"
	"CIHunter/src/config"
	"CIHunter/src/maintainers"
	"CIHunter/src/models"
	"CIHunter/src/scripts"
	"CIHunter/src/usecases"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	models.Init()

	// crawl gitstar-ranking.com
	switch config.STAGE {
	case "index-maintainers":
		maintainers.Index()
	case "index-usecases":
		usecases.Index()
	case "clone-usecases":
		usecases.Clone()
	case "index-scripts":
		scripts.Index()
	case "analyze":
		analyzer.Analyze()
	}
}
