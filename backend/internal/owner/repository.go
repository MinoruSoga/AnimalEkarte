package owner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// OwnerUpdateApplier builds the typed update command from the FOR UPDATE locked row.
// Returning an error aborts the transaction without writing.
type OwnerUpdateApplier func(locked *model.Owner) (UpdateCommand, error)

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
	UpdateAndFind(ctx context.Context, clinicID, id uint64, cmd UpdateCommand) (*model.Owner, error)
	// UpdateAndFindApplying locks the owner FOR UPDATE, lets apply build a command from
	// the locked snapshot, then updates and reloads (SEC-CS-F15 discount recheck).
	UpdateAndFindApplying(ctx context.Context, clinicID, id uint64, apply OwnerUpdateApplier) (*model.Owner, error)
	// LockByIDForUpdate locks a full clinic-scoped owner row. Fail-closed without ambient tx.
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

// LookupRepository contains owner lookups consumed outside the owner use case.
type LookupRepository interface {
	FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error)
	FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error)
}

// LifecycleOwnerRepository contains owner persistence used by owner-lifecycle
// workflows (LINE follow/block, delivery opt-out).
type LifecycleOwnerRepository interface {
	FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error)
	FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error)
	LockLineLinkOwner(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
	UpdateLineFollowedAt(
		ctx context.Context,
		clinicID, id uint64,
		expectedLineUserID string,
		eventAt time.Time,
	) (bool, error)
	UpdateLineBlockedAt(
		ctx context.Context,
		clinicID, id uint64,
		expectedLineUserID string,
		eventAt time.Time,
	) (bool, error)
	RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error
	ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error
}

// Repository is the complete owner persistence capability used at the
// composition boundary. Application services accept narrower interfaces.
type Repository interface {
	ServiceRepository
	LookupRepository
	LifecycleOwnerRepository
	OwnerDeleteLocker
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

// ownerBasicAlias は検索で結ぶ現行 owner basic revision の JOIN 別名。
const ownerBasicAlias = "owner_basic"

func (r *ownerRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners := make([]model.Owner, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return owners, 0, nil
	}

	buildBase := func() *gorm.DB {
		// owners の列は常に修飾する（clinic_id は binding、created_at は subject の列と衝突する）。
		q := r.db.WithContext(ctx).Model(&model.Owner{}).Where("owners.clinic_id IN ?", clinicIDs)
		if search != "" {
			// 氏名・連絡先は owner の現行 basic revision の列なので、検索のときだけ LEFT JOIN で結ぶ。
			// binding・現行 revision の無い owner は述語に一致せず検索結果に現れない
			// （検索なしの一覧は JOIN せず、Hydrate が未解決の owner を unresolved として検出する）。
			// name / name_kana はカナ+空白を畳んだ translate() 式1腕で検索する
			// （生値・空白差・カナ差の旧複数腕を包含する）。
			// name の空白除去腕は姓名の空白有無を吸収する。
			// phone と email は従来どおり正規化済み pattern で比較する。
			// 空白のみは fail-closed で 0 件 (BUG-001)。
			qSearch := textsearch.NormalizeQuerySpaces(search)
			if qSearch == "" {
				q = q.Where("1 = 0")
				return q
			}
			normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(qSearch)) + "%"
			compactPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(textsearch.CompactSearchText(search))) + "%"
			nameExpr := textsearch.FoldedExpr(ownerBasicAlias + ".name")
			q = commonbasic.JoinCurrentOwnerBasic(q, "owners", ownerBasicAlias, true).Where(
				`(`+nameExpr+` ILIKE ? ESCAPE '\'`+
					` OR `+textsearch.FoldedExpr(ownerBasicAlias+".name_kana")+` ILIKE ? ESCAPE '\'`+
					` OR `+ownerBasicAlias+`.phone ILIKE ? ESCAPE '\'`+
					` OR `+ownerBasicAlias+`.email ILIKE ? ESCAPE '\'`+
					` OR regexp_replace(`+nameExpr+`, '`+textsearch.SpaceStripRegexp+`', '', 'g') ILIKE ? ESCAPE '\')`,
				normalizedPattern,
				normalizedPattern,
				normalizedPattern,
				normalizedPattern,
				compactPattern,
			)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "owner", "")
	}
	if err := buildBase().
		Select("owners.*").
		Preload("Pets", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.Paginate(page, limit)).Order("owners.created_at DESC").
		Find(&owners).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "owner", "")
	}
	if err := hydrateOwnerBasics(ctx, r.db, clinicIDs, commonbasic.Ptrs(owners)); err != nil {
		return nil, 0, err
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
	return r.findOwnerByIDWithDB(ctx, persistence.DBOrTx(ctx, r.db), clinicIDs, id)
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
		Preload("Pets.Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pets.Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).Where("id = ?", id).First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	if err := hydrateOwnerBasics(ctx, db, clinicIDs, []*model.Owner{&owner}); err != nil {
		return nil, err
	}
	return &owner, nil
}

