# Security

## Model

`bridge-map-diff` reads two local native JSON snapshots and writes a report to stdout. It performs no network calls, URL loading, environment configuration, shell execution, plugins, database access, automatic fixes, or output-file writes.

Inputs are untrusted. Parser rejects files over 16 MiB, nesting beyond 64, more than 100,000 mappings, strings over 1,024 bytes, NUL, invalid UTF-8, unknown fields, malformed endpoint kinds, and invalid EVM addresses. Mixed-case EVM addresses require EIP-55; uniform case is accepted and normalized lowercase.

Reports describe repository-declared mappings. They do not prove migration safety, authorization, authenticity, or canonical-asset status.

## Reporting

Before publication, report issues privately to repository owner. Do not include secrets or live credentials. Include smallest local fixture that reproduces issue.
