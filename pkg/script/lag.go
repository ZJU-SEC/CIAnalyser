package script

import (
	"CIHunter/pkg/model"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func Lag() {
	// calculate latest version time
	rows, _ := model.DB.Model(&Script{}).Rows()
	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		repo, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
			continue
		}

		var latestVersionTime int64 = 0
		tags, _ := repo.Tags()
		tags.ForEach(func(r *plumbing.Reference) error {
			to, err := repo.TagObject(r.Hash())
			if err == nil {
				if to.Tagger.When.Unix() > latestVersionTime {
					latestVersionTime = to.Tagger.When.Unix()
				}
			} else {
				fmt.Println("[ERR] cannot find tag object!")
			}
			return nil
		})

		if s.LatestVersionTime != latestVersionTime {
			s.LatestVersionTime = latestVersionTime
			model.DB.Save(&s)
		}
	}

	// calculate update lag
	//rows, _ = model.DB.Model(&Usage{}).Rows()
	//for rows.Next() {
	//	var u Usage
	//	model.DB.ScanRows(rows, &u)
	//
	//	var lag int64 = -1
	//	if u.UseBranch {
	//		lag = 0
	//	} else if u.UseTag {
	//
	//	} else if u.UseHash {
	//
	//	}
	//
	//	if u.UpdateLag != lag {
	//		u.UpdateLag = lag
	//		model.DB.Save(&u)
	//	}
	//}
}
