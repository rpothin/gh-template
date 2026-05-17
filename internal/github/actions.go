package github

import (
	"fmt"
	"net/url"
	"sync"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
)

// ─── API response / request types ────────────────────────────────────────────

type repoActionsPermissions struct {
	Enabled            bool   `json:"enabled"`
	AllowedActions     string `json:"allowed_actions"`
	ShaPinningRequired bool   `json:"sha_pinning_required"`
}

type workflowPermissions struct {
	DefaultWorkflowPermissions   string `json:"default_workflow_permissions"`
	CanApprovePullRequestReviews bool   `json:"can_approve_pull_request_reviews"`
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetActionsPermissions fetches Actions permissions settings for a repository by
// reading both the actions/permissions and actions/permissions/workflow endpoints
// concurrently and merging the results into a single ActionsSettings value.
func GetActionsPermissions(client *gogithub.RESTClient, owner, repo string) (*config.ActionsSettings, error) {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		fetchErr error
		ap       repoActionsPermissions
		wp       workflowPermissions
	)

	setErr := func(err error) {
		mu.Lock()
		if fetchErr == nil {
			fetchErr = err
		}
		mu.Unlock()
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := client.Get(
			fmt.Sprintf("repos/%s/%s/actions/permissions", url.PathEscape(owner), url.PathEscape(repo)),
			&ap,
		); err != nil {
			setErr(fmt.Errorf("fetching actions permissions for %s/%s: %w", owner, repo, err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := client.Get(
			fmt.Sprintf("repos/%s/%s/actions/permissions/workflow", url.PathEscape(owner), url.PathEscape(repo)),
			&wp,
		); err != nil {
			setErr(fmt.Errorf("fetching workflow permissions for %s/%s: %w", owner, repo, err))
		}
	}()

	wg.Wait()
	if fetchErr != nil {
		return nil, fetchErr
	}

	return &config.ActionsSettings{
		CanApprovePullRequestReviews: boolPtr(wp.CanApprovePullRequestReviews),
		ShaPinningRequired:           boolPtr(ap.ShaPinningRequired),
		DefaultWorkflowPermissions:   wp.DefaultWorkflowPermissions,
	}, nil
}

// UpdateActionsPermissions applies Actions permissions settings to a repository.
// sha_pinning_required is applied via the actions/permissions endpoint (using a
// read-then-write to preserve existing enabled/allowed_actions values).
// can_approve_pull_request_reviews and default_workflow_permissions are applied
// via the actions/permissions/workflow endpoint.
func UpdateActionsPermissions(client *gogithub.RESTClient, owner, repo string, settings *config.ActionsSettings) error {
	if settings == nil {
		return nil
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		applyErr error
	)

	setErr := func(err error) {
		mu.Lock()
		if applyErr == nil {
			applyErr = err
		}
		mu.Unlock()
	}

	if settings.ShaPinningRequired != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := applyShaPinning(client, owner, repo, *settings.ShaPinningRequired); err != nil {
				setErr(err)
			}
		}()
	}

	if settings.CanApprovePullRequestReviews != nil || settings.DefaultWorkflowPermissions != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := applyWorkflowPermissions(client, owner, repo, settings); err != nil {
				setErr(err)
			}
		}()
	}

	wg.Wait()
	return applyErr
}

// applyShaPinning reads current actions/permissions then PUTs with the new sha_pinning_required
// value while preserving the existing enabled and allowed_actions fields.
func applyShaPinning(client *gogithub.RESTClient, owner, repo string, pinningRequired bool) error {
	var current repoActionsPermissions
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/actions/permissions", url.PathEscape(owner), url.PathEscape(repo)),
		&current,
	); err != nil {
		return fmt.Errorf("reading actions permissions for %s/%s: %w", owner, repo, err)
	}

	payload := map[string]interface{}{
		"enabled":              current.Enabled,
		"sha_pinning_required": pinningRequired,
	}
	if current.AllowedActions != "" {
		payload["allowed_actions"] = current.AllowedActions
	}

	body, err := jsonBody(payload)
	if err != nil {
		return err
	}
	var result interface{}
	if err := client.Put(
		fmt.Sprintf("repos/%s/%s/actions/permissions", url.PathEscape(owner), url.PathEscape(repo)),
		body,
		&result,
	); err != nil {
		return fmt.Errorf("updating actions permissions for %s/%s: %w", owner, repo, err)
	}
	return nil
}

