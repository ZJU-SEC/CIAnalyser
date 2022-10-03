package repo

import (
	"CIAnalyser/config"
	"CIAnalyser/pkg/model"
	"fmt"
	"gorm.io/gorm"
	"path"
	"sync"
)

// Repo schema for repo's metadata
type Repo struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Ref     string `gorm:"uniqueIndex"`
	Checked bool   `gorm:"default:false"`
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

	res := model.DB.Model(&Repo{}).Where("ref = ?", r.Ref).Update("checked", true)
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
	return "https://github.com" + r.Ref + ".git"
}

func (r *Repo) LocalPath() string {
	return path.Join(config.REPOS_PATH, r.Ref[1:])
}

func (r *Repo) WorkflowsPath() string {
	return path.Join(r.LocalPath(), ".github", "workflows")
}
