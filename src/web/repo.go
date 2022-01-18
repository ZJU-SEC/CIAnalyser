package crawler

import (
	"CIHunter/src/config"
	"CIHunter/src/database"
	"CIHunter/src/utils"
	"fmt"
	"github.com/gocolly/colly"
	"gorm.io/gorm"
	"strconv"
	"strings"
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
			crawlGitstarUserOrgRepo(link)
		}
	})

	for i := 1; i <= 50; i++ {
		userURL := fmt.Sprintf("https://gitstar-ranking.com/users?page=%d", i)
		orgURL := fmt.Sprintf("https://gitstar-ranking.com/organizations?page=%d", i)
		c.Visit(userURL)
		c.Visit(orgURL)
	}
}

// crawl the repositories hosted on Gitstar
func crawlGitstarUserOrgRepo(href string) {
	page := getPageOfUserOrg(href)

	c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_full_item" {
			crawlGitHubRepo(link)
		}
	})

	for i := 1; i <= page; i++ {
		url := fmt.Sprintf("https://gitstar-ranking.com%s?page=%d", href, i)
		c.Visit(url)
	}
}

// calculate the total page of user / org
func getPageOfUserOrg(href string) int {
	var num = 0

	c := commonCollector()

	c.OnHTML("h3", func(e *colly.HTMLElement) {
		header := strings.TrimSpace(e.Text)
		var err error
		num, err = strconv.Atoi(strings.Split(header, " ")[0])
		if err != nil {
			num = 0
		}
	})

	url := "https://gitstar-ranking.com" + href
	c.Visit(url)

	return (num-1)/50 + 1
}

func crawlGitHubRepo(href string) {
	repo := database.Repo{}

	res := database.DB.Where("Ref = ?", href).First(&repo)
	if res.Error == gorm.ErrRecordNotFound {
		// not found, create
		fmt.Printf("create %s\n", href)
		repo = database.Repo{
			Ref: href,
		}
		database.DB.Create(&repo)
	} else if config.NOW.Sub(repo.UpdatedAt) > config.UPDATE_DIFF {
		// TODO update
		fmt.Printf("update %s\n", href)
	}
}
