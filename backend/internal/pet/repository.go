package pet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// PetListFilters はペット一覧のフィルタ条件（#266: サーバサイド pagination/search 拡張）。
// Search はペット名/カナ + 飼主名/カナ/電話に加え、空白非依存の飼主フルネーム、
// 飼主No（owners.id の文字列一致）、ペット番号（pets.pet_number）を対象に
// ILIKE + NormalizeKana で部分一致（番号は text 比較）する。
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
	UpdateAndFind(ctx context.Context, clinicID, id uint64, update PetUpdate) (*model.Pet, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
	FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
}

// OwnerReportPetRepository is the narrow persistence capability used by the Owner Report route.
type OwnerReportPetRepository interface {
	// FindOwnerReportPets は認可済み医院内の対象飼主について、Owner Report 用のペット一覧を返す。
	// 飼主とペットの clinic_id を相関させ、破損した cross-clinic owner FK を除外する。
	FindOwnerReportPets(ctx context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error)
}

// ServiceRepository is the complete persistence capability required when
// constructing the pet application service.
type ServiceRepository interface {
	Repository
	OwnerReportPetRepository
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
	ServiceRepository
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

	if err := r.petCountQuery(ctx, clinicIDs, filters).Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	// BRT-70: JOIN 時の owners 列混入 scan を避けるため Find だけ pets.* を明示する。
	if err := r.petListQuery(ctx, clinicIDs, filters).
		Select("pets.*").
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.Paginate(page, limit)).
		// 順序の安定性: owners.name_kana ASC を主キーに、pets.id ASC を一意タイブレーカとする
		// （#266: 同一 kana でもページ送りで行の重複/欠落が起きないようにする）。
		Order("owners.name_kana ASC, pets.id ASC").
		Find(&pets).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	// GORM の batched Preload は親 pets.clinic_id を参照できないため、
	// 認可集合スコープ後に親 clinic 相関で Owner を post-sanitize する（billing と同型）。
	for i := range pets {
		sanitizePetOwnerRelation(&pets[i])
	}
	return pets, total, nil
}

func (r *repository) petCountQuery(ctx context.Context, clinicIDs []uint64, filters PetListFilters) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("pets.clinic_id IN ?", clinicIDs).
		Where("pets.deleted_at IS NULL")
	if filters.OwnerID != nil {
		q = q.Where("pets.owner_id = ?", *filters.OwnerID)
	}
	if filters.AnimalSpeciesID != nil {
		q = q.Where("pets.animal_species_id = ?", *filters.AnimalSpeciesID)
	}
	if !filters.IncludeDeceased {
		q = q.Where("pets.deceased_at IS NULL AND pets.status <> ?", model.PetStatusDeceased)
	}
	if filters.Search == "" {
		return q
	}
	q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.clinic_id IN ? AND owners.deleted_at IS NULL", clinicIDs)
	return applyPetListSearch(q, filters.Search)
}

func (r *repository) petListQuery(ctx context.Context, clinicIDs []uint64, filters PetListFilters) *gorm.DB {
	// owners への LEFT JOIN は search 有無に関わらず常に張る
	// (owners.name_kana ASC を安定順序の主キーにするため、Order 句が常にこの JOIN を要求する)。
	// clinicScopeIn は "clinic_id" を無修飾で参照し pets/owners 両方に同名列を持つため、
	// JOIN 併用時の曖昧列エラーを避けて pets.clinic_id / owners.clinic_id を明示指定する
	// （owners 側も同一 clinicIDs で二重にスコープし、クロステナント JOIN 汚染を防ぐ）。
	// BUG-454: owners.clinic_id = pets.clinic_id 相関で、認可集合に両院が含まれても
	// 破損した pet(A)->owner(B) FK を search/order の JOIN 経由で復元しない。
	// Select("pets.*") は Find 側のみ（Count に付けると COUNT("pets".*) で 42703）。
	q := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("pets.clinic_id IN ?", clinicIDs).
		Where("pets.deleted_at IS NULL").
		Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.clinic_id IN ? AND owners.deleted_at IS NULL", clinicIDs)
	if filters.OwnerID != nil {
		q = q.Where("pets.owner_id = ?", *filters.OwnerID)
	}
	if filters.AnimalSpeciesID != nil {
		q = q.Where("pets.animal_species_id = ?", *filters.AnimalSpeciesID)
	}
	if !filters.IncludeDeceased {
		// deceased_at と status の両方で除外する。seed/旧データは status=deceased でも
		// deceased_at が NULL のことがあり、列片方だけ見ると死亡個体が検索に混入する（BUG-001）。
		q = q.Where("pets.deceased_at IS NULL AND pets.status <> ?", model.PetStatusDeceased)
	}
	return applyPetListSearch(q, filters.Search)
}

