# US010 — Backend unit & integration test specification
# `ResolveDBURL` helper + startup log line integration test

**For BE Dev:** these are the tests you write FIRST (TDD red). Create `services/agent-board/internal/config/dburl_test.go` (new file alongside the new `dburl.go`). Startup log-line integration test lives in `services/agent-board/cmd/api-server/main_test.go` (edit or create). Production code introduces `services/agent-board/internal/config/dburl.go` (new package). `main.go` files are edited; production code changes are expected for this story.

**Integration-test shape choice (architecture.md §5.7):** this spec picks **(b)** — the `run()` helper approach. The dev refactors (or confirms the existing structure of) `main.go` to expose a `run(logger *log.Logger) error` helper, then the test calls `run(l)` with a `log.SetOutput(&buf)` buffer and asserts the captured output contains `"db config: using DATABASE_URL"`. This is faster and more deterministic than subprocess spawning. The optional hard-fail regression test (mcp-server with `DB_URL` set) is included as IT-003 via a subprocess approach because it tests process exit-code, which cannot be easily asserted without spawning a subprocess.

## Coverage matrix

| AC scenario | Layer | Test ID | Package | Function under test |
|---|---|---|---|---|
| `DATABASE_URL` set, `DB_URL` unset — happy path | unit | UT-001 | `internal/config` | `ResolveDBURL` |
| `DB_URL` set, `DATABASE_URL` unset — rename error | unit | UT-002 | `internal/config` | `ResolveDBURL` |
| Both `DB_URL` and `DATABASE_URL` set — disambiguate error | unit | UT-003 | `internal/config` | `ResolveDBURL` |
| Neither set — required error | unit | UT-004 | `internal/config` | `ResolveDBURL` |
| Startup log line emitted before DB ping (api-server) | integration | IT-001 | `cmd/api-server` | `run()` helper |
| Startup log line emitted before DB ping (mcp-server) | integration | IT-002 | `cmd/mcp-server` | `run()` helper |
| mcp-server hard-fails when only DB_URL set (subprocess) | integration | IT-003 | `cmd/mcp-server` | binary startup |
| package coverage ≥95% on `internal/config` | integration | IT-004 | `internal/config` | `dburl.go` |
| full suite still passes | integration | IT-005 | `services/agent-board` | `go test ./...` |

## Unit tests

### UT-001 — `TestResolveDBURL_OnlyDatabaseURLSet_Happy`
- **Service:** `services/agent-board`
- **Package:** `internal/config`
- **File:** `dburl_test.go`
- **Given:**
  ```go
  t.Setenv("DATABASE_URL", "postgres://x")
  // Explicitly unset DB_URL:
  os.Unsetenv("DB_URL")
  ```
- **When:** `url, err := config.ResolveDBURL()`
- **Then:**
  - `assert.NoError(t, err)`
  - `assert.Equal(t, "postgres://x", url)`
- **Architecture cite:** architecture.md §5.6 case 1; §5.4 happy-path row

---

### UT-002 — `TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError`
- **Service:** `services/agent-board`
- **Package:** `internal/config`
- **Given:**
  ```go
  t.Setenv("DB_URL", "postgres://legacy")
  os.Unsetenv("DATABASE_URL")
  ```
- **When:** `url, err := config.ResolveDBURL()`
- **Then:**
  - `assert.Equal(t, "", url)`
  - `assert.EqualError(t, err, "DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)")`
- **Exact error string (locked in architecture.md §5.4):** `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)"`
- **Architecture cite:** architecture.md §5.6 case 2; §5.4 `Only DB_URL set` row

---

### UT-003 — `TestResolveDBURL_BothSet_RejectsWithDisambiguateError`
- **Service:** `services/agent-board`
- **Package:** `internal/config`
- **Given:**
  ```go
  t.Setenv("DATABASE_URL", "postgres://new")
  t.Setenv("DB_URL", "postgres://old")
  ```
