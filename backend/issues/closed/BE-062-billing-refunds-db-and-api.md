# BE-062: billing_refunds テーブル + 返金 API 実装

**Status**: Closed
**Priority**: High
**Affects**: accounting（会計精算）
**Date Created**: 2026-03-26
**Related**: TASK-031, FE-126

## Summary

`billing_refunds` テーブルを新設し、`completed` 会計に対する部分返金・全額返金・複数回返金を管理する API を実装する。
`billings.status` は変更しない（Stripe モデル）。

## 現状のコード

```go
// backend/internal/model/accounting.go（抜粋）
type Billing struct {
    ID              uint64         `gorm:"primaryKey" tygo:"type: number"`
    ClinicID        uint64
    PetID           uint64
    OwnerID         uint64
    MedicalRecordID *uint64
    Status          BillingStatus  `gorm:"type:billing_status"`
    SubTotal        int64
    TaxTotal        int64
    TotalAmount     int64
    ScheduledDate   string
    CompletedAt     *string
    Memo            string
    Items           []BillingItem  `gorm:"foreignKey:BillingID"`
    Payments        []Payment      `gorm:"foreignKey:BillingID"`
    // Refunds は未定義
}
```

```sql
-- backend/migrations/001_init.sql（現状）
-- billing_refunds テーブルなし
-- billings テーブル: id, clinic_id, pet_id, owner_id, medical_record_id,
--   status, subtotal, tax_total, total_amount, scheduled_date, completed_at, memo
```

```go
// backend/internal/handler/accounting_handler.go（現状のルート）
// GET    /v1/accountings
// POST   /v1/accountings
// GET    /v1/accountings/:id
// PATCH  /v1/accountings/:id
// DELETE /v1/accountings/:id
// 返金エンドポイントなし
```

## 必要な変更

### 1. DB マイグレーション

```sql
-- backend/migrations/001_init.sql に追記

-- 返金レコード（Stripe モデル：billing_refunds で独立管理）
CREATE TABLE billing_refunds (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL,
    billing_id   BIGINT NOT NULL,
    amount       BIGINT NOT NULL CHECK (amount > 0),
    reason       TEXT,
    refunded_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_billing_refunds_billing FOREIGN KEY (billing_id) REFERENCES billings(id),
    CONSTRAINT fk_billing_refunds_clinic  FOREIGN KEY (clinic_id)  REFERENCES clinics(id)
);
CREATE INDEX idx_billing_refunds_billing ON billing_refunds(billing_id);
CREATE INDEX idx_billing_refunds_clinic_billing ON billing_refunds(clinic_id, billing_id);
```

### 2. Model 変更

```go
// backend/internal/model/billing_refund.go（新規）
package model

import "time"

type BillingRefund struct {
    ID         uint64     `gorm:"primaryKey"                tygo:"type: number"`
    ClinicID   uint64
    BillingID  uint64
    Amount     int64      // 返金額（正の整数、円）
    Reason     string     // 返金理由（任意）
    RefundedAt time.Time
    CreatedAt  time.Time
}
```

```go
// backend/internal/model/accounting.go — Billing に Refunds リレーション追加
type Billing struct {
    // ... 既存フィールドはそのまま ...
    Items    []BillingItem   `gorm:"foreignKey:BillingID"`
    Payments []Payment       `gorm:"foreignKey:BillingID"`
    Refunds  []BillingRefund `gorm:"foreignKey:BillingID"` // ← 追加
}
```

`make codegen` を実行して `frontend/src/types/generated/models.ts` を更新する。

### 3. Repository 変更

```go
// backend/internal/repository/refund_repository.go（新規）
package repository

import (
    "context"
    "github.com/animal-ekarte/backend/internal/model"
    "gorm.io/gorm"
)

type RefundRepository interface {
    Create(ctx context.Context, refund *model.BillingRefund) error
    FindByBillingID(ctx context.Context, billingID uint64) ([]model.BillingRefund, error)
    SumByBillingID(ctx context.Context, billingID uint64) (int64, error)
}

type refundRepository struct {
    db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) RefundRepository {
    return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *model.BillingRefund) error {
    return r.db.WithContext(ctx).Create(refund).Error
}

func (r *refundRepository) FindByBillingID(ctx context.Context, billingID uint64) ([]model.BillingRefund, error) {
    var refunds []model.BillingRefund
    err := r.db.WithContext(ctx).
        Where("billing_id = ?", billingID).
        Order("refunded_at DESC").
        Find(&refunds).Error
    return refunds, err
}

func (r *refundRepository) SumByBillingID(ctx context.Context, billingID uint64) (int64, error) {
    var total int64
    err := r.db.WithContext(ctx).
        Model(&model.BillingRefund{}).
        Where("billing_id = ?", billingID).
        Select("COALESCE(SUM(amount), 0)").
        Scan(&total).Error
    return total, err
}
```

### 4. Service 変更

