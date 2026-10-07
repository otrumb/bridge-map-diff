package corpus_test

import (
	"os"
	"strings"
	"testing"
)

func TestRepository_enforces_LF_for_release_text(t *testing.T) {
	// Given
	data, err := os.ReadFile("../.gitattributes")
	if err != nil {
		t.Fatal(err)
	}
	lines := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		lines[line] = true
	}

	// When / Then
	for _, pattern := range []string{"*.go", "*.json", "*.yml", "*.yaml", "*.md", "LICENSE", ".gitattributes", "*.golden"} {
		if !lines[pattern+" text eol=lf"] {
			t.Errorf("missing LF policy for %s", pattern)
		}
	}
}

func TestReleaseWorkflowBindsRemoteTagToValidatedCommit(t *testing.T) {
	// Given
	data, err := os.ReadFile("../.github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)

	// When / Then
	required := []string{
		"group: release-${{ github.ref }}",
		"cancel-in-progress: false",
		`git ls-remote origin "refs/tags/${GITHUB_REF_NAME}^{}"`,
		`test "$remote_target" = "$GITHUB_SHA"`,
		`gh release create "$GITHUB_REF_NAME"`,
	}
	for _, contract := range required {
		if !strings.Contains(workflow, contract) {
			t.Errorf("release workflow missing contract %q", contract)
		}
	}

	tagCheck := strings.Index(workflow, `test "$remote_target" = "$GITHUB_SHA"`)
	publish := strings.Index(workflow, `gh release create "$GITHUB_REF_NAME"`)
	if tagCheck < 0 || publish < 0 || tagCheck > publish {
		t.Error("remote tag check must run before release creation")
	}
}
