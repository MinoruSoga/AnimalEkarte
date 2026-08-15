package medicalrecord

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type CreateTreatmentPlanInput struct {
	TreatmentContent string
	Memo             string
	IsInsurance      bool
	UnitPrice        int64
	Quantity         float64
	DiscountRate     float64
	DiscountAmount   int64
	// Subtotal is ignored on write; server always recomputes (MRD-04).
	Subtotal  int64
	SortOrder int
}

type UpdateTreatmentPlanInput struct {
	TreatmentContent *string
	Memo             *string
	IsInsurance      *bool
	UnitPrice        *int64
	Quantity         *float64
	DiscountRate     *float64
	DiscountAmount   *int64
	// Subtotal is ignored on write; server always recomputes when price fields change (MRD-04).
	Subtotal  *int64
	SortOrder *int
	// DiscountEditAllowed is set by the HTTP boundary from discount:edit RBAC.
	// Service rechecks discount fields against the FOR UPDATE locked row (SEC-CS-F10).
	DiscountEditAllowed bool
}

// applyTreatmentPlanDiscountGuard rechecks discount against the locked plan row (SEC-CS-F10).
func applyTreatmentPlanDiscountGuard(locked *model.TreatmentPlan, input *UpdateTreatmentPlanInput) error {
	if input.DiscountRate != nil {
		if !httpapi.FloatEquals(*input.DiscountRate, locked.DiscountRate) && !input.DiscountEditAllowed {
			return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
		}
	}
	if input.DiscountAmount != nil {
		if *input.DiscountAmount != locked.DiscountAmount && !input.DiscountEditAllowed {
			return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
		}
	}
	return nil
}

// computeTreatmentPlanSubtotal is the single source of truth for plan subtotal (MRD-04).
func computeTreatmentPlanSubtotal(unitPrice int64, quantity, discountRate float64, discountAmount int64) int64 {
	return int64(float64(unitPrice)*quantity*(1-discountRate/100)) - discountAmount
}

func validateTreatmentPlanMoney(unitPrice int64, quantity, discountRate float64, discountAmount int64) error {
	if err := validateNonNegativePrice(&unitPrice); err != nil {
		return err
	}
	if quantity <= 0 {
		return apperrors.WrapInvalidInput(errMsgQuantityPositive)
	}
	if err := validateDiscountRate(discountRate); err != nil {
		return err
	}
	if discountAmount < 0 {
		return apperrors.WrapInvalidInput(errMsgPriceZeroOrMore)
	}
	return nil
}

func buildTreatmentPlanUpdate(input *UpdateTreatmentPlanInput) map[string]any {
	fields := map[string]any{}
	if input.TreatmentContent != nil {
		fields["treatment_content"] = *input.TreatmentContent
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.IsInsurance != nil {
		fields["is_insurance"] = *input.IsInsurance
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.DiscountRate != nil {
		fields["discount_rate"] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	// Subtotal is never taken from client input (MRD-04).
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

type TreatmentPlanService interface {
	ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	Create(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error)
	// medicalRecordID / hospitalizationID bind the plan to the URL parent resource (MRD-03).
	Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error)
	Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error
}

type treatmentPlanService struct {
	repo       TreatmentPlanRepository
	transactor Transactor
}

// NewTreatmentPlanService constructs TreatmentPlanService.
// transactor is required so Create/Update write+reload stay in one transaction (MRD-02 / X-01).
func NewTreatmentPlanService(repo TreatmentPlanRepository, transactor Transactor) TreatmentPlanService {
	return &treatmentPlanService{repo: repo, transactor: transactor}
}

func (s *treatmentPlanService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("treatment plan transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}

func (s *treatmentPlanService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	plan, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get treatment plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to get treatment plan")
	}
	return plan, nil
}

func (s *treatmentPlanService) ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	plans, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list treatment plans by medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to list treatment plans by medical record")
	}
	return plans, nil
}

func (s *treatmentPlanService) ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans, err := s.repo.FindByHospitalizationID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list treatment plans by hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to list treatment plans by hospitalization")
	}
	return plans, nil
}

// assertPlanParentMatch ensures plan belongs to the URL parent resource (MRD-03).
func assertPlanParentMatch(plan *model.TreatmentPlan, medicalRecordID, hospitalizationID *uint64) error {
	if medicalRecordID != nil {
		if plan.MedicalRecordID == nil || *plan.MedicalRecordID != *medicalRecordID {
			return apperrors.WrapNotFound("treatment_plan", "parent")
		}
	}
	if hospitalizationID != nil {
		if plan.HospitalizationID == nil || *plan.HospitalizationID != *hospitalizationID {
			return apperrors.WrapNotFound("treatment_plan", "parent")
		}
	}
	return nil
}

