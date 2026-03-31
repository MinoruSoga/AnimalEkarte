# FE-072: Feature hooks ファイル名 kebab-case 化

**Status**: Closed
**Priority**: High
**Affects**: 全 feature の `hooks/` ディレクトリ（11 feature, 21 ファイル）
**Date Created**: 2026-03-18
**Related**: TASK-018, FE-071

## Summary

`frontend/src/features/*/hooks/` 内の 21 ファイルが camelCase で命名されている。プロジェクト規約に準拠するため kebab-case にリネームし、全 import パスと barrel index を更新する。

## 現状のコード

### 違反ファイル一覧（21 ファイル、feature 別）

#### auth（1 ファイル）
```
features/auth/hooks/useAuth.tsx → use-auth.tsx
```

#### owners（1 ファイル）
```
features/owners/hooks/useOwnerForm.ts → use-owner-form.ts
```

#### dashboard（1 ファイル）
```
features/dashboard/hooks/useDashboardKanban.ts → use-dashboard-kanban.ts
```

#### estimates（1 ファイル）
```
features/estimates/hooks/useEstimateForm.ts → use-estimate-form.ts
```

#### examinations（2 ファイル）
```
features/examinations/hooks/useExaminationForm.ts → use-examination-form.ts
features/examinations/hooks/useExaminationRecords.ts → use-examination-records.ts
```

#### hospitalization（5 ファイル）
```
features/hospitalization/hooks/useHospitalizationForm.ts → use-hospitalization-form.ts
features/hospitalization/hooks/useHospitalizationList.ts → use-hospitalization-list.ts
features/hospitalization/hooks/useHospitalizationDetail.ts → use-hospitalization-detail.ts
features/hospitalization/hooks/useHospitalizations.ts → use-hospitalizations.ts
features/hospitalization/hooks/useDailyRecordLogic.ts → use-daily-record-logic.ts
```

#### hospital-settings（1 ファイル）
```
features/hospital-settings/hooks/useClinicSettingsForm.ts → use-clinic-settings-form.ts
```

#### inventory（1 ファイル）
```
features/inventory/hooks/useInventory.ts → use-inventory.ts
```

#### master（1 ファイル）
```
features/master/hooks/useServiceTypeColorMap.ts → use-service-type-color-map.ts
```

#### medical-records（2 ファイル）
```
features/medical-records/hooks/useMedicalRecords.ts → use-medical-records.ts
features/medical-records/hooks/useMedicalRecordForm.ts → use-medical-record-form.ts
```

#### reservations（1 ファイル）
```
features/reservations/hooks/useReservationManagement.ts → use-reservation-management.ts
```

#### trimming（2 ファイル）
```
features/trimming/hooks/useTrimmingForm.ts → use-trimming-form.ts
features/trimming/hooks/useTrimmingRecords.ts → use-trimming-records.ts
```

#### vaccinations（2 ファイル）
```
features/vaccinations/hooks/useVaccinationForm.ts → use-vaccination-form.ts
features/vaccinations/hooks/useVaccinations.ts → use-vaccinations.ts
```

### import 元ファイル一覧（~20 箇所）

#### 相対 import（feature 内部）
```typescript
// features/owners/routes/OwnerForm.tsx:53
import { useOwnerForm } from "../hooks/useOwnerForm";

// features/dashboard/routes/Dashboard.tsx:32
import { useDashboardKanban } from "../hooks/useDashboardKanban";

// features/estimates/routes/EstimateForm.tsx:19
import { useEstimateForm } from '../hooks/useEstimateForm';

// features/examinations/routes/ExaminationForm.tsx:20
import { useExaminationForm } from "../hooks/useExaminationForm";

// features/examinations/routes/Examinations.tsx:22
import { useExaminationRecords } from "../hooks/useExaminationRecords";

// features/hospitalization/routes/HospitalizationForm.tsx:17
import { useHospitalizationForm } from "../hooks/useHospitalizationForm";

// features/hospitalization/routes/HospitalizationList.tsx:17
import { useHospitalizationList } from "../hooks/useHospitalizationList";

// features/hospitalization/routes/HospitalizationDetail.tsx:13
import { useHospitalizationDetail } from "../hooks/useHospitalizationDetail";

// features/hospitalization/hooks/useHospitalizationList.ts:3
import { useHospitalizations } from "./useHospitalizations";

// features/hospitalization/components/DailyRecord/DailyRecordSection.tsx:12
import { useDailyRecordLogic } from "../../hooks/useDailyRecordLogic";

// features/hospital-settings/routes/ClinicSettings.tsx:25
import { useClinicSettingsForm } from "../hooks/useClinicSettingsForm";

// features/inventory/routes/InventoryList.tsx:34
import { useInventory } from "../hooks/useInventory";

// features/medical-records/routes/MedicalRecords.tsx:26
import { useMedicalRecords } from "../hooks/useMedicalRecords";

// features/medical-records/routes/MedicalRecordForm.tsx:29
import { useMedicalRecordForm } from "../hooks/useMedicalRecordForm";

// features/reservations/routes/ReservationManagement.tsx:27
import { useReservationManagement } from "../hooks/useReservationManagement";

// features/trimming/routes/TrimmingForm.tsx:34
import { useTrimmingForm } from "../hooks/useTrimmingForm";

// features/trimming/routes/TrimmingList.tsx:36
import { useTrimmingRecords } from "../hooks/useTrimmingRecords";

// features/vaccinations/routes/VaccinationForm.tsx:29
import { useVaccinationForm } from "../hooks/useVaccinationForm";

// features/vaccinations/routes/VaccinationList.tsx:21
import { useVaccinations } from "../hooks/useVaccinations";
```

