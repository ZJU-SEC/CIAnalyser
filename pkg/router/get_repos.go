package router

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func get_official_repos() {
	// to_insert := make([]AllAction, 0)
	db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Database failed at get official repos stage")
	}
	db.AutoMigrate(&AllAction{})
	var official_actions []OfficialAction
	db.Find(&official_actions)
	c := colly.NewCollector()
	c.OnHTML("aside", func(e *colly.HTMLElement) {
		all_links := e.ChildAttrs("a", "href")
		right_link := ""

		for _, link := range all_links {
			if strings.HasPrefix(link, "/login?") {
				right_link = link
				break
			}
		}

		if right_link != "" {
			identifier := strings.ReplaceAll(right_link[20:], "%2F", "/")
			db.Clauses(clause.OnConflict{DoNothing: true}).Create(&AllAction{
				Identifier: identifier,
				IsOfficial: true,
				Checked:    false,
				Url:        "https://github.com/" + identifier,
			})
			fmt.Println(identifier)
		} else {
			fmt.Fprintf(os.Stderr, "Cannot get repository from %s\n", strings.Join(all_links, ", "))
		}
	})
	for _, action := range official_actions {
		c.Visit(action.Url)
	}
	// db.Create(&to_insert)
}

func get_dependents_repos_all() {
	db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Database failed at get dependents stage")
	}

	db.AutoMigrate(&DependRelation{})
	var to_get []AllAction
	db.Where("checked = ?", false).Find(&to_get)
	// for ; len(to_get) > 0; db.Where("checked = ?", false).Find(&to_get) {
	// 	for _, record := range to_get {
	// 		get_dependents(record.Url+"/network/dependents", db)
	// 	}
	// }
	for _, record := range to_get {
		get_dependents(record.Url+"/network/dependents", db)
		// to update checked value
	}
	// fmt.Println(to_get)
}

func get_dependents(url string, db *gorm.DB) {
	c := colly.NewCollector()
	to_visit := make(map[string](string), 2)
	c.OnHTML("details-menu", func(e *colly.HTMLElement) {
		is_package_list := false
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
				get_package_dependents("https://github.com"+package_path, package_name, db)
			}
		} else {
			fmt.Println("visiting repository dependents " + url)
			get_package_dependents(url, "", db)
		}
	})
	c.Visit(url)
}

func get_package_dependents(url string, package_name string, db *gorm.DB) {

}
