package analyzer

import (
	"CIHunter/src/utils"
	"fmt"
	"syscall"
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

	// FIXME
	if len(runners) == 0 {
		syscall.Pause()
		fmt.Println(job.RunsOn())
	}
	fmt.Println(runners)
}
