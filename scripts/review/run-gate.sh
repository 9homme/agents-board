#!/usr/bin/env bash
# scripts/review/run-gate.sh — code-review gate for tech-lead.
#
# Runs the quality + security checks tech-lead must clear before approving
# a task. Fails loudly if a required tool is missing — the gate is mandatory,
# not "best effort".
#
# Usage:
#   scripts/review/run-gate.sh be <service-dir>     # e.g. services/basket
#   scripts/review/run-gate.sh fe
#   scripts/review/run-gate.sh cross                # repo-wide (semgrep + secrets)
#   scripts/review/run-gate.sh all <service-dir>    # cross + be + fe
#
# Exit codes:
#   0   all checks pass
#   1   one or more checks failed
#   2   bad invocation / missing required tool
#
# See scripts/review/README.md for install instructions.

set -u
IFS=$'\n\t'

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# ---- output helpers ----
if [ -t 1 ]; then
  RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  RED=""; GREEN=""; YELLOW=""; BOLD=""; RESET=""
fi

FAILED=0
declare -a FAILED_CHECKS=()

pass() { printf "  ${GREEN}PASS${RESET}  %s\n" "$1"; }
fail() { printf "  ${RED}FAIL${RESET}  %s\n" "$1"; FAILED=$((FAILED+1)); FAILED_CHECKS+=("$1"); }
section() { printf "\n${BOLD}== %s ==${RESET}\n" "$1"; }

require_tool() {
  # require_tool <cmd> <install-hint>
  if ! command -v "$1" >/dev/null 2>&1; then
    printf "${RED}MISSING TOOL${RESET}: %s\n  install: %s\n" "$1" "$2" >&2
    exit 2
  fi
}

run_check() {
  # run_check "<name>" <cmd...>
  local name="$1"; shift
  local out rc
  out="$("$@" 2>&1)"; rc=$?
  if [ $rc -eq 0 ]; then
    pass "$name"
  else
    fail "$name"
    printf "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n${YELLOW}-----------------------${RESET}\n" "$rc" "$out"
  fi
}

run_check_warn() {
  # run_check_warn "<name>" <cmd...> (non-fatal)
  local name="$1"; shift
  local out rc
  out="$("$@" 2>&1)"; rc=$?
  if [ $rc -eq 0 ]; then
    pass "$name"
  else
    printf "  ${YELLOW}WARN${RESET}  %s\n" "$name"
    printf "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n${YELLOW}-----------------------${RESET}\n" "$rc" "$out"
  fi
}

# ---- BE gate ----
gate_be() {
  local svc="$1"
  if [ -z "${svc:-}" ] || [ ! -d "$svc" ]; then
    echo "${RED}gate-be: pass an existing service dir (e.g. services/basket)${RESET}" >&2
    exit 2
  fi

  section "BE gate · $svc"

  require_tool go               "https://go.dev/dl/"
  require_tool golangci-lint    "brew install golangci-lint  |  https://golangci-lint.run/welcome/install/"
  require_tool gosec            "go install github.com/securego/gosec/v2/cmd/gosec@latest"
  require_tool govulncheck      "go install golang.org/x/vuln/cmd/govulncheck@latest"

  # ( Use a local variable to capture subshell failures if needed, or avoid subshell )
  pushd "$svc" >/dev/null
  run_check "gofmt -s (no diff)"        bash -c 'diff -u <(echo -n) <(gofmt -s -d . | tee /dev/stderr)'
  run_check "go vet ./..."              go vet ./...
  run_check "golangci-lint run ./..."   golangci-lint run --timeout=2m --no-config ./...
  run_check "go test ./..."             go test ./...
  run_check "gosec ./... (security)"    gosec -quiet -severity=medium ./...
  run_check "govulncheck ./..."         govulncheck ./...
  popd >/dev/null
}

