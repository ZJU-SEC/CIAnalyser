package script

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"fmt"
	"golang.org/x/exp/slices"
	"gorm.io/gorm/clause"
	"path"
	"strings"
	"sync"
)

// Script schema for script's metadata
type Script struct {
	// basic
	ID  uint `gorm:"primaryKey;autoIncrement;"`
	Url string

	// crawl
	Ref           string
	Category      string
	OnMarketplace bool   `gorm:"default:false"`
	Verified      bool   `gorm:"default:false"`
	StarCount     string `gorm:"default:0"`

	// clone
	Cloned            bool `gorm:"default:false"`
	Using             string
	VersionCount      int   `gorm:"default:0"`
	LatestVersionTime int64 `gorm:"default:0"` // time for the latest version

	LastVisitedURL string // record last visited page for recovery
}

func (s *Script) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Model(&Script{}).Where(Script{Ref: s.Ref, Url: s.Url})
	if res.RowsAffected == 0 {
		model.DB.Create(s)
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

type Verified struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex"`
}

func manuallyVerify(name string) {
	var mutex sync.Mutex
	mutex.Lock()

	v := Verified{Name: name}
	model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&v)

	mutex.Unlock()
}

func (v *Verified) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(v)

	mutex.Unlock()
}

var verified []string = nil

func IsVerified(name string) bool {
	if verified == nil {
		model.DB.Model(&Verified{}).Select("name").Find(&verified)
		slices.Sort(verified)
	}

	_, exist := slices.BinarySearch(verified, name)
	return exist
}
