package credential

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
	"CIHunter/utils"
	"github.com/shomali11/parallelizer"
	"io/ioutil"
	"strings"
)

func Extract() {
	err := model.DB.AutoMigrate(Credential{}, model.Measure{})
	if err != nil {
		panic(err)
	}

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	authorDirList, _ := ioutil.ReadDir(config.WORKFLOWS_PATH)
	for _, authorDir := range authorDirList {
		if !authorDir.IsDir() {
			continue // not dir, skip
		}

		utils.TraverseAuthor(group, authorDir, analyzeCredential)
	}

	group.Wait()
}

func analyzeCredential(job *model.Job, measure *model.Measure) {
	for _, s := range job.Steps {
		envs := s.GetEnv()

		if envs == nil {
			continue // skip empty strings
		}

		for _, e := range envs {
			if strings.Contains(e, "secrets.") {
				c := Credential{
					MeasureID:  measure.ID,
					Measure:    *measure,
					Credential: e,
				}
				c.Create()
			}
		}
	}
}
