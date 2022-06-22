package router

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/script"

	"github.com/gocolly/colly"
	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// func get_official_actions_all() {
func GetOfficialActionsAll() {
	// db, err := gorm.Open(sqlite.Open(DATABASE_PATH), &gorm.Config{})
	// if err != nil {
	// 	panic("fail to connect to database")
	// }
	db := model.DB
	db.AutoMigrate(&OfficialAction{})
	// db.AutoMigrate(&script.Script{})

	for _, char := range ALLCHARS {
		get_official_with_kerword(string(char), db)
	}
}

func get_official_with_kerword(keyword string, db *gorm.DB) {

	number_of_result := 0
	c := colly.NewCollector()

	to_visit := make([]OfficialAction, 0)
	c.OnHTML("a", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Attr("href"), "/marketplace/actions/") {
			attributes := strings.Split(e.Text, "\n")
			valid_attributes := make([]string, 0)
			for _, attr := range attributes {
				if len(strings.Trim(attr, " \t")) > 0 {
					valid_attributes = append(valid_attributes, strings.Trim(attr, " \t"))
				}
			}

			star_cnt := 0
			if len(valid_attributes) > 4 {
				if len(valid_attributes[4]) > 6 {
					star_cnt, _ = strconv.Atoi(valid_attributes[4][0 : len(valid_attributes[4])-6])
				} else {
					fmt.Printf("%s is not a valid count", valid_attributes[4])
				}
			}

			desc := ""
			if len(valid_attributes) > 3 {
				desc = valid_attributes[3]
			}
			to_visit = append(to_visit, OfficialAction{
				Name:        valid_attributes[1],
				Category:    valid_attributes[0],
				Author:      valid_attributes[2][2:],
				Description: desc,
				NumStars:    uint(star_cnt),
				Url:         "https://github.com" + e.Attr("href"),
			})
		}
	})
	c.OnHTML("span", func(e *colly.HTMLElement) {
		text := e.Text
		if strings.HasSuffix(text, "results") && len(text) < 15 {
			string_of_result := text[0 : len(text)-8]
			number_of_result, _ = strconv.Atoi(string_of_result)
		}
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println(err)
	})
	c.Visit("https://github.com/marketplace?type=actions&query=" + keyword)
	random_delay_ms(1000)

	if number_of_result > 0 {

		number_of_pages := (number_of_result-1)/20 + 1

		if number_of_result <= 1000 {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(number_of_result) + " results, collecting")
			for i := 2; i <= number_of_pages; i++ {
				c.Visit("https://github.com/marketplace?type=actions&query=" + keyword + "&page=" + strconv.Itoa(i))
				random_delay_ms(1000)
			}
			db.Clauses(clause.OnConflict{DoNothing: true}).Create(to_visit)
		} else {
			fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(number_of_result) + " results, refining")
			for _, char := range ALLCHARS {
				get_official_with_kerword(keyword+string(char), db)
			}
		}
	}
}

// this function get official actions' repositories by entering the link in marketplace page
func GetOfficialReposAll() {
	// to_insert := make([]AllAction, 0)
	get_repo := false
	// db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Database failed at get official repos stage")
	// }
	db := model.DB
	db.AutoMigrate(&script.Script{})
	var official_actions []OfficialAction
	db.Find(&official_actions)
	rand.Shuffle(len(official_actions), func(i, j int) { official_actions[i], official_actions[j] = official_actions[j], official_actions[i] })
	c := colly.NewCollector()
	c.OnHTML("aside", func(e *colly.HTMLElement) {
		all_links := e.ChildAttrs("a", "href")
		right_link := ""
		direct_link := ""

		for _, link := range all_links {
			if DEBUG {
				fmt.Println("***" + link + "***")
			}
			if strings.HasPrefix(link, "/login?") {
				right_link = link
				break
			} else if is_direct_repo_link(link) {
				direct_link = link
				break
			}
		}

		if right_link != "" {
			get_repo = true
			identifier := strings.ReplaceAll(right_link[20:], "%2F", "/")
			// db.Clauses(clause.OnConflict{DoNothing: true}).Create(&AllAction{
			// 	Identifier: identifier,
			// 	IsOfficial: true,
			// 	Checked:    false,
			// 	Url:        "https://github.com/" + identifier,
			// })
			db.Clauses(clause.OnConflict{DoNothing: true}).Create(
				&script.Script{
					Ref:      identifier,
					Verified: true,
				},
			)
			fmt.Println("with login: " + identifier)
		} else if direct_link != "" {
			get_repo = true
			identifier := direct_link[19:]
			// db.Clauses(clause.OnConflict{DoNothing: true}).Create(&AllAction{
			// 	Identifier: identifier,
			// 	IsOfficial: true,
			// 	Checked:    false,
			// 	Url:        direct_link,
			// })
			db.Clauses(clause.OnConflict{DoNothing: true}).Create(
				&script.Script{
					Ref:      identifier,
					Verified: true,
				},
			)
			fmt.Println("direct link: " + identifier)
		} else {
			fmt.Fprintf(os.Stderr, "Cannot get repository from %s\n", strings.Join(all_links, ", "))
		}
	})
	for _, action := range official_actions {
		get_repo = false
		c.Visit(action.Url)
		random_delay_ms(300)
		if DEBUG && !get_repo {
			fmt.Printf("didn't get repo at %s\n", action.Url)
		}
	}

}

