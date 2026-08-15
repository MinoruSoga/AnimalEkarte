package lstep

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	// FindAllByClinicID はクリニックスコープで最新順にインポート履歴一覧を返す。
	FindAllByClinicID(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
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
	// NOTE: GORM's Save() keys UPDATE purely on the primary key when it is already populated
	// (imp.ID is a non-zero UUID here) and silently ignores any chained .Where(...) clause —
	// the clinic_id predicate below would have zero effect and allow cross-tenant overwrites.
	// Use an explicit Model+Where+Updates(map) so the clinic_id predicate actually gates the
	// UPDATE's WHERE clause (P4: clinicScope on Update, MANDATORY).
	if err := r.db.WithContext(ctx).
		Model(&model.LstepCsvImport{}).
		Where("id = ? AND clinic_id = ?", imp.ID, imp.ClinicID).
		Updates(map[string]any{
			"csv_type":            imp.CsvType,
			"file_name":           imp.FileName,
			"uploaded_by_user_id": imp.UploadedByUserID,
			"row_count":           imp.RowCount,
			"success_count":       imp.SuccessCount,
			"error_count":         imp.ErrorCount,
			"status":              imp.Status,
			"error_log":           imp.ErrorLog,
			"imported_at":         imp.ImportedAt,
		}).Error; err != nil {
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

func (r *lstepCsvImportRepository) FindAllByClinicID(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	var imports []*model.LstepCsvImport
	err := r.db.WithContext(ctx).
		Omit("error_log").
		Where("clinic_id = ?", clinicID).
		Order("imported_at DESC NULLS LAST").
		Limit(limit).
		Find(&imports).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_csv_import", "list")
	}
	return imports, nil
}
