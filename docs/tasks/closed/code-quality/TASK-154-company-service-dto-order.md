# TASK-154: company_service.go — DTO/定数の定義順序違反

**優先度**: Low
**対象ファイル**: `backend/internal/service/company_service.go`
**チェック項目**: 6（Service の DTO・定数・helper の定義順序）

---

## 問題

プロジェクト規約では Service ファイルの定義順序を以下のように定めている。

```
CreateXxxInput / UpdateXxxInput  (DTO)
const colXxx* = "..."
func buildXxxUpdateFields(...)
type XxxService interface { ... }
type xxxService struct { ... }
```

`company_service.go` では `const (colCompany...)` が `UpdateCompanyInput` DTO より**前**に定義されており、規約と逆転している。

---

## 現状コード（company_service.go 抜粋）

```go
// 14行目: const が先
const (
    colCompanyName                      = "name"
    colCompanyPostalCode                = "postal_code"
    // ...
    colCompanyLogoURL                   = "logo_url"
)

// 29行目: DTO が後
type UpdateCompanyInput struct {
    Name                      *string
    // ...
}

// 44行目: interface
type CompanyService interface { ... }
```

---

## 修正後コード（ファイル全体の構造）

```go
// --- Input DTOs ---

// UpdateCompanyInput は法人情報部分更新の入力DTO
type UpdateCompanyInput struct {
    Name                      *string
    PostalCode                *string
    Address                   *string
    PhoneNumber               *string
    FaxNumber                 *string
    Email                     *string
    Website                   *string
    DirectorName              *string
    RegistrationNumber        *string
    InvoiceRegistrationNumber *string
    LogoURL                   *string
}

// --- DB column constants ---

const (
    colCompanyName                      = "name"
    colCompanyPostalCode                = "postal_code"
    colCompanyAddress                   = "address"
    colCompanyPhoneNumber               = "phone_number"
    colCompanyFaxNumber                 = "fax_number"
    colCompanyEmail                     = "email"
    colCompanyWebsite                   = "website"
    colCompanyDirectorName              = "director_name"
    colCompanyRegistrationNumber        = "registration_number"
    colCompanyInvoiceRegistrationNumber = "invoice_registration_number"
    colCompanyLogoURL                   = "logo_url"
)

func buildCompanyUpdateFields(input *UpdateCompanyInput) map[string]any { ... }

// CompanyService は法人情報のビジネスロジックインターフェース
type CompanyService interface { ... }

type companyService struct { ... }
// ...
```

---

## 修正手順

`company_service.go` で `const (colCompany...)` ブロックを `UpdateCompanyInput` 定義の直後（`CompanyService` インターフェース定義より前）に移動する。

