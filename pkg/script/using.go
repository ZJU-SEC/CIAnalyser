package script

import (
	"CIHunter/pkg/model"
	"gopkg.in/yaml.v3"
	"os"
)

func ParseUsing() {
	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", true).Rows()

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		// parse yaml or yml file
		f, err := os.Open(s.LocalYMLPath())
		if err != nil {
			f, err = os.Open(s.LocalYAMLPath())
			if err != nil {
				continue
			}
		}
		dec := yaml.NewDecoder(f)
		a := model.Action{}
		if err := dec.Decode(&a); err != nil {
			continue
		}

		model.DB.Model(&Script{}).Where("id = ?", s.ID).
			Update("using", a.Runs.Using)
	}
}
