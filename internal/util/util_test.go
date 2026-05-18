package util

import (
	"testing"
)

func TestParseOwnerRepo_Valid(t *testing.T) {
	owner, repo, err := ParseOwnerRepo("acme/my-repo")
	if err != nil {
		t.Fatalf("ParseOwnerRepo() error = %v, want nil", err)
	}
	if owner != "acme" {
		t.Errorf("owner = %q, want %q", owner, "acme")
	}
	if repo != "my-repo" {
		t.Errorf("repo = %q, want %q", repo, "my-repo")
	}
}

func TestParseOwnerRepo_MultiSlash(t *testing.T) {
	// SplitN(s, "/", 2) intentionally keeps anything after the first slash
	// as part of the repo name — callers that pass "a/b/c" get repo="b/c",
	// and the GitHub API will reject the resulting URL. This preserves the
	// original input for error messages rather than silently dropping parts.
	owner, repo, err := ParseOwnerRepo("owner/repo/extra")
	if err != nil {
		t.Fatalf("ParseOwnerRepo() error = %v, want nil (multi-slash passthrough)", err)
	}
	if owner != "owner" {
		t.Errorf("owner = %q, want %q", owner, "owner")
	}
	if repo != "repo/extra" {
		t.Errorf("repo = %q, want %q", repo, "repo/extra")
	}
}

func TestParseOwnerRepo_NoSlash(t *testing.T) {
	_, _, err := ParseOwnerRepo("noslash")
	if err == nil {
		t.Fatal("ParseOwnerRepo() error = nil, want error for missing slash")
	}
}

func TestParseOwnerRepo_EmptyOwner(t *testing.T) {
	_, _, err := ParseOwnerRepo("/repo")
	if err == nil {
		t.Fatal("ParseOwnerRepo() error = nil, want error for empty owner")
	}
}

func TestParseOwnerRepo_EmptyRepo(t *testing.T) {
	_, _, err := ParseOwnerRepo("owner/")
	if err == nil {
		t.Fatal("ParseOwnerRepo() error = nil, want error for empty repo")
	}
}

func TestParseOwnerRepo_Empty(t *testing.T) {
	_, _, err := ParseOwnerRepo("")
	if err == nil {
		t.Fatal("ParseOwnerRepo() error = nil, want error for empty string")
	}
}
