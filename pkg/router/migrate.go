package router

import (
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/repo"

	"gorm.io/gorm/clause"
)

func RelationsToRepos() {
	db := model.DB
	db.AutoMigrate(&repo.Repo{})
	to_insert := make([]DependRelation, 0)
	db.Distinct("dependent_identifier").Find(&to_insert)
	for _, relation := range to_insert {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&repo.Repo{
			Ref: relation.DependentIdentifier,
			// Checked: false,
		})
	}
}
