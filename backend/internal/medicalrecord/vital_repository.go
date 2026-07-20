package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// VitalRepository はバイタル記録のデータアクセスインターフェース。
// Moved from internal/repository (BE9-2D sub-batch④a). The former package-private clinicScope/dbOrTx
// are swapped for repohelpers.ClinicScope/DBOrTx (identical predicate / ambient-tx participation);
// every external caller only ever saw this via the internal/repository facade (VitalRepository alias),
// so no call site changes.
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

// FindByMedicalRecordID は repohelpers.DBOrTx で ambient tx に参加する（BE9-2D ④b）。
// treatmentService の dose 体重解決（resolveDoseWeight）が保存 tx 内から読む read で、旧
// repos.Transaction（tx-bound clone）では暗黙に tx 参加していた。WithTx 化後も同一 tx で読み、
// 並行 vital 変更による dose スナップショット TOCTOU を作らない（#201 B-2 security review MEDIUM-1）。
func (r *vitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	vitals := make([]model.VitalRecord, 0)
	if err := repohelpers.DBOrTx(ctx, r.db).
		Scopes(repohelpers.ClinicScope(clinicID)).
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
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		First(&vital).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))
	}
	return &vital, nil
}

// Create は repohelpers.DBOrTx(ctx, r.db) で ambient tx に参加する（BE-refactor.md X-11）。
// LockByIDForUpdate の行ロック保持 tx 内から呼ばれた場合、別コネクションで INSERT すると
// vital_records.medical_record_id の FK 制約チェックが同一行への FOR UPDATE ロックと
// デッドロックする（FK チェックは FOR KEY SHARE を要求し FOR UPDATE と競合するため）。
func (r *vitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(vital).Error; err != nil {
		return apperrors.FromGORM(err, "vital", "")
	}
	return nil
}

// Update は repohelpers.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.VitalRecord{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

// Delete は repohelpers.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := repohelpers.DBOrTx(ctx, r.db).
		Scopes(repohelpers.ClinicScope(clinicID)).
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
