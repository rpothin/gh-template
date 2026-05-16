package github

import (
	"fmt"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
)

// environmentsResponse is the GitHub API response for listing environments.
type environmentsResponse struct {
	TotalCount   int `json:"total_count"`
	Environments []struct {
		Name            string `json:"name"`
		ProtectionRules []struct {
			Type      string `json:"type"`
			WaitTimer int    `json:"wait_timer"`
		} `json:"protection_rules"`
	} `json:"environments"`
}

// GetEnvironments fetches all environments for a repository.
func GetEnvironments(client *gogithub.RESTClient, owner, repo string) ([]config.Environment, error) {
	var result environmentsResponse
	if err := client.Get(fmt.Sprintf("repos/%s/%s/environments", owner, repo), &result); err != nil {
		return nil, fmt.Errorf("fetching environments for %s/%s: %w", owner, repo, err)
	}

	envs := make([]config.Environment, 0, len(result.Environments))
	for _, e := range result.Environments {
		env := config.Environment{Name: e.Name}
		for _, rule := range e.ProtectionRules {
			if rule.Type == "wait_timer" {
				env.WaitTimer = rule.WaitTimer
				break
			}
		}
		envs = append(envs, env)
	}
	return envs, nil
}

// environmentCreatePayload is the request body for creating/updating an environment.
type environmentCreatePayload struct {
	WaitTimer int `json:"wait_timer"`
}

// CreateOrUpdateEnvironment creates or updates a deployment environment.
func CreateOrUpdateEnvironment(client *gogithub.RESTClient, owner, repo string, env config.Environment) error {
	payload := environmentCreatePayload{WaitTimer: env.WaitTimer}
	body, err := jsonBody(payload)
	if err != nil {
		return fmt.Errorf("encoding environment %q for %s/%s: %w", env.Name, owner, repo, err)
	}
	var result interface{}
	if err := client.Put(
		fmt.Sprintf("repos/%s/%s/environments/%s", owner, repo, env.Name),
		body,
		&result,
	); err != nil {
		return fmt.Errorf("creating/updating environment %q for %s/%s: %w", env.Name, owner, repo, err)
	}
	return nil
}
