package analyzer

import (
	"CIHunter/src/models"
	"fmt"
	"strings"
)

type GHCredential struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	GHJobID    uint
	Job        GHJob `gorm:"foreignKey:GHJobID"`
	Credential string
}

// analyzeCredentials search for envs and get the possible credential usages
func analyzeCredentials(job *Job, ghJob *GHJob) {
	for _, s := range job.Steps {
		envs := s.GetEnv()

		if envs == nil {
			continue // skip empty strings
		}

		for _, e := range envs {
			if strings.Contains(e, "secrets.") {
				models.DB.Create(&GHCredential{
					GHJobID:    ghJob.ID,
					Job:        *ghJob,
					Credential: e,
				})
			}
		}
	}
}

// outputCredentials
func outputCredentials() {
	fmt.Println("\n[Credentials]")

	// id of the row may not be sequential
	// first get the maximum id
	// remove the difference between MAX() and COUNT()
	var ghJobMaxID, ghJobCount int64
	models.DB.Model(&GHJob{}).Count(&ghJobCount)
	models.DB.Model(&GHJob{}).Select("MAX(id)").Row().Scan(&ghJobMaxID)
	ghJobCredentialCount := make([]int, ghJobMaxID+1)

	// scan to rows
	rows, _ := models.DB.Model(&GHCredential{}).Rows()
	for rows.Next() {
		var c GHCredential
		models.DB.ScanRows(rows, &c)

		ghJobCredentialCount[c.GHJobID]++
	}

	//findMax := func(s []int) int {
	//	max := 0
	//	for _, i := range s {
	//		if max < i {
	//			max = i
	//		}
	//	}
	//	return max
	//}

	//maxValue := findMax(ghJobCredentialCount) // get the maximum value
	maxValue := 5
	ghCredentialMetric := make([]int, maxValue+1)

	for _, count := range ghJobCredentialCount {
		if count > maxValue {
			ghCredentialMetric[maxValue]++
		} else {
			ghCredentialMetric[count]++
		}
	}

	// normalize the non-sequential ID & count
	ghCredentialMetric[0] -= int(ghJobMaxID - ghJobCount + 1)

	for credentialCount, jobCount := range ghCredentialMetric {
		if credentialCount != maxValue {
			fmt.Printf("%d:\t%d\n", credentialCount, jobCount)
		} else {
			fmt.Printf(">= %d:\t%d\n", credentialCount, jobCount)
		}
	}
}
