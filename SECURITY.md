# Security

## Model

`bridge-map-diff` reads two local JSON snapshots using the explicitly selected `native` or `uniswap-local-map` schema and writes a report to stdout. It performs no network calls, URL loading, environment configuration, shell execution, plugins, database access, automatic fixes, or output-file writes.

Inputs are untrusted. Parser rejects files over 16 MiB, nesting beyond 64, more than 100,000 mappings, strings over 1,024 bytes, NUL, invalid UTF-8, unknown fields, malformed endpoint kinds, and invalid EVM addresses. Mixed-case EVM addresses require EIP-55; uniform case is accepted and normalized lowercase.

Reports describe repository-declared mappings. They do not prove migration safety, authorization, authenticity, or canonical-asset status.

## Reporting

Report suspected vulnerabilities privately through [GitHub private vulnerability reporting](https://github.com/otrumb/bridge-map-diff/security/advisories/new). Do not open public issues or discussions for vulnerability reports. Do not include secrets or live credentials. Include version, OS, selected schema, reproduction steps, impact, and smallest local fixture that reproduces issue.
