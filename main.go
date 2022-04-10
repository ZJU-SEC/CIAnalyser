package main

import (
	"CIHunter/config"
	"CIHunter/pkg/analyzer"
	"CIHunter/pkg/contributor"
	"CIHunter/pkg/credential"
	"CIHunter/pkg/model"
	"CIHunter/pkg/repo"
	"CIHunter/pkg/script"
	"CIHunter/pkg/verified"
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
	case "clone-script":
		script.Clone()
	case "crawl-contributor":
		contributor.Crawl()
	case "crawl-verified":
		verified.Crawl()
	case "extract-credential":
		credential.Extract()
	case "label-usage":
		script.Label()
	case "analyze":
		analyzer.Analyze()
	default:
		fmt.Println("not a valid stage code")
	}
}
