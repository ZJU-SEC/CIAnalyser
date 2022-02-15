package models

import (
	"fmt"
	"gorm.io/gorm"
	"sync"
)

// GitHubActionMeasure measurement of the GitHub Actions
type GitHubActionMeasure struct {
	ID     uint `gorm:"primaryKey"`
	RepoID uint
	Repo   Repo `gorm:"foreignKey:RepoID"`

	HasPermission       bool
	HasSelfHosted       bool
	HasSecrets          bool
	HasServiceContainer bool
}

// GitHubActionUses
type GitHubActionUses struct {
	ID                    uint `gorm:"primaryKey"`
	GitHubActionMeasureID uint
	GitHubActionMeasure   GitHubActionMeasure `gorm:"foreignKey:GitHubActionMeasureID"`
	Uses                  string
}

func (measure *GitHubActionMeasure) Create(repo *Repo, ghUses []GitHubActionUses) {
	var mutex sync.Mutex
	mutex.Lock()

	if err := DB.Transaction(func(tx *gorm.DB) error {
		measure.RepoID = repo.ID
		measure.Repo = *repo

		if err := tx.Create(measure).Error; err != nil {
			return err
		}

		// create uses data
		for i := 0; i < len(ghUses); i++ {
			ghUses[i].GitHubActionMeasureID = measure.ID
			ghUses[i].GitHubActionMeasure = *measure
		}
		if err := tx.Create(&ghUses).Error; err != nil {
			return err
		}
		return nil // commit the whole transaction
	}); err != nil {
		fmt.Println("[ERR] cannot create", repo.Ref, err)
	}

	mutex.Unlock()
}
