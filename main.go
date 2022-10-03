package main

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/analyzer"
	"CIAnalyser/pkg/contributor"
	"CIAnalyser/pkg/credential"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/repo"
	"CIAnalyser/pkg/script"
	"CIAnalyser/pkg/verified"
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
	case "index-repo":
		repo.Index()
	case "clone-repo":
		repo.Clone()
	case "extract-script":
		script.Extract()
	case "clone-script":
		script.Clone()
	case "categorize-script":
		script.Categorize()
	case "crawl-contributor":
		contributor.Crawl()
	case "crawl-verified":
		verified.Crawl()
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
	default:
		fmt.Println("not a valid stage code")
	}
}
