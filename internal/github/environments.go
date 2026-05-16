package github

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
	"golang.org/x/crypto/nacl/box"
)

// ─── API response types ───────────────────────────────────────────────────────

type environmentsListResponse struct {
	TotalCount   int            `json:"total_count"`
	Environments []envListEntry `json:"environments"`
}

type envListEntry struct {
	Name                   string                `json:"name"`
	PreventSelfReview      bool                  `json:"prevent_self_review"`
	DeploymentBranchPolicy *branchPolicyResponse `json:"deployment_branch_policy"`
	ProtectionRules        []protectionRule      `json:"protection_rules"`
}

type protectionRule struct {
	ID        int              `json:"id"`
	Type      string           `json:"type"`
	WaitTimer int              `json:"wait_timer"`
	Reviewers []reviewerInRule `json:"reviewers"`
}

type reviewerInRule struct {
	Type     string          `json:"type"`
	Reviewer reviewerDetails `json:"reviewer"`
}

type reviewerDetails struct {
	Login        string `json:"login"`
	Slug         string `json:"slug"`
	Organization struct {
		Login string `json:"login"`
	} `json:"organization"`
}

type branchPolicyResponse struct {
	ProtectedBranches    bool `json:"protected_branches"`
	CustomBranchPolicies bool `json:"custom_branch_policies"`
}

type branchPoliciesListResponse struct {
	TotalCount     int            `json:"total_count"`
	BranchPolicies []branchPolicy `json:"branch_policies"`
}

type branchPolicy struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type variablesListResponse struct {
	TotalCount int                `json:"total_count"`
	Variables  []variableResponse `json:"variables"`
}

type variableResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type secretsListResponse struct {
	TotalCount int              `json:"total_count"`
	Secrets    []secretResponse `json:"secrets"`
}

type secretResponse struct {
	Name string `json:"name"`
}

type publicKeyResponse struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
}

// ─── API payload types ────────────────────────────────────────────────────────

type reviewerRef struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetEnvironments fetches all environments for a repository with full detail,
// including reviewers, variables, secret names, and branch policies.
func GetEnvironments(client *gogithub.RESTClient, owner, repo string) ([]config.Environment, error) {
	var result environmentsListResponse
	if err := client.Get(fmt.Sprintf("repos/%s/%s/environments", owner, repo), &result); err != nil {
		return nil, fmt.Errorf("fetching environments for %s/%s: %w", owner, repo, err)
	}

	type envResult struct {
		index int
		env   config.Environment
		err   error
	}

	ch := make(chan envResult, len(result.Environments))
	for i, e := range result.Environments {
		i, e := i, e
		go func() {
			env, err := buildFullEnvironment(client, owner, repo, e)
			ch <- envResult{i, env, err}
		}()
	}

	envs := make([]config.Environment, len(result.Environments))
	for range result.Environments {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		envs[r.index] = r.env
	}
	return envs, nil
}

// buildFullEnvironment constructs a complete config.Environment from an envListEntry
// by making additional concurrent API calls for variables, secret names, and branch patterns.
func buildFullEnvironment(client *gogithub.RESTClient, owner, repo string, e envListEntry) (config.Environment, error) {
	prevent := e.PreventSelfReview
	env := config.Environment{
		Name:              e.Name,
		PreventSelfReview: &prevent,
	}

	for _, rule := range e.ProtectionRules {
		switch rule.Type {
		case "wait_timer":
			env.WaitTimer = rule.WaitTimer
		case "required_reviewers":
			for _, r := range rule.Reviewers {
				switch r.Type {
				case "User":
					env.Reviewers = append(env.Reviewers, r.Reviewer.Login)
				case "Team":
					env.Reviewers = append(env.Reviewers,
						r.Reviewer.Organization.Login+"/"+r.Reviewer.Slug)
				}
			}
		}
	}

	switch {
	case e.DeploymentBranchPolicy == nil:
		env.DeploymentBranchPolicy = "all"
	case e.DeploymentBranchPolicy.ProtectedBranches:
		env.DeploymentBranchPolicy = "protected"
	default:
		env.DeploymentBranchPolicy = "custom"
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		fetchErr error
	)
	setErr := func(err error) {
		mu.Lock()
		if fetchErr == nil {
			fetchErr = err
		}
		mu.Unlock()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		vars, err := getEnvironmentVariables(client, owner, repo, e.Name)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		env.Variables = vars
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		secrets, err := getEnvironmentSecretNames(client, owner, repo, e.Name)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		env.Secrets = secrets
		mu.Unlock()
	}()

	if env.DeploymentBranchPolicy == "custom" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			patterns, err := getDeploymentBranchPatterns(client, owner, repo, e.Name)
			if err != nil {
				setErr(err)
				return
			}
			mu.Lock()
			env.DeploymentBranchPatterns = patterns
			mu.Unlock()
		}()
	}

	wg.Wait()
	if fetchErr != nil {
		return config.Environment{}, fetchErr
	}
	return env, nil
}

