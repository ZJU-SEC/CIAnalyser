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
	reportVersion(f)
	reportCVE(f)

	if err := f.SaveAs(config.REPORT); err != nil {
		fmt.Println("[ERR] cannot save report to", config.REPORT)
	}
}

func reportVersion(f *excelize.File) {
	const sheet = "version"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "Version Count")
	f.SetCellValue(sheet, "B1", "# of Repositories")

	iter := 2
	THRESHOLD := 100
	for bottom := 0; bottom <= THRESHOLD; bottom += 10 {
		var c int64
		up := bottom + 10

		if bottom == THRESHOLD {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", iter),
				fmt.Sprintf(">= %d", bottom))
			model.DB.Model(&script.Script{}).
				Where("version_count >= ?", bottom).Count(&c)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", iter), c)
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", iter),
				fmt.Sprintf("[%d, %d)", bottom, up))
			model.DB.Model(&script.Script{}).
				Where("version_count >= ? AND version_count < ?", bottom, up).Count(&c)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", iter), c)
		}

		iter++
	}
}

func reportCVE(f *excelize.File) {
	const sheet = "CVE"

	CVEmapping := map[string]string{
		"check-spelling/check-spelling":             "CVE-2021-32724",
		"github/codeql-action":                      "CVE-2021-32638",
		"hashicorp/vault-action":                    "CVE-2021-32074",
		"ericcornelissen/git-tag-annotation-action": "CVE-2020-15272",
		"atlassian/gajira-comment":                  "CVE-2020-14189",
		"atlassian/gajira-create":                   "CVE-2020-14188",
	}

	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "CVE")
	f.SetCellValue(sheet, "B1", "Repository")
	f.SetCellValue(sheet, "C1", "Script")
	f.SetCellValue(sheet, "D1", "Usage")

	iter := 0
	for u, cve := range CVEmapping {
		type Result struct {
			Name string
			Ref  string
			Use  string
		}

		var result []Result
		model.DB.Model(&script.Usage{}).
			Joins("LEFT JOIN measures m on m.id = usages.measure_id").
			Joins("LEFT JOIN scripts s on usages.script_id = s.id").
			Select("name", "ref", "use").
			Where("use ILIKE ?", u+"%").
			Where("use NOT ILIKE ?", u+"%@main").
			Where("use NOT ILIKE ?", u+"%@master").
			Where("use NOT ILIKE ?", u+"%@master").
			Where("use NOT ILIKE ?", u+"%@prerelease").
			Where("use NOT ILIKE ?", u+"%@%/%").
			Where("use NOT ILIKE ?", "hashicorp/vault-action@v2.4.0").
			Where("use NOT ILIKE ?", "github/codeql-action/%@v1%").
			Where("use NOT ILIKE ?", "atlassian/gajira-create@v2.0.1").
			Where("use NOT ILIKE ?", "atlassian/gajira-comment@v2.0.2").
			Where("use NOT ILIKE ?", "check-spelling/check-spelling@v0.0.19").
			Where("use NOT ILIKE ?", "check-spelling/check-spelling@v0.0.20%").
			Distinct().
			Scan(&result)

		for _, r := range result {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", iter), cve)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", iter), r.Name)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", iter), r.Ref)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", iter), r.Use)
			iter++
		}
	}
}

func reportUsing(f *excelize.File) {
	const sheet = "using"

	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "Item")
	f.SetCellValue(sheet, "B1", "Docker")
	f.SetCellValue(sheet, "C1", "Node.js")
	f.SetCellValue(sheet, "D1", "Raw Command")

	f.SetCellValue(sheet, "A2", "# of scripts")
	f.SetCellValue(sheet, "A3", "% of scripts")
	f.SetCellValue(sheet, "A4", "# of usage")
	f.SetCellValue(sheet, "A5", "% of usage")

	var NofDockerScript, NofNodeScript, NofRCScript, totalS,
		NofDockerUsage, NofNodeUsage, NofRCUsage, totalR int64

	model.DB.Model(&script.Script{}).Count(&totalS)
	model.DB.Model(&model.Measure{}).Count(&totalR)

	model.DB.Model(&script.Script{}).
		Where("\"using\" ILIKE ?", "docker%").Count(&NofDockerScript)
	model.DB.Model(&script.Script{}).
		Where("\"using\" ILIKE ?", "node%").Count(&NofNodeScript)
	model.DB.Model(&script.Script{}).
		Where("\"using\" ILIKE ?", "composite").Count(&NofRCScript)

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
		Where("\"using\" ILIKE ?", "composite").
		Distinct("measure_id").Count(&NofRCUsage)

	f.SetCellValue(sheet, "B2", NofDockerScript)
	f.SetCellValue(sheet, "C2", NofNodeScript)
	f.SetCellValue(sheet, "D2", NofRCScript)

	f.SetCellValue(sheet, "B3", float64(NofDockerScript)/float64(totalS))
	f.SetCellValue(sheet, "C3", float64(NofNodeScript)/float64(totalS))
	f.SetCellValue(sheet, "D3", float64(NofRCScript)/float64(totalS))

	f.SetCellValue(sheet, "B4", NofDockerUsage)
	f.SetCellValue(sheet, "C4", NofNodeUsage)
	f.SetCellValue(sheet, "D4", NofRCUsage)

	f.SetCellValue(sheet, "B5", float64(NofDockerUsage)/float64(totalR))
	f.SetCellValue(sheet, "C5", float64(NofNodeUsage)/float64(totalR))
	f.SetCellValue(sheet, "D5", float64(NofRCUsage)/float64(totalR))
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
