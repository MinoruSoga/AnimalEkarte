package medicalrecord

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colConsultationName           = "name"
	colConsultationPrice          = "price"
	colConsultationIsActive       = "is_active"
	colConsultationDescription    = "description"
	colConsultationTimeCondition  = "time_condition"
	colConsultationDuration       = "duration"
	colConsultationParentID       = "parent_id"
	colConsultationSortOrder      = "sort_order"
	colConsultationTaxType        = "tax_type"
	colConsultationTaxRate        = "tax_rate"
	colConsultationIsNonInsurance = "is_non_insurance"
)

func buildConsultationUpdate(input *UpdateConsultationInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colConsultationName] = *input.Name
	}
	if input.Price != nil {
		fields[colConsultationPrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colConsultationIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colConsultationDescription] = *input.Description
	}
	if input.TimeCondition != nil {
		fields[colConsultationTimeCondition] = *input.TimeCondition
	}
	if input.Duration != nil {
		fields[colConsultationDuration] = *input.Duration
	}
	setNullableUint64Field(fields, colConsultationParentID, input.ClearParentID, input.ParentID)
	if input.SortOrder != nil {
		fields[colConsultationSortOrder] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields[colConsultationTaxType] = model.TaxType(*input.TaxType)
	}
	if input.TaxRate != nil {
		fields[colConsultationTaxRate] = *input.TaxRate
	}
	if input.IsNonInsurance != nil {
		fields[colConsultationIsNonInsurance] = *input.IsNonInsurance
	}
	setBillDiscountExcludedField(fields, input.IsBillDiscountExcluded)
	return fields
}

// ---- ConsultationService ----

// CreateConsultationInput は診察種別作成のサービス入力 DTO
type CreateConsultationInput struct {
	ActorID       *uint64
	Name          string
	Price         *int64
	IsActive      bool
	Description   string
	TimeCondition string
	Duration      *int
	ParentID      *uint64
	SortOrder     int
	TaxType       *string  // nil = "excluded" (default)
	TaxRate       *float64 // nil = 0.10 (default)

	// IsNonInsurance は品目別保険設定。nil = 新規登録の初期値（保険対象外=true）。
	IsNonInsurance *bool
	// IsBillDiscountExcluded は全体割引除外。nil = 新規登録の初期値 false を明示保存（EMR-447 A01）。
	IsBillDiscountExcluded *bool
}

// UpdateConsultationInput は診察料金更新のサービス入力 DTO
type UpdateConsultationInput struct {
	Version       int64 // Required editing-start version; never replaced with a freshly read version.
	ActorID       *uint64
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	TimeCondition *string
	Duration      *int
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
	TaxType       *string
	TaxRate       *float64

	// IsNonInsurance は nil = 変更なし（DB の未保存/旧欠落 NULL を書き換えない）。
	IsNonInsurance *bool
	// IsBillDiscountExcluded は nil = 変更なし（EMR-447 A01）。
	IsBillDiscountExcluded *bool
}

type ConsultationService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type consultationService struct {
	insuranceWrites MasterInsuranceWriteDependencies
	repo            ConsultationRepository
}