func getDeploymentBranchPatterns(client *gogithub.RESTClient, owner, repo, envName string) ([]string, error) {
	var result branchPoliciesListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/deployment-branch-policies", owner, repo, envName),
		&result,
	); err != nil {
		return nil, fmt.Errorf("fetching branch policies for environment %q: %w", envName, err)
	}
	patterns := make([]string, 0, len(result.BranchPolicies))
	for _, p := range result.BranchPolicies {
		patterns = append(patterns, p.Name)
	}
	return patterns, nil
}

func getEnvironmentVariables(client *gogithub.RESTClient, owner, repo, envName string) ([]config.EnvironmentVariable, error) {
	var result variablesListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/variables", owner, repo, envName),
		&result,
	); err != nil {
		return nil, fmt.Errorf("fetching variables for environment %q: %w", envName, err)
	}
	if len(result.Variables) == 0 {
		return nil, nil
	}
	vars := make([]config.EnvironmentVariable, len(result.Variables))
	for i, v := range result.Variables {
		vars[i] = config.EnvironmentVariable{Name: v.Name, Value: v.Value}
	}
	return vars, nil
}

func getEnvironmentSecretNames(client *gogithub.RESTClient, owner, repo, envName string) ([]config.EnvironmentSecret, error) {
	var result secretsListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/secrets", owner, repo, envName),
		&result,
	); err != nil {
		return nil, fmt.Errorf("fetching secrets for environment %q: %w", envName, err)
	}
	if len(result.Secrets) == 0 {
		return nil, nil
	}
	secrets := make([]config.EnvironmentSecret, len(result.Secrets))
	for i, s := range result.Secrets {
		secrets[i] = config.EnvironmentSecret{Name: s.Name, Value: config.SecretPlaceholder}
	}
	return secrets, nil
}

// ─── Write ────────────────────────────────────────────────────────────────────

