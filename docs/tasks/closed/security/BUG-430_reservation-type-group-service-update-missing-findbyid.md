# BUG-430: reservation_type_group_service の Update で FindByID 存在確認が欠落

## 概要

BUG-424 で指摘した「複数サービスの Update メソッドで FindByID 存在確認が欠落」と同種の問題が
`reservation_type_group_service.go` にも存在する。BUG-424 起票後に発見されたため別チケットとする。

## 問題箇所

```go
// reservation_type_group_service.go:96-107
func (s *reservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildReservationTypeGroupUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    g, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← FindByID 存在確認なし
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update reservation type group")
    }
    // ...
}
```

## 標準パターン（cage_service.go）

```go
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    // ✅ Step 1: 存在確認（テナント検証含む）
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // Step 2: フィールド更新 ...
}
```

## 修正方針

Update メソッドの先頭に存在確認を追加する。

```go
func (s *reservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
    // ← 追加: 存在確認・テナント検証
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get reservation type group")
    }

    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ... 以降は変更なし
}
```

## 影響ファイル

- `backend/internal/service/reservation_type_group_service.go` — 行 96-107（Update メソッド）

## 優先度

**High** — マルチテナント境界の保護が不完全。他クリニックの予約種別グループを上書きできる可能性。

## 関連チケット

- BUG-424（reservation_type / trimming_master×2 / diagnosis_name の同種問題）
- BUG-420（trimming_service.Update の同種問題）
