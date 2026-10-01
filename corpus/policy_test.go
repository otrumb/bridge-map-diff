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
	for _, pattern := range []string{"*.go", "*.json", "*.yml", "*.yaml", "*.md", "LICENSE"} {
		if !lines[pattern+" text eol=lf"] {
			t.Errorf("missing LF policy for %s", pattern)
		}
	}
}
