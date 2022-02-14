package crawler

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/gocolly/colly"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Crawl the repositories
func Crawl() {
	group := parallelizer.NewGroup(parallelizer.WithPoolSize(config.THREAD_SIZE))
	defer group.Close()

	for i := 1; i <= 50; i++ {
		i := i
		group.Add(func() {
			crawlGitstarRepo(i)
		})
		group.Add(func() {
			crawlGitstarUserOrg(i, true)  // crawl organizations
			crawlGitstarUserOrg(i, false) // crawl users
		})
	}

	group.Wait()
}

func commonCollector() *colly.Collector {
	c := colly.NewCollector()

	// set random `User-Agent`
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", utils.RandomString())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	return c
}

func crawlGitstarRepo(page int) {
	c := commonCollector()
	group := parallelizer.NewGroup(parallelizer.WithPoolSize(config.THREAD_SIZE))
	defer group.Close()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			group.Add(func() {
				crawlGitHubRepo(link)
			})
		}
	})

	pageURL := fmt.Sprintf("https://gitstar-ranking.com/repositories?page=%d", page)
	c.Visit(pageURL)
	group.Wait()
}

func crawlGitstarUserOrg(page int, org bool) {
	c := commonCollector()
	group := parallelizer.NewGroup(parallelizer.WithPoolSize(config.THREAD_SIZE))
	defer group.Close()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			group.Add(func() {
				crawlGitstarUserOrgRepo(link)
			})
		}
	})

	var url string
	if org {
		url = fmt.Sprintf("https://gitstar-ranking.com/users?page=%d", page)
	} else {
		url = fmt.Sprintf("https://gitstar-ranking.com/organizations?page=%d", page)
	}
	c.Visit(url)
	group.Wait()
}

// crawl the repositories hosted on Gitstar
func crawlGitstarUserOrgRepo(href string) {
	page := getPageOfUserOrg(href)
	group := parallelizer.NewGroup()
	defer group.Close()

	c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_full_item" {
			group.Add(func() {
				crawlGitHubRepo(link)
			})
		}
	})

	for i := 1; i <= page; i++ {
		url := fmt.Sprintf("https://gitstar-ranking.com%s?page=%d", href, i)
		c.Visit(url)
	}
	group.Wait()
}

// calculate the total page of user / org
func getPageOfUserOrg(href string) int {
	var num = 0

	c := commonCollector()

	c.OnHTML("h3", func(e *colly.HTMLElement) {
		header := strings.TrimSpace(e.Text)
		var err error
		num, err = strconv.Atoi(strings.Split(header, " ")[0])
		if err != nil {
			num = 0
		}
	})

	url := "https://gitstar-ranking.com" + href
	c.Visit(url)

	return (num-1)/50 + 1
}

func crawlGitHubRepo(href string) {
	repo := models.Repo{}

	res := models.DB.Where("Ref = ?", href).First(&repo)
	if res.Error == gorm.ErrRecordNotFound {
		// not found, create

		src, ghMeasure, ghUses, err := analyzeRepoByGit(href)
		if err != nil {
			fmt.Println("[ERROR] when trying analyze", href, err)
		} else {
			// no error happened when analyzing repository
			// create this repository
			err := models.DB.Transaction(func(tx *gorm.DB) error {
				repo.Ref = href
				repo.Source = src

				// create repository data
				if err := tx.Create(&repo).Error; err != nil {
					return err
				}

				// not using GitHub Actions, skip
				if ghMeasure != nil {
					// create measurement
					ghMeasure.RepoID = repo.ID
					ghMeasure.Repo = repo
					if err := tx.Create(ghMeasure).Error; err != nil {
						return err
					}

					// create uses data
					for i := 0; i < len(ghUses); i++ {
						ghUses[i].GitHubActionMeasureID = ghMeasure.ID
						ghUses[i].GitHubActionMeasure = *ghMeasure
					}
					if err := tx.Create(&ghUses).Error; err != nil {
						return err
					}
				}
				return nil // commit the whole transaction
			})
			if err != nil {
				fmt.Println("[ERROR] failed to create", href, "because:", err)
			} else {
				fmt.Println("[SUCCESS]", href, "created")
			}
		}
	} else if config.NOW.Sub(repo.UpdatedAt) > config.UPDATE_DIFF {
		// TODO update
		//fmt.Printf("update %s\n", href)
	}
}

func analyzeRepoByGit(href string) ([]byte, *models.GitHubActionMeasure, []models.GitHubActionUses, error) {
	repoName := strings.ReplaceAll(href[1:], "/", ":")
	repoPath := path.Join(config.DEV_SHM, repoName)

	if _, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:   "https://github.com" + href + ".git",
		Depth: 1,
	}); err != nil { // clone error, fast return
		return nil, nil, nil, err
	}

	// declare variables
	var ghMeasure *models.GitHubActionMeasure = nil
	var ghUses []models.GitHubActionUses

	// analyze workflows
	workflows, _ := ioutil.ReadDir(path.Join(repoPath, ".github/workflows"))
	for _, entry := range workflows {
		// if dir / not yml / not yaml skip
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}

		// open yaml configuration file
		yml, err := os.ReadFile(path.Join(repoPath, ".github/workflows", entry.Name()))
		if err != nil {
			continue
		}

		// parse workflow from yaml file
		w := models.Workflow{}
		if err := yaml.Unmarshal(yml, &w); err != nil {
			continue
		}

		// create ghMeasure
		if ghMeasure == nil {
			ghMeasure = new(models.GitHubActionMeasure)
		}

		// map result from workflow to measure / uses
		for _, job := range w.Jobs {
			for _, step := range job.Steps { // traverse `uses` item, if not empty, record
				if step.Uses != "" {
					ghUses = append(ghUses, models.GitHubActionUses{Uses: step.Uses})
				}
			}
		}
	}

	var src []byte = nil

	if ghMeasure != nil {
		src, _ = utils.SerializeRepo(repoName)
	}

	// remove local repo to save storage
	os.RemoveAll(repoPath)

	return src, ghMeasure, ghUses, nil
}
