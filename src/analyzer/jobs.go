package analyzer

import (
	"CIHunter/src/models"
	"gopkg.in/yaml.v3"
	"os"
)

// analyzeJobs analyze jobs and their underlying attributes, runners, uses, ...
func analyzeJobs(f *os.File, measure *GHMeasure) {
	dec := yaml.NewDecoder(f)
	w := Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

	// map result from workflow to measure / uses
	for _, job := range w.Jobs {
		// create measure record
		ghJob := GHJob{
			GHMeasureID: measure.ID,
			GHMeasure:   *measure,
		}
		models.DB.Create(&ghJob)

		analyzeUses(job, &ghJob)
		analyzeRunners(job, &ghJob)
	}
}
