package crawler

import (
	"CIHunter/src/utils"
	"fmt"
	"github.com/gocolly/colly"
)

// Crawl the repositories
func Crawl() {
	crawlGitstarRepo()
	crawlGitstarUserOrg()
}


func commonCollector() *colly.Collector {
    c := colly.NewCollector()

	// set random `User-Agent`
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", utils.RandomString())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

    return c
}


func crawlGitstarRepo() {
    c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			crawlGitHubRepo(link)
		}
	})


	for i := 1; i <= 50; i++ {
		pageURL := fmt.Sprintf("https://gitstar-ranking.com/repositories?page=%d", i)
		c.Visit(pageURL)
	}
}

func crawlGitstarUserOrg() {
    c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
            fmt.Println(link)
            // TODO parse user / org
		}
    })

    for i := 1; i <= 50; i++ {
		userURL := fmt.Sprintf("https://gitstar-ranking.com/users?page=%d", i)
        orgURL := fmt.Sprintf("https://gitstar-ranking.com/organizations?page=%d", i)
		c.Visit(userURL)
        c.Visit(orgURL)
    }
}

func crawlGitHubRepo(href string) {
    // TODO parse repo
    fmt.Println(href)
}


