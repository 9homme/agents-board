# US007 — Move `@testing-library/dom` from `dependencies` to `devDependencies` in `web/package.json`

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **release engineer producing a production build of `web/`**, I want **test-time tooling to live in `devDependencies`** so that a `npm ci --omit=dev` (or `npm install --production`) install does not ship `@testing-library/dom` into the runtime bundle / lockfile reachable set.

## Acceptance criteria

- **Scenario: `@testing-library/dom` is in `devDependencies`**
  - Given `web/package.json`
  - When the file is read after the story
  - Then `@testing-library/dom` appears in the `devDependencies` block (not `dependencies`)
  - And the version specifier matches or supersedes the current `^10.4.1` (no downgrade)

- **Scenario: `package-lock.json` is regenerated cleanly**
  - Given `web/package-lock.json` is regenerated (`npm install` or `npm install --package-lock-only`)
  - When the lockfile is inspected
  - Then `@testing-library/dom` is still resolved (via the dev path or as a transitive peer of other dev deps such as `@testing-library/react`)
  - And no other top-level dependency is added or removed by the move
  - And the diff to `package-lock.json` is limited to the dep-section reorganisation plus any necessary peer resolution updates — no version churn on unrelated packages

- **Scenario: production install does NOT pull `@testing-library/dom` as a direct dep**
  - Given `cd web && npm ci --omit=dev` (or `npm install --omit=dev`) is run in a clean checkout
  - When `npm ls @testing-library/dom --omit=dev` is inspected
  - Then `@testing-library/dom` is either absent, or present only as a transitive dependency of a runtime package (the audit suggests it was pulled in by mermaid as a transitive peer — that path is acceptable; what we're removing is the explicit top-level `dependencies` entry)

- **Scenario: dev install still works for tests**
  - Given `cd web && npm install` (full install including dev)
  - When `npm test --watchAll=false --forceExit` is run
  - Then all existing Jest tests pass
  - And no test fails with a "Cannot find module '@testing-library/dom'" error
  - And no other dev tooling (typecheck, lint) regresses

- **Scenario: typecheck and lint are clean**
  - Given the package.json change has been applied
  - When `npm run typecheck && npm run lint -- --max-warnings=0` is run
  - Then both pass with zero output

- **Scenario: no runtime imports of `@testing-library/dom` exist**
  - Given a `grep -rn '@testing-library/dom' web/components web/hooks web/lib web/pages` (production source folders only — excluding `test/`, `__mocks__/`, and `*.test.{ts,tsx}`)
  - When the search runs
  - Then there are zero matches (confirms nothing in production code depended on the misplaced dep)

## UI / UX flow expectations

**No UI:** package hygiene. The "flow" is: release engineer runs `npm ci --omit=dev` in CI for prod builds and gets a smaller, test-tool-free install.

## Out of scope
- **Sweeping other potentially misplaced deps.** If others are spotted during this story (e.g. `whatwg-fetch` is in devDeps which is correct; `undici` ditto), leave them alone unless obviously wrong. A broader audit is its own story.
- **Upgrading `@testing-library/dom`** beyond what's needed to land the move.
- **Reorganising the `dependencies` section** alphabetically or stylistically beyond the one move.
- **Adding a `lint:deps` script or a CI check for misplaced deps.** Nice-to-have, but its own story.

## Dependencies
- None. Self-contained edit to `web/package.json` + regenerate `web/package-lock.json`.

## Notes for the team

- **Why this matters:** the audit's §2.2 manual sweep noted `@testing-library/dom` was pulled in by the mermaid install as a transitive peer and ended up at the top level of `dependencies`. That means a production-omitting install still resolves it as a direct dep, polluting the runtime dep graph. Correct home is `devDependencies`.
- **One-line move:** delete `"@testing-library/dom": "^10.4.1"` from `dependencies`, add it to `devDependencies` with the same specifier. Then `npm install` to regenerate the lockfile.
- **Lockfile churn risk:** in rare cases, moving a dep can cause npm to re-resolve a peer. Check the lockfile diff is minimal; if it's large, investigate before committing.
- **Test the prod-install path** by running `npm ci --omit=dev` in a throwaway directory or a CI container — easy to verify the AC.

## Sign-off log
(po-ba appends here on each sign-off pass)
