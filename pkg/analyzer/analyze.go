package analyzer

import (
	"CIHunter/config"
	"CIHunter/pkg/credential"
	"CIHunter/pkg/model"
	"CIHunter/pkg/script"
	"CIHunter/pkg/verified"
	"fmt"
	"github.com/xuri/excelize/v2"
)

func Analyze() {
	f := excelize.NewFile()

	reportVerified(f)
	reportContributor(f)
	reportCredential(f)
	reportMaintainer(f)

	if err := f.SaveAs(config.REPORT); err != nil {
		fmt.Println("[ERR] cannot save report to", config.REPORT)
	}
}

func reportVerified(f *excelize.File) {
	const sheet = "verified"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "A1", "Item")
	f.SetCellValue(sheet, "B1", "Verified")
	f.SetCellValue(sheet, "C1", "Unverified")

	f.SetCellValue(sheet, "A2", "Creators")
	f.SetCellValue(sheet, "A3", "Scripts")
	f.SetCellValue(sheet, "A4", "Influenced Repos")

	var vCreator, tCreator int64
	model.DB.Model(&verified.Verified{}).Count(&vCreator)
	model.DB.Model(&script.Script{}).Distinct("maintainer").Count(&tCreator)
	f.SetCellValue(sheet, "B2",
		fmt.Sprintf("%d (%.2f%%)",
			vCreator, float64(vCreator*100)/float64(tCreator)))
	f.SetCellValue(sheet, "C2",
		fmt.Sprintf("%d (%.2f%%)",
			tCreator-vCreator, float64((tCreator-vCreator)*100)/float64(tCreator)))

	var vScript, tScript int64
	model.DB.Model(&script.Script{}).Count(&tScript)
	model.DB.Model(&script.Script{}).Where("verified = ?", true).Count(&vScript)
	f.SetCellValue(sheet, "B3",
		fmt.Sprintf("%d (%.2f%%)",
			vScript, float64(vScript*100)/float64(tScript)))
	f.SetCellValue(sheet, "C3",
		fmt.Sprintf("%d (%.2f%%)",
			tScript-vScript, float64((tScript-vScript)*100)/float64(tScript)))

	var vRepo, uvRepo, tRepo int64
	model.DB.Model(&model.Measure{}).Count(&tRepo)
	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("verified = ?", true).Distinct("measure_id").
		Count(&vRepo)
	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("verified = ?", false).Distinct("measure_id").
		Count(&uvRepo)

	f.SetCellValue(sheet, "B4",
		fmt.Sprintf("%d (%.2f%%)",
			vRepo, float64(vRepo*100)/float64(tRepo)))
	f.SetCellValue(sheet, "C4",
		fmt.Sprintf("%d (%.2f%%)",
			uvRepo, float64(uvRepo*100)/float64(tRepo)))

	f.SetCellValue(sheet, "A6", "Repos Importing Only Verified Scripts")

	var voRepo int64
	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Group("measure_id").
		Having("SUM(CASE WHEN verified THEN 1 ELSE 0 END) = COUNT(*)").
		Distinct("measure_id").
		Count(&voRepo)
	f.SetCellValue(sheet, "B6",
		fmt.Sprintf("%d (%.2f%%)",
			voRepo, float64(voRepo*100)/float64(tRepo)))
}

// TODO
func reportContributor(f *excelize.File) {
	const sheet = "contributor"
	f.NewSheet(sheet)

}

func reportCredential(f *excelize.File) {
	const sheet = "credential"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "A1", "# of credentials")
	f.SetCellValue(sheet, "B1", "# of repos")
	f.SetCellValue(sheet, "C1", "% of repos")

	// calculate number of total repositories
	var totalR int64
	model.DB.Model(&model.Measure{}).Count(&totalR)

	const MAX_CRED = 5
	for i := 1; i <= MAX_CRED; i++ {
		var c int64

		if i == MAX_CRED {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", i+1),
				fmt.Sprintf(">=%d", i))
			model.DB.Model(&credential.Credential{}).Group("measure_id").
				Having("count(*) >= ?", i).Distinct("measure_id").Count(&c)
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", i+1), i)
			model.DB.Model(&credential.Credential{}).Group("measure_id").
				Having("count(*) = ?", i).Distinct("measure_id").Count(&c)
		}

		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+1), c)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+1),
			fmt.Sprintf("%.2f%%", float64(c)/float64(totalR)*100))
	}
}

// TODO
func reportVul(f *excelize.File) {

}

func reportMaintainer(f *excelize.File) {
	const sheet = "maintainer"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "A1", "maintainer")
	f.SetCellValue(sheet, "B1", "# of influenced repos")
	f.SetCellValue(sheet, "C1", "% of influenced repos")

	f.SetCellValue(sheet, "E1", "unverified maintainer")
	f.SetCellValue(sheet, "F1", "# of influenced repos")
	f.SetCellValue(sheet, "G1", "% of influenced repos")

	var totalR int64
	model.DB.Model(&model.Measure{}).Count(&totalR)

	type Result struct {
		Maintainer string
		Count      int
	}
	var result []Result

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("verified = ?", true).
		Group("maintainer").
		Select("COUNT(DISTINCT(measure_id)) AS count", "maintainer").
		Order("count DESC").
		Limit(10).
		Scan(&result)

	for i := 0; i < len(result); i++ {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), result[i].Maintainer)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), result[i].Count)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2),
			float64(result[i].Count)/float64(totalR))
	}

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("verified = ?", false).
		Group("maintainer, verified").
		Select("COUNT(DISTINCT(measure_id)) AS count", "maintainer", "verified").
		Order("count DESC").
		Limit(10).
		Scan(&result)

	for i := 0; i < len(result); i++ {
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+2), result[i].Maintainer)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+2), result[i].Count)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+2),
			float64(result[i].Count)/float64(totalR))
	}
}
