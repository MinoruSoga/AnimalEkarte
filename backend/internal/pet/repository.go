package pet

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// PetListFilters はペット一覧のフィルタ条件（#266: サーバサイド pagination/search 拡張）。
// Search はペット名/カナ + 飼主名/カナ/電話を対象に ILIKE + NormalizeKana で部分一致検索する。
type PetListFilters struct {
	OwnerID         *uint64
	Search          string
	AnimalSpeciesID *uint64
	// IncludeDeceased: false（既定）は deceased_at IS NULL（生存のみ）に絞る。
	IncludeDeceased bool
}

type Repository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) のペットを検索する。clinicIDs はハンドラ層で所属検証済みであること。
	// 順序は owners.name_kana ASC, pets.id ASC で安定（#266: ページ送りでの重複/欠落防止）。
	FindAll(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error)
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

// LifecycleWriter is the typed pet lifecycle capability consumed by LSTEP.
// It deliberately exposes only the two permitted status transitions.
type LifecycleWriter interface {
	RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error
	ClearDeath(ctx context.Context, clinicID, petID uint64) error
}

// CompleteRepository is the concrete composition surface returned by NewRepository.
// The legacy repository facade intentionally narrows this to Repository until central
// composition cuts over to the typed lifecycle capability.
type CompleteRepository interface {
	Repository
	LifecycleWriter
}

type repository struct {
	db        *gorm.DB
	petWriter Creator
}

// NewRepository constructs the pet persistence and lifecycle capability.
func NewRepository(db *gorm.DB) CompleteRepository {
	return NewRepositoryWithWriter(db, NewWriter(db))
}

// NewRepositoryWithWriter allows focused write-owner capability tests.
func NewRepositoryWithWriter(db *gorm.DB, petWriter Creator) CompleteRepository {
	return &repository{db: db, petWriter: petWriter}
}

func (r *repository) FindAll(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error) {
	pets := make([]model.Pet, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return pets, 0, nil
	}

	buildBase := func() *gorm.DB {
		// owners への LEFT JOIN は search 有無に関わらず常に張る
		// (owners.name_kana ASC を安定順序の主キーにするため、Order 句が常にこの JOIN を要求する)。
		// clinicScopeIn は "clinic_id" を無修飾で参照し pets/owners 両方に同名列を持つため、
		// JOIN 併用時の曖昧列エラーを避けて pets.clinic_id / owners.clinic_id を明示指定する
		// （owners 側も同一 clinicIDs で二重にスコープし、クロステナント JOIN 汚染を防ぐ）。
		q := r.db.WithContext(ctx).Model(&model.Pet{}).
			Where("pets.clinic_id IN ?", clinicIDs).
			Where("pets.deleted_at IS NULL").
			Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id IN ? AND owners.deleted_at IS NULL", clinicIDs)
		if filters.OwnerID != nil {
			q = q.Where("pets.owner_id = ?", *filters.OwnerID)
		}
		if filters.AnimalSpeciesID != nil {
			q = q.Where("pets.animal_species_id = ?", *filters.AnimalSpeciesID)
		}
		if !filters.IncludeDeceased {
			q = q.Where("pets.deceased_at IS NULL")
		}
		if filters.Search != "" {
			// NormalizeKana で検索語のカタカナをひらがなに正規化。
			// DB 列は translate() でひらがなに正規化済みのため、双方を統一して比較する。
			pattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(filters.Search)) + "%"
			q = q.Where(
				`(pets.name ILIKE ? ESCAPE '\'`+
					` OR translate(pets.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR owners.name ILIKE ? ESCAPE '\'`+
					` OR translate(owners.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR owners.phone ILIKE ? ESCAPE '\')`,
				pattern,
				textsearch.KanaSourceChars, textsearch.KanaTargetChars, pattern,
				pattern,
				textsearch.KanaSourceChars, textsearch.KanaTargetChars, pattern,
				pattern,
			)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	if err := buildBase().
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(repohelpers.Paginate(page, limit)).
		// 順序の安定性: owners.name_kana ASC を主キーに、pets.id ASC を一意タイブレーカとする
		// （#266: 同一 kana でもページ送りで行の重複/欠落が起きないようにする）。
		Order("owners.name_kana ASC, pets.id ASC").
		Find(&pets).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	return pets, total, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, []uint64{clinicID}, id)
}

func (r *repository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, clinicIDs, id)
}

// findPetByID は認可済みクリニック集合を受け取りペットを1件取得する共通実装。
// Preload する飼主と保険マスタも同じ集合で clinic 隔離する
// （破損した owner_id / insurance_id から別クリニックのデータが混入するのを防止）。
func (r *repository) findPetByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	var pet model.Pet
	err := r.db.WithContext(ctx).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(repohelpers.ClinicScopeIn(clinicIDs)).Where("id = ?", id).First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	return &pet, nil
}

func (r *repository) FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Preload("AnimalSpecies").
		Where("owner_id = ? AND deceased_at IS NULL AND deleted_at IS NULL", ownerID).
		Order("created_at ASC").
		Find(&pets).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	return pets, nil
}

func (r *repository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

// CountLivingByOwner は指定オーナーの生存ペット数（deceased_at IS NULL）を返す。
// ISSUE-007: CreateCheckupSync のサーバ側二重防御で死亡ペットのみの飼い主を除外するために使用する。
func (r *repository) CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error) {
	if len(ownerIDs) == 0 {
		return map[uint64]int64{}, nil
	}
	var rows []ownerPetCount
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("animal_species_id = ? AND deleted_at IS NULL", speciesID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

func (r *repository) Create(ctx context.Context, pet *model.Pet) error {
	if r.petWriter == nil {
		return apperrors.WrapInternalServerError("pet writer is not configured")
	}
	created, err := r.petWriter.Create(ctx, CreateIntent{
		ClinicID: pet.ClinicID,
		OwnerID:  pet.OwnerID,
		Pet:      CreatePetDraftFromModel(*pet),
	})
	if err != nil {
		if apperrors.IsAlreadyExists(err) {
			return apperrors.WrapAlreadyExists("pet", "pet number already registered")
		}
		return err
	}
	*pet = *created
	return nil
}

// Update は BUG-407 の fail-closed 化（lstepLifecycleService.HandlePetDeath/HandlePetRevival が
// status/deceased_at 更新と監査書込を同一 tx で原子化する）のため dbOrTx(ctx, r.db) を使う。
// ambient tx が無い呼び出し（大多数の既存経路）では r.db.WithContext(ctx) と等価（後方互換）。
func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	db := repohelpers.DBOrTx(ctx, r.db)
	result := db.
		Model(&model.Pet{}).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := db.Model(&model.Pet{}).
			Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	return r.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at":     deceasedAt,
		"deceased_reason": reason,
		"status":          model.PetStatusDeceased,
	})
}

func (r *repository) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	return r.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at":     nil,
		"deceased_reason": nil,
		"status":          model.PetStatusAlive,
	})
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.Pet{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	return nil
}

// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
func (r *repository) FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error) {
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).
		Table("pets AS p").
		Joins("JOIN owners AS o ON o.id = p.owner_id AND o.clinic_id = p.clinic_id AND o.deleted_at IS NULL").
		Where("p.clinic_id = ? AND p.deceased_at IS NULL AND p.deleted_at IS NULL", clinicID).
		Where("EXTRACT(month FROM p.birth_date) = ? AND EXTRACT(day FROM p.birth_date) = ?", month, day).
		Distinct("p.owner_id").
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
