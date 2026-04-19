# BUG-396: vaccine_service.Create が validateRequiredName を呼び出していない

## 概要
`vaccine_service.go` の `Create` メソッドが `validateRequiredName(input.Name)` を呼び出していない。他の全マスタサービス（14ファイル）は `validateRequiredName` または `validateOptionalName` を Create/Update の先頭で呼び出しているが、vaccine_service だけが欠落しており、空文字や空白のみの名前が保存される可能性がある。

## 再現手順
1. `POST /masters/vaccines` に `{"name": ""}` または `{"name": " "}` を送信
2. **結果**: 200/201 が返り、空文字の Name で vaccine レコードが作成される
3. **期待**: 400 Bad Request が返る

## 期待する動作
- 空文字・空白のみの Name は `apperrors.WrapInvalidInput` で拒否される
- 最小長 1・最大長 255 のバリデーションが適用される

## 現状コード

### `backend/internal/service/vaccine_service.go:56`（問題箇所）
```go
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
    if input.Species != nil {
        if err := validateVaccineSpecies(*input.Species); err != nil {
            return nil, err
        }
    }
    // validateRequiredName(input.Name) が欠落 ← ここに追加が必要
    vaccine := &model.Vaccine{
        ClinicID:    clinicID,
        Name:        input.Name,
        ...
    }
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/checkup_type_service.go:Create
func (s *checkupTypeService) Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
    if err := validateRequiredName(input.Name); err != nil {  // ← 先頭でバリデーション
        return nil, err
    }
    ...
}

// backend/internal/service/validators.go
func validateRequiredName(name string) error {
    if strings.TrimSpace(name) == "" {
        return apperrors.WrapInvalidInput("名前を入力してください")
    }
    if len([]rune(name)) > 255 {
        return apperrors.WrapInvalidInput("名前は255文字以内で入力してください")
    }
    return nil
}
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/service/vaccine_service.go:56` | Create メソッド先頭に `validateRequiredName(input.Name)` 追加 |
| `backend/internal/service/vaccine_service.go:Update` | Update メソッドで `validateOptionalName` が呼ばれているか確認（同様の漏れがないか） |
| `backend/internal/service/vaccine_service_test.go` | 空文字名のバリデーションテスト追加 |

## 修正方針

### `vaccine_service.go:Create` — validateRequiredName 追加
```go
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
    if err := validateRequiredName(input.Name); err != nil {  // ← 追加
        return nil, err
    }
    if input.Species != nil {
        if err := validateVaccineSpecies(*input.Species); err != nil {
            return nil, err
        }
    }
    ...
}
```

## 優先度
**Medium** — バリデーション欠落。空の Name が保存されると UI 表示・フィルタリングに問題が生じる。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/vaccine_service.go:56` — 問題箇所
- `backend/internal/service/validators.go` — validateRequiredName の定義場所
- `backend/internal/service/checkup_type_service.go` — 参照実装
