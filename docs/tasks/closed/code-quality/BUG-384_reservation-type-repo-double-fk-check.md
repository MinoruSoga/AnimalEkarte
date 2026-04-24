# BUG-384: reservation_type_repository.Delete に FK 依存チェックが二重実装

## 概要
`reservation_type_service.Delete` はサービス層で `ExistsByReservationTypeID` による事前チェックを行っているが、`reservation_type_repository.Delete` 内でも `isFKConstraintErr` によるFK制約違反のハンドリングを二重に実装している。他の全マスタ（animal_species, cage, chief_complaint 等）はサービス層のみで依存チェックを行うパターンに統一されており、reservation_type だけが例外となっている。

## 再現手順
コードレビューで確認可能。

1. サービス層で `ExistsByReservationTypeID` が `true` を返す → `WrapConflict` で 409 返却
2. （仮に競合状態で）リポジトリ層の Delete まで到達した場合 → `isFKConstraintErr` で別メッセージの `WrapConflict` を返却
3. **問題**: 同一エラーに対して2つの異なるエラーメッセージが存在し、コードパスが不明確

## 期待する動作
- 依存チェックはサービス層のみで実施する（単一責任）
- リポジトリ層は FK 制約違反を `apperrors.FromGORM` で変換し、サービス層に委譲する

## 現状コード

### `backend/internal/service/reservation_type_service.go:283-296`（サービス層チェック）
```go
func (s *reservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    exists, err := s.reservationRepo.ExistsByReservationTypeID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check reservation dependency")
    }
    if exists {
        return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
    }
    // ...削除実行
}
```

### `backend/internal/repository/reservation_type_repository.go:77-90`（リポジトリ層の二重チェック）
```go
func (r *reservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
    if result.Error != nil {
        // BUG-030: ON DELETE RESTRICT の FK 制約違反は 409 Conflict に変換する
        if isFKConstraintErr(result.Error) {
            return apperrors.WrapConflict("このサービス種別は予約に使用されているため削除できません")
            // ↑ サービス層と異なるエラーメッセージ
        }
        return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
    }
    // ...
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/repository/animal_species_repository.go — シンプルな Delete
func (r *animalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
    result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AnimalSpecies{})
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "animal_species", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.WrapNotFound("animal_species", fmt.Sprintf("%d", id))
    }
    return nil
}
// サービス層が CountByAnimalSpeciesID で事前チェック済み → リポジトリは FK チェックしない
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/reservation_type_repository.go:81-83` | isFKConstraintErr による二重チェック | 要削除 |
| `backend/internal/service/reservation_type_service.go:283-296` | 既存の事前チェック（正しい） | 変更不要 |

## 修正方針

### `backend/internal/repository/reservation_type_repository.go:81-83` — isFKConstraintErr ブロックを削除

```go
// 修正前
func (r *reservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
    if result.Error != nil {
        if isFKConstraintErr(result.Error) {
            return apperrors.WrapConflict("このサービス種別は予約に使用されているため削除できません")
        }
        return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
    }
    // ...
}

// 修正後
func (r *reservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
    }
    return nil
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — マスタ削除の FK 依存チェック
> マスタ削除時は必ず依存レコードの存在をサービス層でチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。
> リポジトリ層は FK 制約違反のハンドリングを行わない（単一責任）。

### プロジェクト内参照実装
`backend/internal/repository/animal_species_repository.go` — リポジトリは `FromGORM` のみで処理

## 優先度
**Medium** — 機能上の問題はないが、二重実装により同一エラーに対して異なるメッセージが存在し、デバッグと保守を困難にする。サービス層の事前チェックが競合状態（race condition）で抜け落ちた場合にリポジトリ側のメッセージが露出する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/reservation_type_repository.go:77-90` — 問題箇所（二重チェック）
- `backend/internal/service/reservation_type_service.go:283-296` — 正しい事前チェック（変更不要）
