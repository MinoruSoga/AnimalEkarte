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

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
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
	return s.repos.Hospitalization.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return s.repos.Hospitalization.FindByID(ctx, clinicID, id)
}

func (s *hospitalizationService) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	if err := s.repos.Hospitalization.Create(ctx, hospitalization); err != nil {
		return apperrors.Wrap(err, "failed to create hospitalization")
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID))
	return nil
}

func (s *hospitalizationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildHospitalizationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	hosp, err := s.repos.Hospitalization.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update hospitalization")
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))
	return hosp, nil
}

// UpdateHospitalizationInput は入院更新のサービス入力 DTO
type UpdateHospitalizationInput struct {
	OwnerID             *uint64
	PetID               *uint64
	HospitalizationType *model.HospitalizationType
	StartDate           *time.Time
	EndDate             *time.Time
	Status              *model.HospitalizationStatus
	CageID              *uint64
	DoctorID            *uint64
	Memo                *string
	OwnerRequest        *string
	StaffNotes          *string
}

func buildHospitalizationUpdateFields(input *UpdateHospitalizationInput) map[string]any {
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
	return fields
}

func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repos.Hospitalization.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete hospitalization")
	}
	return nil
}

// DischargeWithBilling は退院処理を行い、オプションで会計レコードを自動生成する。
// care_plan_items を billing_items に変換してトランザクション内で原子的に実行する。
func (s *hospitalizationService) DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
	// 入院レコード取得
	hosp, err := s.repos.Hospitalization.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	if hosp.Status == model.HospitalizationStatusDischarged {
		return nil, apperrors.WrapInvalidInput("hospitalization is already discharged")
	}

	result := &DischargeWithBillingResult{
		HospitalizationID: id,
		Status:            string(model.HospitalizationStatusDischarged),
	}

	err = s.repos.Transaction(func(txRepos *repository.Repositories) error {
		// 1. 退院ステータスに更新
		dischargedStatus := model.HospitalizationStatusDischarged
		dischargeFields := map[string]any{
			"status":   dischargedStatus,
			"end_date": input.DischargeDate,
		}
		if _, err := txRepos.Hospitalization.UpdateFields(ctx, clinicID, id, dischargeFields); err != nil {
			return apperrors.Wrap(err, "failed to discharge hospitalization")
		}

		if !input.CreateAccounting {
			return nil
		}

		// 2. ケアプラン取得
		carePlanItems, err := txRepos.CarePlanItem.ListByHospitalizationID(ctx, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get care plan items")
		}

		// 3. 会計レコード作成
		billing := &model.Billing{
			ClinicID:          clinicID,
			HospitalizationID: &id,
			PetID:             &hosp.PetID,
			OwnerID:           &hosp.OwnerID,
			Status:            model.BillingStatusWaiting,
			ScheduledDate:     input.DischargeDate,
		}
		if err := txRepos.Accounting.Create(ctx, clinicID, billing); err != nil {
			return apperrors.Wrap(err, "failed to create billing")
		}

		// 4. ケアプラン → 会計明細に変換
		var totalAmount int64
		for i, item := range carePlanItems {
			billingItem := &model.BillingItem{
				BillingID: billing.ID,
				Category:  model.ItemCategoryOther,
				Name:      item.Name,
				UnitPrice: item.UnitPrice,
				Quantity:  1.0,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   0.10,
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
			taxTotal := int64(float64(totalAmount) * 0.10)
			if err := txRepos.BillingItem.UpdateBillingTotals(ctx, billing.ID, totalAmount, taxTotal, totalAmount+taxTotal); err != nil {
				return apperrors.Wrap(err, "failed to update billing totals")
			}
		}

		result.AccountingID = &billing.ID
		return nil
	})

	if err != nil {
		return nil, apperrors.Wrap(err, "failed to discharge hospitalization with billing")
	}

	slog.InfoContext(ctx, "hospitalization discharged",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("create_accounting", input.CreateAccounting))

	return result, nil
}
