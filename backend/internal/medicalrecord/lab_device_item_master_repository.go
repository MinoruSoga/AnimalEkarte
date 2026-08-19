package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LabDeviceItemMasterRepository persists clinic-scoped device item mappings.
type LabDeviceItemMasterRepository interface {
	List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.LabDeviceItemMaster, error)
	FindByClinicSourceCodes(ctx context.Context, clinicID uint64, sourceType string, codes []string) ([]model.LabDeviceItemMaster, error)
	EnsureCatalog(ctx context.Context, rows []model.LabDeviceItemMaster) (int64, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.LabDeviceItemMaster, error)
	FindExamTypeField(ctx context.Context, clinicID, fieldID uint64) (*model.ExamTypeField, error)
	FindExamTypeFields(ctx context.Context, clinicID uint64, fieldIDs []uint64) (map[uint64]model.ExamTypeField, error)
	FindExamType(ctx context.Context, clinicID, examTypeID uint64) (*model.ExaminationType, error)
	ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error)
	FindDeviceByID(ctx context.Context, clinicID, id uint64) (*model.LabDevice, error)
	CreateDevice(ctx context.Context, row *model.LabDevice) error
	UpdateDevice(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.LabDevice, error)
	EnsureDevices(ctx context.Context, rows []model.LabDevice) (int64, error)
}

type labDeviceItemMasterRepository struct{ db *gorm.DB }

// NewLabDeviceItemMasterRepository initializes a LabDeviceItemMasterRepository.
func NewLabDeviceItemMasterRepository(db *gorm.DB) LabDeviceItemMasterRepository {
	return &labDeviceItemMasterRepository{db: db}
}

func labDeviceMasterOrder(q *gorm.DB) *gorm.DB {
	return q.Order(`
CASE source_type
  WHEN 'fuji_nx600' THEN 1
  WHEN 'fuji_au10v' THEN 2
  WHEN 'arkray_pu4010' THEN 3
  ELSE 9
END, sort_order ASC, device_item_code ASC`)
}

func (r *labDeviceItemMasterRepository) List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error) {
	rows := make([]model.LabDeviceItemMaster, 0)
	q := persistence.DBOrTx(ctx, r.db).Model(&model.LabDeviceItemMaster{}).
		Scopes(persistence.ClinicScope(clinicID))
	if sourceType != "" {
		q = q.Where("source_type = ?", sourceType)
	}
	if err := labDeviceMasterOrder(q).Limit(persistence.MaxMasterListRows).Find(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lab_device_item_master", "")
	}
	return rows, nil
}

func (r *labDeviceItemMasterRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.LabDeviceItemMaster, error) {
	var row model.LabDeviceItemMaster
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_device_item_master", fmt.Sprintf("%d", id))
	}
	return &row, nil
}

func (r *labDeviceItemMasterRepository) FindByClinicSourceCodes(
	ctx context.Context,
	clinicID uint64,
	sourceType string,
	codes []string,
) ([]model.LabDeviceItemMaster, error) {
	rows := make([]model.LabDeviceItemMaster, 0)
	if len(codes) == 0 {
		return rows, nil
	}
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("source_type = ?", sourceType).
		Where("device_item_code IN ?", codes).
		Limit(persistence.MaxMasterListRows).
		Find(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_device_item_master", "")
	}
	return rows, nil
}

func (r *labDeviceItemMasterRepository) EnsureCatalog(ctx context.Context, rows []model.LabDeviceItemMaster) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := persistence.DBOrTx(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "clinic_id"},
			{Name: "source_type"},
			{Name: "device_item_code"},
		},
		DoNothing: true,
	}).Create(&rows)
	if result.Error != nil {
		return 0, apperrors.FromGORM(result.Error, "lab_device_item_master", "")
	}
	return result.RowsAffected, nil
}

func (r *labDeviceItemMasterRepository) Update(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
) (*model.LabDeviceItemMaster, error) {
	if err := persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
		&model.LabDeviceItemMaster{},
		"lab_device_item_master",
		clinicID,
		id,
		fields,
	); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *labDeviceItemMasterRepository) FindExamTypeField(
	ctx context.Context,
	clinicID, fieldID uint64,
) (*model.ExamTypeField, error) {
	var field model.ExamTypeField
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id = ?", clinicID, fieldID).
		First(&field).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	return &field, nil
}

func (r *labDeviceItemMasterRepository) FindExamTypeFields(
	ctx context.Context,
	clinicID uint64,
	fieldIDs []uint64,
) (map[uint64]model.ExamTypeField, error) {
	out := make(map[uint64]model.ExamTypeField, len(fieldIDs))
	if len(fieldIDs) == 0 {
		return out, nil
	}
	var fields []model.ExamTypeField
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ?", clinicID).
		Where("id IN ?", fieldIDs).
		Limit(persistence.MaxMasterListRows).
		Find(&fields).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", "")
	}
	for i := range fields {
		out[fields[i].ID] = fields[i]
	}
	return out, nil
}

func (r *labDeviceItemMasterRepository) FindExamType(
	ctx context.Context,
	clinicID, examTypeID uint64,
) (*model.ExaminationType, error) {
	var row model.ExaminationType
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", examTypeID).
		First(&row).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type", fmt.Sprintf("%d", examTypeID))
	}
	return &row, nil
}

func (r *labDeviceItemMasterRepository) ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error) {
	rows := make([]model.LabDevice, 0)
	err := persistence.DBOrTx(ctx, r.db).Model(&model.LabDevice{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, id ASC").
		Limit(persistence.MaxMasterListRows).
		Find(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_device", "")
	}
	return rows, nil
}

func (r *labDeviceItemMasterRepository) FindDeviceByID(ctx context.Context, clinicID, id uint64) (*model.LabDevice, error) {
	var row model.LabDevice
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_device", fmt.Sprintf("%d", id))
	}
	return &row, nil
}

func (r *labDeviceItemMasterRepository) CreateDevice(ctx context.Context, row *model.LabDevice) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(row).Error; err != nil {
		return apperrors.FromGORM(err, "lab_device", "")
	}
	return nil
}

func (r *labDeviceItemMasterRepository) UpdateDevice(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
) (*model.LabDevice, error) {
	if err := persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
		&model.LabDevice{},
		"lab_device",
		clinicID,
		id,
		fields,
	); err != nil {
		return nil, err
	}
	return r.FindDeviceByID(ctx, clinicID, id)
}

func (r *labDeviceItemMasterRepository) EnsureDevices(ctx context.Context, rows []model.LabDevice) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := persistence.DBOrTx(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "clinic_id"},
			{Name: "source_type"},
		},
		DoNothing: true,
	}).Create(&rows)
	if result.Error != nil {
		return 0, apperrors.FromGORM(result.Error, "lab_device", "")
	}
	return result.RowsAffected, nil
}
