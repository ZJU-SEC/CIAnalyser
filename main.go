package main

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/analyzer"
	"CIAnalyser/pkg/contributor"
	"CIAnalyser/pkg/credential"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/repo"
	"CIAnalyser/pkg/router"
	"CIAnalyser/pkg/script"
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
	case "migrate":
		router.RelationsToRepos()
	default:
		fmt.Println("not a valid stage code")
	}
}
