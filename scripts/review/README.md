# scripts/review/ — code-review gate

Static-analysis gate that tech-lead **must** run before approving any task in
`Status: in_review`. The gate is mandatory: missing tools cause a hard failure
(exit 2) so tech-lead can't silently skip checks.

## Usage

```sh
scripts/review/run-gate.sh be services/<name>   # BE task: quality + security
scripts/review/run-gate.sh fe                   # FE task: quality + security + CSR-only check
scripts/review/run-gate.sh cross                # repo-wide: semgrep + gitleaks
scripts/review/run-gate.sh all services/<name>  # everything (cross + be + fe)
```

Exit codes: `0` = all checks pass · `1` = at least one check failed · `2` = missing tool / bad invocation.

Tech-lead pastes the final `REVIEW GATE: PASS` / `REVIEW GATE: FAIL` line + the
list of failed checks into the task's `### Review pass N` entry.

## What it runs

**BE (`gate be <service-dir>`)** — runs from inside the service module:

| Check | Tool | Purpose |
|---|---|---|
| `gofmt -s` (no diff) | `gofmt` (Go stdlib) | formatting drift |
| `go vet ./...` | `go` | suspicious constructs |
| `golangci-lint run` | `golangci-lint` | bundle: staticcheck, errcheck, ineffassign, unused, gocritic, revive (config at repo-root `.golangci.yml`) |
| `go test ./...` | `go` | DoD already requires this |
| `gosec ./...` | `gosec` | SAST: SQLi, hardcoded creds, weak crypto, unsafe use |
| `govulncheck ./...` | `govulncheck` | reachable CVEs in deps (official Go vuln DB) |

**FE (`gate fe`)** — runs from inside `web/`:

| Check | Tool | Purpose |
|---|---|---|
| `npm run typecheck` | `tsc --noEmit` | DoD already requires this |
| `npm run lint -- --max-warnings=0` | eslint + `eslint-plugin-security` + `eslint-plugin-no-secrets` | quality + security rules |
| `npm test -- --watchAll=false` | jest | DoD already requires this |
| `npm audit --omit=dev --audit-level=high` | npm | dep vulns |
| CSR-only scan | grep | no `getServerSideProps` / `getStaticProps` / `getInitialProps` / `web/pages/api/` |
| `fetch()` boundary scan | grep | no raw `fetch()` outside `web/lib/api/` |

**Cross (`gate cross`)** — repo-wide:

| Check | Tool | Purpose |
|---|---|---|
| `semgrep --error` | `semgrep` | OWASP top 10 + golang + typescript + react rule packs |
| `gitleaks detect --no-git` | `gitleaks` | hardcoded secrets in the working tree |

## Install

```sh
# Go side
brew install go golangci-lint
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Cross-cutting
brew install semgrep gitleaks
# or: pipx install semgrep

# FE side — npm is enough; eslint plugins live in web/package.json devDependencies.
# Make sure web/package.json has eslint-plugin-security and eslint-plugin-no-secrets,
# and the eslint config extends:
#   "plugin:security/recommended", "plugin:no-secrets/all"
```

## Wiring into tech-lead

The agent prompt at `.claude/agents/tech-lead.md` (sync'd to `.gemini/agents/tech-lead.md`
by `scripts/sync-gemini.py`) makes the gate mandatory in review mode. Tech-lead
cannot issue an `approved` verdict without pasting the gate output. A parity
test under `.claude/evals/tests/test_agent_integrity.py` asserts both platform
versions reference `scripts/review/run-gate.sh`.
