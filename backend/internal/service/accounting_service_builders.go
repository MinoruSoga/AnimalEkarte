package service

import (
	"fmt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// hasPaymentFields は UpdateAccountingInput に Payment 関連フィールドが含まれているか判定する。
func hasPaymentFields(input *UpdateAccountingInput) bool {
	return len(input.PaymentSplits) > 0 ||
		input.PaymentMethod != nil ||
		input.InsuranceRatio != nil ||
		input.InsuranceAmount != nil ||
		input.BillingAmount != nil ||
		input.ReceivedAmount != nil ||
		input.ChangeAmount != nil ||
		input.DiscountAmount != nil
}

// representativeMethod は splits から payments.method（代表手段）を導出する仕様ロジック（PO判断B 2026-05-25: 確定）。
// 優先順位は仕様として固定: cash > credit_card > electronic_money
func representativeMethod(splits []PaymentSplitInput) model.PaymentMethod {
	for _, s := range splits {
		if s.Method == model.PaymentMethodCash {
			return model.PaymentMethodCash
		}
	}
	for _, s := range splits {
		if s.Method == model.PaymentMethodCreditCard {
			return model.PaymentMethodCreditCard
		}
	}
	return model.PaymentMethodElectronicMoney
}

// validatePaymentSplits は splits の整合性を検証する。
func validatePaymentSplits(splits []PaymentSplitInput, billingAmount *int64) error {
	if len(splits) == 0 {
		return nil
	}
	seen := make(map[model.PaymentMethod]bool, len(splits))
	var total int64
	for _, s := range splits {
		if seen[s.Method] {
			return apperrors.WrapInvalidInput(fmt.Sprintf("支払い手段 %s が重複しています", s.Method))
		}
		seen[s.Method] = true
		if s.Amount <= 0 {
			return apperrors.WrapInvalidInput("各支払い金額は1円以上でなければなりません")
		}
		total += s.Amount
		if s.Method == model.PaymentMethodCash {
			if s.ReceivedAmount < s.Amount {
				return apperrors.WrapInvalidInput("現金の預り金が不足しています")
			}
			if s.ChangeAmount != s.ReceivedAmount-s.Amount {
				return apperrors.WrapInvalidInput("お釣り計算が不正です")
			}
		}
	}
	if billingAmount != nil && total != *billingAmount {
		return apperrors.WrapInvalidInput(fmt.Sprintf("支払い内訳の合計（%d）が請求金額（%d）と一致しません", total, *billingAmount))
	}
	return nil
}

// buildPaymentFromInput は UpdateAccountingInput から Payment モデルを構築する。
// splits がある場合は代表支払い手段・受領額・お釣りを splits から導出する。
func buildPaymentFromInput(input *UpdateAccountingInput) *model.Payment {
	p := &model.Payment{
		BillingID: input.ID,
		PaidBy:    input.StaffID,
	}
	if input.Subtotal != nil {
		p.Subtotal = *input.Subtotal
	}
	if input.TaxTotal != nil {
		p.TaxTotal = *input.TaxTotal
	}
	if input.TotalAmount != nil {
		p.TotalAmount = *input.TotalAmount
	}
	if input.InsuranceName != nil {
		p.InsuranceName = *input.InsuranceName
	}
	if input.InsuranceRatio != nil {
		p.InsuranceRatio = *input.InsuranceRatio
	}
	if input.InsuranceAmount != nil {
		p.InsuranceAmount = *input.InsuranceAmount
	}
	if input.DiscountAmount != nil {
		p.DiscountAmount = *input.DiscountAmount
	}
	if input.BillingAmount != nil {
		p.BillingAmount = *input.BillingAmount
	}

	if len(input.PaymentSplits) > 0 {
		// 混在会計: splits の一次情報から代表手段・受領額・お釣りを導出（仕様）
		p.Method = representativeMethod(input.PaymentSplits)
		for _, s := range input.PaymentSplits {
			if s.Method == model.PaymentMethodCash {
				p.ReceivedAmount = s.ReceivedAmount
				p.ChangeAmount = s.ChangeAmount
				break
			}
		}
	} else {
		if input.ReceivedAmount != nil {
			p.ReceivedAmount = *input.ReceivedAmount
		}
		if input.ChangeAmount != nil {
			p.ChangeAmount = *input.ChangeAmount
		}
		if input.PaymentMethod != nil {
			p.Method = *input.PaymentMethod
		}
	}
	return p
}

// buildPaymentSplits は UpdateAccountingInput から PaymentSplit モデルのスライスを構築する。
// PaymentSplits が空の場合は単一支払いフィールドから1行を生成する（backward compat）。
func buildPaymentSplits(input *UpdateAccountingInput) []model.PaymentSplit {
	if len(input.PaymentSplits) > 0 {
		splits := make([]model.PaymentSplit, 0, len(input.PaymentSplits))
		for _, s := range input.PaymentSplits {
			splits = append(splits, model.PaymentSplit{
				ClinicID:        input.ClinicID,
				BillingID:       input.ID,
				Method:          s.Method,
				PaymentMethodID: s.PaymentMethodID,
				Amount:          s.Amount,
				ReceivedAmount:  s.ReceivedAmount,
				ChangeAmount:    s.ChangeAmount,
				PaidBy:          input.StaffID,
			})
		}
		return splits
	}
	// 単一支払い backward compat — BillingAmount が設定されている場合のみ生成
	if input.BillingAmount == nil || *input.BillingAmount <= 0 {
		return nil
	}
	method := model.PaymentMethodCash
	if input.PaymentMethod != nil {
		method = *input.PaymentMethod
	}
	var received, change int64
	if input.ReceivedAmount != nil {
		received = *input.ReceivedAmount
	}
	if input.ChangeAmount != nil {
		change = *input.ChangeAmount
	}
	return []model.PaymentSplit{
		{
			ClinicID:       input.ClinicID,
			BillingID:      input.ID,
			Method:         method,
			Amount:         *input.BillingAmount,
			ReceivedAmount: received,
			ChangeAmount:   change,
			PaidBy:         input.StaffID,
		},
	}
}

// buildAccountingUpdate は UpdateAccountingInput から nil でないフィールドのみ抽出する。
func buildAccountingUpdate(input *UpdateAccountingInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.HospitalizationID != nil {
		fields["hospitalization_id"] = *input.HospitalizationID
	}
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.Subtotal != nil {
		fields["subtotal"] = *input.Subtotal
	}
	if input.TaxTotal != nil {
		fields["tax_total"] = *input.TaxTotal
	}
	if input.TotalAmount != nil {
		fields["total_amount"] = *input.TotalAmount
	}
	if input.HasInsurance != nil {
		fields["has_insurance"] = *input.HasInsurance
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.ScheduledDate != nil {
		fields["scheduled_date"] = *input.ScheduledDate
	}
	if input.CompletedAt != nil {
		fields["completed_at"] = *input.CompletedAt
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	return fields
}