#### cross-feature import（注意: 規約上は非推奨だが既存コード）
```typescript
// features/auth/index.ts:1
export { AuthProvider, useAuth } from "./hooks/useAuth";

// components/shared/Layout/Sidebar.tsx:31
import { useAuth } from "@/features/auth/hooks/useAuth";

// components/shared/Layout/Layout.tsx:3
import { useAuth } from "@/features/auth/hooks/useAuth";

// features/accounting/routes/AccountingDetail.tsx:45
import { useAuth } from "@/features/auth/hooks/useAuth";

// features/reservations/routes/ReservationManagement.tsx:28
import { useServiceTypeColorMap } from "@/features/master/hooks/useServiceTypeColorMap";

// features/reservations/components/ReservationDetailModal.tsx:26
import { useServiceTypeColorMap } from "@/features/master/hooks/useServiceTypeColorMap";
```

#### barrel index 更新が必要
```typescript
// features/hospitalization/hooks/index.ts:1
export { useHospitalizations } from "./useHospitalizations";

// features/hospitalization/hooks/index.ts:5
export { useDailyRecordLogic } from "./useDailyRecordLogic";
```

## 必要な変更

### 1. ファイル rename（git mv）

各 feature ディレクトリで `git mv` を実行。例：

```bash
cd frontend/src/features/auth/hooks
git mv useAuth.tsx use-auth.tsx

cd frontend/src/features/owners/hooks
git mv useOwnerForm.ts use-owner-form.ts

# ... 以下 21 ファイル分
```

### 2. import パス更新

全ファイルで import パスの camelCase 部分を kebab-case に置換。

| 変更前パターン | 変更後パターン |
|--------------|--------------|
| `../hooks/useOwnerForm` | `../hooks/use-owner-form` |
| `../hooks/useDashboardKanban` | `../hooks/use-dashboard-kanban` |
| `@/features/auth/hooks/useAuth` | `@/features/auth/hooks/use-auth` |
| `@/features/master/hooks/useServiceTypeColorMap` | `@/features/master/hooks/use-service-type-color-map` |
| （他 17 パターン） | |

### 3. barrel index 更新

```typescript
// features/hospitalization/hooks/index.ts
// Before:
export { useHospitalizations } from "./useHospitalizations";
export { useDailyRecordLogic } from "./useDailyRecordLogic";

// After:
export { useHospitalizations } from "./use-hospitalizations";
export { useDailyRecordLogic } from "./use-daily-record-logic";
```

```typescript
// features/auth/index.ts
// Before:
export { AuthProvider, useAuth } from "./hooks/useAuth";

// After:
export { AuthProvider, useAuth } from "./hooks/use-auth";
```

**注意**: 関数名（`useOwnerForm` 等）は変更しない。ファイル名と import パスのみ。

## プロジェクトルール遵守チェック

- [ ] ファイル名が kebab-case
- [ ] import パスが新ファイル名を参照
- [ ] barrel index が更新済み
- [ ] 関数名は変更しない（camelCase のまま正しい）
- [ ] cross-feature import のパスも更新済み

## 依存関係

- FE-071 との依存はない（並列着手可能）
- ただし FE-071 を先に完了すると、shared hooks の rename が済んだ状態で feature hooks に集中できる

## 完了条件

- [ ] 21 ファイルが kebab-case にリネーム済み
- [ ] 全 import パスが更新済み（相対・絶対・cross-feature）
- [ ] barrel index（`hospitalization/hooks/index.ts`, `auth/index.ts`）が更新済み
- [ ] `npm run build` パス
- [ ] `npm run lint` パス
- [ ] `grep -rn "hooks/use[A-Z]" frontend/src --include="*.ts" --include="*.tsx"` で camelCase の import パスが 0 件

## クローズ情報

- **Closed At**: 2026-03-18
- **変更ファイル**:
  - 21 feature hook ファイルを kebab-case にリネーム（auth, owners, dashboard, estimates, examinations x2, hospitalization x5, hospital-settings, inventory, master, medical-records x2, reservations, trimming x2, vaccinations x2）
  - ~25箇所の import パス更新（相対・絶対・cross-feature）
  - barrel index 更新: hospitalization/hooks/index.ts, vaccinations/hooks/index.ts, examinations/hooks/index.ts, trimming/hooks/index.ts, medical-records/hooks/index.ts, auth/index.ts
