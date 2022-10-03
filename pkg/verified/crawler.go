package verified

import (
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"fmt"
	"github.com/gocolly/colly"
	"strconv"
	"strings"
	"time"
)

func Crawl() {
	err := model.DB.AutoMigrate(Verified{})
	if err != nil {
		panic(err)
	}

	indexManually("github")
	indexManually("Azure")

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

	for _, c := range utils.ShuffleStringArray(categories) {
		category := c
		indexCategory(category)
	}
}

func indexManually(name string) {
	v := Verified{Name: name}
	v.Create()
}

// indexCategory collect the index for scripts
func indexCategory(category string) {
	const MARKETPLACE_URL = "https://github.com/marketplace"
	num := parseScriptsNum(fmt.Sprintf("%s?category=%s&type=actions", MARKETPLACE_URL, category))

	if num > 1000 {
		queries := []string{
			"sort%3Amatch-desc",
			"sort%3Acreated-desc",
			"sort%3Apopularity-desc",
			"sort%3Amatch-asc",
			"sort%3Acreated-asc",
			"sort%3Apopularity-asc",
		}
		for _, q := range queries {
			indexByQuery(category, q, num)
		}
	} else {
		indexByQuery(category, "", num)
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

func indexByQuery(category string, query string, num int) {
	var totPage int

	if num <= 1000 {
		totPage = (num-1)/20 + 1 // calculate the number of total pages
	} else {
		totPage = 50
	}

	for _, i := range utils.RandomIntArray(1, totPage) {
		page := i
		// complete the query URL
		const MARKETPLACE_URL = "https://github.com/marketplace"
		url := fmt.Sprintf("%s?category=%s&query=%s&page=%d&type=actions",
			MARKETPLACE_URL, category, query, page)

		c := utils.CommonCollector()

		c.Limit(&colly.LimitRule{
			RandomDelay: 5 * time.Second,
		})

		c.OnHTML("p.color-fg-muted.text-small.lh-condensed.mb-1", func(e *colly.HTMLElement) {
			//fmt.Println(e)
			name := strings.Split(e.Text, " ")[1]
			name = strings.ReplaceAll(name, "\n", "")
			span := e.ChildAttr("span.tooltipped.tooltipped-s", "aria-label")

			if span != "" {
				v := Verified{Name: name}
				v.Create()
			}
		})

		c.Visit(url)
	}
}
