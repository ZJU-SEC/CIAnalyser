package script

import (
	"CIHunter/pkg/model"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func Lag() {
	scriptMap := make(map[string]Script)

	// calculate latest version time
	rows, _ := model.DB.Model(&Script{}).Where("checked = ?", true).Rows()
	for rows.Next() {
		var s Script
		model.DB.ScanRows(rows, &s)

		repo, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
			continue
		}
		scriptMap[s.Ref] = s

		var latestVersionTime int64 = 0
		tags, _ := repo.Tags()

		tags.ForEach(func(r *plumbing.Reference) error {
			var commitObj *object.Commit
			var tagObj *object.Tag
			var err error

			commitObj, err = repo.CommitObject(r.Hash())
			if err == nil {
				if latestVersionTime < commitObj.Author.When.Unix() {
					latestVersionTime = commitObj.Author.When.Unix()
				}
				return nil
			}

			tagObj, err = repo.TagObject(r.Hash())
			if err == nil {
				if latestVersionTime < tagObj.Tagger.When.Unix() {
					latestVersionTime = tagObj.Tagger.When.Unix()
				}
				return nil
			}
			fmt.Println("[ERR] no such tag", err)
			return nil
		})

		if s.LatestVersionTime != latestVersionTime {
			s.LatestVersionTime = latestVersionTime
			model.DB.Save(&s)
		}
	}

	// calculate update lag
	rows, _ = model.DB.Model(&Usage{}).Rows()
	for rows.Next() {
		var u Usage
		model.DB.ScanRows(rows, &u)

		s, ok := scriptMap[u.ScriptRef()]
		if !ok {
			continue
		}

		repo, err := git.PlainOpen(s.LocalPath())
		if err != nil {
			fmt.Println("[ERR] cannot open", s.SrcRef(), "as Git repository")
			continue
		}

		// declare time & lag
		var t, lag int64
		lag = -1
		version := u.Version()

		if u.UseBranch {
			t, err = getTimeByBranch(repo, version)
		} else if u.UseTag {
			t, err = getTimeByTag(repo, version)
		} else if u.UseHash {
			t, err = getTimeByHash(repo, version)
		}

		if err == nil {
			if t < s.LatestVersionTime {
				lag = s.LatestVersionTime - t
			} else {
				lag = 0
			}
		}

		if u.UpdateLag != lag {
			u.UpdateLag = lag
			model.DB.Save(&u)
		}
	}
}

func getTimeByBranch(r *git.Repository, v string) (int64, error) {
	hash, err := r.ResolveRevision(plumbing.Revision("remotes/origin/" + v))
	if err != nil {
		fmt.Println("[ERR] cannot resolve the branch", v, err)
		return -1, err
	}
	commitObj, _ := r.CommitObject(*hash)
	return commitObj.Author.When.Unix(), nil
}

func getTimeByHash(r *git.Repository, v string) (int64, error) {
	commitObj, err := r.CommitObject(plumbing.NewHash(v))
	if err != nil {
		fmt.Println("[ERR] cannot resolve the commit", v, err)
		return -1, err
	}
	return commitObj.Author.When.Unix(), nil
}

func getTimeByTag(r *git.Repository, v string) (int64, error) {
	var err error
	tag, err := r.Tag(v)
	if err != nil {
		fmt.Println("[ERR] cannot resolve the tag", v, err)
		return -1, err
	}

	var tagObj *object.Tag
	var commitObj *object.Commit

	tagObj, err = r.TagObject(tag.Hash())
	if err == nil {
		return tagObj.Tagger.When.Unix(), nil
	}

	commitObj, err = r.CommitObject(tag.Hash())
	if err == nil {
		return commitObj.Author.When.Unix(), nil
	}

	return -1, err
}
