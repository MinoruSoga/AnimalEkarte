# FE-053: src/types/ 共有型ファイルの整理 — models.ts 導出化

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

`src/types/index.ts` に27個の手書きドメイン型が残存し、models.ts を import していない。`src/types/owner.ts` も models.ts 非依存。整理が必要。

## 現状

### src/types/index.ts（27個の手書き型）
- ClinicInfo, MenuItem, Appointment, ColumnData
- MedicalRecord, Hospitalization, CarePlanItem, DailyRecord, TreatmentPlan
- ReservationAppointment, TrimmingRecord
- ExaminationRecord, ExaminationItem, VaccinationRecord
- MasterItem, StaffMember, StaffRole
- InsuranceCompany, CoverageRate
- TrimmingCourse, TrimmingOption, TargetSize, Combinable
- ExaminationTypeInspection, ExaminationType, InventoryItem

### src/types/owner.ts（models.ts 非依存）
- Owner, CreateOwnerRequest, UpdateOwnerRequest — 全て手書き

### 移行済み（問題なし）
- `src/types/pet.ts` — models.ts import 済み ✅
- `src/types/diagnosis.ts` — models.ts import 済み ✅
- `src/types/medicine.ts` — models.ts import 済み ✅
- `src/types/treatment.ts` — models.ts import 済み ✅
- `src/types/service-type.ts` — models.ts import 済み ✅
- `src/types/trimming.ts` — models.ts import 済み ✅

## 必要な変更

### 1. src/types/owner.ts
- `Owner` → models.ts の `Owner` を import して使用
- `CreateOwnerRequest` → `Omit<Owner, 'id' | ...>` で導出
- `UpdateOwnerRequest` → `Partial<CreateOwnerRequest>` で導出

### 2. src/types/index.ts
- models.ts に対応がある型（MedicalRecord, Hospitalization, Vaccination 等）→ models.ts から re-export
- UI固有型（ColumnData, MenuItem 等）→ 各 feature の types/ に移動 or そのまま残す
- 使われていない型 → 削除

## 完了条件

- [ ] `src/types/owner.ts` が models.ts から型を導出
- [ ] `src/types/index.ts` の手書き型のうち、models.ts に対応があるものは import に置換
- [ ] 使用箇所の型参照が更新されている
- [ ] `npm run build` 成功・型エラーなし
