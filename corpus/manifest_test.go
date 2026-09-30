package corpus_test

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

type manifest struct {
	Version int          `json:"version"`
	Cases   []caseRecord `json:"cases"`
}

type caseRecord struct {
	ID            string                     `json:"id"`
	Repository    string                     `json:"repository"`
	Base          string                     `json:"base"`
	Head          string                     `json:"head"`
	Paths         []string                   `json:"paths"`
	License       string                     `json:"license"`
	Category      string                     `json:"category"`
	Rule          string                     `json:"rule"`
	SemanticValue bool                       `json:"semanticValue"`
	Confidence    string                     `json:"confidence"`
	Before        map[string]json.RawMessage `json:"before"`
	After         map[string]json.RawMessage `json:"after"`
	Rationale     string                     `json:"rationale"`
}

func loadManifest(t *testing.T) manifest {
	t.Helper()

	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var corpus manifest
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	return corpus
}

func TestManifest_has_frozen_case_inventory(t *testing.T) {
	// Given
	corpus := loadManifest(t)
	wantIDs := map[string]bool{
		"uniswap-monad-additions":              true,
		"optimism-cbeth-base-goerli-addition":  true,
		"wormhole-token-list-elon-retarget":    true,
		"uniswap-unichain-lsk-removal":         true,
		"uniswap-excluded-token-conflict":      true,
		"megaeth-canonical-case-normalization": true,
		"wormhole-prime-layout-move":           true,
		"wormhole-layerzero-usdc-price-only":   true,
	}

	// When
	gotIDs := make(map[string]bool, len(corpus.Cases))
	for _, record := range corpus.Cases {
		gotIDs[record.ID] = true
	}

	// Then
	if corpus.Version != 1 || len(corpus.Cases) != 8 {
		t.Fatalf("want version 1 with 8 cases, got version %d with %d", corpus.Version, len(corpus.Cases))
	}
	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("duplicate or missing IDs: got %v", gotIDs)
	}
	for id := range wantIDs {
		if !gotIDs[id] {
			t.Errorf("missing case %q", id)
		}
	}
}

func TestManifest_meets_gate_categories(t *testing.T) {
	// Given
	corpus := loadManifest(t)
	counts := map[string]int{}
	semanticValue := 0

	// When
	for _, record := range corpus.Cases {
		counts[record.Category]++
		if record.SemanticValue {
			semanticValue++
		}
	}

	// Then
	if counts["addition_migration"] < 2 {
		t.Errorf("addition/migration cases = %d, want at least 2", counts["addition_migration"])
	}
	if counts["review"] < 2 {
		t.Errorf("review cases = %d, want at least 2", counts["review"])
	}
	if counts["negative_control"] < 1 {
		t.Errorf("negative controls = %d, want at least 1", counts["negative_control"])
	}
	if semanticValue < 2 {
		t.Errorf("semantic-value cases = %d, want at least 2", semanticValue)
	}
}

func TestManifest_records_are_reproducible(t *testing.T) {
	// Given
	corpus := loadManifest(t)
	sha := regexp.MustCompile(`^[0-9a-f]{40}$`)
	licenses := map[string]bool{"Apache-2.0": true, "GPL-3.0-or-later": true, "MIT": true}
	confidences := map[string]bool{"high": true, "medium": true}

	// When / Then
	for _, record := range corpus.Cases {
		t.Run(record.ID, func(t *testing.T) {
			if record.Repository == "" || len(record.Paths) == 0 || record.Rationale == "" {
				t.Error("repository, source paths, and rationale must be present")
			}
			if !sha.MatchString(record.Base) || !sha.MatchString(record.Head) || record.Base == record.Head {
				t.Errorf("invalid immutable comparison %q..%q", record.Base, record.Head)
			}
			if !licenses[record.License] {
				t.Errorf("unapproved license %q", record.License)
			}
			if !confidences[record.Confidence] {
				t.Errorf("invalid confidence %q", record.Confidence)
			}
			if record.Rule == "" || len(record.Before) == 0 || len(record.After) == 0 {
				t.Error("semantic rule and before/after facts must be present")
			}
		})
	}
}
