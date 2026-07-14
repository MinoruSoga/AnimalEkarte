package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type PetRepository interface {
	FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	// FindByIDForClinics は複数医院スコープでペットを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
	FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
	CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// CountLivingByOwner は指定オーナーの生存ペット数（deceased_at IS NULL）を返す。
	// ISSUE-007: CreateCheckupSync のサーバ側二重防御で誤配信を防ぐ。
	CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// CountLivingByOwnerIDs は複数オーナーの生存ペット数を一括取得する（N+1 解消用）。
	// 返り値は ownerID → 生存ペット数のマップ。存在しない ownerID は 0 として扱う。
	CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error)
	CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
	FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{db: db}
}

func (r *petRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	pets := make([]model.Pet, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Pet{}).Where("pets.clinic_id = ?", clinicID)
		if ownerID != nil {
			q = q.Where("pets.owner_id = ?", *ownerID)
		}
		if search != "" {
			// NormalizeKana で検索語のカタカナをひらがなに正規化。
			// DB 列は translate() でひらがなに正規化済みのため、双方を統一して比較する。
			pattern := "%" + escapeLike(NormalizeKana(search)) + "%"
			q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.deleted_at IS NULL").
				Where(
					`(pets.name ILIKE ? ESCAPE '\'`+
						` OR translate(pets.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
						` OR owners.name ILIKE ? ESCAPE '\'`+
						` OR translate(owners.name_kana, ?, ?) ILIKE ? ESCAPE '\')`,
					pattern,
					kanaSourceChars, kanaTargetChars, pattern,
					pattern,
					kanaSourceChars, kanaTargetChars, pattern,
				)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	if err := buildBase().Preload("Owner", "deleted_at IS NULL").Preload("AnimalSpecies").Preload("Insurance", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(paginate(page, limit)).Order("pets.created_at DESC").Find(&pets).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	return pets, total, nil
}

func (r *petRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, []uint64{clinicID}, id)
}

func (r *petRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, clinicIDs, id)
}

// findPetByID は認可済みクリニック集合を受け取りペットを1件取得する共通実装。
// Preload する保険マスタも同じ集合で clinic 隔離する（別クリニックの保険マスタ混入防止）。
func (r *petRepository) findPetByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	var pet model.Pet
	err := r.db.WithContext(ctx).
		Preload("Owner", "deleted_at IS NULL").
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(clinicScopeIn(clinicIDs)).Where("id = ?", id).First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	return &pet, nil
}

func (r *petRepository) FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Preload("AnimalSpecies").
		Where("owner_id = ? AND deceased_at IS NULL AND deleted_at IS NULL", ownerID).
		Order("created_at ASC").
		Find(&pets).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	return pets, nil
}

func (r *petRepository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

// CountLivingByOwner は指定オーナーの生存ペット数（deceased_at IS NULL）を返す。
// ISSUE-007: CreateCheckupSync のサーバ側二重防御で死亡ペットのみの飼い主を除外するために使用する。
func (r *petRepository) CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ? AND deceased_at IS NULL AND deleted_at IS NULL", ownerID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

type ownerPetCount struct {
	OwnerID uint64
	Count   int64
}

func (r *petRepository) CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error) {
	if len(ownerIDs) == 0 {
		return map[uint64]int64{}, nil
	}
	var rows []ownerPetCount
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
		Select("owner_id, COUNT(*) AS count").
		Where("owner_id IN ? AND deceased_at IS NULL AND deleted_at IS NULL", ownerIDs).
		Group("owner_id").
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	result := make(map[uint64]int64, len(ownerIDs))
	for _, row := range rows {
		result[row.OwnerID] = row.Count
	}
	return result, nil
}

func (r *petRepository) CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("animal_species_id = ? AND deleted_at IS NULL", speciesID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

func (r *petRepository) Create(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("pet", "pet number already registered")
		}
		return apperrors.FromGORM(err, "pet", "")
	}
	loaded, err := r.FindByID(ctx, pet.ClinicID, pet.ID)
	if err != nil {
		return apperrors.Wrap(err, "reload pet after create")
	}
	*pet = *loaded
	return nil
}

func (r *petRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Pet{}).
			Scopes(clinicScope(clinicID)).
			Where("id = ?", id).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
		}
		if count == 0 {
			return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
		}
		// レコードは存在するが値が変わらなかった → success
	}
	return nil
}

func (r *petRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Pet{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	return nil
}

// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
func (r *petRepository) FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error) {
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
		Where("deceased_at IS NULL AND deleted_at IS NULL").
		Where("EXTRACT(month FROM birth_date) = ? AND EXTRACT(day FROM birth_date) = ?", month, day).
		Distinct("owner_id").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("clinic=%d birthday=%02d-%02d", clinicID, month, day))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}
