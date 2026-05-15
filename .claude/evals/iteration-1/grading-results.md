# Eval Grading Results — iteration 1

Hand-graded after running each fixture in `skill_evals.json` and `agent_evals.json` from a fresh session.
Verdict legend: ✅ PASS · ⚠️ PARTIAL · ❌ FAIL · ⏭️ SKIPPED (with reason).

## Skills

| ID | Skill | Platform | Verdict | Notes |
|----|-------|----------|---------|-------|
| 1 | agile-product-owner | claude | _pending_ | |
| 1 | agile-product-owner | gemini | _pending_ | |
| 2 | focused-fix | claude | _pending_ | |
| 2 | focused-fix | gemini | _pending_ | |
| 3 | karpathy-coder | claude | _pending_ | |
| 3 | karpathy-coder | gemini | _pending_ | |
| 4 | senior-architect | claude | _pending_ | |
| 4 | senior-architect | gemini | _pending_ | |
| 5 | senior-backend | claude | _pending_ | |
| 5 | senior-backend | gemini | _pending_ | |
| 6 | senior-frontend | claude | _pending_ | |
| 6 | senior-frontend | gemini | _pending_ | |
| 7 | senior-qa | claude | _pending_ | |
| 7 | senior-qa | gemini | _pending_ | |
| 8 | tdd-guide | claude | _pending_ | |
| 8 | tdd-guide | gemini | _pending_ | |

## Agents

| ID | Agent | Scenario | Platform | Verdict | Notes |
|----|-------|----------|----------|---------|-------|
| po-ba-1 | po-ba | intake | claude | _pending_ | |
| po-ba-1 | po-ba | intake | gemini | _pending_ | Watch for: gemini po-ba has no AskUserQuestion equivalent in its tools list — does it still ask before writing? |
| po-ba-2 | po-ba | sign-off | claude | _pending_ | |
| po-ba-2 | po-ba | sign-off | gemini | _pending_ | |
| system-architect-1 | system-architect | draft | claude | _pending_ | |
| system-architect-1 | system-architect | draft | gemini | _pending_ | |
| tech-lead-1 | tech-lead | plan | claude | _pending_ | |
| tech-lead-1 | tech-lead | plan | gemini | _pending_ | |
| tech-lead-2 | tech-lead | review | claude | _pending_ | |
| tech-lead-2 | tech-lead | review | gemini | _pending_ | |
| tester-1 | tester | author | claude | _pending_ | |
| tester-1 | tester | author | gemini | _pending_ | |
| be-dev-1 | be-dev | happy path | claude | _pending_ | |
| be-dev-1 | be-dev | happy path | gemini | _pending_ | |
| be-dev-2 | be-dev | wrong track | claude | _pending_ | |
| be-dev-2 | be-dev | wrong track | gemini | _pending_ | |
| fe-dev-1 | fe-dev | happy path | claude | _pending_ | |
| fe-dev-1 | fe-dev | happy path | gemini | _pending_ | |

## Detailed grading

### Skill — _example_ (replace with real results)

**Assertions met:**
- [ ] Hits every bullet in the fixture's `expected_output`.

**Notes:**

---

### Agent — _example_ (replace with real results)

**Artifacts produced vs expected:**
- [ ] File path matches the glob in `expected_artifacts`.
- [ ] Status header transitions match the state machine.

**Behavior:**
- [ ] Each bullet in `expected_behavior`.

**Anti-patterns triggered:**
- [ ] None — list any that were.

**Verdict:** ✅ / ⚠️ / ❌

**Notes:**

---

## Conclusions

(Filled in after the iteration.)

- Cross-platform parity:
- Known issues:
- Action items for iteration-2:
