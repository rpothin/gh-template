package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
)

// ─── API types ────────────────────────────────────────────────────────────────

// SecurityAnalysis represents the security_and_analysis object returned by
// GET /repos/{owner}/{repo}. It is nil when no Advanced Security features are
// enabled or available for the repository.
type SecurityAnalysis struct {
	DependabotSecurityUpdates    *securityAnalysisEntry `json:"dependabot_security_updates"`
	SecretScanning               *securityAnalysisEntry `json:"secret_scanning"`
	SecretScanningPushProtection *securityAnalysisEntry `json:"secret_scanning_push_protection"`
	DependencyGraph              *securityAnalysisEntry `json:"dependency_graph"`
}

type securityAnalysisEntry struct {
	Status string `json:"status"`
}

// privateVulnerabilityReportingResponse is the response body for
// GET /repos/{owner}/{repo}/private-vulnerability-reporting.
type privateVulnerabilityReportingResponse struct {
	Enabled bool `json:"enabled"`
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetVulnerabilityAlertsEnabled reports whether Dependabot vulnerability alerts
// are enabled on the repository.
// A 404 response is treated as "disabled" (not an error).
func GetVulnerabilityAlertsEnabled(client *gogithub.RESTClient, owner, repo string) (bool, error) {
	var result interface{}
	err := client.Get(
		fmt.Sprintf("repos/%s/%s/vulnerability-alerts", url.PathEscape(owner), url.PathEscape(repo)),
		&result,
	)
	if err == nil {
		return true, nil
	}

	var httpErr *gogithub.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("checking vulnerability alerts for %s/%s: %w", owner, repo, err)
}

// GetPrivateVulnerabilityReportingEnabled reports whether private vulnerability
// reporting is enabled on the repository.
// A 404 or 422 response is treated as "not enabled / not supported" (not an error).
// Private vulnerability reporting is primarily supported for public repositories.
func GetPrivateVulnerabilityReportingEnabled(client *gogithub.RESTClient, owner, repo string) (bool, error) {
	var result privateVulnerabilityReportingResponse
	err := client.Get(
		fmt.Sprintf("repos/%s/%s/private-vulnerability-reporting", url.PathEscape(owner), url.PathEscape(repo)),
		&result,
	)
	if err == nil {
		return result.Enabled, nil
	}

	var httpErr *gogithub.HTTPError
	if errors.As(err, &httpErr) && (httpErr.StatusCode == http.StatusNotFound || httpErr.StatusCode == http.StatusUnprocessableEntity) {
		return false, nil
	}
	return false, fmt.Errorf("checking private vulnerability reporting for %s/%s: %w", owner, repo, err)
}

// RepoInfoToSecurity converts the security_and_analysis portion of a RepoInfo
// plus the vulnerability-alerts and private-vulnerability-reporting enabled flags
// into a SecuritySettings value.
// Returns nil if no security data is present.
func RepoInfoToSecurity(info *RepoInfo, vulnAlertsEnabled bool, privVulnReportingEnabled bool) *config.SecuritySettings {
	s := &config.SecuritySettings{
		DependabotAlerts:              boolPtr(vulnAlertsEnabled),
		PrivateVulnerabilityReporting: boolPtr(privVulnReportingEnabled),
	}

	if info.SecurityAndAnalysis == nil {
		// Only Dependabot alerts and private vulnerability reporting status are
		// known; leave GHAS fields nil.
		return s
	}

	sa := info.SecurityAndAnalysis
	if sa.DependabotSecurityUpdates != nil {
		s.DependabotSecurityUpdates = boolPtr(sa.DependabotSecurityUpdates.Status == "enabled")
	}
	if sa.SecretScanning != nil {
		s.SecretScanning = boolPtr(sa.SecretScanning.Status == "enabled")
	}
	if sa.SecretScanningPushProtection != nil {
		s.SecretScanningPushProtection = boolPtr(sa.SecretScanningPushProtection.Status == "enabled")
	}
	if sa.DependencyGraph != nil {
		s.DependencyGraph = boolPtr(sa.DependencyGraph.Status == "enabled")
	}
	return s
}

// ─── Write ────────────────────────────────────────────────────────────────────

// UpdateSecuritySettings applies security configuration to a repository.
// Dependabot alerts are toggled via /vulnerability-alerts (PUT/DELETE).
// Private vulnerability reporting is toggled via /private-vulnerability-reporting (PUT/DELETE).
// All other fields are applied via security_and_analysis in PATCH /repos.
// Fields that are not supported for the repository type (e.g. secret scanning
// on a private repo without GitHub Advanced Security) are silently skipped by
// the GitHub API.
func UpdateSecuritySettings(client *gogithub.RESTClient, owner, repo string, settings *config.SecuritySettings) error {
	if settings == nil {
		return nil
	}

	if settings.DependabotAlerts != nil {
		if err := setVulnerabilityAlerts(client, owner, repo, *settings.DependabotAlerts); err != nil {
			return err
		}
	}

	if settings.PrivateVulnerabilityReporting != nil {
		if err := setPrivateVulnerabilityReporting(client, owner, repo, *settings.PrivateVulnerabilityReporting); err != nil {
			return err
		}
	}

	// Build security_and_analysis sub-object for PATCH /repos.
	sa := map[string]interface{}{}
	if settings.DependabotSecurityUpdates != nil {
		status := "disabled"
		if *settings.DependabotSecurityUpdates {
			status = "enabled"
		}
		sa["dependabot_security_updates"] = map[string]interface{}{"status": status}
	}
	if settings.SecretScanning != nil {
		status := "disabled"
		if *settings.SecretScanning {
			status = "enabled"
		}
		sa["secret_scanning"] = map[string]interface{}{"status": status}
	}
	if settings.SecretScanningPushProtection != nil {
		status := "disabled"
		if *settings.SecretScanningPushProtection {
			status = "enabled"
		}
		sa["secret_scanning_push_protection"] = map[string]interface{}{"status": status}
	}
	if settings.DependencyGraph != nil {
		status := "disabled"
		if *settings.DependencyGraph {
			status = "enabled"
		}
		sa["dependency_graph"] = map[string]interface{}{"status": status}
	}

	if len(sa) == 0 {
		return nil
	}

	body, err := jsonBody(map[string]interface{}{"security_and_analysis": sa})
	if err != nil {
		return err
	}
	var result interface{}
	if err := client.Patch(
		fmt.Sprintf("repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo)),
		body,
		&result,
	); err != nil {
		return fmt.Errorf("updating security settings for %s/%s: %w", owner, repo, err)
	}
	return nil
}

func setVulnerabilityAlerts(client *gogithub.RESTClient, owner, repo string, enable bool) error {
	path := fmt.Sprintf("repos/%s/%s/vulnerability-alerts", url.PathEscape(owner), url.PathEscape(repo))
	if enable {
		body, _ := jsonBody(map[string]interface{}{})
		var result interface{}
		if err := client.Put(path, body, &result); err != nil {
			return fmt.Errorf("enabling vulnerability alerts for %s/%s: %w", owner, repo, err)
		}
	} else {
		if err := client.Delete(path, nil); err != nil {
			return fmt.Errorf("disabling vulnerability alerts for %s/%s: %w", owner, repo, err)
		}
	}
	return nil
}

func setPrivateVulnerabilityReporting(client *gogithub.RESTClient, owner, repo string, enable bool) error {
	path := fmt.Sprintf("repos/%s/%s/private-vulnerability-reporting", url.PathEscape(owner), url.PathEscape(repo))
	if enable {
		body, _ := jsonBody(map[string]interface{}{})
		var result interface{}
		if err := client.Put(path, body, &result); err != nil {
			return fmt.Errorf("enabling private vulnerability reporting for %s/%s: %w", owner, repo, err)
		}
	} else {
		if err := client.Delete(path, nil); err != nil {
			return fmt.Errorf("disabling private vulnerability reporting for %s/%s: %w", owner, repo, err)
		}
	}
	return nil
}
