# Legacy amortization test corpus (not compiled)

These 7 files are the orphaned test helpers and contract/integration suites that
used to live at `internal/api/handlers/amortization_tests/`. They declared
`package amortization` while sitting outside `internal/amortization`, so none of
it ever compiled; moving it into `internal/api/handlers` surfaced **65 distinct
compile errors** against APIs that were never implemented (for example
`LedgerEntryResponse`, `AmortizationRunResponse`, `BaseCI`, removed
`AccumulatedDepreciationBefore/After` fields, and the old
`middleware.AuthenticatedUser` context API).

They are kept under `testdata/` because the Go toolchain ignores that directory:
the code no longer breaks `go build`/`go vet`/`go test` and needs no CI
exclusion, while the intent of the original suites stays inspectable.

Reconciling them with the real `amortization.Service` interface and
`AmortizationHandler` endpoints is tracked as a separate follow-up issue. The
right end state is a rewritten suite that exercises the implemented API, not a
mechanical port of these expectations.
