package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type OwnerRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error)
	FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error)
	CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

type ownerRepository struct {
	db *gorm.DB
}

func NewOwnerRepository(db *gorm.DB) OwnerRepository {
	return &ownerRepository{db: db}
}

func (r *ownerRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners := make([]model.Owner, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Owner{}).Where("clinic_id = ?", clinicID)
		if search != "" {
			pattern := "%" + escapeLike(search) + "%"
			q = q.Where(
				`(owner_name ILIKE ? ESCAPE '\' OR phone ILIKE ? ESCAPE '\' OR email ILIKE ? ESCAPE '\')`,
				pattern, pattern, pattern,
			)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count owners")
	}
	if err := buildBase().
		Preload("Pets").Preload("Pets.AnimalSpecies").Preload("Pets.Insurance").
		Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").
		Find(&owners).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find owners")
	}
	return owners, total, nil
}

func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).Preload("Pets").Preload("Pets.AnimalSpecies").Preload("Pets.Insurance").First(&owner, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

func (r *ownerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).First(&owner, "clinic_id = ? AND email = ?", clinicID, email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, "find owner by email")
	}
	return &owner, nil
}

// FindByPhone は clinic_id + phone に一致するオーナーを返す。見つからない場合は nil を返す。
func (r *ownerRepository) FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).First(&owner, "clinic_id = ? AND phone = ?", clinicID, phone).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, "find owner by phone")
	}
	return &owner, nil
}

func (r *ownerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 飼主を作成
		if err := tx.Create(owner).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return apperrors.WrapAlreadyExists("owner", "email already registered")
			}
			return apperrors.Wrap(err, "create owner")
		}
		// 2. ペットを順次作成（owner_id, clinic_id をサーバー側でセット）
		for i := range pets {
			pets[i].OwnerID = owner.ID
			pets[i].ClinicID = owner.ClinicID
			if err := tx.Create(&pets[i]).Error; err != nil {
				return apperrors.Wrap(err, "create pet")
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "create owner with pets")
	}
	// トランザクションコミット後に全リレーションをロードして呼び出し元に反映
	loaded, err := r.FindByID(ctx, owner.ClinicID, owner.ID)
	if err != nil {
		return apperrors.Wrap(err, "reload owner after create")
	}
	*owner = *loaded
	return nil
}

// escapeLike escapes LIKE wildcard characters in s for use with PostgreSQL ILIKE ... ESCAPE '\'.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *ownerRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update owner")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("owner", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *ownerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Owner{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete owner")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("owner", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountPetsByOwnerID は指定されたオーナーに紐付いているペット数を返す
func (r *ownerRepository) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count pets by owner")
	}
	return count, nil
}
