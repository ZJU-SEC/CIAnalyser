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

		gh_measure, gh_uses, err := analyzeRepoByGit(href)
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
				gh_measure.RepoID = repo.ID
				gh_measure.Repo = repo
				if err := tx.Create(&gh_measure).Error; err != nil {
					return err
				}

				// create uses data
				for i := 0; i <= len(gh_uses); i++ {
					gh_uses[i].GitHubActionMeasureID = gh_measure.ID
					gh_uses[i].GitHubActionMeasure = gh_measure
				}
				if err := tx.Create(&gh_uses).Error; err != nil {
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

	git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: "https://github.com" + href,
	})

	// analyze workflows
	workflows, err := fs.ReadDir(".github/workflows")
	for _, file := range workflows {
		// if dir / not yml / not yaml skip
		if file.IsDir() ||
			filepath.Ext(file.Name()) != "yml" ||
			filepath.Ext(file.Name()) != "yaml" {
			continue
		}
		// TODO
	}

	if err != nil {
		fmt.Println("[ERROR] clone", href, "ended with", err)
	}

	return models.GitHubActionMeasure{}, nil, nil
}
