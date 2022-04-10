package script

import (
	"CIHunter/pkg/model"
	"gopkg.in/yaml.v3"
	"os"
)

// TODO ParseEnv
func ParseEnv() {
	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", true).Rows()

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		f, err := os.Open(s.LocalPath())
		if err != nil {
			continue
		}
		dec := yaml.NewDecoder(f)
		a := model.Action{}
		if err := dec.Decode(&a); err != nil {
			continue
		}

		s.Using = a.Using
	}
}
