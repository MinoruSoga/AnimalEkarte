# [型定義] 自動生成型 models.ts への完全移行

## 背景

バックエンド（Go）のモデル struct から TypeScript 型を自動生成するパイプラインを導入した。

```
backend/internal/model/*.go
    ↓ make codegen（tygo）
frontend/src/types/generated/models.ts
```

`make codegen` を実行するだけで、Go モデルの変更が TypeScript 型に自動反映される。
手動で型定義を書く必要はなくなった。

## 現在の状態

### 移行済み（models.ts を使用中）

| Feature | 移行済みの型 |
|---|---|
| `owners` | `Owner`, `Pet`, `AnimalSpecies`, `Insurance` |
| `pets` | `Pet` |
| `vaccinations` | `Vaccination` |
| `examinations` | `Exam`, `ExamItem` |
| `inventory` | `InventoryItem` |
| `hospital-settings` | `Clinic`, `Staff` |

### 未移行（手書き Backend* 型を使用中）

以下の feature はまだ手書きの `Backend*` interface を使用している。
Go モデルと乖離するリスクがあるため、`models.ts` の型に置き換えること。

| Feature | 手書き型ファイル | 置き換え対象 |
|---|---|---|
| `accounting` | `api/types.ts` | `BackendAccounting` → `Billing`, `BackendAccountingItem` → `BillingItem` |
| `hospitalization` | `api/types.ts` | `BackendHospitalization` → `Hospitalization`, `BackendCarePlanItem` → `CarePlanItem` など |
| `reservations` | `api/types.ts` | `BackendReservation` → `ReservationAppointment` |
| `dashboard` | `api/types.ts` | `BackendDashboardReservation` → `ReservationAppointment` |
| `medical-records` | `api/types.ts` | `BackendMedicalRecord` → `MedicalRecord` |
| `trimming` | `api/types.ts` | `BackendTrimming` → `TrimmingRecord` |
| `master` | `api/types.ts` | `BackendStaff` → `Staff` など |

## 移行方法

### 1. models.ts で利用可能な型を確認

```typescript
// frontend/src/types/generated/models.ts（自動生成）
export interface Owner { ... }
export interface Pet { ... }
export interface Billing { ... }  // 会計
export interface Hospitalization { ... }
export interface ReservationAppointment { ... }  // 予約
export interface MedicalRecord { ... }
export interface TrimmingRecord { ... }
// ... 全 Go モデルが揃っている
```

### 2. 移行パターン

```typescript
// Before（手書き）
export interface BackendAccounting {
  id: string;
  status: "未収" | "保留" | "回収済" | "キャンセル";
  // ... 手動で Go と同期が必要
}

// After（自動生成型を使用）
import type { Billing } from "@/types/generated/models";
export type BackendAccounting = Billing;
```

### 3. 型名の対応表

models.ts の型名は Go の struct 名と一致する。

| 画面上の概念 | models.ts の型名 |
|---|---|
| 会計 | `Billing` |
| 会計明細 | `BillingItem` |
| 入院 | `Hospitalization` |
| ケアプラン項目 | `CarePlanItem` |
| 入院プラン | `HospitalizationPlan` |
| 予約 | `ReservationAppointment` |
| カルテ | `MedicalRecord` |
| トリミング | `TrimmingRecord` |
| スタッフ | `Staff` |
| 診断カテゴリ | `DiagnosisCategory` |
| 薬剤 | `Medicine` |
| ワクチン | `Vaccine` |

### 4. 注意事項

**ID フィールドの型が変わる**

models.ts では ID が `number`（Go の uint64 由来）。
フロントエンド内部では文字列として扱うため、transforms で変換が必要。

```typescript
// transforms.ts でのパターン
id: String(data.id ?? 0),
owner_id: String(data.owner_id ?? 0),
```

**enum 型は string 型エイリアス + const**

Go の enum は以下のように生成される。

```typescript
// models.ts（自動生成）
export type BillingStatus = string;
export const BillingStatusWaiting: BillingStatus = "waiting";
export const BillingStatusCompleted: BillingStatus = "completed";
```

フロントエンドで literal union が必要な場合は `as` キャストを使う。

```typescript
status: data.status as "waiting" | "completed" | "cancelled",
```

**Request 型（Create/Update）は models.ts に存在しない**

`CreateAccountingRequest` のような入力型は Go モデルと別物なので手書きのまま残してよい。
置き換えるのは「バックエンドのレスポンス型」のみ。

## 完了条件

- [ ] 全 feature の `api/types.ts` から手書き `BackendXxx` interface が削除されている
- [ ] 全て `models.ts` からの import に置き換わっている
- [ ] `docker compose exec frontend npm run build` がエラーゼロで通る
- [ ] `make codegen` 実行後もビルドが通る（型の自動更新が機能している）

## 参考

- 移行済みサンプル: `frontend/src/features/owners/api/types.ts`
- 生成型: `frontend/src/types/generated/models.ts`（直接編集禁止）
- 生成コマンド: `make codegen`（backend/internal/model/ が更新されたら実行）
