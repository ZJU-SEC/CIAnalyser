package scripts

import (
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"fmt"
	"github.com/gocolly/colly"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func Index() {
	// AutoMigrate Script table
	err := models.DB.AutoMigrate(models.Script{})
	if err != nil {
		panic(err)
	}

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

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			href := e.Attr("href")
			if strings.Contains(href, "/marketplace/actions/") {
				parseEntry(href)
			}
		})

		c.Visit(url)
	}
}

func parseEntry(href string) {
	uniqueRef := strings.TrimPrefix(href, "/marketplace/actions/")
	res := models.DB.Where("ref = ?", uniqueRef).First(&models.Script{})
	if res.Error != gorm.ErrRecordNotFound {
		return
	}

	const DOMAIN = "https://github.com"
	c := utils.CommonCollector()
	script := models.Script{Ref: uniqueRef}

	c.OnHTML("h1[class]", func(e *colly.HTMLElement) {
		if e.Attr("class") == "f1 text-normal mb-1" {
			script.Name = e.Text
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("href"), "https://github.com/") &&
			e.Attr("class") == "d-block mb-2" {
			script.SrcRef = strings.TrimPrefix(e.Attr("href"), DOMAIN)
		}
	})

	c.Visit(DOMAIN + href)
	if len(script.SrcRef) > 0 && len(script.Name) > 0 {
		script.Create()
	}
}
