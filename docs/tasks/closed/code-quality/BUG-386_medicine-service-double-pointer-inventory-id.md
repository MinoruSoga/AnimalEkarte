# BUG-386: medicine_service.UpdateMedicineInput.InventoryID が **uint64 二重ポインタ設計

## 概要
`medicine_service.go` の `UpdateMedicineInput` で `InventoryID` フィールドが `**uint64`（二重ポインタ）として定義されている。この設計は「nil = 未指定」「&nil = NULL クリア」「&&val = 値セット」の3状態を表現するが、Go の慣用句では非常に難解。同種の問題を解決している `exam_type_service.go` は `*uint64 + ClearParentID bool` フラグパターンを採用しており、こちらが推奨パターンだ。

## 再現手順
コードレビューで確認可能。

## 期待する動作
- `UpdateMedicineInput.InventoryID` を `*uint64` + `ClearInventoryID bool` フラグパターンに変更する
- `buildMedicineUpdateFields` 内でフラグを参照して適切に処理する

## 現状コード

### `backend/internal/service/medicine_service.go:59`（二重ポインタ）
```go
type UpdateMedicineInput struct {
    Name            *string
    IsActive        *bool
    Description     *string
    ParentID        *uint64
    ClearParentID   bool
    InventoryID     **uint64 // nil = 未指定, &nil = NULL クリア, &&val = 値セット
    // ...
}
```

### `backend/internal/service/medicine_service.go:101-103`（buildUpdateFields での処理）
```go
if input.InventoryID != nil {
    fields[colMedicineInventoryID] = *input.InventoryID // *uint64 (nil = NULL)
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/exam_type_service.go
type UpdateExamTypeInput struct {
    Name          *string
    IsActive      *bool
    ParentID      *uint64
    ClearParentID bool  // true = parent_id を NULL にクリア
    // ...
}

// buildExamTypeUpdateFields での処理
if input.ClearParentID {
    fields["parent_id"] = nil
} else if input.ParentID != nil {
    fields["parent_id"] = *input.ParentID
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/service/medicine_service.go:59` | UpdateMedicineInput.InventoryID 型定義 | 要修正 |
| `backend/internal/service/medicine_service.go:101-103` | buildMedicineUpdateFields の InventoryID 処理 | 要修正 |
| `backend/internal/handler/medicine_handler.go` | UpdateMedicine ハンドラの InventoryID 設定方法 | 確認・修正 |
| `backend/internal/service/medicine_service_test.go` | テストの修正 | 要修正 |

## 修正方針

### 1. `backend/internal/service/medicine_service.go:59` — UpdateMedicineInput を変更
```go
type UpdateMedicineInput struct {
    Name             *string
    IsActive         *bool
    Description      *string
    ParentID         *uint64
    ClearParentID    bool
    InventoryID      *uint64 // nil = 未指定（変更なし）
    ClearInventoryID bool    // true = inventory_id を NULL にクリア
    // ...
}
```

### 2. `backend/internal/service/medicine_service.go` — buildMedicineUpdateFields を修正
```go
// 修正前
if input.InventoryID != nil {
    fields[colMedicineInventoryID] = *input.InventoryID
}

// 修正後
if input.ClearInventoryID {
    fields[colMedicineInventoryID] = nil
} else if input.InventoryID != nil {
    fields[colMedicineInventoryID] = *input.InventoryID
}
```

### 3. `backend/internal/handler/medicine_handler.go` — ハンドラの InventoryID セット方法を変更
ハンドラ側でどのように `UpdateMedicineInput.InventoryID` を設定しているか確認し、`ClearInventoryID bool` フラグを活用するよう更新する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/go-language.md` — GORM PATCH（ポインタ型 + buildUpdateFields）
> UpdateXxxInput はポインタ型を使用し、NULL クリアには bool フラグ（`ClearXxxField bool`）を使用する。

### プロジェクト内参照実装
`backend/internal/service/exam_type_service.go` — `ClearParentID bool` フラグパターンの正しい実装

## 優先度
**Medium** — 機能は動作しているが、コードの可読性・保守性が低い。新規開発者が `**uint64` パターンを誤解してバグを混入させるリスクがある。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/medicine_service.go:59,101-103` — 問題箇所
- `backend/internal/service/exam_type_service.go` — 参照実装
- `backend/internal/handler/medicine_handler.go` — 影響確認が必要
- `backend/internal/service/medicine_service_test.go` — テスト更新が必要
