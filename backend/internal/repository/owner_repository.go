package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type OwnerRepository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) の飼主を検索する。
	// clinicIDs はハンドラ層で所属検証済みであること。
	FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	// FindByIDForClinics は複数医院スコープで飼主を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error)
	FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error)
	FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error)
	FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error)
	// FindByLineUserID は LINE User ID で飼主を検索する（Lステップ連携用）。
	FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
	// FindAllWithLineUserID は line_user_id が設定されている飼主を全件返す（タグ一括同期用）。
	FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error)
	// FindAllWithLineUserIDCursor は line_user_id が設定されている飼主を id カーソルページネーションで返す
	// （PERF-FOLLOWUP-02: 大規模クリニックでの無制限全件取得を回避するバッチ処理用）。
	// afterID より大きい id を昇順で最大 limit 件返す。afterID=0 で先頭ページ。
	FindAllWithLineUserIDCursor(ctx context.Context, clinicID uint64, afterID uint64, limit int) ([]model.Owner, error)
	CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	// UpdateLineUserID は飼主の LINE User ID を更新する。nil を渡すと連携解除。
	UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	// FindAllByLineUserID は LINE User ID で飼主を全クリニック横断検索する（Webhook用）。
	FindAllByLineUserID(ctx context.Context, lineUserID string) ([]model.Owner, error)
	// UpdateLineFollowedAt は飼主の LINE フォロー日時を更新する。
	UpdateLineFollowedAt(ctx context.Context, clinicID, id uint64, t time.Time) error
	// UpdateLineBlockedAt は飼主の LINE ブロック日時を更新する。
	UpdateLineBlockedAt(ctx context.Context, clinicID, id uint64, t time.Time) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// FindByIDs は複数 ID でオーナーを一括取得する（タグ一括同期の N+1 解消用）。
	// Preload なしの軽量クエリ。返り値の順序は ids の順序と一致しない場合がある。
	FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error)
}

type ownerRepository struct {
	db *gorm.DB
}

func NewOwnerRepository(db *gorm.DB) OwnerRepository {
	return &ownerRepository{db: db}
}

func (r *ownerRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners := make([]model.Owner, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return owners, 0, nil
	}

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Owner{}).Scopes(clinicScopeIn(clinicIDs))
		if search != "" {
			// NormalizeKana で検索語のカタカナをひらがなに正規化。
			// DB 列は translate() でひらがなに正規化済みのため、双方を統一して比較する。
			pattern := "%" + escapeLike(NormalizeKana(search)) + "%"
			q = q.Where(
				`(name ILIKE ? ESCAPE '\'`+
					` OR translate(name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR phone ILIKE ? ESCAPE '\'`+
					` OR email ILIKE ? ESCAPE '\')`,
				pattern,
				kanaSourceChars, kanaTargetChars, pattern,
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
		Preload("Pets", "deleted_at IS NULL").Preload("Pets.AnimalSpecies").Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").
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
// Preload する保険マスタも同じ集合で clinic 隔離する（別クリニックの保険マスタ混入防止）。
func (r *ownerRepository) findOwnerByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("owner", fmt.Sprintf("%d", id))
	}
	var owner model.Owner
	err := r.db.WithContext(ctx).
		Preload("Pets", "deleted_at IS NULL").
		Preload("Pets.AnimalSpecies").
		Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.Owner", "deleted_at IS NULL").
		Scopes(clinicScopeIn(clinicIDs)).Where("id = ?", id).First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

func (r *ownerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).First(&owner, "email = ?", email).Error
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
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).First(&owner, "phone = ?", phone).Error
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
		Scopes(clinicScope(clinicID)).
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
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// 1. 飼主を作成
		if err := tx.Create(owner).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return apperrors.WrapAlreadyExists("owner", "email already registered")
			}
			return apperrors.FromGORM(err, "owner", "")
		}
		// 2. ペットを順次作成（owner_id, clinic_id をサーバー側でセット）
		for i := range pets {
			pets[i].OwnerID = owner.ID
			pets[i].ClinicID = owner.ClinicID
			if err := tx.Create(&pets[i]).Error; err != nil {
				return apperrors.FromGORM(err, "pet", "")
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to create owner with pets")
	}
	// トランザクションコミット後に全リレーションをロードして呼び出し元に反映
	loaded, err := r.FindByID(ctx, owner.ClinicID, owner.ID)
	if err != nil {
		return apperrors.Wrap(err, "reload owner after create")
	}
	*owner = *loaded
	return nil
}

func (r *ownerRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, r.db, &model.Owner{}, "owner", clinicID, id, fields)
}

func (r *ownerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Owner{}, "owner", clinicID, id)
}

// CountPetsByOwnerID は指定されたオーナーに紐付いているペット数を返す
func (r *ownerRepository) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
		Where("line_user_id IS NOT NULL AND deleted_at IS NULL").
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}

// FindAllWithLineUserIDCursor は line_user_id が設定されている飼主を id カーソルページネーションで返す
// （PERF-FOLLOWUP-02）。id 昇順で afterID より大きいものを最大 limit 件返す。
func (r *ownerRepository) FindAllWithLineUserIDCursor(ctx context.Context, clinicID uint64, afterID uint64, limit int) ([]model.Owner, error) {
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
		Where("id IN ?", ids).
		Find(&owners).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}
