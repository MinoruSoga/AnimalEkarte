package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// PaymentMethodSystemKeys は payment_method ENUM と payment_methods マスタの system_key を対応づける（#197）。
// system_key は rename 耐性のある安定識別子（migration 009 で追加）。
// name（表示専用）への依存を廃止し、クリニックが既定 name を改名しても会計確定・集計が正常動作する。
var PaymentMethodSystemKeys = map[model.PaymentMethod]string{
	model.PaymentMethodCash:            "cash",
	model.PaymentMethodCreditCard:      "credit_card",
	model.PaymentMethodElectronicMoney: "electronic_money",
	model.PaymentMethodBankTransfer:    "bank_transfer",
}

// resolvePaymentMethodMasterID は method(ENUM) を当該 clinic の payment_methods マスタ id に解決する（#128/#197）。
// 真実の源泉は検証済みの method(ENUM)。systemKeyToID は当該 clinic の payment_methods マスタ system_key→id マップ。
// existing（リクエストが明示供給した payment_method_id。ハンドラで無検証バインドされる）が非 nil の場合は、
// 当該 clinic の当該 method に解決される id と一致することを検証し、不一致なら拒否する
// （他クリニックの master id や method と矛盾する id の混入を防ぐ＝マルチテナント隔離）。
// 対応するマスタ行が無い場合は明示エラーを返す（payment_method_id=NULL のまま保存して集計で現金に倒す挙動を防ぐ）。
func resolvePaymentMethodMasterID(method model.PaymentMethod, existing *uint64, systemKeyToID map[string]uint64) (*uint64, error) {
	key, ok := PaymentMethodSystemKeys[method]
	if !ok {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("未知の支払方法のためマスタ解決ができません: %s", method))
	}
	id, ok := systemKeyToID[key]
	if !ok {
		// クリニックに既定支払方法マスタが存在しない設定不整合。誤って現金扱いせず、会計確定を止める。
		return nil, apperrors.WrapInternalServerError(fmt.Sprintf("支払方法マスタ (system_key=%s) が見つかりません", key))
	}
	// 明示供給された id は当該 clinic の当該 method マスタと一致する場合のみ許可する。
	if existing != nil && *existing != id {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("指定された支払方法 id (%d) は当該クリニックの支払方法 (system_key=%s) と一致しません", *existing, key))
	}
	resolved := id
	return &resolved, nil
}

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
// 優先順位は仕様として固定: cash > credit_card > bank_transfer > electronic_money（#198 で bank_transfer を明示追加）。
func representativeMethod(splits []PaymentSplitInput) model.PaymentMethod {
	for _, p := range []model.PaymentMethod{
		model.PaymentMethodCash,
		model.PaymentMethodCreditCard,
		model.PaymentMethodBankTransfer,
		model.PaymentMethodElectronicMoney,
	} {
		for _, s := range splits {
			if s.Method == p {
				return p
			}
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
			if err := validateCashSplit(s); err != nil {
				return err
			}
		}
	}
	if billingAmount != nil && total != *billingAmount {
		return apperrors.WrapInvalidInput(fmt.Sprintf("支払い内訳の合計（%d）が請求金額（%d）と一致しません", total, *billingAmount))
	}
	return nil
}

