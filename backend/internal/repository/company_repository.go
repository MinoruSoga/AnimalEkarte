package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CompanyRepository は法人情報のデータアクセスインターフェース
type CompanyRepository interface {
	Get(ctx context.Context) (*model.Company, error)
	Update(ctx context.Context, fields map[string]any) error
}

type companyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository は CompanyRepository を初期化して返す
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// Get は company テーブルの先頭レコードを返す。レコードがなければ WrapNotFound を返す。
func (r *companyRepository) Get(ctx context.Context) (*model.Company, error) {
	var company model.Company
	if err := r.db.WithContext(ctx).Limit(1).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("company", "1")
		}
		return nil, apperrors.Wrap(err, "get company")
	}
	return &company, nil
}

// Update は先頭レコード（id = 1）を map[string]any で部分更新する。
// RowsAffected == 0 の場合は WrapNotFound を返す。
func (r *companyRepository) Update(ctx context.Context, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("id = ?", 1).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update company")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("company", "1")
	}
	return nil
}
