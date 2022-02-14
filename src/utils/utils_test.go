package utils

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"github.com/go-git/go-git/v5"
	"os"
	"testing"
)

func TestDeserializeRepo(t *testing.T) {
	config.Init()
	models.Init()

	var repos []models.Repo
	models.DB.Where("source IS NOT NULL").Find(&repos)
	for _, repo := range repos {
		// deserialize tarball from database
		if err := DeserializeRepo(repo.Source); err != nil {
			os.RemoveAll(repo.LocalPath())
			t.Errorf("deserialize %s ended with error, %s", repo.Name(), err)
		}
		if _, err := git.PlainOpen(repo.LocalPath()); err != nil {
			os.RemoveAll(repo.LocalPath())
			t.Errorf("open %s as git repo ended with error, %s", repo.Name(), err)
		}
		os.RemoveAll(repo.LocalPath())
	}
}