// applyWorkflowPermissions PUTs can_approve and/or default_workflow_permissions.
func applyWorkflowPermissions(client *gogithub.RESTClient, owner, repo string, settings *config.ActionsSettings) error {
	payload := map[string]interface{}{}
	if settings.CanApprovePullRequestReviews != nil {
		payload["can_approve_pull_request_reviews"] = *settings.CanApprovePullRequestReviews
	}
	if settings.DefaultWorkflowPermissions != "" {
		payload["default_workflow_permissions"] = settings.DefaultWorkflowPermissions
	}
	body, err := jsonBody(payload)
	if err != nil {
		return err
	}
	var result interface{}
	if err := client.Put(
		fmt.Sprintf("repos/%s/%s/actions/permissions/workflow", url.PathEscape(owner), url.PathEscape(repo)),
		body,
		&result,
	); err != nil {
		return fmt.Errorf("updating workflow permissions for %s/%s: %w", owner, repo, err)
	}
	return nil
}

// ─── Repo-level variables ─────────────────────────────────────────────────────

func listRepoVariables(client *gogithub.RESTClient, owner, repo string) ([]variableResponse, error) {
	const perPage = 100

	var variables []variableResponse
	for page := 1; ; page++ {
		var result variablesListResponse
		if err := client.Get(
			fmt.Sprintf(
				"repos/%s/%s/actions/variables?per_page=%d&page=%d",
				url.PathEscape(owner),
				url.PathEscape(repo),
				perPage,
				page,
			),
			&result,
		); err != nil {
			return nil, fmt.Errorf("fetching Actions variables for %s/%s (page %d): %w", owner, repo, page, err)
		}
		variables = append(variables, result.Variables...)
		if len(result.Variables) < perPage {
			break
		}
	}
	return variables, nil
}

// GetRepoVariables fetches all Actions variables for a repository.
func GetRepoVariables(client *gogithub.RESTClient, owner, repo string) ([]config.EnvironmentVariable, error) {
	variables, err := listRepoVariables(client, owner, repo)
	if err != nil {
		return nil, err
	}
	if len(variables) == 0 {
		return nil, nil
	}
	vars := make([]config.EnvironmentVariable, len(variables))
	for i, v := range variables {
		vars[i] = config.EnvironmentVariable{Name: v.Name, Value: v.Value}
	}
	return vars, nil
}

// ApplyRepoVariables creates or updates repository-level Actions variables.
func ApplyRepoVariables(client *gogithub.RESTClient, owner, repo string, vars []config.EnvironmentVariable) error {
	if len(vars) == 0 {
		return nil
	}

	existingVariables, err := listRepoVariables(client, owner, repo)
	if err != nil {
		return err
	}

	liveNames := make(map[string]struct{}, len(existingVariables))
	for _, v := range existingVariables {
		liveNames[v.Name] = struct{}{}
	}

	for _, v := range vars {
		body, err := jsonBody(map[string]interface{}{"name": v.Name, "value": v.Value})
		if err != nil {
			return err
		}
		var result interface{}
		if _, exists := liveNames[v.Name]; exists {
			if err := client.Patch(
				fmt.Sprintf(
					"repos/%s/%s/actions/variables/%s",
					url.PathEscape(owner),
					url.PathEscape(repo),
					url.PathEscape(v.Name),
				),
				body,
				&result,
			); err != nil {
				return fmt.Errorf("updating variable %q for %s/%s: %w", v.Name, owner, repo, err)
			}
		} else {
			if err := client.Post(
				fmt.Sprintf("repos/%s/%s/actions/variables", url.PathEscape(owner), url.PathEscape(repo)),
				body,
				&result,
			); err != nil {
				return fmt.Errorf("creating variable %q for %s/%s: %w", v.Name, owner, repo, err)
			}
		}
	}
	return nil
}

