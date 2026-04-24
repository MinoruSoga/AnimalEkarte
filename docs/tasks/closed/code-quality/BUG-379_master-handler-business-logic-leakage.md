# BUG-379: マスタハンドラにビジネスロジックが漏出（insurance, procedure）

## 概要
`insurance_handler.go` と `procedure_handler.go` の Create ハンドラにビジネスロジック（デフォルト値設定・型変換）が混入している。`CoverageRate` のデフォルト値設定と `TaxType` の文字列型変換はサービス層の責務であり、ハンドラはリクエストパースと委譲のみを担うべきだ。

## 再現手順
コードレビューで確認可能。実行時に問題が顕在化するシナリオ:
1. `CreateInsurance` ハンドラ: `req.CoverageRate` が `nil` の場合に `coverageRate = 0` を設定（ハンドラがデフォルト値を決める）
2. `CreateProcedure` ハンドラ: `input.TaxType != ""` の場合に `*string` 型へ変換（ハンドラが型変換を行う）

## 期待する動作
- ハンドラはリクエストボディをパースし、Input DTO にそのまま詰めてサービスに委譲する
- デフォルト値の設定・型変換はサービス層が担う

## 現状コード

### `backend/internal/handler/insurance_handler.go:61-69`（デフォルト値設定がハンドラに）
```go
coverageRate := 0
if req.CoverageRate != nil {
    coverageRate = *req.CoverageRate
}
svcInput := &service.CreateInsuranceInput{
    Name:         req.Name,
    IsActive:     req.IsActive,
    Description:  req.Description,
    CoverageRate: coverageRate, // ← ハンドラがデフォルト値を決定
    ContactPhone: req.ContactPhone,
    SortOrder:    req.SortOrder,
}
```

### `backend/internal/handler/procedure_handler.go:61-77`（型変換がハンドラに）
```go
var taxType *string
if input.TaxType != "" {
    t := input.TaxType
    taxType = &t // ← ハンドラが string → *string 変換
}
svcInput := &service.CreateProcedureInput{
    Name:        input.Name,
    Price:       input.Price,
    IsActive:    input.IsActive,
    Description: input.Description,
    Duration:    input.Duration,
    Anesthesia:  input.Anesthesia,
    ParentID:    input.ParentID,
    SortOrder:   input.SortOrder,
    TaxType:     taxType,
    TaxRate:     input.TaxRate,
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/handler/cage_handler.go — ハンドラは直接 Input に詰める
svcInput := &service.CreateCageInput{
    Name:      req.Name,
    CageType:  req.CageType,
    CageSize:  req.CageSize,
    SortOrder: req.SortOrder,
}
cage, err := h.svc.Cage.Create(c.Request.Context(), clinicID, svcInput)
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/insurance_handler.go:61-69` | CoverageRate デフォルト値設定 | 要修正 |
| `backend/internal/handler/procedure_handler.go:61-65` | TaxType 型変換 | 要修正 |
| `backend/internal/service/insurance_service.go` | CreateInsuranceInput・Create() | 変更が必要 |
| `backend/internal/service/procedure_service.go` | CreateProcedureInput・Create() | 変更が必要 |

## 修正方針

### 1. `backend/internal/service/insurance_service.go` — Input DTO をポインタ型に変更
```go
type CreateInsuranceInput struct {
    Name         string  `json:"name"`
    IsActive     bool    `json:"is_active"`
    Description  *string `json:"description"`
    CoverageRate *int    `json:"coverage_rate"` // nil 許容に変更
    ContactPhone *string `json:"contact_phone"`
    SortOrder    int     `json:"sort_order"`
}

func (s *insuranceService) Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
    coverageRate := 0 // サービス層でデフォルト値を決定
    if input.CoverageRate != nil {
        coverageRate = *input.CoverageRate
    }
    // ...
}
```

### 2. `backend/internal/handler/insurance_handler.go:61-72` — ハンドラを簡素化
```go
svcInput := &service.CreateInsuranceInput{
    Name:         req.Name,
    IsActive:     req.IsActive,
    Description:  req.Description,
    CoverageRate: req.CoverageRate, // nil をそのまま渡す
    ContactPhone: req.ContactPhone,
    SortOrder:    req.SortOrder,
}
```

### 3. `backend/internal/service/procedure_service.go` — TaxType 変換をサービス層に移動
```go
type CreateProcedureInput struct {
    Name        string  `json:"name"`
    TaxType     string  `json:"tax_type"` // string のまま
    // ...
}

func (s *procedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
    var taxType *string
    if input.TaxType != "" {
        t := input.TaxType
        taxType = &t // サービス層で変換
    }
    // ...
}
```

### 4. `backend/internal/handler/procedure_handler.go:61-77` — ハンドラを簡素化
```go
svcInput := &service.CreateProcedureInput{
    Name:        input.Name,
    Price:       input.Price,
    IsActive:    input.IsActive,
    Description: input.Description,
    Duration:    input.Duration,
    Anesthesia:  input.Anesthesia,
    ParentID:    input.ParentID,
    SortOrder:   input.SortOrder,
    TaxType:     input.TaxType, // そのまま渡す
    TaxRate:     input.TaxRate,
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — アーキテクチャ
> handler → service → repository の軽量レイヤードを徹底。ビジネスロジックはサービス層に集約する。

### `.claude/rules/go-language.md` — インターフェース設計
> handler 層は HTTP 関心事のみ担う。ビジネスロジックはサービス層に配置する。

### プロジェクト内参照実装
`backend/internal/handler/cage_handler.go` — ハンドラが Input に直接詰めるシンプルパターン

## 優先度
**Medium** — ビジネスロジックの責任分離違反。テスト不能なコードがハンドラに混入しており、仕様変更時に変更箇所が分散するリスクがある。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/insurance_handler.go:61-69` — 問題箇所（ハンドラ内デフォルト値）
- `backend/internal/handler/procedure_handler.go:61-65` — 問題箇所（ハンドラ内型変換）
- `backend/internal/service/insurance_service.go` — 修正先（デフォルト値移動先）
- `backend/internal/service/procedure_service.go` — 修正先（型変換移動先）
