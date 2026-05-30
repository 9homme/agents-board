# TDG Configuration

This repo is a polyglot monorepo. Backend tasks (Track: BE) live under `services/<name>/` (Go), frontend tasks (Track: FE) live under `web/` (Next.js / TypeScript). Pick the section that matches your task's `Track:` field.

---

## Track: BE — Go microservice (`services/<service-name>/`)

### Project Information
- Language: Go (latest stable)
- Framework: Echo (HTTP), standard `testing` + `github.com/stretchr/testify`
- Test Framework: `go test`

### Build Command
```bash
cd services/<service-name> && go build ./...
```

### Test Command
```bash
cd services/<service-name> && go test ./...
```

### Single Test Command
```bash
cd services/<service-name> && go test ./<pkg> -run '^TestName$' -v
```

### Coverage Command
```bash
cd services/<service-name> && go test ./... -cover -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1
```

### Test File Patterns
- Test files: `*_test.go` (colocated with the code they exercise)
- Test directory: same package as the source

### Lint / Vet
```bash
cd services/<service-name> && go vet ./... && gofmt -s -l .
```

---

## Track: FE — Next.js Pages Router CSR (`web/`)

### Project Information
- Language: TypeScript (strict)
- Framework: Next.js Pages Router, CSR-only (no SSR / SSG / API routes)
- Test Framework: Jest + React Testing Library + MSW

### Build Command
```bash
cd web && npm run build
```

### Test Command
```bash
cd web && npm test -- --watchAll=false
```

### Single Test Command
```bash
cd web && npm test -- --watchAll=false --testPathPattern='<relative/path/to/File.test.tsx>' -t '<test name regex>'
```

### Coverage Command
```bash
cd web && npm test -- --watchAll=false --coverage
```

### Test File Patterns
- Test files: `*.test.tsx` / `*.test.ts` (colocated next to the component/hook they exercise)
- MSW handlers: `web/test/msw/handlers.ts`

### Lint / Typecheck
```bash
cd web && npm run typecheck && npm run lint
```

---

## Issue traceability

This project tracks work by task path, not GitHub issue numbers. For dev agents working a Phase-3 task, use the **user story ID** (e.g. `US001`) as the traceability tag in commit messages instead of `#42`. Example:

```
red: test spec for ProjectHeader empty-name fallback (US004)
green: render fallback when project.name is empty (US004)
refactor: extract ProjectHeader name resolver (US004)
```
