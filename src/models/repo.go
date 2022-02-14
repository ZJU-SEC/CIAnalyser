package models

import (
	"CIHunter/src/config"
	"path"
	"strings"
	"time"
)

// Repo schema for repo's metadata
type Repo struct {
	ID        uint   `gorm:"primaryKey"`
	Ref       string `gorm:"uniqueIndex"`
	UpdatedAt time.Time

	// source of the repo
	Source []byte `gorm:"type:bytea"`

	ContributorCount uint
	CommitCount      uint
	WatchCount       uint
	ForkCount        uint
	StarCount        uint
}

func (r *Repo) Name() string {
	return strings.ReplaceAll(r.Ref[1:], "/", ":")
}

func (r *Repo) LocalPath() string {
	return path.Join(config.DEV_SHM, r.Name())
}

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
