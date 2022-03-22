package analyzer

import (
	"CIHunter/src/models"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

// outputMaintainersInfluence scan the uses and calculate the maintainers' influence
// maintainer A --> script A/A --- influenced usecases
//              --> script A/B --- influenced usecases
// maintainer B --> script B/A --- influenced usecases
func outputMaintainersInfluence(f *excelize.File) {
	fmt.Println("\n[Maintainers' Influence]")

	extract := func(s string) (string, string) {
		return strings.Split(s, "/")[0], strings.Split(s, "@")[0]
	}

	// influenceMap record maintainers' influence on scripts and usecases
	influenceMap := make(map[string]map[string]int)

	rows, _ := models.DB.Model(&GHUse{}).
		Where("use LIKE ? AND use NOT LIKE ?", "%@%", "docker://%").Rows()

	for rows.Next() {
		var u GHUse
		models.DB.ScanRows(rows, &u)
		maintainer, script := extract(u.Use)

		if maintainerMap, ok := influenceMap[maintainer]; ok {
			if _, ok := maintainerMap[script]; ok {
				maintainerMap[script] += 1 // increment scriptCnt
			} else {
				maintainerMap[script] = 1 // create scriptMap
			}
		} else {
			influenceMap[maintainer] = make(map[string]int)
			influenceMap[maintainer][script] = 1
		}
	}

	//for
	idx := 1
	for maintainer, maintainerMap := range influenceMap {
		usecaseTot := 0
		for _, usecaseCount := range maintainerMap {
			usecaseTot += usecaseCount
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", idx), maintainer)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", idx), len(maintainerMap))
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", idx), usecaseTot)

		idx++
	}
}