// func get_dependents_repos_all() {
func GetDependentsReposAll() {
	// db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Database failed at get dependents stage")
	// }
	db := model.DB

	db.AutoMigrate(&DependRelation{})
	db.AutoMigrate(&CheckedPackage{})
	// var to_get []AllAction
	var to_get []script.Script
	db.Where("checked = ?", false).Find(&to_get)

	for _, record := range to_get {
		// get_dependents(record.Url+"/network/dependents", db, record.Identifier)
		// db.Model(&AllAction{})
		//   .Where("identifier = ?", record.Identifier)
		//   .Update("checked", true)
		get_dependents("https://github.com/"+record.Ref+"/network/dependents", db, record.Ref)
		db.Model(&script.Script{}).Where("identifier = ?", record.Ref).Update("checked", true)
		random_delay_ms(1000)
	}
}

func get_dependents(url string, db *gorm.DB, repo_id string) {
	db.AutoMigrate(&DependRelation{})
	c := colly.NewCollector()
	to_visit := make(map[string](string), 2)
	is_package_list := false
	c.OnHTML("details-menu", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, elem *colly.HTMLElement) {
			if elem.Attr("class") == "select-menu-list" {
				elem.ForEach("a", func(_ int, inside *colly.HTMLElement) {
					// fmt.Println(inside.Attr("href"))
					// fmt.Println(strings.Trim(inside.Text, "\n "))
					to_visit[strings.Trim(inside.Text, "\n ")] = inside.Attr("href")
				})
			} else if elem.Attr("class") == "select-menu-header" && strings.Contains(elem.Text, "Packages") {
				is_package_list = true
				// fmt.Println("get package")
			}
		})
		if is_package_list {
			for package_name, package_path := range to_visit {
				if DEBUG {
					fmt.Println("visiting package dependents" + " https://github.com" + package_path)
				}
				random_delay_ms(500)
				get_package_dependents("https://github.com"+package_path, package_name, db, repo_id)
			}
		}
	})
	c.Visit(url)
	if !is_package_list {
		fmt.Println("visiting repository dependents " + url)
		random_delay_ms(500)
		get_package_dependents(url, "", db, repo_id)
	}
}

