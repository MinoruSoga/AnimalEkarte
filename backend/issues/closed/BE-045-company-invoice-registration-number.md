# BE-045: company テーブルに invoice_registration_number カラム追加

**Status**: Closed
**Priority**: Medium
**Affects**: 法人情報 API（GET/PATCH /v1/company）
**Date Created**: 2026-03-18
**Related**: TASK-020, FE-073

## Summary

`company` テーブルにインボイス番号（適格請求書発行事業者登録番号）カラムを追加し、GET/PATCH API で読み書きできるようにする。

## 現状のコード

### DB スキーマ

```sql
-- backend/migrations/001_init.sql:93-107
CREATE TABLE company (
    id                  BIGSERIAL   PRIMARY KEY,
    name                text        NOT NULL,
    postal_code         text        NOT NULL DEFAULT '',
    address             text        NOT NULL DEFAULT '',
    phone_number        text        NOT NULL DEFAULT '',
    fax_number          text        NOT NULL DEFAULT '',
    registration_number text        NOT NULL DEFAULT '',
    director_name       text        NOT NULL DEFAULT '',
    email               text        NOT NULL DEFAULT '',
    website             text        NOT NULL DEFAULT '',
    logo_url            text        NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
```

### Go モデル

```go
// backend/internal/model/company.go:5-20
type Company struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"  json:"id"`
	Name               string    `gorm:"not null"                  json:"name"`
	PostalCode         string    `gorm:"default:''"                json:"postal_code"`
	Address            string    `gorm:"default:''"                json:"address"`
	PhoneNumber        string    `gorm:"default:''"                json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                json:"fax_number"`
	Email              string    `gorm:"default:''"                json:"email"`
	Website            string    `gorm:"default:''"                json:"website"`
	DirectorName       string    `gorm:"default:''"                json:"director_name"`
	RegistrationNumber string    `gorm:"default:''"                json:"registration_number"`
	LogoURL            string    `gorm:"default:''"                json:"logo_url"`
	CreatedAt          time.Time `gorm:"autoCreateTime"            json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"            json:"updated_at"`
}
```

### Handler request/response

```go
// backend/internal/handler/company_request.go:3-14
type updateCompanyRequest struct {
	Name               *string `json:"name"`
	PostalCode         *string `json:"postal_code"`
	Address            *string `json:"address"`
	PhoneNumber        *string `json:"phone_number"`
	FaxNumber          *string `json:"fax_number"`
	Email              *string `json:"email"`
	Website            *string `json:"website"`
	DirectorName       *string `json:"director_name"`
	RegistrationNumber *string `json:"registration_number"`
	LogoURL            *string `json:"logo_url"`
}

// backend/internal/handler/company_response.go:10-24
type companyResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	PostalCode         string    `json:"postal_code"`
	Address            string    `json:"address"`
	PhoneNumber        string    `json:"phone_number"`
	FaxNumber          string    `json:"fax_number"`
	Email              string    `json:"email"`
	Website            string    `json:"website"`
	DirectorName       string    `json:"director_name"`
	RegistrationNumber string    `json:"registration_number"`
	LogoURL            string    `json:"logo_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
```

### Service

```go
// backend/internal/service/company_service.go:13-24
type UpdateCompanyInput struct {
	Name               *string
	PostalCode         *string
	Address            *string
	PhoneNumber        *string
	FaxNumber          *string
	Email              *string
	Website            *string
	DirectorName       *string
	RegistrationNumber *string
	LogoURL            *string
}

// backend/internal/service/company_service.go:59-92
func buildCompanyUpdateFields(input *UpdateCompanyInput) map[string]any {
	fields := map[string]any{}
	// ... 既存フィールド ...
	if input.RegistrationNumber != nil {
		fields["registration_number"] = *input.RegistrationNumber
	}
	if input.LogoURL != nil {
		fields["logo_url"] = *input.LogoURL
	}
	return fields
}
```

## 必要な変更

### 1. DB マイグレーション

```sql
-- backend/migrations/001_init.sql の company テーブル定義
-- registration_number と director_name の間に挿入:
    registration_number text        NOT NULL DEFAULT '',
    invoice_registration_number text NOT NULL DEFAULT '',  -- ← 新規追加
    director_name       text        NOT NULL DEFAULT '',
```

### 2. Model 変更

