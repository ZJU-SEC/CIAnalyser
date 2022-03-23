package models

import (
	"fmt"
	"gorm.io/gorm"
	"sync"
)

type Maintainer struct {
	ID       uint `gorm:"primaryKey;autoIncrement;"`
	Name     string
	Verified bool `gorm:"default:false"`
}

func (m *Maintainer) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	res := DB.Where("name = ?", m.Name).First(&Maintainer{})

	if res.Error == gorm.ErrRecordNotFound {
		if err := DB.Create(m).Error; err != nil {
			fmt.Println("[ERR] cannot index maintainer", m.Name, err)
		} else {
			fmt.Println("✔", m.Name, "created")
		}
	}

	mutex.Unlock()
}
