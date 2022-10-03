package utils

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TraverseAuthor traverse author's directories
func TraverseAuthor(group *parallelizer.Group, authorDir fs.FileInfo, fn func(*model.Job, *model.Measure)) {
	repoDirList, _ := ioutil.ReadDir(path.Join(config.WORKFLOWS_PATH,
		authorDir.Name()))
	for _, repoDir := range repoDirList {
		if repoDir.IsDir() {
			repoPath := path.Join(config.WORKFLOWS_PATH, authorDir.Name(), repoDir.Name())

			// analyze this repository specifically
			group.Add(func() {
				TraverseRepo(repoPath, fn)
			})
		}
	}
}

func TraverseRepo(repoPath string, fn func(*model.Job, *model.Measure)) {
	measure := model.Measure{
		Name: strings.TrimPrefix(repoPath, config.WORKFLOWS_PATH),
	}
	measure.FetchOrCreate()

	filepath.Walk(repoPath, func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		if err != nil || info.IsDir() || (ext != ".yml" && ext != ".yaml") {
			return err
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}

		TraverseJob(f, &measure, fn)

		return nil
	})
}

func TraverseJob(f *os.File, measure *model.Measure, fn func(*model.Job, *model.Measure)) {
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

		fn(job, measure)
	}
}
