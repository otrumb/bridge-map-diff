package corpus_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/diff"
	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

type replayCorpus struct {
	Version int          `json:"version"`
	Cases   []replayCase `json:"cases"`
}

type replayCase struct {
	ID               string          `json:"id"`
	Fidelity         string          `json:"fidelity"`
	Excluded         string          `json:"excluded"`
	DestinationChain string          `json:"destinationChain"`
	Before           json.RawMessage `json:"before"`
	After            json.RawMessage `json:"after"`
	AdapterBefore    json.RawMessage `json:"adapterBefore"`
	AdapterAfter     json.RawMessage `json:"adapterAfter"`
	Want             []string        `json:"want"`
}

func TestReplay_matches_frozen_semantic_outcomes(t *testing.T) {
	// Given
	data, err := os.ReadFile("replay.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus replayCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Version != 1 || len(corpus.Cases) != 8 {
		t.Fatalf("replay inventory = version %d, %d cases", corpus.Version, len(corpus.Cases))
	}

	for _, record := range corpus.Cases {
		t.Run(record.ID, func(t *testing.T) {
			if record.Excluded != "" {
				if record.Before != nil || record.After != nil {
					t.Fatal("excluded case must not contain invented fixtures")
				}
				return
			}
			before, err := snapshot.Parse(record.Before)
			if err != nil {
				t.Fatalf("parse before: %v", err)
			}
			after, err := snapshot.Parse(record.After)
			if err != nil {
				t.Fatalf("parse after: %v", err)
			}

			// When
			report := diff.Compare(before, after)

			// Then
			got := make([]string, len(report.Findings))
			for index, finding := range report.Findings {
				got[index] = finding.Code
			}
			if !equalStrings(got, record.Want) {
				t.Fatalf("codes = %v, want %v (%s)", got, record.Want, record.Fidelity)
			}
			if record.AdapterBefore != nil {
				adapterBefore, err := snapshot.ParseUniswapLocalMap(record.AdapterBefore, record.DestinationChain)
				if err != nil {
					t.Fatalf("parse adapter before: %v", err)
				}
				adapterAfter, err := snapshot.ParseUniswapLocalMap(record.AdapterAfter, record.DestinationChain)
				if err != nil {
					t.Fatalf("parse adapter after: %v", err)
				}
				adapterReport := diff.Compare(adapterBefore, adapterAfter)
				if !equalStrings(codes(adapterReport), record.Want) {
					t.Fatalf("adapter codes = %v, want %v", codes(adapterReport), record.Want)
				}
			}
		})
	}
}

func codes(report diff.Report) []string {
	result := make([]string, len(report.Findings))
	for index, finding := range report.Findings {
		result[index] = finding.Code
	}
	return result
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
