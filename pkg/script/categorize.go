package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"golang.org/x/exp/slices"
	"strings"
)

func Categorize() {
	err := model.DB.AutoMigrate(Script{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, _ := model.DB.Model(&Script{}).Rows()
	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		group.Add(func() {
			s := s

			catgegories := getMarketplace(&s)
			if slices.Contains(catgegories, "deployment") {
				s.IsDeployment = true
			}
			if slices.Contains(catgegories, "publishing") {
				s.IsRelease = true
			}
			model.DB.Save(&s)
		})
	}

	group.Wait()
}

func getMarketplace(s *Script) []string {
	var categories []string
	c := utils.CommonCollector()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.HasPrefix(href, "/marketplace/actions") {
			url := "https://github.com" + href
			categories = getCategories(url)
		}
	})

	c.Visit(s.SrcURL())
	return categories
}

func getCategories(url string) []string {
	var categories []string
	c := utils.CommonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.HasPrefix(href, "/marketplace/category/") {
			splits := strings.Split(href, "/")
			categories = append(categories, splits[len(splits)-1])
		}
	})

	c.Visit(url)
	return categories
}
