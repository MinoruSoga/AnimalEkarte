# BUG-387: マスタサービスの slog 構造化ログが不統一（ErrorContext 禁止・Reorder count 未記録）

## 概要
マスタ関連サービス層の slog ログに2つの不統一がある。(1) `medicine_service.go` が `slog.ErrorContext` を使用しているが、サービス層では `apperrors.Wrap` または `slog.WarnContext` を使うべき。(2) `trimming_master_service.go`（2箇所）・`merchandise_item_service.go`・`medicine_service.go` の Reorder メソッドが `slog.Int("count", len(ids))` を記録しておらず、他マスタと統一されていない。

## 再現手順
コードレビューで確認可能。

## 期待する動作
- サービス層で `slog.ErrorContext` を使用しないこと（`apperrors.Wrap` または `slog.WarnContext` を使用）
- 全マスタの Reorder メソッドで `slog.Int("count", len(ids))` を記録すること

## 現状コード

### 問題1: slog.ErrorContext の使用（medicine_service.go）

```go
// backend/internal/service/medicine_service.go:205-209
if err := s.inventoryRepo.Create(ctx, clinicID, invItem); err != nil {
    slog.ErrorContext(ctx, "failed to create inventory item",
        slog.Uint64("clinic_id", clinicID), slog.String("error", err.Error()))
    // best-effort: 在庫作成失敗は medicine 作成エラーにしない
}

// backend/internal/service/medicine_service.go:291-296
if err := s.inventoryRepo.DeleteByNameAndMedicineCategory(ctx, clinicID, *medicine.Name); err != nil {
    slog.ErrorContext(ctx, "failed to delete linked inventory (BUG-381)",
        slog.Uint64("clinic_id", clinicID), slog.String("error", err.Error()))
    // best-effort: 在庫削除失敗は medicine 削除エラーにしない
}
```

**問題**: `slog.ErrorContext` はシステムエラーレベルのログ。best-effort パターン（失敗しても処理継続）には不適切。`WarnContext` が正しい。

### 問題2: Reorder count 未記録（4箇所）

```go
// backend/internal/service/trimming_master_service.go:119（CourseReorder）
slog.InfoContext(ctx, "trimming courses reordered", slog.Uint64("clinic_id", clinicID))
// count が欠落

// backend/internal/service/trimming_master_service.go:275（OptionReorder）
slog.InfoContext(ctx, "trimming options reordered", slog.Uint64("clinic_id", clinicID))
// count が欠落

// backend/internal/service/merchandise_item_service.go:187
slog.InfoContext(ctx, "merchandise items reordered", slog.Uint64("clinic_id", clinicID))
// count が欠落

// backend/internal/service/medicine_service.go:251
slog.InfoContext(ctx, "medicines reordered", slog.Uint64("clinic_id", clinicID))
// count が欠落
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/cage_service.go:135
slog.InfoContext(ctx, "cage reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids))) // count を記録

// backend/internal/service/animal_species_service.go:128
slog.InfoContext(ctx, "animal_species reordered",
    slog.Int("count", len(ids)))
```

## 影響範囲

| 対象 | 問題 | 状態 |
|------|------|------|
| `backend/internal/service/medicine_service.go:205-209` | slog.ErrorContext → WarnContext に変更 | 要修正 |
| `backend/internal/service/medicine_service.go:291-296` | slog.ErrorContext → WarnContext に変更 | 要修正 |
| `backend/internal/service/trimming_master_service.go:119,275` | Reorder の count 追加 | 要修正 |
| `backend/internal/service/merchandise_item_service.go:187` | Reorder の count 追加 | 要修正 |
| `backend/internal/service/medicine_service.go:251` | Reorder の count 追加 | 要修正 |

## 修正方針

### 1. `medicine_service.go:205-209` — ErrorContext → WarnContext
```go
// 修正前
slog.ErrorContext(ctx, "failed to create inventory item", ...)

// 修正後
slog.WarnContext(ctx, "failed to create inventory item (best-effort)",
    slog.Uint64("clinic_id", clinicID), slog.String("error", err.Error()))
```

### 2. `medicine_service.go:291-296` — ErrorContext → WarnContext
```go
// 修正前
slog.ErrorContext(ctx, "failed to delete linked inventory (BUG-381)", ...)

// 修正後
slog.WarnContext(ctx, "failed to delete linked inventory (best-effort)",
    slog.Uint64("clinic_id", clinicID), slog.String("error", err.Error()))
```

### 3. 4箇所の Reorder ログに count を追加（例: trimming_master_service.go:119）
```go
// 修正前
slog.InfoContext(ctx, "trimming courses reordered", slog.Uint64("clinic_id", clinicID))

// 修正後
slog.InfoContext(ctx, "trimming courses reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — slog 使い方
> - `slog.InfoContext`: リソース作成/更新完了の情報
> - `slog.WarnContext`: best-effort パターン（失敗しても処理継続）
> - `slog.ErrorContext`: 禁止（サービス層。apperrors.Wrap で処理）

### `.claude/rules/go-language.md` — ログ（slog構造化ログ）
> 原則 service 層のみ。`InfoContext`, `ErrorContext` でコンテキストを適切に伝播させる。
> best-effort の在庫連携失敗は WarnContext が適切。

### プロジェクト内参照実装
`backend/internal/service/cage_service.go:135` — Reorder の count ログの正しい実装

## 優先度
**Low** — 機能への影響なし。ログの品質・監視のしやすさに影響する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/medicine_service.go:205-209,251,291-296` — 修正対象（ErrorContext + count）
- `backend/internal/service/trimming_master_service.go:119,275` — 修正対象（count）
- `backend/internal/service/merchandise_item_service.go:187` — 修正対象（count）