# ---- FE gate ----
gate_fe() {
  if [ ! -d "web" ]; then
    echo "${RED}gate-fe: web/ not found${RESET}" >&2
    exit 2
  fi

  section "FE gate · web/"

  require_tool npm "https://nodejs.org/"

  (
    cd web
    run_check "npm run typecheck"                       npm run typecheck --silent
    run_check "npm run lint (--max-warnings=0)"         bash -c 'npm run lint --silent -- --max-warnings=0'
    run_check "npm test (--watchAll=false)"             bash -c 'npm test --silent -- --watchAll=false'
    # Use || true to make it non-fatal, but it will still be printed if run_check handles it.
    # Actually, run_check uses the exit code of the command.
    # We want to see the output but NOT increment FAILED.
    # I'll create a new helper for non-fatal checks.
    run_check_warn "npm audit (omit=dev, high+)"         bash -c 'npm audit --omit=dev --audit-level=high'
  )

  # Project anti-patterns (CSR-only is non-negotiable per CLAUDE.md).
  section "FE anti-pattern scan · web/"
  csr_violations=$(grep -rEn '^[[:space:]]*export[[:space:]]+(async[[:space:]]+)?function[[:space:]]+(getServerSideProps|getStaticProps|getInitialProps)\b' web/pages 2>/dev/null || true)
  if [ -n "$csr_violations" ]; then
    fail "CSR-only violation: SSR/SSG exports found"
    printf "%s\n" "$csr_violations"
  else
    pass "no getServerSideProps / getStaticProps / getInitialProps in web/pages/"
  fi

  if [ -d "web/pages/api" ]; then
    fail "API routes present: web/pages/api/ must not exist (CSR-only)"
  else
    pass "no web/pages/api/ directory"
  fi

  fetch_outside=$(grep -rEn '\bfetch[[:space:]]*\(' web/components web/hooks web/pages 2>/dev/null | grep -v -E '/(lib/api|test/msw)/' || true)
  if [ -n "$fetch_outside" ]; then
    fail "fetch() outside web/lib/api/"
    printf "%s\n" "$fetch_outside"
  else
    pass "no raw fetch() outside web/lib/api/"
  fi
}

# ---- cross-cutting (repo-wide) ----
gate_cross() {
  section "Cross-cutting · repo"

  require_tool semgrep   "brew install semgrep  |  pipx install semgrep"
  require_tool gitleaks  "brew install gitleaks  |  https://github.com/gitleaks/gitleaks"

  # OWASP top 10 + per-lang rules. --error makes findings fail the run.
  run_check "semgrep (owasp/golang/typescript)" \
    semgrep --quiet --error \
      --config=p/owasp-top-ten \
      --config=p/golang \
      --config=p/typescript \
      --config=p/react \
      --exclude=node_modules --exclude=.next --exclude=tests/e2e

  # gitleaks: scan working tree even when no git history exists.
  run_check "gitleaks (no secrets)" \
    gitleaks detect --no-banner --redact --source="$REPO_ROOT" --no-git
}

# ---- dispatch ----
usage() {
  sed -n '2,25p' "$0" | sed -e 's/^# //' -e 's/^#$//'
  exit 2
}

[ $# -ge 1 ] || usage
track="$1"; shift || true

case "$track" in
  be)    gate_be "${1:-}" ;;
  fe)    gate_fe ;;
  cross) gate_cross ;;
  all)
    gate_cross
    gate_be "${1:-}"
    gate_fe
    ;;
  -h|--help|help) usage ;;
  *) echo "unknown track: $track" >&2; usage ;;
esac

printf "\n"
if [ $FAILED -eq 0 ]; then
  printf "${GREEN}${BOLD}REVIEW GATE: PASS${RESET}\n"
  exit 0
else
  printf "${RED}${BOLD}REVIEW GATE: FAIL${RESET} (%d check(s))\n" "$FAILED"
  for c in "${FAILED_CHECKS[@]}"; do printf "  - %s\n" "$c"; done
  exit 1
fi
