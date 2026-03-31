package repository

import (
	"context"

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
	err := r.db.WithContext(ctx).Limit(1).First(&company).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "1")
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
		return apperrors.FromGORM(result.Error, "company", "1")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("company", "1")
	}
	return nil
}
