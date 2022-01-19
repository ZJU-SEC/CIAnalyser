package crawler

import (
	"CIHunter/src/config"
	"CIHunter/src/models"
	"CIHunter/src/utils"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gocolly/colly"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"path/filepath"
	"strconv"
	"strings"
)

// Crawl the repositories
func Crawl() {
	crawlGitstarRepo()
	crawlGitstarUserOrg()
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

func crawlGitstarRepo() {
	c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			crawlGitHubRepo(link)
		}
	})

	for i := 1; i <= 50; i++ {
		pageURL := fmt.Sprintf("https://gitstar-ranking.com/repositories?page=%d", i)
		c.Visit(pageURL)
	}
}

func crawlGitstarUserOrg() {
	c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_item" {
			crawlGitstarUserOrgRepo(link)
		}
	})

	for i := 1; i <= 50; i++ {
		userURL := fmt.Sprintf("https://gitstar-ranking.com/users?page=%d", i)
		orgURL := fmt.Sprintf("https://gitstar-ranking.com/organizations?page=%d", i)
		c.Visit(userURL)
		c.Visit(orgURL)
	}
}

// crawl the repositories hosted on Gitstar
func crawlGitstarUserOrgRepo(href string) {
	page := getPageOfUserOrg(href)

	c := commonCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Attr("class") == "list-group-item paginated_full_item" {
			crawlGitHubRepo(link)
		}
	})

	for i := 1; i <= page; i++ {
		url := fmt.Sprintf("https://gitstar-ranking.com%s?page=%d", href, i)
		c.Visit(url)
	}
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
		fmt.Printf("create %s\n", href)

		ghMeasure, ghUses, err := analyzeRepoByGit(href)
		if err != nil {

		} else {
			// no error happened when analyzing repository
			// create this repository
			models.DB.Transaction(func(tx *gorm.DB) error {
				repo.Ref = href

				// create repository data
				if err := tx.Create(&repo).Error; err != nil {
					return err
				}

				// create measurement
				ghMeasure.RepoID = repo.ID
				ghMeasure.Repo = repo
				if err := tx.Create(&ghMeasure).Error; err != nil {
					return err
				}

				// create uses data
				for i := 0; i <= len(ghUses); i++ {
					ghUses[i].GitHubActionMeasureID = ghMeasure.ID
					ghUses[i].GitHubActionMeasure = ghMeasure
				}
				if err := tx.Create(&ghUses).Error; err != nil {
					return err
				}

				return nil // commit the whole transaction
			})
		}
	} else if config.NOW.Sub(repo.UpdatedAt) > config.UPDATE_DIFF {
		// TODO update
		fmt.Printf("update %s\n", href)
	}
}

func analyzeRepoByGit(href string) (models.GitHubActionMeasure, []models.GitHubActionUses, error) {
	// clone git repo to memory
	fs := memfs.New()

	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: "https://github.com" + href,
	}); err != nil { // clone error, fast return
		return models.GitHubActionMeasure{}, nil, err
	}

	// declare variables
	var ghMeasure models.GitHubActionMeasure
	var ghUses []models.GitHubActionUses

	// analyze workflows
	workflows, err := fs.ReadDir(".github/workflows")
	if err != nil {
		return models.GitHubActionMeasure{}, nil, err
	}

	for _, entry := range workflows {
		// if dir / not yml / not yaml skip
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}

		// open yaml configuration file
		yml, err := fs.Open(".github/workflows/" + entry.Name())
		if err != nil {
			return models.GitHubActionMeasure{}, nil, err
		}

		// parse workflow from yaml file
		w := models.Workflow{}
		if err := yaml.NewDecoder(yml).Decode(&w); err != nil {
			return models.GitHubActionMeasure{}, nil, err
		}

		// map result from workflow to measure / uses

	}

	return ghMeasure, ghUses, nil
}
