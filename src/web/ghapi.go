package web

import (
	"CIHunter/src/config"
	"github.com/shomali11/parallelizer"
)

func CrawlGHAPI() {
	group := parallelizer.NewGroup(
		parallelizer.WithPoolSize(config.WORKER),
		parallelizer.WithJobQueueSize(config.QUEUE_SIZE),
	)
	defer group.Close()

	group.Wait()
}
