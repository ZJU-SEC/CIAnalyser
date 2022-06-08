package router

var DATABASE_PATH = "test.db"

type OfficialAction struct {
	Name        string
	Author      string
	Category    string
	NumStars    uint
	Description string
	Url         string `gorm:"uniqueIndex:url,sort:desc"`
}

type AllAction struct {
	Identifier string `gorm:"uniqueIndex:identifier,sort:desc"`
	IsOfficial bool
	Checked    bool
	Url        string
}

type DependRelation struct {
	DependencyIdentifier string
	PackageIdentifier    string
	DependentIdentifier  string
}
