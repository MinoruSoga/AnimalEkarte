package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type VaccinationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	// FindByOwner は飼い主の生存ワクチン記録（ペット経由）を全件返す（ISSUE-004 タグ再同期用）。
	// 飼い主のペットがすべて削除済みの場合は空配列を返す。
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindOwnersByVaccineDeadline はワクチン次回接種日（next_date）が targetDate の飼い主IDリストを返す（FEAT-383）。
	FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

type vaccinationRepository struct {
	db *gorm.DB
}

func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

func (r *vaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Vaccination{}).
			Where("vaccinations.clinic_id = ?", clinicID)
		if petID != nil {
			q = q.Where("vaccinations.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
		}
		if startDate != nil {
			q = q.Where("vaccinations.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("vaccinations.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}

	vaccinations := make([]model.Vaccination, 0)
	if err := buildBase().
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.Owner", "deleted_at IS NULL").
		Preload("Doctor", "deleted_at IS NULL").
		Offset((page - 1) * limit).Limit(limit).Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}
	return vaccinations, total, nil
}

// FindByOwner は飼い主の全生存ワクチン記録を返す（ISSUE-004 タグ再同期用）。
// pets テーブルを JOIN し、飼い主に属する生存ペットのワクチンのみ取得する。
func (r *vaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	vaccinations := make([]model.Vaccination, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.deleted_at IS NULL").
		Where("vaccinations.clinic_id = ? AND pets.owner_id = ?", clinicID, ownerID).
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("owner=%d", ownerID))
	}
	return vaccinations, nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	var vaccination model.Vaccination
	err := r.db.WithContext(ctx).
		Where("vaccinations.id = ? AND vaccinations.clinic_id = ?", id, clinicID).
		Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Pet", "deleted_at IS NULL").Preload("Pet.Owner", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").
		First(&vaccination).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", id))
	}
	return &vaccination, nil
}

func (r *vaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	if err := r.db.WithContext(ctx).Create(vaccination).Error; err != nil {
		return apperrors.FromGORM(err, "vaccination", "")
	}
	return nil
}

func (r *vaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	if err := updateScopedByID(ctx, r.db, &model.Vaccination{}, "vaccination", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *vaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Vaccination{}, "vaccination", clinicID, id)
}

// FindOwnersByVaccineDeadline はワクチン次回接種日（next_date）が targetDate の飼い主IDリストを返す（FEAT-383）。
// pets テーブルを JOIN し、生存ペット経由で飼い主IDを取得する。
func (r *vaccinationRepository) FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	target := targetDate.Format("2006-01-02")
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.deleted_at IS NULL AND pets.deceased_at IS NULL").
		Where("vaccinations.clinic_id = ? AND vaccinations.deleted_at IS NULL", clinicID).
		Where("vaccinations.next_date::date = ?::date", target).
		Distinct("pets.owner_id AS owner_id").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("clinic=%d vaccine_deadline=%s", clinicID, target))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}
