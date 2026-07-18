// Package company owns companies データアクセス (BE8-4 batch15 — leaf domain).
package company

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Repository は法人情報のデータアクセスインターフェース
type Repository interface {
	FindSingleton(ctx context.Context) (*model.Company, error)
	Update(ctx context.Context, fields map[string]any) error
}

type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindSingleton は company テーブルの先頭レコードを返す。レコードがなければ WrapNotFound を返す。
func (r *repository) FindSingleton(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := r.db.WithContext(ctx).First(&company, 1).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "1")
	}
	return &company, nil
}

// Update は先頭レコード（id = 1）を map[string]any で部分更新する。
// RowsAffected == 0 の場合は WrapNotFound を返す。
func (r *repository) Update(ctx context.Context, fields map[string]any) error {
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
