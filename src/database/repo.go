package database

import "time"

// Repo schema for repo's metadata
type Repo struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Author    string
	UpdatedAt time.Time

	ContributorCount uint
	CommitCount      uint
	WatchCount       uint
	ForkCount        uint
	StarCount        uint
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
}
