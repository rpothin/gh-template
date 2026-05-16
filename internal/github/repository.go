package github

import (
	"fmt"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
)

// RepoInfo represents the GitHub API response for a repository.
type RepoInfo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	Description                string           `json:"description"`
	Private                    bool             `json:"private"`
	Visibility                 string           `json:"visibility"`
	HTMLURL                    string           `json:"html_url"`
	HasIssues                  bool             `json:"has_issues"`
	HasProjects                bool             `json:"has_projects"`
	HasWiki                    bool             `json:"has_wiki"`
	HasDiscussions             bool             `json:"has_discussions"`
	HasPullRequests            bool             `json:"has_pull_requests"`
	PullRequestCreationPolicy  string           `json:"pull_request_creation_policy"`
	AllowSquashMerge           bool             `json:"allow_squash_merge"`
	AllowMergeCommit           bool             `json:"allow_merge_commit"`
	AllowRebaseMerge           bool             `json:"allow_rebase_merge"`
	AllowAutoMerge             bool             `json:"allow_auto_merge"`
	DeleteBranchOnMerge        bool             `json:"delete_branch_on_merge"`
	AllowUpdateBranch          bool             `json:"allow_update_branch"`
	SecurityAndAnalysis        *SecurityAnalysis `json:"security_and_analysis"`
}

// GetRepository fetches repository metadata from the GitHub API.
func GetRepository(client *gogithub.RESTClient, owner, repo string) (*RepoInfo, error) {
	var result RepoInfo
	if err := client.Get(fmt.Sprintf("repos/%s/%s", owner, repo), &result); err != nil {
		return nil, fmt.Errorf("fetching repository %s/%s: %w", owner, repo, err)
	}
	return &result, nil
}

// repoUpdatePayload builds a map of only the non-nil fields from RepoSettings.
func repoUpdatePayload(s config.RepoSettings) map[string]interface{} {
	p := make(map[string]interface{})
	if s.HasWiki != nil {
		p["has_wiki"] = *s.HasWiki
	}
	if s.HasIssues != nil {
		p["has_issues"] = *s.HasIssues
	}
	if s.HasProjects != nil {
		p["has_projects"] = *s.HasProjects
	}
	if s.HasDiscussions != nil {
		p["has_discussions"] = *s.HasDiscussions
	}
	if s.HasPullRequests != nil {
		p["has_pull_requests"] = *s.HasPullRequests
	}
	if s.PullRequestCreationPolicy != "" {
		p["pull_request_creation_policy"] = s.PullRequestCreationPolicy
	}
	if s.AllowSquashMerge != nil {
		p["allow_squash_merge"] = *s.AllowSquashMerge
	}
	if s.AllowMergeCommit != nil {
		p["allow_merge_commit"] = *s.AllowMergeCommit
	}
	if s.AllowRebaseMerge != nil {
		p["allow_rebase_merge"] = *s.AllowRebaseMerge
	}
	if s.AllowAutoMerge != nil {
		p["allow_auto_merge"] = *s.AllowAutoMerge
	}
	if s.DeleteBranchOnMerge != nil {
		p["delete_branch_on_merge"] = *s.DeleteBranchOnMerge
	}
	if s.AllowUpdateBranch != nil {
		p["allow_update_branch"] = *s.AllowUpdateBranch
	}
	if s.Visibility != "" {
		p["visibility"] = s.Visibility
		if s.Visibility == "private" {
			p["private"] = true
		} else {
			p["private"] = false
		}
	}
	if s.Description != "" {
		p["description"] = s.Description
	}
	return p
}

// UpdateRepository applies repository settings via the GitHub API.
func UpdateRepository(client *gogithub.RESTClient, owner, repo string, settings config.RepoSettings) error {
	payload := repoUpdatePayload(settings)
	if len(payload) == 0 {
		return nil
	}
	body, err := jsonBody(payload)
	if err != nil {
		return fmt.Errorf("encoding repository settings for %s/%s: %w", owner, repo, err)
	}
	var result interface{}
	if err := client.Patch(fmt.Sprintf("repos/%s/%s", owner, repo), body, &result); err != nil {
		return fmt.Errorf("updating repository settings for %s/%s: %w", owner, repo, err)
	}
	return nil
}