// crawl package identifier from url.
// such interface design makes recovery easy (but not that easy thanks to the complex webpage design)
func get_package_dependents(url string, package_name string, db *gorm.DB, repo_id string) {
	// next_page := ""
	c := colly.NewCollector()
	to_insert := make([]DependRelation, 0)
	finished := false
	last_visited := url
	if DEBUG {
		fmt.Println("visiting " + url)
	}

	// use the same crawler to visit next page if there is one
	c.OnHTML("a", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Next") && e.Attr("rel") == "nofollow" {
			if strings.Contains(e.Attr("href"), "/network/dependents") {
				if DEBUG {
					fmt.Println("visiting dependents subpage: " + e.Attr("href"))
				}
				random_delay_ms(500)
				last_visited = e.Attr("href")
				c.Visit(e.Attr("href"))
				// next_page = e.Attr("href")
			}
		} else if e.Attr("data-hovercard-type") == "repository" && e.Attr("class") == "text-bold" {
			to_insert = append(to_insert, DependRelation{
				DependencyIdentifier: url_to_identifier(url),
				PackageIdentifier:    package_name,
				DependentIdentifier:  url_to_identifier(e.Attr("href")),
			})

		}
	})

	star_cnt := make([]int, 0)
	fork_cnt := make([]int, 0)

	// get stars count and fork count
	c.OnHTML("span", func(e *colly.HTMLElement) {
		if e.ChildAttr("svg", "class") == "octicon octicon-star" {
			star_count_string := strings.Trim(e.Text, " \t\n\r")
			count, err := strconv.Atoi(strings.ReplaceAll(star_count_string, ",", ""))
			if err != nil {
				fmt.Fprintln(os.Stderr, star_count_string+" is not a valid star count!")
			}
			star_cnt = append(star_cnt, count)
		} else if e.ChildAttr("svg", "class") == "octicon octicon-repo-forked" {
			fork_count_string := strings.Trim(e.Text, " \t\n\r")
			count, err := strconv.Atoi(strings.ReplaceAll(fork_count_string, ",", ""))
			if err != nil {
				fmt.Fprintln(os.Stderr, fork_count_string+" is not a valid fork count!")
			}
			fork_cnt = append(fork_cnt, count)
		}
	})

	// if the `Next` button goes gray, all the dependents are collected
	c.OnHTML("button", func(e *colly.HTMLElement) {
		if e.Attr("disabled") == "disabled" && e.Text == "Next" {
			finished = true
		}
	})

	// if no dependents, the info will show, but not a disabled next button
	c.OnHTML("p", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "We haven’t found any dependents for this repository yet.") || strings.Contains(e.Text, "keep looking!") {
			finished = true
			if DEBUG {
				fmt.Println("visiting package with no dependents: " + url)
			}
		}
	})

	c.Visit(url)

	// redundant check
	// have no error in the 8 billion depend relations
	// but who knows when bugs will occur
	if len(fork_cnt) != len(to_insert) {
		fmt.Fprintf(os.Stderr, "fork count and to insert count don't match!\n forkcount: %d\n starcount: %d\n recordcount: %d\n", len(fork_cnt), len(star_cnt), len(to_insert))
		fmt.Println(to_insert)
	}

	for i := 0; i < len(fork_cnt); i++ {
		to_insert[i].ForkCount = uint32(fork_cnt[i])
		to_insert[i].StarCount = uint32(star_cnt[i])
	}

	if !finished {
		if DEBUG {
			// fmt.Println("crawl " + repo_id + " failed while accessing " + last_visited)
			// if next_page == "" {
			fmt.Println("crawl " + repo_id + " failed while accessing " + last_visited)
			// }
			// fmt.Println("update last_visited to " + last_visited)
		}
		old_checked := make([]CheckedPackage, 0)
		db.Where("repo_identifier = ", repo_id).Find(&old_checked)
		old_count := 0
		if len(old_checked) > 0 {
			if last_visited == old_checked[0].LastVisitedUrl && strings.HasSuffix(last_visited, "/network/dependents") {
				old_count = int(old_checked[0].FailedTimes)
			}
		}
		db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&CheckedPackage{
			RepoIdentifier:    repo_id,
			PackageIdentifier: package_name,
			Finished:          false,
			LastVisitedUrl:    last_visited,
			FailedTimes:       uint32(old_count) + 1,
		})
	} else {
		if DEBUG {
			fmt.Println("crawl " + repo_id + " done with no error")
		}
		db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&CheckedPackage{
			RepoIdentifier:    repo_id,
			PackageIdentifier: package_name,
			Finished:          true,
			LastVisitedUrl:    last_visited,
			FailedTimes:       0,
		})
	}

	if len(to_insert) > 0 {
		db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&to_insert, 100)
	}
}

func RecoverCrawlAll() bool {

	db := model.DB
	to_visit := make([]CheckedPackage, 0)
	db.Where("finished = ? AND failed_times < ?", false, 5).Find(&to_visit)

	for _, pack := range to_visit {
		if DEBUG {
			fmt.Println("recovering " + pack.RepoIdentifier + ": " + pack.PackageIdentifier + " from " + pack.LastVisitedUrl)
		}
		// this means the packages may not be get properly
		// e.g. get |  repo_identifier | package_identifier | finished | last_visited_url       | failed_times |
		//			| actions/checkout | 				    | false    | .../network/dependents | 1			   |
		// but not all the packages
		// so function get_dependents() instead of get_package_dependents() should be applied
		if strings.HasSuffix(pack.LastVisitedUrl, "/network/dependents") {
			check_repo_records := make([]CheckedPackage, 0)
			db.Where("repo_identifier = ?", pack.RepoIdentifier).Find(&check_repo_records)
			// make sure no package of current repo is collected.
			// e.g. get |  repo_identifier | package_identifier | finished | last_visited_url       | failed_times |
			//			| actions/checkout | 				    | false    | .../network/dependents | 1			   |
			// 			| actions/checkout | checkout			| false    | .../something<HASH>    | 1            |
			if len(check_repo_records) == 1 && check_repo_records[0].PackageIdentifier == "" {
				get_dependents(pack.LastVisitedUrl, db, pack.RepoIdentifier)
			}
		} else {
			// all the packages is added to the `checked_packages` table, only need to get package dependents
			get_package_dependents(pack.LastVisitedUrl, pack.PackageIdentifier, db, pack.RepoIdentifier)
		}
		random_delay_ms(1000)
	}

	if len(to_visit) > 0 {
		// get something
		return true
	} else {
		// can get nothing
		return false
	}
}
