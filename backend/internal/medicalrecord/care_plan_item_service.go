package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateCarePlanItemInput はケアプランアイテム作成のサービス入力DTO
type CreateCarePlanItemInput struct {
	Type                  string
	Name                  string
	Description           string
	Timing                []string
	Status                string
	Notes                 string
	MedicineID            *uint64
	ProcedureID           *uint64
	HospitalizationPlanID *uint64
	UnitPrice             int64
	Category              string
	SortOrder             int
}

// UpdateCarePlanItemInput はケアプランアイテム更新のサービス入力DTO
type UpdateCarePlanItemInput struct {
	Type                  *string
	Name                  *string
	Description           *string
	Timing                []string // nil = not changed, empty slice = clear
	Status                *string
	Notes                 *string
	MedicineID            *uint64
	ProcedureID           *uint64
	HospitalizationPlanID *uint64
	UnitPrice             *int64
	Category              *string
	SortOrder             *int
}

func buildCarePlanItemUpdate(input *UpdateCarePlanItemInput) map[string]any {
	fields := map[string]any{}
	if input.Type != nil {
		fields["type"] = *input.Type
	}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Timing != nil {
		fields["timing"] = pq.StringArray(input.Timing)
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	if input.MedicineID != nil {
		fields["medicine_id"] = *input.MedicineID
	}
	if input.ProcedureID != nil {
		fields["procedure_id"] = *input.ProcedureID
	}
	if input.HospitalizationPlanID != nil {
		fields["hospitalization_plan_id"] = *input.HospitalizationPlanID
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Category != nil {
		fields["category"] = *input.Category
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

// CarePlanItemService はケアプランアイテムのビジネスロジックインターフェース
type CarePlanItemService interface {
	List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	Create(ctx context.Context, clinicID, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error)
	Update(ctx context.Context, clinicID, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error)
	Delete(ctx context.Context, clinicID, hospitalizationID, itemID uint64) error
}

type carePlanItemService struct {
	repo          CarePlanItemRepository
	hospRepo      HospitalizationRepository
	medicineRepo  medicineFinder
	procedureRepo procedureFinder
	hospPlanRepo  HospitalizationPlanRepository
	transactor    Transactor
	auditTx       AuditTxLogger
}

// NewCarePlanItemService は CarePlanItemService を初期化して返す。
// transactor is required for Create/Update write+reload atomicity (MRA-02).
// auditTx is required so hard-delete leaves a fail-closed audit trail (MRA-01).
func NewCarePlanItemService(
	repo CarePlanItemRepository,
	hospRepo HospitalizationRepository,
	medicineRepo medicineFinder,
	procedureRepo procedureFinder,
	hospPlanRepo HospitalizationPlanRepository,
	transactor Transactor,
	auditTx AuditTxLogger,
) CarePlanItemService {
	return &carePlanItemService{
		repo:          repo,
		hospRepo:      hospRepo,
		medicineRepo:  medicineRepo,
		procedureRepo: procedureRepo,
		hospPlanRepo:  hospPlanRepo,
		transactor:    transactor,
		auditTx:       auditTx,
	}
}

func (s *carePlanItemService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("care plan item transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}

func carePlanItemAuditValue(item *model.CarePlanItem) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":                      item.ID,
		"hospitalization_id":      item.HospitalizationID,
		"type":                    string(item.Type),
		"name":                    item.Name,
		"description":             item.Description,
		"timing":                  []string(item.Timing),
		"status":                  string(item.Status),
		"notes":                   item.Notes,
		"medicine_id":             item.MedicineID,
		"procedure_id":            item.ProcedureID,
		"hospitalization_plan_id": item.HospitalizationPlanID,
		"unit_price":              item.UnitPrice,
		"category":                item.Category,
		"sort_order":              item.SortOrder,
	}
}

func (s *carePlanItemService) auditDeleteTx(ctx context.Context, clinicID uint64, item *model.CarePlanItem) error {
	// MRA-01: hard-delete without durable recovery requires fail-closed audit.
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("care plan item audit dependency is required")
	}
	resourceID := item.ID
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorType:  auditActorTypeFor(nil),
		Action:     model.AuditActionCarePlanItemDelete,
		Resource:   model.AuditResourceCarePlanItem,
		ResourceID: &resourceID,
		OldValue:   carePlanItemAuditValue(item),
		Metadata: map[string]any{
			"hospitalization_id": item.HospitalizationID,
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		return apperrors.Wrap(err, "failed to audit care plan item delete")
	}
	return nil
}

// validateMasterFKs は request 由来の clinic-scoped マスタFK (medicine/procedure/hospitalization_plan)
// の所有権を検証する。別 clinic のマスタ参照は NotFound で遮断し入院ケアの cross-tenant mislink を防ぐ。
func (s *carePlanItemService) validateMasterFKs(ctx context.Context, clinicID uint64, medicineID, procedureID, hospitalizationPlanID *uint64) error {
	if err := validateOwnedMasterFK(ctx, "medicine", clinicID, medicineID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.medicineRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "procedure", clinicID, procedureID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.procedureRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	return validateOwnedMasterFK(ctx, "hospitalization plan", clinicID, hospitalizationPlanID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.hospPlanRepo.FindByID(actx, cid, mid)
			return err
		})
}

func (s *carePlanItemService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	items, err := s.repo.FindByHospitalizationID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list care plan items", "error", err)
		return nil, apperrors.Wrap(err, "failed to list care plan items")
	}
	return items, nil
}

func (s *carePlanItemService) Create(ctx context.Context, clinicID, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error) {
	planType := model.CarePlanType(input.Type)
	if err := validateCarePlanType(planType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate care plan type")
	}

	status := model.CarePlanStatusActive
	if input.Status != "" {
		planStatus := model.CarePlanStatus(input.Status)
		if err := validateCarePlanStatus(planStatus); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate care plan status")
		}
		status = planStatus
	}

	// テナント所有権検証（クロステナント write 防止）。
	// care_plan_items は自前 clinic_id を持たず hospitalizations 経由で隔離するため、
	// 親入院の所有権を Create でも明示検証する（repo.Create は clinicScope できない）。
	if _, err := s.hospRepo.FindByID(ctx, clinicID, hospitalizationID); err != nil {
		slog.ErrorContext(ctx, "failed to verify hospitalization ownership", "error", err)
		return nil, apperrors.Wrap(err, "failed to verify hospitalization ownership")
	}

	// クロステナント write 防止: medicine/procedure/hospitalization_plan マスタが caller の clinic に属することを検証する。
	if err := s.validateMasterFKs(ctx, clinicID, input.MedicineID, input.ProcedureID, input.HospitalizationPlanID); err != nil {
		return nil, err
	}

	item := &model.CarePlanItem{
		HospitalizationID:     hospitalizationID,
		Type:                  planType,
		Name:                  input.Name,
		Description:           input.Description,
		Timing:                pq.StringArray(input.Timing),
		Status:                status,
		Notes:                 input.Notes,
		MedicineID:            input.MedicineID,
		ProcedureID:           input.ProcedureID,
		HospitalizationPlanID: input.HospitalizationPlanID,
		UnitPrice:             input.UnitPrice,
		Category:              input.Category,
		SortOrder:             input.SortOrder,
	}
	// MRA-02: write + response re-fetch before commit.
	var created *model.CarePlanItem
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, item); err != nil {
			slog.ErrorContext(txCtx, "failed to create care plan item", "error", err)
			return apperrors.Wrap(err, "failed to create care plan item")
		}
		reloaded, err := s.repo.FindByID(txCtx, clinicID, item.ID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get created care plan item", "error", err)
			return apperrors.Wrap(err, "failed to get created care plan item")
		}
		created = reloaded
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "care plan item created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", item.ID))
	return created, nil
}

