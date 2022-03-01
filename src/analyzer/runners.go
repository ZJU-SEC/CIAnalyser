package analyzer

import (
	"CIHunter/src/utils"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"syscall"
)

type GHRunner struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	GHJobID    uint
	Job        GHJob `gorm:"foreignKey:GHJobID"`
	Runner     string
	SelfHosted bool
}

func analyzeRunners(f *os.File, measure *GHMeasure) {
	dec := yaml.NewDecoder(f)
	w := Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

	// map result from workflow to measure / uses
	for _, job := range w.Jobs {
		runners := utils.TrimRunner(job.RunsOn())

		// FIXME
		if len(runners) == 0 {
			syscall.Pause()
			fmt.Println(job.RunsOn())
		}
		fmt.Println(runners)

		//for _, step := range job.Steps { // traverse `uses` item, if not empty, record
		//	if step.Uses != "" {
		//		ghRunners = append(ghRunners, GHRunner{
		//			GHMeasureID: measure.ID,
		//			GHMeasure:   *measure,
		//			Runner:      job.RawRunsOn,
		//		})
		//	}
		//}
	}
}
