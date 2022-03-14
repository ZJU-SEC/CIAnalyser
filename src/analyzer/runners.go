package analyzer

import (
	"CIHunter/src/models"
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

func outputRunners() {
	var totRunners int64
	var c int64
	models.DB.Model(&GHRunner{}).Count(&totRunners)
	fmt.Printf("Total occurrences of runners: %d\n", totRunners)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "ubuntu%").Count(&c)
	fmt.Printf("Occurrences of ubuntu runners: %d, Ratio: %.2f%%\n", c, float64(c)/float64(totRunners)*100)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND (runner ILIKE ? OR runner ILIKE ?)",
		false, "ubuntu-20.04", "ubuntu-latest").Count(&c)
	fmt.Printf("\tubuntu-20.04: %d\n", c)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "ubuntu-18.04").Count(&c)
	fmt.Printf("\tubuntu-18.04: %d\n", c)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "ubuntu-16.04").Count(&c)
	fmt.Printf("\tubuntu-16.04: %d\n", c)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "macos%").Count(&c)
	fmt.Printf("Occurrences of macos runners: %d, Ratio: %.2f%%\n", c, float64(c)/float64(totRunners)*100)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND (runner ILIKE ? OR runner ILIKE ?)",
		false, "macos-11%", "macos-latest").Count(&c)
	fmt.Printf("\tmacos-11: %d\n", c)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "macos-10.15%").Count(&c)
	fmt.Printf("\tmacos-10.15: %d\n", c)

	models.DB.Model(&GHRunner{}).Where("self_hosted = ? AND runner ILIKE ?", false, "windows%").Count(&c)
	fmt.Printf("Occurrences of windows runners: %d, Ratio: %.2f%%\n", c, float64(c)/float64(totRunners)*100)
	models.DB.Model(&GHRunner{}).Where("self_hosted = ?", true).Count(&c)
	fmt.Printf("Occurrences of self-hosted runners: %d, Ratio: %.2f%%\n", c, float64(c)/float64(totRunners)*100)
	fmt.Printf("Occurrences of self-hosted runners: %d, Ratio: %.2f%%\n", c, float64(c)/float64(totRunners)*100)
}
