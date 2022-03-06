package scripts

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"strconv"
	"strings"
)

func Index() {
	// AutoMigrate Script table
	err := models.DB.AutoMigrate(models.Script{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	categories := []string{
		"",
		"api-management",
		"chat",
		"code-quality",
		"code-review",
		"continuous-integration",
		"container-ci",
		"mobile-ci",
		"dependency-management",
		"deployment",
		"ides",
		"learning",
		"localization",
		"mobile",
		"monitoring",
		"project-management",
		"publishing",
		"security",
		"support",
		"testing",
		"utilities",
	}

	//queries := []string{
	//	"sort%3Amatch-desc",
	//	"sort%3Acreated-desc",
	//	"sort%3Apopularity-desc",
	//	"sort%3Amatch-asc",
	//	"sort%3Acreated-asc",
	//	"sort%3Apopularity-asc",
	//}

	for _, c := range categories {
		category := c
		group.Add(func() {
			indexCategory(category)
		})
	}

	group.Wait()
}

func indexCategory(category string) {
	const MARKETPLACE_URL = "https://github.com/marketplace"
	num := parseScriptsNum(fmt.Sprintf("%s?category=%s&type=actions", MARKETPLACE_URL, category))

	fmt.Println(num)

	if num > 1000 {

	} else {

	}

}

// parseScriptsNum get the number of the scripts according to the
func parseScriptsNum(url string) int {
	c := utils.CommonCollector()

	page := 0
	c.OnHTML("span[class=text-bold]", func(e *colly.HTMLElement) {
		page, _ = strconv.Atoi(strings.Split(e.Text, " ")[0])
	})

	c.Visit(url)
	return page
}
