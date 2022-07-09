package script

import (
	"CIAnalyser/pkg/model"
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
		origin, _ := repo.Remote("origin")
		tags, _ := repo.Tags()
		commits, _ := repo.CommitObjects()

		commits.ForEach(func(c *object.Commit) error {
			hash := c.Hash.String()
			longHashMap[s.Ref] = append(longHashMap[s.Ref], hash)
			shortHashMap[s.Ref] = append(shortHashMap[s.Ref], hash[0:6])
			return nil
		})

		refList, _ := origin.List(&git.ListOptions{})
		for _, ref := range refList {
			if !strings.HasPrefix(ref.Name().String(), "refs/heads/") {
				continue
			}
			b := strings.TrimPrefix(ref.Name().String(), "refs/heads/")
			branchMap[s.Ref] = append(branchMap[s.Ref], b)
		}

		s.VersionCount = 0 // zero out count of version to persist idempotence
		tags.ForEach(func(r *plumbing.Reference) error {
			s.VersionCount++
			t := strings.TrimPrefix(r.Name().String(), "refs/tags/")
			tagMap[s.Ref] = append(tagMap[s.Ref], t)
			return nil
		})

		model.DB.Save(&s)
	}

	fmt.Println("[INFO] cache finished")

	// traverse usage
	rows, _ = model.DB.Model(&Usage{}).Rows()
	for rows.Next() {
		var u Usage
		model.DB.ScanRows(rows, &u)

		change := false
		u.UseTag = false
		u.UseBranch = false
		u.UseHash = false

		// branch
		if branches, ok := branchMap[u.ScriptRef()]; ok {
			if slices.Contains(branches, u.Version()) && !u.UseBranch {
				u.UseBranch = true
				change = true
			}
		}

		// tag
		if tags, ok := tagMap[u.ScriptRef()]; ok {
			if slices.Contains(tags, u.Version()) && !u.UseTag {
				u.UseTag = true
				change = true
			}
		}

		// short hash
		if len(u.Version()) == 7 {
			if hashes, ok := shortHashMap[u.ScriptRef()]; ok {
				if slices.Contains(hashes, u.Version()) && !u.UseHash {
					u.UseHash = true
					change = true
				}
			}
		}

		// long hash
		if len(u.Version()) == 40 {
			if hashes, ok := longHashMap[u.ScriptRef()]; ok {
				if slices.Contains(hashes, u.Version()) && !u.UseHash {
					u.UseHash = true
					change = true
				}
			}
		}

		if change {
			model.DB.Save(&u)
		}
	}
}