// CreateOrUpdateEnvironment creates or updates a deployment environment with all
// configured protection rules, variables, and secrets. It returns any warnings
// (e.g. secrets initialized with a placeholder value) alongside any error.
func CreateOrUpdateEnvironment(client *gogithub.RESTClient, owner, repo string, env config.Environment) (warnings []string, err error) {
	// Resolve reviewer usernames/team-slugs to numeric IDs.
	reviewerRefs, err := resolveReviewers(client, env.Reviewers)
	if err != nil {
		return nil, err
	}

	// Build protection-rules payload (full replacement via PUT).
	preventSelfReview := false
	if env.PreventSelfReview != nil {
		preventSelfReview = *env.PreventSelfReview
	}
	payload := map[string]interface{}{
		"wait_timer":          env.WaitTimer,
		"prevent_self_review": preventSelfReview,
		"reviewers":           reviewerRefs,
	}
	switch env.DeploymentBranchPolicy {
	case "protected":
		payload["deployment_branch_policy"] = map[string]interface{}{
			"protected_branches":    true,
			"custom_branch_policies": false,
		}
	case "custom":
		payload["deployment_branch_policy"] = map[string]interface{}{
			"protected_branches":    false,
			"custom_branch_policies": true,
		}
	default:
		payload["deployment_branch_policy"] = nil
	}

	body, err := jsonBody(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding environment %q: %w", env.Name, err)
	}
	var putResult interface{}
	if err := client.Put(
		fmt.Sprintf("repos/%s/%s/environments/%s", owner, repo, env.Name),
		body,
		&putResult,
	); err != nil {
		return nil, fmt.Errorf("creating/updating environment %q: %w", env.Name, err)
	}

	// Run post-PUT reconciliation steps concurrently.
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		warns   []string
		postErr error
	)
	addWarn := func(w string) {
		mu.Lock()
		warns = append(warns, w)
		mu.Unlock()
	}
	setPostErr := func(e error) {
		mu.Lock()
		if postErr == nil {
			postErr = e
		}
		mu.Unlock()
	}

	if env.DeploymentBranchPolicy == "custom" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := reconcileCustomBranchPatterns(client, owner, repo, env.Name, env.DeploymentBranchPatterns); err != nil {
				setPostErr(err)
			}
		}()
	}

	if len(env.Variables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := applyEnvironmentVariables(client, owner, repo, env.Name, env.Variables); err != nil {
				setPostErr(err)
			}
		}()
	}

	if len(env.Secrets) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w, err := applyEnvironmentSecrets(client, owner, repo, env.Name, env.Secrets)
			if err != nil {
				setPostErr(err)
				return
			}
			for _, warn := range w {
				addWarn(warn)
			}
		}()
	}

	wg.Wait()
	if postErr != nil {
		return nil, postErr
	}
	return warns, nil
}

func resolveReviewers(client *gogithub.RESTClient, refs []string) ([]reviewerRef, error) {
	if len(refs) == 0 {
		return []reviewerRef{}, nil
	}

	type refResult struct {
		index int
		ref   reviewerRef
		err   error
	}

	ch := make(chan refResult, len(refs))
	for i, ref := range refs {
		i, ref := i, ref
		go func() {
			r, err := resolveReviewer(client, ref)
			ch <- refResult{i, r, err}
		}()
	}

	resolved := make([]reviewerRef, len(refs))
	for range refs {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		resolved[r.index] = r.ref
	}
	return resolved, nil
}

func resolveReviewer(client *gogithub.RESTClient, ref string) (reviewerRef, error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 1 {
		var user struct {
			ID int `json:"id"`
		}
		if err := client.Get(fmt.Sprintf("users/%s", ref), &user); err != nil {
			return reviewerRef{}, fmt.Errorf("resolving user reviewer %q: %w", ref, err)
		}
		return reviewerRef{Type: "User", ID: user.ID}, nil
	}
	org, team := parts[0], parts[1]
	var teamResp struct {
		ID int `json:"id"`
	}
	if err := client.Get(fmt.Sprintf("orgs/%s/teams/%s", org, team), &teamResp); err != nil {
		return reviewerRef{}, fmt.Errorf("resolving team reviewer %q: %w", ref, err)
	}
	return reviewerRef{Type: "Team", ID: teamResp.ID}, nil
}

func reconcileCustomBranchPatterns(client *gogithub.RESTClient, owner, repo, envName string, desired []string) error {
	var result branchPoliciesListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/deployment-branch-policies", owner, repo, envName),
		&result,
	); err != nil {
		return fmt.Errorf("fetching branch policies for environment %q: %w", envName, err)
	}

	liveByName := make(map[string]int, len(result.BranchPolicies))
	for _, p := range result.BranchPolicies {
		liveByName[p.Name] = p.ID
	}
	desiredSet := make(map[string]struct{}, len(desired))
	for _, p := range desired {
		desiredSet[p] = struct{}{}
	}

	for _, pattern := range desired {
		if _, exists := liveByName[pattern]; exists {
			continue
		}
		body, err := jsonBody(map[string]interface{}{"name": pattern})
		if err != nil {
			return err
		}
		var created interface{}
		if err := client.Post(
			fmt.Sprintf("repos/%s/%s/environments/%s/deployment-branch-policies", owner, repo, envName),
			body, &created,
		); err != nil {
			return fmt.Errorf("creating branch policy pattern %q in environment %q: %w", pattern, envName, err)
		}
	}

	for name, id := range liveByName {
		if _, wanted := desiredSet[name]; wanted {
			continue
		}
		if err := client.Delete(
			fmt.Sprintf("repos/%s/%s/environments/%s/deployment-branch-policies/%d", owner, repo, envName, id),
			nil,
		); err != nil {
			return fmt.Errorf("deleting branch policy pattern %q from environment %q: %w", name, envName, err)
		}
	}
	return nil
}

