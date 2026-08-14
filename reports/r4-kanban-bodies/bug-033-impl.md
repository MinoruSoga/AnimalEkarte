# [実装] BUG-033 completed examination lock residual

## Scope (single BUG)
Only BUG-033: completed examination must not allow result edit / item delete / item add / destructive save when first-pass seal applies. UAT 2026-08-13 reproduced on exam id 1014563 (completed still editable in browser).

## SoT
- bug.md BUG-033 / residual notes (2026-08-12/13)
- frontend/src/features/examinations/lib/examination-lock.ts
- ExaminationForm, ExaminationFormFields, ExamItemsTable, use-examination-form
- backend medicalrecord examination lock (already has RejectsCompletedSeal tests)

## Context (Phase0)
- Focused FE unit tests on main already PASS for lock helpers.
- Browser still NG on real completed exam → find gap (e.g. seal only when currentRevisionVersion==null, status JP/EN map, disabled not applied to table actions, BE status != FE).
- Do not treat as pure ENV_STALE; fix real residual if present. If truly data-specific (revision history intentional unlock), document and leave minimal guarded behavior matching product intent with test.

## DoD
1. Minimal FE and/or BE change so completed seal locks results/add/remove consistently for UAT case.
2. Scoped tests green (existing examination-lock + form fields + any new case for 1014563-like completed).
3. No other BUG scope. No migrate apply. No merge/push (orchestrator lands after verify).
4. Handoff note: branch tip SHA, files, test commands+exit, residual risks.
5. Status at most IMPLEMENTED_UNVERIFIED in notes; never VERIFIED_FIXED.

## Forbidden
- Other BUG IDs in same PR
- force-push, staging merge, make migrate, DB reset
- Linear Done / VERIFIED_FIXED
