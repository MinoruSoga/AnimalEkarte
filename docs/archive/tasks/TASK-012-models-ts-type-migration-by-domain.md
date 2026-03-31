# TASK-012: ドメイン別 models.ts 型移行

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー

---

## 概要

各ドメイン（feature）のコードで手書きの型定義を廃止し、`frontend/src/types/generated/models.ts`（tygo自動生成）から導出した型を使用するように統一する。

## 依頼内容（原文）

> ドメイン毎に、ドメイン関連のコードの型使用にて、frontend/src/types/generated/models.ts の型を使用するようにしてください。例えば、1ドメイン=飼主・ペット　みたいな感じです。

## 型パイプライン（プロジェクト標準）

```
backend/internal/model/*.go
    ↓ make codegen（tygo）
src/types/generated/models.ts     ← 自動生成（直接編集禁止）
    ↓ Omit / Partial / Pick で導出
features/xxx/api/types.ts         ← APIリクエスト型（models.tsから導出）
    ↓ transform関数
src/lib/transforms/xxx.ts         ← ReturnType<typeof transform> で型推論
```

## ドメイン別移行状況

| # | ドメイン | feature | 移行状況 | 問題点 | イシュー |
|---|---------|---------|---------|-------|---------|
| 1 | 飼主・ペット | owners, pets | 🟡 部分的 | api/types.ts は models.ts 済み。types/index.ts に PetGender/MembershipType/PetFormData/OwnerData 等5個手書き残存、src/types/owner.ts も models.ts 非依存 | FE-054 |
| 2 | 見積書 | estimates | ❌ 未着手 | api/types.ts に BackendEstimate/BackendEstimateItem 完全手書き、types/index.ts に EstimateStatus/Estimate/EstimateLineItem 3個手書き | FE-040 |
| 3 | 予約 | reservations | 🟡 部分的 | transforms/create/update で models.ts import 済み。api/types.ts の Request 型 + types/index.ts に ReservationFormData 等3個手書き | FE-041 |
| 4 | ダッシュボード | dashboard | 🟡 部分的 | transforms/get-dashboard で models.ts import 済み。api/types.ts に DashboardAppointment/DashboardColumn/UpdateAppointmentStatusRequest 3個手書き | FE-042 |
| 5 | カルテ | medical-records | 🟡 部分的 | api/types.ts は models.ts 済み。Request 型手書き + types/index.ts に Treatment/BillingReview/Vital 等10個手書き | FE-043 |
| 6 | 検査 | examinations | 🟡 部分的 | api/types.ts は models.ts 済み。Request 型手書き + transforms が src/types/index.ts の ExaminationRecord 手書き型を使用 | FE-044 |
| 7 | 予防接種 | vaccinations | 🟡 部分的 | api/types.ts は models.ts 済み。Request 型手書き | FE-045 |
| 8 | 入院 | hospitalization | 🟡 部分的 | api/types.ts は models.ts 済み。Request 型手書き + types/index.ts に Task/TimelineItem/HospitalizationFormData 等10個手書き | FE-046 |
| 9 | 会計 | accounting | 🟡 部分的 | api/types.ts は models.ts 済み。Request 型手書き + types/index.ts に AccountingStatus/AccountingItem 等5個手書き | FE-047 |
| 10 | トリミング・在庫 | trimming, inventory | 🟡 部分的 | Request 型が src/types/ に手書き | FE-048 |
| 11 | 認証 | auth | ❌ 未着手 | 8個の手書き型、models.ts 完全非接続 | FE-049 |
| 12 | 病院設定 | hospital-settings | 🟡 部分的 | Request 型（UpdateClinic, CreateStaff等）手書き + types/index.ts に ClinicInfo 1個手書き | FE-050 |
| 13 | マスタ設定 | master | 🟡 部分的 | 旧STI型（CreateMasterItemRequest等）が残存 | FE-051 |
| 14 | シフト | shifts | ❌ 未着手 | 全型手書き、models.ts 完全非接続 | FE-052 |
| 15 | 共有型 | src/types/ | ❌ 未着手 | index.ts に27個手書き型、owner.ts も models.ts 非依存 | FE-053 |

## 影響範囲

### DB / Backend
- 変更なし（models.ts は既存の tygo 生成済み）

### Frontend
- 各 feature の `api/types.ts` — 手書き interface → models.ts Omit/Partial 導出
- `src/types/index.ts` — 旧UI型を各 feature に移動 or 削除
- `src/types/*.ts` — 手書き型を models.ts 導出に書き換え
- `src/lib/transforms/` — transform 関数の入出力型を models.ts に統一

## 移行済み（対応不要）

- **src/types/pet.ts, diagnosis.ts, medicine.ts, treatment.ts, service-type.ts, trimming.ts** — models.ts import 済み
- **src/lib/transforms/** — 全て models.ts 依存、ReturnType パターン実装済み

※ owners/pets は api 層は移行済みだが types/index.ts + src/types/owner.ts に手書き残存（FE-054）

## 実装順序

1. FE-053（共有型整理 — 他ドメインの移行で参照されるため先に実施）
2. FE-040〜042, FE-049, FE-052（完全未着手ドメイン）
3. FE-043〜048, FE-050〜051（部分移行済みドメイン — 並行可）

## 関連イシュー（15件）

### 未着手（Priority: High） — models.ts import が api/types.ts にすらない
- FE-040: [estimates 型移行](../../frontend/issues/open/FE-040-estimates-type-migration.md)
- FE-049: [auth 型移行](../../frontend/issues/open/FE-049-auth-type-migration.md)
- FE-052: [shifts 型移行](../../frontend/issues/open/FE-052-shifts-type-migration.md)

### 部分的移行（Priority: Medium） — api/types.ts は models.ts 済み、Request型 + feature types/ に手書き残存
- FE-041: [reservations 型移行](../../frontend/issues/open/FE-041-reservations-type-migration.md)
- FE-042: [dashboard 型移行](../../frontend/issues/open/FE-042-dashboard-type-migration.md)
- FE-043: [medical-records 型移行](../../frontend/issues/open/FE-043-medical-records-type-migration.md)
- FE-044: [examinations 型移行](../../frontend/issues/open/FE-044-examinations-type-migration.md)
- FE-045: [vaccinations 型移行](../../frontend/issues/open/FE-045-vaccinations-type-migration.md)
- FE-046: [hospitalization 型移行](../../frontend/issues/open/FE-046-hospitalization-type-migration.md)
- FE-047: [accounting 型移行](../../frontend/issues/open/FE-047-accounting-type-migration.md)
- FE-048: [trimming・inventory 型移行](../../frontend/issues/open/FE-048-trimming-inventory-type-migration.md)
- FE-050: [hospital-settings 型移行](../../frontend/issues/open/FE-050-hospital-settings-type-migration.md)
- FE-051: [master 旧STI型整理](../../frontend/issues/open/FE-051-master-type-cleanup.md)
- FE-054: [owners 手書き型移行](../../frontend/issues/open/FE-054-owners-type-migration.md)

### 共有型整理（先行実施）
- FE-053: [src/types/ 共有型整理](../../frontend/issues/open/FE-053-shared-types-cleanup.md)
