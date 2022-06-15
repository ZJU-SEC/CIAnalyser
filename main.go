package main

import (
	"CIHunter/config"
	"CIHunter/pkg/analyzer"
	"CIHunter/pkg/contributor"
	"CIHunter/pkg/credential"
	"CIHunter/pkg/model"
	"CIHunter/pkg/repo"
	"CIHunter/pkg/router"
	"CIHunter/pkg/script"
	"fmt"
	"os"
)

func main() {
	// initialize configurations
	config.Init()

	// initialize models
	model.Init()

	if len(os.Args) < 2 {
		panic("require an argument")
	}

	switch os.Args[1] {
	case "crawl-script":
		script.Crawl()
	case "extract-script":
		script.Extract()
	case "clone-script":
		script.Clone()
	//case "dependent":
	//	repo.GetDependentsReposAll()
	case "clone-repo":
		repo.Clone()
	case "crawl-contributor":
		contributor.Crawl()
	case "extract-credential":
		credential.Extract()
	case "label-usage":
		script.Label()
	case "label-lag":
		script.Lag()
	case "parse-using":
		script.ParseUsing()
	case "analyze":
		analyzer.Analyze()
	case "recover":
		for router.RecoverCrawlAll() {
			// do nothing
		}
	default:
		fmt.Println("not a valid stage code")
	}
}
