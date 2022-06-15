package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/gorm/clause"
)

func Crawl() {
	err := model.DB.AutoMigrate(&Script{}, &Verified{})
	if err != nil {
		panic(err)
	}

	// crawl scripts on marketplace
	for _, char := range utils.ALLCHARS {
		getScriptsWithKeyword(string(char))
	}

	// select
	rows, _ := model.DB.Model(&Script{}).Rows()
	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)
		getScriptSource(&s)
		model.DB.Save(&s)
	}
}

func getScriptsWithKeyword(keyword string) {
	resultNum := 0
	c := colly.NewCollector()

	toVisit := make([]Script, 0)
	c.OnHTML("a", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Attr("href"), "/marketplace/actions/") {
			attrs := make([]string, 0)
			for _, attr := range strings.Split(e.Text, "\n") {
				if len(strings.Trim(attr, " \t")) > 0 {
					attrs = append(attrs, strings.Trim(attr, " \t"))
				}
			}

			starCount := 0
			if len(attrs) > 4 {
				if len(attrs[4]) > 6 {
					starCount, _ = strconv.Atoi(attrs[4][0 : len(attrs[4])-6])
				} else {
					fmt.Printf("%s is not a valid count", attrs[4])
				}
			}

			toVisit = append(toVisit, Script{
				Category:      attrs[0],
				StarCount:     starCount,
				OnMarketplace: true,
				Url:           "https://github.com" + e.Attr("href"),
			})
		}
	})
	c.OnHTML("span", func(e *colly.HTMLElement) {
		text := e.Text
		if strings.HasSuffix(text, "results") && len(text) < 15 {
			resultStr := text[0 : len(text)-8]
			resultNum, _ = strconv.Atoi(resultStr)
		}
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

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println(err)
	})
	c.Visit("https://github.com/marketplace?type=actions&query=" + keyword)

	if resultNum > 0 {
		pageNum := (resultNum-1)/20 + 1

		if resultNum <= 1000 {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(resultNum) + " results, collecting")
			for i := 2; i <= pageNum; i++ {
				c.Visit("https://github.com/marketplace?type=actions&query=" + keyword + "&page=" + strconv.Itoa(i))
			}
			model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(toVisit)
		} else {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(resultNum) + " results, refining")
			for _, char := range utils.ALLCHARS {
				getScriptsWithKeyword(keyword + string(char))
			}
		}
	}
}

// this function get official scripts' repositories by entering the link in marketplace page
func getScriptSource(s *Script) {
	get_repo := false

	c := colly.NewCollector()
	c.OnHTML("aside", func(e *colly.HTMLElement) {
		direct_link := ""

		for _, link := range e.ChildAttrs("a", "href") {
			if config.DEBUG {
				fmt.Println("***" + link + "***")
			}
			if isDirectRepoLink(link) {
				direct_link = link
				break
			}
		}

		if direct_link != "" {
			get_repo = true
			identifier := direct_link[19:]
			s.Ref = identifier
			s.Verified = IsVerified(strings.Split(s.Ref, "/")[0])
		}
	})
	c.Visit(s.Url)
	if config.DEBUG && !get_repo {
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
