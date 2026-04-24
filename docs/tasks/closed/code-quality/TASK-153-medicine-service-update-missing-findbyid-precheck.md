# TASK-153: medicine_service.go — Update メソッドに FindByID precheck がない

**優先度**: Medium
**対象ファイル**: `backend/internal/service/medicine_service.go`
**チェック項目**: 1（Handler の責務）+ 規約一貫性

---

## 問題

プロジェクト標準実装（`reservation_type_service.go` など）では、`Update` メソッド冒頭で `FindByID` を呼び出して対象レコードの存在を確認してから更新処理に進む。

`medicine_service.go` の `Update` メソッドはこの precheck を行わず、直接 `repo.UpdateFields` を呼び出している。

`repo.UpdateFields` 内で `RowsAffected == 0` を `WrapNotFound` に変換するため機能的には 404 が返るが、以下の問題がある。

1. **一貫性の欠如**: 同プロジェクト内の他サービス（`vaccine_service`、`reservation_type_service`、`animal_species_service` など）はすべて Update 冒頭で FindByID を呼ぶパターンを採用している。
2. **エラーコンテキストの明確さ**: precheck がない場合、バリデーションエラーより先に存在チェックが走らないため、「存在しないIDに対するバリデーションエラー」と「存在するIDへの正当な更新」が同一フローを通る。
3. **将来の拡張リスク**: Update 前に取得したオブジェクトを使ってビジネスルール判定を追加したい場合にリファクタリングが必要になる。

---

## 現状コード（medicine_service.go 222〜240行）

```go
func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildMedicineUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }

    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update medicine")
    }
    // ...
    return result, nil
}
```

---

## 修正後コード

```go
func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
    // 存在確認（参照実装と一貫性を保つ）
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get medicine")
    }
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildMedicineUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }

    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update medicine")
    }
    slog.InfoContext(ctx, "medicine updated",
        slog.Uint64("clinic_id", clinicID),
        slog.Uint64("medicine_id", id),
    )
    return result, nil
}
```

---

## 修正手順

`medicine_service.go` の `Update` メソッド冒頭に `s.repo.FindByID(ctx, clinicID, id)` の存在確認を追加する。

