# TASK-185: accounting_service.go の DTO・定義順序違反

## 優先度: Low

## 概要
`accounting_service.go` で `buildAccountingUpdateFields` ヘルパー関数が
`AccountingService` interface の前に定義されている。
規約では定義順序は:
```
CreateXxxInput → UpdateXxxInput → const → buildXxxUpdateFields → interface → struct → methods
```

## 対象ファイル
`backend/internal/service/accounting_service.go`

## 現状コード（定義順序）

```
L14:  CreateAccountingInput struct
L32:  UpdateAccountingInput struct
L60:  buildAccountingUpdateFields func    ← ここ
L101: AccountingService interface         ← interface が後
L115: accountingService struct
L119: NewAccountingService func
L123: methods...
```

## 問題の詳細
- `buildAccountingUpdateFields`（private helper）が `AccountingService`（public interface）の前にある
- 規約の順序に従えば `interface` → `struct` → `methods` の後ろに helper を置くか、
  または `const` の直後・`interface` の前に helper を置く（後者が規約に合致）
- 現状は `UpdateAccountingInput` の直後に helper があるため、
  interface より先に来てしまっている

## 修正後の定義順序

```go
// 1. CreateAccountingInput
type CreateAccountingInput struct { ... }

// 2. UpdateAccountingInput
type UpdateAccountingInput struct { ... }

// 3. (const があれば)

// 4. buildAccountingUpdateFields（helper 関数）
func buildAccountingUpdateFields(input *UpdateAccountingInput) map[string]any { ... }

// 5. AccountingService interface
type AccountingService interface { ... }

// 6. accountingService struct
type accountingService struct { ... }

// 7. NewAccountingService
func NewAccountingService(...) AccountingService { ... }

// 8. methods (List, GetByID, Create, Update, ...)

// 9. hasPaymentFields / buildPaymentFromInput 等の private helpers
func hasPaymentFields(...) bool { ... }
func buildPaymentFromInput(...) *model.Payment { ... }
```

現在の `buildAccountingUpdateFields` の位置は許容範囲（DTO の直後）ですが、
`hasPaymentFields` と `buildPaymentFromInput` はメソッド群の後（L214〜L265）にあり
こちらは規約に沿っています。
`buildAccountingUpdateFields` もメソッド群の後ろに移動して統一することを推奨します。
