package github

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
	"github.com/rpothin/gh-template/internal/config"
)

// contentResponse maps the GitHub Contents API response for a single file.
type contentResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	Name     string `json:"name"`
}

// FetchManifestFromRepo fetches template-metadata.yml from the root of owner/repo
// via the GitHub Contents API and parses it into a Manifest.
func FetchManifestFromRepo(client *gogithub.RESTClient, owner, repo string) (*config.Manifest, error) {
	var resp contentResponse
	if err := client.Get(
		fmt.Sprintf("repos/%s/%s/contents/template-metadata.yml", url.PathEscape(owner), url.PathEscape(repo)),
		&resp,
	); err != nil {
		return nil, fmt.Errorf("fetching template-metadata.yml from %s/%s: %w", owner, repo, err)
	}

	if resp.Type != "file" {
		return nil, fmt.Errorf("template-metadata.yml in %s/%s is not a file (type: %s)", owner, repo, resp.Type)
	}

	// GitHub base64-encodes content with embedded newlines; strip them before decoding.
	cleaned := strings.ReplaceAll(resp.Content, "\n", "")
	raw, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("decoding template-metadata.yml from %s/%s: %w", owner, repo, err)
	}

	manifest, err := config.ParseManifestBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing template-metadata.yml from %s/%s: %w", owner, repo, err)
	}

	return manifest, nil
}
