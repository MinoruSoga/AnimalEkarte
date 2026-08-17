# [実装] BUG-035 finalized medical record edit lock residual

## Scope
Only BUG-035: finalized medical record must not allow clinical field edit or primary save. UAT 2026-08-13: record 1425559 showed finalized banner but chief complaint / treatment policy disabled=false and save remained.

## SoT
- bug.md BUG-035
- MedicalRecordForm fieldset lock (no display:contents)
- use-medical-record-form isFinalized (status === 確定済)
- MedicalRecordTabsArea / floating actions / inquiry notes path (BUG-034 already landed)

## Context
- permissions unit tests PASS on main.
- Browser residual → inspect nested controls ignoring fieldset, wrong status string, save button outside gate, portal, or status not 確定済 while UI label looks finalized.

## DoD
1. Minimal fix so finalized MR clinical inputs disabled and primary save hidden/blocked.
2. Scoped vitest green (+ new regression if gap found).
3. Single BUG only. No migrate/merge/push.
4. Handoff: SHA, files, tests, risks.
5. Never VERIFIED_FIXED.

## Forbidden
Other BUGs, force-push, staging, migrate apply, Linear Done.
