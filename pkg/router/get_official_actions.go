package router

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func get_official_all() {
	db, err := gorm.Open(sqlite.Open(DATABASE_PATH), &gorm.Config{})
	if err != nil {
		panic("fail to connect to database")
	}
	db.AutoMigrate(&OfficialAction{})

	for _, char := range ALLCHARS {
		get_official_with_kerword(string(char), db)
	}
}

func get_official_with_kerword(keyword string, db *gorm.DB) {
	number_of_result := -1
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

			to_visit = append(to_visit, OfficialAction{
				Name:        valid_attributes[1],
				Category:    valid_attributes[0],
				Author:      valid_attributes[2][2:],
				Description: valid_attributes[3],
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

	if number_of_result <= 0 {
		return
	}

	number_of_pages := (number_of_result-1)/20 + 1

	if number_of_result <= 1000 {
		fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(number_of_result) + " results, collecting")
		for i := 2; i <= number_of_pages; i++ {
			c.Visit("https://github.com/marketplace?type=actions&query=" + keyword + "&page=" + strconv.Itoa(i))
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(to_visit)
	} else {
		fmt.Println("Keyword " + keyword + " have " + strconv.Itoa(number_of_result) + " results, refining")
		for _, char := range ALLCHARS {
			get_official_with_kerword(keyword+string(char), db)
		}
	}
}
