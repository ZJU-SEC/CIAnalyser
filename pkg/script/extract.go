package script

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Extract() {
	err := model.DB.AutoMigrate(Script{}, Usage{}, Measure{})
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

		traverseAuthor(group, authorDir)
	}

	group.Wait()
}

// traverseAuthor traverse author's directories
func traverseAuthor(group *parallelizer.Group, authorDir fs.FileInfo) {
	repoDirList, _ := ioutil.ReadDir(path.Join(config.WORKFLOWS_PATH,
		authorDir.Name()))
	for _, repoDir := range repoDirList {
		if repoDir.IsDir() {
			repoPath := path.Join(config.WORKFLOWS_PATH, authorDir.Name(), repoDir.Name())

			// analyze this repository specifically
			group.Add(func() {
				analyzeRepo(repoPath)
			})
		}
	}
}

func analyzeRepo(repoPath string) {
	measure := Measure{
		Name: strings.TrimPrefix(repoPath, config.WORKFLOWS_PATH),
	}
	measure.create()

	filepath.Walk(repoPath, func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		if err != nil || info.IsDir() || (ext != ".yml" && ext != ".yaml") {
			return err
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}

		analyzeJobs(f, &measure)

		return nil
	})
}

func analyzeJobs(f *os.File, measure *Measure) {
	dec := yaml.NewDecoder(f)
	w := model.Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

	// map result from workflow to measure / uses
	for _, job := range w.Jobs {
		if job == nil {
			continue // skip null jobs
		}

		analyzeUses(job, measure)
	}
}

// analyzeUses analyzes how 3rd-party scripts are imported
func analyzeUses(job *model.Job, measure *Measure) {
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
		script.Ref = strings.Split(step.Uses, "@")[0]
		script.Maintainer = strings.Split(script.Ref, "/")[0]
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
