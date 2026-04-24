# TASK-007: Go モデルの ENUM 定数欠落 — staff_type.trimmer / billing_status.pending / payment_method.electronic_money

## 概要

migration SQL で定義された ENUM 値が Go モデルの定数として定義されておらず、DB に該当値が格納されていても Go 側でハンドリングできない型安全性の破綻が発生している。

## 優先度

CRITICAL（型安全性破綻・runtime 時に未定義値扱いになるリスク）

## 影響ファイル

| ファイル | 欠落している定数 |
|---------|---------------|
| `backend/internal/model/staff.go:16` | `StaffTypeTrimmer StaffType = "trimmer"` |
| `backend/internal/model/accounting.go:20` | `BillingStatusPending BillingStatus = "pending"` |
| `backend/internal/model/accounting.go:30` | `PaymentMethodElectronicMoney PaymentMethod = "electronic_money"` |

## 規約違反

`.claude/rules/go-language.md`:
> GORM モデルのフィールドが migration の DDL と一致すること。ENUM 定数はすべての値を網羅すること。

## SQL定義（001_init.sql）

```sql
CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'trimmer', 'resource');
CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money');
```

## 修正方針

### staff.go

```go
const (
    StaffTypeDoctor   StaffType = "doctor"
    StaffTypeNurse    StaffType = "nurse"
    StaffTypeTrimmer  StaffType = "trimmer"   // 追加
    StaffTypeResource StaffType = "resource"
)
```

### accounting.go

```go
// BillingStatus
const (
    BillingStatusWaiting   BillingStatus = "waiting"
    BillingStatusCompleted BillingStatus = "completed"
    BillingStatusCancelled BillingStatus = "cancelled"
    BillingStatusPending   BillingStatus = "pending"   // 追加
)

// PaymentMethod
const (
    PaymentMethodCash             PaymentMethod = "cash"
    PaymentMethodCreditCard       PaymentMethod = "credit_card"
    PaymentMethodElectronicMoney  PaymentMethod = "electronic_money"  // 追加
)
```

## 注意

`make codegen` を実行して `frontend/src/types/generated/models.ts` も自動更新すること。フロントエンド側でも欠落した ENUM 値に対する UI 対応が必要か確認すること。

## テスト

- 各 ENUM 定数が `switch` 文で網羅されているかの静的解析（`exhaustive` lint ルール検討）
- `trimmer` タイプのスタッフ登録・取得の動作確認
