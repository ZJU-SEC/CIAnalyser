package router

type OfficialAction struct {
	Name        string
	Author      string
	Category    string
	NumStars    uint
	Description string
	Url         string `gorm:"primaryKey"`
}

type AllAction struct {
	Identifier string `gorm:"primaryKey"`
	IsOfficial bool
	Checked    bool
	Url        string
}

type DependRelation struct {
	DependencyIdentifier string `gorm:"primaryKey"`
	PackageIdentifier    string `gorm:"primaryKey"`
	DependentIdentifier  string `gorm:"primaryKey"`
	StarCount            uint32
	ForkCount            uint32
}

type ActionRelatedRepository struct {
	Identifier string `gorm:"primaryKey"`
	Analyzed   bool
}

type RawDependRelation struct {
	ScriptIdentifier    string `gorm:"PrimaryKey"`
	TagIdentifier       string `gorm:"PrimaryKey"`
	DependentIdentifier string `gorm:"PrimaryKey"`
}

type CheckedPackage struct {
	RepoIdentifier    string `gorm:"PrimaryKey"`
	PackageIdentifier string `gorm:"PrimaryKey"`
	Finished          bool
	LastVisitedUrl    string
	FailedTimes       uint32
}
type CveAction struct {
	ID            string `gorm:"PrimaryKey"`
	Action        string
	AffectVersion string
	Date          string
}