func (s *carePlanItemService) Update(ctx context.Context, clinicID, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error) {
	// Verify item belongs to this clinic + hospitalization
	existing, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get care plan item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get care plan item")
	}
	if existing.HospitalizationID != hospitalizationID {
		return nil, apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", itemID))
	}

	if input.Type != nil {
		if err := validateCarePlanType(model.CarePlanType(*input.Type)); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate care plan type")
		}
	}
	if input.Status != nil {
		if err := validateCarePlanStatus(model.CarePlanStatus(*input.Status)); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate care plan status")
		}
	}

	// クロステナント write 防止: 貼り替え先 medicine/procedure/hospitalization_plan マスタの所有権を検証する。
	if err := s.validateMasterFKs(ctx, clinicID, input.MedicineID, input.ProcedureID, input.HospitalizationPlanID); err != nil {
		return nil, err
	}

	fields := buildCarePlanItemUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	// MRA-02: write + response re-fetch before commit.
	var updated *model.CarePlanItem
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Update(txCtx, clinicID, itemID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update care plan item", "error", err)
			return apperrors.Wrap(err, "failed to update care plan item")
		}
		reloaded, err := s.repo.FindByID(txCtx, clinicID, itemID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get updated care plan item", "error", err)
			return apperrors.Wrap(err, "failed to get updated care plan item")
		}
		updated = reloaded
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "care plan item updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", itemID))
	return updated, nil
}

func (s *carePlanItemService) Delete(ctx context.Context, clinicID, hospitalizationID, itemID uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get care plan item")
	}
	if existing.HospitalizationID != hospitalizationID {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", itemID))
	}

	// MRA-01: hard-delete + fail-closed audit in one transaction.
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, itemID); err != nil {
			slog.ErrorContext(txCtx, "failed to delete care plan item", "error", err, "id", itemID, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete care plan item")
		}
		return s.auditDeleteTx(txCtx, clinicID, existing)
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "care plan item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", itemID))

	return nil
}

func validateCarePlanType(t model.CarePlanType) error {
	switch t {
	case model.CarePlanTypeFood, model.CarePlanTypeMedicine, model.CarePlanTypeTreatment,
		model.CarePlanTypeInstruction, model.CarePlanTypeItem:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid care plan type: %s", t))
	}
}

func validateCarePlanStatus(s model.CarePlanStatus) error {
	switch s {
	case model.CarePlanStatusActive, model.CarePlanStatusCompleted, model.CarePlanStatusDiscontinued:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid care plan status: %s", s))
	}
}
