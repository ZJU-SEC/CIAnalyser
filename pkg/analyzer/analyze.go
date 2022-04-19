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
	reportCategory(f)
	reportUsing(f)

	if err := f.SaveAs(config.REPORT); err != nil {
		fmt.Println("[ERR] cannot save report to", config.REPORT)
	}
}

func reportUsing(f *excelize.File) {
	const sheet = "using"

	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "Item")
	f.SetCellValue(sheet, "B1", "Docker")
	f.SetCellValue(sheet, "C1", "Node.js")
	f.SetCellValue(sheet, "D1", "Others")

	f.SetCellValue(sheet, "A2", "# of scripts")
	f.SetCellValue(sheet, "A3", "% of scripts")
	f.SetCellValue(sheet, "A4", "# of usage")
	f.SetCellValue(sheet, "A5", "% of usage")

	var NofDockerScript, NofNodeScript, totalS,
		NofDockerUsage, NofNodeUsage, NofOtherUsage, totalR int64

	model.DB.Model(&script.Script{}).Count(&totalS)
	model.DB.Model(&model.Measure{}).Count(&totalR)

	model.DB.Model(&script.Script{}).
		Where("\"using\" ILIKE ?", "docker%").Count(&NofDockerScript)
	model.DB.Model(&script.Script{}).
		Where("\"using\" ILIKE ?", "node%").Count(&NofNodeScript)

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("\"using\" ILIKE ?", "docker").
		Distinct("measure_id").Count(&NofDockerUsage)
	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("\"using\" ILIKE ?", "node%").
		Distinct("measure_id").Count(&NofNodeUsage)
	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("\"using\" NOT ILIKE ? AND \"using\" NOT ILIKE ?", "docker", "node%").
		Distinct("measure_id").Count(&NofOtherUsage)

	f.SetCellValue(sheet, "B2", NofDockerScript)
	f.SetCellValue(sheet, "C2", NofNodeScript)
	f.SetCellValue(sheet, "D2", totalS-NofDockerScript-NofNodeScript)

	f.SetCellValue(sheet, "B4", NofDockerUsage)
	f.SetCellValue(sheet, "C4", NofNodeUsage)
	f.SetCellValue(sheet, "D4", NofOtherUsage)
}

// TODO
func reportVul(f *excelize.File) {
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
		Distinct("measure_id").Count(&voRepo)
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
	var i int
	for i = 1; i <= MAX_CRED; i++ {
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

	var m int64
	f.SetCellValue(sheet, fmt.Sprintf("A%d", i+3),
		"Maximum existence of credentials")
	model.DB.Model(&credential.Credential{}).Group("measure_id").
		Select("COUNT(*) as coc").Order("coc DESC").Limit(1).Scan(&m)
	f.SetCellValue(sheet, fmt.Sprintf("B%d", i+3), m)
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
		Where("verified = ?", true).Group("maintainer").
		Select("COUNT(DISTINCT(measure_id)) AS count", "maintainer").
		Order("count DESC").Limit(10).Scan(&result)

	for i := 0; i < len(result); i++ {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), result[i].Maintainer)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), result[i].Count)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2),
			float64(result[i].Count)/float64(totalR))
	}

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("verified = ?", false).Group("maintainer, verified").
		Select("COUNT(DISTINCT(measure_id)) AS cm", "maintainer", "verified").
		Order("cm DESC").Limit(10).Scan(&result)

	for i := 0; i < len(result); i++ {
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+2), result[i].Maintainer)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+2), result[i].Count)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+2),
			float64(result[i].Count)/float64(totalR))
	}
}

func reportCategory(f *excelize.File) {
	const sheet = "category"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "B1", "deployment")
	f.SetCellValue(sheet, "C1", "artifact")
	f.SetCellValue(sheet, "A2", "# of script")
	f.SetCellValue(sheet, "A3", "% of script")
	f.SetCellValue(sheet, "A4", "# of usage")
	f.SetCellValue(sheet, "A5", "% of usage")

	var DeploymentScript, DeploymentUsage, ArtifactScript, ArtifactUsage int64
	model.DB.Model(&script.Script{}).
		Where("is_deployment = ?", true).Count(&DeploymentScript)
	model.DB.Model(&script.Script{}).
		Where("is_release = ?", true).Count(&ArtifactScript)

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("is_deployment = ?", true).
		Distinct("measure_id").Count(&DeploymentUsage)

	model.DB.Model(&script.Usage{}).
		Joins("LEFT JOIN scripts ON scripts.id = usages.script_id").
		Where("is_release = ?", true).
		Distinct("measure_id").Count(&ArtifactUsage)

	var totalS, totalR int64
	model.DB.Model(&script.Script{}).Count(&totalS)
	model.DB.Model(&model.Measure{}).Count(&totalR)

	f.SetCellValue(sheet, "B2", DeploymentScript)
	f.SetCellValue(sheet, "C2", ArtifactScript)

	f.SetCellValue(sheet, "B3", float64(DeploymentScript)/float64(totalS))
	f.SetCellValue(sheet, "C3", float64(ArtifactScript)/float64(totalS))

	f.SetCellValue(sheet, "B4", DeploymentUsage)
	f.SetCellValue(sheet, "C4", ArtifactUsage)

	f.SetCellValue(sheet, "B5", float64(DeploymentUsage)/float64(totalR))
	f.SetCellValue(sheet, "C5", float64(ArtifactUsage)/float64(totalR))
}
