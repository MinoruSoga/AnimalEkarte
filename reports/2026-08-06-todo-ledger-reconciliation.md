# todo.md ledger reconciliation (2026-08-06)

## Prompt-craft-graph / multi-agent posture

- Requested: clear all of `todo.md` under multi-agent prompt-craft-graph.
- Finding: **no remaining agent-implementable product-code unit** that is READY_AGENT without USER/clinical/PO gates.
- Therefore: **no v2 feature graph was started** (would be empty-diff / policy-blocked). Regeneration is appropriate only after TASK-033 or TASK-021 gates open.
- Providers available (admission only): Grok subscription OAuth; Claude Max subscription OAuth.

## Classification (todo.md body vs git main)

| Bucket | Count | Action taken |
|--------|------:|--------------|
| Product code / docs DONE on main | ~20 TASK sections | Moved to "対応済み" table; detail sections removed from open residual |
| USER / ops residual | ~15 rows | Kept as open residual with owner USER/ops/PO |
| BLOCKED clinical (TASK-033) | 1 | Kept; no implementation |
| BLOCKED destructive (TASK-021) | 1 | Kept; needs PO |
| Browser / UAT residual | TASK-010/020/022–024 | Kept; linked to BROWSER_VERIFICATION_BACKLOG |

## Verified commits on main (sample)

See `git merge-base --is-ancestor` for TASK-025/026/028/029/030 evidence hashes in prior agent audit.

## What "all done" means

| Layer | Status |
|-------|--------|
| Agent product code open | **0** |
| todo.md open residual | USER gates only |
| bug.md | 32/32 IU (browser deferred) |
| VERIFIED_FIXED / seed apply / migrate | USER |

## Next agent entry points

1. TASK-033 after clinical SoT + DB review
2. TASK-021 after external-use inventory + destroy approval
3. Browser batch (not a code graph)

## Post-audit correction

- TASK-026 evidence hash `2a8aca33c` is **not** in git; land evidence is merge `e572e941c`.
- Explore audit confirmed agent-implementable open = 0.
