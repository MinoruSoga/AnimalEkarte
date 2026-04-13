# BUG-357: frontend デッドコード一括削除

## 概要

デッドコードスキャン（2026-04-14）により、frontend/src/ 配下で 40+ のデッドシンボル・10+ のデッドファイルを検出。全て grep で参照ゼロを確認済み。

## 優先度

**MEDIUM** — バンドルサイズへの直接影響は tree-shaking で軽減されるが、保守性・可読性・開発者の混乱防止の観点で削除すべき。

---

## A. 削除対象ファイル（参照ゼロ）

| # | ファイル | 理由 |
|---|---------|------|
| 1 | `features/owners/api/get-animal-species.ts` | `hooks/use-animal-species.ts` に置換済み |
| 2 | `features/owners/api/get-owners.ts` | loader パターンに移行済み |
| 3 | `features/medical-records/api/estimates.ts` | `save-estimate.ts` に置換済み |
| 4 | `features/accounting/api/delete-accounting.ts` | UI 未実装のまま放置 |
| 5 | `features/hospitalization/components/DailyRecord/` (8ファイル) | `DailyRecordsTab/` に完全置換済み |
| 6 | `features/hospitalization/hooks/use-daily-record-logic.ts` | DailyRecord/ のみが使用（共に死亡） |
| 7 | `features/trimming/api/types.ts` | `@/types/trimming` から直接 import に移行済み |
| 8 | `features/shifts/routes/ShiftCalendar.tsx` | alias ファイル、参照ゼロ |
| 9 | `features/master/api/examination-types.ts` | endpoint 未接続 |
| 10 | `features/hospital-settings/types/index.ts` | `Clinic` 型を直接使用に移行済み |
| 11 | `features/reservations/api/get-reservation.ts` | 単体取得未使用（リスト取得のみ） |

## B. 削除対象シンボル（ファイルは残す）

### features/

| ファイル | デッドシンボル |
|---------|--------------|
| `owners/index.ts` | `OwnerData`, `PetFormData`, `MembershipType` の export 行（内部使用は継続） |
| `reservations/api/get-reservation-types.ts` | `ReservationTypeRaw`, `GroupedReservationTypes` interface |
| `reservations/api/get-on-duty-staffs.ts` | `OnDutyStaff` interface |
| `reservations/types/index.ts` | `createDefaultReservationFormData`, `ReservationFormSaveHandler` |
| `reception/api/transforms.ts` | `COLUMN_ID_TO_TITLE` export（テスト専用） |
| `hospitalization/types/index.ts` | `CreateHospitalizationDTO`, `UpdateHospitalizationDTO` |
| `accounting/api/update-billing-item.ts` | `useUpdateBillingItem` hook |
| `estimates/api/types.ts` | `BackendOwnerSummary` type |
| `shifts/api/create-shift.ts` | `useCreateShift` hook |
| `shifts/api/update-shift.ts` | `useUpdateShift` hook |
| `shifts/index.ts` | `getShiftTemplates`, `createShiftTemplate`, `updateShiftTemplate`, `deleteShiftTemplate`, `reorderShiftTemplates`, `SHIFT_TYPE_COLORS` の re-export |
| `master/index.ts` | `useReorderReservationTypeGroups` re-export |

### lib / hooks / utils / types

| ファイル | デッドシンボル |
|---------|--------------|
| `lib/query-keys.ts` | 13 factory groups（owners, pets, reservations, medicalRecords, vaccinations, examinations, hospitalization, inventory, trimming, estimates, checkups, reception, shifts）+ masters 配下 4 sub-factories（staffs, animalSpecies, cages, medicines） |
| `hooks/use-sortable-data.ts` | `ActiveSortItem` の export keyword |
| `utils/format/date.ts` | `formatDateJapanese` 関数 |
| `types/index.ts` | `StaffMember`, `InsuranceCompany`, `TrimmingCourse`, `TrimmingOption`, `ExaminationType`, `BackendHospitalization`, `BackendCarePlanItem`, `BackendDailyRecord`, `BackendMedicalRecord`, `MasterCategory` |

## 検出方法

4 並列 refactor-cleaner エージェントによる全数 grep 検証（2026-04-14）

## タグ

- `dead-code`
- `bundle-size`
- `refactor`
- `medium-priority`
