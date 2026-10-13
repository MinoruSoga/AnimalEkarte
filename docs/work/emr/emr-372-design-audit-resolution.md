# EMR-372: Design-state audit — corrected finding resolution

## Authorization and scope

- User authorized fixing valid EMR-372 findings after Lead review corrected the
  original audit. The user requested all source findings corrected and a new main
  PR. Codex Lead owns PR creation and Plane state.
- Base: `ed04c0237f81ddc2890d39b528d3608a58180c8b` (clean baseline).
- Binding spec: `docs/spec/design-states.md` §1 (keyboard-visible focus contract)
  and `docs/spec/design-system.md` §3.4 (font-weight rules), read against
  `frontend/src/lib/design-tokens.ts`.
- Brief revisions: `1-test-first` (adjacent contract tests, Lead RED), then
  `2-implementation` (bounded fixes below).
- This document does not repeat the original route inventory; the original Plane
  audit is retained as the source of record.

## Corrected dispositions (IU-001 … IU-007)

Review corrections applied to the original audit before implementation:

- Shared `Textarea` already had a focus shadow and `CharCountTextarea` already
  had a visible ring. The valid defect is pointer-gated `focus:` indicators plus
  a missing ring offset — not total absence of focus indication. The withdrawn
  "missing ring" claim is not repeated; IU-004 is merged into the IU-002 root
  fix as removal of the now-redundant manual ring override.
- `HelpChat` belongs to the support widget
  (`AuthenticatedSupportWidget → SupportWidget/SupportPanel → ManualHelpTab →
  HelpChat`, gated by `chatStatus.enabled`), not `ChatHistoryPage`.

| ID   | Disposition | Change |
|------|-------------|--------|
| IU-001 | Implemented | `OwnerAccountingHistoryParts` summary `completedTotal`: `font-bold` → `font-semibold` (design-system §3.4: 700 prohibited for non-heading; precedent = `DailyAccountingTab` totals). |
| IU-002 | Implemented | `ui/textarea.tsx`: removed pointer `focus:bg-white` + `STATE.focusBorderLegacyAccent` + `STATE.focusRingActionPrimary`; applied canonical `focus-visible:ring-2` + `C.focusVisibleRingActionPrimary` + `focus-visible:ring-offset-1`. `STYLE.textarea`: replaced `C.focusBorderAccent` with the same canonical trio. Raw `LineSendPanel` textarea now composes `cn(STYLE.textarea, "font-sans", C.textPlaceholder)` (preserves non-mono typography and placeholder token). `STATE` tokens and `STYLE.formInput` untouched (still used by `Input`). |
| IU-003 | Implemented | `C.text40` → `C.text60` on the specified messages only: `MedicalRecordExamination` loading + empty, `MedicalRecordTabsShared` saved-gate, `DiagnosisHeaderChiefComplaint` empty. |
| IU-004 | Implemented (merged) | `CharCountTextarea`: removed redundant `focus-visible:ring-2` + `C.focusRingAccent40` override; canonical ring now inherited centrally from `STYLE.textarea`. `C.focusRingAccent40` is the same literal as `C.focusVisibleRingActionPrimary` — no `/40` difference claimed. |
| IU-005 | Implemented | `C.text40` → `C.text60` on specified messages only: `MedicineDoseParamsEditor` overdose warning + loading, `StaffBasicInfoSection` missing occupations, `LabDeviceItemMasterSidePanel` missing items, `ReservationTypeSidePanel` child-reservation instruction. |
| IU-006 | Implemented | `ShiftTemplateSidePanelFields` "休憩なし" value: `C.text40` → `C.text60`. |
| IU-007 | Implemented | `HelpChat`: 4 native buttons `disabled:opacity-40` → `disabled:opacity-50 disabled:pointer-events-none` (suggested questions, retry, submit, history reset). |

## Binding source paths

- `frontend/src/components/ui/textarea.tsx`
- `frontend/src/lib/design-tokens.ts` (`STYLE.textarea` only)
- `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.tsx`
- `frontend/src/features/owners/components/LineSendPanel.tsx`
- `frontend/src/features/accounting/components/OwnerAccountingHistoryParts.tsx`
- `frontend/src/features/medical-records/components/{MedicalRecordExamination,MedicalRecordTabsShared,DiagnosisHeaderChiefComplaint}.tsx`
- `frontend/src/features/master/components/{MedicineDoseParamsEditor,StaffBasicInfoSection,LabDeviceItemMasterSidePanel,ReservationTypeSidePanel}.tsx`
- `frontend/src/features/shifts/components/ShiftTemplateSidePanelFields.tsx`
- `frontend/src/features/support/components/HelpChat.tsx`

## Adjacent contract tests (authored revision 1)

- `frontend/src/components/ui/textarea.test.tsx`
- `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.test.tsx`

## Verification status

- Lead RED: `/tmp/ae-emr372-fusion-20261013/red.log` + `red-candidate.json` —
  isolated offline read-only-source Docker (image
  `sha256:73df0c1c01928468aa7115dc31cd7addbe5e4232359b1c1b018e37fa5faf383a`,
  `ekarte-frontend-node-modules` volume), `vitest --configLoader native
  --no-cache`, exit 1: 5 failing focus-contract tests, 9 passing native-behavior
  tests. RED actor: Codex Lead.
- Lead GREEN: 9 selected test files, 77 tests PASS (exit 0), including the
  two adjacent focus/native-behavior tests and existing owner LINE panel,
  owner accounting history/pagination, medical-record examination, staff,
  reservation-type and HelpChat regressions.
- Changed-file ESLint and Prettier: PASS (exit 0). Scoped TypeScript program
  rooted in the 16 changed production/test files plus Vite/testing declarations:
  0 diagnostics (exit 0); this is not the full application type-check.
- Coverage on `ui/textarea.tsx` and `CharCountTextarea.tsx` only:
  lines/statements/functions/branches 100% (14 selected tests, exit 0).
  This does not establish feature-wide or full-project coverage.
- `node scripts/design-system-audit.mjs`: C1–C24 PASS, 0 violations (exit 0).
  This proves only the script's mechanical rules, not visual/semantic acceptance.
- All checks above ran by Codex Lead inside Docker, with the exact worktree
  source and existing dependency volume mounted read-only, network disabled,
  no capabilities, no installs and no migrations. Image and dependency volume
  are the same as RED. Command/output/exit records are kept locally in
  `/tmp/ae-emr372-fusion-20261013/verification-results.json` and sibling logs.
- Rendering, CI, feature-level E2E and UAT remain unverified here. Browser
  rendering/focus/contrast acceptance remains the separate EMR-373 work item.

## Browser review

- Separate work item EMR-373, status UNKNOWN. Current Stagehand settings:
  `page_to_llm=false`, `external_writes=false` — unchanged by this task.

## Out of scope

- No global audit PASS claim; only the listed IU units were remediated.
- No blanket `C.text40` replacement; only source-matched specified messages.
- No clinical function, ARIA, error-state, mutation, config, hook, migration,
  or dependency changes.
