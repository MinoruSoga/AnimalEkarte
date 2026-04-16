# BUG-348: reservation_type_group_service の読み取り系メソッドが apperrors.Wrap を欠落

## 概要

`reservation_type_group_service.go` の `List`、`GetByID`、`Reorder`、`Delete` の末尾 `repo.Delete` 呼び出しで、
リポジトリエラーを `apperrors.Wrap` でラッピングせずに直接返している。
プロジェクト規約「Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング」に違反。

## 違反箇所

```go
// backend/internal/service/reservation_type_group_service.go

// L43-44: List — Wrap なし
func (s *reservationTypeGroupService) List(...) ([]model.ReservationTypeGroup, error) {
    return s.repo.FindAll(ctx, clinicID) // ← apperrors.Wrap 欠落
}

// L47-48: GetByID — Wrap なし
func (s *reservationTypeGroupService) GetByID(...) (*model.ReservationTypeGroup, error) {
    return s.repo.FindByID(ctx, clinicID, id) // ← apperrors.Wrap 欠落
}

// L101: Delete の最終 repo.Delete — Wrap なし
func (s *reservationTypeGroupService) Delete(...) error {
    // ...
    return s.repo.Delete(ctx, clinicID, id) // ← apperrors.Wrap 欠落
}

// L104-105: Reorder — Wrap なし
func (s *reservationTypeGroupService) Reorder(...) error {
    return s.repo.Reorder(ctx, clinicID, ids) // ← apperrors.Wrap 欠落
}
```

## 修正方針

```go
func (s *reservationTypeGroupService) List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
    groups, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list reservation type groups")
    }
    return groups, nil
}

func (s *reservationTypeGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
    group, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get reservation type group")
    }
    return group, nil
}

// Delete の末尾
if err := s.repo.Delete(ctx, clinicID, id); err != nil {
    return apperrors.Wrap(err, "failed to delete reservation type group")
}
return nil

// Reorder
func (s *reservationTypeGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder reservation type groups")
    }
    return nil
}
```

## 影響範囲

- 機能的影響は限定的（リポジトリが `apperrors.FromGORM` で変換済みのエラーを返すため、ハンドラの `RespondError` は正しく動作する）
- ただしエラーメッセージのコンテキスト（どのサービスで失敗したか）が失われるため、デバッグが困難になる

## 優先度

**MEDIUM** — 機能バグではなく規約違反。デバッグ品質への影響。

## 関連ファイル

- `backend/internal/service/reservation_type_group_service.go:43-48,101,104-105`