```go
// backend/internal/service/refund_service.go（新規）
package service

import (
    "context"
    "fmt"

    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/animal-ekarte/backend/internal/repository"
    "log/slog"
)

type CreateRefundInput struct {
    ClinicID  uint64
    BillingID uint64
    Amount    int64  // 必須、> 0
    Reason    string // 任意
}

type RefundService interface {
    CreateRefund(ctx context.Context, input CreateRefundInput) (*model.BillingRefund, error)
    ListRefunds(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
}

type refundService struct {
    repo           repository.RefundRepository
    accountingRepo repository.AccountingRepository // ← BillingRepository ではなく AccountingRepository（accounting_repository.go:14）
}

func NewRefundService(repo repository.RefundRepository, accountingRepo repository.AccountingRepository) RefundService {
    return &refundService{repo: repo, accountingRepo: accountingRepo}
}

func (s *refundService) CreateRefund(ctx context.Context, input CreateRefundInput) (*model.BillingRefund, error) {
    if input.Amount <= 0 {
        return nil, fmt.Errorf("amount must be positive: %w", apperrors.ErrInvalidInput)
    }

    // billing の存在確認 + status チェック（clinicID も渡してマルチテナント認可）
    billing, err := s.accountingRepo.FindByID(ctx, input.ClinicID, input.BillingID)
    if err != nil {
        return nil, fmt.Errorf("billing not found: %w", apperrors.ErrNotFound)
    }
    if billing.Status != model.BillingStatusCompleted {
        return nil, fmt.Errorf("refund only allowed for completed billing: %w", apperrors.ErrInvalidInput)
    }

    // 返金可能額チェック
    refunded, err := s.repo.SumByBillingID(ctx, input.BillingID)
    if err != nil {
        return nil, fmt.Errorf("failed to sum refunds: %w", err)
    }
    refundable := billing.TotalAmount - refunded
    if input.Amount > refundable {
        return nil, fmt.Errorf("refund amount %d exceeds refundable amount %d: %w",
            input.Amount, refundable, apperrors.ErrInvalidInput)
    }

    refund := &model.BillingRefund{
        ClinicID:  input.ClinicID,
        BillingID: input.BillingID,
        Amount:    input.Amount,
        Reason:    input.Reason,
    }
    if err := s.repo.Create(ctx, refund); err != nil {
        return nil, fmt.Errorf("failed to create refund: %w", err)
    }

    slog.InfoContext(ctx, "refund created",
        slog.Uint64("billing_id", input.BillingID),
        slog.Int64("amount", input.Amount),
        slog.Int64("remaining", refundable-input.Amount),
    )
    return refund, nil
}

func (s *refundService) ListRefunds(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
    // billing が clinicID に属することを先に確認（認可チェック）
    if _, err := s.accountingRepo.FindByID(ctx, clinicID, billingID); err != nil {
        return nil, fmt.Errorf("billing not found: %w", apperrors.ErrNotFound)
    }
    return s.repo.FindByBillingID(ctx, billingID)
}
```

### 5. Handler 変更

```go
// backend/internal/handler/refund_handler.go（新規）
package handler

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/service"
)

type RefundHandler struct {
    service service.RefundService
}

func NewRefundHandler(s service.RefundService) *RefundHandler {
    return &RefundHandler{service: s}
}

// POST /v1/accountings/:id/refunds
func (h *RefundHandler) Create(c *gin.Context) {
    var req CreateRefundRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, fmt.Errorf("invalid request: %w", apperrors.ErrInvalidInput))
        return
    }
    billingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, fmt.Errorf("invalid id: %w", apperrors.ErrInvalidInput))
        return
    }

    refund, err := h.service.CreateRefund(c.Request.Context(), service.CreateRefundInput{
        ClinicID:  getClinicID(c), // ← req.ClinicID は json:"-" で常に0。既存パターンと同様 getClinicID(c) で取得
        BillingID: billingID,
        Amount:    req.Amount,
        Reason:    req.Reason,
    })
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toRefundResponse(refund))
}

// GET /v1/accountings/:id/refunds
func (h *RefundHandler) List(c *gin.Context) {
    billingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, fmt.Errorf("invalid id: %w", apperrors.ErrInvalidInput))
        return
    }
    clinicID := getClinicID(c) // 既存パターン参照
    refunds, err := h.service.ListRefunds(c.Request.Context(), clinicID, billingID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toRefundListResponse(refunds))
}
```

```go
// backend/internal/handler/refund_request.go（新規）
package handler

// ClinicID はフィールドに持たない（json:"-" でゼロ値になるため）。
// handler で getClinicID(c) を使って取得し、service.CreateRefundInput に直接セットする。
type CreateRefundRequest struct {
    Amount int64  `json:"amount" binding:"required,min=1"`
    Reason string `json:"reason"`
}
```

