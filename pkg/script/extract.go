package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/utils"
	"github.com/shomali11/parallelizer"
	"io/ioutil"
	"strings"
)

func Extract() {
	if !model.DB.Migrator().HasTable(&Verified{}) {
		panic("does not have the table: `verified`")
	}

	err := model.DB.AutoMigrate(Script{}, Usage{}, model.Measure{})
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

		utils.TraverseAuthor(group, authorDir, analyzeUses)
	}

	group.Wait()
}

// analyzeUses analyzes how 3rd-party scripts are imported
func analyzeUses(job *model.Job, measure *model.Measure) {
	// map result from workflow to measure / uses
	for _, step := range job.Steps {
		if step.Uses == "" ||
			strings.HasPrefix(step.Uses, ".") ||
			strings.HasPrefix(step.Uses, "/") ||
			strings.Contains(step.Uses, "docker:") ||
			!strings.Contains(step.Uses, "@") ||
			len(strings.Split(step.Uses, "/")) < 2 {
			continue
		}

		// record this script
		script := Script{}
		script.OnMarketplace = false
		script.Ref = strings.Split(step.Uses, "@")[0]
		script.Verified = IsVerified(strings.Split(script.Ref, "/")[0])
		script.fetchOrCreate()

		usage := Usage{
			MeasureID: measure.ID,
			Measure:   *measure,
			ScriptID:  script.ID,
			Script:    script,
			Use:       step.Uses,
		}
		usage.create()
	}
}
