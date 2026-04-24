# CODE-QUALITY-209: Service Update メソッドの `input == nil` チェック欠落（9サービス）

## 概要

複数の Service `Update` メソッドで `input *UpdateXxxInput` が `nil` の場合のチェックが欠落している。
`input.Name` などのフィールドアクセスが即 nil pointer dereference パニックを起こす。
Handler からは nil が渡らないが、テストや将来の呼び出し元からの呼び出し時に防御できない。

## 優先度

HIGH

## 影響ファイル（欠落が確認されたもの）

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `backend/internal/service/animal_species_service.go` | Update | ~L104 |
| `backend/internal/service/cage_service.go` | Update | ~L133 |
| `backend/internal/service/reservation_type_service.go` | Update | ~L266 |
| `backend/internal/service/permission_group_service.go` | Update | ~L101 |
| `backend/internal/service/diagnosis_service.go` | diagnosisTypeService.Update | ~L124 |
| `backend/internal/service/diagnosis_service.go` | diagnosisNameService.Update | ~L273 |
| `backend/internal/service/medicine_service.go` | Update | ~L222 |
| `backend/internal/service/merchandise_item_service.go` | Update | ~L152 |
| `backend/internal/service/trimming_service.go` | TrimmingCourse.Update / TrimmingOption.Update | ~L174 |

---

## 正しいパターン（参照実装: exam_type_service.go）

```go
func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    ...
}
```

`payment_method_master_service.go` および `consultation_service.go` は実装済み。

---

## 修正方針

全9サービスの `Update` メソッド先頭に以下を追加する。

```go
if input == nil {
    return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
}
```

`ErrMsgInputNotNil` は `validators.go` または `errors.go` に定義されている共通定数を使用すること。

---

## 規約参照

- `.claude/rules/go-language.md`: エラーハンドリング

## テスト

- 各サービスの Update に nil を渡した場合に `400 Bad Request` が返ることを確認
- 既存テストが引き続き通ることを確認
