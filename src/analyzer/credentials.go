package analyzer

import (
	"CIHunter/src/models"
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
