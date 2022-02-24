package analyzer

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"fmt"
	"io/ioutil"
)

// Analyze the collected data
func Analyze() {
	// analyze the global data
	analyzeGlobal()

	// traverse the workflows
	traverse()

	r.print()
}

type Result struct {
	TotalReposInCentralIndex int
	TotalReposProcessed      int

	TotalAuthor int

	TotalReposWithGHAction int
}

var r = Result{
	TotalReposInCentralIndex: 0,
	TotalReposProcessed:      0,
	TotalAuthor:              0,
	TotalReposWithGHAction:   0,
}

func analyzeGlobal() {
	var c int64

	// count all repos in the central index
	models.DB.Model(&models.Repo{}).Count(&c)
	r.TotalReposInCentralIndex = int(c)

	// count all repos that is checked
	models.DB.Model(&models.Repo{}).Where("checked = ?", true).Count(&c)
	r.TotalReposProcessed = int(c)
}

func traverse() {
	authorDirList, _ := ioutil.ReadDir(config.WORKFLOWS_PATH)
	for _, authorDir := range authorDirList {
		if authorDir.IsDir() {
			r.TotalAuthor++
			repoDirList, _ := ioutil.ReadDir(authorDir.Name())
			for _, repoDir := range repoDirList {
				if repoDir.IsDir() {
					r.TotalReposWithGHAction++
				}
			}
		}
	}
}

func (r *Result) print() {
	fmt.Println("[Global]")
	fmt.Printf("Total repos in the central index: %d\n", r.TotalReposInCentralIndex)
	fmt.Printf("Total repos processed: %d\n", r.TotalReposProcessed)

	fmt.Println("[How CI/CD are configured]")
	fmt.Printf("Total number of the authors: %d\n", r.TotalAuthor)
}
