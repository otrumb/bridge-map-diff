# bridge-map-diff

Offline semantic diff CLI for native bridge-map snapshots. Tool reports factual mapping changes and review conditions; it never asserts migration safety, authorization, authenticity, or canonical-asset status.

## Usage

```text
bridge-map-diff --schema native [--format json|text] BEFORE AFTER
bridge-map-diff --schema uniswap-local-map --destination-chain CHAIN [--format json|text] BEFORE AFTER
```

Exit codes: `0` no findings, `1` informational changes, `2` review findings, `3` usage/input/output error. Input paths must be local files. Output is stdout only.

Native snapshot shape:

```json
{"version":1,"mappings":[{"registry":"example","scope":"","route":"canonical","origin":{"chain":"ethereum","kind":"evm","address":"0xde709f2102306220921060314715629080e2fb77","decimals":18},"destination":{"chain":"base","kind":"evm","address":"0x1111111111111111111111111111111111111111","decimals":18},"symbol":"TOK"}]}
```

Endpoint `kind` is `evm` or `opaque`. Semantic identity is `(registry, scope, route, origin.chain, origin.address, destination.chain)`. `supersedes` may name an exact prior identity to declare migration grouping. Tool never infers migration from symbol, timing, or address similarity.

Supported schemas are `native` and `uniswap-local-map`. Selection is always explicit. `uniswap-local-map` consumes Uniswap `src/local_mappings/*.json` objects keyed by Ethereum origin address with `childToken` and `decimals`; destination chain comes from required CLI context. Adapter does not read or generate `bridgeInfo`.

## Gate

Verdict: **GO**.

- 8 license-safe cases
- 3 addition or migration cases
- 3 removal, retarget, or conflict-review cases
- 2 semantic no-op negative controls
- 4 cases where semantic analysis adds value beyond schema validation, `jq`, or line diff

Run `CGO_ENABLED=0 go test -shuffle=on -count=1 ./...` to verify engine and frozen corpus replay. Seven native cases execute, and Monad also executes an additional real `uniswap-local-map` source-shape replay. Conflict-suppression remains inventory-only because endpoint facts are absent; inventing them would weaken provenance.

## Competitor Boundary

`@uniswap/token-list-bridge-utils` at `3b9e97f0644126e7ea1b22a6eecaa22214a81015` already handles schema-shaped token lists, mapping providers, case-insensitive token deduplication, manual mapping precedence, reciprocal `bridgeInfo` verification, and explicit child-token exclusion. Generic normalization or JSON diff would duplicate existing behavior.

Corpus preserves distinct semantic opportunities:

1. Collapse source-layout moves into no-op when normalized chain/address mapping is unchanged (`wormhole-prime-layout-move`).
2. Classify coordinated old-pair removal plus new-pair addition as attested migration, not unrelated churn (`wormhole-token-list-elon-retarget`).
3. Flag runtime suppressions and many-to-one destination collisions that plain mapping-file diff cannot explain (`uniswap-excluded-token-conflict`).
4. Compare EVM addresses case-insensitively while retaining classifier changes caused by case-sensitive code (`megaeth-canonical-case-normalization`).

See `PROVENANCE.md` and `corpus/manifest.json` for evidence.

## Limits

- 16 MiB per file, depth 64, 100,000 mappings, 1,024 bytes per string, 10,000 report findings.
- EVM mixed case must satisfy EIP-55; lowercase and uppercase normalize lowercase. Opaque identifiers retain case after surrounding whitespace removal.
- Destination collisions are route-aware. Parallel routes with empty scope require review. Duplicate active identities and EVM zero addresses require review.
- Omitted or `null` decimals are unknown; numeric `0` is known zero.
- No network, URL, environment, shell, plugin, database, output-file, or mutation features.

## Release

Only exact tag `v0.1.0` publishes. Windows and Ubuntu validate tagged commit with CGO disabled before release job receives `contents: write`. Linux and Windows amd64 archives include `LICENSE`; `SHA256SUMS` covers archives only.
