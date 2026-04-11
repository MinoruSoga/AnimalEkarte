package service

import (
	"context"
	"log/slog"
	"strconv"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ─── Input DTOs ───────────────────────────────────────────────────────────────

// CreateTreatmentInput は治療項目作成の入力DTO（HTTPを知らない）
type CreateTreatmentInput struct {
	ItemType       model.TreatmentItemType
	ConsultationID *uint64
	ProcedureID    *uint64
	MedicineID     *uint64
	InventoryID    *uint64
	UnitPrice      int64
	Quantity       float64
	IsSelected      bool
	Status         string
	Content        string
	Memo           string
	AdminRoute     string
	IsInsurance     bool
	DiscountRate   float64
	DiscountAmount int64
	SortOrder      int
}

// UpdateTreatmentInput は治療項目更新の入力DTO（ポインタ型 = nil は未送信）
type UpdateTreatmentInput struct {
	ItemType       *model.TreatmentItemType
	ConsultationID *uint64
	ProcedureID    *uint64
	MedicineID     *uint64
	InventoryID    *uint64
	UnitPrice      *int64
	Quantity       *float64
	IsSelected      *bool
	Status         *string
	Content        *string
	Memo           *string
	AdminRoute     *string
	IsInsurance     *bool
	DiscountRate   *float64
	DiscountAmount *int64
	SortOrder      *int
}

// BulkUpdateTreatmentsInput は並び順一括更新の入力DTO
type BulkUpdateTreatmentsInput struct {
	Treatments []BulkTreatmentItem
}

// BulkTreatmentItem は一括更新の個別項目
type BulkTreatmentItem struct {
	ID        uint64
	SortOrder int
}

// ─── Interface ────────────────────────────────────────────────────────────────

// TreatmentService は治療項目のビジネスロジックインターフェース
type TreatmentService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error)
	Update(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error
	BulkUpdateSortOrder(ctx context.Context, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type treatmentService struct {
	repos *repository.Repositories
}

// NewTreatmentService はTreatmentServiceを初期化して返す
func NewTreatmentService(repos *repository.Repositories) TreatmentService {
	return &treatmentService{
		repos: repos,
	}
}

func (s *treatmentService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	treatments, err := s.repos.Treatment.ListByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list treatments")
	}
	return treatments, nil
}

func (s *treatmentService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error) {
	if err := validateTreatmentItemType(input.ItemType); err != nil {
		return nil, err
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
	}
	if input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("数量は0より大きい値を入力してください")
	}
	if input.DiscountRate < 0 || input.DiscountRate > 100 {
		return nil, apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}

	status := model.TreatmentStatusPending
	if input.Status != "" {
		s, err := parseTreatmentStatus(input.Status)
		if err != nil {
			return nil, err
		}
		status = s
	}

	var treatment *model.Treatment

	// ─── Transaction ───
	err := s.repos.Transaction(func(txRepos *repository.Repositories) error {
		treatment = &model.Treatment{
			MedicalRecordID: medicalRecordID,
			ItemType:        input.ItemType,
			ConsultationID:  input.ConsultationID,
			ProcedureID:     input.ProcedureID,
			MedicineID:      input.MedicineID,
			InventoryID:     input.InventoryID,
			UnitPrice:       input.UnitPrice,
			Quantity:        input.Quantity,
			IsSelected:      input.IsSelected,
			Status:          status,
			Content:         input.Content,
			Memo:            input.Memo,
			AdminRoute:      input.AdminRoute,
			IsInsurance:     input.IsInsurance,
			DiscountRate:    input.DiscountRate,
			DiscountAmount:  input.DiscountAmount,
			SortOrder:       input.SortOrder,
		}

		// 1. Create Treatment
		if err := txRepos.Treatment.Create(ctx, treatment); err != nil {
			return apperrors.Wrap(err, "failed to create treatment")
		}

		// 2. Decrease Stock (if applicable)
		if input.MedicineID != nil || input.InventoryID != nil {
			var targetInvID uint64
			if input.InventoryID != nil {
				targetInvID = *input.InventoryID
			} else {
				targetInvID = *input.MedicineID
			}

			if targetInvID > 0 {
				if err := txRepos.Inventory.DecreaseStock(ctx, targetInvID, input.Quantity); err != nil {
					return apperrors.Wrap(err, "failed to decrease inventory stock")
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create treatment")
	}

	slog.InfoContext(ctx, "treatment created with atomic inventory sync",
		slog.Uint64("treatment_id", treatment.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return treatment, nil
}

func (s *treatmentService) Update(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
	// 所属確認（clinic_id + id で検索）
	existing, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	if input.ItemType != nil {
		if err := validateTreatmentItemType(*input.ItemType); err != nil {
			return nil, err
		}
	}
	if input.Status != nil {
		if _, err := parseTreatmentStatus(*input.Status); err != nil {
			return nil, err
		}
	}
	if input.Quantity != nil && *input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("quantity must be greater than 0")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
	}
	if input.DiscountRate != nil && (*input.DiscountRate < 0 || *input.DiscountRate > 100) {
		return nil, apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}

	fields := buildTreatmentUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repos.Treatment.Update(ctx, clinicID, treatmentID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update treatment")
	}

	slog.InfoContext(ctx, "treatment updated",
		slog.Uint64("treatment_id", treatmentID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	treatment, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated treatment")
	}
	return treatment, nil
}

func (s *treatmentService) Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error {
	existing, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	if err := s.repos.Treatment.Delete(ctx, clinicID, treatmentID); err != nil {
		return apperrors.Wrap(err, "failed to delete treatment")
	}

	slog.InfoContext(ctx, "treatment deleted",
		slog.Uint64("treatment_id", treatmentID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return nil
}

func (s *treatmentService) BulkUpdateSortOrder(ctx context.Context, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
	updates := make([]repository.TreatmentSortUpdate, 0, len(input.Treatments))
	for _, item := range input.Treatments {
		updates = append(updates, repository.TreatmentSortUpdate{
			ID:        item.ID,
			SortOrder: item.SortOrder,
		})
	}

	if err := s.repos.Treatment.BulkUpdateSortOrder(ctx, updates); err != nil {
		return apperrors.Wrap(err, "failed to bulk update treatment sort order")
	}

	slog.InfoContext(ctx, "treatments bulk sort_order updated",
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Int("count", len(updates)))

	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// buildTreatmentUpdateFields は非nilポインタフィールドだけをGORM向けmapに変換する。
// GORMのzero-value問題（false/0/"" がスキップされる）を回避するために使用する。
func buildTreatmentUpdateFields(input *UpdateTreatmentInput) map[string]any {
	fields := map[string]any{}
	if input.ItemType != nil {
		fields["item_type"] = *input.ItemType
	}
	if input.ConsultationID != nil {
		fields["consultation_id"] = *input.ConsultationID
	}
	if input.ProcedureID != nil {
		fields["procedure_id"] = *input.ProcedureID
	}
	if input.MedicineID != nil {
		fields["medicine_id"] = *input.MedicineID
	}
	if input.InventoryID != nil {
		fields["inventory_id"] = *input.InventoryID
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.IsSelected != nil {
		fields["is_selected"] = *input.IsSelected
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Content != nil {
		fields["content"] = *input.Content
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.AdminRoute != nil {
		fields["admin_route"] = *input.AdminRoute
	}
	if input.IsInsurance != nil {
		fields["is_insurance"] = *input.IsInsurance
	}
	if input.DiscountRate != nil {
		fields["discount_rate"] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

func validateTreatmentItemType(t model.TreatmentItemType) error {
	switch t {
	case model.TreatmentItemTypeConsultation,
		model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine,
		model.TreatmentItemTypeOther:
		return nil
	}
	return apperrors.WrapInvalidInput("invalid item_type: " + string(t))
}

func parseTreatmentStatus(s string) (model.TreatmentStatus, error) {
	switch model.TreatmentStatus(s) {
	case model.TreatmentStatusPending,
		model.TreatmentStatusCompleted,
		model.TreatmentStatusNotApplicable:
		return model.TreatmentStatus(s), nil
	}
	return "", apperrors.WrapInvalidInput("invalid status: " + s)
}
