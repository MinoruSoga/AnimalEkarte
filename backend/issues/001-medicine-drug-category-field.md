# [Issue #001] medicines テーブルに薬効分類（drug_category）フィールドが欠損

## 概要

薬剤マスタ画面（`/settings/medicine`）のUIデザインでは、薬剤を **薬効分類**（例: 抗生剤・消炎剤・利尿剤）でグループ化した階層テーブルを表示する。
しかし現在の `medicines` テーブルには薬効分類フィールドが存在しない。

## 現在の状態

| フィールド | 型 | 意味 |
|-----------|-----|------|
| `dosage_form` | `string` | **剤形**（物理的形状）: `tablet` / `liquid` / `injection` / `topical` / `powder` |
| ~~`drug_category`~~ | — | **薬効分類**（治療分類）: 存在しない ❌ |

現在フロントエンドは `dosage_form`（剤形）を暫定的なグループキーとして使用しているが、
UIデザインの意図は「抗生剤」「消炎剤」「利尿剤」といった **治療上の分類** によるグループ化である。

## 影響

- 薬剤マスタ画面のグループヘッダーが「錠剤」「液剤」等の剤形表示になる（デザインと相違）
- 「抗生剤」「消炎剤」等の治療分類による絞り込み・並び替えが不可能

## 必要な対応

### 1. マイグレーション

```sql
ALTER TABLE medicines
  ADD COLUMN drug_category VARCHAR(100) NULL;

COMMENT ON COLUMN medicines.drug_category IS '薬効分類（例: 抗生剤, 消炎剤, 利尿剤）';
```

### 2. Goモデル更新（`backend/internal/model/medicine.go`）

```go
DrugCategory *string `gorm:"column:drug_category" json:"drug_category,omitempty"`
```

### 3. ハンドラ・サービス更新

- `CreateMedicineInput` / `UpdateMedicineInput` に `DrugCategory *string` を追加
- `*_request.go` でバインド追加
- `*_response.go` でレスポンスに含める

### 4. コードジェン再実行

```bash
make codegen
```

### 5. フロントエンド更新

- `Medicine` 型に `drugCategory?: string` を追加（`transformBackendPetToFrontend` パターンで）
- `MedicineSettings.tsx` のグループキーを `dosageForm` → `drugCategory` に変更
- `CreateMedicineRequest` / `UpdateMedicineRequest` に `drug_category` を追加

## 優先度

**Medium** — UI機能的には現在の `dosage_form` グループ化で動作するが、デザイン意図と乖離がある。
次回スプリントで対応推奨。

## 参考

- 参考UIデザイン: `https://sweep-tart-35834055.figma.site/settings/medicine`
- 現在のフロントエンド暫定実装: `frontend/src/features/master/routes/MedicineSettings.tsx`
