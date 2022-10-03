package verified

import (
	"CIAnalyser/pkg/model"
	"golang.org/x/exp/slices"
	"gorm.io/gorm/clause"
	"sync"
)

type Verified struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex"`
}

func (v *Verified) Create() {
	var mutex sync.Mutex
	mutex.Lock()

	model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(v)

	mutex.Unlock()
}

var verified []string = nil

func Exist(name string) bool {
	if verified == nil {
		model.DB.Model(&Verified{}).Select("name").Find(&verified)
		slices.Sort(verified)
	}

	_, exist := slices.BinarySearch(verified, name)
	return exist
}
