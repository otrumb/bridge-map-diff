package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/otrumb/bridge-map-diff/internal/diff"
	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func Run(args []string, stdout, stderr io.Writer) int {
	options, paths, ok := parseArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "usage: bridge-map-diff --schema native|uniswap-local-map [--destination-chain CHAIN] [--format json|text] BEFORE AFTER")
		return 3
	}
	before, err := load(paths[0], options)
	if err != nil {
		fmt.Fprintf(stderr, "before: %v\n", err)
		return 3
	}
	after, err := load(paths[1], options)
	if err != nil {
		fmt.Fprintf(stderr, "after: %v\n", err)
		return 3
	}
	report := diff.Compare(before, after)
	if err := render(stdout, options.format, report); err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 3
	}
	if report.Summary.Review > 0 {
		return 2
	}
	if report.Summary.Changes > 0 {
		return 1
	}
	return 0
}

type options struct {
	format           string
	schema           string
	destinationChain string
}

func parseArgs(args []string) (options, []string, bool) {
	parsed := options{format: "text"}
	for len(args) >= 2 && strings.HasPrefix(args[0], "--") {
		switch args[0] {
		case "--format":
			parsed.format = args[1]
		case "--schema":
			parsed.schema = args[1]
		case "--destination-chain":
			parsed.destinationChain = args[1]
		default:
			return options{}, nil, false
		}
		args = args[2:]
	}
	validSchema := parsed.schema == "native" || parsed.schema == "uniswap-local-map"
	validContext := parsed.schema != "uniswap-local-map" || parsed.destinationChain != ""
	return parsed, args, len(args) == 2 && validSchema && validContext && (parsed.format == "json" || parsed.format == "text")
}

func load(path string, options options) (snapshot.Snapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return snapshot.Snapshot{}, fmt.Errorf("open input: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 16<<20+1))
	if err != nil {
		return snapshot.Snapshot{}, fmt.Errorf("read input: %w", err)
	}
	var parsed snapshot.Snapshot
	if options.schema == "native" {
		parsed, err = snapshot.Parse(data)
	} else {
		parsed, err = snapshot.ParseUniswapLocalMap(data, options.destinationChain)
	}
	if err != nil {
		return snapshot.Snapshot{}, fmt.Errorf("parse input: %w", err)
	}
	return parsed, nil
}

func render(output io.Writer, format string, report diff.Report) error {
	if format == "json" {
		encoder := json.NewEncoder(output)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(report)
	}
	if _, err := fmt.Fprintf(output, "bridge-map-diff: %d change(s), %d review, %d info\n", report.Summary.Changes, report.Summary.Review, report.Summary.Info); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	for _, finding := range report.Findings {
		line := strings.ToUpper(finding.Severity) + " " + finding.Code + " " + finding.Identity
		if finding.Before != "" {
			line += " before=" + finding.Before
		}
		if finding.After != "" {
			line += " after=" + finding.After
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return fmt.Errorf("write finding: %w", err)
		}
	}
	return nil
}
