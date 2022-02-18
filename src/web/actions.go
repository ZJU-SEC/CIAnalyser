package web

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/otiai10/copy"
	"github.com/shomali11/parallelizer"
	"os"
	"path"
)

func CrawlActions() int {
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

	var count = 0
	for rows.Next() && count < config.BATCH_SIZE {
		var repo models.Repo
		models.DB.ScanRows(rows, &repo)

		if !repo.Checked {
			group.Add(func() {
				downloadRepo(&repo)
			})
			count++
		}
	}

	group.Wait()

	if count == 0 {
		return 0
	} else {
		return 1
	}
}

// analyze the repository
func downloadRepo(repo *models.Repo) {
	if _, err := git.PlainClone(repo.LocalPath(), false, &git.CloneOptions{
		URL:   repo.GitURL(),
		Depth: 1,
	}); err != nil {
		switch err {
		case transport.ErrEmptyRemoteRepository, transport.ErrAuthenticationRequired:
			repo.Delete()
		default:
			fmt.Println("[ERR] cannot clone", repo.Ref, err)
		}
		return
	}

	if utils.DirExists(repo.WorkflowsPath()) {
		copy.Copy(repo.WorkflowsPath(), path.Join(config.WORKFLOWS_PATH, repo.Ref[1:]))
	}
	os.RemoveAll(repo.LocalPath())
	repo.Check()
}
