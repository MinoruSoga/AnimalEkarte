package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- toDailySummaryResponse ----

func TestToDailySummaryResponse(t *testing.T) {
	tests := []struct {
		name string
		in   *DailySummaryResult
		want dailySummaryResponse
	}{
		{
			name: "normal: payment/category totals populated",
			in: &DailySummaryResult{
				PaymentTotals:  []PaymentMethodTotal{{Method: "cash", Total: 1000}, {Method: "credit_card", Total: 2000}},
				CategoryTotals: []CategoryTotal{{Category: "examination", Total: 1500}},
				BillingCount:   3,
				GrandTotal:     3000,
			},
			want: dailySummaryResponse{
				PaymentTotals:  []paymentMethodTotalResponse{{Method: "cash", Total: 1000}, {Method: "credit_card", Total: 2000}},
				CategoryTotals: []categoryTotalResponse{{Category: "examination", Total: 1500}},
				BillingCount:   3,
				GrandTotal:     3000,
			},
		},
		{
			name: "zero value: empty slices, zero counts",
			in:   &DailySummaryResult{},
			want: dailySummaryResponse{
				PaymentTotals:  []paymentMethodTotalResponse{},
				CategoryTotals: []categoryTotalResponse{},
				BillingCount:   0,
				GrandTotal:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDailySummaryResponse(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- toDailySummaryForClinicsResponse ----

func TestToDailySummaryForClinicsResponse(t *testing.T) {
	tests := []struct {
		name string
		in   []ClinicDailySummary
		want int // expected len(PerClinic)
	}{
		{
			name: "normal: multiple clinics",
			in: []ClinicDailySummary{
				{ClinicID: 1, Summary: &DailySummaryResult{BillingCount: 1, GrandTotal: 100}},
				{ClinicID: 2, Summary: &DailySummaryResult{BillingCount: 2, GrandTotal: 200}},
			},
			want: 2,
		},
		{
			name: "empty: no clinics",
			in:   []ClinicDailySummary{},
			want: 0,
		},
		{
			name: "nil slice",
			in:   nil,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDailySummaryForClinicsResponse(tt.in)
			assert.Len(t, got.PerClinic, tt.want)
			if tt.want == 2 {
				assert.Equal(t, uint64(1), got.PerClinic[0].ClinicID)
				assert.Equal(t, int64(100), got.PerClinic[0].Summary.GrandTotal)
				assert.Equal(t, uint64(2), got.PerClinic[1].ClinicID)
				assert.Equal(t, int64(200), got.PerClinic[1].Summary.GrandTotal)
			}
		})
	}
}

// ---- toUnpaidByOwnerResponse ----

func TestToUnpaidByOwnerResponse(t *testing.T) {
	tests := []struct {
		name    string
		items   []UnpaidOwnerAggregate
		total   int64
		page    int
		limit   int
		summary UnpaidSummary
		wantLen int
	}{
		{
			name: "normal: one owner aggregate",
			items: []UnpaidOwnerAggregate{
				{OwnerID: 1, OwnerName: "田中太郎", Count: 2, TotalAmount: 5000, OldestScheduled: "2026-01-01", LatestScheduled: "2026-01-15"},
			},
			total:   1,
			page:    1,
			limit:   20,
			summary: UnpaidSummary{TotalAmount: 5000, BillingCount: 2, OwnerCount: 1},
			wantLen: 1,
		},
		{
			name:    "empty items",
			items:   []UnpaidOwnerAggregate{},
			total:   0,
			page:    1,
			limit:   20,
			summary: UnpaidSummary{},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUnpaidByOwnerResponse(tt.items, tt.total, tt.page, tt.limit, tt.summary)
			assert.Len(t, got.Data, tt.wantLen)
			assert.Equal(t, tt.total, got.Total)
			assert.Equal(t, tt.page, got.Page)
			assert.Equal(t, tt.limit, got.Limit)
			assert.Equal(t, tt.summary.TotalAmount, got.Summary.TotalAmount)
			assert.Equal(t, tt.summary.BillingCount, got.Summary.BillingCount)
			assert.Equal(t, tt.summary.OwnerCount, got.Summary.OwnerCount)
			if tt.wantLen == 1 {
				assert.Equal(t, uint64(1), got.Data[0].OwnerID)
				assert.Equal(t, "田中太郎", got.Data[0].OwnerName)
				assert.Equal(t, "2026-01-01", got.Data[0].OldestScheduled)
				assert.Equal(t, "2026-01-15", got.Data[0].LatestScheduled)
			}
		})
	}
}

// ---- toOwnerUnpaidBalanceResponse ----

func TestToOwnerUnpaidBalanceResponse(t *testing.T) {
	tests := []struct {
		name string
		in   OwnerUnpaidBalance
		want ownerUnpaidBalanceResponse
	}{
		{
			name: "normal: non-zero balance",
			in:   OwnerUnpaidBalance{TotalAmount: 12000, Count: 3},
			want: ownerUnpaidBalanceResponse{UnpaidTotal: 12000, UnpaidCount: 3},
		},
		{
			name: "zero value: no unpaid balance",
			in:   OwnerUnpaidBalance{},
			want: ownerUnpaidBalanceResponse{UnpaidTotal: 0, UnpaidCount: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toOwnerUnpaidBalanceResponse(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- toMonthlyUnpaidCarryoverResponse ----

func TestToMonthlyUnpaidCarryoverResponse(t *testing.T) {
	petID := uint64(9)
	tests := []struct {
		name       string
		items      []MonthlyUnpaidOwnerPet
		summary    MonthlyUnpaidSummary
		wantLen    int
		wantPetNil bool
	}{
		{
			name: "normal: pet_id present",
			items: []MonthlyUnpaidOwnerPet{
				{OwnerID: 1, OwnerName: "田中太郎", PetID: &petID, PetName: "ポチ", PrevMonthCarryover: 1000, CurrentMonthUnpaid: 2000, NextMonthCarryover: 3000},
			},
			summary: MonthlyUnpaidSummary{PrevMonthCarryover: 1000, CurrentMonthUnpaid: 2000, NextMonthCarryover: 3000},
			wantLen: 1,
		},
		{
			name: "pet_id nil: owner-level record without a specific pet",
			items: []MonthlyUnpaidOwnerPet{
				{OwnerID: 2, OwnerName: "鈴木花子", PetID: nil},
			},
			wantLen:    1,
			wantPetNil: true,
		},
		{
			name:    "empty items",
			items:   []MonthlyUnpaidOwnerPet{},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMonthlyUnpaidCarryoverResponse(tt.items, int64(tt.wantLen), 1, 20, tt.summary)
			assert.Len(t, got.Data, tt.wantLen)
			assert.Equal(t, tt.summary.PrevMonthCarryover, got.Summary.PrevMonthCarryover)
			if tt.wantLen == 1 {
				if tt.wantPetNil {
					assert.Nil(t, got.Data[0].PetID)
				} else {
					require.NotNil(t, got.Data[0].PetID)
					assert.Equal(t, petID, *got.Data[0].PetID)
				}
			}
		})
	}
}

// ---- ToRefundResponse ----

func TestToRefundResponse(t *testing.T) {
	refundedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC)
	staffID := uint64(5)
	cash := model.PaymentMethodCash

	tests := []struct {
		name           string
		in             *model.BillingRefund
		wantStaffName  string
		wantPayMethod  *string
		wantPayMethNil bool
	}{
		{
			name: "normal: refunded_by_staff and payment_method populated",
			in: &model.BillingRefund{
				ID: 1, ClinicID: 1, BillingID: 10, Amount: 3000, Reason: "過剰請求",
				RefundedBy:      &staffID,
				RefundedByStaff: &model.Staff{ID: staffID, Name: "山田スタッフ"},
				PaymentMethod:   &cash,
				RefundedAt:      refundedAt,
				CreatedAt:       createdAt,
			},
			wantStaffName: "山田スタッフ",
		},
		{
			name: "nil relations: RefundedByStaff and PaymentMethod nil",
			in: &model.BillingRefund{
				ID: 2, ClinicID: 1, BillingID: 11, Amount: 1000, Reason: "",
				RefundedAt: refundedAt,
				CreatedAt:  createdAt,
			},
			wantStaffName:  "",
			wantPayMethNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRefundResponse(tt.in)
			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.in.Amount, got.Amount)
			assert.Equal(t, tt.wantStaffName, got.RefundedByName)
			if tt.wantPayMethNil {
				assert.Nil(t, got.PaymentMethod)
			} else {
				require.NotNil(t, got.PaymentMethod)
				assert.Equal(t, string(cash), *got.PaymentMethod)
			}
		})
	}
}

// ---- ToBillingItemResponse ----

func TestToBillingItemResponse(t *testing.T) {
	tests := []struct {
		name          string
		in            *model.BillingItem
		wantSubtotal  int64
		wantTaxAmount int64
	}{
		{
			name: "normal: excluded tax, no discount",
			in: &model.BillingItem{
				ID: 1, BillingID: 10, Category: model.ItemCategoryExamination, Name: "診察",
				UnitPrice: 1000, Quantity: 2, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
			},
			wantSubtotal:  2000,
			wantTaxAmount: 200,
		},
		{
			name: "discount larger than subtotal clamps to zero (#85)",
			in: &model.BillingItem{
				ID: 2, BillingID: 10, Category: model.ItemCategoryExamination, Name: "割引品",
				UnitPrice: 100, Quantity: 1, DiscountAmount: 500, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
			},
			wantSubtotal:  0,
			wantTaxAmount: 0,
		},
		{
			name: "included tax computed from post-discount subtotal",
			in: &model.BillingItem{
				ID: 3, BillingID: 10, Category: model.ItemCategoryExamination, Name: "内税品",
				UnitPrice: 1100, Quantity: 1, TaxType: model.TaxTypeIncluded, TaxRate: 0.10,
			},
			wantSubtotal:  1100,
			wantTaxAmount: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToBillingItemResponse(tt.in)
			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.wantSubtotal, got.Subtotal)
			assert.Equal(t, tt.wantTaxAmount, got.TaxAmount)
			assert.Equal(t, string(tt.in.Category), got.Category)
			assert.Equal(t, string(tt.in.TaxType), got.TaxType)
		})
	}
}

// #251 U2a-1 / BUG-436: 物販マスタ由来の provenance は応答DTOでも往復しなければならない。
// 応答に載らないと frontend 側の型だけが当該フィールドを持ち、実行時は常に undefined になる
// （BUG-433 と同型のサイレントな乖離）。
func TestToBillingItemResponse_MerchandiseItemIDRoundTrips(t *testing.T) {
	merchandiseID := uint64(77)

	t.Run("emits merchandise_item_id when the item carries it", func(t *testing.T) {
		got := ToBillingItemResponse(&model.BillingItem{
			ID: 1, BillingID: 10, Category: model.ItemCategoryGoods, Name: "療法食",
			UnitPrice: 1200, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
			MerchandiseItemID: &merchandiseID,
		})
		require.NotNil(t, got.MerchandiseItemID)
		assert.Equal(t, merchandiseID, *got.MerchandiseItemID)
	})

	t.Run("omits merchandise_item_id for manually entered items", func(t *testing.T) {
		got := ToBillingItemResponse(&model.BillingItem{
			ID: 2, BillingID: 10, Category: model.ItemCategoryOther, Name: "手入力",
			UnitPrice: 500, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
		})
		assert.Nil(t, got.MerchandiseItemID)
	})
}

// ---- toPaymentResponse ----

func TestToPaymentResponse(t *testing.T) {
	staffID := uint64(7)
	tests := []struct {
		name          string
		in            *model.Payment
		wantStaffName string
	}{
		{
			name: "normal: paid_by_staff populated",
			in: &model.Payment{
				ID: 1, BillingID: 10, TotalAmount: 5000, Method: model.PaymentMethodCash,
				PaidBy: &staffID, PaidByStaff: &model.Staff{ID: staffID, Name: "受付スタッフ"},
			},
			wantStaffName: "受付スタッフ",
		},
		{
			name: "nil PaidByStaff",
			in: &model.Payment{
				ID: 2, BillingID: 11, TotalAmount: 3000, Method: model.PaymentMethodCash,
			},
			wantStaffName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPaymentResponse(tt.in)
			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.in.TotalAmount, got.TotalAmount)
			assert.Equal(t, tt.wantStaffName, got.PaidByName)
			assert.Equal(t, string(tt.in.Method), got.Method)
		})
	}
}

