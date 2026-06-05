# US015 — E2E test specification

No new e2e: tech-debt backfill scope per architecture §1.2 (anti-scope). No new .robot files.

The e2e regression bar (3 consecutive clean runs of `make e2e-up && make e2e-seed && make e2e-run && make e2e-down`) is the live-e2e gate for US015 and is tracked as IT-002 in `US015_be_unit_tests.md`. This is a reuse of existing Robot Framework tests — no new scenarios are added.
