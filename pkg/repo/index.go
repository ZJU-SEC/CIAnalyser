package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Index collect the index of usecases
func Index() {
	const FromGitstar = false

	// AutoMigrate Repo table
	err := model.DB.AutoMigrate(Repo{})
	if err != nil {
		panic(err)
	}

	if FromGitstar {
		indexFromGitstar()
	} else {
		indexFromAPI()
	}
}

// IndexFromGitstar crawl the repositories according to gitstar-ranking.com
func indexFromGitstar() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	for _, i := range utils.RandomIntArray(1, 50) {
		page := i
		group.Add(func() {
			crawlGitstarRepo(page)
		})
		group.Add(func() {
			crawlGitstarUserOrg(page, true)  // crawl organizations
			crawlGitstarUserOrg(page, false) // crawl users
		})
	}

	group.Wait()
}

func crawlGitstarRepo(page int) {
	c := utils.CommonCollector()
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)

	defer group.Close()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			group.Add(func() {
				CreateRepo(link)
			})
		}
	})

	pageURL := fmt.Sprintf("https://gitstar-ranking.com/repositories?page=%d", page)
	c.Visit(pageURL)
	group.Wait()
}

func crawlGitstarUserOrg(page int, org bool) {
	c := utils.CommonCollector()
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			group.Add(func() {
				crawlGitstarUserOrgRepo(link)
			})
		}
	})

	var url string
	if org {
		url = fmt.Sprintf("https://gitstar-ranking.com/users?page=%d", page)
	} else {
		url = fmt.Sprintf("https://gitstar-ranking.com/organizations?page=%d", page)
	}
	c.Visit(url)
	group.Wait()
}

// crawl the repositories hosted on Gitstar
func crawlGitstarUserOrgRepo(href string) {
	page := getPageOfUserOrg(href)
	group := parallelizer.NewGroup()
	defer group.Close()

	c := utils.CommonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_full_item" {
			group.Add(func() {
				CreateRepo(link)
			})
		}
	})

	for _, i := range utils.RandomIntArray(1, page) {
		url := fmt.Sprintf("https://gitstar-ranking.com%s?page=%d", href, i)
		c.Visit(url)
	}

	group.Wait()
}

// calculate the total page of user / org
func getPageOfUserOrg(href string) int {
	var num = 0

	c := utils.CommonCollector()

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

func indexFromAPI() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rand.Seed(time.Now().UnixNano())
	since := rand.Intn(config.SINCE_INTERVAL/100)*100 + 250000000
	for since <= config.MAX_SINCE {
		group.Add(func() {
			parseAPI(since)
		})

		since += config.SINCE_INTERVAL
	}

	group.Wait()
}

func parseAPI(since int) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	c := utils.CommonCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Authorization", "token "+utils.RequestGitHubToken())
	})

	c.OnResponse(func(r *colly.Response) {
		// parse repo refs from response body
		type Repo struct {
			Ref string `json:"full_name"`
		}

		var repos []Repo
		json.Unmarshal(r.Body, &repos)

		for _, repo := range repos {
			CreateRepo("/" + repo.Ref)
		}
	})

	c.Visit(fmt.Sprintf("https://api.github.com/repositories?since=%d", since))
}
