package github

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
)

// TemplateSummary is a lightweight representation of a template repository
// used by the list and search commands.
type TemplateSummary struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	StarCount   int    `json:"star_count,omitempty"`
}

// visibilityOf returns a display-friendly visibility string, falling back to
// the private boolean when the visibility field is absent.
func visibilityOf(vis string, private bool) string {
	if vis != "" {
		return vis
	}
	if private {
		return "private"
	}
	return "public"
}

// --- list user-owned template repos -----------------------------------------

// userRepoPage is the per-item shape returned by GET /user/repos.
type userRepoPage struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	Private     bool   `json:"private"`
	IsTemplate  bool   `json:"is_template"`
	Archived    bool   `json:"archived"`
}

// ListOwnedTemplateRepos fetches all template repositories owned by the
// authenticated user (type=owner, all pages). When includeArchived is false,
// archived repositories are excluded from the results.
func ListOwnedTemplateRepos(client *gogithub.RESTClient, includeArchived bool) ([]TemplateSummary, error) {
	const perPage = 100
	var results []TemplateSummary

	for page := 1; ; page++ {
		var repos []userRepoPage
		path := fmt.Sprintf("user/repos?type=owner&per_page=%d&page=%d", perPage, page)
		if err := client.Get(path, &repos); err != nil {
			return nil, fmt.Errorf("listing user repositories (page %d): %w", page, err)
		}
		for _, r := range repos {
			if r.IsTemplate && (includeArchived || !r.Archived) {
				results = append(results, TemplateSummary{
					FullName:    r.FullName,
					Description: r.Description,
					Visibility:  visibilityOf(r.Visibility, r.Private),
				})
			}
		}
		if len(repos) < perPage {
			break
		}
	}
	return results, nil
}

// --- list org template repos -------------------------------------------------

// orgSummary is the per-item shape returned by GET /user/orgs.
type orgSummary struct {
	Login string `json:"login"`
}

// ListOrgTemplateRepos fetches template repositories from all organisations the
// authenticated user belongs to. It uses a bounded pool of 4 concurrent
// workers. If fetching repos for an specific org fails, warningFn is called
// with the error message and that org is skipped. When includeArchived is
// false, archived repositories are excluded from the results.
func ListOrgTemplateRepos(client *gogithub.RESTClient, includeArchived bool, warningFn func(string)) ([]TemplateSummary, error) {
	var orgs []orgSummary
	if err := client.Get("user/orgs?per_page=100", &orgs); err != nil {
		return nil, fmt.Errorf("listing user organisations: %w", err)
	}

	if len(orgs) == 0 {
		return nil, nil
	}

	type orgResult struct {
		repos []TemplateSummary
		err   error
		org   string
	}

	const workers = 4
	jobs := make(chan string, len(orgs))
	results := make(chan orgResult, len(orgs))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for orgLogin := range jobs {
				repos, err := fetchOrgTemplates(client, orgLogin, includeArchived)
				results <- orgResult{repos: repos, err: err, org: orgLogin}
			}
		}()
	}

	for _, o := range orgs {
		jobs <- o.Login
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []TemplateSummary
	for res := range results {
		if res.err != nil {
			if warningFn != nil {
				warningFn(fmt.Sprintf("skipping org %s: %v", res.org, res.err))
			}
			continue
		}
		all = append(all, res.repos...)
	}
	return all, nil
}

func fetchOrgTemplates(client *gogithub.RESTClient, orgLogin string, includeArchived bool) ([]TemplateSummary, error) {
	const perPage = 100
	var results []TemplateSummary
	for page := 1; ; page++ {
		var repos []userRepoPage
		path := fmt.Sprintf("orgs/%s/repos?type=all&per_page=%d&page=%d", orgLogin, perPage, page)
		if err := client.Get(path, &repos); err != nil {
			return nil, err
		}
		for _, r := range repos {
			if r.IsTemplate && (includeArchived || !r.Archived) {
				results = append(results, TemplateSummary{
					FullName:    r.FullName,
					Description: r.Description,
					Visibility:  visibilityOf(r.Visibility, r.Private),
				})
			}
		}
		if len(repos) < perPage {
			break
		}
	}
	return results, nil
}

// --- search public template repos --------------------------------------------

// searchResponse is the top-level shape of GET /search/repositories.
type searchResponse struct {
	TotalCount        int              `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items             []searchRepoItem `json:"items"`
}

type searchRepoItem struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	Private     bool   `json:"private"`
	StarCount   int    `json:"stargazers_count"`
	Archived    bool   `json:"archived"`
}

// SearchTemplateRepos searches public template repositories using the GitHub
// Search API. query is appended to the mandatory "template:true" qualifier.
// When includeArchived is false, "archived:false" is added to the query so the
// API filters archived repositories server-side.
// If incomplete_results is returned by the API, warnFn (if non-nil) is called.
func SearchTemplateRepos(client *gogithub.RESTClient, query string, limit int, includeArchived bool, warnFn func(string)) ([]TemplateSummary, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	q := "template:true"
	if !includeArchived {
		q += " archived:false"
	}
	if trimmed := strings.TrimSpace(query); trimmed != "" {
		q += " " + trimmed
	}

	path := fmt.Sprintf(
		"search/repositories?q=%s&sort=stars&order=desc&per_page=%d",
		url.QueryEscape(q),
		limit,
	)

	var resp searchResponse
	if err := client.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("searching template repositories: %w", err)
	}

	if resp.IncompleteResults && warnFn != nil {
		warnFn("search results may be incomplete (GitHub index timeout)")
	}

	results := make([]TemplateSummary, 0, len(resp.Items))
	for _, item := range resp.Items {
		results = append(results, TemplateSummary{
			FullName:    item.FullName,
			Description: item.Description,
			Visibility:  visibilityOf(item.Visibility, item.Private),
			StarCount:   item.StarCount,
		})
	}
	return results, nil
}
