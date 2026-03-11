package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Clinic, error)
	GetClinicInfo(ctx context.Context) (*model.ClinicInfo, error)
	UpdateClinicInfo(ctx context.Context, info *model.ClinicInfo) error
	Create(ctx context.Context, clinic *model.Clinic) error
	Update(ctx context.Context, clinic *model.Clinic) error
}

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	var clinics []model.Clinic
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&clinics).Error; err != nil {
		return nil, apperrors.Wrap(err, "find clinics")
	}
	return clinics, nil
}

func (r *clinicRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Clinic, error) {
	var clinic model.Clinic
	if err := r.db.WithContext(ctx).First(&clinic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinic", id.String())
		}
		return nil, apperrors.Wrap(err, "find clinic by id")
	}
	return &clinic, nil
}

func (r *clinicRepository) GetClinicInfo(ctx context.Context) (*model.ClinicInfo, error) {
	var info model.ClinicInfo
	if err := r.db.WithContext(ctx).First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinic_info", "singleton")
		}
		return nil, apperrors.Wrap(err, "get clinic info")
	}
	return &info, nil
}

func (r *clinicRepository) UpdateClinicInfo(ctx context.Context, info *model.ClinicInfo) error {
	var existing model.ClinicInfo
	err := r.db.WithContext(ctx).First(&existing).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// レコードが存在しない場合は新規作成
		if info.ID == (uuid.UUID{}) {
			info.ID = uuid.New()
		}
		if err := r.db.WithContext(ctx).Create(info).Error; err != nil {
			return apperrors.Wrap(err, "create clinic info")
		}
		return nil
	}
	if err != nil {
		return apperrors.Wrap(err, "get clinic info for update")
	}
	info.ID = existing.ID
	if err := r.db.WithContext(ctx).Save(info).Error; err != nil {
		return apperrors.Wrap(err, "update clinic info")
	}
	return nil
}

func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	if err := r.db.WithContext(ctx).Create(clinic).Error; err != nil {
		return apperrors.Wrap(err, "create clinic")
	}
	return nil
}

func (r *clinicRepository) Update(ctx context.Context, clinic *model.Clinic) error {
	if err := r.db.WithContext(ctx).Save(clinic).Error; err != nil {
		return apperrors.Wrap(err, "update clinic")
	}
	return nil
}
