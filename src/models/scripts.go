package models

import (
	"fmt"
	"gorm.io/gorm"
	"sync"
)

// Script schema for script's metadata
type Script struct {
	ID      uint   `gorm:"primaryKey;autoIncrement;"`
	Ref     string `gorm:"uniqueIndex"`
	Name    string
	SrcRef  string // ref for the source code
	Checked bool   `gorm:"default:false"`
	Using   string
}

func (s *Script) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	res := DB.Where("ref = ?", s.Ref).First(&Script{})

	if res.Error == gorm.ErrRecordNotFound {
		if err := DB.Create(s).Error; err != nil {
			fmt.Println("[ERR] cannot index script", s.Ref, err)
		} else {
			fmt.Println("✔", s.Ref, "created")
		}
	}

	mutex.Unlock()
}
