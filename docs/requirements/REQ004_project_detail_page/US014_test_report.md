# US014 — Markdown + Mermaid rendering · Test Report

**Generated:** 2026-05-30 (Asia/Bangkok)
**Commit at capture:** `HEAD` of `main` after both US014 tech-lead approvals (markdown: `ffc0c69`; mermaid: review pass entry transcribed by orchestrator)
**Story status at capture:** both FE tasks `Status: completed`; no BE tasks for US014
**Capture driver:** orchestrator (Phase 3c)

---

## Backend (Go — `services/agent-board`)

**N/A.** US014 has no BE tasks. The full BE suite remains green (107 passed, 0 failed) at the captured HEAD from US013 capture; no regressions introduced by the US014 FE-only work.

---

## Frontend (Jest + React Testing Library — `web`)

US014-scoped command: `cd web && npm test -- --watchAll=false --forceExit --testPathPatterns='MarkdownRenderer|Mermaid|DocumentPreviewer'`
Result: **`Test Suites: 3 passed, 3 total; Tests: 28 passed, 28 total`** (2.7 s).

Full FE suite: **107 tests / 16 suites pass** (US012 + US013 + US014 combined; no regressions).

| Spec ID | Test file(s) | Outcome |
|---|---|---|
| FCT-US014-001 (basic markdown renders as HTML, not text) | `web/components/ProjectDetail/MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-002 (GFM tables / task lists / strikethrough) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-003 (links autolink + safe target/rel) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-004 (inline code formatting) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-005 (fenced code blocks) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-006 (syntax highlighting — `language-go` + `hljs` + `<span>` tokens) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-007 (empty content placeholder unchanged) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-008 (mermaid code block → SVG; raw source hidden) | `MermaidDiagram.test.tsx` (×2) + `MarkdownRenderer.test.tsx` code-override routing (×3) | PASS |
| FCT-US014-009 (mermaid error → friendly fallback; no propagation) | `MermaidDiagram.test.tsx` (×2) | PASS |
| FCT-US014-010 (`mermaid.render` NOT called for non-mermaid code) | `MarkdownRenderer.test.tsx` (Go fence) | PASS |
| FCT-US014-011 (key-driven doc switch drops stale SVG, shows new SVG) | `DocumentPreviewer.test.tsx` | PASS |
| FCT-US014-012 (sanitization — `<script>alert(1)</script>` stripped; `window.alert` not called) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-013 (`javascript:` URL scheme on `[link](…)` neutralised) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-014 (`<img onerror=…>` handler stripped) | `MarkdownRenderer.test.tsx` | PASS |
| FCT-US014-015 (error boundary catches renderer throw without taking down the page) | `MarkdownRenderer.test.tsx` (boundary case) | PASS |

**Aux gates (per both FE tech-lead reviews):** `npm run typecheck` clean; `npm run lint -- --max-warnings=0` clean (`ESLint: No issues found`); `bash scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS`.

**Pre-existing FE gate-script hang:** unchanged from US012/US013 — `scripts/review/run-gate.sh fe` hangs at `npm test --watchAll=false` because MSW keeps an open handle and the gate doesn't pass `--forceExit`. All FE tech-lead reviews verified constituent checks individually.

**XSS evidence quality (per markdown review):** FCT-US014-012/013/014 exercise actual attacker payloads AND assert `window.alert` was not called — not lazy whitelist checks. Pipeline ordering verified: `rehype-sanitize` runs BEFORE `rehype-highlight` so attacker HTML is gone before highlighter classes are added. Schema allow-list whitelists `language-*` and `hljs(-*)` classes on `<code>`/`<pre>`/`<span>` so the highlighter's later-added classes survive sanitisation.

**Mermaid security stance (per mermaid review):** `MermaidDiagram` uses sanctioned `dangerouslySetInnerHTML` on mermaid-emitted SVG (architecture D-004 explicit). Dev relied on `mermaid.initialize({ securityLevel: 'strict' })` rather than double-sanitising through `rehype-sanitize` — architecture allows this.

**Open follow-up (non-blocking tech-debt, surfaced in mermaid review log):** `@testing-library/dom` was added to `dependencies` during mermaid install (npm peer-resolver surfaced an unrelated missing peer). Test-time tool only; should be moved to `devDependencies`. Doesn't break the build, doesn't ship to browser.

---

## E2E (Robot Framework — `tests/e2e/REQ004_project_detail_page`)

Command: `robot --dryrun --include US014 tests/e2e/REQ004_project_detail_page/`
Outcome: **PASS** (1/1 tests parsed cleanly; the US012 import-path fix `31f162d` made US014's suite parse-ready alongside US012's).

| Spec ID | Test (Robot) | Outcome |
|---|---|---|
| E2E-US014-001 (markdown + mermaid render in the previewer end-to-end) | `tests/e2e/REQ004_project_detail_page/US014_markdown_mermaid.robot` | BLOCKED (no live stack) |

**Status:** dry-run clean (1/1 parsed). Live execution NOT performed — orchestrator does not automate `web` + `api-server` + seeded-DB stack-up. Same release-gate posture as US012 pass 2 and US013 pass 2.

---

## Skipped tests — called out

- **No BE tests applicable (US014 is FE-only).**
- **No FE tests skipped.**
- **E2E (E2E-US014-001):** not executed against a live stack; dry-run passes. Live execution is a release-gate item that requires standing up `cd web && npm run dev` (CSR) + `cd services/agent-board && go run ./cmd/api-server` against a seeded DB.
