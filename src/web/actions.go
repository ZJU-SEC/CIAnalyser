package web

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

func CrawlActions() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, err := models.DB.Model(&models.Repo{}).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var repo models.Repo
		models.DB.ScanRows(rows, &repo)

		if !repo.Checked {
			group.Add(func() {
				analyzeRepo(&repo)
			})
		}
	}

	group.Wait()
}

// analyze the repository
func analyzeRepo(repo *models.Repo) {
	fs := memfs.New()
	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   "https://github.com" + repo.Ref + ".git",
		Depth: 1,
	}); err != nil {
		fmt.Println("[ERR] cannot clone", repo.Ref, err)
		return
	}

	// declare variables
	var ghMeasure *models.GitHubActionMeasure = nil
	var ghUses []models.GitHubActionUses

	// analyze workflows
	workflows, _ := fs.ReadDir(".github/workflows")
	for _, entry := range workflows {
		// if dir / not yml / not yaml skip
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}

		// open yaml configuration file
		yml, err := fs.Open(".github/workflows/" + entry.Name())
		if err != nil {
			continue
		}

		// parse workflow from yaml file
		w := models.Workflow{}
		if err := yaml.NewDecoder(yml).Decode(&w); err != nil {
			continue
		}

		// create ghMeasure
		if ghMeasure == nil {
			ghMeasure = new(models.GitHubActionMeasure)
		}

		// map result from workflow to measure / uses
		for _, job := range w.Jobs {
			for _, step := range job.Steps { // traverse `uses` item, if not empty, record
				if step.Uses != "" {
					ghUses = append(ghUses, models.GitHubActionUses{Uses: step.Uses})
				}
			}
		}
	}

	// create measure if using github actions
	if ghMeasure != nil {
		ghMeasure.Create(repo, ghUses)
	}
	repo.Check()
}
