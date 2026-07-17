package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// VitalRepository はバイタル記録のデータアクセスインターフェース
type VitalRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error)
	Create(ctx context.Context, vital *model.VitalRecord) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vitalRepository struct {
	db *gorm.DB
}

// NewVitalRepository はVitalRepositoryを初期化して返す
func NewVitalRepository(db *gorm.DB) VitalRepository {
	return &vitalRepository{db: db}
}

func (r *vitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	vitals := make([]model.VitalRecord, 0)
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.medical_record_id = ? AND vital_records.deleted_at IS NULL", medicalRecordID).
		Order("vital_records.recorded_at ASC").
		Find(&vitals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vital", "")
	}
	return vitals, nil
}

func (r *vitalRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error) {
	var vital model.VitalRecord
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		First(&vital).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))
	}
	return &vital, nil
}

// Create は dbOrTx(ctx, r.db) で ambient tx に参加する（BE-refactor.md X-11）。
// LockByIDForUpdate の行ロック保持 tx 内から呼ばれた場合、別コネクションで INSERT すると
// vital_records.medical_record_id の FK 制約チェックが同一行への FOR UPDATE ロックと
// デッドロックする（FK チェックは FOR KEY SHARE を要求し FOR UPDATE と競合するため）。
func (r *vitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	if err := dbOrTx(ctx, r.db).Create(vital).Error; err != nil {
		return apperrors.FromGORM(err, "vital", "")
	}
	return nil
}

// Update は dbOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := dbOrTx(ctx, r.db).
		Model(&model.VitalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vital", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete は dbOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := dbOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ?", id).
		Delete(&model.VitalRecord{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vital", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}