// hydrateOwnerBasics は owners と入れ子の pets（動物種を含む）へ現行 basic revision を載せる。
// 1 件でも解決できなければ全体が unresolved（422）で、空値や旧列へのフォールバックはしない。
// 入れ子の pets は owners と同じ clinicIDs で preload 済みであること。
func hydrateOwnerBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, owners []*model.Owner) error {
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, owners); err != nil {
		return err
	}
	var pets []*model.Pet
	for _, owner := range owners {
		for i := range owner.Pets {
			pets = append(pets, &owner.Pets[i])
		}
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, pets)
}

func (r *ownerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	return r.findFirstByBasicColumn(ctx, clinicID, "email", email)
}

// FindByPhone は clinic_id + phone に一致するオーナーを返す。見つからない場合は nil を返す。
func (r *ownerRepository) FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error) {
	return r.findFirstByBasicColumn(ctx, clinicID, "phone", phone)
}

// findFirstByBasicColumn は現行 basic revision の column が value に一致する最初の飼主を返す。
// 見つからなければ nil を返す。column は SQL へ直接埋め込むため、コンパイル時定数だけを渡すこと。
func (r *ownerRepository) findFirstByBasicColumn(ctx context.Context, clinicID uint64, column, value string) (*model.Owner, error) {
	var owner model.Owner
	err := commonbasic.JoinCurrentOwnerBasic(r.db.WithContext(ctx).Model(&model.Owner{}), "owners", ownerBasicAlias, true).
		Select("owners.*").
		Where("owners.clinic_id = ?", clinicID).
		Where(ownerBasicAlias+"."+column+" = ?", value).
		First(&owner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "owner", value)
	}
	if err := commonbasic.HydrateOwners(ctx, r.db, []uint64{clinicID}, []*model.Owner{&owner}); err != nil {
		return nil, err
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
	err := commonbasic.JoinCurrentOwnerBasic(r.db.WithContext(ctx).Model(&model.Owner{}), "owners", ownerBasicAlias, true).
		Select("owners.*").
		Where("owners.clinic_id = ?", clinicID).
		Where(ownerBasicAlias+".name = ? AND "+ownerBasicAlias+".phone = ? AND owners.deleted_at IS NULL", name, phone).
		Limit(2). // 2件以上あるかだけ判定すればよい
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%s/%s", name, phone))
	}
	if len(owners) != 1 {
		return nil, nil // 0件 or 複数件 → 自動紐付け不可
	}
	if err := commonbasic.HydrateOwners(ctx, r.db, []uint64{clinicID}, []*model.Owner{&owners[0]}); err != nil {
		return nil, err
	}
	return &owners[0], nil
}

