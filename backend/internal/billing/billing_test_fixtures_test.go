package billing

// billing_test_fixtures_test.go — internal/repository の同名 fixture の複製
// （accounting 系残留 test が原本を継続使用・B④移動時に片側へ統合）。

// billing_test_fixtures_test.go — BE-refactor.md R30
//
// 「最小構成の Billing を INSERT する」テストビルダーが、status/金額/日付フィールドだけ異なる
// 状態で makeBilling / makeBillingRet / makeBillingForAccountingTx / makeBillingForItemTx /
// makeBillingForRefund / makeTrimmingBilling / makeTrimmingBillingWithCompletedAt の 7 変種として
// 併存していた。本ファイルは共通の組み立てロジックを makeBillingWith に一本化し、
// 各変種は既存のシグネチャ（呼び出し元は無変更）を保ったまま thin wrapper として委譲する。
//
// 各フィールドの既定値は billingFixtureOpts のゼロ値がそのまま model.Billing のゼロ値として
// INSERT される（例: TotalAmount=0, OwnerID=nil, PetID=nil, CompletedAt=nil）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// billingFixtureOpts は makeBillingWith に渡す Billing 組み立てオプション。
type billingFixtureOpts struct {
	ClinicID      uint64
	OwnerID       *uint64
	PetID         *uint64
	TotalAmount   int64
	Status        model.BillingStatus
	ScheduledDate time.Time
	CompletedAt   *time.Time
}

// makeBillingWith は billingFixtureOpts に基づいて最小構成の Billing を作成する共有テストヘルパー。
func makeBillingWith(t *testing.T, db *gorm.DB, opts billingFixtureOpts) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      opts.ClinicID,
		OwnerID:       opts.OwnerID,
		PetID:         opts.PetID,
		TotalAmount:   opts.TotalAmount,
		Status:        opts.Status,
		ScheduledDate: opts.ScheduledDate,
		CompletedAt:   opts.CompletedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}