func applyPetListSearch(q *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return q
	}
	// 空白のみは fail-closed で 0 件（空フィルタ扱いで全件返さない）。
	compactSearch := compactSearchText(search)
	if compactSearch == "" {
		return q.Where("1 = 0")
	}
	// raw name の同一表記一致は既存の trgm index を利用可能な形で残し、
	// translate() した name/name_kana との比較でカナ表記をまたぐ一致を補う。
	// 空白除去形は「姓 名」入力の半角/全角/連続空白差を順序保持で吸収する（BUG-001）。
	// 飼主No は独立カラムではなく owners.id の text 一致。pet_number は文字列列。
	// いずれもユーザ入力を数値パースせずバインドする。
	qSearch := textsearch.NormalizeQuerySpaces(search)
	rawPattern := "%" + textsearch.EscapeLike(qSearch) + "%"
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(qSearch)) + "%"
	compactPattern := "%" + textsearch.EscapeLike(compactSearch) + "%"
	compactNormalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(compactSearch)) + "%"
	trimmedSearch := strings.TrimSpace(search)
	return q.Where(
		`(pets.name ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR owners.name ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR owners.phone ILIKE ? ESCAPE '\'`+
			// POSIX [[:space:]] is ASCII-only; include ideographic space U+3000 so
			// stored full-width spaces match Go compactSearchText (unicode.IsSpace).
			` OR regexp_replace(owners.name, '[[:space:]　]+', '', 'g') ILIKE ? ESCAPE '\'`+
			` OR regexp_replace(translate(owners.name, ?, ?), '[[:space:]　]+', '', 'g') ILIKE ? ESCAPE '\'`+
			` OR CAST(owners.id AS text) = ?`+
			` OR pets.pet_number ILIKE ? ESCAPE '\')`,
		rawPattern,
		textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		rawPattern,
		textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		normalizedPattern,
		compactPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, compactNormalizedPattern,
		trimmedSearch,
		"%"+textsearch.EscapeLike(trimmedSearch)+"%",
	)
}

func (r *repository) FindOwnerReportPets(ctx context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("owner", fmt.Sprintf("%d", ownerID))
	}

	var owner model.Owner
	if err := r.db.WithContext(ctx).
		Select("id", "clinic_id").
		Where("id = ? AND clinic_id IN ? AND deleted_at IS NULL", ownerID, clinicIDs).
		First(&owner).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", ownerID))
	}

	pets := make([]model.Pet, 0)
	// BRT-70: JOIN 併用時は pets.* を明示（owners 列の混入 scan を避ける）
	if err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Select("pets.*").
		Joins("INNER JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.deleted_at IS NULL").
		Where("pets.owner_id = ? AND pets.clinic_id = ? AND pets.deleted_at IS NULL", ownerID, owner.ClinicID).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id = ? AND deleted_at IS NULL", owner.ClinicID).
		Order("pets.created_at ASC, pets.id ASC").
		Find(&pets).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	return pets, nil
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
// BUG-454: 認可集合に両院が含まれても pet.clinic_id と一致しない Owner は復元しない。
func (r *repository) findPetByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	var pet model.Pet
	err := r.db.WithContext(ctx).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).Where("id = ?", id).First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	// GORM batched Preload は親 pets.clinic_id を参照できないため post-sanitize で相関する。
	sanitizePetOwnerRelation(&pet)
	return &pet, nil
}

