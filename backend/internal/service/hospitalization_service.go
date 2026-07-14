package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// DischargeWithBillingInput は退院+会計作成の入力DTO
type DischargeWithBillingInput struct {
	DischargeDate    time.Time
	CreateAccounting bool
}

// DischargeWithBillingResult は退院+会計作成のレスポンスDTO
type DischargeWithBillingResult struct {
	HospitalizationID uint64
	AccountingID      *uint64
	Status            string
}

// CreateHospitalizationInput は入院作成の入力DTO
type CreateHospitalizationInput struct {
	OwnerID              uint64
	PetID                uint64
	HospitalizationType  model.HospitalizationType
	StartDate            time.Time
	EndDate              time.Time
	Status               model.HospitalizationStatus
	CageID               *uint64
	DoctorID             *uint64
	Memo                 string
	OwnerRequest         string
	StaffNotes           string
	IsInsurance          bool
	InsuranceCompanyName *string
	InsuranceNumber      *string
}

// UpdateHospitalizationInput は入院更新のサービス入力 DTO
type UpdateHospitalizationInput struct {
	OwnerID              *uint64
	PetID                *uint64
	HospitalizationType  *model.HospitalizationType
	StartDate            *time.Time
	EndDate              *time.Time
	Status               *model.HospitalizationStatus
	CageID               *uint64
	DoctorID             *uint64
	Memo                 *string
	OwnerRequest         *string
	StaffNotes           *string
	IsInsurance          *bool
	InsuranceCompanyName *string
	InsuranceNumber      *string
}

// resolveFinalHospitalizationOwnerPet は PATCH 入力と現在値から最終 Owner/Pet を求める（AUD-004）。
func resolveFinalHospitalizationOwnerPet(existing *model.Hospitalization, input *UpdateHospitalizationInput) (ownerID, petID *uint64) {
	o, p := existing.OwnerID, existing.PetID
	ownerID, petID = &o, &p
	if input.OwnerID != nil {
		ownerID = input.OwnerID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	return ownerID, petID
}

func buildHospitalizationUpdate(input *UpdateHospitalizationInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.HospitalizationType != nil {
		fields["hospitalization_type"] = *input.HospitalizationType
	}
	if input.StartDate != nil {
		fields["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		fields["end_date"] = *input.EndDate
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.CageID != nil {
		fields["cage_id"] = *input.CageID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.OwnerRequest != nil {
		fields["owner_request"] = *input.OwnerRequest
	}
	if input.StaffNotes != nil {
		fields["staff_notes"] = *input.StaffNotes
	}
	if input.IsInsurance != nil {
		if !*input.IsInsurance {
			// 保険なしに切り替えた場合は保険情報を NULL にする
			fields["insurance_company_name"] = nil
			fields["insurance_number"] = nil
		} else {
			if input.InsuranceCompanyName != nil {
				fields["insurance_company_name"] = *input.InsuranceCompanyName
			}
			if input.InsuranceNumber != nil {
				fields["insurance_number"] = *input.InsuranceNumber
			}
		}
	} else {
		// IsInsurance が nil でも保険フィールド単体の更新は許容する
		if input.InsuranceCompanyName != nil {
			fields["insurance_company_name"] = *input.InsuranceCompanyName
		}
		if input.InsuranceNumber != nil {
			fields["insurance_number"] = *input.InsuranceNumber
		}
	}
	return fields
}

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error)
}

type hospitalizationService struct {
	repos *repository.Repositories
}

func NewHospitalizationService(repos *repository.Repositories) HospitalizationService {
	return &hospitalizationService{repos: repos}
}

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	result, total, err := s.repos.Hospitalization.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list hospitalizations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list hospitalizations")
	}
	return result, total, nil
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	result, err := s.repos.Hospitalization.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	return result, nil
}

func (s *hospitalizationService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
	status := input.Status
	if status == "" {
		status = model.HospitalizationStatusReserved
	}
	// is_insurance == false の場合は保険情報を NULL にする
	var insuranceCompanyName *string
	var insuranceNumber *string
	if input.IsInsurance {
		insuranceCompanyName = input.InsuranceCompanyName
		insuranceNumber = input.InsuranceNumber
	}

	// クロステナント write 防止: request 由来 Owner/Pet の clinic 所有と Owner-Pet 整合を検証する（AUD-004）。
	ownerID, petID := input.OwnerID, input.PetID
	if err := validateReservationOwnerPetLinks(ctx, s.repos.Reservation, clinicID, &ownerID, &petID); err != nil {
		return nil, err
	}

	// クロステナント write 防止: request 由来の cage マスタが caller の clinic に属することを検証する。
	if err := validateOwnedMasterFK(ctx, "cage", clinicID, input.CageID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Cage.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return nil, err
	}

	hospitalization := &model.Hospitalization{
		ClinicID:             clinicID,
		OwnerID:              input.OwnerID,
		PetID:                input.PetID,
		HospitalizationType:  input.HospitalizationType,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		Status:               status,
		CageID:               input.CageID,
		DoctorID:             input.DoctorID,
		Memo:                 input.Memo,
		OwnerRequest:         input.OwnerRequest,
		StaffNotes:           input.StaffNotes,
		InsuranceCompanyName: insuranceCompanyName,
		InsuranceNumber:      insuranceNumber,
	}
	if err := s.repos.Hospitalization.Create(ctx, hospitalization); err != nil {
		slog.ErrorContext(ctx, "failed to create hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to create hospitalization")
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID))
	return hospitalization, nil
}

