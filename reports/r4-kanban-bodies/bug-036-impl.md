# [実装] BUG-036 shift empty start/end times

## Scope
BUG-036 only. Shift create/update must require start_time and end_time for non-off shift types; empty must show error and not persist. DB evidence: shift_entries id=2460 date=2026-08-01 start/end NULL.

## SoT
- bug.md BUG-036
- frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx
- BE shift create/update validation if missing

## DoD
1. FE validation (and BE fail-closed if needed) for required times when shift_type needs times (not off/paid-leave if product hides times).
2. Scoped tests.
3. Single BUG. No migrate/merge/push.
4. Handoff with evidence.

## Forbidden
Other BUGs, force-push, staging, migrate apply, Done.
