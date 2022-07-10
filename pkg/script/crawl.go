package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

func Crawl() {
	err := model.DB.AutoMigrate(&Script{}, &Verified{})
	if err != nil {
		panic(err)
	}

	// manually add 2 verified creators
	manuallyVerify("github")
	manuallyVerify("azure")

	// crawl scripts on marketplace
	for _, char := range utils.ALLCHARS {
		getScriptsWithKeyword(string(char))
	}

	// parse scripts' details
	rows, _ := model.DB.Model(&Script{}).Rows()
	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)
		getScriptDetail(&s)
		model.DB.Save(&s)
	}
}

func getScriptsWithKeyword(keyword string) {
	resultNum := 0
	c := utils.CommonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Attr("href"), "/marketplace/actions/") {
			starStr := e.ChildText("span.text-small.color-fg-muted.text-bold")

			s := Script{
				Url:           "https://github.com" + e.Attr("href"),
				StarCount:     strings.Split(starStr, " ")[0],
				OnMarketplace: true,
			}
			s.Create()
		}
	})

	// parse number of results
	c.OnHTML("span[class=text-bold]", func(e *colly.HTMLElement) {
		resultNum, _ = strconv.Atoi(strings.Split(e.Text, " ")[0])
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

	c.Visit("https://github.com/marketplace?type=actions&query=" + keyword)

	if resultNum > 0 {
		pageNum := (resultNum-1)/20 + 1

		if resultNum <= 1000 {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(resultNum) + " results, collecting ...")
			for i := 2; i <= pageNum; i++ {
				c.Visit("https://github.com/marketplace?type=actions&query=" + keyword + "&page=" + strconv.Itoa(i))
			}
		} else {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(resultNum) + " results, splitting ...")
			for _, char := range utils.ALLCHARS {
				getScriptsWithKeyword(keyword + string(char))
			}
		}
	}
}

// this function get official scripts' repositories by entering the link in marketplace page
func getScriptDetail(s *Script) {
	getDetail := false

	c := utils.CommonCollector()
	c.OnHTML("aside", func(e *colly.HTMLElement) {
		directLink := ""

		for _, link := range e.ChildAttrs("a", "href") {
			if config.DEBUG {
				fmt.Println("***" + link + "***")
			}
			if isDirectRepoLink(link) {
				directLink = link
				break
			}
		}

		getDetail = true
		identifier := directLink[19:]
		s.Ref = identifier
		s.Verified = IsVerified(strings.Split(s.Ref, "/")[0])
	})

	c.OnHTML("a.topic-tag.topic-tag-link.f6", func(e *colly.HTMLElement) {
		getDetail = true
		s.Category = s.Category + strings.TrimSpace(e.Text) + ";"
	})

	c.Visit(s.Url)
	if config.DEBUG && !getDetail {
		fmt.Printf("didn't get repo at %s\n", s.Url)
	}
}

func isDirectRepoLink(candidate string) bool {
	if config.DEBUG {
		fmt.Println("determining candidate: " + candidate)
	}
	if strings.HasPrefix(candidate, "https://github.com/") && strings.Count(candidate, "/") == 4 {
		return true
	} else {
		return false
	}
}
