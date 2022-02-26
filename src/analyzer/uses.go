package analyzer

import "os"

type GHUse struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID uint
	GHMeasure   GHMeasure `gorm:"foreignKey:GHMeasureID"`
	Use         string
}

func analyzeUses(f *os.File) {

}
