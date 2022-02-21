package web

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"math/rand"
	"time"
)

func CrawlGHAPI() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rand.Seed(time.Now().UnixNano())
	since := rand.Intn(config.SINCE_INTERVAL/100) * 100
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
		r.Headers.Add("Authorization", "token "+config.GITHUB_TOKEN)
	})

	c.OnResponse(func(r *colly.Response) {
		// parse repo refs from response body
		type Repo struct {
			Ref string `json:"full_name"`
		}

		var repos []Repo
		json.Unmarshal(r.Body, &repos)

		for _, repo := range repos {
			models.CreateRepo("/" + repo.Ref)
		}
	})

	c.Visit(fmt.Sprintf("https://api.github.com/repositories?since=%d", since))
}
