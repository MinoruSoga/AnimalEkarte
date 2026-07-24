package owner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// ServiceRepository is the minimal owner persistence view consumed by the
// owner use case.
type ServiceRepository interface {
	FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error)
	FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error)
	FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error)
	FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
	CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error
	// UpdateAndFind updates and reloads an owner in one transaction. A reload
	// failure rolls the write back.
	UpdateAndFind(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Owner, error)
	UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

// LookupRepository contains owner lookups consumed outside the owner use case.
type LookupRepository interface {
	FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error)
	FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error)
}

// LstepRepository contains owner persistence consumed by LSTEP workflows.
type LstepRepository interface {
	FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error)
	FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error)
	FindAllByLineUserID(ctx context.Context, lineUserID string) ([]model.Owner, error)
	UpdateLineFollowedAt(ctx context.Context, clinicID, id uint64, t time.Time) error
	UpdateLineBlockedAt(ctx context.Context, clinicID, id uint64, t time.Time) error
	RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error
	ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error
}

// Repository is the complete owner persistence capability used at the
// composition boundary. Application services accept narrower interfaces.
type Repository interface {
	ServiceRepository
	LookupRepository
	LstepRepository
}

type ownerRepository struct {
	db           *gorm.DB
	petRegistrar PetRegistrar
}

// NewRepository constructs the owner persistence capability. petRegistrar is
// required only by CreateWithPets; reads and owner-only writes do not use it.
func NewRepository(db *gorm.DB, petRegistrar PetRegistrar) Repository {
	return &ownerRepository{db: db, petRegistrar: petRegistrar}
}

func (r *ownerRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners := make([]model.Owner, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return owners, 0, nil
	}

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Owner{}).Scopes(repohelpers.ClinicScopeIn(clinicIDs))
		if search != "" {
			// NormalizeKana で検索語のカタカナをひらがなに正規化。
			// DB 列は translate() でひらがなに正規化済みのため、双方を統一して比較する。
			pattern := "%" + repohelpers.EscapeLike(repohelpers.NormalizeKana(search)) + "%"
			q = q.Where(
				`(name ILIKE ? ESCAPE '\'`+
					` OR translate(name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR phone ILIKE ? ESCAPE '\'`+
					` OR email ILIKE ? ESCAPE '\')`,
				pattern,
				repohelpers.KanaSourceChars, repohelpers.KanaTargetChars, pattern,
				pattern,
				pattern,
			)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "owner", "")
	}
	if err := buildBase().
		Preload("Pets", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.AnimalSpecies").
		Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(repohelpers.Paginate(page, limit)).Order("created_at DESC").
		Find(&owners).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "owner", "")
	}
	return owners, total, nil
}

func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return r.findOwnerByID(ctx, []uint64{clinicID}, id)
}

func (r *ownerRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	return r.findOwnerByID(ctx, clinicIDs, id)
}

// findOwnerByID は認可済みクリニック集合を受け取り飼主を1件取得する共通実装。
// clinicIDs は呼び出し側で検証済みの集合（単一は []uint64{clinicID}、拠点横断#86は全所属）。
// Preload するペット・飼主・保険マスタも同じ集合で clinic 隔離する
// （破損したowner/pet関連やinsurance_idから別クリニックのデータが混入するのを防止）。
func (r *ownerRepository) findOwnerByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	return r.findOwnerByIDWithDB(ctx, repohelpers.DBOrTx(ctx, r.db), clinicIDs, id)
}

func (r *ownerRepository) findOwnerByIDWithDB(
	ctx context.Context,
	db *gorm.DB,
	clinicIDs []uint64,
	id uint64,
) (*model.Owner, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("owner", fmt.Sprintf("%d", id))
	}
	var owner model.Owner
	err := db.WithContext(ctx).
		Preload("Pets", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.AnimalSpecies").
		Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(repohelpers.ClinicScopeIn(clinicIDs)).Where("id = ?", id).First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

func (r *ownerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).First(&owner, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "owner", email)
	}
	return &owner, nil
}

// FindByPhone は clinic_id + phone に一致するオーナーを返す。見つからない場合は nil を返す。
func (r *ownerRepository) FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).First(&owner, "phone = ?", phone).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "owner", phone)
	}
	return &owner, nil
}

// FindByNameAndPhone は clinic_id + owner_name + phone に完全一致するオーナーを返す。
// 0件 or 複数件の場合は nil を返す（1件の場合のみ返す）。
func (r *ownerRepository) FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	if name == "" || phone == "" {
		return nil, nil
	}
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("name = ? AND phone = ? AND deleted_at IS NULL", name, phone).
		Limit(2). // 2件以上あるかだけ判定すればよい
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%s/%s", name, phone))
	}
	if len(owners) != 1 {
		return nil, nil // 0件 or 複数件 → 自動紐付け不可
	}
	return &owners[0], nil
}

func (r *ownerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := repohelpers.WithTxValue(ctx, tx)
		// 1. 飼主を作成
		if err := tx.Omit(clause.Associations).Create(owner).Error; err != nil {
			if repohelpers.IsUniqueConstraintErr(err) {
				return apperrors.WrapAlreadyExists("owner", "email already registered")
			}
			return apperrors.FromGORM(err, "owner", "")
		}
		if r.petRegistrar == nil {
			return apperrors.WrapInternalServerError("owner registration pet registrar is not configured")
		}
		// 2. ペット側の write owner に、同じ ambient transaction 内で
		// owner-registration intent だけを委譲する。
		if _, err := r.petRegistrar.CreateForOwnerRegistration(txCtx, PetRegistrationIntent{
			ClinicID: owner.ClinicID,
			OwnerID:  owner.ID,
			Pets:     ownerRegistrationPetDrafts(pets),
		}); err != nil {
			return err
		}
		loaded, err := r.findOwnerByIDWithDB(
			txCtx,
			tx,
			[]uint64{owner.ClinicID},
			owner.ID,
		)
		if err != nil {
			return apperrors.Wrap(err, "reload owner after create")
		}
		*owner = *loaded
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to create owner with pets")
	}
	return nil
}

