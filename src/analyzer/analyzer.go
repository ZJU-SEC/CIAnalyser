package analyzer

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"fmt"
	"github.com/shomali11/parallelizer"
	"io/ioutil"
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
	models.DB.Migrator().CreateTable(&GHCredential{})

	// traverse the workflows
	traverse()
}

func finish() {
	models.DB.Migrator().DropTable(&GHRunner{})
	models.DB.Migrator().DropTable(&GHUse{})
	models.DB.Migrator().DropTable(&GHJob{})
	models.DB.Migrator().DropTable(&GHMeasure{})
	models.DB.Migrator().DropTable(&GHCredential{})
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

	n := 10
	fmt.Println("\n[Popular", n, "scripts]")
	outputPopularNthUses(n)

	//fmt.Println("\n[Possible scripts containing CVEs]")
	//analyzeCVE()

	fmt.Println("\n[Runtime Environments]")
	outputRunners()

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
