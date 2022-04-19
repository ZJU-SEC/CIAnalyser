package script

import (
	"CIHunter/config"
	"CIHunter/pkg/model"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path"
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
				if parseDockerfile(path.Join(config.SCRIPTS_PATH, s.Ref, "Dockerfile")) {
					s.Using = "docker"
				}
			}
		}

		if err == nil {
			dec := yaml.NewDecoder(f)
			a := model.Action{}
			if err := dec.Decode(&a); err != nil {
				fmt.Println(s.Ref)
				continue
			}
			s.Using = a.Runs.Using
		}

		model.DB.Save(&s)
	}
}

func parseDockerfile(path string) bool {
	_, err := os.Open(path)
	if err != nil {
		return false
	}
	return true
}
