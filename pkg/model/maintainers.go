package model

type Maintainer struct {
	ID       uint `gorm:"primaryKey;autoIncrement;"`
	Name     string
	Verified bool `gorm:"default:false"`
}
