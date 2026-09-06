package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// VitalRepository はバイタル記録のデータアクセスインターフェース。
// Moved from internal/repository (BE9-2D sub-batch④a). The former package-private clinicScope/dbOrTx
// are swapped for persistence.ClinicScope/DBOrTx (identical predicate / ambient-tx participation);
// every external caller only ever saw this via the internal/repository facade (VitalRepository alias),
// so no call site changes.
type VitalRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error)
	Create(ctx context.Context, vital *model.VitalRecord) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateVitalInput) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vitalRepository struct {
	db *gorm.DB
}

// NewVitalRepository はVitalRepositoryを初期化して返す
func NewVitalRepository(db *gorm.DB) VitalRepository {
	return &vitalRepository{db: db}
}

// FindByMedicalRecordID は persistence.DBOrTx で ambient tx に参加する（BE9-2D ④b）。
// treatmentService の dose 体重解決（resolveDoseWeight）が保存 tx 内から読む read で、旧
// repos.Transaction（tx-bound clone）では暗黙に tx 参加していた。WithTx 化後も同一 tx で読み、
// 並行 vital 変更による dose スナップショット TOCTOU を作らない（#201 B-2 security review MEDIUM-1）。
func (r *vitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	vitals := make([]model.VitalRecord, 0)
	// Parent medical_records clinic correlation (SEC-SWEEP-02-MR-B1): child clinic alone
	// is insufficient when medical_record_id is a corrupt cross-tenant FK.
	// Qualify vital_records.clinic_id (not ClinicScope) so JOIN medical_records does not
	// make the bare clinic_id predicate ambiguous.
	if err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = vital_records.medical_record_id AND medical_records.clinic_id = vital_records.clinic_id AND medical_records.deleted_at IS NULL").
		Where("vital_records.clinic_id = ? AND vital_records.medical_record_id = ? AND vital_records.deleted_at IS NULL", clinicID, medicalRecordID).
		Order("vital_records.recorded_at ASC").
		Find(&vitals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vital", "")
	}
	return vitals, nil
}

// FindByID participates in an ambient transaction so Update can complete its
// response re-fetch before commit and roll back when that re-fetch fails.
func (r *vitalRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error) {
	var vital model.VitalRecord
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		First(&vital).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))
	}
	return &vital, nil
}

// Create は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（BE-refactor.md X-11）。
// LockByIDForUpdate の行ロック保持 tx 内から呼ばれた場合、別コネクションで INSERT すると
// vital_records.medical_record_id の FK 制約チェックが同一行への FOR UPDATE ロックと
// デッドロックする（FK チェックは FOR KEY SHARE を要求し FOR UPDATE と競合するため）。
func (r *vitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(vital).Error; err != nil {
		return apperrors.FromGORM(err, "vital", "")
	}
	return nil
}

// Update は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateVitalInput) error {
	return r.update(ctx, clinicID, id, buildVitalUpdate(&cmd))
}

func (r *vitalRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.VitalRecord{}).
		Scopes(persistence.ClinicScope(clinicID)).
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

// Delete は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *vitalRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
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
