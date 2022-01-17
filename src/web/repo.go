package crawler

import (
	"CIHunter/src/utils"
	"github.com/gocolly/colly"
)

// Crawl the repositories
func Crawl() {
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", utils.RandomString())
	})

}
