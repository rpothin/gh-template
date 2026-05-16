package util

import (
	"fmt"
	"strings"
)

// ParseOwnerRepo splits "owner/repo" into (owner, repo).
func ParseOwnerRepo(ownerRepo string) (string, string, error) {
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format %q: expected owner/repo", ownerRepo)
	}
	return parts[0], parts[1], nil
}