func validateCashSplit(s PaymentSplitInput) error {
	if s.ReceivedAmount < s.Amount {
		return apperrors.WrapInvalidInput("現金の預り金が不足しています")
	}
	if s.ChangeOverride {
		// #188: お釣り直接上書きモード。レジ実機（カルテ非連動）の誤差を会計記録に反映するため
		// change == received - amount の整合検証を意図的に緩和する。下限ガード（change >= 0）は維持する。
		if s.ChangeAmount < 0 {
			return apperrors.WrapInvalidInput("お釣りは0円以上でなければなりません")
		}
		return nil
	}
	if s.ChangeAmount != s.ReceivedAmount-s.Amount {
		return apperrors.WrapInvalidInput("お釣り計算が不正です")
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
		// P10: status = completed 時は server time で completed_at を上書き（client clock skew 防止）
		if *input.Status == model.BillingStatusCompleted {
			now := time.Now()
			fields["completed_at"] = now
		}
	} else if input.CompletedAt != nil {
		// status 変更なし、かつ completed_at 送信時のみ受け取る（completed 状態での再設定は許可しない）
		fields["completed_at"] = *input.CompletedAt
	}
	if input.ScheduledDate != nil {
		fields["scheduled_date"] = *input.ScheduledDate
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	return fields
}

// validateAccountingRelatedFKs は会計の関連 FK（medical_record / hospitalization / owner / pet）の
// clinic 所有と相互整合を検証する（AUD-002）。nil 関連は既存契約どおり許可する。
// Owner/Pet は reservation.ValidateReservationOwnerPetLinksWithRepo（AUD-001）を再利用する。
func (s *accountingService) validateAccountingRelatedFKs(
	ctx context.Context,
	clinicID uint64,
	medicalRecordID, hospitalizationID, ownerID, petID *uint64,
) error {
	var mr *model.MedicalRecord
	if medicalRecordID != nil {
		if s.medicalRecordRepo == nil {
			return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", *medicalRecordID))
		}
		record, err := s.medicalRecordRepo.FindByID(ctx, clinicID, *medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify medical record ownership")
		}
		mr = record
	}

	var hosp *model.Hospitalization
	if hospitalizationID != nil {
		if s.hospRepo == nil {
			return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", *hospitalizationID))
		}
		h, err := s.hospRepo.FindByID(ctx, clinicID, *hospitalizationID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify hospitalization ownership")
		}
		hosp = h
	}

	if s.reservationRepo != nil {
		if err := reservation.ValidateReservationOwnerPetLinksWithRepo(ctx, s.reservationRepo, clinicID, ownerID, petID); err != nil {
			return err
		}
	}

	if mr != nil {
		if err := AssertBillingLinksMatchMedicalRecord(mr, ownerID, petID); err != nil {
			return err
		}
	}
	if hosp != nil {
		if err := assertBillingLinksMatchHospitalization(hosp, ownerID, petID); err != nil {
			return err
		}
	}
	return nil
}

// AssertBillingLinksMatchMedicalRecord は AUD-005 系の請求リンク整合ガード
// （B②で分離→B④で原本がここへ帰還・estimate 側も本関数を消費）。
func AssertBillingLinksMatchMedicalRecord(mr *model.MedicalRecord, ownerID, petID *uint64) error {
	if ownerID != nil && mr.OwnerID != nil && *ownerID != *mr.OwnerID {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", mr.ID))
	}
	if petID != nil && mr.PetID != nil && *petID != *mr.PetID {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", mr.ID))
	}
	return nil
}

func assertBillingLinksMatchHospitalization(h *model.Hospitalization, ownerID, petID *uint64) error {
	if ownerID != nil && *ownerID != h.OwnerID {
		return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", h.ID))
	}
	if petID != nil && *petID != h.PetID {
		return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", h.ID))
	}
	return nil
}

// resolveFinalAccountingRelatedIDs は Update の最終関連 FK 集合を既存値と input から合成する（AUD-002）。
func resolveFinalAccountingRelatedIDs(existing *model.Billing, input *UpdateAccountingInput) (medicalRecordID, hospitalizationID, ownerID, petID *uint64) {
	medicalRecordID = existing.MedicalRecordID
	hospitalizationID = existing.HospitalizationID
	ownerID = existing.OwnerID
	petID = existing.PetID
	if input.MedicalRecordID != nil {
		medicalRecordID = input.MedicalRecordID
	}
	if input.HospitalizationID != nil {
		hospitalizationID = input.HospitalizationID
	}
	if input.OwnerID != nil {
		ownerID = input.OwnerID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	return medicalRecordID, hospitalizationID, ownerID, petID
}
