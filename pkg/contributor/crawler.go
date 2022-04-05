package contributor

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
	"CIHunter/pkg/script"
	"CIHunter/utils"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
)

func Crawl() {
	err := model.DB.AutoMigrate(Contributor{}, Contribution{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, _ := model.DB.Model(&script.Script{}).Rows()
	for rows.Next() {
		var s script.Script
		model.DB.ScanRows(rows, &s)

		group.Add(func() {
			s := s
			crawl(&s)
		})
	}
	group.Wait()
}

func crawl(s *script.Script) {
	c := utils.CommonCollector()

	var contributors []Contributor

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Authorization", "token "+utils.RequestGitHubToken())
	})

	c.OnResponse(func(r *colly.Response) {
		err := json.Unmarshal(r.Body, &contributors)
		if err != nil {
			return
		}

		for i := 0; i < len(contributors); i++ {
			contributors[i].fetchOrCreate()

			contribution := Contribution{
				ContributorID: contributors[i].ID,
				Contributor:   contributors[i],
				ScriptID:      s.ID,
				Script:        *s,
			}
			contribution.create()
		}
	})

	c.Visit(fmt.Sprintf("https://api.github.com/repos/%s/contributors", s.SrcRef()))
}

//import (
//	"CIHunter/pkg/model"
//	"CIHunter/utils"
//	"fmt"
//	"github.com/gocolly/colly"
//	"strconv"
//	"strings"
//	"time"
//)
//
//package maintainers
//
//import (
//"CIHunter/pkg/model"
//"CIHunter/utils"
//"fmt"
//"github.com/gocolly/colly"
//"strconv"
//"strings"
//"time"
//)
//
