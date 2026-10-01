package diff

import "github.com/otrumb/bridge-map-diff/internal/snapshot"

const maxFindings = 10000

type Report struct {
	Version  int       `json:"version"`
	Summary  Summary   `json:"summary"`
	Findings []Finding `json:"findings"`
}

type Summary struct {
	Changes int `json:"changes"`
	Review  int `json:"review"`
	Info    int `json:"info"`
}

type Finding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Identity string `json:"identity"`
	Before   string `json:"before,omitempty"`
	After    string `json:"after,omitempty"`
}

func newReport(findings []Finding) Report {
	report := Report{Version: 1}
	for _, finding := range findings {
		report.Summary.Changes++
		if finding.Severity == "review" {
			report.Summary.Review++
		} else {
			report.Summary.Info++
		}
	}
	if len(findings) > maxFindings {
		findings = findings[:maxFindings]
	}
	report.Findings = findings
	return report
}

func endpointValue(endpoint snapshot.Endpoint) string {
	return endpoint.Chain + ":" + endpoint.Address
}
