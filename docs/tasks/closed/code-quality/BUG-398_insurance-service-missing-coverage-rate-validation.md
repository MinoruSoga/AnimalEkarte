# BUG-398: insurance_service.Create が CoverageRate のバリデーションを行っていない

## 概要
`insurance_service.go` の `Create` および `Update` メソッドで `CoverageRate`（保険補償率）の範囲バリデーション（0〜100）が実装されていない。`validators.go` には同様のロジック（`validateDiscountRate`）が存在するが再利用されていない。無効な補償率（負数・100超）が DB に保存されると、請求額計算のバグに直結する。

## 再現手順
1. `POST /masters/insurances` に `{"name": "テスト保険", "coverage_rate": -10}` を送信
2. **結果**: 201 Created — バリデーションなしで -10 が保存される
3. `POST /masters/insurances` に `{"name": "テスト保険", "coverage_rate": 150}` を送信
4. **結果**: 201 Created — 150% という不正な補償率が保存される

## 期待する動作
- `CoverageRate < 0` または `CoverageRate > 100` の場合は 400 Bad Request を返す

## 現状コード

### `backend/internal/service/insurance_service.go:56`（問題箇所）
```go
func (s *insuranceService) Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    insurance := &model.Insurance{
        ClinicID:     clinicID,
        Name:         input.Name,
        IsActive:     input.IsActive,
        Description:  input.Description,
        CoverageRate: input.CoverageRate,  // ← 0〜100 のバリデーションなし
        SortOrder:    input.SortOrder,
    }
    ...
}
```

### `backend/internal/service/validators.go`（活用すべき既存関数）
```go
// 同様の範囲バリデーション（0〜100 を検証）が存在するが insurance では使われていない
func validateDiscountRate(rate *float64) error {
    if rate == nil {
        return nil
    }
    if *rate < 0 || *rate > 100 {
        return apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
    }
    return nil
}
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/service/insurance_service.go:Create` | CoverageRate のバリデーション追加 |
| `backend/internal/service/insurance_service.go:Update` | buildInsuranceUpdateFields 前に CoverageRate の範囲チェック追加 |
| `backend/internal/service/validators.go` | `validateCoverageRate` 関数を新規追加（または `validateDiscountRate` を汎用化） |
| `backend/internal/service/insurance_service_test.go` | 範囲外バリデーションのテスト追加 |

## 修正方針

### 1. `validators.go` — validateCoverageRate 追加
```go
func validateCoverageRate(rate int) error {
    if rate < 0 || rate > 100 {
        return apperrors.WrapInvalidInput("補償率は0〜100の範囲で入力してください")
    }
    return nil
}
```

### 2. `insurance_service.go:Create` — バリデーション追加
```go
func (s *insuranceService) Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    if err := validateCoverageRate(input.CoverageRate); err != nil {  // ← 追加
        return nil, err
    }
    ...
}
```

### 3. `insurance_service.go:Update` — buildInsuranceUpdateFields 前にも追加
```go
if input.CoverageRate != nil {
    if err := validateCoverageRate(*input.CoverageRate); err != nil {
        return nil, err
    }
}
```

## 優先度
**Medium** — 不正な補償率が DB に保存されると、保険請求額の計算（`請求額 × (100 - coverage_rate) / 100`）がバグになる。会計の正確性に直結する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/insurance_service.go` — 問題箇所
- `backend/internal/service/validators.go` — validateCoverageRate の追加場所
- `backend/internal/service/insurance_service_test.go` — テスト追加が必要
