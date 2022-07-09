package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"CIAnalyser/pkg/script"
	"fmt"
	"path"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// Repo schema for repo's metadata
type Repo struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Ref       string `gorm:"uniqueIndex"`
	Cloned    bool   `gorm:"default:false"`
	StarCount uint   `gorm:"default:0"`
	ForkCount uint   `gorm:"default:0"`
}

type Dependency struct {
	RepoID   uint
	Repo     Repo `gorm:"foreignKey:RepoID"`
	ScriptID uint
	Script   script.Script `gorm:"foreignKey:ScriptID"`
}

// CreateRepo a repo
func CreateRepo(href string) {
	var mutex sync.Mutex
	mutex.Lock()

	repo := Repo{}
	res := model.DB.Where("ref = ?", href).First(&repo)

	if res.Error == gorm.ErrRecordNotFound {
		repo.Ref = href
		if err := model.DB.Create(&repo).Error; err != nil {
			fmt.Println("[ERR] cannot index usecase", href, err)
		} else {
			fmt.Println("✔", href, "created")
		}
	}

	mutex.Unlock()
}

func (r *Repo) Check() {
	var mutex sync.Mutex
	mutex.Lock()

	// res := model.DB.Model(&Repo{}).Where("ref = ?", r.Ref).Update("checked", true)
	res := model.DB.Model(&Repo{}).Where("ref = ?", r.Ref).Update("cloned", true)
	if res.Error != nil {
		fmt.Println("[ERR] cannot check", r.Ref, res.Error)
	} else {
		fmt.Println("✔", r.Ref, "processed")
	}

	mutex.Unlock()
}

func (r *Repo) Delete() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Delete(r)
	if res.Error != nil {
		fmt.Println("[ERR] cannot delete", r.Ref, res.Error)
	}

	mutex.Unlock()
}

func (r *Repo) GitURL() string {
	// if start with a slash, remove it first
	ref := "/" + strings.TrimPrefix(r.Ref, "/")
	return "https://github.com" + ref + ".git"
}

func (r *Repo) LocalPath() string {
	// if start with a slash, remove it first
	path_with_slash := "/" + strings.TrimPrefix(r.Ref, "/")
	return path.Join(config.REPOS_PATH, path_with_slash[1:])
}

func (r *Repo) WorkflowsPath() string {
	return path.Join(r.LocalPath(), ".github", "workflows")
}
