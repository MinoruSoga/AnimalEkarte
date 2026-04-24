# TASK-078: reservation_type_group — slog 順序違反 + service Update パターン不統一

## 優先度

LOW

---

## 概要

`reservation_type_group_service.go` に slog の clinic_id 順序違反と、
サービス Update メソッドのパターン不統一がある。

---

## 1. slog 順序違反（TASK-057 パターン）

### reservation_type_group_service.go — Delete

```go
// ❌ clinic_id が 2 番目
slog.InfoContext(ctx, "reservation_type_group deleted",
    slog.Uint64("reservation_type_group_id", id),  // ← 1番目
    slog.Uint64("clinic_id", clinicID))             // ← 2番目

// ✅ 修正後
slog.InfoContext(ctx, "reservation_type_group deleted",
    slog.Uint64("clinic_id", clinicID),                     // ← 1番目
    slog.Uint64("reservation_type_group_id", id))
```

---

## 2. Repository Update パターン不統一

### 現状

`reservation_type_group_repository.go` の Update は `error` のみを返す：

```go
// ❌ 現状: error のみ返す
Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
```

そのため `reservation_type_group_service.go` では Update 後に `FindByID` を呼ぶ（2クエリ）：

```go
if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
    return nil, apperrors.Wrap(err, "...")
}
g, err := s.repo.FindByID(ctx, clinicID, id)  // ← 追加クエリ
```

### 参照実装（medicine_repository.go）

```go
// ✅ UpdateFields が更新後エンティティを返す（1クエリ）
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
```

### 修正方針

`reservation_type_group_repository.go` の `Update` を `UpdateFields` に改名し、
更新後エンティティを返すよう変更する。

```go
// ✅ 修正後
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationTypeGroup, error)
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `reservation_type_group_service.go` | Delete の slog 順序修正（clinic_id を先頭に） |
| `reservation_type_group_repository.go` | `Update` → `UpdateFields`、返り値を `(*model.ReservationTypeGroup, error)` に変更 |
| `reservation_type_group_service.go` | Update 後の FindByID 呼び出しを削除、UpdateFields の返り値を使用 |