func ownerRegistrationPetDrafts(pets []model.Pet) []PetRegistrationDraft {
	drafts := make([]PetRegistrationDraft, 0, len(pets))
	for i := range pets {
		pet := pets[i]
		drafts = append(drafts, PetRegistrationDraft{
			AnimalSpeciesID: pet.AnimalSpeciesID,
			Name:            pet.Name,
			NameKana:        pet.NameKana,
			Breed:           pet.Breed,
			Color:           pet.Color,
			BloodType:       pet.BloodType,
			MicrochipNumber: pet.MicrochipNumber,
			Gender:          pet.Gender,
			Status:          pet.Status,
			BirthDate:       pet.BirthDate,
			Weight:          pet.Weight,
			NeuteredDate:    pet.NeuteredDate,
			AcquisitionType: pet.AcquisitionType,
			DangerLevel:     pet.DangerLevel,
			Food:            pet.Food,
			Environment:     pet.Environment,
			Phone:           pet.Phone,
			InsuranceID:     pet.InsuranceID,
			Remarks:         pet.Remarks,
		})
	}
	return drafts
}

func (r *ownerRepository) UpdateAndFind(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
) (*model.Owner, error) {
	var loaded *model.Owner
	err := repohelpers.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := repohelpers.WithTxValue(ctx, tx)
		if err := repohelpers.UpdateScopedByID(
			txCtx,
			tx,
			&model.Owner{},
			"owner",
			clinicID,
			id,
			fields,
		); err != nil {
			return err
		}

		var err error
		loaded, err = r.findOwnerByIDWithDB(txCtx, tx, []uint64{clinicID}, id)
		if err != nil {
			return apperrors.Wrap(err, "reload owner after update")
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update and reload owner")
	}
	return loaded, nil
}

func (r *ownerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Owner{}, "owner", clinicID, id)
}

// CountPetsByOwnerID は指定されたオーナーに紐付いているペット数を返す
func (r *ownerRepository) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

// FindByLineUserID は LINE User ID で飼主を検索する（Lステップ連携用）。
func (r *ownerRepository) FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("line_user_id = ? AND deleted_at IS NULL", lineUserID).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", lineUserID)
	}
	return &owner, nil
}

// FindAllWithLineUserID は line_user_id が設定されている飼主を全件返す（タグ一括同期用）。
func (r *ownerRepository) FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error) {
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("line_user_id IS NOT NULL AND deleted_at IS NULL").
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}

// FindAllWithLineUserIDCursor は line_user_id が設定されている飼主を id カーソルページネーションで返す
// （PERF-FOLLOWUP-02）。id 昇順で afterID より大きいものを最大 limit 件返す。
func (r *ownerRepository) FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error) {
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("line_user_id IS NOT NULL AND deleted_at IS NULL AND id > ?", afterID).
		Order("id ASC").
		Limit(limit).
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}

// FindAllByLineUserID は LINE User ID で飼主を全クリニック横断検索する（Webhook用）。
func (r *ownerRepository) FindAllByLineUserID(ctx context.Context, lineUserID string) ([]model.Owner, error) {
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Where("line_user_id = ? AND deleted_at IS NULL", lineUserID).
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", lineUserID)
	}
	return owners, nil
}

// UpdateLineFollowedAt は飼主の LINE フォロー日時を更新する。
func (r *ownerRepository) UpdateLineFollowedAt(ctx context.Context, clinicID, id uint64, t time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"line_followed_at": t, "line_blocked_at": nil}).Error
	if err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return nil
}

// UpdateLineBlockedAt は飼主の LINE ブロック日時を更新する。
func (r *ownerRepository) UpdateLineBlockedAt(ctx context.Context, clinicID, id uint64, t time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("line_blocked_at", t).Error
	if err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return nil
}

// UpdateLineUserID は飼主の LINE User ID を更新する。nil を渡すと連携解除。
func (r *ownerRepository) UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	err := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("line_user_id", lineUserID).Error
	if err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *ownerRepository) FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var owners []*model.Owner
	if err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id IN ?", ids).
		Find(&owners).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}

// RecordLstepOptOut persists the owner lifecycle intent on the caller's
// ambient transaction when one exists.
func (r *ownerRepository) RecordLstepOptOut(
	ctx context.Context,
	clinicID, ownerID uint64,
	at time.Time,
	reason string,
) error {
	return repohelpers.UpdateScopedByID(
		ctx,
		repohelpers.DBOrTx(ctx, r.db),
		&model.Owner{},
		"owner",
		clinicID,
		ownerID,
		map[string]any{
			colLstepOptOut:       true,
			colLstepOptOutAt:     at,
			colLstepOptOutReason: reason,
		},
	)
}

// ClearLstepOptOut clears the owner lifecycle intent on the caller's ambient
// transaction when one exists.
func (r *ownerRepository) ClearLstepOptOut(
	ctx context.Context,
	clinicID, ownerID uint64,
) error {
	return repohelpers.UpdateScopedByID(
		ctx,
		repohelpers.DBOrTx(ctx, r.db),
		&model.Owner{},
		"owner",
		clinicID,
		ownerID,
		map[string]any{
			colLstepOptOut:       false,
			colLstepOptOutAt:     nil,
			colLstepOptOutReason: nil,
		},
	)
}
