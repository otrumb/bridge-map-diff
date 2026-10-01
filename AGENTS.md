# AGENTS.md

Go 1.23+ offline CLI for bridge-map semantic review.

## Commands
- `CGO_ENABLED=0 go test -shuffle=on -count=1 ./...` verifies parser, engine, CLI, and corpus replay.
- `CGO_ENABLED=0 go vet ./...` checks Go sources.
- `CGO_ENABLED=0 go build -trimpath ./cmd/bridge-map-diff` builds release binary.

## Scope
- `internal/snapshot` owns strict native parsing and normalization.
- `internal/diff` owns deterministic semantic findings.
- `internal/cli` owns local file boundary, reports, and exit codes.
- `corpus/manifest.json` remains frozen normalized factual evidence.
- `PROVENANCE.md` records immutable upstream provenance and rationale.
- No network, URL, environment, shell, plugin, DB, output file, or automatic mutation.
- Keep source files at or below 250 pure lines.
