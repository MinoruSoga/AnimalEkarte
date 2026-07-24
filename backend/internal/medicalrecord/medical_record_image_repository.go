package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MedicalRecordImageRepository は診療画像のデータアクセス層。
// Moved from internal/repository (BE9-2D sub-batch④a). The former package-private dbOrTx/
// medicalRecordTenantScope are swapped for persistence.DBOrTx/MedicalRecordTenantScope (identical
// ambient-tx participation / join predicate); every external caller only ever saw this via the
// internal/repository facade (MedicalRecordImageRepository alias), so no call site changes.
type MedicalRecordImageRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	Create(ctx context.Context, image *model.MedicalRecordImage) error
	Delete(ctx context.Context, clinicID, id uint64) error
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordImage, error)
}

type medicalRecordImageRepository struct {
	db *gorm.DB
}

// NewMedicalRecordImageRepository は MedicalRecordImageRepository を初期化して返す
func NewMedicalRecordImageRepository(db *gorm.DB) MedicalRecordImageRepository {
	return &medicalRecordImageRepository{db: db}
}

func (r *medicalRecordImageRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	images := make([]model.MedicalRecordImage, 0)
	if err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = medical_record_images.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND medical_record_images.medical_record_id = ?", clinicID, medicalRecordID).
		Preload("Staff", "deleted_at IS NULL").
		Order("medical_record_images.sort_order ASC, medical_record_images.created_at ASC").
		Find(&images).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_image", "")
	}
	return images, nil
}

// Create は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（SD-2 / BE-refactor.md X-11、
// examination_repository.go Create と同じ理由 — LockByIDForUpdate の行ロックと同一 tx で
// 直列化しないとデッドロックしうる）。
func (r *medicalRecordImageRepository) Create(ctx context.Context, image *model.MedicalRecordImage) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(image).Error; err != nil {
		return apperrors.FromGORM(err, "medical_record_image", "")
	}
	return nil
}

// Delete は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、SD-2）。
func (r *medicalRecordImageRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Where("medical_record_images.id = ? AND medical_record_images.medical_record_id IN "+
			"(SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Delete(&model.MedicalRecordImage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medical_record_image", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record_image", fmt.Sprintf("%d", id))
	}
	return nil
}

// FindByID は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Delete の事前所有権チェックが
// 同一 tx 内の一貫した読み取りになるよう、Create/Delete と同じ理由で揃える）。
func (r *medicalRecordImageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordImage, error) {
	var image model.MedicalRecordImage
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.MedicalRecordTenantScope("medical_record_images", clinicID)).
		Where("medical_record_images.id = ?", id).
		Preload("Staff", "deleted_at IS NULL").
		First(&image).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_image", fmt.Sprintf("%d", id))
	}
	return &image, nil
}
