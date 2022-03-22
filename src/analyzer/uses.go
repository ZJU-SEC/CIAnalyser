package analyzer

import (
	"CIHunter/src/models"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"sort"
	"strings"
)

type GHUse struct {
	ID      uint `gorm:"primaryKey;autoIncrement"`
	GHJobID uint
	GHJob   GHJob `gorm:"foreignKey:GHJobID"`
	Use     string
}

// analyzeUses analyzes how 3rd-party scripts are imported
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

// outputPopularNthUses
func outputPopularNthUses(f *excelize.File, n int) {
	fmt.Println("\n[Popular", n, "scripts]")

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

	// calculate total scripts
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for i := 0; i < n; i++ {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), ss[i].Key)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), findReposCountByScript(ss[i].Key))
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+1), findJobsCountByScript(ss[i].Key))
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+1), ss[i].Value)
	}
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

	return float64(count) / float64(totRepos)
}

func findJobsCountByScript(s string) int {
	var c int64
	models.DB.Model(&GHUse{}).Where("use LIKE ?", s+"%").
		Distinct("gh_job_id").Count(&c)
	return int(c)
}

func findReposCountByScript(s string) int {
	var c int64
	models.DB.Model(&GHUse{}).Select("use, gh_measure_id").
		Joins("left join gh_jobs ON gh_uses.gh_job_id = gh_jobs.id").
		Where("use LIKE ?", s+"%").Distinct("gh_measure_id").Count(&c)

	return int(c)
}
