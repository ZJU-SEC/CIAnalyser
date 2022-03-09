package analyzer

import (
	"fmt"
	"strings"
)

type GHRunner struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	GHJobID    uint
	Job        GHJob `gorm:"foreignKey:GHJobID"`
	Runner     string
	SelfHosted bool
}

func analyzeRunners(job *Job, ghJob *GHJob) {
	runners := TrimRunner(job.RunsOn())
	fmt.Println(runners)
	// TODO create runner records in database
}

// TrimRunner first regulars the runner labels, then removes the duplicate runners
func TrimRunner(runners []string) []string {
	var finalRunners []string

	contains := func(s []string, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}

	trimDuplicate := func(s []string) []string {
		keys := make(map[string]bool)
		res := []string{}
		for _, e := range s {
			if _, v := keys[e]; !v {
				keys[e] = true
				res = append(res, e)
			}
		}

		return res
	}

	latestMapping := map[string]string{
		"ubuntu-latest":  "ubuntu-20.04",
		"macos-latest":   "macos-11",
		"windows-latest": "windows-2019",
	}

	for _, runner := range runners {
		// lowercase the macOS runner
		if runner == "macOS-11" || runner == "macOS-10.15" || runner == "macOS-latest" {
			runner = strings.ToLower(runner)
		}

		if val, ok := latestMapping[runner]; ok {
			runner = val
		}

		if !contains(finalRunners, runner) {
			finalRunners = append(finalRunners, runner)
		}
	}

	return trimDuplicate(finalRunners)
}
