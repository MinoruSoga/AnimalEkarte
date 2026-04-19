# TASK-005: ReservationTypeGroup Delete に slog ログを追加

## 概要

`reservation_type_group_service.go` の `Delete` 処理に `slog.InfoContext` が記録されていない。同ファイルの `Create` や他のマスタサービス（`exam_type_service.go:80`, `checkup_type_service.go:82` 等）はすべて Delete 成功後にログを記録しており、実装が不統一になっている。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/service/reservation_type_group_service.go` | L115-127（Delete 成功後のログなし） |

## 規約違反

`.claude/rules/go-language.md`:
> 構造化ログ `slog` はサービス層のみで使用し、重要なミューテーション操作には InfoContext を記録する。

## 修正方針

```go
// 現状
func (s *reservationTypeGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
    // FK 依存チェック
    count, err := s.repo.CountUsageByGroupID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to count usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("このグループには予約種別が紐付いているため削除できません")
    }
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete reservation type group")
    }
    return nil  // ← ログなし
}

// 修正後
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete reservation type group")
    }
    slog.InfoContext(ctx, "reservation type group deleted",
        slog.Uint64("id", id),
        slog.Uint64("clinic_id", clinicID))
    return nil
```

## テスト

- Delete 成功時に slog が呼ばれることを検証（slogtest または zap-observer 相当で確認）