func NewConsultationService(repo ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

// NewConsultationServiceWithAudit enables atomic item-setting writes and durable audit.
func NewConsultationServiceWithAudit(repo ConsultationRepository, deps MasterInsuranceWriteDependencies) ConsultationService {
	return &consultationService{repo: repo, insuranceWrites: deps}
}

func (s *consultationService) List(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list consultations")
	}
	return items, nil
}
func (s *consultationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get consultation")
	}
	return result, nil
}
func (s *consultationService) Create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	var result *model.Consultation
	err := s.insuranceWrites.write(ctx, clinicID, input.ActorID, "consultation", "consultation.create", func(txCtx context.Context) (masterInsuranceChange, error) {
		if err := lockMasterWriteRows(txCtx, s.repo, clinicID, 0, input.ParentID); err != nil {
			return masterInsuranceChange{}, err
		}
		var err error
		result, err = s.create(txCtx, clinicID, input)
		if err != nil {
			return masterInsuranceChange{}, err
		}
		return masterInsuranceChange{
			ID:           result.ID,
			After:        result.IsNonInsurance,
			BillDiscount: &billDiscountChange{After: result.IsBillDiscountExcluded},
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *consultationService) create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	taxType := model.TaxTypeExcluded
	if input.TaxType != nil && *input.TaxType != "" {
		taxType = model.TaxType(*input.TaxType)
	}
	taxRate := DefaultTaxRate
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	// EMR-454: 新規登録の初期値は保険対象外（true）。DB に DEFAULT を持たせず
	// API 層で明示保存する。明示指定（false=保険対象）は上書きしない。
	isNonInsurance := true
	if input.IsNonInsurance != nil {
		isNonInsurance = *input.IsNonInsurance
	}
	var exactTaxRate model.ExactNumeric
	if err := exactTaxRate.Scan(strconv.FormatFloat(taxRate, 'f', -1, 64)); err != nil {
		return nil, apperrors.WrapInvalidInput("invalid tax rate")
	}
	sortOrder := input.SortOrder
	consultation := &model.Consultation{
		ClinicID:      clinicID,
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		TimeCondition: input.TimeCondition,
		Duration:      input.Duration,
		ParentID:      input.ParentID,
		SortOrder:     &sortOrder,
		TaxType:       taxType,
		TaxRate:       exactTaxRate,

		IsNonInsurance:         &isNonInsurance,
		IsBillDiscountExcluded: createBillDiscountExcluded(input.IsBillDiscountExcluded),
	}
	if err := s.repo.Create(ctx, consultation); err != nil {
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Name,
			apperrors.ConstraintConsultationName,
			apperrors.CodeConsultationNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to create consultation")
	}
	slog.InfoContext(ctx, "consultation created", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", consultation.ID))
	return s.repo.FindByID(ctx, clinicID, consultation.ID)
}
func (s *consultationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if input.Version < 1 {
		return nil, apperrors.WrapInvalidInput("positive editing-start version is required")
	}
	var result *model.Consultation
	err := s.insuranceWrites.write(ctx, clinicID, input.ActorID, "consultation", "consultation.update", func(txCtx context.Context) (masterInsuranceChange, error) {
		if err := lockMasterWriteRows(txCtx, s.repo, clinicID, id, input.ParentID); err != nil {
			return masterInsuranceChange{}, err
		}
		before, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return masterInsuranceChange{}, err
		}
		var oldSetting *bool
		if before.IsNonInsurance != nil {
			value := *before.IsNonInsurance
			oldSetting = &value
		}
		result, err = s.update(txCtx, clinicID, id, input)
		if err != nil {
			return masterInsuranceChange{}, err
		}
		var billDiscount *billDiscountChange
		if input.IsBillDiscountExcluded != nil {
			billDiscount = &billDiscountChange{BeforeExists: true, Before: before.IsBillDiscountExcluded, After: result.IsBillDiscountExcluded}
		}
		return masterInsuranceChange{ID: result.ID, BeforeExists: true, Before: oldSetting, After: result.IsNonInsurance, SkipInsuranceAudit: input.IsNonInsurance == nil, BillDiscount: billDiscount}, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *consultationService) update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get consultation")
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	fields := buildConsultationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	consultation, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintConsultationName,
			apperrors.CodeConsultationNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to update consultation")
	}
	slog.InfoContext(ctx, "consultation updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", id))
	return consultation, nil
}
func (s *consultationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if s.insuranceWrites.Transactor == nil {
		return apperrors.WrapInternalServerError("consultation write transaction dependency is required")
	}
	if err := s.insuranceWrites.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := lockMasterWriteRows(txCtx, s.repo, clinicID, id, nil); err != nil {
			return err
		}
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to find consultation")
		}
		childCount, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check consultation children")
		}
		if childCount > 0 {
			return apperrors.WrapConflict("この診察項目にはサブ項目が登録されているため削除できません")
		}
		count, err := s.repo.CountUsageByConsultationID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check consultation dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この診察項目はカルテで使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete consultation")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "consultation deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", id))
	return nil
}

func (s *consultationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder consultations")
	}
	slog.InfoContext(ctx, "consultations reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *consultationService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent consultation", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}
