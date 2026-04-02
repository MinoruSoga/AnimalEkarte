# 法人情報（Company）API 実装

## 概要
運営法人情報の取得・更新 API を実装する。システム内にシングルトンとして存在し、一覧エンドポイントは不要。
`model/company.go` の `Company` struct は実装済み。handler・service・repository が未実装。

## 優先度
low

## 関連テーブル
- `company` (id, name NOT NULL, postal_code DEFAULT '', address DEFAULT '', phone_number DEFAULT '', fax_number DEFAULT '', email DEFAULT '', website DEFAULT '', director_name DEFAULT '', registration_number DEFAULT '', logo_url DEFAULT '', created_at, updated_at)
  - シングルトン: レコードは常に 1 件のみ（id=1 固定、または `LIMIT 1` で取得）

## 実装内容

### モデル
`model/company.go` は実装済み。変更不要。

```go
type Company struct {
    ID                 uint64
    Name               string
    PostalCode         string
    Address            string
    PhoneNumber        string
    FaxNumber          string
    Email              string
    Website            string
    DirectorName       string
    RegistrationNumber string
    LogoURL            string
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
```

### リポジトリ
新規ファイル `repository/company_repository.go`:
```go
type CompanyRepository interface {
    Get(ctx context.Context) (*model.Company, error)
    Update(ctx context.Context, fields map[string]any) error
}
```
- `Get`: `SELECT * FROM company LIMIT 1`。レコードが存在しない場合は `apperrors.WrapNotFound("company", "1")` を返す。
- `Update`: `UPDATE company SET ... WHERE id = (SELECT id FROM company LIMIT 1)` または最初の 1 件を対象に更新。`RowsAffected == 0` の場合は `WrapNotFound` を返す。

`repository/repositories.go` に `Company CompanyRepository` を追加。

### サービス
新規ファイル `service/company_service.go`:
```go
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

type CompanyService interface {
    Get(ctx context.Context) (*model.Company, error)
    Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error)
}
```

`buildCompanyUpdateFields(input *UpdateCompanyInput) map[string]any` を実装する。
`Update` で `len(fields) == 0` の場合は `apperrors.WrapInvalidInput("at least one field must be provided")` を返す。

### ハンドラ
新規ファイル `handler/company_handler.go`:
```go
func (h *Handler) GetCompany(c *gin.Context)
func (h *Handler) UpdateCompany(c *gin.Context)
func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/company_request.go`:
```go
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
```

新規ファイル `handler/company_response.go`:
```go
type companyResponse struct {
    ID                 uint64    `json:"id"`
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

func toCompanyResponse(c *model.Company) companyResponse { ... }
```

### ルート登録
`cmd/api/main.go` に追加（認証必要）:
```go
v1.GET("/company",   authMiddleware, h.GetCompany)
v1.PATCH("/company", authMiddleware, h.UpdateCompany)
```

## 完了条件
- `GET /v1/company` が法人情報を返す（DB にレコードがない場合は 404）
- `PATCH /v1/company` で法人情報を部分更新できる
- フィールドが 1 つも送信されない場合は 400 エラーを返す
- 未認証リクエストは 401 を返す
