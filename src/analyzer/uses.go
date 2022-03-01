package analyzer

type GHUse struct {
	ID      uint `gorm:"primaryKey;autoIncrement"`
	GHJobID uint
	GHJob   GHJob `gorm:"foreignKey:GHJobID"`
	Use     string
}

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

	// TODO create ghUses
}
