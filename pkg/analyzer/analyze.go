package analyzer

import (
	"CIHunter/config"
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

func reportContributor(f *excelize.File) {
	const sheet = "contributor"
	f.NewSheet(sheet)

	// TODO
}
