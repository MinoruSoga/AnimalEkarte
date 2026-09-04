package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// このファイルは accounting_service_builders.go の純粋関数（representativeMethod /
// buildPaymentFromInput / buildAccountingUpdate）に対する直接テストを対象とする。
// これらはリポジトリ依存を持たないため、モックなしで直接呼び出して検証する。

func TestAccountingServiceBuilders_RepresentativeMethod(t *testing.T) {
	tests := []struct {
		name   string
		splits []PaymentSplitInput
		want   model.PaymentMethod
	}{
		{
			name:   "空のsplitsはelectronic_moneyを返す",
			splits: nil,
			want:   model.PaymentMethodElectronicMoney,
		},
		{
			name: "cashが最優先",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodElectronicMoney, Amount: 100},
				{Method: model.PaymentMethodCash, Amount: 200},
				{Method: model.PaymentMethodCreditCard, Amount: 300},
			},
			want: model.PaymentMethodCash,
		},
		{
			name: "cashがない場合はcredit_cardが優先",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodBankTransfer, Amount: 100},
				{Method: model.PaymentMethodCreditCard, Amount: 200},
			},
			want: model.PaymentMethodCreditCard,
		},
		{
			name: "cash/credit_cardがない場合はbank_transferが優先(#198)",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodElectronicMoney, Amount: 100},
				{Method: model.PaymentMethodBankTransfer, Amount: 200},
			},
			want: model.PaymentMethodBankTransfer,
		},
		{
			name: "electronic_moneyのみの場合はelectronic_money",
			splits: []PaymentSplitInput{
				{Method: model.PaymentMethodElectronicMoney, Amount: 100},
			},
			want: model.PaymentMethodElectronicMoney,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, representativeMethod(tt.splits))
		})
	}
}

func TestBuildPaymentFromInput(t *testing.T) {
	billingID := uint64(1)
	staffID := uint64(2)

	t.Run("フィールドが全てnilの場合はBillingID/PaidByのみ設定", func(t *testing.T) {
		input := &UpdateAccountingInput{ID: billingID, StaffID: &staffID}
		p := buildPaymentFromInput(input)
		assert.Equal(t, billingID, p.BillingID)
		assert.Equal(t, &staffID, p.PaidBy)
		assert.Equal(t, int64(0), p.Subtotal)
		assert.Equal(t, model.PaymentMethod(""), p.Method)
	})

	t.Run("単一支払いフィールドが全て設定されている場合", func(t *testing.T) {
		subtotal := int64(1000)
		taxTotal := int64(100)
		totalAmount := int64(1100)
		insuranceName := "ペット保険"
		insuranceRatio := 0.5
		insuranceAmount := int64(550)
		discountAmount := int64(50)
		billingAmount := int64(1050)
		receivedAmount := int64(2000)
		changeAmount := int64(950)
		method := model.PaymentMethodCash

		input := &UpdateAccountingInput{
			ID:              billingID,
			StaffID:         &staffID,
			Subtotal:        &subtotal,
			TaxTotal:        &taxTotal,
			TotalAmount:     &totalAmount,
			InsuranceName:   &insuranceName,
			InsuranceRatio:  &insuranceRatio,
			InsuranceAmount: &insuranceAmount,
			DiscountAmount:  &discountAmount,
			BillingAmount:   &billingAmount,
			ReceivedAmount:  &receivedAmount,
			ChangeAmount:    &changeAmount,
			PaymentMethod:   &method,
		}

		p := buildPaymentFromInput(input)

		assert.Equal(t, subtotal, p.Subtotal)
		assert.Equal(t, taxTotal, p.TaxTotal)
		assert.Equal(t, totalAmount, p.TotalAmount)
		assert.Equal(t, insuranceName, p.InsuranceName)
		assert.Equal(t, insuranceRatio, p.InsuranceRatio)
		assert.Equal(t, insuranceAmount, p.InsuranceAmount)
		assert.Equal(t, discountAmount, p.DiscountAmount)
		assert.Equal(t, billingAmount, p.BillingAmount)
		assert.Equal(t, receivedAmount, p.ReceivedAmount)
		assert.Equal(t, changeAmount, p.ChangeAmount)
		assert.Equal(t, method, p.Method)
	})

	t.Run("PaymentSplitsにcashが含まれる場合はcashのReceived/Changeを採用", func(t *testing.T) {
		input := &UpdateAccountingInput{
			ID: billingID,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, Amount: 500},
				{Method: model.PaymentMethodCash, Amount: 300, ReceivedAmount: 500, ChangeAmount: 200},
			},
		}

		p := buildPaymentFromInput(input)

		assert.Equal(t, model.PaymentMethodCash, p.Method) // cash が代表手段
		assert.Equal(t, int64(500), p.ReceivedAmount)
		assert.Equal(t, int64(200), p.ChangeAmount)
	})

	t.Run("PaymentSplitsにcashが含まれない場合はReceived/Changeはゼロのまま", func(t *testing.T) {
		input := &UpdateAccountingInput{
			ID: billingID,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCreditCard, Amount: 500},
			},
		}

		p := buildPaymentFromInput(input)

		assert.Equal(t, model.PaymentMethodCreditCard, p.Method)
		assert.Equal(t, int64(0), p.ReceivedAmount)
		assert.Equal(t, int64(0), p.ChangeAmount)
	})
}

func TestBuildAccountingUpdate(t *testing.T) {
	medicalRecordID := uint64(1)
	hospitalizationID := uint64(2)
	ownerID := uint64(3)
	petID := uint64(4)
	subtotal := int64(1000)
	taxTotal := int64(100)
	totalAmount := int64(1100)
	hasInsurance := true
	scheduledDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	memo := "備考"

	t.Run("全フィールドnilの場合は空map", func(t *testing.T) {
		got := buildAccountingUpdate(&UpdateAccountingInput{})
		assert.Empty(t, got)
	})

	t.Run("Status以外の全フィールド設定", func(t *testing.T) {
		input := &UpdateAccountingInput{
			MedicalRecordID:   &medicalRecordID,
			HospitalizationID: &hospitalizationID,
			OwnerID:           &ownerID,
			PetID:             &petID,
			Subtotal:          &subtotal,
			TaxTotal:          &taxTotal,
			TotalAmount:       &totalAmount,
			HasInsurance:      &hasInsurance,
			ScheduledDate:     &scheduledDate,
			Memo:              &memo,
		}
		got := buildAccountingUpdate(input)

		assert.Equal(t, medicalRecordID, got["medical_record_id"])
		assert.Equal(t, hospitalizationID, got["hospitalization_id"])
		assert.Equal(t, ownerID, got["owner_id"])
		assert.Equal(t, petID, got["pet_id"])
		assert.Equal(t, subtotal, got["subtotal"])
		assert.Equal(t, taxTotal, got["tax_total"])
		assert.Equal(t, totalAmount, got["total_amount"])
		assert.Equal(t, hasInsurance, got["has_insurance"])
		assert.Equal(t, scheduledDate, got["scheduled_date"])
		assert.Equal(t, memo, got["memo"])
		assert.NotContains(t, got, "status")
		assert.NotContains(t, got, "completed_at")
	})

	t.Run("Statusとcompleted_atは汎用更新mapに出ない", func(t *testing.T) {
		status := model.BillingStatusCompleted
		completedAt := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
		got := buildAccountingUpdate(&UpdateAccountingInput{Status: &status, CompletedAt: &completedAt})

		assert.NotContains(t, got, "status")
		assert.NotContains(t, got, "completed_at")
	})
}
