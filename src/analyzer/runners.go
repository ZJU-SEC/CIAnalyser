package analyzer

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GHRunner struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	GHMeasureID uint
	GHMeasure   GHMeasure `gorm:"foreignKey:GHMeasureID"`
	Runner      string
}

func analyzeRunners(f *os.File) {
	dec := yaml.NewDecoder(f)
	w := Workflow{}
	if err := dec.Decode(&w); err != nil {
		return
	}

}
