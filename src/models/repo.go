package models

import (
	"CIHunter/src/config"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

// Repo schema for repo's metadata
type Repo struct {
	ID      uint   `gorm:"primaryKey"`
	Ref     string `gorm:"uniqueIndex"`
	Checked bool   `gorm:"default:false"`
}

// Create a repo
func Create(href string) {
	var mutex sync.Mutex
	mutex.Lock()

	repo := Repo{}
	res := DB.Where("ref = ?", href)

	if res.Error == gorm.ErrRecordNotFound {
		repo.Ref = href
		if err := DB.Create(&repo).Error; err != nil {
			fmt.Println("[ERR] cannot create️", href, err)
		} else {
			fmt.Println("✔", href, "created")
		}
	} else if config.DEBUG {
		fmt.Println("⚙️", href)
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

// TarballURL returns the url for tarball archive according to repo & branch
func (r *Repo) TarballURL(branch string) string {
	const DOMAIN = "https://codeload.github.com"
	return fmt.Sprintf("%s%s/tar.gz/refs/heads/%s", DOMAIN, r.Ref, branch)
}

// API return the url for api.github.com according to repo
func (r *Repo) API() string {
	const DOMAIN = "https://api.github.com/repos"
	return fmt.Sprintf("%s%s", DOMAIN, r.Ref)
}
