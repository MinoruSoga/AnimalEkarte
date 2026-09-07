package clinic

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type companyRepository struct {
	db *gorm.DB
}

// FindSingleton は company テーブルの先頭レコードを返す。レコードがなければ WrapNotFound を返す。
func (r *companyRepository) FindSingleton(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := r.db.WithContext(ctx).First(&company, 1).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "1")
	}
	return &company, nil
}

// Update は先頭レコード（id = 1）を typed command で部分更新する。
// RowsAffected == 0 の場合は WrapNotFound を返す。
func (r *companyRepository) Update(ctx context.Context, cmd UpdateCompanyInput) error {
	return r.update(ctx, BuildCompanyUpdate(&cmd))
}

func (r *companyRepository) update(ctx context.Context, fields map[string]any) error {
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
