package analyzer

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"fmt"
	"github.com/shomali11/parallelizer"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// Analyze the collected data
func Analyze() {
	prepare()

	output()

	//finish()
}

type GHMeasure struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement"`
	RepoRef            string `gorm:"uniqueIndex"`
	ConfigurationCount int    `gorm:"default:0"`
}

type GHJob struct {
	ID             uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID    uint
	GHMeasure      GHMeasure `gorm:"foreignKey:GHMeasureID"`
	PassCredential bool      `gorm:"default:false"`
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

// countAuthor increment the count for the total authors
func countAuthor() {
	var mutex sync.Mutex
	mutex.Lock()

	Count.TotalAuthors++

	mutex.Unlock()
}

// countRepo increment the count for the total repositories
func countRepo() {
	var mutex sync.Mutex
	mutex.Lock()

	Count.TotalHasGHAction++

	mutex.Unlock()
}

// prepare tables
func prepare() {
	models.DB.Migrator().CreateTable(&GHMeasure{})
	models.DB.Migrator().CreateTable(&GHJob{})
	models.DB.Migrator().CreateTable(&GHRunner{})
	models.DB.Migrator().CreateTable(&GHUse{})

	// traverse the workflows
	traverse()
}

func finish() {
	models.DB.Migrator().DropTable(&GHRunner{})
	models.DB.Migrator().DropTable(&GHUse{})
	models.DB.Migrator().DropTable(&GHJob{})
	models.DB.Migrator().DropTable(&GHMeasure{})
}

func output() {
	fmt.Println("[Global]")
	var c int64

	// count all repos in the central index
	models.DB.Model(&models.Repo{}).Count(&c)
	fmt.Printf("Total repos in the central index: %d\n", c)

	// count all repos that is checked
	models.DB.Model(&models.Repo{}).Where("checked = ?", true).Count(&c)
	fmt.Printf("Total repos processed: %d\n", c)

	fmt.Println("\n[How CI/CD are configured]")
	fmt.Printf("Total number of the authors: %d\n", Count.TotalAuthors)

	models.DB.Model(&GHMeasure{}).Count(&c)
	fmt.Printf("Total repos using GitHub Actions: %d\n", c)
	models.DB.Model(&GHJob{}).Count(&c)
	fmt.Printf("Total jobs: %d\n", c)

	//------//
	// uses //
	//------//
	fmt.Println("\n[How scripts are imported]")
	models.DB.Model(&GHUse{}).Count(&c)
	fmt.Printf("Total occurrences of `uses` field: %d\n", c)
	models.DB.Model(&GHUse{}).Where("use LIKE ?", "docker://%").Count(&c)
	fmt.Printf("Total occurrences of docker images: %d\n", c)
	models.DB.Model(&GHUse{}).Where("use NOT LIKE ? AND use NOT LIKE ?", "%@%", "docker://%").Count(&c)
	fmt.Printf("Total occurrences of self-written scripts: %d\n", c)

	n := 20
	fmt.Println("\n[Popular", n, "scripts]")
	analyzePopularNthUses(n)

	//fmt.Println("\n[Possible scripts containing CVEs]")
	//analyzeCVE()
}

func traverse() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	authorDirList, _ := ioutil.ReadDir(config.WORKFLOWS_PATH)
	for _, authorDir := range authorDirList {
		if !authorDir.IsDir() {
			continue // not dir, skip
		}

		group.Add(func() {
			traverseAuthor(authorDir)
		})
	}
}

func traverseAuthor(authorDir fs.FileInfo) {
	countAuthor()

	repoDirList, _ := ioutil.ReadDir(path.Join(config.WORKFLOWS_PATH, authorDir.Name()))
	for _, repoDir := range repoDirList {
		if repoDir.IsDir() {
			countRepo()

			repoPath := path.Join(config.WORKFLOWS_PATH, authorDir.Name(), repoDir.Name())

			// analyze this repository specifically
			analyzeRepo(repoPath)
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