// sanitizePetOwnerRelation enforces exact pet→owner clinic correlation after GORM's
// batched Owner preload. Multi-clinic authorization may safely query the authorized
// clinic set, but a pet from clinic A must never attach an owner from clinic B
// (BUG-454: multi-clinic viewers must not restore a broken cross-clinic FK graph).
func sanitizePetOwnerRelation(pet *model.Pet) {
	if pet.Owner != nil && pet.Owner.ClinicID != pet.ClinicID {
		pet.Owner = nil
	}
}

// compactSearchText removes all Unicode whitespace while preserving character order.
// Used only to derive a comparison form for BUG-001 owner-name matching; does not
// mutate the original filter value.
func compactSearchText(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (r *repository) FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
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
	for key := range fields {
		if isDangerFieldKey(key) || isStructuralPetFieldKey(key) {
			return apperrors.WrapInvalidInput("protected pet fields require the typed pet update capability")
		}
	}
	db := persistence.DBOrTx(ctx, r.db)
	if expectedStatus, conflictMessage, ok := legacyLifecycleTransition(fields); ok {
		return updateLegacyLifecycleFieldsWithDB(db, clinicID, id, expectedStatus, fields, conflictMessage)
	}
	return updatePetFieldsWithDB(db, clinicID, id, fields)
}

func legacyLifecycleTransition(fields map[string]any) (model.PetStatus, string, bool) {
	if len(fields) != 3 {
		return "", "", false
	}
	if _, ok := fields["deceased_at"]; !ok {
		return "", "", false
	}
	if _, ok := fields["deceased_reason"]; !ok {
		return "", "", false
	}
	status, ok := fields["status"]
	if !ok {
		return "", "", false
	}

	switch status {
	case model.PetStatusDeceased:
		return model.PetStatusAlive, "死亡記録は既に登録されています", true
	case model.PetStatusAlive:
		return model.PetStatusDeceased, "死亡記録が登録されていないため解除できません", true
	default:
		return "", "", false
	}
}

func updateLegacyLifecycleFieldsWithDB(
	db *gorm.DB,
	clinicID, petID uint64,
	expectedStatus model.PetStatus,
	fields map[string]any,
	conflictMessage string,
) error {
	result := db.
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status = ?", petID, expectedStatus).
		Select("deceased_at", "deceased_reason", "status").
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", petID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict(conflictMessage)
	}
	return nil
}

func isDangerFieldKey(key string) bool {
	switch key {
	case "danger_level", "DangerLevel", "danger_reason", "DangerReason":
		return true
	default:
		return false
	}
}

func isStructuralPetFieldKey(key string) bool {
	switch key {
	case "clinic_id", "ClinicID", "owner_id", "OwnerID", "insurance_id", "InsuranceID":
		return true
	default:
		return false
	}
}

