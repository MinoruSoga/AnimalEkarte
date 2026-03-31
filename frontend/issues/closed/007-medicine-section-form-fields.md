# 007: MedicineSection フォームフィールド追加

**ステータス:** open
**優先度:** medium
**関連API:** `PATCH /v1/masters/medicines/:id`

## 背景

バックエンドのハンドラに `dosage_form` と `medicine_unit` が追加された。
マスタ設定画面（`/settings/medicine`）の `MedicineSection` コンポーネントにUI入力フィールドが未実装。

## 追加するフィールド

| フィールド | UIコンポーネント | 値 | デフォルト |
|---|---|---|---|
| `dosage_form` | `Select` | tablet/liquid/injection/topical/powder | tablet |
| `medicine_unit` | `Select` | per_tablet/per_ml/per_dose/per_gram | per_tablet |

## 表示ラベル（仕様書準拠）

| フィールド | ラベル | 選択肢 |
|---|---|---|
| `dosage_form` | 剤形 | 錠剤 / 液剤 / 注射 / 外用薬 / 粉末 |
| `medicine_unit` | 単位 | 1錠あたり / 1mLあたり / 1回分 / 1gあたり |

## 実装場所

- `MasterItemFormSections` のディスパッチャーに `medicine` ケースを追加または更新
- `MedicineSection` コンポーネントを実装（`SectionWrapper` / `NotionPropertyRow` 使用）

## 型定義

`models.ts` に `DosageForm` / `MedicineUnit` enumは既に定義済み。

## 注意

`default_quantity` / `inventory_id` は将来フェーズ（在庫連携）で対応。このチケットのスコープ外。
