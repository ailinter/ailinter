# redteam_demo — adversarial-loop demo package

Dogfood demonstration of the review ladder: a Unicode string-utility package
(`stringutils`) that went through a red-team hardening loop before shipping.

## History

- **Round 1 (red-team):** found `ContainsAny` byte-vs-rune false positives
  (é/è, 日/本, 🙂/🙃) and an O(n·m) quadratic path → fixed with rune-set
  matching (O(n+m)) + early-exit guards.
- **Round 2 (red-team):** SURVIVED — 50 probes + 180k differential fuzz pairs
  vs stdlib.
- **Code review (2026-09-07):** caught two remaining helpers of the same
  byte-vs-rune class (`Reverse`, `MaxLen` emitted invalid UTF-8) → made both
  rune-safe with multibyte corpus tests + a UTF-8 property test.

## Isolation

This is a **nested module** (its own `go.mod`), so it is intentionally
excluded from the parent module's `go build ./...`, `go vet`, `go test`,
coverage, and staticcheck surfaces — a demo artifact must not add permanent
CI weight to the product.

Run its tests independently:

```sh
cd redteam_demo && go test ./... -count=1
```
