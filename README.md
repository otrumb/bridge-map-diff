# bridge-map-diff

Frozen evidence corpus for future bridge-map semantic diff work. Current scope contains no semantic engine, network client, registry integration, or canonical-asset assertions.

## Gate

Verdict: **GO**.

- 8 license-safe cases
- 3 addition or migration cases
- 3 removal, retarget, or conflict-review cases
- 2 semantic no-op negative controls
- 4 cases where semantic analysis adds value beyond schema validation, `jq`, or line diff

Run `go test -race -shuffle=on -count=1 ./...` to verify frozen counts, categories, commit pairs, source paths, licenses, and mapping facts.

## Competitor Boundary

`@uniswap/token-list-bridge-utils` at `3b9e97f0644126e7ea1b22a6eecaa22214a81015` already handles schema-shaped token lists, mapping providers, case-insensitive token deduplication, manual mapping precedence, reciprocal `bridgeInfo` verification, and explicit child-token exclusion. Generic normalization or JSON diff would duplicate existing behavior.

Corpus preserves distinct semantic opportunities:

1. Collapse source-layout moves into no-op when normalized chain/address mapping is unchanged (`wormhole-prime-layout-move`).
2. Classify coordinated old-pair removal plus new-pair addition as attested migration, not unrelated churn (`wormhole-token-list-elon-retarget`).
3. Flag runtime suppressions and many-to-one destination collisions that plain mapping-file diff cannot explain (`uniswap-excluded-token-conflict`).
4. Compare EVM addresses case-insensitively while retaining classifier changes caused by case-sensitive code (`megaeth-canonical-case-normalization`).

See `PROVENANCE.md` and `corpus/manifest.json` for evidence.
