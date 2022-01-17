package crawler

import (
	"CIHunter/src/utils"
	"fmt"
	"github.com/gocolly/colly"
)

// Crawl the repositories
func Crawl() {
	crawlGitstarRepo()
}

func crawlGitstarRepo() {
	c := colly.NewCollector()

	// set random `User-Agent`
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", utils.RandomString())
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println(link)
		// TODO parse more

	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	for i := 1; i <= 50; i++ {
		pageURL := fmt.Sprintf("https://gitstar-ranking.com/repositories?page=%d", i)
		c.Visit(pageURL)
	}
}
