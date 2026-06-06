// Package handler — internal test helpers re-exporting unexported methods so that
// the external test package (handler_test) can call sendError and sendToolResultError
// directly for UT-010..UT-013 (US008). This file is a _test.go file and does NOT
// constitute a production-code change (architecture.md §8.3 / D-012).
package handler

// SendError re-exports (*Handler).sendError for use in handler_test tests.
// Method value bound at test-package init; callers pass `h` as the first argument.
var SendError = (*Handler).sendError

// SendToolResultError re-exports (*Handler).sendToolResultError for use in
// handler_test tests.
var SendToolResultError = (*Handler).sendToolResultError
