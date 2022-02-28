package analyzer

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GHUse struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID uint
	GHMeasure   GHMeasure `gorm:"foreignKey:GHMeasureID"`
	Use         string
}

func analyzeUses(f *os.File, measure *GHMeasure) {
	dec := yaml.NewDecoder(f)
	w := Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

	var ghUses []GHUse

	// map result from workflow to measure / uses
	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			if step.Uses != "" {
				ghUses = append(ghUses, GHUse{
					GHMeasureID: measure.ID,
					GHMeasure:   *measure,
					Use:         step.Uses,
				})
			}
		}
	}
}
