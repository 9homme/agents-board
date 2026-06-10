# US038 — Backend unit & integration test specification
# Makefile consolidation + `PG_CONN ?=` flip + retire `startup.sh`/`shutdown.sh`

**For BE Dev:** this story is a Makefile + repo-root-script + documentation edit. There are no Go test files to write. The "test spec" is the Makefile-target regression bar defined below. Production code changes: `Makefile` (new targets, `PG_CONN ?=` flip, new `DEV_PG_CONN ?=` variable), deletion of `startup.sh` and `shutdown.sh`, and doc/agent-definition sweeps. No new `*_test.go` files are created by this story.

## Regression bar (architecture.md §10.1, §3 US038 row)

All of the following MUST pass green before the story flips to `in_review`.

| Check | Test ID | Command | What it asserts |
|---|---|---|---|
| `startup.sh` deleted | UT-001 | `ls startup.sh` exits non-zero | file does not exist |
| `shutdown.sh` deleted | UT-002 | `ls shutdown.sh` exits non-zero | file does not exist |
| No references to `startup.sh`/`shutdown.sh` remain | UT-003 | `git grep -nE 'startup\.sh\|shutdown\.sh'` returns zero hits | sweep complete |
| `PG_CONN` uses `?=` (not `:=`) | UT-004 | `grep 'PG_CONN ?=' Makefile` matches | e2e env-overridability correct |
| `DEV_PG_CONN` uses `?=` | UT-005 | `grep 'DEV_PG_CONN ?=' Makefile` matches | dev env-overridability correct |
| Zero `DB_URL` in any new recipe | UT-006 | `grep 'DB_URL' Makefile` returns zero matches in new recipe lines | US034 alignment |
| `make dev-up` target exists | UT-007 | `grep 'dev-up:' Makefile` matches | new target present |
| `make dev-down` target exists | UT-008 | `grep 'dev-down:' Makefile` matches | new target present |
| `make dev-migrate` target exists | UT-009 | `grep 'dev-migrate:' Makefile` matches | new target present |
| `make dev-seed` target exists | UT-010 | `grep 'dev-seed:' Makefile` matches | new target present |
| Existing `make e2e-*` bodies byte-identical | IT-001 | `git diff` e2e-* recipe bodies | zero lines changed in recipe bodies |
| `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` green (3 consecutive runs) | IT-002 | e2e pipeline | compose stack + Robot Framework tests unaffected |
| `PG_CONN` env-override respected by e2e-seed | IT-003 | `PG_CONN=postgres://custom make e2e-seed` (verify psql target) | `?=` semantics work |
| `DEV_PG_CONN` env-override respected by dev-migrate | IT-004 | `DEV_PG_CONN=postgres://custom make dev-migrate` (verify psql target) | `?=` semantics work for dev family |

## Structural checks (UT-001 through UT-010)

These are shell-level assertions the BE Dev runs manually during implementation; tech-lead repeats them at review. They are NOT Go tests — they are Makefile/filesystem verification commands.

### UT-001 — `startup.sh` deleted
- **Command:** `ls startup.sh` → should exit non-zero (file not found)
- **Architecture cite:** architecture.md §3 US038 row DELETE marker; US038 AC `startup.sh` scenario

---

### UT-002 — `shutdown.sh` deleted
- **Command:** `ls shutdown.sh` → should exit non-zero
- **Architecture cite:** architecture.md §3 US038 row DELETE marker

---

### UT-003 — No remaining references to deleted scripts
- **Command:** `git grep -nE 'startup\.sh|shutdown\.sh'`
- **Expect:** zero output (excluding the US038 story file and sign-off log per US038 AC exact wording)
- **Architecture cite:** US038 AC "git grep -nE 'startup\.sh|shutdown\.sh' returns zero hits across the working tree (excluding this story file and the REQ006 sign-off log)"

---

### UT-004 — `PG_CONN` uses `?=`
- **Command:** `grep 'PG_CONN ?=' Makefile`
- **Expect:** one match confirming the `:=` → `?=` flip was applied
- **Default value must be byte-identical:** `postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable`
- **Architecture cite:** US038 AC `PG_CONN ?=` scenario; D-014 Option A union; tech_debt.md line 86

---

### UT-005 — `DEV_PG_CONN` uses `?=`
- **Command:** `grep 'DEV_PG_CONN ?=' Makefile`
- **Expect:** one match with default `postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable`
- **Architecture cite:** architecture.md §3 US038 row; D-013 Q1/Q3

---

### UT-006 — Zero `DB_URL` in any new recipe
- **Command:** `grep 'DB_URL' Makefile`
- **Expect:** zero matches in the new `dev-*` recipe bodies or new variables — if any `DB_URL` reference exists it must be from pre-existing untouched `e2e-*` recipes (verify by reading the Makefile)
- **Architecture cite:** US038 AC "DATABASE_URL is the only DB env var used in any new dev-* recipe"

---

### UT-007 through UT-010 — New targets exist
- **Commands:** `grep 'dev-up:' Makefile`, `grep 'dev-down:' Makefile`, `grep 'dev-migrate:' Makefile`, `grep 'dev-seed:' Makefile`
- **Expect:** each returns one match
- **Architecture cite:** architecture.md §3 US038 row

## Integration tests

### IT-001 — Existing `make e2e-*` targets byte-identical
- **Command:** `git diff HEAD~1 -- Makefile | grep -A5 'e2e-up:\|e2e-down:\|e2e-seed:\|e2e-run:'`
- **Expect:** zero recipe-body lines changed for any `e2e-*` target
- **Note:** variable additions above the recipes that do not affect resolved values are acceptable (e.g. adding `DEV_PG_CONN ?=` on a new line does not change the `e2e-*` recipe bodies).
- **Architecture cite:** US038 AC "existing make e2e-* targets are byte-identical"

---

### IT-002 — E2E pipeline regression (3 consecutive clean runs)
- **Command:**
  ```bash
  make e2e-up && make e2e-seed && make e2e-run && make e2e-down
  make e2e-up && make e2e-seed && make e2e-run && make e2e-down
  make e2e-up && make e2e-seed && make e2e-run && make e2e-down
  ```
- **Expect:** all three runs pass green (zero Robot Framework failures)
- **Architecture cite:** architecture.md §10.1 live-e2e + 3-clean-run mandate for US038

---

### IT-003 — `PG_CONN` env-override for e2e-seed
- **Command:** `PG_CONN=postgres://testuser:pass@localhost:15432/testdb make e2e-seed --dry-run` (or equivalent to verify the variable is interpolated)
- **Expect:** the resolved `psql` command uses the overridden URL
- **Architecture cite:** US038 AC "PG_CONN env-override respected"

---

### IT-004 — `DEV_PG_CONN` env-override for dev-migrate
- **Command:** `DEV_PG_CONN=postgres://testuser:pass@localhost:5432/testdb make dev-migrate --dry-run`
- **Expect:** the resolved `psql` command uses the overridden URL
- **Architecture cite:** US038 AC "DEV_PG_CONN env-override respected"

## Coverage exemptions

N/A — this story does not add Go test files. The regression bar is the e2e pipeline plus the structural Makefile checks above.
