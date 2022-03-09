package analyzer

import (
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
	//runners := utils.TrimRunner(job.RunsOn())

	// FIXME interoperate yaml syntax based on js runtime
	fmt.Println(job.RunsOn())
}
