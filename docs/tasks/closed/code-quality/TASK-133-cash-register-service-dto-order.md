# TASK-133: cash_register_service — DTO 定義順序違反・List/GetByID の apperrors.Wrap 漏れ

## 優先度
**Medium**

## 対象ファイル
`backend/internal/service/cash_register_service.go`

---

## 問題 1: インターフェースが DTO より先に定義されている

### チェック項目
- **Service の DTO・定数・helper の定義順序**: `Input DTO → const → builder → interface → struct → メソッド` の順が規約。

### 現状コード（service.go L15–98）
```go
// ❌ インターフェースが最初
type CashRegisterService interface {  // L15
    GetPreview(...)
    Close(...)
    List(...)
    GetByID(...)
}

// ❌ レスポンス DTO・入力 DTO がインターフェースより後
type CashRegisterPreview struct { ... }         // L23
type CloseAggregateSummary struct { ... }       // L35
type CloseBillingDetail struct { ... }          // L43
type CloseRegisterInput struct { ... }          // L58
type periodAggregate struct { ... }             // L67 (internal helper)

type cashRegisterService struct { ... }         // L78
```

### 修正後コード
参照実装に合わせて並び替え。

```go
// ---- Input DTOs ----
type CloseRegisterInput struct { ... }

// ---- Response / Internal types ----
type CashRegisterPreview struct { ... }
type CloseAggregateSummary struct { ... }
type CloseBillingDetail struct { ... }
type periodAggregate struct { ... }

// ---- CashRegisterService ----
type CashRegisterService interface { ... }

type cashRegisterService struct { ... }

func NewCashRegisterService(...) CashRegisterService { ... }

// ---- メソッド実装 ----
...
```

---

## 問題 2: `List` / `GetByID` で `apperrors.Wrap` を省略している

### チェック項目
- **Service のエラーハンドリング**: repo 呼び出しエラーは `apperrors.Wrap` でラップする。

### 現状コード
```go
// ❌ Wrap なし
func (s *cashRegisterService) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
    return s.closeRepo.FindAll(ctx, clinicID, startDate, endDate, page, limit)  // L271
}

func (s *cashRegisterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
    return s.closeRepo.FindByID(ctx, clinicID, id)  // L275
}
```

### 修正後コード
```go
func (s *cashRegisterService) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
    items, total, err := s.closeRepo.FindAll(ctx, clinicID, startDate, endDate, page, limit)
    if err != nil {
        return nil, 0, apperrors.Wrap(err, "failed to list cash register closes")
    }
    return items, total, nil
}

func (s *cashRegisterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
    record, err := s.closeRepo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get cash register close")
    }
    return record, nil
}
```

---

## 備考
- `Close` メソッドの `validatePeriod`・二重締めチェック・`apperrors.WrapConflict` は正しく実装されている。
- `GetPreview` の `validatePeriod` 呼び出しも Service 内で正しく行われている。
- TASK-131 で Handler から日付パースが Service に移動されると、`GetPreview` の引数シグネチャ変更が必要になるため、TASK-131 と合わせて対応すること。
