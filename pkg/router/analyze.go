package router

import (
	"fmt"
	// "os"
	"strings"

	"github.com/gocolly/colly"
	// "gorm.io/driver/sqlite"
	"CIAnalyser/pkg/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func update_related_repos() {
	// db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "fail to open database at update related repo stage")
	// }
	db := model.DB
	db.AutoMigrate(&ActionRelatedRepository{})
	all_relations := make([]DependRelation, 0)
	db.Select("dependent_identifier").Find(&all_relations)
	to_insert := make([]ActionRelatedRepository, 0)
	for _, relation := range all_relations {
		to_insert = append(to_insert, ActionRelatedRepository{
			Identifier: relation.DependentIdentifier,
			Analyzed:   false,
		})
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&to_insert, 100)
}

func analyze_all() uint64 {
	var ret uint64 = 0
	// db, err := gorm.Open(sqlite.Open(DATABASE_PATH))
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "fail to open database at analyze all stage")
	// }
	db := model.DB
	db.AutoMigrate(&RawDependRelation{})
	unchecked_repos := make([]ActionRelatedRepository, 0)
	db.Find(&unchecked_repos)
	c := colly.NewCollector()
	c.OnHTML("a", func(e *colly.HTMLElement) {
		if strings.HasSuffix(e.Text, ".yaml") || strings.HasSuffix(e.Text, ".yml") {
			// fmt.Println(e.Attr("href"))
			get_repo_dependencies("https://github.com"+e.Attr("href"), db)
		}
	})
	for _, repo := range unchecked_repos {
		c.Visit("https://github.com/" + repo.Identifier + "/tree/master/.github/workflows")
		db.Model(&ActionRelatedRepository{}).Where("identifier = ?", repo.Identifier).Update("analyzed", true)
		ret++
	}

	return ret
}

func get_repo_dependencies(url string, db *gorm.DB) {
	repo_id := url_to_identifier(url)
	c := colly.NewCollector()
	to_insert := make([]RawDependRelation, 0)
	repos_fount := make([]AllAction, 0)

	c.OnHTML("td", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "uses") {
			id_with_tag := strings.Trim(strings.TrimPrefix(strings.Trim(e.Text, " \t-#"), "uses: "), "\"'")
			at_pos := strings.Index(id_with_tag, "@")
			if at_pos != -1 {
				dependency_id := id_with_tag[:at_pos]
				dependency_tag := id_with_tag[at_pos+1:]
				to_insert = append(to_insert, RawDependRelation{
					ScriptIdentifier:    dependency_id,
					TagIdentifier:       dependency_tag,
					DependentIdentifier: repo_id,
				})

				if DEBUG {
					fmt.Println("find use: " + dependency_id + "@" + dependency_tag)
				}

				if is_repo_identifier(dependency_id) {
					repos_fount = append(repos_fount, AllAction{
						Identifier: dependency_id,
						IsOfficial: false,
						Checked:    false,
						Url:        "https://github.com/" + dependency_id,
					})
					if DEBUG {
						fmt.Println("The URL is https://github.com/" + dependency_id)
					}
				}
			}
		}
	})
	c.Visit(url)
	if len(to_insert) > 0 {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&to_insert)
	}
	if len(repos_fount) > 0 {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&repos_fount)
	}
}

func is_repo_identifier(candidate string) bool {
	ret := true
	if strings.Count(candidate, "/") != 1 {
		ret = false
	}
	return ret
}
