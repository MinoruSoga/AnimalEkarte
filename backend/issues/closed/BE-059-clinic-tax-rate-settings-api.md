# BE-059: Clinic 税率設定 API（GET/PATCH）

**Status**: Closed
**Priority**: High
**Affects**: clinic_handler, clinic_service, clinic_repository
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-058（先行必須）, FE-121

## Summary

`clinics` テーブルに追加した `standard_tax_rate`・`reduced_tax_rate` を取得・更新できる API を実装する。
既存の Clinic GET/PATCH エンドポイントのリクエスト/レスポンスに2フィールドを追加する形で対応する。

## 現状のコード

```go
// backend/internal/handler/clinic_handler.go（既存）
// GET /v1/clinics/:id  → ClinicResponse
// PATCH /v1/clinics/:id → UpdateClinicRequest → ClinicResponse

// backend/internal/handler/clinic_response.go（想定構造）
type ClinicResponse struct {
    ID             string `json:"id"`
    Name           string `json:"name"`
    // ... 既存フィールド
    // StandardTaxRate, ReducedTaxRate は未存在
}

// backend/internal/handler/clinic_request.go（想定構造）
type UpdateClinicRequest struct {
    Name *string `json:"name"`
    // ... 既存フィールド
    // StandardTaxRate, ReducedTaxRate は未存在
}
```

## 必要な変更

### 1. clinic_response.go の変更

```go
// 既存の ClinicResponse に追加
type ClinicResponse struct {
    // ... 既存フィールドはそのまま
    StandardTaxRate float64 `json:"standard_tax_rate"` // 例: 0.10
    ReducedTaxRate  float64 `json:"reduced_tax_rate"`  // 例: 0.08
}

// toClinicResponse() 関数に追加
func toClinicResponse(c *model.Clinic) ClinicResponse {
    return ClinicResponse{
        // ... 既存フィールド
        StandardTaxRate: c.StandardTaxRate,
        ReducedTaxRate:  c.ReducedTaxRate,
    }
}
```

### 2. clinic_request.go の変更

```go
// UpdateClinicRequest に追加（ポインタ型で PATCH セマンティクス）
type UpdateClinicRequest struct {
    // ... 既存フィールド
    StandardTaxRate *float64 `json:"standard_tax_rate"`
    ReducedTaxRate  *float64 `json:"reduced_tax_rate"`
}
```

### 3. clinic_service.go の変更

```go
// UpdateClinicInput に追加
type UpdateClinicInput struct {
    // ... 既存フィールド
    StandardTaxRate *float64
    ReducedTaxRate  *float64
}

// buildClinicUpdateFields() に追加
func buildClinicUpdateFields(input UpdateClinicInput) map[string]any {
    fields := make(map[string]any)
    // ... 既存フィールド
    if input.StandardTaxRate != nil {
        if *input.StandardTaxRate <= 0 || *input.StandardTaxRate > 1 {
            return nil, fmt.Errorf("standard_tax_rate must be between 0 and 1: %w", errors.ErrInvalidInput)
        }
        fields["standard_tax_rate"] = *input.StandardTaxRate
    }
    if input.ReducedTaxRate != nil {
        if *input.ReducedTaxRate <= 0 || *input.ReducedTaxRate > 1 {
            return nil, fmt.Errorf("reduced_tax_rate must be between 0 and 1: %w", errors.ErrInvalidInput)
        }
        fields["reduced_tax_rate"] = *input.ReducedTaxRate
    }
    return fields
}
```

## API レスポンス形式

```json
// GET /v1/clinics/:id
{
  "id": "1",
  "name": "動物病院サンプル",
  "standard_tax_rate": 0.10,
  "reduced_tax_rate": 0.08,
  // ... 既存フィールド
}

// PATCH /v1/clinics/:id
// Request
{
  "standard_tax_rate": 0.10,
  "reduced_tax_rate": 0.08
}
// Response: 更新後の ClinicResponse
```

## フロントエンド影響

- FE-121 で Clinic 型の `standard_tax_rate`, `reduced_tax_rate` を使用する
- `make codegen` 後（BE-058）に models.ts に既に含まれているため追加 codegen 不要

## 完了条件

- [ ] GET /v1/clinics/:id のレスポンスに `standard_tax_rate`, `reduced_tax_rate` が含まれる
- [ ] PATCH /v1/clinics/:id で両フィールドが更新できる
- [ ] 税率が 0〜1 の範囲外の場合 400 が返る
- [ ] `docker compose exec backend go test ./... -v` が通る