// TopicsResponse is the GitHub API response for repository topics.
type TopicsResponse struct {
	Names []string `json:"names"`
}

// GetTopics fetches the topics of a repository.
func GetTopics(client *gogithub.RESTClient, owner, repo string) ([]string, error) {
	var result TopicsResponse
	if err := client.Get(fmt.Sprintf("repos/%s/%s/topics", owner, repo), &result); err != nil {
		return nil, fmt.Errorf("fetching topics for %s/%s: %w", owner, repo, err)
	}
	return result.Names, nil
}

// SetTopics replaces the topics of a repository.
func SetTopics(client *gogithub.RESTClient, owner, repo string, topics []string) error {
	if topics == nil {
		topics = []string{}
	}
	payload := map[string]interface{}{"names": topics}
	body, err := jsonBody(payload)
	if err != nil {
		return fmt.Errorf("encoding topics for %s/%s: %w", owner, repo, err)
	}
	var result interface{}
	if err := client.Put(fmt.Sprintf("repos/%s/%s/topics", owner, repo), body, &result); err != nil {
		return fmt.Errorf("setting topics for %s/%s: %w", owner, repo, err)
	}
	return nil
}

// GenerateRepoPayload is the request body for creating a repo from a template.
type GenerateRepoPayload struct {
	Owner              string `json:"owner"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Private            bool   `json:"private"`
	IncludeAllBranches bool   `json:"include_all_branches"`
}

// CreateFromTemplate creates a new repository from a template repository.
func CreateFromTemplate(client *gogithub.RESTClient, templateOwner, templateRepo, newOwner, newName string, private bool) (*RepoInfo, error) {
	payload := GenerateRepoPayload{
		Owner:   newOwner,
		Name:    newName,
		Private: private,
	}
	body, err := jsonBody(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding template repository payload for %s/%s: %w", templateOwner, templateRepo, err)
	}
	var result RepoInfo
	if err := client.Post(
		fmt.Sprintf("repos/%s/%s/generate", templateOwner, templateRepo),
		body,
		&result,
	); err != nil {
		return nil, fmt.Errorf("creating repository from template %s/%s: %w", templateOwner, templateRepo, err)
	}
	return &result, nil
}

// GetAuthenticatedUser returns the login of the authenticated user.
func GetAuthenticatedUser(client *gogithub.RESTClient) (string, error) {
	var result struct {
		Login string `json:"login"`
	}
	if err := client.Get("user", &result); err != nil {
		return "", fmt.Errorf("fetching authenticated user: %w", err)
	}
	return result.Login, nil
}

// RepoInfoToSettings converts a RepoInfo to a config.RepoSettings with pointer fields.
func RepoInfoToSettings(info *RepoInfo) config.RepoSettings {
	return config.RepoSettings{
		HasWiki:                   boolPtr(info.HasWiki),
		HasIssues:                 boolPtr(info.HasIssues),
		HasProjects:               boolPtr(info.HasProjects),
		HasDiscussions:            boolPtr(info.HasDiscussions),
		HasPullRequests:           boolPtr(info.HasPullRequests),
		PullRequestCreationPolicy: info.PullRequestCreationPolicy,
		AllowSquashMerge:          boolPtr(info.AllowSquashMerge),
		AllowMergeCommit:          boolPtr(info.AllowMergeCommit),
		AllowRebaseMerge:          boolPtr(info.AllowRebaseMerge),
		AllowAutoMerge:            boolPtr(info.AllowAutoMerge),
		DeleteBranchOnMerge:       boolPtr(info.DeleteBranchOnMerge),
		AllowUpdateBranch:         boolPtr(info.AllowUpdateBranch),
		Visibility:                info.Visibility,
		Description:               info.Description,
	}
}

func boolPtr(b bool) *bool { return &b }
