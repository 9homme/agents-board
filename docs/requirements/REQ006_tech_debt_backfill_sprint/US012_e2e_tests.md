# US012 — E2E test specification

No new e2e: tech-debt backfill scope per architecture §1.2 (anti-scope). Existing tests/e2e/ tests must remain green; no new .robot files.

Note: `make e2e-up && make e2e-run && make e2e-down` (3 consecutive clean runs) is the live-e2e regression bar for US012 per architecture.md §10.1, and is tracked in `US012_be_unit_tests.md` IT-005. This is a reuse of existing Robot Framework tests — no new `.robot` files are created.
