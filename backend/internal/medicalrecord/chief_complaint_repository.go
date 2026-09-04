package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ChiefComplaintTypeRepository is the data access interface for chief complaint types.
// Moved from internal/repository/chiefcomplaint — BE9-2C roll-up (see
// diagnosis_type_repository.go's header for the rename rationale).
type ChiefComplaintTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, category *model.ChiefComplaintType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type chiefComplaintTypeRepository struct{ db *gorm.DB }

// NewChiefComplaintTypeRepository constructs a ChiefComplaintTypeRepository.
func NewChiefComplaintTypeRepository(db *gorm.DB) ChiefComplaintTypeRepository {
	return &chiefComplaintTypeRepository{db: db}
}

func (r *chiefComplaintTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	categories := make([]model.ChiefComplaintType, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return categories, nil
}

func (r *chiefComplaintTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	return persistence.FindByIDScoped[model.ChiefComplaintType](ctx, r.db, "chief_complaint_type", clinicID, id)
}

func (r *chiefComplaintTypeRepository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	db := r.db.WithContext(ctx)
	wantActive := category.IsActive
	if err := db.Create(category).Error; err != nil {
		return apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	if !wantActive {
		if err := db.Model(category).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "chief_complaint_type", fmt.Sprintf("%d", category.ID))
		}
		category.IsActive = false
	}
	return nil
}

func (r *chiefComplaintTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *chiefComplaintTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, ids, "sort_order")
}

// CountUsageByChiefComplaintTypeID returns inquiry references.
// inquiries lack clinic_id; tenant isolation is via medical_records JOIN.
func (r *chiefComplaintTypeRepository) CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Inquiry{}).
		Scopes(persistence.MedicalRecordTenantScope("inquiries", clinicID)).
		Where("inquiries.chief_complaint_type_id = ? AND inquiries.deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inquiry", "")
	}
	return count, nil
}

func (r *chiefComplaintTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM inquiries
			JOIN medical_records ON medical_records.id = inquiries.medical_record_id
			  AND medical_records.clinic_id = ?
			  AND medical_records.deleted_at IS NULL
			WHERE inquiries.chief_complaint_type_id = chief_complaint_types.id
		)`, clinicID).
		Delete(&model.ChiefComplaintType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "chief_complaint_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeChiefComplaintDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *chiefComplaintTypeRepository) normalizeChiefComplaintDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この主訴カテゴリは問診記録で使用中のため削除できません")
}
