# TASK-132: payment_method_master_service — DTO 定義順序違反・apperrors.Wrap 漏れ

## 優先度
**Medium**

## 対象ファイル
`backend/internal/service/payment_method_master_service.go`

---

## 問題 1: DTO・builder の定義順序がインターフェースより後になっている

### チェック項目
- **Service の DTO・定数・helper の定義順序**: `Input DTO → const → builder → interface → struct → メソッド` の順が規約。

### 現状コード（service.go L12–78）
```go
// ❌ インターフェースが先に定義されている
type PaymentMethodMasterService interface {  // L12
    ...
}

// ❌ DTO がインターフェースより後
type CreatePaymentMethodInput struct {  // L21
    ...
}

type UpdatePaymentMethodInput struct {  // L27
    ...
}

type paymentMethodMasterService struct { ... }  // L33

// ❌ builder 関数がメソッド実装の後
func buildPaymentMethodUpdateFields(...) map[string]any {  // L66
    ...
}
```

### 修正後コード
参照実装 (`reservation_type_service.go`) に合わせて下記順序に並び替える。

```go
// ---- Input DTOs ----
type CreatePaymentMethodInput struct { ... }
type UpdatePaymentMethodInput struct { ... }

// ---- builder ----
func buildPaymentMethodUpdateFields(input UpdatePaymentMethodInput) map[string]any { ... }

// ---- PaymentMethodMasterService ----
type PaymentMethodMasterService interface { ... }

type paymentMethodMasterService struct { ... }

func NewPaymentMethodMasterService(...) PaymentMethodMasterService { ... }

// ---- メソッド実装 ----
func (s *paymentMethodMasterService) List(...) ...
...
```

---

## 問題 2: List / Create / Update / GetByID で `apperrors.Wrap` を省略している

### チェック項目
- **Service のエラーハンドリング**: Service 内の repo 呼び出しエラーは `apperrors.Wrap(err, "context message")` でラップする必要がある（単純な `return s.repo.Xxx(...)` は文脈情報なしで上位に伝播する）。

### 現状コード
```go
// ❌ エラーをそのまま返している（Wrap なし）
func (s *paymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
    return s.repo.FindAll(ctx, clinicID)  // L43
}

func (s *paymentMethodMasterService) Create(...) (*model.PaymentMethodMaster, error) {
    ...
    return s.repo.Create(ctx, m)  // L55 (Wrap なし)
}

func (s *paymentMethodMasterService) Update(...) (*model.PaymentMethodMaster, error) {
    fields := buildPaymentMethodUpdateFields(input)
    if len(fields) == 0 {
        return s.repo.FindByID(ctx, clinicID, id)  // L61 (Wrap なし)
    }
    return s.repo.UpdateFields(ctx, clinicID, id, fields)  // L63 (Wrap なし)
}

func (s *paymentMethodMasterService) Delete(...) error {
    ...
    return s.repo.Delete(ctx, clinicID, id)  // L91 (Wrap なし)
}
```

### 修正後コード
```go
func (s *paymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
    items, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list payment methods")
    }
    return items, nil
}

func (s *paymentMethodMasterService) Create(ctx context.Context, clinicID uint64, input CreatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
    slog.InfoContext(ctx, "creating payment method",
        slog.Uint64("clinic_id", clinicID),
        slog.String("name", input.Name))
    m := &model.PaymentMethodMaster{
        ClinicID:     clinicID,
        Name:         input.Name,
        DisplayOrder: input.DisplayOrder,
    }
    result, err := s.repo.Create(ctx, m)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to create payment method")
    }
    slog.InfoContext(ctx, "payment method created", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", result.ID))
    return result, nil
}

func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
    fields := buildPaymentMethodUpdateFields(input)
    if len(fields) == 0 {
        result, err := s.repo.FindByID(ctx, clinicID, id)
        if err != nil {
            return nil, apperrors.Wrap(err, "failed to get payment method")
        }
        return result, nil
    }
    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update payment method")
    }
    return result, nil
}

func (s *paymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
    ...
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete payment method")
    }
    return nil
}
```

---

## 備考
- `Delete` での依存チェック (`CountUsageByID`) と `apperrors.WrapConflict` は正しく実装されている。
- `List` の `apperrors.Wrap` 漏れは軽微だが、エラー追跡で文脈が失われるため修正が望ましい。
