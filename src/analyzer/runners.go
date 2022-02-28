package analyzer

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type GHRunner struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID uint
	GHMeasure   GHMeasure `gorm:"foreignKey:GHMeasureID"`
	Runner      string
}

func analyzeRunners(f *os.File, measure *GHMeasure) {
	dec := yaml.NewDecoder(f)
	w := Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

	// map result from workflow to measure / uses
	for _, job := range w.Jobs {
		//if job.Matrix() != nil {
		//	fmt.Println(job.Matrix())
		//}
		fmt.Println(job.RunsOn())
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