```go
// backend/internal/model/company.go
// RegistrationNumber と LogoURL の間に挿入:
	RegistrationNumber        string    `gorm:"default:''"                json:"registration_number"`
	InvoiceRegistrationNumber string    `gorm:"default:''"                json:"invoice_registration_number"` // ← 新規追加
	LogoURL                   string    `gorm:"default:''"                json:"logo_url"`
```

### 3. Handler Request 変更

```go
// backend/internal/handler/company_request.go
// RegistrationNumber と LogoURL の間に挿入:
	RegistrationNumber        *string `json:"registration_number"`
	InvoiceRegistrationNumber *string `json:"invoice_registration_number"` // ← 新規追加
	LogoURL                   *string `json:"logo_url"`
```

### 4. Handler Response 変更

```go
// backend/internal/handler/company_response.go
// companyResponse — RegistrationNumber と LogoURL の間に挿入:
	RegistrationNumber        string    `json:"registration_number"`
	InvoiceRegistrationNumber string    `json:"invoice_registration_number"` // ← 新規追加
	LogoURL                   string    `json:"logo_url"`

// toCompanyResponse() に追加:
	InvoiceRegistrationNumber: c.InvoiceRegistrationNumber,
```

### 5. Handler 変更

```go
// backend/internal/handler/company_handler.go:28-39
// UpdateCompany() 内の service.UpdateCompanyInput{} — RegistrationNumber と LogoURL の間に挿入:
		RegistrationNumber:        req.RegistrationNumber,
		InvoiceRegistrationNumber: req.InvoiceRegistrationNumber, // ← 新規追加
		LogoURL:                   req.LogoURL,
```

### 6. Service 変更

```go
// backend/internal/service/company_service.go
// UpdateCompanyInput — RegistrationNumber と LogoURL の間に挿入:
	RegistrationNumber        *string
	InvoiceRegistrationNumber *string // ← 新規追加
	LogoURL                   *string

// buildCompanyUpdateFields() — RegistrationNumber ブロックと LogoURL ブロックの間に挿入:
	if input.InvoiceRegistrationNumber != nil {
		fields["invoice_registration_number"] = *input.InvoiceRegistrationNumber
	}
```

### 7. Repository 層

変更不要。`repository.Update(ctx, fields)` は `map[string]any` を受け取るため、service 層で `buildCompanyUpdateFields()` に追加するだけで自動的に対応される。

### 8. codegen 実行

```bash
make codegen
```

## API レスポンス形式

```json
// GET /v1/company
{
  "id": "1",
  "name": "株式会社テスト",
  "registration_number": "1234567890123",
  "invoice_registration_number": "T1234567890123",
  ...
}

// PATCH /v1/company
// Request:
{ "invoice_registration_number": "T1234567890123" }
// Response: 上記と同形式
```

## フロントエンド影響

- `make codegen` で `models.ts` の `Company` 型に `invoice_registration_number` フィールドが追加される
- FE-073 で UI 対応が必要

## 完了条件

- [ ] `001_init.sql` に `invoice_registration_number` カラム追加
- [ ] `model/company.go` にフィールド追加
- [ ] `company_request.go`, `company_response.go` にフィールド追加
- [ ] `company_handler.go` の `UpdateCompany()` で input に渡す
- [ ] `company_service.go` の `UpdateCompanyInput` + `buildCompanyUpdateFields()` にフィールド追加
- [ ] `make codegen` で `models.ts` 更新
- [ ] GET /v1/company のレスポンスに `invoice_registration_number` が含まれる
- [ ] PATCH /v1/company で `invoice_registration_number` を更新できる

## クローズ情報

- **Closed At**: 2026-03-18
- **変更ファイル**:
  - `backend/migrations/001_init.sql` — company テーブルに `invoice_registration_number` カラム追加
  - `backend/internal/model/company.go` — `InvoiceRegistrationNumber` フィールド追加
  - `backend/internal/handler/company_request.go` — `InvoiceRegistrationNumber` フィールド追加
  - `backend/internal/handler/company_response.go` — `InvoiceRegistrationNumber` フィールド追加 + `toCompanyResponse()` マッピング追加
  - `backend/internal/handler/company_handler.go` — `UpdateCompany()` で input に渡す
  - `backend/internal/service/company_service.go` — `UpdateCompanyInput` + `buildCompanyUpdateFields()` にフィールド追加
  - `frontend/src/types/generated/models.ts` — `make codegen` で自動更新
