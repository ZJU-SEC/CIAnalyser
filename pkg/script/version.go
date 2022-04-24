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
	branchMap := make(map[string][]string)
	tagMap := make(map[string][]string)
	longHashMap := make(map[string][]string)
	shortHashMap := make(map[string][]string)

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
		commits, _ := repo.CommitObjects()

		commits.ForEach(func(c *object.Commit) error {
			hash := c.Hash.String()
			longHashMap[s.Ref] = append(longHashMap[s.Ref], hash)
			shortHashMap[s.Ref] = append(shortHashMap[s.Ref], hash[0:6])

			return nil
		})

		branches.ForEach(func(r *plumbing.Reference) error {
			b := strings.TrimPrefix(r.Name().String(), "refs/heads/")
			branchMap[s.Ref] = append(branchMap[s.Ref], b)
			return nil
		})

		s.VersionCount = 0 // zero out count of version to persist idempotence
		tags.ForEach(func(r *plumbing.Reference) error {
			s.VersionCount++
			t := strings.TrimPrefix(r.Name().String(), "refs/tags/")
			tagMap[s.Ref] = append(tagMap[s.Ref], t)
			return nil
		})

		model.DB.Save(&s)
	}

	// traverse usage
	rows, _ = model.DB.Model(&Usage{}).Rows()
	for rows.Next() {
		var u Usage
		model.DB.ScanRows(rows, &u)

		// check branch
		if branches, ok := branchMap[u.ScriptRef()]; ok {
			if slices.Contains(branches, u.Version()) {
				u.UseBranch = true
			}
		}

		if tags, ok := branchMap[u.ScriptRef()]; ok {
			if slices.Contains(tags, u.Version()) {
				u.UseTag = true
			}
		}

		// short / long hash
		if len(u.Version()) == 7 {
			if hashes, ok := shortHashMap[u.ScriptRef()]; ok {
				if slices.Contains(hashes, u.Version()) {
					u.UseHash = true
				}
			}
		}

		if len(u.Version()) == 40 {
			if hashes, ok := longHashMap[u.ScriptRef()]; ok {
				if slices.Contains(hashes, u.Version()) {
					u.UseHash = true
				}
			}
		}

		model.DB.Save(&u)
	}
}