func (s *treatmentPlanService) Create(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	if err := validateTreatmentPlanMoney(input.UnitPrice, input.Quantity, input.DiscountRate, input.DiscountAmount); err != nil {
		return nil, err
	}
	subtotal := computeTreatmentPlanSubtotal(input.UnitPrice, input.Quantity, input.DiscountRate, input.DiscountAmount)

	plan := &model.TreatmentPlan{
		ClinicID:          clinicID,
		MedicalRecordID:   medicalRecordID,
		HospitalizationID: hospitalizationID,
		TreatmentContent:  input.TreatmentContent,
		Memo:              input.Memo,
		IsInsurance:       input.IsInsurance,
		UnitPrice:         input.UnitPrice,
		Quantity:          input.Quantity,
		DiscountRate:      input.DiscountRate,
		DiscountAmount:    input.DiscountAmount,
		Subtotal:          subtotal,
		SortOrder:         input.SortOrder,
	}

	// MRD-02: write + response re-fetch in one transaction (vital Update pattern).
	var result *model.TreatmentPlan
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, plan); err != nil {
			slog.ErrorContext(txCtx, "failed to create treatment plan", "error", err)
			return apperrors.Wrap(err, "failed to create treatment plan")
		}
		reloaded, err := s.repo.FindByID(txCtx, clinicID, plan.ID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to reload treatment plan after create", "error", err)
			return apperrors.Wrap(err, "failed to reload treatment plan after create")
		}
		result = reloaded
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "treatment plan created", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", plan.ID))
	return result, nil
}

func (s *treatmentPlanService) Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	// Hospitalization-nested plans are create-time snapshots (W-002): no PATCH after create.
	if hospitalizationID != nil {
		return nil, apperrors.WrapConflict("入院に紐づく治療プランは登録時スナップショットのため変更・削除できません")
	}
	// MRD-02 + MRD-03 + MRD-04: parent bind, money validation, write+reload in one tx.
	// SEC-CS-F10: lock plan row and recheck discount against the locked snapshot.
	var result *model.TreatmentPlan
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		existing, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to lock treatment plan", "error", err)
			return apperrors.Wrap(err, "failed to lock treatment plan")
		}
		if existing == nil {
			return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
		}
		if err := assertPlanParentMatch(existing, medicalRecordID, hospitalizationID); err != nil {
			return err
		}
		if err := applyTreatmentPlanDiscountGuard(existing, input); err != nil {
			return err
		}

		// Unprivileged equal-to-locked discount is a no-op: drop it so money merge cannot
		// race-write stale values. Differing values already failed the guard above.
		effective := *input
		if input.DiscountRate != nil && httpapi.FloatEquals(*input.DiscountRate, existing.DiscountRate) && !input.DiscountEditAllowed {
			effective.DiscountRate = nil
		}
		if input.DiscountAmount != nil && *input.DiscountAmount == existing.DiscountAmount && !input.DiscountEditAllowed {
			effective.DiscountAmount = nil
		}

		// Merge money fields for validation / subtotal recompute.
		unitPrice := existing.UnitPrice
		quantity := existing.Quantity
		discountRate := existing.DiscountRate
		discountAmount := existing.DiscountAmount
		moneyTouched := false
		if effective.UnitPrice != nil {
			unitPrice = *effective.UnitPrice
			moneyTouched = true
		}
		if effective.Quantity != nil {
			quantity = *effective.Quantity
			moneyTouched = true
		}
		if effective.DiscountRate != nil {
			discountRate = *effective.DiscountRate
			moneyTouched = true
		}
		if effective.DiscountAmount != nil {
			discountAmount = *effective.DiscountAmount
			moneyTouched = true
		}
		if moneyTouched {
			if err := validateTreatmentPlanMoney(unitPrice, quantity, discountRate, discountAmount); err != nil {
				return err
			}
		}

		fields := buildTreatmentPlanUpdate(&effective)
		if moneyTouched {
			fields["unit_price"] = unitPrice
			fields["quantity"] = quantity
			fields["discount_rate"] = discountRate
			fields["discount_amount"] = discountAmount
			fields["subtotal"] = computeTreatmentPlanSubtotal(unitPrice, quantity, discountRate, discountAmount)
		}
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput("at least one field must be provided")
		}
		if err := s.repo.Update(txCtx, clinicID, id, medicalRecordID, hospitalizationID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update treatment plan", "error", err)
			return apperrors.Wrap(err, "failed to update treatment plan")
		}
		reloaded, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to reload treatment plan after update", "error", err)
			return apperrors.Wrap(err, "failed to reload treatment plan after update")
		}
		result = reloaded
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "treatment plan updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", id))
	return result, nil
}

func (s *treatmentPlanService) Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
	// Hospitalization-nested plans are create-time snapshots (W-002): no DELETE after create.
	if hospitalizationID != nil {
		return apperrors.WrapConflict("入院に紐づく治療プランは登録時スナップショットのため変更・削除できません")
	}
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		existing, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to find treatment plan")
		}
		if err := assertPlanParentMatch(existing, medicalRecordID, hospitalizationID); err != nil {
			return err
		}
		if err := s.repo.Delete(txCtx, clinicID, id, medicalRecordID, hospitalizationID); err != nil {
			slog.ErrorContext(txCtx, "failed to delete treatment plan", "error", err, "clinic_id", clinicID, "treatment_plan_id", id)
			return apperrors.Wrap(err, "failed to delete treatment plan")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "treatment plan deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", id))
	return nil
}

// compile-time check
var _ TreatmentPlanService = (*treatmentPlanService)(nil)