func (s *hospitalizationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	existing, err := s.repos.Hospitalization.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to find hospitalization")
	}

	// クロステナント write 防止: 最終マージ後の Owner/Pet を検証する（AUD-004）。
	if input.OwnerID != nil || input.PetID != nil {
		finalOwnerID, finalPetID := resolveFinalHospitalizationOwnerPet(existing, input)
		if err := validateReservationOwnerPetLinks(ctx, s.repos.Reservation, clinicID, finalOwnerID, finalPetID); err != nil {
			return nil, err
		}
	}

	// クロステナント write 防止: 貼り替え先 cage マスタの所有権を検証する。
	if err := validateOwnedMasterFK(ctx, "cage", clinicID, input.CageID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Cage.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return nil, err
	}

	fields := buildHospitalizationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	hosp, err := s.repos.Hospitalization.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to update hospitalization")
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))
	return hosp, nil
}
func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repos.Hospitalization.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find hospitalization")
	}
	// FK依存チェック: 入院に紐付く日次記録が存在する場合は削除を拒否
	dailyCount, err := s.repos.Hospitalization.CountDailyRecordsByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check daily record dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check daily record dependencies")
	}
	if dailyCount > 0 {
		return apperrors.WrapConflict("日次記録が紐付いているため削除できません。先に日次記録を削除してください")
	}

	// FK依存チェック: 入院に紐付く治療計画が存在する場合は削除を拒否
	planCount, err := s.repos.Hospitalization.CountTreatmentPlansByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check treatment plan dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check treatment plan dependencies")
	}
	if planCount > 0 {
		return apperrors.WrapConflict("治療計画が紐付いているため削除できません。先に治療計画を削除してください")
	}

	// FK依存チェック: 入院に紐付くケアプラン項目が存在する場合は削除を拒否
	itemCount, err := s.repos.Hospitalization.CountCarePlanItemsByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check care plan item dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check care plan item dependencies")
	}
	if itemCount > 0 {
		return apperrors.WrapConflict("ケアプランが紐付いているため削除できません。先にケアプランを削除してください")
	}

	if err := s.repos.Hospitalization.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete hospitalization", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete hospitalization")
	}

	slog.InfoContext(ctx, "hospitalization deleted",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}

// DischargeWithBilling は退院処理を行い、オプションで会計レコードを自動生成する。
// care_plan_items を billing_items に変換してトランザクション内で原子的に実行する。
func (s *hospitalizationService) DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
	// 入院レコード取得
	hosp, err := s.repos.Hospitalization.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	if hosp.Status == model.HospitalizationStatusDischarged {
		return nil, apperrors.WrapInvalidInput("hospitalization is already discharged")
	}

	result := &DischargeWithBillingResult{
		HospitalizationID: id,
		Status:            string(model.HospitalizationStatusDischarged),
	}

	err = s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		// 0. TOCTOU 対策: tx 内で再取得し、locked の Owner/Pet で検証・会計作成する（AUD-004 Q2-A）。
		locked, err := txRepos.Hospitalization.FindByID(ctx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get hospitalization in transaction")
		}
		if locked.Status == model.HospitalizationStatusDischarged {
			return apperrors.WrapInvalidInput("hospitalization is already discharged")
		}

		// 1. 汚染行対策: CreateAccounting 有無に関わらず、Update 前に Owner/Pet の clinic 所有を再検証する（AUD-004）。
		if err := validateReservationOwnerPetLinks(ctx, txRepos.Reservation, clinicID, &locked.OwnerID, &locked.PetID); err != nil {
			return err
		}

		// 2. 退院ステータスに更新
		dischargedStatus := model.HospitalizationStatusDischarged
		dischargeFields := map[string]any{
			"status":   dischargedStatus,
			"end_date": input.DischargeDate,
		}
		if _, err := txRepos.Hospitalization.Update(ctx, clinicID, id, dischargeFields); err != nil {
			return apperrors.Wrap(err, "failed to discharge hospitalization")
		}

		if !input.CreateAccounting {
			return nil
		}

		// 2. ケアプラン取得
		carePlanItems, err := txRepos.CarePlanItem.FindByHospitalizationID(ctx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get care plan items")
		}

		// 3. 会計レコード作成
		billing := &model.Billing{
			ClinicID:          clinicID,
			HospitalizationID: &id,
			PetID:             &locked.PetID,
			OwnerID:           &locked.OwnerID,
			Status:            model.BillingStatusWaiting,
			ScheduledDate:     input.DischargeDate,
		}
		if err := txRepos.Accounting.Create(ctx, clinicID, billing); err != nil {
			return apperrors.Wrap(err, "failed to create billing")
		}

		// 4. ケアプラン → 会計明細に変換
		var totalAmount int64
		for i := range carePlanItems {
			item := &carePlanItems[i]
			billingItem := &model.BillingItem{
				BillingID: billing.ID,
				Category:  model.ItemCategoryOther,
				Name:      item.Name,
				UnitPrice: item.UnitPrice,
				Quantity:  1.0,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   DefaultTaxRate,
				Source:    model.ItemSourceHospitalization,
				SortOrder: i,
			}
			if err := txRepos.BillingItem.Create(ctx, billingItem); err != nil {
				return apperrors.Wrap(err, "failed to create billing item")
			}
			totalAmount += item.UnitPrice
		}

		// 5. 合計金額更新
		if len(carePlanItems) > 0 {
			taxTotal := int64(float64(totalAmount) * DefaultTaxRate)
			if err := txRepos.BillingItem.UpdateBillingTotals(ctx, clinicID, billing.ID, totalAmount, taxTotal, totalAmount+taxTotal); err != nil {
				return apperrors.Wrap(err, "failed to update billing totals")
			}
		}

		result.AccountingID = &billing.ID
		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to discharge hospitalization with billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to discharge hospitalization with billing")
	}

	slog.InfoContext(ctx, "hospitalization discharged",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("create_accounting", input.CreateAccounting))

	return result, nil
}
