# US021 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). US021 is a package.json hygiene story with no runtime behaviour change. The "tests" here are structural assertions rather than RTL component tests. They can be implemented as Jest tests that read the filesystem (using `fs.readFileSync` and JSON parsing) or as `npm ls` output parsers, or as part of a CI assertion step. They live in `web/` alongside other Jest tests (e.g. `web/package-hygiene.test.ts` or similar).

## Coverage matrix

| AC scenario | Test ID | What it asserts |
|---|---|---|
| `@testing-library/dom` absent from `dependencies` | FCT-US021-001 | `web/package.json` `dependencies` block has no `@testing-library/dom` key |
| `@testing-library/dom` present in `devDependencies` | FCT-US021-002 | `web/package.json` `devDependencies` block contains `@testing-library/dom` with specifier `^10.4.1` or higher |
| No production source files import `@testing-library/dom` | FCT-US021-003 | `grep` over production source folders returns zero hits |
| Dev install still resolves `@testing-library/dom` (tests pass) | FCT-US021-004 | All Jest tests pass with the moved dep (meta-assertion: `npm test` exits 0) |

## Component tests

### FCT-US021-001 — `@testing-library/dom` is absent from `dependencies`

- **Test type:** structural assertion (Jest test reading `package.json`)
- **File:** create `web/package-hygiene.test.ts` (or add to an existing structural test file if one exists).
- **Given:** `const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, '../package.json'), 'utf-8'))`
- **Then:**
  - `pkg.dependencies` does NOT have the key `@testing-library/dom`.
  - `Object.keys(pkg.dependencies).includes('@testing-library/dom')` is `false`.
- **Architecture cite:** US021 AC "Scenario: `@testing-library/dom` is in `devDependencies`"; architecture §2 US021 row.

### FCT-US021-002 — `@testing-library/dom` is in `devDependencies` with correct version

- **Test type:** structural assertion
- **Given:** same `pkg` as above
- **Then:**
  - `pkg.devDependencies['@testing-library/dom']` is defined.
  - The version specifier starts with `^10.4` or higher (satisfies `>=10.4.1`).
- **Architecture cite:** US021 AC — "version specifier matches or supersedes the current `^10.4.1`".

### FCT-US021-003 — No production source imports `@testing-library/dom`

- **Test type:** structural assertion (file-content scan)
- **Given:** read all `.ts` and `.tsx` files under `web/components/`, `web/hooks/`, `web/lib/`, and `web/pages/` (production folders only — exclude `web/test/`, `web/__mocks__/`, and any `*.test.*` file).
- **Then:**
  - Zero files contain the string `'@testing-library/dom'` or `"@testing-library/dom"`.
- **Notes:** Use `fs` + a recursive directory walker or `glob` to enumerate files. Do not shell-exec `grep` inside the Jest test (non-deterministic in CI environments without the binary).
- **Architecture cite:** US021 AC "Scenario: no runtime imports of `@testing-library/dom` exist".

### FCT-US021-004 — Dev install still resolves `@testing-library/dom` (meta-assertion)

- **Test type:** meta-assertion (documented, not a Jest test case)
- **Notes:** This is not a Jest test in the traditional sense — it is the assertion that `npm install` (full install including dev) completes without error and that all pre-existing Jest tests continue to pass after the move. The dev verifies this by running `cd web && npm install && npm test --watchAll=false --forceExit` locally and in CI. Document this verification step in the test report (not in a test file). If the lockfile churn is larger than expected, the dev raises `ARCHITECTURE_GAP_FOUND` per architecture risk R2.
- **Architecture cite:** US021 AC "Scenario: dev install still works for tests".
