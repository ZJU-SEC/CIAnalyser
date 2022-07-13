package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/script"
	"CIAnalyser/utils"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
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

// getDependents crawl package identifier from url.
// such interface design makes recovery easy (but not that easy thanks to the complex webpage design)
func getDependents(packageURL string, s *script.Script) {
	if config.DEBUG {
		fmt.Println("[DEBUG] on", s.Ref, "visiting", packageURL)
	}

	c := utils.CommonCollector()
	//finished := false
	s.LastVisitedURL = packageURL

	// parse dependents
	c.OnHTML("div.Box-row.d-flex.flex-items-center", func(e *colly.HTMLElement) {
		childInfo := e.ChildAttrs("span.color-fg-muted.text-bold.pl-3", "#text")
		fmt.Println(childInfo)
	})

	c.Visit(packageURL)

	//if !finished {
	//	// TODO recovery mechanism
	//}
	//
	////use the same crawler to visit next page if there is one
	//c.OnHTML("a", func(e *colly.HTMLElement) {
	//	if strings.Contains(e.Text, "Next") && e.Attr("rel") == "nofollow" {
	//		if strings.Contains(e.Attr("href"), "/network/dependents") {
	//			if config.DEBUG {
	//				fmt.Println("visiting dependents subpage: " + e.Attr("href"))
	//			}
	//			lastVisited = e.Attr("href")
	//			c.Visit(e.Attr("href"))
	//		}
	//	} else if e.Attr("data-hovercard-type") == "repository" && e.Attr("class") == "text-bold" {
	//		toInsert = append(toInsert, Repo{
	//			DependencyIdentifier: url_to_identifier(packageURL),
	//			PackageIdentifier:    packageName,
	//			DependentIdentifier:  url_to_identifier(e.Attr("href")),
	//		})
	//	}
	//})
	//
	//starCount := make([]int, 0)
	//forkCount := make([]int, 0)
	//
	//// get stars count and fork count
	//c.OnHTML("span", func(e *colly.HTMLElement) {
	//	if e.ChildAttr("svg", "class") == "octicon octicon-star" {
	//		starCountStr := strings.Trim(e.Text, " \t\n\r")
	//		count, err := strconv.Atoi(strings.ReplaceAll(starCountStr, ",", ""))
	//		if err != nil {
	//			fmt.Fprintln(os.Stderr, starCountStr+" is not a valid star count!")
	//		}
	//		starCount = append(starCount, count)
	//	} else if e.ChildAttr("svg", "class") == "octicon octicon-repo-forked" {
	//		forkCountStar := strings.Trim(e.Text, " \t\n\r")
	//		count, err := strconv.Atoi(strings.ReplaceAll(forkCountStar, ",", ""))
	//		if err != nil {
	//			fmt.Fprintln(os.Stderr, forkCountStar+" is not a valid fork count!")
	//		}
	//		forkCount = append(forkCount, count)
	//	}
	//})
	//
	//// if the `Next` button goes gray, all the dependents are collected
	//c.OnHTML("button", func(e *colly.HTMLElement) {
	//	if e.Attr("disabled") == "disabled" && e.Text == "Next" {
	//		finished = true
	//	}
	//})
	//
	//// if no dependents, the info will show, but not a disabled next button
	//c.OnHTML("p", func(e *colly.HTMLElement) {
	//	if strings.Contains(e.Text, "We haven’t found any dependents for this repository yet.") || strings.Contains(e.Text, "keep looking!") {
	//		finished = true
	//		if config.DEBUG {
	//			fmt.Println("visiting package with no dependents: " + packageURL)
	//		}
	//	}
	//})
	//
	//c.Visit(packageURL)
	//
	//if !finished {
	//	if config.DEBUG {
	//		fmt.Println("crawl dependents of" + scriptRef + " failed while accessing " + lastVisited)
	//	}
	//	old_checked := make([]CheckedPackage, 0)
	//	model.DB.Where("repo_identifier = ", scriptRef).Find(&old_checked)
	//	old_count := 0
	//	if len(old_checked) > 0 {
	//		if lastVisited == old_checked[0].LastVisitedUrl && strings.HasSuffix(lastVisited, "/network/dependents") {
	//			old_count = int(old_checked[0].FailedTimes)
	//		}
	//	}
	//	model.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&CheckedPackage{
	//		RepoIdentifier:    repo_id,
	//		PackageIdentifier: packageName,
	//		Finished:          false,
	//		LastVisitedUrl:    lastVisited,
	//		FailedTimes:       uint32(old_count) + 1,
	//	})
	//} else {
	//	if config.DEBUG {
	//		fmt.Println("crawl " + repo_id + " done with no error")
	//	}
	//	model.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&CheckedPackage{
	//		RepoIdentifier:    repo_id,
	//		PackageIdentifier: packageName,
	//		Finished:          true,
	//		LastVisitedUrl:    lastVisited,
	//		FailedTimes:       0,
	//	})
	//}
	//
	//if len(toInsert) > 0 {
	//	model.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&toInsert, 100)
	//}
}
