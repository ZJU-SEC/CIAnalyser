package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"fmt"
	"gorm.io/gorm/clause"
	"path"
	"strings"
	"sync"
)

// Script schema for script's metadata
type Script struct {
	ID                uint   `gorm:"primaryKey;autoIncrement;"`
	Ref               string `gorm:"uniqueIndex"`
	Maintainer        string
	Verified          bool `gorm:"default:false"`
	Checked           bool `gorm:"default:false"`
	Using             string
	IsDeployment      bool  `gorm:"default:false"`
	IsRelease         bool  `gorm:"default:false"`
	VersionCount      int   `gorm:"default:0"`
	LatestVersionTime int64 `gorm:"default:0"` // time for the latest version
}

func (s *Script) fetchOrCreate() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(s)
	if res.Error != nil { // create failed
		fmt.Println("[ERR] cannot create script", s.Ref, res.Error)
	} else if res.RowsAffected == 0 { // create nothing, fetch
		model.DB.Where(Script{Ref: s.Ref}).First(s)
	}

	mutex.Unlock()
}

func (s *Script) Check() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Model(s).Update("checked", true)
	if res.Error != nil {
		fmt.Println("[ERR] cannot check", s.Ref, res.Error)
	} else {
		fmt.Println("✔", s.Ref, "processed")
	}

	mutex.Unlock()
}

func (s *Script) Delete() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Delete(s)
	if res.Error != nil {
		fmt.Println("[ERR] cannot delete", s.Ref, res.Error)
	}

	mutex.Unlock()
}

func (s *Script) SrcRef() string {
	ss := strings.Split(s.Ref, "/")
	return ss[0] + "/" + ss[1]
}

func (s *Script) LocalPath() string {
	return path.Join(config.SCRIPTS_PATH, s.SrcRef())
}

func (s *Script) LocalYMLPath() string {
	return path.Join(config.SCRIPTS_PATH, s.Ref, "action.yml")
}

func (s *Script) LocalYAMLPath() string {
	return path.Join(config.SCRIPTS_PATH, s.Ref, "action.yaml")
}

func (s *Script) GitURL() string {
	return "https://github.com/" + s.SrcRef() + ".git"
}

func (s *Script) SrcURL() string {
	return "https://github.com/" + s.SrcRef()
}

type Usage struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	MeasureID uint
	Measure   model.Measure `gorm:"foreignKey:MeasureID"`
	ScriptID  uint
	Script    Script `gorm:"foreignKey:ScriptID"`
	Use       string
	UseBranch bool  `gorm:"default:false"`
	UseTag    bool  `gorm:"default:false"`
	UseHash   bool  `gorm:"default:false"`
	UpdateLag int64 // record usage's update lag
}

func (u *Usage) create() {
	var mutex sync.Mutex
	mutex.Lock()

	if err := model.DB.Create(u).Error; err != nil {
		fmt.Println("[ERR] cannot create usage", u, err)
	}

	mutex.Unlock()
}

func (u *Usage) Update() {
	var mutex sync.Mutex
	mutex.Lock()

	model.DB.Save(u)

	mutex.Unlock()
}

// ScriptRef return the script in this usage
func (u *Usage) ScriptRef() string {
	script := strings.Split(u.Use, "@")[0]
	return script
}

func (u *Usage) Version() string {
	return strings.Split(u.Use, "@")[1]
}
