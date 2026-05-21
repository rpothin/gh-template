package github

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
)

// ─── path helpers ─────────────────────────────────────────────────────────────

// contentPath encodes each slash-separated segment of a repository-relative
// path individually, preserving the "/" separator so that the GitHub Contents
// API routes correctly.  Using url.PathEscape on the entire string would encode
// "/" as "%2F", which breaks the API routing.
func contentPath(p string) string {
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}

// normalizePath strips leading and trailing slashes from a path.
func normalizePath(p string) string {
	return strings.Trim(p, "/")
}

// validateCommonFilePath returns an error when path is unsuitable for use as a
// common_files entry: empty, absolute, contains ".." traversal, or references
// the ".git" directory.
func validateCommonFilePath(p string) error {
	if p == "" {
		return fmt.Errorf("common_files path must not be empty")
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("common_files path must be relative, got %q", p)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return fmt.Errorf("common_files path must not contain '..', got %q", p)
		}
		if seg == ".git" {
			return fmt.Errorf("common_files path must not reference .git, got %q", p)
		}
	}
	return nil
}

// ─── error helpers ────────────────────────────────────────────────────────────

// isNotFound returns true when err is an HTTP 404 response from the GitHub API.
func isNotFound(err error) bool {
	var httpErr *gogithub.HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == 404
}

// ─── Contents API types ───────────────────────────────────────────────────────

