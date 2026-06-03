# US007 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US007 moves `@testing-library/dom` from `dependencies` to `devDependencies` in `web/package.json`. This is a package.json hygiene change with zero runtime behaviour impact:

- The moved package is a test-time library; it is not imported by any production source file (verified by FCT-US007-003).
- The change has no effect on the compiled Next.js bundle served at `http://localhost:3000` or on the api-server. The docker-compose stack (US008) would see no difference whether this story is landed or not.
- The Browser library and RequestsLibrary have no surface to exercise against a `package.json` change.
- The correctness guarantees (correct dep location, no production import, dev install still works) are fully structural assertions covered by FCT-US007-001 through FCT-US007-003 plus the manual FCT-US007-004 meta-assertion.

**Verdict: No e2e scenarios. Package hygiene; covered by FCT-US007-001 through FCT-US007-004.**
