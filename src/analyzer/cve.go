package analyzer

import (
	"CIHunter/src/models"
	"github.com/olekukonko/tablewriter"
	"os"
)

// CVEList records mapping from CVE to certain script
var CVEList = map[string]string{
	"CVE-2021-32724": "check-spelling/check-spelling",
	//"CVE-2021-32638": "github/codeql-action",
	"CVE-2021-32074": "hashicorp/vault-action",
	"CVE-2020-15272": "ericcornelissen/git-tag-annotation-action",
	"CVE-2020-14189": "atlassian/gajira-comment",
	"CVE-2020-14188": "atlassian/gajira-create",
}

func analyzeCVE() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CVE", "Repository", "Script"})

	// traverse CVEs
	for cve, script := range CVEList {
		var uses []GHUse

		// match `uses`
		models.DB.Where("usecases LIKE ?", script+"%").Find(&uses)

		for _, u := range uses {
			var job GHJob
			var measure GHMeasure
			models.DB.First(&job, u.GHJobID)
			models.DB.First(&measure, job.GHMeasureID)
			table.Append([]string{
				cve,
				measure.RepoRef,
				u.Use,
			})
		}
	}

	table.Render()
}
