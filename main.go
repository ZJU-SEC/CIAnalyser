package main

import (
	"CIHunter/src/analyzer"
	"CIHunter/src/config"
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
	case 1:
		usecases.Index()
	case 2:
		usecases.Clone()
	case 3:
		scripts.Index()
	case 4:
		scripts.Clone()
	case 5:
		analyzer.Analyze()
	}
}
