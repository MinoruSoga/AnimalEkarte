package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicRepository クリニックリポジトリインターフェース
type ClinicRepository interface {
	GetAllClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id string) (*model.Clinic, error)
	CreateClinic(ctx context.Context, clinic *model.Clinic) error
	UpdateClinic(ctx context.Context, clinic *model.Clinic) error
	DeleteClinic(ctx context.Context, id string) error
}

// clinicRepository クリニックリポジトリ実装
type clinicRepository struct {
	db *gorm.DB
}

// NewClinicRepository 新しいクリニックリポジトリを作成
func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

// GetAllClinics 全てのクリニックを取得
func (r *clinicRepository) GetAllClinics(ctx context.Context) ([]model.Clinic, error) {
	var clinics []model.Clinic
	result := r.db.WithContext(ctx).
		Preload("Staffs").
		Order("created_at DESC").
		Find(&clinics)

	if result.Error != nil {
		return nil, result.Error
	}

	return clinics, nil
}

// GetClinicByID IDでクリニックを取得
func (r *clinicRepository) GetClinicByID(ctx context.Context, id string) (*model.Clinic, error) {
	var clinic model.Clinic
	result := r.db.WithContext(ctx).
		Preload("Staffs").
		First(&clinic, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinic with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get clinic")
	}

	return &clinic, nil
}

// CreateClinic クリニックを作成
func (r *clinicRepository) CreateClinic(ctx context.Context, clinic *model.Clinic) error {
	// Generate UUID if not set
	if clinic.ID == uuid.Nil {
		clinic.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(clinic)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateClinic クリニックを更新
func (r *clinicRepository) UpdateClinic(ctx context.Context, clinic *model.Clinic) error {
	result := r.db.WithContext(ctx).Save(clinic)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteClinic クリニックを削除
func (r *clinicRepository) DeleteClinic(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Clinic{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// StaffRepository スタッフリポジトリインターフェース
type StaffRepository interface {
	GetAllStaff(ctx context.Context) ([]model.Staff, error)
	GetStaffByID(ctx context.Context, id string) (*model.Staff, error)
	GetStaffByClinicID(ctx context.Context, clinicID string) ([]model.Staff, error)
	CreateStaff(ctx context.Context, staff *model.Staff) error
	UpdateStaff(ctx context.Context, staff *model.Staff) error
	DeleteStaff(ctx context.Context, id string) error
}

// staffRepository スタッフリポジトリ実装
type staffRepository struct {
	db *gorm.DB
}

// NewStaffRepository 新しいスタッフリポジトリを作成
func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

// GetAllStaff 全てのスタッフを取得
func (r *staffRepository) GetAllStaff(ctx context.Context) ([]model.Staff, error) {
	var staffs []model.Staff
	result := r.db.WithContext(ctx).
		Preload("Clinic").
		Order("created_at DESC").
		Find(&staffs)

	if result.Error != nil {
		return nil, result.Error
	}

	return staffs, nil
}

// GetStaffByID IDでスタッフを取得
func (r *staffRepository) GetStaffByID(ctx context.Context, id string) (*model.Staff, error) {
	var staff model.Staff
	result := r.db.WithContext(ctx).
		Preload("Clinic").
		First(&staff, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("staff with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get staff")
	}

	return &staff, nil
}

// GetStaffByClinicID クリニックIDでスタッフを取得
func (r *staffRepository) GetStaffByClinicID(ctx context.Context, clinicID string) ([]model.Staff, error) {
	var staffs []model.Staff
	result := r.db.WithContext(ctx).
		Preload("Clinic").
		Where("clinic_id = ?", clinicID).
		Order("created_at DESC").
		Find(&staffs)

	if result.Error != nil {
		return nil, result.Error
	}

	return staffs, nil
}

// CreateStaff スタッフを作成
func (r *staffRepository) CreateStaff(ctx context.Context, staff *model.Staff) error {
	// Generate UUID if not set
	if staff.ID == uuid.Nil {
		staff.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(staff)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateStaff スタッフを更新
func (r *staffRepository) UpdateStaff(ctx context.Context, staff *model.Staff) error {
	result := r.db.WithContext(ctx).Save(staff)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteStaff スタッフを削除
func (r *staffRepository) DeleteStaff(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Staff{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
