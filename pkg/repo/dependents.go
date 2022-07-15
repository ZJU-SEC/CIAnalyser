package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/script"
	"CIAnalyser/utils"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"strconv"
	"strings"
)

// GetDependents search all dependent repos upon a CI script
func GetDependents() {
	err := model.DB.AutoMigrate(&Repo{}, &Dependency{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, _ := model.DB.Model(&script.Script{}).Where("checked = ?", false).Rows()

	for rows.Next() {
		var s script.Script
		model.DB.ScanRows(rows, &s)

		group.Add(func() {
			getPackages(&s)
		})
	}

	group.Wait()
}

// getPackages reach the list of packages in dependents menu
func getPackages(s *script.Script) {
	// has last visited url, continue it
	if s.LastVisitedURL != "" {
		getDependents(s.LastVisitedURL, s)
	} else {
		dependentURL := s.SrcURL() + "/network/dependents"
		c := utils.CommonCollector()

		isList := false

		// find package list
		c.OnHTML("a.select-menu-item", func(e *colly.HTMLElement) {
			isList = true
			url := e.Attr("href")
			name := strings.Trim(e.Text, "\n ")
			if name == s.Ref {
				getDependents("https://github.com"+url, s)
			}
		})

		c.Visit(dependentURL)

		if !isList {
			getDependents(dependentURL, s)
		}
	}
}

// getDependents crawl package identifier from url.
// such interface design makes recovery easy (but not that easy thanks to the complex webpage design)
func getDependents(packageURL string, s *script.Script) {
	if config.DEBUG {
		fmt.Println("[DEBUG] on", s.Ref, "visiting", packageURL)
	}

	c := utils.CommonCollector()

	// parse dependents
	c.OnHTML("div.Box-row.d-flex.flex-items-center", func(e *colly.HTMLElement) {
		childInfo := strings.Fields(e.ChildText("span.color-fg-muted.text-bold.pl-3"))
		star, _ := strconv.Atoi(childInfo[0])
		fork, _ := strconv.Atoi(childInfo[1])
		repoRef := e.ChildAttr("a.text-bold", "href")

		repo := Repo{
			Ref:       repoRef,
			StarCount: star,
			ForkCount: fork,
		}

		repo.FetchOrCreate()

		relation := Dependency{
			RepoID:   repo.ID,
			Repo:     repo,
			ScriptID: s.ID,
			Script:   *s,
		}

		relation.Create()
		fmt.Println("dependency:", repo.Ref, "<->", s.Ref, "created")
	})

	// parse next page
	c.OnHTML("a.btn.btn-outline.BtnGroup-item", func(e *colly.HTMLElement) {
		if e.Text == "Next" {
			nextURL := e.Attr("href")
			s.LastVisitedURL = nextURL
			c.Visit(nextURL)
		}
	})

	// disabled button -> finished -> clear URL
	c.OnHTML("button.btn.btn-outline.BtnGroup-item", func(e *colly.HTMLElement) {
		if e.Text == "Next" {
			s.LastVisitedURL = ""
			s.Checked = true
		}
	})
	s.Update()

	c.Visit(packageURL)
}
