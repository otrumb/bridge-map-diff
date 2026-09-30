# AGENTS.md

Go 1.23+ corpus for bridge-map semantic diff research.

## Commands
- `go test -race -shuffle=on -count=1 ./...` verifies frozen corpus gates.
- `go vet ./...` checks Go sources.

## Scope
- `corpus/manifest.json` is frozen normalized factual evidence.
- `PROVENANCE.md` records immutable upstream provenance and rationale.
- No semantic diff engine belongs in this repository phase.
- Keep source files at or below 250 pure lines.
