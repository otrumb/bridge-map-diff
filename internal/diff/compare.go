package diff

import (
	"slices"
	"strconv"

	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func Compare(before, after snapshot.Snapshot) Report {
	oldByID, oldDuplicates := uniqueActiveByIdentity(before.Mappings)
	newByID, newDuplicates := uniqueActiveByIdentity(after.Mappings)
	consumedOld := map[string]bool{}
	consumedNew := map[string]bool{}
	findings := duplicateFindings(oldDuplicates, "before")
	findings = append(findings, duplicateFindings(newDuplicates, "after")...)
	findings = append(findings, migrations(oldByID, newByID, after.Mappings, consumedOld, consumedNew)...)

	for identity, oldMapping := range oldByID {
		newMapping, found := newByID[identity]
		if consumedOld[identity] {
			continue
		}
		if !found {
			code := "removal"
			if hasInactiveIdentity(after.Mappings, identity) {
				code = "deactivated"
			}
			findings = append(findings, Finding{Severity: "review", Code: code, Identity: identity, Before: endpointValue(oldMapping.Destination)})
			continue
		}
		findings = append(findings, changed(oldMapping, newMapping)...)
		consumedNew[identity] = true
	}
	for identity, mapping := range newByID {
		if !consumedNew[identity] {
			findings = append(findings, Finding{Severity: "info", Code: "addition", Identity: identity, After: endpointValue(mapping.Destination)})
		}
	}
	findings = append(findings, integrityFindings(after.Mappings)...)
	slices.SortFunc(findings, func(left, right Finding) int {
		if left.Code != right.Code {
			return compareString(left.Code, right.Code)
		}
		if left.Identity != right.Identity {
			return compareString(left.Identity, right.Identity)
		}
		return compareString(left.After, right.After)
	})
	return newReport(findings)
}

func uniqueActiveByIdentity(mappings []snapshot.Mapping) (map[string]snapshot.Mapping, map[string]int) {
	buckets := make(map[string][]snapshot.Mapping, len(mappings))
	for _, mapping := range mappings {
		if mapping.IsActive() {
			buckets[mapping.Identity()] = append(buckets[mapping.Identity()], mapping)
		}
	}
	unique := make(map[string]snapshot.Mapping, len(buckets))
	duplicates := map[string]int{}
	for identity, mappings := range buckets {
		if len(mappings) == 1 {
			unique[identity] = mappings[0]
		} else {
			duplicates[identity] = len(mappings)
		}
	}
	return unique, duplicates
}

func migrations(oldByID, newByID map[string]snapshot.Mapping, after []snapshot.Mapping, consumedOld, consumedNew map[string]bool) []Finding {
	var findings []Finding
	claims := map[string]int{}
	for _, replacement := range newByID {
		if replacement.Supersedes != "" {
			claims[replacement.Supersedes]++
		}
	}
	for identity, replacement := range newByID {
		oldMapping, found := oldByID[replacement.Supersedes]
		if replacement.Supersedes == "" {
			continue
		}
		valid := found && identity != replacement.Supersedes && !migrationCycle(identity, oldMapping, oldByID) && claims[replacement.Supersedes] == 1 && !hasActiveIdentity(after, replacement.Supersedes) && compatibleMigration(oldMapping, replacement)
		if !valid {
			findings = append(findings, Finding{Severity: "review", Code: "invalid_migration", Identity: identity, Before: replacement.Supersedes})
			continue
		}
		findings = append(findings, Finding{Severity: "review", Code: "migration", Identity: identity, Before: oldMapping.Identity(), After: identity})
		consumedOld[replacement.Supersedes] = true
		consumedNew[identity] = true
	}
	return findings
}

func changed(before, after snapshot.Mapping) []Finding {
	var findings []Finding
	if before.Destination.Address != after.Destination.Address {
		findings = append(findings, Finding{Severity: "review", Code: "destination_changed", Identity: before.Identity(), Before: before.Destination.Address, After: after.Destination.Address})
	}
	if decimalsValue(before.Destination.Decimals) != decimalsValue(after.Destination.Decimals) {
		findings = append(findings, Finding{Severity: "review", Code: "destination_decimals_changed", Identity: before.Identity(), Before: decimalsValue(before.Destination.Decimals), After: decimalsValue(after.Destination.Decimals)})
	}
	if decimalsValue(before.Origin.Decimals) != decimalsValue(after.Origin.Decimals) {
		findings = append(findings, Finding{Severity: "review", Code: "origin_decimals_changed", Identity: before.Identity(), Before: decimalsValue(before.Origin.Decimals), After: decimalsValue(after.Origin.Decimals)})
	}
	if before.Symbol != after.Symbol {
		findings = append(findings, Finding{Severity: "info", Code: "symbol_changed", Identity: before.Identity(), Before: before.Symbol, After: after.Symbol})
	}
	return findings
}

func decimalsValue(value *uint8) string {
	if value == nil {
		return "unknown"
	}
	return strconv.Itoa(int(*value))
}

func hasActiveIdentity(mappings []snapshot.Mapping, identity string) bool {
	for _, mapping := range mappings {
		if mapping.IsActive() && mapping.Identity() == identity {
			return true
		}
	}
	return false
}

func hasInactiveIdentity(mappings []snapshot.Mapping, identity string) bool {
	for _, mapping := range mappings {
		if !mapping.IsActive() && mapping.Identity() == identity {
			return true
		}
	}
	return false
}

func compatibleMigration(before, after snapshot.Mapping) bool {
	return before.Registry == after.Registry && before.Scope == after.Scope && before.Route == after.Route && before.Destination.Chain == after.Destination.Chain
}

func migrationCycle(replacementID string, predecessor snapshot.Mapping, oldByID map[string]snapshot.Mapping) bool {
	seen := map[string]bool{}
	for predecessor.Supersedes != "" {
		if predecessor.Supersedes == replacementID || seen[predecessor.Supersedes] {
			return true
		}
		seen[predecessor.Supersedes] = true
		next, found := oldByID[predecessor.Supersedes]
		if !found {
			return false
		}
		predecessor = next
	}
	return false
}

func duplicateFindings(duplicates map[string]int, snapshotName string) []Finding {
	findings := make([]Finding, 0, len(duplicates))
	for identity, count := range duplicates {
		findings = append(findings, Finding{Severity: "review", Code: "duplicate_active_mapping_" + snapshotName, Identity: identity, After: strconv.Itoa(count)})
	}
	return findings
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
