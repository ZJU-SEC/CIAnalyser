package analyzer

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Analyze the collected data
func Analyze() {
	prepare()

	// traverse the workflows
	traverse()

	output()
}

type GHMeasure struct {
	ID                 uint `gorm:"primaryKey;autoIncrement"`
	RepoRef            string
	ConfigurationCount int `gorm:"default:0"`
}

type GHJob struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID uint
	GHMeasure   GHMeasure `gorm:"foreignKey:GHMeasureID"`
}

type GlobalCount struct {
	TotalCentralIndex int
	TotalProcessed    int
	TotalAuthors      int
	TotalHasGHAction  int
}

var Count = GlobalCount{
	TotalCentralIndex: 0,
	TotalProcessed:    0,
	TotalAuthors:      0,
	TotalHasGHAction:  0,
}

// prepare tables
func prepare() {
	models.DB.Migrator().CreateTable(&GHMeasure{})
	models.DB.Migrator().CreateTable(&GHJob{})
	models.DB.Migrator().CreateTable(&GHRunner{})
	models.DB.Migrator().CreateTable(&GHUse{})
}

func output() {
	fmt.Println("[Global]")
	var c int64

	// count all repos in the central index
	models.DB.Model(&models.Repo{}).Count(&c)
	fmt.Printf("Total repos in the central index: %d\n", c)

	// count all repos that is checked
	models.DB.Model(&models.Repo{}).Where("checked = ?", true).Count(&c)
	fmt.Printf("Total repos processed: %d\n\n", c)

	fmt.Println("[How CI/CD are configured]")
	fmt.Printf("Total number of the authors: %d\n", Count.TotalAuthors)

	models.DB.Model(&GHMeasure{}).Count(&c)
	fmt.Printf("Total repos using GitHub Actions: %d\n", c)

	models.DB.Migrator().DropTable(&GHRunner{})
	models.DB.Migrator().DropTable(&GHUse{})
	models.DB.Migrator().DropTable(&GHMeasure{})
}

func traverse() {
	authorDirList, _ := ioutil.ReadDir(config.WORKFLOWS_PATH)
	for _, authorDir := range authorDirList {
		if !authorDir.IsDir() {
			continue // not dir, skip
		}

		Count.TotalAuthors++ // count this author

		repoDirList, _ := ioutil.ReadDir(path.Join(config.WORKFLOWS_PATH, authorDir.Name()))
		for _, repoDir := range repoDirList {
			if repoDir.IsDir() {
				Count.TotalHasGHAction++
				repoPath := path.Join(config.WORKFLOWS_PATH, authorDir.Name(), repoDir.Name())

				// analyze this repository specifically
				analyzeRepo(repoPath)
			}
		}
	}
}

// analyzeRepo glob the given path, check yaml files and process
func analyzeRepo(repoPath string) {
	// create measure record
	measure := GHMeasure{
		RepoRef:            strings.TrimPrefix(repoPath, config.WORKFLOWS_PATH),
		ConfigurationCount: 0,
	}

	models.DB.Create(&measure)

	filepath.Walk(repoPath, func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		if err != nil || info.IsDir() || (ext != ".yml" && ext != ".yaml") {
			return err
		}

		measure.ConfigurationCount++
		models.DB.Save(&measure)

		f, err := os.Open(p)
		if err != nil {
			return err
		}

		analyzeJobs(f, &measure)

		return nil
	})
}
