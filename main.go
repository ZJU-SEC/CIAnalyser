package main

import (
	"CIHunter/config"
	"CIHunter/pkg/contributor"
	"CIHunter/pkg/model"
	"CIHunter/pkg/repo"
	"CIHunter/pkg/script"
	"fmt"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	model.Init()

	// crawl gitstar-ranking.com
	switch config.STAGE {
	case "index-repo":
		repo.Index()
	case "clone-repo":
		repo.Clone()
	case "extract-script":
		script.Extract()
	case "crawl-contributor":
		contributor.Crawl()
	default:
		fmt.Println("not a valid stage code")
	}
}
