package analyzer

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Author -> Repo -> Jobs

// traverseAuthor traverse author's directories
func traverseAuthor(authorDir fs.FileInfo) {
	countAuthor()

	repoDirList, _ := ioutil.ReadDir(path.Join(config.WORKFLOWS_PATH, authorDir.Name()))
	for _, repoDir := range repoDirList {
		if repoDir.IsDir() {
			countRepo()

			repoPath := path.Join(config.WORKFLOWS_PATH, authorDir.Name(), repoDir.Name())

			// analyze this repository specifically
			analyzeRepo(repoPath)
		}
	}
}

// analyzeRepo glob the given path, check yaml files and process
func analyzeRepo(repoPath string) {
	// create measure record
	measure := GHMeasure{
		RepoRef:            strings.TrimPrefix(repoPath, config.WORKFLOWS_PATH),
		ConfigurationCount: 0,
	}

	models.DB.Create(&measure)

	filepath.Walk(repoPath, func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		if err != nil || info.IsDir() || (ext != ".yml" && ext != ".yaml") {
			return err
		}

		measure.ConfigurationCount++
		models.DB.Save(&measure)

		f, err := os.Open(p)
		if err != nil {
			return err
		}

		analyzeJobs(f, &measure)

		return nil
	})
}

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
		analyzeCredentials(job, &ghJob)
	}
}
