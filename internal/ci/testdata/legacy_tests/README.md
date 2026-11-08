# Legacy service tests (not compiled)

Same story as `internal/api/testdata/legacy_tests/` (issue #5): these suites call
constructors with the pre-refactor signatures and build against removed mock
types, so they have not compiled or run since the service dependency refactor.

Kept under `testdata/` so the Go toolchain ignores them; replacing them is part
of the follow-up issue. `internal/ci` currently has no running tests.
