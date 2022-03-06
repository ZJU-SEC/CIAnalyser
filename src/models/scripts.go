package models

// Script schema for script's metadata
type Script struct {
	ID      uint `gorm:"primaryKey;autoIncrement;"`
	Name    string
	Ref     string
	Checked bool `gorm:"default:false"`
}
