package billing

import (
	"context"
	"fmt"

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

// hasPaymentFields は UpdateAccountingInput に Payment 関連フィールドが含まれているか判定する。
func hasPaymentFields(input *UpdateAccountingInput) bool {
	return len(input.PaymentSplits) > 0 ||
		input.PaymentMethod != nil ||
		input.InsuranceRatio != nil ||
		input.InsuranceName != nil ||
		input.InsuranceAmount != nil ||
		input.BillingAmount != nil ||
		input.ReceivedAmount != nil ||
		input.ChangeAmount != nil ||
		input.DiscountAmount != nil
}

func accountingUpdateFromInput(input *UpdateAccountingInput) AccountingUpdate {
	return AccountingUpdate{
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		ScheduledDate:     input.ScheduledDate,
		Memo:              input.Memo,
	}
}

func (cmd AccountingUpdate) toFields() map[string]any {
	fields := make(map[string]any)
	if cmd.MedicalRecordID != nil {
		fields["medical_record_id"] = *cmd.MedicalRecordID
	}
	if cmd.HospitalizationID != nil {
		fields["hospitalization_id"] = *cmd.HospitalizationID
	}
	if cmd.OwnerID != nil {
		fields["owner_id"] = *cmd.OwnerID
	}
	if cmd.PetID != nil {
		fields["pet_id"] = *cmd.PetID
	}
	if cmd.ScheduledDate != nil {
		fields["scheduled_date"] = *cmd.ScheduledDate
	}
	if cmd.Memo != nil {
		fields["memo"] = *cmd.Memo
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
