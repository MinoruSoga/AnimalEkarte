package billing

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func applyExamBillingProvenance(
	ctx context.Context,
	repo BillingItemRepository,
	input *CreateBillingItemInput,
	item *model.BillingItem,
) error {
	if input == nil || input.ExamID == nil {
		return nil
	}
	if input.MerchandiseItemID != nil ||
		input.TreatmentID != nil ||
		input.AppointmentID != nil ||
		input.TrimmingCourseID != nil ||
		input.TrimmingOptionID != nil {
		return invalidBillingItemReferenceCombination()
	}
	if err := repo.ValidateExamCreateReference(ctx, input.ClinicID, input.BillingID, *input.ExamID); err != nil {
		return err
	}
	values, err := examBillingValuesFor(ctx, repo, input.ClinicID, *input.ExamID)
	if err != nil {
		return err
	}
	item.Category = model.ItemCategoryTest
	item.Source = model.ItemSourceMedicalRecord
	item.ExamID = input.ExamID
	item.Name = values.Name
	item.UnitPrice = values.UnitPrice
	clinicID := input.ClinicID
	item.ClinicID = &clinicID
	return nil
}

type examBillingValues struct {
	Name      string
	UnitPrice int64
}

type examBillingValuesProvider interface {
	ExamBillingValues(ctx context.Context, clinicID, examID uint64) (*examBillingValues, error)
}

func examBillingValuesFor(ctx context.Context, repo BillingItemRepository, clinicID, examID uint64) (*examBillingValues, error) {
	provider, ok := repo.(examBillingValuesProvider)
	if !ok {
		return nil, apperrors.WrapInternalServerError("exam billing values provider is required")
	}
	return provider.ExamBillingValues(ctx, clinicID, examID)
}

func (r *billingItemRepository) ValidateExamCreateReference(
	ctx context.Context,
	clinicID, billingID, examID uint64,
) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("exam billing validation requires an active transaction")
	}
	tx = tx.WithContext(ctx)

	var billingRef struct {
		MedicalRecordID *uint64
		PetID           *uint64
		Status          model.BillingStatus
	}
	if err := tx.
		Table("billings").
		Select("medical_record_id", "pet_id", "status").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", billingID, clinicID).
		Take(&billingRef).Error; err != nil {
		return apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	if billingRef.Status == model.BillingStatusCompleted ||
		billingRef.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("確定済みまたは取消済みの会計には検査を追加できません")
	}
	if billingRef.PetID == nil {
		return invalidBillingItemReferenceCombination()
	}

	var examRef struct {
		MedicalRecordID *uint64
		PetID           *uint64
		ExamTypeID      uint64
	}
	if err := tx.
		Table("exams").
		Select("medical_record_id", "pet_id", "exam_type_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examID, clinicID).
		Take(&examRef).Error; err != nil {
		return apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examID))
	}
	if examRef.PetID == nil || *examRef.PetID != *billingRef.PetID || examRef.MedicalRecordID == nil {
		return invalidBillingItemReferenceCombination()
	}
	if billingRef.MedicalRecordID != nil && *billingRef.MedicalRecordID != *examRef.MedicalRecordID {
		return invalidBillingItemReferenceCombination()
	}

	var confirmationStatus string
	if err := tx.
		Table("billing_confirmations").
		Select("status").
		Where("medical_record_id = ?", *examRef.MedicalRecordID).
		Take(&confirmationStatus).Error; err != nil || confirmationStatus != string(model.ConfirmationStatusConfirmed) {
		return apperrors.WrapConflict("会計確認前の検査は請求できません")
	}

	var examTypeRef struct {
		Name  string
		Price *int64
	}
	if err := tx.
		Table("exam_types").
		Select("name", "price").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examRef.ExamTypeID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Take(&examTypeRef).Error; err != nil {
		return apperrors.FromGORM(err, "exam_type", fmt.Sprintf("%d", examRef.ExamTypeID))
	}
	if strings.TrimSpace(examTypeRef.Name) == "" || examTypeRef.Price == nil || *examTypeRef.Price < 0 {
		return apperrors.WrapInternalServerError("exam type master is not billable")
	}

	var existingCount int64
	if err := tx.
		Table("billing_items AS bi").
		Where("bi.exam_id = ? AND bi.clinic_id = ? AND bi.deleted_at IS NULL", examID, clinicID).
		Count(&existingCount).Error; err != nil {
		return apperrors.FromGORM(err, "billing_item", fmt.Sprintf("exam:%d", examID))
	}
	if existingCount > 0 {
		return apperrors.WrapConflict("この検査は既に会計明細へ取り込まれています")
	}
	return nil
}

// ExamBillingValues reads canonical master values in the transaction that
// validated the exam reference; request name and price are never authoritative.
func (r *billingItemRepository) ExamBillingValues(ctx context.Context, clinicID, examID uint64) (*examBillingValues, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("exam billing values require an active transaction")
	}
	var values examBillingValues
	if err := tx.Table("exams").
		Select("exam_types.name", "exam_types.price AS unit_price").
		Joins("JOIN exam_types ON exam_types.id = exams.exam_type_id AND exam_types.clinic_id = exams.clinic_id AND exam_types.deleted_at IS NULL").
		Where("exams.id = ? AND exams.clinic_id = ? AND exams.deleted_at IS NULL", examID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Take(&values).Error; err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examID))
	}
	if strings.TrimSpace(values.Name) == "" || values.UnitPrice < 0 {
		return nil, apperrors.WrapInternalServerError("exam type master is not billable")
	}
	return &values, nil
}
