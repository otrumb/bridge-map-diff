package diff_test

import (
	"encoding/json"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/diff"
	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func TestCompare_distinguishes_unknown_decimals_from_known_zero(t *testing.T) {
	// Given
	before := mapping("canonical", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
	after := before
	after.Origin.Decimals = nil

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{before}}, snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{after}})

	// Then
	if len(report.Findings) != 1 || report.Findings[0].Code != "origin_decimals_changed" || report.Findings[0].After != "unknown" {
		t.Fatalf("findings = %#v", report.Findings)
	}
}

func TestCompare_reports_removal_and_deactivation_as_review(t *testing.T) {
	// Given
	removed := mapping("canonical", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
	deactivated := mapping("alternate", "0x3333333333333333333333333333333333333333", "0x4444444444444444444444444444444444444444")
	inactive := deactivated
	active := false
	inactive.Active = &active

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{removed, deactivated}}, snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{inactive}})

	// Then
	if len(report.Findings) != 2 || report.Findings[0].Code != "deactivated" || report.Findings[1].Code != "removal" {
		t.Fatalf("findings = %#v", report.Findings)
	}
	if report.Summary.Review != 2 {
		t.Fatalf("summary = %#v", report.Summary)
	}
}

func TestCompare_duplicate_active_identities_are_order_independent_in_both_snapshots(t *testing.T) {
	// Given
	first := mapping("canonical", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
	second := first
	second.Destination.Address = "0x3333333333333333333333333333333333333333"
	forward := []snapshot.Mapping{first, second}
	reverse := []snapshot.Mapping{second, first}

	// When
	left := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: forward}, snapshot.Snapshot{Version: 1, Mappings: forward})
	right := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: reverse}, snapshot.Snapshot{Version: 1, Mappings: reverse})
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)

	// Then
	if string(leftJSON) != string(rightJSON) {
		t.Fatalf("forward = %s\nreverse = %s", leftJSON, rightJSON)
	}
	if len(left.Findings) != 2 || left.Findings[0].Code != "duplicate_active_mapping_after" || left.Findings[1].Code != "duplicate_active_mapping_before" {
		t.Fatalf("findings = %#v", left.Findings)
	}
}

func TestCompare_invalid_migration_does_not_suppress_churn(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*snapshot.Mapping, snapshot.Mapping)
	}{
		{"predecessor remains active", func(replacement *snapshot.Mapping, old snapshot.Mapping) {}},
		{"incompatible route", func(replacement *snapshot.Mapping, old snapshot.Mapping) { replacement.Route = "alternate" }},
		{"self reference", func(replacement *snapshot.Mapping, old snapshot.Mapping) {
			replacement.Supersedes = replacement.Identity()
		}},
		{"cycle", func(replacement *snapshot.Mapping, old snapshot.Mapping) {}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			old := mapping("canonical", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
			replacement := mapping("canonical", "0x3333333333333333333333333333333333333333", "0x4444444444444444444444444444444444444444")
			replacement.Supersedes = old.Identity()
			test.mutate(&replacement, old)
			if test.name == "cycle" {
				old.Supersedes = replacement.Identity()
			}
			afterMappings := []snapshot.Mapping{replacement}
			if test.name == "predecessor remains active" {
				afterMappings = append(afterMappings, old)
			}

			// When
			report := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{old}}, snapshot.Snapshot{Version: 1, Mappings: afterMappings})

			// Then
			codes := findingCodes(report)
			if !codes["invalid_migration"] || !codes["addition"] {
				t.Fatalf("findings = %#v", report.Findings)
			}
			if test.name != "predecessor remains active" && !codes["removal"] {
				t.Fatalf("findings = %#v, want removal", report.Findings)
			}
		})
	}
}

func TestCompare_duplicate_predecessor_claims_and_cycles_are_invalid(t *testing.T) {
	// Given
	old := mapping("canonical", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
	first := mapping("canonical", "0x3333333333333333333333333333333333333333", "0x4444444444444444444444444444444444444444")
	second := mapping("canonical", "0x5555555555555555555555555555555555555555", "0x6666666666666666666666666666666666666666")
	first.Supersedes = old.Identity()
	second.Supersedes = old.Identity()

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{old}}, snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{first, second}})

	// Then
	codes := findingCodes(report)
	if !codes["invalid_migration"] || !codes["removal"] || !codes["addition"] || codes["migration"] {
		t.Fatalf("findings = %#v", report.Findings)
	}
}

func findingCodes(report diff.Report) map[string]bool {
	codes := map[string]bool{}
	for _, finding := range report.Findings {
		codes[finding.Code] = true
	}
	return codes
}