func updatePetFieldsWithDB(db *gorm.DB, clinicID, id uint64, fields map[string]any) error {
	result := db.
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := db.Model(&model.Pet{}).
			Scopes(persistence.ClinicScope(clinicID)).
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

func (r *repository) UpdateAndFind(
	ctx context.Context,
	clinicID, id uint64,
	update PetUpdate,
) (*model.Pet, error) {
	var loaded *model.Pet
	err := withPetUpdateTransaction(ctx, r.db, func(txCtx context.Context, tx *gorm.DB) error {
		var locked model.Pet
		if err := tx.WithContext(txCtx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).
			First(&locked).Error; err != nil {
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
		}

		effectiveLevel := locked.DangerLevel
		if update.dangerLevel != nil {
			effectiveLevel = *update.dangerLevel
		}
		effectiveReason := locked.DangerReason
		if update.dangerReason != nil {
			effectiveReason = *update.dangerReason
		}
		normalizedReason, err := normalizeDangerReason(effectiveLevel, effectiveReason)
		if err != nil {
			return err
		}

		fields := make(map[string]any, len(update.fields))
		for key, value := range update.fields {
			if isDangerFieldKey(key) {
				continue
			}
			fields[key] = value
		}
		if update.dangerLevel != nil {
			fields["danger_level"] = effectiveLevel
		}
		if update.dangerReason != nil {
			fields["danger_reason"] = normalizedReason
		}

		// POC-03 / X-05: re-validate request-derived clinic FKs inside the same
		// write transaction with SHARE locks so concurrent soft-delete cannot
		// race past the pre-tx service checks.
		if err := revalidatePetUpdateForeignKeys(tx.WithContext(txCtx), clinicID, fields); err != nil {
			return err
		}

		if err := updatePetFieldsWithDB(tx.WithContext(txCtx), clinicID, id, fields); err != nil {
			return err
		}
		loaded, err = loadPetGraph(txCtx, tx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "reload pet after update")
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.Wrap(err, "transaction failed"),
			"failed to update and reload pet",
		)
	}
	return loaded, nil
}

func withPetUpdateTransaction(
	ctx context.Context,
	db *gorm.DB,
	fn func(context.Context, *gorm.DB) error,
) error {
	return persistence.DBOrTx(ctx, db).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx), tx)
	})
}

func revalidatePetUpdateForeignKeys(db *gorm.DB, clinicID uint64, fields map[string]any) error {
	if rawOwner, ok := fields[colPetOwnerID]; ok {
		ownerID, err := uint64FromField(rawOwner)
		if err != nil {
			return apperrors.WrapInvalidInput("owner not found in this clinic")
		}
		if err := lockOwnerForRegistration(db, clinicID, ownerID); err != nil {
			return err
		}
	}
	if rawInsurance, ok := fields[colPetInsuranceID]; ok {
		if rawInsurance == nil {
			return nil
		}
		switch v := rawInsurance.(type) {
		case *uint64:
			if v == nil {
				return nil
			}
			return lockInsuranceForPetUpdate(db, clinicID, *v)
		default:
			insuranceID, err := uint64FromField(rawInsurance)
			if err != nil {
				return apperrors.WrapInvalidInput("insurance not found in this clinic")
			}
			return lockInsuranceForPetUpdate(db, clinicID, insuranceID)
		}
	}
	return nil
}

func uint64FromField(raw any) (uint64, error) {
	switch v := raw.(type) {
	case uint64:
		return v, nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative id")
		}
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("negative id")
		}
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("unsupported id type %T", raw)
	}
}

func lockInsuranceForPetUpdate(db *gorm.DB, clinicID, insuranceID uint64) error {
	var insurance model.Insurance
	err := db.
		Select("id", "clinic_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", insuranceID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		First(&insurance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.WrapInvalidInput("insurance not found in this clinic")
	}
	if err != nil {
		return apperrors.FromGORM(err, "insurance", "")
	}
	return nil
}

func (r *repository) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status = ?", petID, model.PetStatusAlive).
		Select("deceased_at", "deceased_reason", "status").
		Updates(&model.Pet{
			DeceasedAt:     &deceasedAt,
			DeceasedReason: &reason,
			Status:         model.PetStatusDeceased,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", petID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict("死亡記録は既に登録されています")
	}
	return nil
}

func (r *repository) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status = ?", petID, model.PetStatusDeceased).
		Select("deceased_at", "deceased_reason", "status").
		Updates(&model.Pet{Status: model.PetStatusAlive})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", petID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict("死亡記録が登録されていないため解除できません")
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.Pet{})
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
