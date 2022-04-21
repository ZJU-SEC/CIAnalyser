package script

import (
	"CIHunter/pkg/model"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/exp/slices"
	"strings"
)

func Label() {
	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", true).Rows()

	// record scripts' tags & branches
	tagMap := make(map[string][]string)
	branchMap := make(map[string][]string)

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		repo, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
			continue
		}
		// traverse tags & branches
		branches, _ := repo.Branches()
		tags, _ := repo.Tags()
		branches.ForEach(func(r *plumbing.Reference) error {
			b := strings.TrimPrefix(r.Name().String(), "refs/heads/")
			branchMap[s.Ref] = append(branchMap[s.Ref], b)
			return nil
		})

		tags.ForEach(func(r *plumbing.Reference) error {
			t := strings.TrimPrefix(r.Name().String(), "refs/tags/")
			tagMap[s.Ref] = append(tagMap[s.Ref], t)
			co, err := repo.CommitObject(r.Hash())
			if err == nil {
				if uint(co.Author.When.Unix()) > s.ReleaseAt {
					s.ReleaseAt = uint(co.Author.When.Unix())
				}
			}
			return nil
		})

		// find update_at
		cIter, _ := repo.CommitObjects()
		cIter.ForEach(func(c *object.Commit) error {
			if uint(c.Author.When.Unix()) > s.UpdateAt {
				s.UpdateAt = uint(c.Author.When.Unix())
			}
			return nil
		})
		model.DB.Save(&s)
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
