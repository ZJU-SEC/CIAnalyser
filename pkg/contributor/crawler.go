package contributor

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
	"CIHunter/pkg/script"
	"CIHunter/utils"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
)

func Crawl() {
	err := model.DB.AutoMigrate(Contributor{}, Contribution{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, _ := model.DB.Model(&script.Script{}).Rows()
	for rows.Next() {
		var s script.Script
		model.DB.ScanRows(rows, &s)

		group.Add(func() {
			s := s
			crawl(&s)
		})
	}
	group.Wait()
}

func crawl(s *script.Script) {
	c := utils.CommonCollector()

	var contributors []Contributor

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Authorization", "token "+utils.RequestGitHubToken())
	})

	c.OnResponse(func(r *colly.Response) {
		err := json.Unmarshal(r.Body, &contributors)
		if err != nil {
			return
		}

		for i := 0; i < len(contributors); i++ {
			contributors[i].fetchOrCreate()

			contribution := Contribution{
				ContributorID: contributors[i].ID,
				Contributor:   contributors[i],
				ScriptID:      s.ID,
				Script:        *s,
			}
			contribution.create()
		}
	})

	c.Visit(fmt.Sprintf("https://api.github.com/repos/%s/contributors", s.SrcRef()))
}

//import (
//	"CIHunter/pkg/model"
//	"CIHunter/utils"
//	"fmt"
//	"github.com/gocolly/colly"
//	"strconv"
//	"strings"
//	"time"
//)
//
//package maintainers
//
//import (
//"CIHunter/pkg/model"
//"CIHunter/utils"
//"fmt"
//"github.com/gocolly/colly"
//"strconv"
//"strings"
//"time"
//)
//
//func Index() {
//	// AutoMigrate Maintainer table
//	err := model.DB.AutoMigrate(model.Maintainer{})
//	if err != nil {
//		panic(err)
//	}
//
//	categories := []string{
//		"",
//		"api-management",
//		"chat",
//		"code-quality",
//		"code-review",
//		"continuous-integration",
//		"container-ci",
//		"mobile-ci",
//		"dependency-management",
//		"deployment",
//		"ides",
//		"learning",
//		"localization",
//		"mobile",
//		"monitoring",
//		"project-management",
//		"publishing",
//		"security",
//		"support",
//		"testing",
//		"utilities",
//	}
//
//	for _, c := range utils.ShuffleStringArray(categories) {
//		category := c
//		indexCategory(category)
//	}
//}
//
//// indexCategory collect the index for scripts
//func indexCategory(category string) {
//	const MARKETPLACE_URL = "https://github.com/marketplace"
//	num := parseScriptsNum(fmt.Sprintf("%s?category=%s&type=actions", MARKETPLACE_URL, category))
//
//	if num > 1000 {
//		queries := []string{
//			"sort%3Amatch-desc",
//			"sort%3Acreated-desc",
//			"sort%3Apopularity-desc",
//			"sort%3Amatch-asc",
//			"sort%3Acreated-asc",
//			"sort%3Apopularity-asc",
//		}
//		for _, q := range queries {
//			indexByQuery(category, q, num)
//		}
//	} else {
//		indexByQuery(category, "", num)
//	}
//}
//
//// parseScriptsNum get the number of the scripts according to the
//func parseScriptsNum(url string) int {
//	c := utils.CommonCollector()
//
//	page := 0
//	c.OnHTML("span[class=text-bold]", func(e *colly.HTMLElement) {
//		page, _ = strconv.Atoi(strings.Split(e.Text, " ")[0])
//	})
//
//	c.Visit(url)
//	return page
//}
//
//func indexByQuery(category string, query string, num int) {
//	var totPage int
//
//	if num <= 1000 {
//		totPage = (num-1)/20 + 1 // calculate the number of total pages
//	} else {
//		totPage = 50
//	}
//
//	for _, i := range utils.RandomIntArray(1, totPage) {
//		page := i
//		// complete the query URL
//		const MARKETPLACE_URL = "https://github.com/marketplace"
//		url := fmt.Sprintf("%s?category=%s&query=%s&page=%d&type=actions",
//			MARKETPLACE_URL, category, query, page)
//
//		c := utils.CommonCollector()
//
//		c.Limit(&colly.LimitRule{
//			RandomDelay: 5 * time.Second,
//		})
//
//		c.OnHTML("p.color-fg-muted.text-small.lh-condensed.mb-1", func(e *colly.HTMLElement) {
//			//fmt.Println(e)
//			name := strings.Split(e.Text, " ")[1]
//			name = strings.ReplaceAll(name, "\n", "")
//			span := e.ChildAttr("span.tooltipped.tooltipped-s", "aria-label")
//
//			m := model.Maintainer{Name: name}
//			if span != "" {
//				m.Verified = true
//			}
//
//			m.Create()
//		})
//
//		c.Visit(url)
//	}
//}
