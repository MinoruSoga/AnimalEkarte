# [実装] BUG-037 hospitalization required fields (cage)

## Scope
BUG-037 only. New hospitalization must not save without required cage (and align care-plan required rules with product). Evidence: hospitalizations id=8 cage_id=NULL after empty-ish submit from /hospitalization/new?petId=1000019.

## SoT
- bug.md BUG-037
- use-hospitalization-form.ts validation
- HospitalizationForm UI errors
- BE create hospitalization validation if absent

## DoD
1. FE blocks with field errors; BE fail-closed preferred.
2. Scoped tests.
3. Single BUG. No migrate/merge/push.
4. Handoff.

## Forbidden
Other BUGs, force-push, staging, migrate, Done.
