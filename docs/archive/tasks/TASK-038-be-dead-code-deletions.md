# TASK-038: BE 確定デッドコード削除

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 高
**領域**: Backend

---

## 概要

静的解析で確定した Go バックエンドのデッドコード（空ファイル・重複実装・未使用ラッパー）を削除する。

---

## 対応項目

### 1. `handler/helper.go` — 完全削除

```go
// 内容
package handler
```

`package` 宣言のみで本体が完全に空。いかなる関数・型も定義していない。

**対応**: ファイルごと削除。

---

### 2. `ClinicService.GetCompany / UpdateCompany` を削除

**背景**:
- `ClinicService` インターフェースに `GetCompany` / `UpdateCompany` が定義されている
- `company_handler.go` は `h.svc.Company`（`CompanyService`）を使用しており、`h.svc.Clinic.GetCompany/UpdateCompany` は production コードで **一切呼ばれない**
- `CompanyService` と完全重複

**削除対象**:

```go
// service/clinic_service.go
// interface から削除（L86-87）
GetCompany(ctx context.Context) (*model.Company, error)
UpdateCompany(ctx context.Context, company *model.Company) error

// 実装を削除（L139-144）
func (s *clinicService) GetCompany(ctx context.Context) (*model.Company, error) { ... }
func (s *clinicService) UpdateCompany(ctx context.Context, company *model.Company) error { ... }

// clinic_service.go 内で repo.GetCompany を呼んでいる箇所（L108）も削除
```

```go
// repository/clinic_repository.go
// interface から削除（L17-18）
GetCompany(ctx context.Context) (*model.Company, error)
UpdateCompany(ctx context.Context, company *model.Company) error

// 実装を削除（L51-80）
func (r *clinicRepository) GetCompany(...) { ... }
func (r *clinicRepository) UpdateCompany(...) { ... }
```

---

### 3. `errors/errors.go` の汎用 `Is` / `As` ラッパー削除

```go
// L83-90 削除対象
func Is(err, target error) bool {
    return errors.Is(err, target)
}

func As(err error, target any) bool {
    return errors.As(err, target)
}
```

`apperrors.Is/As` は production コードで未使用。
`apperrors.IsNotFound` / `apperrors.IsInvalidInput` 等は alive なので残す。

---

## 受入条件

- [ ] `handler/helper.go` が削除されている
- [ ] `ClinicService` インターフェースから `GetCompany` / `UpdateCompany` が削除されている
- [ ] `clinicService` の実装から `GetCompany` / `UpdateCompany` が削除されている
- [ ] `ClinicRepository` インターフェースから `GetCompany` / `UpdateCompany` が削除されている
- [ ] `clinicRepository` の実装から `GetCompany` / `UpdateCompany` が削除されている
- [ ] `errors.go` から汎用 `Is` / `As` 関数が削除されている
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./...` 全テストパス
