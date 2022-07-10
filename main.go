package main

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/analyzer"
	"CIAnalyser/pkg/credential"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/repo"
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
	// data collection
	case "crawl-script": // 1st way to get CI scripts
		script.Crawl()
	case "extract-script": // 2nd way to get CI scripts
		script.Extract()
	case "clone-script": // clone scripts
		script.Clone()
	case "dependent":
		repo.GetDependents()
	case "clone-repo":
		repo.Clone()
	// data analysis
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
