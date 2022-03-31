package script

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
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

		group.Add(func() {
			// TODO clone repo
		})
	}

	group.Wait()
}