// CreateWithPets は飼主とペットの新規登録 writer が未接続の間、書き込みの前に固定 422（F3）で拒否する。
// owners 行だけを作ると共通 subject・binding・basic revision の無い孤児になり、以降の読み取りがすべて
// unresolved になる。model.Pet は基本情報の列を持たないため、pets 引数からペットの基本情報も復元できない。
// writer を接続するときは、PetRegistrationDraft を受け取る経路へ置き換える。
func (r *ownerRepository) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return errOwnerRegistrationPending()
}

func (r *ownerRepository) UpdateAndFind(
	ctx context.Context,
	clinicID, id uint64,
	cmd UpdateCommand,
) (*model.Owner, error) {
	return r.UpdateAndFindApplying(ctx, clinicID, id, func(_ *model.Owner) (UpdateCommand, error) {
		return cmd, nil
	})
}

// LockByIDForUpdate は clinic-scoped FOR UPDATE で owner 全カラムを固定する（SEC-CS-F15）。
func (r *ownerRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("owner lock requires an ambient transaction")
	}
	var owner model.Owner
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

func (r *ownerRepository) UpdateAndFindApplying(
	ctx context.Context,
	clinicID, id uint64,
	apply OwnerUpdateApplier,
) (*model.Owner, error) {
	if apply == nil {
		return nil, apperrors.WrapInternalServerError("owner update applier is required")
	}
	var loaded *model.Owner
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		locked, err := r.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return err
		}
		cmd, err := apply(locked)
		if err != nil {
			return err
		}
		fields := updateCommandFields(cmd)
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput("at least one field must be provided")
		}
		if err := persistence.UpdateScopedByID(
			txCtx,
			tx,
			&model.Owner{},
			"owner",
			clinicID,
			id,
			fields,
		); err != nil {
			if mapped := mapOwnerUniqueConstraintErr(err); mapped != nil {
				return mapped
			}
			return err
		}

		loaded, err = r.findOwnerByIDWithDB(txCtx, tx, []uint64{clinicID}, id)
		if err != nil {
			return apperrors.Wrap(err, "reload owner after update")
		}
		return nil
	})
	if err != nil {
		if mapped := mapOwnerUniqueConstraintErr(err); mapped != nil {
			return nil, mapped
		}
		// Preserve Forbidden/InvalidInput from apply without double-wrapping.
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, apperrors.Wrap(err, "failed to update and reload owner")
	}
	return loaded, nil
}

// mapOwnerUniqueConstraintErr maps owners partial unique indexes (email/phone)
// to AlreadyExists. POC-06 / U-X05-OWNER-PHONE.
// BUG-024: client-facing JA message (no English "%s '%s' already exists" template).
func mapOwnerUniqueConstraintErr(err error) error {
	if err == nil || !persistence.IsUniqueConstraintErr(err) {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "uk_owners_clinic_phone":
			// BUG-019 / BUG-024: クライアント表示用の自然な日本語（英語テンプレ埋め込み禁止）
			return apperrors.WrapAlreadyExistsMessage("この電話番号はすでに登録されています")
		case "uk_owners_clinic_email":
			return apperrors.WrapAlreadyExistsMessage("このメールアドレスはすでに登録されています")
		}
	}
	// Constraint name missing (driver wrap) — still fail closed as already-exists.
	return apperrors.WrapAlreadyExistsMessage("この飼主はすでに登録されています")
}

func (r *ownerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
		&model.Owner{},
		"owner",
		clinicID,
		id,
	)
}

// CountPetsByOwnerID は指定されたオーナーに紐付いているペット数を返す
func (r *ownerRepository) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

// LockForDelete serializes deletion with writers that take a row lock on an
// active, clinic-scoped owner. It fails closed without the service-owned
// ambient transaction.
func (r *ownerRepository) LockForDelete(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Owner, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("owner delete lock requires an ambient transaction")
	}
	var owner model.Owner
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "clinic_id", "line_user_id").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

