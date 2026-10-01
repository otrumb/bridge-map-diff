package diff_test

import (
	"fmt"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/diff"
	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func mapping(route, origin, destination string) snapshot.Mapping {
	return snapshot.Mapping{Registry: "r", Route: route, Origin: snapshot.Endpoint{Chain: "ethereum", Kind: "evm", Address: origin, Decimals: decimals(18)}, Destination: snapshot.Endpoint{Chain: "base", Kind: "evm", Address: destination, Decimals: decimals(18)}, Symbol: "TOK"}
}

func decimals(value uint8) *uint8 { return &value }

func TestCompare_preserves_review_summary_when_findings_are_truncated(t *testing.T) {
	// Given
	mappings := make([]snapshot.Mapping, 10001)
	for index := range mappings {
		mappings[index] = snapshot.Mapping{
			Registry: "r", Route: "x",
			Origin:      snapshot.Endpoint{Chain: "a", Kind: "opaque", Address: fmt.Sprintf("origin-%05d", index), Decimals: decimals(1)},
			Destination: snapshot.Endpoint{Chain: fmt.Sprintf("b-%05d", index), Kind: "opaque", Address: "destination", Decimals: decimals(1)},
		}
	}
	mappings[10000].Origin = snapshot.Endpoint{Chain: "ethereum", Kind: "evm", Address: "0x0000000000000000000000000000000000000000", Decimals: decimals(18)}

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1}, snapshot.Snapshot{Version: 1, Mappings: mappings})

	// Then
	if len(report.Findings) != 10000 {
		t.Fatalf("bounded findings = %d, want 10000", len(report.Findings))
	}
	if report.Summary.Review != 1 || report.Summary.Changes != 10002 {
		t.Fatalf("summary = %#v, want full 10002 findings including review", report.Summary)
	}
}

func TestCompare_reports_semantic_changes(t *testing.T) {
	// Given
	origin := "0x1111111111111111111111111111111111111111"
	oldDestination := "0x2222222222222222222222222222222222222222"
	newDestination := "0x3333333333333333333333333333333333333333"
	before := snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{mapping("canonical", origin, oldDestination)}}
	afterMapping := mapping("canonical", origin, newDestination)
	afterMapping.Origin.Decimals = decimals(6)
	afterMapping.Destination.Decimals = decimals(8)
	afterMapping.Symbol = "NEW"
	after := snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{afterMapping}}

	// When
	report := diff.Compare(before, after)

	// Then
	want := []string{"destination_changed", "destination_decimals_changed", "origin_decimals_changed", "symbol_changed"}
	if len(report.Findings) != len(want) {
		t.Fatalf("findings = %#v", report.Findings)
	}
	for index, code := range want {
		if report.Findings[index].Code != code {
			t.Fatalf("finding %d code = %q, want %q", index, report.Findings[index].Code, code)
		}
	}
}

func TestCompare_groups_only_explicit_supersedes_as_migration(t *testing.T) {
	// Given
	old := mapping("wormhole", "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222")
	replacement := mapping("wormhole", "0x3333333333333333333333333333333333333333", "0x4444444444444444444444444444444444444444")
	replacement.Supersedes = old.Identity()

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{old}}, snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{replacement}})

	// Then
	if len(report.Findings) != 1 || report.Findings[0].Code != "migration" {
		t.Fatalf("findings = %#v, want migration only", report.Findings)
	}
}

func TestCompare_flags_collision_duplicate_zero_and_unscoped_parallel(t *testing.T) {
	// Given
	zero := "0x0000000000000000000000000000000000000000"
	destination := "0x2222222222222222222222222222222222222222"
	first := mapping("canonical", zero, destination)
	second := mapping("canonical", "0x1111111111111111111111111111111111111111", destination)
	parallel := mapping("alternate", zero, destination)
	after := snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{first, first, second, parallel}}

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1}, after)

	// Then
	codes := map[string]bool{}
	for _, finding := range report.Findings {
		codes[finding.Code] = true
	}
	for _, code := range []string{"destination_collision", "duplicate_active_mapping_after", "unscoped_parallel", "zero_address"} {
		if !codes[code] {
			t.Errorf("missing %q in %#v", code, report.Findings)
		}
	}
}

func TestCompare_preserves_route_aware_fanout(t *testing.T) {
	// Given
	origin := "0x1111111111111111111111111111111111111111"
	base := mapping("canonical", origin, "0x2222222222222222222222222222222222222222")
	arbitrum := base
	arbitrum.Destination.Chain = "arbitrum"
	arbitrum.Destination.Address = "0x3333333333333333333333333333333333333333"

	// When
	report := diff.Compare(snapshot.Snapshot{Version: 1}, snapshot.Snapshot{Version: 1, Mappings: []snapshot.Mapping{base, arbitrum}})

	// Then
	if len(report.Findings) != 2 || report.Findings[0].Code != "addition" || report.Findings[1].Code != "addition" {
		t.Fatalf("findings = %#v, want two route-aware additions", report.Findings)
	}
}
