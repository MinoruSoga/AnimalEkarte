# [実装] BUG-035 — finalized medical record still editable + save visible

## Scope
- Only BUG-035. UAT record 1425559: banner says locked but inputs disabled=false and save still shown.
- Likely FE: fieldset disabled={isFinalized} with className **contents** (display:contents breaks disabled cascade).
- Also ensure BE rejects updates on finalized (defense in depth) if gap.
- Out of scope: exam 033, 034 policy content, merge, push, migrate.

## DoD
1. When status 確定済: all clinical inputs non-editable (not rely solely on broken fieldset contents); save/finalize hidden or no-op; addendum path remains for allowed edits.
2. Vitest proves disabled/read-only when finalized.
3. BE update reject on finalized if not already (test).
4. Docker --entrypoint ''.
5. Handoff + review-required. No merge/push.
