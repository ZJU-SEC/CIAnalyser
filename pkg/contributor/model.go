package contributor

import (
	"CIHunter/pkg/model"
	"CIHunter/pkg/script"
	"fmt"
	"gorm.io/gorm/clause"
	"sync"
)

type Contributor struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueKey" json:"login"`
}

func (c *Contributor) fetchOrCreate() {
	var mutex sync.Mutex
	mutex.Lock()

	res := model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(c)
	if res.Error != nil {
		fmt.Println("[ERR] cannot create contributor", c, res.Error)
	} else if res.RowsAffected == 0 {
		model.DB.Where(Contributor{Name: c.Name}).First(c)
	}

	mutex.Unlock()
}

type Contribution struct {
	ContributorID uint
	Contributor   Contributor
	ScriptID      uint
	Script        script.Script
}

func (c *Contribution) create() {
	var mutex sync.Mutex
	mutex.Lock()

	if err := model.DB.Create(c).Error; err != nil {
		fmt.Println("[ERR] cannot create contribution", c, err)
	}

	mutex.Unlock()
}
