package analyzer

import (
	"CIHunter/src/utils"
	"fmt"
)

type GHRunner struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	GHJobID    uint
	Job        GHJob `gorm:"foreignKey:GHJobID"`
	Runner     string
	SelfHosted bool
}

func analyzeRunners(job *Job, ghJob *GHJob) {
	runners := utils.TrimRunner(job.RunsOn())

	// FIXME interoperate yaml syntax based on js runtime
	if len(runners) == 0 {
		fmt.Println(job.RunsOn())
	}
}
