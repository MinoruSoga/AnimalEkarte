# 005: HospitalizationSection フォームフィールド追加

**ステータス:** open
**優先度:** medium
**関連API:** `PATCH /v1/masters/hospitalization-plans/:id`

## 背景

バックエンドのハンドラ・DBには `body_size` と `billing_unit` が既に実装済み。
マスタ設定画面（`/settings/hospitalization`）の `HospitalizationSection` コンポーネントにUI入力フィールドが未実装。

## 追加するフィールド

| フィールド | UIコンポーネント | 値 | デフォルト |
|---|---|---|---|
| `body_size` | `Select` | small/medium/large | small |
| `billing_unit` | `Select` | per_day/per_night | per_day |

## 表示ラベル（仕様書準拠）

| フィールド | ラベル | 選択肢 |
|---|---|---|
| `body_size` | 対象体格 | 小型 / 中型 / 大型 |
| `billing_unit` | 料金単位 | 1日あたり / 1泊あたり |

## 実装場所

- `MasterItemFormSections` のディスパッチャーに `hospitalization` ケースを追加
- `HospitalizationSection` コンポーネントを実装（`SectionWrapper` / `NotionPropertyRow` 使用）

## 型定義

`models.ts` に `BodySize` / `BillingUnit` enumは既に定義済み。
