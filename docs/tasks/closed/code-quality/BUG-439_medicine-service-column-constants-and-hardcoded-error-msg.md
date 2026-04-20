# BUG-439: medicine_service のカラム定数先頭配置・エラーメッセージハードコード

## 概要

`medicine_service.go` に 2 つの軽微な規約違反がある。いずれも既存の
BUG-423/BUG-428 と同種の問題が medicine_service に追加で存在する。

---

## 問題 1: カラム定数がファイル先頭に定義（BUG-423/BUG-437 追加分）

```go
// medicine_service.go:14-29（ファイル先頭）
// --- DB column constants ---

const (
    colMedicineName            = "name"
    colMedicineParentID        = "parent_id"
    colMedicinePrice           = "price"
    colMedicineIsActive        = "is_active"
    colMedicineDescription     = "description"
    colMedicineDosageForm      = "dosage_form"
    colMedicineMedicineUnit    = "medicine_unit"
    colMedicineInventoryID     = "inventory_id"
    colMedicineDefaultQuantity = "default_quantity"
    colMedicineSortOrder       = "sort_order"
    colMedicineTaxType         = "tax_type"
    colMedicineTaxRate         = "tax_rate"
)

// --- Input DTOs ---
type CreateMedicineInput struct { ... }
```

**規約**: 列名定数は `buildXxxUpdateFields()` 関数の直前（ファイル末尾付近）にまとめる。

**修正**: const ブロックを `buildMedicineUpdateFields()` 関数の直前に移動。

---

## 問題 2: ErrMsgAtLeastOneField をハードコード（BUG-428 追加分）

```go
// medicine_service.go:228（Update メソッド内）
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")  // ← ハードコード
}
```

**規約**: `validators.go` に定義された `ErrMsgAtLeastOneField` 定数を使用すること。

**修正**:
```go
return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
```

---

## 補足: BUG-429 修正済みの確認

medicine_service の Create/Delete は既に `WithTx` でラップ済み（行 192, 283）。
BUG-429 は実装完了済み。

## 影響ファイル

- `backend/internal/service/medicine_service.go` — 行 14-29（定数配置）、行 228（ハードコード）

## 優先度

**Low** — 機能影響なし。BUG-423/BUG-428 の修正時に medicine_service も合わせて対応する。

## 関連チケット

- BUG-423（consultation/diagnosis のカラム定数先頭配置）
- BUG-437（animal_species/merchandise_item の同種問題）
- BUG-428（エラーメッセージ定数未使用・英語混在）
- BUG-429（medicine WithTx — 修正済み）
