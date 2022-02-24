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
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
)

func CrawlActions() {
	os.RemoveAll(config.REPOS_PATH)

	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	// get database iterator
	rows, err := models.DB.Model(&models.Repo{}).Where("checked = ?", false).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var notCheckedCount int64
	models.DB.Model(&models.Repo{}).Where("checked = ?", false).Count(&notCheckedCount)
	rand.Seed(time.Now().UnixNano())
	randomSkip := rand.Intn(int(notCheckedCount))
	fmt.Println("randomly skip", randomSkip, "rows")

	count := 0
	for rows.Next() && count < randomSkip {
		count++
	}

	fmt.Println("start processing ...")
	time.Sleep(1 * time.Second)
	count = 0
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
}

// analyze the repository
func downloadRepo(repo *models.Repo) {
	c := make(chan error, 1)

	// clone worker
	go func() {
		err := clone(repo)
		c <- err
	}()

	select {
	case res := <-c:
		if res != nil {
			os.RemoveAll(repo.LocalPath())
			return
		}
	case <-time.After(time.Duration(config.TIMEOUT) * time.Second):
		fmt.Println("- skipped", repo.Ref)
		os.RemoveAll(repo.LocalPath())
		adjustTimeout()
		return
	}

	if utils.DirExists(repo.WorkflowsPath()) {
		copy.Copy(repo.WorkflowsPath(), path.Join(config.WORKFLOWS_PATH, repo.Ref[1:]))
	}
	os.RemoveAll(repo.LocalPath())
	repo.Check()
}

func clone(repo *models.Repo) error {
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
		return err
	}

	return nil
}

var timeoutCount = 0

func adjustTimeout() {
	var mutex sync.Mutex
	mutex.Lock()

	if timeoutCount > config.TIMEOUT_THRESHOLD {
		config.TIMEOUT *= 2
		timeoutCount = 0
	} else {
		timeoutCount++
	}

	mutex.Unlock()
}
