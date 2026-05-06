package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepCsvImportRepository は lstep_csv_imports テーブルの永続化インターフェース。
type LstepCsvImportRepository interface {
	// Create は新規 CSV インポート履歴を作成する。
	Create(ctx context.Context, imp *model.LstepCsvImport) error
	// Update は CSV インポート履歴を更新する。
	Update(ctx context.Context, imp *model.LstepCsvImport) error
	// FindByID はクリニックスコープで ID に一致するインポート履歴を返す。
	FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error)
	// ListByClinic はクリニックスコープで最新順にインポート履歴一覧を返す。
	ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

type lstepCsvImportRepository struct{ db *gorm.DB }

// NewLstepCsvImportRepository は LstepCsvImportRepository を初期化して返す。
func NewLstepCsvImportRepository(db *gorm.DB) LstepCsvImportRepository {
	return &lstepCsvImportRepository{db: db}
}

func (r *lstepCsvImportRepository) Create(ctx context.Context, imp *model.LstepCsvImport) error {
	if err := r.db.WithContext(ctx).Create(imp).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_csv_import", "")
	}
	return nil
}

func (r *lstepCsvImportRepository) Update(ctx context.Context, imp *model.LstepCsvImport) error {
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", imp.ClinicID).
		Save(imp).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_csv_import", imp.ID.String())
	}
	return nil
}

func (r *lstepCsvImportRepository) FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error) {
	var imp model.LstepCsvImport
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&imp).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_csv_import", id.String())
	}
	return &imp, nil
}

func (r *lstepCsvImportRepository) ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	var imports []*model.LstepCsvImport
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("imported_at DESC NULLS LAST").
		Limit(limit).
		Find(&imports).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_csv_import", "list")
	}
	return imports, nil
}