// ─── Repo-level secrets ───────────────────────────────────────────────────────

func listRepoSecrets(client *gogithub.RESTClient, owner, repo string) ([]secretResponse, error) {
	const perPage = 100

	var secrets []secretResponse
	for page := 1; ; page++ {
		var result secretsListResponse
		if err := client.Get(
			fmt.Sprintf(
				"repos/%s/%s/actions/secrets?per_page=%d&page=%d",
				url.PathEscape(owner),
				url.PathEscape(repo),
				perPage,
				page,
			),
			&result,
		); err != nil {
			return nil, fmt.Errorf("fetching Actions secrets for %s/%s (page %d): %w", owner, repo, page, err)
		}
		secrets = append(secrets, result.Secrets...)
		if len(result.Secrets) < perPage {
			break
		}
	}
	return secrets, nil
}

// GetRepoSecretNames fetches the names of all Actions secrets for a repository.
// Secret values are never returned by the GitHub API, so Value is left empty.
func GetRepoSecretNames(client *gogithub.RESTClient, owner, repo string) ([]config.EnvironmentSecret, error) {
	secretNames, err := listRepoSecrets(client, owner, repo)
	if err != nil {
		return nil, err
	}
	if len(secretNames) == 0 {
		return nil, nil
	}
	secrets := make([]config.EnvironmentSecret, len(secretNames))
	for i, s := range secretNames {
		secrets[i] = config.EnvironmentSecret{Name: s.Name}
	}
	return secrets, nil
}

// ApplyRepoSecrets initializes any missing repository-level Actions secrets
// with a placeholder value. Existing secrets are left untouched.
// Returns a warning message for each secret initialized with a placeholder.
func ApplyRepoSecrets(client *gogithub.RESTClient, owner, repo string, secrets []config.EnvironmentSecret) (warnings []string, err error) {
	if len(secrets) == 0 {
		return nil, nil
	}

	existingSecrets, err := listRepoSecrets(client, owner, repo)
	if err != nil {
		return nil, err
	}

	liveNames := make(map[string]struct{}, len(existingSecrets))
	for _, s := range existingSecrets {
		liveNames[s.Name] = struct{}{}
	}

	var pk *publicKeyResponse
	for _, s := range secrets {
		if _, exists := liveNames[s.Name]; !exists {
			key, err := getRepoPublicKey(client, owner, repo)
			if err != nil {
				return nil, err
			}
			pk = &key
			break
		}
	}

	for _, s := range secrets {
		if _, exists := liveNames[s.Name]; exists {
			continue
		}
		encrypted, err := encryptSecret(pk.Key, config.SecretPlaceholder)
		if err != nil {
			return nil, fmt.Errorf("encrypting secret %q: %w", s.Name, err)
		}
		body, err := jsonBody(map[string]interface{}{
			"encrypted_value": encrypted,
			"key_id":          pk.KeyID,
		})
		if err != nil {
			return nil, err
		}
		var result interface{}
		if err := client.Put(
			fmt.Sprintf(
				"repos/%s/%s/actions/secrets/%s",
				url.PathEscape(owner),
				url.PathEscape(repo),
				url.PathEscape(s.Name),
			),
			body,
			&result,
		); err != nil {
			return nil, fmt.Errorf("creating secret %q for %s/%s: %w", s.Name, owner, repo, err)
		}
		warnings = append(warnings, fmt.Sprintf("⚠  Secret %q initialized with placeholder value — update it before use in workflows", s.Name))
	}
	return warnings, nil
}

func getRepoPublicKey(client *gogithub.RESTClient, owner, repo string) (publicKeyResponse, error) {
	var pk publicKeyResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/actions/secrets/public-key", url.PathEscape(owner), url.PathEscape(repo)),
		&pk,
	); err != nil {
		return publicKeyResponse{}, fmt.Errorf("fetching public key for %s/%s: %w", owner, repo, err)
	}
	return pk, nil
}