func applyEnvironmentVariables(client *gogithub.RESTClient, owner, repo, envName string, vars []config.EnvironmentVariable) error {
	var existing variablesListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/variables", owner, repo, envName),
		&existing,
	); err != nil {
		return fmt.Errorf("fetching variables for environment %q: %w", envName, err)
	}

	liveVarNames := make(map[string]struct{}, len(existing.Variables))
	for _, v := range existing.Variables {
		liveVarNames[v.Name] = struct{}{}
	}

	for _, v := range vars {
		body, err := jsonBody(map[string]interface{}{"name": v.Name, "value": v.Value})
		if err != nil {
			return err
		}
		var result interface{}
		if _, exists := liveVarNames[v.Name]; exists {
			if err := client.Patch(
				fmt.Sprintf("repos/%s/%s/environments/%s/variables/%s", owner, repo, envName, v.Name),
				body, &result,
			); err != nil {
				return fmt.Errorf("updating variable %q in environment %q: %w", v.Name, envName, err)
			}
		} else {
			if err := client.Post(
				fmt.Sprintf("repos/%s/%s/environments/%s/variables", owner, repo, envName),
				body, &result,
			); err != nil {
				return fmt.Errorf("creating variable %q in environment %q: %w", v.Name, envName, err)
			}
		}
	}
	return nil
}

func applyEnvironmentSecrets(client *gogithub.RESTClient, owner, repo, envName string, secrets []config.EnvironmentSecret) (warnings []string, err error) {
	var existing secretsListResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/secrets", owner, repo, envName),
		&existing,
	); err != nil {
		return nil, fmt.Errorf("fetching secrets for environment %q: %w", envName, err)
	}

	liveSecretNames := make(map[string]struct{}, len(existing.Secrets))
	for _, s := range existing.Secrets {
		liveSecretNames[s.Name] = struct{}{}
	}

	// Fetch public key only if we have secrets to create.
	var pk *publicKeyResponse
	for _, s := range secrets {
		if _, exists := liveSecretNames[s.Name]; !exists {
			key, err := getEnvironmentPublicKey(client, owner, repo, envName)
			if err != nil {
				return nil, err
			}
			pk = &key
			break
		}
	}

	for _, s := range secrets {
		if _, exists := liveSecretNames[s.Name]; exists {
			continue // leave existing secret untouched
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
			fmt.Sprintf("repos/%s/%s/environments/%s/secrets/%s", owner, repo, envName, s.Name),
			body, &result,
		); err != nil {
			return nil, fmt.Errorf("creating secret %q in environment %q: %w", s.Name, envName, err)
		}
		warnings = append(warnings, fmt.Sprintf("⚠  Secret %q initialized with placeholder value — update it before use in workflows", s.Name))
	}
	return warnings, nil
}

func getEnvironmentPublicKey(client *gogithub.RESTClient, owner, repo, envName string) (publicKeyResponse, error) {
	var pk publicKeyResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/environments/%s/secrets/public-key", owner, repo, envName),
		&pk,
	); err != nil {
		return publicKeyResponse{}, fmt.Errorf("fetching public key for environment %q: %w", envName, err)
	}
	return pk, nil
}

func encryptSecret(publicKeyBase64, value string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return "", fmt.Errorf("decoding public key: %w", err)
	}
	var recipientKey [32]byte
	copy(recipientKey[:], keyBytes)
	encrypted, err := box.SealAnonymous(nil, []byte(value), &recipientKey, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("encrypting value: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
