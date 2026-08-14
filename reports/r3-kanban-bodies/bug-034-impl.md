# [実装] BUG-034 — treatment policy lost after finalize reload

## Scope
- Only BUG-034. UAT: save 「UAT再検証 治療方針」 on 問診「治療方針」, finalize, reload → resets to「# 治療方針」. Chief complaint kept.
- Investigate dual fields: treatmentPolicy (DEFAULT_TREATMENT_POLICY / inquiry notes) vs plan (clinical_plan.treatment_policy).
- Minimal fix: persist + hydrate the field the UI labels 治療方針 so finalize reload keeps user value.
- Out of scope: 035 lock, other tabs, merge, push, migrate.

## Hints
- use-medical-record-save-action: 問診 notes = treatmentPolicy if !== default
- use-apply-medical-record: setTreatmentPolicy only if existingRecord.notes truthy
- Finalize path may omit notes or overwrite; clinical_plan vs inquiry mapping mismatch

## DoD
1. Root cause identified in handoff (fact).
2. Reload after save+finalize keeps treatment policy text (unit/vitest and/or focused integration).
3. No regression on chief complaint / clinical_plan physical_exam.
4. Docker tests with --entrypoint ''.
5. Handoff + review-required. No merge/push.
