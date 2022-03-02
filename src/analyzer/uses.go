package analyzer

import "CIHunter/src/models"

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
