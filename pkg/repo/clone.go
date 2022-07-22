package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/storage/memory"
	"math/rand"

	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/shomali11/parallelizer"
)

func Clone() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	// get database iterator
	rows, err := model.DB.Model(&Repo{}).Where("cloned = ?", false).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var notCheckedCount int64
	var randomSkip int

	model.DB.Model(&Repo{}).Where("checked = ?", false).Count(&notCheckedCount)
	rand.Seed(time.Now().UnixNano())
	if notCheckedCount == 0 {
		randomSkip = 0
	} else {
		randomSkip = rand.Intn(int(notCheckedCount))
	}
	fmt.Println("randomly skip", randomSkip, "rows")

	count := 0
	for rows.Next() && count < randomSkip {
		count++
	}

	fmt.Println("start processing ...")
	time.Sleep(1 * time.Second)
	count = 0
	for rows.Next() && count < config.BATCH_SIZE {
		var repo Repo
		model.DB.ScanRows(rows, &repo)

		if !repo.Cloned {
			group.Add(func() {
				downloadRepo(&repo)
			})
			count++
		}
	}

	group.Wait()
}

// analyze the repository
func downloadRepo(repo *Repo) {
	c := make(chan error, 1)

	// clone worker
	go func() {
		err := clone(repo)
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			return
		}
	case <-time.After(time.Duration(config.TIMEOUT) * time.Second):
		fmt.Println("- skipped", repo.Ref)
		adjustTimeout()
		return
	}

	repo.Check()
}

func clone(repo *Repo) error {
	fs := memfs.New()

	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: repo.GitURL(),
	}); err != nil {
		if !(err == transport.ErrEmptyRemoteRepository ||
			err == transport.ErrAuthenticationRequired) {
			fmt.Println("[ERR] cannot clone", repo.Ref, err)
		}
		return err
	}

	// TODO add operations

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
