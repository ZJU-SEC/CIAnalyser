package model

import (
	"fmt"
	"gorm.io/gorm/clause"
	"sync"
)

type Measure struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"uniqueIndex"`
	Influential bool   `gorm:"default:false"`
}

func (m *Measure) FetchOrCreate() {
	var mutex sync.Mutex
	mutex.Lock()

	res := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(m)
	if res.Error != nil {
		fmt.Println("[ERR] cannot create measure", m.Name, res.Error)
	} else if res.RowsAffected == 0 {
		DB.Where(Measure{Name: m.Name}).First(m)
	}

	mutex.Unlock()
}