// ---- toPaymentSplitResponse ----

func TestToPaymentSplitResponse(t *testing.T) {
	staffID := uint64(8)
	tests := []struct {
		name          string
		in            *model.PaymentSplit
		wantStaffName string
	}{
		{
			name: "normal: paid_by_staff populated",
			in: &model.PaymentSplit{
				ID: 1, ClinicID: 1, BillingID: 10, Method: model.PaymentMethodCreditCard, Amount: 1200,
				PaidBy: &staffID, PaidByStaff: &model.Staff{ID: staffID, Name: "レジ担当"},
			},
			wantStaffName: "レジ担当",
		},
		{
			name: "nil PaidByStaff",
			in: &model.PaymentSplit{
				ID: 2, ClinicID: 1, BillingID: 11, Method: model.PaymentMethodCash, Amount: 800,
			},
			wantStaffName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPaymentSplitResponse(tt.in)
			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.in.Amount, got.Amount)
			assert.Equal(t, tt.wantStaffName, got.PaidByName)
			assert.Equal(t, string(tt.in.Method), got.Method)
		})
	}
}

// ---- toAccountingResponse ----

func TestToAccountingResponse(t *testing.T) {
	ownerID := uint64(1)
	petID := uint64(2)
	scheduledDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	t.Run("normal: owner/pet/items/payments/splits/refunds all populated", func(t *testing.T) {
		b := &model.Billing{
			ID: 1, ClinicID: 1, OwnerID: &ownerID, PetID: &petID,
			Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100, TotalRefundedAmount: 200,
			HasInsurance: true, Status: model.BillingStatusCompleted,
			ScheduledDate: scheduledDate, CompletedAt: &completedAt, Memo: "備考",
			Owner: &model.Owner{ID: ownerID, Name: "田中太郎"},
			Pet:   &model.Pet{ID: petID, Name: "ポチ"},
			Items: []model.BillingItem{
				{ID: 1, BillingID: 1, Category: model.ItemCategoryExamination, Name: "診察", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
			},
			Payments: []model.Payment{
				{ID: 1, BillingID: 1, TotalAmount: 1100, Method: model.PaymentMethodCash},
			},
			PaymentSplits: []model.PaymentSplit{
				{ID: 1, ClinicID: 1, BillingID: 1, Method: model.PaymentMethodCash, Amount: 1100},
			},
			Refunds: []model.BillingRefund{
				{ID: 1, ClinicID: 1, BillingID: 1, Amount: 200, RefundedAt: scheduledDate},
			},
		}

		got := toAccountingResponse(b)

		assert.Equal(t, uint64(1), got.ID)
		require.NotNil(t, got.Owner)
		assert.Equal(t, "田中太郎", got.Owner.OwnerName)
		require.NotNil(t, got.Pet)
		assert.Equal(t, "ポチ", got.Pet.Name)
		assert.Equal(t, int64(200), got.TotalRefundedAmount)
		assert.True(t, got.HasInsurance)
		assert.Equal(t, string(model.BillingStatusCompleted), got.Status)
		assert.Equal(t, "備考", got.Memo)
		assert.Len(t, got.Items, 1)
		assert.Len(t, got.Payments, 1)
		assert.Len(t, got.PaymentSplits, 1)
		assert.Len(t, got.Refunds, 1)
		require.NotNil(t, got.CompletedAt)
	})

	t.Run("zero value: nil owner/pet and empty relation slices", func(t *testing.T) {
		b := &model.Billing{
			ID: 2, ClinicID: 1, ScheduledDate: scheduledDate,
		}

		got := toAccountingResponse(b)

		assert.Nil(t, got.Owner)
		assert.Nil(t, got.Pet)
		assert.Nil(t, got.CompletedAt)
		assert.Empty(t, got.Items)
		assert.Empty(t, got.Payments)
		assert.Empty(t, got.PaymentSplits)
		assert.Empty(t, got.Refunds)
	})
}