// FindByLineUserID は LINE User ID で飼主を検索する（Lステップ連携用）。
func (r *ownerRepository) FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
	var owner model.Owner
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_user_id = ? AND deleted_at IS NULL", lineUserID).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return &owner, nil
}

// FindAllWithLineUserID は line_user_id が設定されている飼主を全件返す（タグ一括同期用）。
func (r *ownerRepository) FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error) {
	var owners []model.Owner
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_user_id IS NOT NULL AND deleted_at IS NULL AND id > ?", afterID).
		Order("id ASC").
		Limit(limit).
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return owners, nil
}

// UpdateLineFollowedAt は、署名元 clinic と現在の LINE User ID が一致し、
// eventAt が既存の follow/block 時刻より新しい場合だけフォロー日時を更新する。
// 更新対象がない（重複、古いイベント、再連携済み）場合は (false, nil) を返す。
func (r *ownerRepository) UpdateLineFollowedAt(
	ctx context.Context,
	clinicID, id uint64,
	expectedLineUserID string,
	eventAt time.Time,
) (bool, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Owner{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND line_user_id = ? AND deleted_at IS NULL", id, expectedLineUserID).
		Where("(line_followed_at IS NULL OR line_followed_at < ?)", eventAt).
		Where("(line_blocked_at IS NULL OR line_blocked_at < ?)", eventAt).
		Updates(map[string]any{"line_followed_at": eventAt, "line_blocked_at": nil})
	if result.Error != nil {
		return false, apperrors.FromGORM(result.Error, "owner", fmt.Sprintf("%d", id))
	}
	return result.RowsAffected == 1, nil
}

// UpdateLineBlockedAt は、署名元 clinic と現在の LINE User ID が一致し、
// eventAt が follow 以上かつ既存 block より新しい場合だけブロック日時を更新する。
// 同時刻の follow/unfollow では unfollow を優先し、更新対象がなければ (false, nil) を返す。
func (r *ownerRepository) UpdateLineBlockedAt(
	ctx context.Context,
	clinicID, id uint64,
	expectedLineUserID string,
	eventAt time.Time,
) (bool, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Owner{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND line_user_id = ? AND deleted_at IS NULL", id, expectedLineUserID).
		Where("(line_followed_at IS NULL OR line_followed_at <= ?)", eventAt).
		Where("(line_blocked_at IS NULL OR line_blocked_at < ?)", eventAt).
		Update("line_blocked_at", eventAt)
	if result.Error != nil {
		return false, apperrors.FromGORM(result.Error, "owner", fmt.Sprintf("%d", id))
	}
	return result.RowsAffected == 1, nil
}

// LockLineLinkOwner serializes public LINE account linking for one clinic-scoped
// owner. It fails closed without the service-owned ambient transaction.
func (r *ownerRepository) LockLineLinkOwner(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Owner, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("LINE link owner lock requires an ambient transaction")
	}
	var owner model.Owner
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &owner, nil
}

// UpdateLineUserID は飼主の LINE User ID の更新（連携・解除）を受け付ける。
// owners.line_user_id は drop 済みで、置換先 owner_line_binding_states の writer はこの branch に無い。
// 成功を装わないため、飼主の存在確認だけ行い、書き込みの前に line_binding_pending（422）で停止する。
func (r *ownerRepository) UpdateLineUserID(ctx context.Context, clinicID, id uint64, _ *string) error {
	var owner model.Owner
	if err := persistence.DBOrTx(ctx, r.db).
		Select("id").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&owner).Error; err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return apperrors.WrapUnprocessableWithCode(apperrors.CodeLineBindingPending, "LINE 連携の保存は準備中のため、この操作はできません")
}

func (r *ownerRepository) FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*model.Owner, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var owners []*model.Owner
	if err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id IN ?", ids).
		Find(&owners).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	if err := commonbasic.HydrateOwners(ctx, r.db, []uint64{clinicID}, owners); err != nil {
		return nil, err
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
	return persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
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
	return persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
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
