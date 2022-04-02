package credential

import (
	"CIHunter/pkg/model"
	"fmt"
	"sync"
)

type Credential struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	MeasureID  uint
	Measure    model.Measure `gorm:"foreignKey:MeasureID"`
	Credential string
}

func (c *Credential) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	err := model.DB.Create(c).Error
	if err != nil {
		fmt.Println("[ERR] cannot record credential", c)
	}

	mutex.Unlock()
}
