package analyzer

import (
	"CIHunter/src/models"
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

	if runners == nil {
		return
	}

	for _, r := range runners {
		models.DB.Create(&GHRunner{
			GHJobID:    ghJob.ID,
			Job:        *ghJob,
			Runner:     r,
			SelfHosted: isSelfHosted(r),
		})
	}
}

// TrimRunner first regulars the runner labels, then removes the duplicate runners
func TrimRunner(runners []string) []string {
	if runners == nil {
		return nil
	}

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

	for _, runner := range runners {
		if runner == "undefined" {
			continue // skip `undefined` runners
		}

		if !contains(finalRunners, runner) {
			finalRunners = append(finalRunners, runner)
		}
	}

	return trimDuplicate(finalRunners)
}

// isSelfHosted checks if a runner is self-hosted
func isSelfHosted(runner string) bool {
	splitLabels := strings.Split(runner, "-")

	if len(splitLabels) != 2 {
		return true
	}

	os := strings.ToLower(splitLabels[0])
	ver := strings.ToLower(splitLabels[1])

	switch os {
	case "ubuntu":
		if ver == "latest" || ver == "20.04" || ver == "18.04" || ver == "16.04" {
			return false
		}
	case "macos":
		if ver == "latest" || ver == "11" || ver == "11.0" || ver == "10.15" {
			return false
		}
	case "windows":
		if ver == "latest" || ver == "2016" || ver == "2019" || ver == "2022" {
			return false
		}
	default:
		return true
	}

	return false
}
