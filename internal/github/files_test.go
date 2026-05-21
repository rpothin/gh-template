package github

import (
	"testing"

	gogithub "github.com/cli/go-gh/v2/pkg/api"
)

// makeHTTPError creates a *gogithub.HTTPError with the given status code for testing.
func makeHTTPError(statusCode int) error {
	return &gogithub.HTTPError{StatusCode: statusCode}
}

// ─── contentPath ──────────────────────────────────────────────────────────────

func TestContentPath_Simple(t *testing.T) {
	got := contentPath("AGENTS.md")
	if got != "AGENTS.md" {
		t.Errorf("contentPath(%q) = %q, want %q", "AGENTS.md", got, "AGENTS.md")
	}
}

func TestContentPath_Nested(t *testing.T) {
	got := contentPath(".github/workflows/ci.yml")
	if got != ".github/workflows/ci.yml" {
		t.Errorf("contentPath(%q) = %q, want %q", ".github/workflows/ci.yml", got, ".github/workflows/ci.yml")
	}
}

func TestContentPath_SpecialChars(t *testing.T) {
	// Spaces and brackets in filenames should be percent-encoded per segment.
	got := contentPath("docs/my file.md")
	want := "docs/my%20file.md"
	if got != want {
		t.Errorf("contentPath(%q) = %q, want %q", "docs/my file.md", got, want)
	}
}

func TestContentPath_PreservesSlashes(t *testing.T) {
	// Slashes must not be encoded as %2F — that would break the Contents API routing.
	got := contentPath("a/b/c")
	if got != "a/b/c" {
		t.Errorf("contentPath(%q) = %q, want %q", "a/b/c", got, "a/b/c")
	}
}

// ─── normalizePath ────────────────────────────────────────────────────────────

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"AGENTS.md", "AGENTS.md"},
		{"/AGENTS.md", "AGENTS.md"},
		{"AGENTS.md/", "AGENTS.md"},
		{"/AGENTS.md/", "AGENTS.md"},
		{".github/workflows/", ".github/workflows"},
		{".github/workflows", ".github/workflows"},
	}
	for _, tc := range cases {
		got := normalizePath(tc.input)
		if got != tc.want {
			t.Errorf("normalizePath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── validateCommonFilePath ───────────────────────────────────────────────────

func TestValidateCommonFilePath_Valid(t *testing.T) {
	valid := []string{
		"AGENTS.md",
		".github/workflows",
		".github/workflows/ci.yml",
		"docs/skills",
	}
	for _, p := range valid {
		if err := validateCommonFilePath(p); err != nil {
			t.Errorf("validateCommonFilePath(%q) unexpected error: %v", p, err)
		}
	}
}

func TestValidateCommonFilePath_Invalid(t *testing.T) {
	cases := []struct {
		input   string
		wantSub string
	}{
		{"", "must not be empty"},
		{"/absolute/path", "must be relative"},
		{"../traversal", "must not contain '..'"},
		{".github/../etc/passwd", "must not contain '..'"},
		{".git/config", "must not reference .git"},
		{".git", "must not reference .git"},
	}
	for _, tc := range cases {
		err := validateCommonFilePath(tc.input)
		if err == nil {
			t.Errorf("validateCommonFilePath(%q) = nil, want error containing %q", tc.input, tc.wantSub)
			continue
		}
		if msg := err.Error(); len(msg) == 0 {
			t.Errorf("validateCommonFilePath(%q) returned empty error string", tc.input)
		}
	}
}

// ─── isNotFound ───────────────────────────────────────────────────────────────

func TestIsNotFound_WithNilError(t *testing.T) {
	if isNotFound(nil) {
		t.Error("isNotFound(nil) = true, want false")
	}
}

func TestIsNotFound_WithNon404HTTPError(t *testing.T) {
	err := makeHTTPError(403)
	if isNotFound(err) {
		t.Errorf("isNotFound(403 error) = true, want false")
	}
}

func TestIsNotFound_With404HTTPError(t *testing.T) {
	err := makeHTTPError(404)
	if !isNotFound(err) {
		t.Errorf("isNotFound(404 error) = false, want true")
	}
}