// fileContent is the GitHub Contents API response shape for a single file.
type fileContent struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// dirEntry is a single item in a Contents API directory listing response.
type dirEntry struct {
	Type string `json:"type"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
}

// ─── Contents API helpers ─────────────────────────────────────────────────────

// getFileOrDir fetches a path from the GitHub Contents API.
// It returns exactly one non-nil result:
//   - (*fileContent, nil) when the path is a file
//   - (nil, []dirEntry)  when the path is a directory
//
// The file-vs-directory distinction is made by inspecting the first non-space
// byte of the JSON response ("{" = file object, "[" = directory array).
func getFileOrDir(client *gogithub.RESTClient, owner, repo, path string) (*fileContent, []dirEntry, error) {
	apiPath := fmt.Sprintf(
		"repos/%s/%s/contents/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		contentPath(path),
	)

	var raw json.RawMessage
	if err := client.Get(apiPath, &raw); err != nil {
		return nil, nil, fmt.Errorf("fetching contents of %s from %s/%s: %w", path, owner, repo, err)
	}

	if strings.HasPrefix(strings.TrimSpace(string(raw)), "[") {
		var entries []dirEntry
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, nil, fmt.Errorf("parsing directory listing for %s in %s/%s: %w", path, owner, repo, err)
		}
		return nil, entries, nil
	}

	var fc fileContent
	if err := json.Unmarshal(raw, &fc); err != nil {
		return nil, nil, fmt.Errorf("parsing file content for %s in %s/%s: %w", path, owner, repo, err)
	}
	if fc.Encoding != "base64" {
		return nil, nil, fmt.Errorf("unexpected encoding %q for %s in %s/%s (want base64)", fc.Encoding, path, owner, repo)
	}
	if fc.Content == "" {
		return nil, nil, fmt.Errorf("file %s in %s/%s is empty or too large to retrieve via the Contents API", path, owner, repo)
	}
	return &fc, nil, nil
}

// collectLeafFiles recursively expands path into all descendant file entries.
// Symlinks and submodule entries are silently skipped.
func collectLeafFiles(client *gogithub.RESTClient, owner, repo, path string) ([]fileContent, error) {
	fc, entries, err := getFileOrDir(client, owner, repo, path)
	if err != nil {
		return nil, err
	}
	if fc != nil {
		return []fileContent{*fc}, nil
	}

	var files []fileContent
	for _, entry := range entries {
		switch entry.Type {
		case "file":
			child, _, err := getFileOrDir(client, owner, repo, entry.Path)
			if err != nil {
				return nil, err
			}
			files = append(files, *child)
		case "dir":
			sub, err := collectLeafFiles(client, owner, repo, entry.Path)
			if err != nil {
				return nil, err
			}
			files = append(files, sub...)
		}
	}
	return files, nil
}

// getTargetSHA returns the current git-blob SHA of a file in the target repo,
// or ("", nil) when the file does not exist (404).  All other errors are
// returned unchanged so the caller can decide whether to abort.
func getTargetSHA(client *gogithub.RESTClient, owner, repo, path string) (string, error) {
	apiPath := fmt.Sprintf(
		"repos/%s/%s/contents/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		contentPath(path),
	)
	var result fileContent
	if err := client.Get(apiPath, &result); err != nil {
		if isNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("checking existing file %s in %s/%s: %w", path, owner, repo, err)
	}
	return result.SHA, nil
}

// putFile creates or updates a file in the target repository.
// existingSHA must be the current git-blob SHA when updating; empty when creating.
// branch specifies the branch to commit to.
func putFile(client *gogithub.RESTClient, owner, repo, path string, content []byte, existingSHA, commitMsg, branch string) error {
	encoded := base64.StdEncoding.EncodeToString(content)
	payload := map[string]interface{}{
		"message": commitMsg,
		"content": encoded,
		"branch":  branch,
	}
	if existingSHA != "" {
		payload["sha"] = existingSHA
	}

	body, err := jsonBody(payload)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf(
		"repos/%s/%s/contents/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		contentPath(path),
	)
	var result interface{}
	if err := client.Put(apiPath, body, &result); err != nil {
		return fmt.Errorf("writing file %s to %s/%s on branch %s: %w", path, owner, repo, branch, err)
	}
	return nil
}

// ─── Branch management ────────────────────────────────────────────────────────

type gitRef struct {
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

// EnsureBranch creates branch on owner/repo if it does not already exist,
// branching from the repository's default branch HEAD.
// Returns the branch name for use in downstream calls.
func EnsureBranch(client *gogithub.RESTClient, owner, repo, branch string) error {
	checkPath := fmt.Sprintf(
		"repos/%s/%s/git/ref/heads/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		contentPath(branch),
	)
	var existing gitRef
	if err := client.Get(checkPath, &existing); err == nil {
		return nil // branch already exists
	} else if !isNotFound(err) {
		return fmt.Errorf("checking branch %s on %s/%s: %w", branch, owner, repo, err)
	}

	// Resolve the SHA of the default branch HEAD.
	defaultBranch, err := GetDefaultBranch(client, owner, repo)
	if err != nil {
		return err
	}
	var defaultRef gitRef
	if err := client.Get(
		fmt.Sprintf(
			"repos/%s/%s/git/ref/heads/%s",
			url.PathEscape(owner),
			url.PathEscape(repo),
			contentPath(defaultBranch),
		),
		&defaultRef,
	); err != nil {
		return fmt.Errorf("resolving default branch %s on %s/%s: %w", defaultBranch, owner, repo, err)
	}

	body, err := jsonBody(map[string]interface{}{
		"ref": "refs/heads/" + branch,
		"sha": defaultRef.Object.SHA,
	})
	if err != nil {
		return err
	}
	var result interface{}
	if err := client.Post(
		fmt.Sprintf("repos/%s/%s/git/refs", url.PathEscape(owner), url.PathEscape(repo)),
		body,
		&result,
	); err != nil {
		return fmt.Errorf("creating branch %s on %s/%s: %w", branch, owner, repo, err)
	}
	return nil
}

// ─── High-level sync ─────────────────────────────────────────────────────────

// SyncCommonFiles copies the files and directories listed in paths from
// srcOwner/srcRepo to dstOwner/dstRepo on the specified branch.
//
// Each path is relative to the repository root.  Directory paths are expanded
// recursively.  Files whose git-blob SHA matches the source are skipped to
// avoid empty commits.
//
// Errors for individual files are collected and returned alongside the list of
// successfully synced file paths; the caller decides whether to abort or
// continue.
func SyncCommonFiles(
	client *gogithub.RESTClient,
	srcOwner, srcRepo,
	dstOwner, dstRepo string,
	paths []string,
	branch string,
) (synced []string, errs []error) {
	commitMsg := fmt.Sprintf("chore: sync common files from %s/%s", srcOwner, srcRepo)

	for _, rawPath := range paths {
		p := normalizePath(rawPath)
		if err := validateCommonFilePath(p); err != nil {
			errs = append(errs, err)
			continue
		}

		files, err := collectLeafFiles(client, srcOwner, srcRepo, p)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, f := range files {
			cleaned := strings.ReplaceAll(f.Content, "\n", "")
			decoded, err := base64.StdEncoding.DecodeString(cleaned)
			if err != nil {
				errs = append(errs, fmt.Errorf("decoding %s from %s/%s: %w", f.Path, srcOwner, srcRepo, err))
				continue
			}

			existingSHA, err := getTargetSHA(client, dstOwner, dstRepo, f.Path)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			// Skip if the git-blob SHAs match — content is identical.
			if existingSHA != "" && existingSHA == f.SHA {
				continue
			}

			if err := putFile(client, dstOwner, dstRepo, f.Path, decoded, existingSHA, commitMsg, branch); err != nil {
				errs = append(errs, err)
				continue
			}

			synced = append(synced, f.Path)
		}
	}

	return synced, errs
}
