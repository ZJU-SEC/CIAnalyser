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
	repoMap := make(map[string]*git.Repository)
	branchMap := make(map[string][]string)
	scriptMap := make(map[string]Script)

	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		repo, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
			continue
		}
		repoMap[s.Ref] = repo
		scriptMap[s.Ref] = s
		// traverse tags & branches
		branches, _ := repo.Branches()
		tags, _ := repo.Tags()
		branches.ForEach(func(r *plumbing.Reference) error {
			b := strings.TrimPrefix(r.Name().String(), "refs/heads/")
			branchMap[s.Ref] = append(branchMap[s.Ref], b)
			return nil
		})

		s.VersionCount = 0 // zero out count of version to persist idempotence
		tags.ForEach(func(r *plumbing.Reference) error {
			s.VersionCount++
			co, err := repo.TagObject(r.Hash())
			if err == nil {
				if co.Tagger.When.Unix() > s.LatestVersionTime {
					s.LatestVersionTime = co.Tagger.When.Unix()
				}
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

		// check branch
		if branches, ok := branchMap[u.ScriptRef()]; ok {
			if slices.Contains(branches, u.Version()) {
				u.UseBranch = true
				u.UpdateLag = 0
			}
		}

		// check tag
		if repo, ok := repoMap[u.ScriptRef()]; ok {
			if verTag, err := repo.Tag(u.Version()); err == nil {
				// is a tag
				u.UseLatest = true
				verObj, err := repo.TagObject(verTag.Hash())
				if err == nil {
					verTime := verObj.Tagger.When.Unix()

					tagObjs, _ := repo.TagObjects() // get tag iterators
					tagObjs.ForEach(func(tag *object.Tag) error {
						iterTime := tag.Tagger.When.Unix()
						if verTime < iterTime {
							u.UseLatest = false
							if -verTime > u.UpdateLag {
								u.UpdateLag = iterTime - verTime
							}
						}
						return nil
					})
				}
			}

			if commObj, err := repo.CommitObject(plumbing.NewHash(u.Version())); err == nil {
				// is a commit object
				commTime := commObj.Author.When.Unix()
				s := scriptMap[u.ScriptRef()]
				if commTime >= s.LatestVersionTime {
					u.UseLatest = true
				} else {
					u.UpdateLag = s.LatestVersionTime - commTime
				}
			}
		}
		u.Update()
	}
}
