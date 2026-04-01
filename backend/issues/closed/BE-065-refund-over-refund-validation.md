# BE-065: 返金サービスに過剰返金チェックを追加

**Status**: Open
**Priority**: High
**Affects**: 返金機能（`POST /v1/accountings/:id/refunds`）
**Date Created**: 2026-03-26
**Related**: TASK-031, FE-127

## Summary

`refundService.Create` が返金可能残額を検証していないため、元の請求金額を超える返金が登録できてしまう。
`SumByBillingID` はリポジトリに実装済みのため、サービス層に残額チェックを追加するだけで解決する。

## 現状のコード

```go
// backend/internal/service/refund_service.go:28-52
func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, amount int64, reason string) (*model.BillingRefund, error) {
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	// 請求が存在するか確認（マルチテナント保護）
	if _, err := s.accountRepo.FindByID(ctx, clinicID, billingID); err != nil {
		return nil, err
	}

	// ← ここに残額チェックが存在しない
	// 現状では元の請求額を超える返金が登録できてしまう

	refund := &model.BillingRefund{
		ClinicID:  clinicID,
		BillingID: billingID,
		Amount:    amount,
		Reason:    reason,
	}
	if err := s.repo.Create(ctx, refund); err != nil {
		return nil, err
	}
	...
}
```

## 必要な変更

### `backend/internal/service/refund_service.go`

```go
// Before（:28-36）
func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, amount int64, reason string) (*model.BillingRefund, error) {
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	// 請求が存在するか確認（マルチテナント保護）
	if _, err := s.accountRepo.FindByID(ctx, clinicID, billingID); err != nil {
		return nil, err
	}

// After
func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, amount int64, reason string) (*model.BillingRefund, error) {
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	// 請求が存在するか確認（マルチテナント保護）
	billing, err := s.accountRepo.FindByID(ctx, clinicID, billingID)
	if err != nil {
		return nil, err
	}

	// 返金可能残額チェック（過剰返金防止）
	if len(billing.Payments) > 0 {
		alreadyRefunded, err := s.repo.SumByBillingID(ctx, billingID)
		if err != nil {
			return nil, fmt.Errorf("sum refunds: %w", err)
		}
		available := billing.Payments[0].TotalAmount - alreadyRefunded
		if amount > available {
			return nil, apperrors.WrapInvalidInput(
				fmt.Sprintf("refund amount %d exceeds available balance %d", amount, available))
		}
	}
```

### import ブロックへの追加

```go
import (
	"context"
	"fmt"   // ← 追加
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)
```

## 前提知識

- `AccountingRepository.FindByID` は `Payments` を `Preload` 済み（`accounting_repository.go:92`）
- `RefundRepository.SumByBillingID` は `billing_id` の返金合計を返す（`refund_repository.go:47`）
- `Payment.TotalAmount int64` が精算済みの確定請求金額（`model/accounting.go:111`）
- `Billing.Payments` が空の場合（未精算）は返金操作自体が発生しないため、ガードは `len(billing.Payments) > 0` で十分

## フロントエンド影響

- BE がエラーを返すと FE の `catch` ブランチが `toast.error("返金の登録に失敗しました")` を表示（既実装）
- FE-127 で残額表示・ボタン無効化を追加することで、エラーに至る前に UI 側でも抑止できる

## 完了条件

- [ ] `refundService.Create` に `SumByBillingID` による残額チェックを追加
- [ ] `fmt` import 追加
- [ ] 返金可能残額 ¥0 の請求に返金しようとすると `400 Bad Request` が返ること
- [ ] 過剰額（例: ¥1,000 の残額に ¥1,001 を送信）で `400 Bad Request` が返ること
- [ ] 正常な返金（残額以内）は従来通り `201 Created` が返ること
