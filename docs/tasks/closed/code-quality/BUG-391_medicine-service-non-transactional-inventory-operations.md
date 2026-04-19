# BUG-391: medicine_service の在庫連携操作がトランザクション外で実行される

## 概要
`medicine_service.go` の `Create` と `Delete` は、薬品レコードの操作後に在庫アイテムを別操作（非トランザクション）で連携する。`Create` は薬品作成が成功した後に在庫作成を試みるが、在庫作成が失敗しても薬品は作成済みのまま残る（薬品が在庫と紐づかない孤児状態）。コードに `best-effort` コメントがあるが、`slog.ErrorContext` を使用しており、意図（警告レベル）と実装（エラーレベル）が矛盾している。

## 再現手順
1. 在庫テーブルへの書き込みを強制的に失敗させた状態で `POST /masters/medicines` を送信
2. **結果**: 薬品レコードは作成されるが、対応する在庫アイテムが存在しない状態になる

## 期待する動作
- best-effort パターンとして継続するなら、`slog.WarnContext` を使用し、エラーコメントも統一する
- 強整合性が必要なら、薬品作成と在庫作成を単一トランザクションでまとめる

## 現状コード

### 問題1: `backend/internal/service/medicine_service.go:190-210` — Create（非トランザクション）
```go
if err := s.repo.Create(ctx, medicine); err != nil {
    return nil, apperrors.Wrap(err, "failed to create medicine")
}

// BUG-320: 薬品作成時に在庫アイテムを自動作成
inventoryItem := &model.InventoryItem{...}
if err := s.inventoryRepo.Create(ctx, clinicID, inventoryItem); err != nil {
    slog.ErrorContext(ctx, "failed to create inventory item",   // ← ErrorContext は不適切
        slog.Uint64("medicine_id", medicine.ID),
        slog.String("error", err.Error()))
    // best-effort: 薬品は作成済みなので、エラーは警告レベル
    // ← コメントは「警告レベル」と言っているが実装は ErrorContext
}
```

### 問題2: `backend/internal/service/medicine_service.go:285-297` — Delete（非トランザクション）
```go
// 薬品削除を先に実行
if err := s.repo.Delete(ctx, clinicID, id); err != nil {
    return apperrors.Wrap(err, "failed to delete medicine")
}
// 在庫削除は best-effort（薬品削除後に失敗すると孤児在庫が残る）
if err := s.inventoryRepo.DeleteByNameAndMedicineCategory(ctx, clinicID, m.Name); err != nil {
    slog.ErrorContext(ctx, "failed to delete linked inventory (BUG-381)",  // ← ErrorContext は不適切
        slog.Uint64("clinic_id", clinicID),
        ...
    )
}
```

### 問題3: BUG-387 との重複
`slog.ErrorContext` → `slog.WarnContext` の修正は BUG-387 でも指摘されている。本チケットはトランザクション設計の問題に焦点を当てる。

## 影響範囲

| 対象 | 問題 |
|------|------|
| `backend/internal/service/medicine_service.go:204-209` | `slog.ErrorContext` → `slog.WarnContext`（BUG-387 と重複、こちらで対処） |
| `backend/internal/service/medicine_service.go:291-296` | `slog.ErrorContext` → `slog.WarnContext`（BUG-387 と重複、こちらで対処） |
| 設計上の問題 | Create・Delete が非トランザクション設計であることの明示的なコメント欠如 |

## 修正方針

### Option A: best-effort を維持（推奨 — 変更コスト低）
コメントと slog レベルを統一する。BUG-387 の修正で解決する。

```go
if err := s.inventoryRepo.Create(ctx, clinicID, inventoryItem); err != nil {
    slog.WarnContext(ctx, "failed to create inventory item (best-effort)",
        slog.Uint64("clinic_id", clinicID),
        slog.Uint64("medicine_id", medicine.ID),
        slog.String("error", err.Error()))
    // best-effort: 在庫作成失敗は medicine 作成エラーにしない
    // 孤児は在庫一覧 UI から手動修復可能
}
```

### Option B: 完全整合性（変更コスト高）
`medicineRepository.Create` をトランザクションで包み、在庫作成も同一トランザクションに含める。
```go
return s.db.Transaction(func(tx *gorm.DB) error {
    if err := txRepo.Create(ctx, medicine); err != nil {
        return apperrors.Wrap(err, "failed to create medicine")
    }
    if err := txInventoryRepo.Create(ctx, clinicID, inventoryItem); err != nil {
        return apperrors.Wrap(err, "failed to create inventory item")
    }
    return nil
})
```

**推奨**: まず Option A（BUG-387 対応）を実施し、トランザクション化は別チケットで計画する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — slog 使い方
> `slog.WarnContext`: best-effort パターン（失敗しても処理継続）
> `slog.ErrorContext`: 禁止（サービス層。apperrors.Wrap で処理）

### `backend/CLAUDE.md` — トランザクション
> GORM `Transaction` で原子性を保証する。

## 優先度
**Medium** — best-effort パターン自体は許容可能な設計だが、ログレベルの誤りとトランザクション設計の明文化が必要。BUG-387 と合わせて対応することで即座に改善できる。

## 関連チケット
- **BUG-387**: slog.ErrorContext → WarnContext の修正（ログレベル部分は BUG-387 で対処）

## 関連ファイル
- `backend/internal/service/medicine_service.go:190-210` — Create の非トランザクション設計
- `backend/internal/service/medicine_service.go:285-297` — Delete の非トランザクション設計
