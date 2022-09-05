package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/shomali11/parallelizer"
)

func Clone() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", false).Rows()

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		group.Add(func() error {
			s := s
			cloneScript(&s)
			return nil
		})
	}

	group.Wait()
}

func cloneScript(script *Script) {
	if _, err := os.Stat(script.LocalPath()); !os.IsNotExist(err) {
		script.Check()
		return
	}

	if _, err := git.PlainClone(script.LocalPath(), false, &git.CloneOptions{
		URL: script.GitURL(),
	}); err != nil {
		switch err {
		//case transport.ErrEmptyRemoteRepository, transport.ErrAuthenticationRequired:
		//	script.Delete()
		default:
			fmt.Println("[ERR] cannot clone", script.SrcRef(), err)
		}
		os.RemoveAll(script.LocalPath())
		return
	}
	script.Check()
}
