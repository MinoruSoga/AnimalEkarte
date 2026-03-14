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
	UnitPrice      float64
	Quantity       int
	Selected       bool
	Status         string
	Content        string
	Memo           string
	Insurance      bool
	DiscountRate   float64
	DiscountAmount float64
	SortOrder      int
}

// UpdateTreatmentInput は治療項目更新の入力DTO（ポインタ型 = nil は未送信）
type UpdateTreatmentInput struct {
	ItemType       *model.TreatmentItemType
	ConsultationID *uint64
	ProcedureID    *uint64
	MedicineID     *uint64
	InventoryID    *uint64
	UnitPrice      *float64
	Quantity       *int
	Selected       *bool
	Status         *string
	Content        *string
	Memo           *string
	Insurance      *bool
	DiscountRate   *float64
	DiscountAmount *float64
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
	List(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error)
	Update(ctx context.Context, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error)
	Delete(ctx context.Context, medicalRecordID, treatmentID uint64) error
	BulkUpdateSortOrder(ctx context.Context, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type treatmentService struct {
	repo repository.TreatmentRepository
}

// NewTreatmentService はTreatmentServiceを初期化して返す
func NewTreatmentService(repo repository.TreatmentRepository) TreatmentService {
	return &treatmentService{repo: repo}
}

func (s *treatmentService) List(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error) {
	return s.repo.ListByMedicalRecordID(ctx, medicalRecordID)
}

func (s *treatmentService) Create(ctx context.Context, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error) {
	if err := validateTreatmentItemType(input.ItemType); err != nil {
		return nil, err
	}

	status := model.TreatmentStatusIncomplete
	if input.Status != "" {
		s, err := parseTreatmentStatus(input.Status)
		if err != nil {
			return nil, err
		}
		status = s
	}

	treatment := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ItemType:        input.ItemType,
		ConsultationID:  input.ConsultationID,
		ProcedureID:     input.ProcedureID,
		MedicineID:      input.MedicineID,
		InventoryID:     input.InventoryID,
		UnitPrice:       input.UnitPrice,
		Quantity:        input.Quantity,
		Selected:        input.Selected,
		Status:          status,
		Content:         input.Content,
		Memo:            input.Memo,
		Insurance:       input.Insurance,
		DiscountRate:    input.DiscountRate,
		DiscountAmount:  input.DiscountAmount,
		SortOrder:       input.SortOrder,
	}

	if err := s.repo.Create(ctx, treatment); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "treatment created",
		slog.Uint64("treatment_id", treatment.ID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return treatment, nil
}

func (s *treatmentService) Update(ctx context.Context, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
	// 所属確認
	existing, err := s.repo.FindByID(ctx, treatmentID)
	if err != nil {
		return nil, err
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

	fields := buildTreatmentUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, treatmentID, fields); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "treatment updated",
		slog.Uint64("treatment_id", treatmentID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return s.repo.FindByID(ctx, treatmentID)
}

func (s *treatmentService) Delete(ctx context.Context, medicalRecordID, treatmentID uint64) error {
	existing, err := s.repo.FindByID(ctx, treatmentID)
	if err != nil {
		return err
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	if err := s.repo.Delete(ctx, treatmentID); err != nil {
		return err
	}

	slog.InfoContext(ctx, "treatment deleted",
		slog.Uint64("treatment_id", treatmentID),
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

	if err := s.repo.BulkUpdateSortOrder(ctx, updates); err != nil {
		return err
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
	if input.Selected != nil {
		fields["selected"] = *input.Selected
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
	if input.Insurance != nil {
		fields["insurance"] = *input.Insurance
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
	case model.TreatmentStatusIncomplete,
		model.TreatmentStatusComplete,
		model.TreatmentStatusNA:
		return model.TreatmentStatus(s), nil
	}
	return "", apperrors.WrapInvalidInput("invalid status: " + s)
}
