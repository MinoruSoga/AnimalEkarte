package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type OwnerRepository interface {
	FindAll(ctx context.Context, clinicID uuid.UUID, page, limit int, search string) ([]model.Owner, int64, error)
	FindByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Owner, error)
	Create(ctx context.Context, owner *model.Owner) error
	Update(ctx context.Context, owner *model.Owner) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type ownerRepository struct {
	db *gorm.DB
}

func NewOwnerRepository(db *gorm.DB) OwnerRepository {
	return &ownerRepository{db: db}
}

func (r *ownerRepository) FindAll(ctx context.Context, clinicID uuid.UUID, page, limit int, search string) ([]model.Owner, int64, error) {
	var owners []model.Owner
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Owner{}).Where("clinic_id = ?", clinicID)
	if search != "" {
		q = q.Where("owner_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count owners")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&owners).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find owners")
	}
	return owners, total, nil
}

func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Owner, error) {
	var owner model.Owner
	if err := r.db.WithContext(ctx).Preload("Pets").First(&owner, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("owner", id.String())
		}
		return nil, apperrors.Wrap(err, "find owner by id")
	}
	return &owner, nil
}

func (r *ownerRepository) Create(ctx context.Context, owner *model.Owner) error {
	if err := r.db.WithContext(ctx).Create(owner).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("owner", owner.Email)
		}
		return apperrors.Wrap(err, "create owner")
	}
	return nil
}

func (r *ownerRepository) Update(ctx context.Context, owner *model.Owner) error {
	if err := r.db.WithContext(ctx).Save(owner).Error; err != nil {
		return apperrors.Wrap(err, "update owner")
	}
	return nil
}

func (r *ownerRepository) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Owner{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete owner")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("owner", id.String())
	}
	return nil
}
