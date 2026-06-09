# Shared ref — circuit breaker (3 strikes)

Applies to **tech-lead-reviewer** (per task, counting `### Review pass N` entries) and **po-ba** (per story, counting `### Sign-off pass N` entries). Three consecutive `changes_requested` verdicts on the same task or story trip the breaker.

## Rule

Before issuing a `changes_requested` verdict, count the existing same-artifact log entries whose verdict was `changes_requested`.

If this would be the **3rd consecutive `changes_requested`** (the artifact has already failed twice and is failing again):

1. **DO NOT** issue a 3rd `changes_requested`.
2. Set the artifact to `Status: blocked_circuit_breaker`.
3. Append a final log entry titled `CIRCUIT BREAKER TRIPPED` listing:
   - the recurring issue(s) across the three passes,
   - what was tried each time,
   - your hypothesis for why the loop is stuck (architecture wrong? spec wrong? requirement/AC wrong? skill gap?).
4. **Stop. Report `CIRCUIT_BREAKER_TRIPPED`** to the orchestrator with the artifact path.

## Reset

A "consecutive" streak resets to zero only on an `approved` pass. If you approve and a *later* round of rework comes back, the counter starts fresh.

## Non-negotiable

Never bypass the breaker. Even if you think the fix is one small change away, say so in your report — but still trip the breaker so a human decides. The orchestrator pauses the whole pipeline for that requirement and surfaces the three failed passes + your hypothesis to the human via `AskUserQuestion`; resume only on explicit human direction. **The orchestrator must never override the breaker.**
