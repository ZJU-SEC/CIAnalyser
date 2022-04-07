package script

import (
	"CIHunter/pkg/model"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/exp/slices"
	"strings"
)

func Label() {
	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", true).Rows()
	tagMap := make(map[string][]string)
	branchMap := make(map[string][]string)

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		r, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
		} else {
			branches, _ := r.Branches()
			tags, _ := r.Tags()
			branches.ForEach(func(r *plumbing.Reference) error {
				b := strings.TrimPrefix(r.Name().String(), "refs/heads/")
				branchMap[s.Ref] = append(branchMap[s.Ref], b)
				return nil
			})
			tags.ForEach(func(r *plumbing.Reference) error {
				t := strings.TrimPrefix(r.Name().String(), "refs/tags/")
				tagMap[s.Ref] = append(tagMap[s.Ref], t)
				return nil
			})
		}
	}

	// traverse usage
	rows, _ = model.DB.Model(&Usage{}).Rows()
	for rows.Next() {
		var u Usage
		model.DB.ScanRows(rows, &u)

		changes := false
		ref := u.SrcRef()

		if branches, ok := branchMap[ref]; ok {
			if slices.Contains(branches, u.Version()) {
				u.UseBranch = true
				changes = true
			}
		}

		if tags, ok := tagMap[ref]; ok {
			if tags[len(tags)-1] == u.Version() {
				u.UseLatest = true
				changes = true
			}
		}

		if changes {
			u.Update()
		}
	}
}
