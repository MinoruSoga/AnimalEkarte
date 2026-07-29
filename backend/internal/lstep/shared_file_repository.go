package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// SharedFileRepository は shared_files テーブルへのアクセスインターフェース。
type SharedFileRepository interface {
	Create(ctx context.Context, f *model.SharedFile) error
	FindByID(ctx context.Context, clinicID, id uint64) (*model.SharedFile, error)
	FindAll(ctx context.Context, clinicID uint64) ([]*model.SharedFile, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindExpired は created_at が threshold より古く未削除のレコードを返す（期限切れバッチ用）。
	FindExpired(ctx context.Context, thresholdUnix int64) ([]*model.SharedFile, error)
}

type sharedFileRepository struct{ db *gorm.DB }

// NewSharedFileRepository は SharedFileRepository を初期化して返す。
func NewSharedFileRepository(db *gorm.DB) SharedFileRepository {
	return &sharedFileRepository{db: db}
}

func (r *sharedFileRepository) Create(ctx context.Context, f *model.SharedFile) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return apperrors.FromGORM(err, "shared_file", "create")
	}
	return nil
}

func (r *sharedFileRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.SharedFile, error) {
	var f model.SharedFile
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&f).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shared_file", fmt.Sprintf("%d", id))
	}
	return &f, nil
}

// sharedFileListMax / sharedFileExpiredMax are hard safety caps (G2F-06).
const (
	sharedFileListMax    = 200
	sharedFileExpiredMax = 500
)

func (r *sharedFileRepository) FindAll(ctx context.Context, clinicID uint64) ([]*model.SharedFile, error) {
	var files []*model.SharedFile
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("created_at DESC").
		Limit(sharedFileListMax).
		Find(&files).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shared_file", fmt.Sprintf("clinic=%d", clinicID))
	}
	return files, nil
}

func (r *sharedFileRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.SharedFile{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shared_file", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shared_file", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *sharedFileRepository) FindExpired(ctx context.Context, thresholdUnix int64) ([]*model.SharedFile, error) {
	var files []*model.SharedFile
	err := r.db.WithContext(ctx).
		Where("EXTRACT(EPOCH FROM created_at) < ?", thresholdUnix).
		Order("created_at ASC").
		Limit(sharedFileExpiredMax).
		Find(&files).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shared_file", "expired")
	}
	return files, nil
}