- **When:** `url, err := config.ResolveDBURL()`
- **Then:**
  - `assert.Equal(t, "", url)`
  - `assert.EqualError(t, err, "DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)")`
- **Exact error string (locked in architecture.md §5.4):** `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)"`
- **Architecture cite:** architecture.md §5.6 case 3; §5.4 `Both set` row

---

### UT-004 — `TestResolveDBURL_NeitherSet_RejectsWithRequiredError`
- **Service:** `services/agent-board`
- **Package:** `internal/config`
- **Given:**
  ```go
  os.Unsetenv("DATABASE_URL")
  os.Unsetenv("DB_URL")
  ```
- **When:** `url, err := config.ResolveDBURL()`
- **Then:**
  - `assert.Equal(t, "", url)`
  - `assert.EqualError(t, err, "DATABASE_URL environment variable is required")`
- **Exact error string (locked in architecture.md §5.4):** `"DATABASE_URL environment variable is required"`
- **Architecture cite:** architecture.md §5.6 case 4; §5.4 `Neither set` row

## Integration tests

### IT-001 — startup log line (api-server)
- **Service:** `services/agent-board/cmd/api-server`
- **File:** `cmd/api-server/main_test.go`
- **Shape:** approach (b) — `run()` helper with captured logger
- **Given:**
  ```go
  t.Setenv("DATABASE_URL", "postgres://localhost/test_db")
  // DB_URL unset
  var buf bytes.Buffer
  log.SetOutput(&buf)
  defer log.SetOutput(os.Stderr)
  ```
- **When:** call the `run()` helper (which calls `config.ResolveDBURL()` then logs the happy-path line then attempts DB ping)
  - NOTE: The DB ping will fail (no real DB); that is expected. The assertion is on the log output BEFORE the ping failure.
- **Then:**
  - `assert.Contains(t, buf.String(), "db config: using DATABASE_URL")`
  - The log line appears in `buf.String()` before any ping-failure log line
- **Architecture cite:** architecture.md §5.3; §5.7 approach (b)

---

### IT-002 — startup log line (mcp-server)
- **Service:** `services/agent-board/cmd/mcp-server`
- **File:** `cmd/mcp-server/main_test.go`
- **Shape:** same as IT-001 (approach b)
- **Given:** `DATABASE_URL` set, `DB_URL` unset
- **Then:** captured log output contains `"db config: using DATABASE_URL"`
- **Architecture cite:** architecture.md §5.3; §5.7

---

### IT-003 — mcp-server hard-fail regression (subprocess)
- **Service:** `services/agent-board/cmd/mcp-server`
- **File:** `cmd/mcp-server/main_test.go`
- **Shape:** `os/exec` subprocess spawn — the only IT in this spec that uses subprocess
- **Given:**
  ```go
  cmd := exec.Command("go", "run", "./cmd/mcp-server")
  cmd.Env = append(os.Environ(), "DB_URL=postgres://x")
  // Remove DATABASE_URL from env
  filteredEnv := []string{}
  for _, e := range cmd.Env {
      if !strings.HasPrefix(e, "DATABASE_URL=") {
          filteredEnv = append(filteredEnv, e)
      }
  }
  cmd.Env = filteredEnv
  ```
- **When:** `output, err := cmd.CombinedOutput()`
- **Then:**
  - `err` is non-nil (process exits non-zero)
  - `string(output)` contains `"DB_URL is no longer supported"` AND `"rename to DATABASE_URL"`
- **Purpose:** guards against a future refactor that silently re-accepts `DB_URL`.
- **Architecture cite:** architecture.md §5.7 optional hard-fail regression test; §5.4

---

### IT-004 — package coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/config -coverprofile=/tmp/config.out
  go tool cover -func=/tmp/config.out | grep dburl.go
  ```
- **Expect:** `dburl.go` shows ≥95% statement coverage. Four tests over a four-branch switch makes 100% trivially achievable.

---

### IT-005 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

None. The four-branch switch in `ResolveDBURL` is fully covered by UT-001 through UT-004. No unreachable lines anticipated.
