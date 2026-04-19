# BUG-402: reservation_type_service.LinkOccupation が occupationID の存在チェックを行わない

## 概要
`reservation_type_service.go` の `LinkOccupation` メソッドは `reservationTypeID` の存在確認（`s.repo.FindByID`）を行うが、`occupationID` の存在確認を行わない。存在しない occupationID が渡された場合、DB の FK 制約エラーが発生して apperrors でラップされていない生の GORM エラーが返るか、FK 制約がなければ不正なレコードが作成される。

## 再現手順
1. 存在しない occupationID を使って `POST /masters/reservation-types/:id/occupations` に `{"occupation_id": 99999}` を送信
2. **結果（FK 制約あり）**: DB エラーが 500 Internal Server Error として返る（4xx ではない）
3. **期待**: 422/400 で「職種が見つかりません」が返る

## 現状コード

### `backend/internal/service/reservation_type_service.go:382-408`
```go
func (s *reservationTypeService) LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
    // reservationTypeID の存在確認あり ✅
    if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
        return nil, apperrors.Wrap(err, "failed to get reservation type")
    }
    // occupationID の存在確認なし ❌
    o := &model.ReservationTypeOccupation{
        ClinicID:          clinicID,
        ReservationTypeID: reservationTypeID,
        OccupationID:      occupationID,
    }
    if err := s.occupationRepo.Create(ctx, o); err != nil {
        return nil, apperrors.Wrap(err, "failed to link occupation")
        // FK 制約エラー → wrapped されるが ErrNotFound ではなく内部エラーとして扱われる
    }
    ...
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/diagnosis_service.go (DiagnosisName.Create での typeID チェック)
if _, err := s.typeRepo.FindByID(ctx, clinicID, input.DiagnosisTypeID); err != nil {
    return nil, apperrors.Wrap(err, "diagnosis type not found")
}
// ↑ 親リソースの存在を確認してから子を作成
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/service/reservation_type_service.go:382-408` | `occupationID` の存在確認（`FindByID`）を `Create` 前に追加 |
| エラーレスポンス | FK 制約エラーが 500 ではなく 404/422 で返るようになる |

## 修正方針

### `reservation_type_service.go:LinkOccupation` — occupationID 存在確認追加
```go
func (s *reservationTypeService) LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
        return nil, apperrors.Wrap(err, "failed to get reservation type")
    }
    // occupationID の存在確認を追加
    if _, err := s.occupationRepo.FindByID(ctx, clinicID, occupationID); err != nil {
        return nil, apperrors.Wrap(err, "occupation not found")
        // ErrNotFound が上位で 404 にマッピングされる
    }
    o := &model.ReservationTypeOccupation{...}
    ...
}
```

**前提**: `reservationTypeService` が `OccupationRepository`（または `OccupationService`）への参照を持っているかを確認。現在 `occupationRepo` を保持しているなら追加不要。保持していない場合は DI 変更も必要。

## 優先度
**Medium** — 存在しない occupationID の紐付けが DB エラー（500）になるか不正レコードになる。FK 制約がある場合は 500 エラーとして露出し、UX が悪い。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/reservation_type_service.go:382-408` — 問題箇所
- `backend/internal/service/diagnosis_service.go` — 参照実装（親リソースの存在確認パターン）