```go
// backend/internal/handler/refund_response.go（新規）
package handler

import (
    "github.com/animal-ekarte/backend/internal/model"
)

type RefundResponse struct {
    ID         uint64 `json:"id"`
    BillingID  uint64 `json:"billing_id"`
    Amount     int64  `json:"amount"`
    Reason     string `json:"reason"`
    RefundedAt string `json:"refunded_at"`
    CreatedAt  string `json:"created_at"`
}

func toRefundResponse(r *model.BillingRefund) RefundResponse {
    return RefundResponse{
        ID:         r.ID,
        BillingID:  r.BillingID,
        Amount:     r.Amount,
        Reason:     r.Reason,
        RefundedAt: r.RefundedAt.Format("2006-01-02T15:04:05Z07:00"),
        CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
    }
}

type RefundListResponse struct {
    Data []RefundResponse `json:"data"`
}

func toRefundListResponse(refunds []model.BillingRefund) RefundListResponse {
    items := make([]RefundResponse, len(refunds))
    for i, r := range refunds {
        items[i] = toRefundResponse(&r)
    }
    return RefundListResponse{Data: items}
}
```

### 6. DI 配線（cmd/api/main.go）

既存の accountings ルートグループに追加:

```go
// backend/cmd/api/main.go（追記部分のみ）
refundRepo    := repository.NewRefundRepository(db)
refundSvc     := service.NewRefundService(refundRepo, accountingRepo) // accountingRepo は既存の DI から取得
refundHandler := handler.NewRefundHandler(refundSvc)

// accountings グループ内に追加
accountings.GET("/:id/refunds", refundHandler.List)
accountings.POST("/:id/refunds", refundHandler.Create)
```

### 7. GET /accountings レスポンスに total_refunded_amount を追加

**⚠️ 注意**: `FindAll` は `Preload("Owner").Preload("Pet").Preload("Payments").Preload("Items")` を使用している。ここに `LEFT JOIN + GROUP BY` を追加すると GORM の Preload と競合してクエリが壊れる。サブクエリ方式を採用する。

```go
// backend/internal/model/accounting.go — 仮想フィールド追加
type Billing struct {
    // ... 既存フィールドはそのまま ...
    TotalRefundedAmount int64 `gorm:"-" json:"total_refunded_amount"` // 仮想フィールド（DB列なし）
}
```

```go
// backend/internal/repository/accounting_repository.go — FindAll の Select を修正
// ⚠️ Select の追加は q.Count(&total) の後に行うこと。
//   Count より前に Select を付けると GORM がサブクエリをラップして
//   SELECT count(*) FROM (SELECT billings.*, ...) AS count になる恐れがある。
//
// 修正後の FindAll（変更箇所のみ）:
//   if err := q.Count(&total).Error; err != nil { ... }      ← Count は Select なしで実行
//   q = q.Select(`billings.*,
//       (SELECT COALESCE(SUM(amount), 0) FROM billing_refunds WHERE billing_id = billings.id) AS total_refunded_amount`)
//   if err := q.Preload(...).Find(&billings).Error; err != nil { ... }  ← Find にサブクエリ適用
//
// GROUP BY 不要。Preload はそのまま維持。
```

```go
// backend/internal/handler/accounting_response.go（既存ファイル修正）
// response 型に total_refunded_amount を追加
type AccountingListItem struct {
    // ... 既存フィールド ...
    TotalRefundedAmount int64 `json:"total_refunded_amount"` // ← 追加
}
// toAccountingListItem() でも billing.TotalRefundedAmount を詰める
```

## API レスポンス形式

**POST /v1/accountings/:id/refunds**
```json
{
  "id": 1,
  "billing_id": 42,
  "amount": 2000,
  "reason": "診断ミスのため",
  "refunded_at": "2026-03-26T10:00:00Z",
  "created_at": "2026-03-26T10:00:00Z"
}
```

**GET /v1/accountings/:id/refunds**
```json
{
  "data": [
    {
      "id": 2,
      "billing_id": 42,
      "amount": 3000,
      "reason": "",
      "refunded_at": "2026-03-26T11:00:00Z",
      "created_at": "2026-03-26T11:00:00Z"
    },
    {
      "id": 1,
      "billing_id": 42,
      "amount": 2000,
      "reason": "診断ミスのため",
      "refunded_at": "2026-03-26T10:00:00Z",
      "created_at": "2026-03-26T10:00:00Z"
    }
  ]
}
```

## フロントエンド影響

- `make codegen` で `BillingRefund` 型が `models.ts` に追加される
- FE-126 で対応が必要

## 完了条件

- [ ] DB マイグレーション適用（`billing_refunds` テーブル作成）
- [ ] `BillingRefund` モデル + `Billing.Refunds` リレーション追加
- [ ] `make codegen` で `models.ts` が更新される（`BillingRefund` 型が生成される）
- [ ] `POST /v1/accountings/:id/refunds` で返金レコードが作成できる
- [ ] `GET /v1/accountings/:id/refunds` で返金一覧が取得できる
- [ ] 返金額超過時に 400 エラーが返る
- [ ] `completed` 以外の billings に対して 400 エラーが返る
- [ ] `GET /v1/accountings` レスポンスに `total_refunded_amount` が含まれる
- [ ] 既存テストが通る（`docker compose exec backend go test ./... -v`）
