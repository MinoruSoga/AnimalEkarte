# [実装] BUG-038 clinic master list empty vs DB

## Scope
BUG-038 only. /settings/clinic shows 0 clinics while DB has 4 active clinics. Fix list path (API scope/permissions/transform/error handling) so authorized staff with hospital-settings view see clinics.

## SoT
- bug.md BUG-038
- ClinicMasterSettings + useGetClinics (scope=all)
- BE GET /v1/clinics scope=all + hospital-settings.can_view

## DoD
1. Root cause fixed with minimal change; empty list only when truly none or unauthorized (clear error, not silent zero).
2. Scoped tests FE and/or BE.
3. Single BUG. No migrate/merge/push.
4. Handoff: how to verify with demo user.

## Forbidden
Other BUGs, force-push, staging, migrate, Done, secrets in logs.
