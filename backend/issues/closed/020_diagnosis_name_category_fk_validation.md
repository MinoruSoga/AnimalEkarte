---
status: open
---

# DiagnosisName: diagnosis_category_id の存在確認（外部キー整合性チェック）

## 背景

`POST /v1/masters/diagnosis-names` および `PATCH /v1/masters/diagnosis-names/:id` の実装において、
`diagnosis_category_id` に対する外部キー整合性チェックが存在しない。

pet service では `ownerRepo.FindByID()` で同一 clinic_id のオーナー存在を事前確認している。

## 問題

```go
// 現在の diagnosis_service.go（確認なし）
func (s *diagnosisNameService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
    name := &model.DiagnosisName{
        ClinicID:            clinicID,
        DiagnosisCategoryID: input.DiagnosisCategoryID,  // ← 存在確認なし
        // ...
    }
    // ...
}
```

存在しない、または別 clinic の `diagnosis_category_id` を指定しても DB レベルでエラーになるが、
アプリ層で適切なエラーメッセージを返せない（500ではなく400を返すべき）。

## 修正方針

```go
// service 内で事前チェック
func (s *diagnosisNameService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
    // カテゴリの存在確認（clinic_id スコープ）
    if _, err := s.categoryRepo.FindByID(ctx, clinicID, input.DiagnosisCategoryID); err != nil {
        return nil, apperrors.WrapInvalidInput("diagnosis category not found in this clinic")
    }
    // ...
}
```

同様に `Update` でも `DiagnosisCategoryID` 変更時に確認する。

## 完了条件

- [ ] `diagnosisNameService` に `categoryRepo` フィールドを追加
- [ ] `Create` で `diagnosis_category_id` の存在確認（clinic_id スコープ）
- [ ] `Update` で `DiagnosisCategoryID` が変更される場合も確認
- [ ] 存在しない場合は `400 Bad Request` + 日本語エラーメッセージ
- [ ] `service.go` の DI 配線を更新
