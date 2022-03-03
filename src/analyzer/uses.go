package analyzer

import (
	"CIHunter/src/models"
	"fmt"
	"github.com/olekukonko/tablewriter"
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

		// calculate the use frequency
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
			"None",
		})
	}
	table.Render()
}
