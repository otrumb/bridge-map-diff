package diff

import (
	"slices"
	"strings"

	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func integrityFindings(mappings []snapshot.Mapping) []Finding {
	active := make([]snapshot.Mapping, 0, len(mappings))
	for _, mapping := range mappings {
		if mapping.IsActive() {
			active = append(active, mapping)
		}
	}
	return append(append(collisions(active), unscopedParallel(active)...), zeroAddresses(active)...)
}

func collisions(mappings []snapshot.Mapping) []Finding {
	groups := map[string]map[string]bool{}
	for _, mapping := range mappings {
		key := mapping.Registry + "|" + mapping.Scope + "|" + mapping.Route + "|" + endpointValue(mapping.Destination)
		if groups[key] == nil {
			groups[key] = map[string]bool{}
		}
		groups[key][endpointValue(mapping.Origin)] = true
	}
	var findings []Finding
	for key, origins := range groups {
		if len(origins) > 1 {
			findings = append(findings, Finding{Severity: "review", Code: "destination_collision", Identity: key, After: sortedSet(origins)})
		}
	}
	return findings
}

func unscopedParallel(mappings []snapshot.Mapping) []Finding {
	groups := map[string]map[string]bool{}
	for _, mapping := range mappings {
		if mapping.Scope != "" {
			continue
		}
		key := mapping.Registry + "||" + endpointValue(mapping.Origin) + "|" + endpointValue(mapping.Destination)
		if groups[key] == nil {
			groups[key] = map[string]bool{}
		}
		groups[key][mapping.Route] = true
	}
	var findings []Finding
	for key, routes := range groups {
		if len(routes) > 1 {
			findings = append(findings, Finding{Severity: "review", Code: "unscoped_parallel", Identity: key, After: sortedSet(routes)})
		}
	}
	return findings
}

func zeroAddresses(mappings []snapshot.Mapping) []Finding {
	var findings []Finding
	for _, mapping := range mappings {
		if isZeroEVM(mapping.Origin) {
			findings = append(findings, Finding{Severity: "review", Code: "zero_address", Identity: mapping.Identity(), After: "origin"})
		}
		if isZeroEVM(mapping.Destination) {
			findings = append(findings, Finding{Severity: "review", Code: "zero_address", Identity: mapping.Identity(), After: "destination"})
		}
	}
	return findings
}

func isZeroEVM(endpoint snapshot.Endpoint) bool {
	return endpoint.Kind == "evm" && endpoint.Address == "0x"+strings.Repeat("0", 40)
}

func sortedSet(values map[string]bool) string {
	items := make([]string, 0, len(values))
	for value := range values {
		items = append(items, value)
	}
	slices.Sort(items)
	return strings.Join(items, ",")
}
