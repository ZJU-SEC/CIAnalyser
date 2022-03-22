package models

import (
	"CIHunter/src/config"
	"fmt"
	"gorm.io/gorm"
	"path"
	"sync"
)

// Repo schema for repo's metadata
type Repo struct {
	ID          uint   `gorm:"primaryKey;autoIncrement;"`
	Ref         string `gorm:"unique"`
	Checked     bool   `gorm:"default:false"`
	Influential bool   `gorm:"default:false"` // mark the usecase as influential
}

// CreateRepo a repo
func CreateRepo(href string) {
	var mutex sync.Mutex
	mutex.Lock()

	repo := Repo{}
	res := DB.Where("ref = ?", href).First(&repo)

	if res.Error == gorm.ErrRecordNotFound {
		repo.Ref = href
		if err := DB.Create(&repo).Error; err != nil {
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

	res := DB.Model(&Repo{}).Where("ref = ?", r.Ref).Update("checked", true)
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

	res := DB.Delete(r)
	if res.Error != nil {
		fmt.Println("[ERR] cannot delete", r.Ref, res.Error)
	}

	mutex.Unlock()
}

// TarballURL returns the url for tarball archive according to repo & branch
func (r *Repo) TarballURL(branch string) string {
	const DOMAIN = "https://codeload.github.com"
	return fmt.Sprintf("%s%s/tar.gz/refs/heads/%s", DOMAIN, r.Ref, branch)
}

func (r *Repo) GitURL() string {
	return "https://github.com" + r.Ref + ".git"
}

// API return the url for api.github.com according to repo
func (r *Repo) API() string {
	const DOMAIN = "https://api.github.com/repos"
	return fmt.Sprintf("%s%s", DOMAIN, r.Ref)
}

func (r *Repo) WorkflowURL(branch string) string {
	const DOMAIN = "https://github.com"
	return fmt.Sprintf("%s%s/tree/%s/.github/workflows", DOMAIN, r.Ref, branch)
}

func (r *Repo) LocalPath() string {
	return path.Join(config.REPOS_PATH, r.Ref[1:])
}

func (r *Repo) WorkflowsPath() string {
	return path.Join(r.LocalPath(), ".github", "workflows")
}

func (r *Repo) GitPath() string {
	return path.Join(r.LocalPath(), ".git")
}
