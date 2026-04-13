# BUG-259: マスタ削除時の FK 依存チェック欠如

## 概要

マスタエンティティの Delete 処理で、依存レコードの存在チェックが欠落している。
FK 制約によるDBエラーが発生するか、データ整合性が崩れる可能性がある。

## 影響範囲

| ファイル | メソッド | 問題 |
|---------|---------|------|
| `service/reservation_staff_service.go` | Delete | 予約（`reservation_appointments`）で使用中のスタッフを削除可能 |
| `service/inquiry_template_service.go` | Delete | FK 依存チェックなし。将来 inquiry から参照される場合にデータ不整合 |

## 修正方針

### reservation_staff_service.go

```go
func (s *reservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.GetByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to find staff")
    }
    // FK 依存チェック追加
    exists, err := s.reservationRepo.ExistsByStaffID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check reservation dependency")
    }
    if exists {
        return apperrors.WrapConflict("このスタッフは予約データで使用中のため削除できません")
    }
    if err := s.repo.SoftDelete(ctx, id); err != nil {
        return apperrors.Wrap(err, "failed to delete staff")
    }
    slog.InfoContext(ctx, "reservation staff deleted", slog.Uint64("staff_id", id))
    return nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — マスタ削除の FK 依存チェック (MANDATORY)
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。

## 優先度

**High** — マスタデータ削除による参照整合性の破壊。

## 関連チケット

- BUG-253: 親チケット

## 補足

`animal_species_repository.go` の Delete 内に FK チェックが存在するが、
同時に `animal_species_service.go` でも同じチェックを実施しており重複している。
Service 層に一元化し、Repository 層の重複チェックは削除すべき。
