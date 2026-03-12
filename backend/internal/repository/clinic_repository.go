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
	GetCompany(ctx context.Context) (*model.Company, error)
	UpdateCompany(ctx context.Context, company *model.Company) error
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

func (r *clinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	if err := r.db.WithContext(ctx).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("company", "singleton")
		}
		return nil, apperrors.Wrap(err, "get company")
	}
	return &company, nil
}

func (r *clinicRepository) UpdateCompany(ctx context.Context, company *model.Company) error {
	var existing model.Company
	err := r.db.WithContext(ctx).First(&existing).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// レコードが存在しない場合は新規作成
		if company.ID == (uuid.UUID{}) {
			company.ID = uuid.New()
		}
		if err := r.db.WithContext(ctx).Create(company).Error; err != nil {
			return apperrors.Wrap(err, "create company")
		}
		return nil
	}
	if err != nil {
		return apperrors.Wrap(err, "get company for update")
	}
	company.ID = existing.ID
	if err := r.db.WithContext(ctx).Save(company).Error; err != nil {
		return apperrors.Wrap(err, "update company")
	}
	return nil
}

func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	if err := r.db.WithContext(ctx).Create(clinic).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
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
