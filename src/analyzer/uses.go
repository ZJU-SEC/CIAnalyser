package analyzer

import (
	"CIHunter/src/models"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"gorm.io/gorm"
	"os"
	"sort"
	"strings"
)

type GHUse struct {
	ID      uint `gorm:"primaryKey;autoIncrement"`
	GHJobID uint
	GHJob   GHJob `gorm:"foreignKey:GHJobID"`
	Use     string
}

// analyzeUses analyzes how 3ed-party scripts are imported
func analyzeUses(job *Job, ghJob *GHJob) {
	var ghUses []GHUse

	// map result from workflow to measure / uses
	for _, step := range job.Steps {
		if step.Uses != "" {
			ghUses = append(ghUses, GHUse{
				GHJobID: ghJob.ID,
				GHJob:   *ghJob,
				Use:     step.Uses,
			})
		}
	}

	// create ghUses

	models.DB.Create(&ghUses)
}

// analyzePopularNthUses
func analyzePopularNthUses(n int) {
	m := make(map[string]int)

	rows, _ := models.DB.Model(&GHUse{}).Rows()

	for rows.Next() {
		var use GHUse
		models.DB.ScanRows(rows, &use)

		var body string
		if strings.Contains(use.Use, "docker://") { // using docker directly
			// parse the body of image
			body = strings.Split(use.Use, ":")[1][2:]
		} else if strings.Contains(use.Use, "@") { // common gh actions
			body = strings.Split(use.Use, "@")[0]
		} else {
			continue // fast return
		}

		// calculate the usecases frequency
		if val, ok := m[body]; ok {
			m[body] = val + 1
		} else {
			m[body] = 1
		}
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Scripts", "Occurrences", "Coverage"})
	for i := 0; i < n; i++ {
		table.Append([]string{
			fmt.Sprint(ss[i].Key),
			fmt.Sprint(ss[i].Value),
			//fmt.Sprintf("%.2f%", findUsesCoverage(ss[i].Key)*100),
			"none",
		})
	}
	table.Render()
}

// findUsesCoverage
func findUsesCoverage(script string) float64 {
	rows, _ := models.DB.Model(&GHMeasure{}).Rows()
	count := 0
	var totRepos int64 = 0
	models.DB.Model(&GHMeasure{}).Count(&totRepos)

	for rows.Next() {
		var measure GHMeasure
		var jobs []GHJob

		// retrieve jobs according to each measure
		models.DB.ScanRows(rows, &measure)
		models.DB.Where("gh_measure_id = ?", measure.ID).Find(&jobs)

		for _, j := range jobs {
			err := models.DB.Where("gh_job_id = ? AND use LIKE ?", j.ID, script+"%").First(&GHUse{}).Error

			if err != gorm.ErrRecordNotFound {
				count++
				break
			}
		}
	}

	fmt.Println("FUCK")
	return float64(count) / float64(totRepos)
}
