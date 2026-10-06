package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LabMappingDraftStore persists drafts, confirmed maps, and the generation switch.
type LabMappingDraftStore interface {
	GenerationStopped(ctx context.Context, clinicID uint64) (bool, error)
	SetGenerationStopped(ctx context.Context, clinicID uint64, stopped bool) error
	InsertDraft(ctx context.Context, draft *model.LabDeviceMappingDraft, items []model.LabDeviceMappingDraftItem) error
	GetDraft(ctx context.Context, clinicID, id uint64) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error)
	LockDraft(ctx context.Context, clinicID, id uint64) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error)
	MarkConfirmed(ctx context.Context, clinicID, id, staffID uint64, at time.Time) error
	UpdateConfirmedItems(ctx context.Context, clinicID, draftID uint64, items []model.LabDeviceMappingDraftItem) error
	UpsertDeviceMaster(ctx context.Context, row model.LabDeviceItemMaster) error
	UpsertSpeciesSpec(ctx context.Context, spec model.LabDeviceMappingSpec) error
	FindExamField(ctx context.Context, clinicID, fieldID uint64) (*model.ExamTypeField, error)
	ListFieldCandidates(ctx context.Context, clinicID uint64) ([]model.ExamTypeField, error)
}

type labMappingDraftRepository struct{ db *gorm.DB }

// NewLabMappingDraftRepository initializes a LabMappingDraftStore.
func NewLabMappingDraftRepository(db *gorm.DB) LabMappingDraftStore {
	return &labMappingDraftRepository{db: db}
}

func (r *labMappingDraftRepository) GenerationStopped(ctx context.Context, clinicID uint64) (bool, error) {
	var row model.LabDeviceMappingGeneration
	err := persistence.DBOrTx(ctx, r.db).Where("clinic_id = ?", clinicID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, apperrors.FromGORM(err, "lab_device_mapping_generation", fmt.Sprintf("%d", clinicID))
	}
	return row.Stopped, nil
}

func (r *labMappingDraftRepository) SetGenerationStopped(ctx context.Context, clinicID uint64, stopped bool) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	row := model.LabDeviceMappingGeneration{ClinicID: clinicID, Stopped: stopped, UpdatedAt: time.Now().UTC()}
	err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"stopped", "updated_at"}),
	}).Create(&row).Error
	if err != nil {
		return apperrors.FromGORM(err, "lab_device_mapping_generation", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *labMappingDraftRepository) InsertDraft(
	ctx context.Context,
	draft *model.LabDeviceMappingDraft,
	items []model.LabDeviceMappingDraftItem,
) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	db := tx.WithContext(ctx)
	if err := db.Create(draft).Error; err != nil {
		return apperrors.FromGORM(err, "lab_device_mapping_draft", "")
	}
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].DraftID = draft.ID
		items[i].ClinicID = draft.ClinicID
	}
	if err := db.Create(&items).Error; err != nil {
		return apperrors.FromGORM(err, "lab_device_mapping_draft_item", "")
	}
	return nil
}

func (r *labMappingDraftRepository) GetDraft(
	ctx context.Context,
	clinicID, id uint64,
) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	return r.loadDraft(persistence.DBOrTx(ctx, r.db), clinicID, id, false)
}

func (r *labMappingDraftRepository) LockDraft(
	ctx context.Context,
	clinicID, id uint64,
) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, nil, apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	return r.loadDraft(tx.WithContext(ctx), clinicID, id, true)
}

func (r *labMappingDraftRepository) loadDraft(
	db *gorm.DB,
	clinicID, id uint64,
	lock bool,
) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	var draft model.LabDeviceMappingDraft
	q := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&draft).Error; err != nil {
		return nil, nil, apperrors.FromGORM(err, "lab_device_mapping_draft", fmt.Sprintf("%d", id))
	}
	items := make([]model.LabDeviceMappingDraftItem, 0)
	err := db.Scopes(persistence.ClinicScope(clinicID)).
		Where("draft_id = ?", id).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, nil, apperrors.FromGORM(err, "lab_device_mapping_draft_item", fmt.Sprintf("%d", id))
	}
	return &draft, items, nil
}

func (r *labMappingDraftRepository) MarkConfirmed(ctx context.Context, clinicID, id, staffID uint64, at time.Time) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	result := tx.WithContext(ctx).Model(&model.LabDeviceMappingDraft{}).
		Where("clinic_id = ? AND id = ? AND status = ?", clinicID, id, model.LabMappingDraftStatusUnconfirmed).
		Updates(map[string]any{
			"status":                model.LabMappingDraftStatusConfirmed,
			"confirmed_at":          at,
			"confirmed_by_staff_id": staffID,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lab_device_mapping_draft", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict("mapping draft is already confirmed")
	}
	return nil
}

func (r *labMappingDraftRepository) UpdateConfirmedItems(
	ctx context.Context,
	clinicID, draftID uint64,
	items []model.LabDeviceMappingDraftItem,
) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	db := tx.WithContext(ctx)
	for i := range items {
		result := db.Model(&model.LabDeviceMappingDraftItem{}).
			Where("clinic_id = ? AND draft_id = ? AND device_item_code = ?", clinicID, draftID, items[i].DeviceItemCode).
			Updates(map[string]any{
				"exam_type_field_id":  items[i].ExamTypeFieldID,
				"chart_unit":          items[i].ChartUnit,
				"proposed_field_name": items[i].ProposedFieldName,
				"confirmed":           true,
			})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "lab_device_mapping_draft_item", items[i].DeviceItemCode)
		}
		if result.RowsAffected == 0 {
			return apperrors.WrapNotFound("lab_device_mapping_draft_item", items[i].DeviceItemCode)
		}
	}
	return nil
}

func (r *labMappingDraftRepository) UpsertDeviceMaster(ctx context.Context, row model.LabDeviceItemMaster) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	row.IsActive = true
	err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "clinic_id"}, {Name: "source_type"}, {Name: "device_item_code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"unit", "value_shape", "exam_type_field_id", "is_active", "updated_at",
		}),
	}).Create(&row).Error
	if err != nil {
		return apperrors.FromGORM(err, "lab_device_item_master", row.DeviceItemCode)
	}
	return nil
}

func (r *labMappingDraftRepository) UpsertSpeciesSpec(ctx context.Context, spec model.LabDeviceMappingSpec) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("lab mapping write requires an ambient transaction")
	}
	err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "clinic_id"}, {Name: "source_type"}, {Name: "species"}, {Name: "device_item_code"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"unit", "value_shape", "exam_type_field_id", "confirmed_draft_id", "updated_at",
		}),
	}).Create(&spec).Error
	if err != nil {
		return apperrors.FromGORM(err, "lab_device_mapping_spec", spec.DeviceItemCode)
	}
	return nil
}

func (r *labMappingDraftRepository) FindExamField(ctx context.Context, clinicID, fieldID uint64) (*model.ExamTypeField, error) {
	var field model.ExamTypeField
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id = ?", clinicID, fieldID).
		First(&field).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	return &field, nil
}

func (r *labMappingDraftRepository) ListFieldCandidates(ctx context.Context, clinicID uint64) ([]model.ExamTypeField, error) {
	rows := make([]model.ExamTypeField, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("name ASC, id ASC").
		Limit(labMappingCandidateCap).
		Find(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", "")
	}
	return rows, nil
}
