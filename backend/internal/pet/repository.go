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
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// PetListFilters はペット一覧のフィルタ条件（#266: サーバサイド pagination/search 拡張）。
// Search はペット名/カナ + 飼主名/カナ/電話に加え、空白非依存の飼主フルネーム、
// 飼主No（owners.id の文字列一致）、ペット番号（pets.pet_number）を対象に
// ILIKE + NormalizeKana で部分一致（番号は text 比較）する。
// 半角/全角/連続空白で分割した複数語は各語が既存フィールドのいずれかに一致する AND。
// 1語はフィールド間 OR（飼主フルネームの空白除去一致を含む）。
type PetListFilters struct {
	OwnerID         *uint64
	Search          string
	AnimalSpeciesID *uint64
	// IncludeDeceased: false（既定）は現行 basic revision が status='alive' かつ deceased_at IS NULL の pet（生存のみ）に絞る。
	IncludeDeceased bool
	// CheckupHistory: EMR-197-01 健診受診履歴の列挙フィルタ。空文字はフィルタ無し。
	CheckupHistory CheckupHistoryFilter
}

type Repository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) のペットを検索する。clinicIDs はハンドラ層で所属検証済みであること。
	// 順序は owner の現行 basic revision の name_kana ASC（NULL は空文字扱い）, pets.id ASC で安定（#266: ページ送りでの重複/欠落防止）。
	FindAll(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	// FindByIDForClinics は複数医院スコープでペットを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
	FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
	CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// CountLivingByOwner は指定オーナーの生存ペット数（現行 basic revision が status='alive' かつ deceased_at IS NULL）を返す。
	// ISSUE-007: CreateCheckupSync のサーバ側二重防御で誤配信を防ぐ。
	CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// CountLivingByOwnerIDs は複数オーナーの生存ペット数を一括取得する（N+1 解消用）。
	// 返り値は ownerID → 生存ペット数のマップ。存在しない ownerID は 0 として扱う。
	CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error)
	CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdatePetInput) error
	UpdateAndFind(ctx context.Context, clinicID, id uint64, update PetUpdate) (*model.Pet, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
	FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
	// OwnerHasAlivePetBirthday は指定月日が誕生日の生存ペットを飼い主が持つかを返す（EMR-330 適格性再読）。
	OwnerHasAlivePetBirthday(ctx context.Context, clinicID, ownerID uint64, month, day int) (bool, error)
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

// RowLockingRepository は同一トランザクション内の臨床状態検証を直列化するための
// FOR UPDATE 読取面（Codex PR606 #26 再指摘: cage swap の死亡判定）。
// Repository（広い読取面）には含めず、必要な消費者だけが型アサーションで取り出す
// narrow capability として CompleteRepository にだけ載せる。
type RowLockingRepository interface {
	// LockByIDForUpdate は FOR UPDATE でペットを行ロック取得する。
	// ambient tx（persistence.TxFromContext）必須 — autocommit ではロックが
	// SELECT 終了と同時に解放され直列化にならないため fail-closed（MRB-02 先例:
	// hospitalizationRepository.LockByIDForUpdate と同じ契約）。
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// CompleteRepository is the concrete composition surface returned by NewRepository.
// The legacy repository facade intentionally narrows this to Repository until central
// composition cuts over to the typed lifecycle capability.
type CompleteRepository interface {
	ServiceRepository
	LifecycleWriter
	RowLockingRepository
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

	// EMR-197-01: checkup_history の JST 窓境界は count/list で同一 now から導出する
	// （別々に time.Now() すると JST 日付またぎで total とページ行がずれ得る）。
	now := time.Now()
	if err := r.petCountQuery(ctx, clinicIDs, filters, now).Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	// BRT-70: JOIN 時の owners 列混入 scan を避けるため Find だけ pets.* を明示する。
	if err := r.petListQuery(ctx, clinicIDs, filters, now).
		Select("pets.*").
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.Paginate(page, limit)).
		// 順序の安定性: owner の現行 basic revision の name_kana ASC を主キーに、pets.id ASC を一意タイブレーカとする
		// （#266: 同一 kana でもページ送りで行の重複/欠落が起きないようにする）。
		// name_kana が NULL の飼主は、廃止前の空文字と同じ位置（先頭）に並べる。
		Order("COALESCE("+ownerBasicAlias+".name_kana, '') ASC, pets.id ASC").
		Find(&pets).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "pet", "")
	}
	// GORM の batched Preload は親 pets.clinic_id を参照できないため、
	// 認可集合スコープ後に親 clinic 相関で Owner を post-sanitize する（billing と同型）。
	for i := range pets {
		sanitizePetOwnerRelation(&pets[i])
	}
	// 現行 basic revision（pet・動物種・nested owner）を載せる。1 件でも未解決なら一覧全体を F1（422）で失敗させ、
	// 行を黙って落とさない。
	if err := hydratePetBasics(ctx, r.db, clinicIDs, commonbasic.Ptrs(pets)); err != nil {
		return nil, 0, err
	}
	return pets, total, nil
}

func (r *repository) petCountQuery(ctx context.Context, clinicIDs []uint64, filters PetListFilters, now time.Time) *gorm.DB {
	q := commonbasic.JoinCurrentPetBasic(r.db.WithContext(ctx).Model(&model.Pet{}), "pets", petBasicAlias, true).
		Where("pets.clinic_id IN ?", clinicIDs).
		Where("pets.deleted_at IS NULL")
	if filters.OwnerID != nil {
		q = q.Where("pets.owner_id = ?", *filters.OwnerID)
	}
	if filters.AnimalSpeciesID != nil {
		q = q.Where(petBasicAlias+".animal_species_id = ?", *filters.AnimalSpeciesID)
	}
	if !filters.IncludeDeceased {
		q = q.Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive)
	}
	q = applyPetCheckupHistoryFilter(q, filters.CheckupHistory, now)
	// 検索の飼主側は applyPetListSearch 内の副問い合わせで完結するため、count 側に owners JOIN は不要
	// （petListQuery 側は name_kana ソートのため owners と owner の現行 basic revision を常時 JOIN）。
	// pet 側の basic revision は種・生死の絞り込みと検索が参照するため、count 側にも JOIN する。
	return applyPetListSearch(q, filters.Search)
}

func (r *repository) petListQuery(ctx context.Context, clinicIDs []uint64, filters PetListFilters, now time.Time) *gorm.DB {
	// owners への LEFT JOIN は search 有無に関わらず常に張る
	// (owner の現行 basic revision の name_kana を安定順序の主キーにするため、Order 句が常にこれらの JOIN を要求する)。
	// clinicScopeIn は "clinic_id" を無修飾で参照し pets/owners 両方に同名列を持つため、
	// JOIN 併用時の曖昧列エラーを避けて pets.clinic_id / owners.clinic_id を明示指定する
	// （owners 側も同一 clinicIDs で二重にスコープし、クロステナント JOIN 汚染を防ぐ）。
	// BUG-454: owners.clinic_id = pets.clinic_id 相関で、認可集合に両院が含まれても
	// 破損した pet(A)->owner(B) FK を search/order の JOIN 経由で復元しない。
	// Select("pets.*") は Find 側のみ（Count に付けると COUNT("pets".*) で 42703）。
	q := commonbasic.JoinCurrentPetBasic(r.db.WithContext(ctx).Model(&model.Pet{}), "pets", petBasicAlias, true).
		Where("pets.clinic_id IN ?", clinicIDs).
		Where("pets.deleted_at IS NULL").
		Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.clinic_id IN ? AND owners.deleted_at IS NULL", clinicIDs)
	// owner の氏名カナは owners の列ではなく現行 basic revision にあるため、並び順用に LEFT JOIN で結ぶ
	// （binding・現行 revision の無い owner も行として残し、未解決は Hydrate が F1 として検出する）。
	q = commonbasic.JoinCurrentOwnerBasic(q, "owners", ownerBasicAlias, true)
	if filters.OwnerID != nil {
		q = q.Where("pets.owner_id = ?", *filters.OwnerID)
	}
	if filters.AnimalSpeciesID != nil {
		q = q.Where(petBasicAlias+".animal_species_id = ?", *filters.AnimalSpeciesID)
	}
	if !filters.IncludeDeceased {
		// 生存は現行 basic revision の status='alive' かつ deceased_at IS NULL の両方で判定する（BUG-001 の OR 契約の後継）。
		// status が NULL・revision の無い pet は生存として扱わない（既定の一覧から外れる側＝過小側に倒す）。
		q = q.Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive)
	}
	q = applyPetCheckupHistoryFilter(q, filters.CheckupHistory, now)
	return applyPetListSearch(q, filters.Search)
}

func applyPetListSearch(q *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return q
	}
	// 空白のみは fail-closed で 0 件（空フィルタ扱いで全件返さない）。
	if compactSearchText(search) == "" {
		return q.Where("1 = 0")
	}
	// strings.Fields は半角/全角/連続 Unicode 空白で分割する。
	// 複数語は各語が既存検索フィールドのいずれかに一致する AND。
	// 1語は従来のフィールド間 OR（compact は語単位。飼主+ペット連結には使わない）。
	tokens := strings.Fields(search)
	if len(tokens) == 0 {
		return q.Where("1 = 0")
	}
	for _, token := range tokens {
		q = applyPetListSearchToken(q, token)
	}
	return q
}

func applyPetListSearchToken(q *gorm.DB, token string) *gorm.DB {
	compactSearch := compactSearchText(token)
	// name/name_kana はカナ+空白を畳んだ translate() 式の1腕で検索する
	// （生値・空白差・カナ差の旧複数腕を包含し、GIN 式インデックスが効く）。
	// 空白除去形は「姓 名」入力の半角/全角/連続空白差を順序保持で吸収する（BUG-001）。
	// 飼主No は独立カラムではなく owners.id の text 一致。pet_number は文字列列。
	// いずれもユーザ入力を数値パースせずバインドする。
	qSearch := textsearch.NormalizeQuerySpaces(token)
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(qSearch)) + "%"
	compactNormalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(compactSearch)) + "%"
	trimmedSearch := strings.TrimSpace(token)
	// owners 側の腕は (owner_id, clinic_id) のペア IN で組み立てる。
	// JOIN+OR だと OR が2テーブルにまたがり GIN 式インデックスが効かず
	// JOIN 済み全行を逐行評価する（2026-10-01 障害の型）。ペア IN は内側を
	// owners の trgm インデックスで駆動し一致集合を先に計算する。
	// ペア一致は LEFT JOIN ON の id/clinic_id/deleted_at 相関と同義で、
	// BUG-454 のクロステナント FK 復元防止を維持する。
	// EMR-491: 飼主の氏名・電話は owners の列ではなく現行 basic revision にあるため、副問い合わせで
	// owners を現行 revision へ LEFT JOIN して述語を当てる（飼主No は owners.id のまま）。
	// binding・現行 revision の無い飼主は氏名・電話の腕に一致しない。
	searchedOwners := commonbasic.JoinCurrentOwnerBasic(
		q.Session(&gorm.Session{NewDB: true}).Table("owners AS searched_owner"),
		"searched_owner", searchedOwnerBasicAlias, true,
	).
		Select("searched_owner.id", "searched_owner.clinic_id").
		Where("searched_owner.deleted_at IS NULL").
		Where(
			`(`+textsearch.FoldedExpr(searchedOwnerBasicAlias+".name")+` ILIKE ? ESCAPE '\'`+
				` OR `+textsearch.FoldedExpr(searchedOwnerBasicAlias+".name_kana")+` ILIKE ? ESCAPE '\'`+
				` OR `+searchedOwnerBasicAlias+`.phone ILIKE ? ESCAPE '\'`+
				// POSIX [[:space:]] is ASCII-only; include ideographic space U+3000 so
				// stored full-width spaces match Go compactSearchText (unicode.IsSpace).
				` OR regexp_replace(`+textsearch.FoldedExpr(searchedOwnerBasicAlias+".name")+`, '`+textsearch.SpaceStripRegexp+`', '', 'g') ILIKE ? ESCAPE '\'`+
				` OR CAST(searched_owner.id AS text) = ?)`,
			normalizedPattern,
			normalizedPattern,
			normalizedPattern,
			compactNormalizedPattern,
			trimmedSearch,
		)
	return q.Where(
		`(`+textsearch.FoldedExpr(petBasicAlias+".name")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr(petBasicAlias+".name_kana")+` ILIKE ? ESCAPE '\'`+
			` OR pets.pet_number ILIKE ? ESCAPE '\'`+
			` OR (pets.owner_id, pets.clinic_id) IN (?))`,
		normalizedPattern,
		normalizedPattern,
		"%"+textsearch.EscapeLike(trimmedSearch)+"%",
		searchedOwners,
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
		Preload("Insurance", "clinic_id = ? AND deleted_at IS NULL", owner.ClinicID).
		Order("pets.created_at ASC, pets.id ASC").
		Find(&pets).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	// 現行 basic revision（名前・動物種・生死など）を載せる。1 件でも未解決なら F1（422）で、行を黙って落とさない。
	if err := hydratePetBasics(ctx, r.db, []uint64{owner.ClinicID}, commonbasic.Ptrs(pets)); err != nil {
		return nil, err
	}
	return pets, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, []uint64{clinicID}, id)
}

func (r *repository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	return r.findPetByID(ctx, clinicIDs, id)
}

// LockByIDForUpdate は FOR UPDATE でペットを行ロック取得する（RowLockingRepository）。
// Owner/Insurance の Preload は載せない — 呼び出し側が必要とするのは行スナップショット
// （pets 行のスカラーのみ。生死は pets の列ではなく sharedkernel.PetLifeStateReader が読む）で、リレーション読込はロックを広げるだけ。
func (r *repository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	// MRB-02: FOR UPDATE is meaningless outside an ambient transaction (autocommit releases
	// the row lock at statement end). Match hospitalizationRepository.LockByIDForUpdate fail-closed.
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("pet lock requires an ambient transaction")
	}
	var pet model.Pet
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	return &pet, nil
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
	// ambient tx が存在すれば同一接続で再読する（delivery-owner lock 内の
	// イベントペット最終適格性チェックは第 2 接続を要求してはならない）。
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Insurance", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).Where("id = ?", id).First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	// GORM batched Preload は親 pets.clinic_id を参照できないため post-sanitize で相関する。
	sanitizePetOwnerRelation(&pet)
	// 現行 basic revision（pet・動物種・nested owner）を載せる。未解決なら F1（422）で、空値は返さない。
	if err := hydratePetBasics(ctx, r.db, clinicIDs, []*model.Pet{&pet}); err != nil {
		return nil, err
	}
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
	// EMR-491: 生存判定は現行 basic revision の status='alive' かつ deceased_at IS NULL（D6）。
	// status が NULL・binding 欠落の pet は生存として数えない。動物種は Hydrate が revision から載せる。
	err := commonbasic.JoinCurrentPetBasic(r.db.WithContext(ctx).Model(&model.Pet{}), "pets", petBasicAlias, true).
		Select("pets.*").
		Where("pets.clinic_id = ? AND pets.owner_id = ? AND pets.deleted_at IS NULL", clinicID, ownerID).
		Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive).
		Order("pets.created_at ASC").
		Find(&pets).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	if err := hydratePetBasics(ctx, r.db, []uint64{clinicID}, commonbasic.Ptrs(pets)); err != nil {
		return nil, err
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

// CountLivingByOwner は指定オーナーの生存ペット数（現行 basic revision が status='alive' かつ deceased_at IS NULL）を返す。
// ISSUE-007: CreateCheckupSync のサーバ側二重防御で死亡ペットのみの飼い主を除外するために使用する。
// EMR-491: status が NULL・binding 欠落の pet は生存として数えない（過小側に倒す）。
func (r *repository) CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	// ambient tx が存在すれば同一接続で再読する（delivery-owner lock 内の
	// living-owner 最終適格性チェックは第 2 接続を要求してはならない）。
	if err := commonbasic.JoinCurrentPetBasic(persistence.DBOrTx(ctx, r.db).Model(&model.Pet{}), "pets", petBasicAlias, true).
		Where("pets.clinic_id = ? AND pets.owner_id = ? AND pets.deleted_at IS NULL", clinicID, ownerID).
		Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive).
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
	if err := commonbasic.JoinCurrentPetBasic(r.db.WithContext(ctx).Model(&model.Pet{}), "pets", petBasicAlias, true).
		Select("pets.owner_id AS owner_id, COUNT(*) AS count").
		Where("pets.clinic_id = ? AND pets.owner_id IN ? AND pets.deleted_at IS NULL", clinicID, ownerIDs).
		Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive).
		Group("pets.owner_id").
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}
	result := make(map[uint64]int64, len(ownerIDs))
	for _, row := range rows {
		result[row.OwnerID] = row.Count
	}
	return result, nil
}

// CountUsageByAnimalSpeciesID は動物種を参照する basic revision の行数を返す。使用中の判定は animal_species の
// 削除述語（animal_species_repository.go）と同じく、現行でない revision と soft-delete 済み pet の revision も
// 数える（履歴行が参照している限り FK で削除できないため）。
func (r *repository) CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).Model(&model.CommonPetBasicRevision{}).
		Where("animal_species_id = ?", speciesID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

// Create は既存飼主へのペット直接登録の writer が未接続の間、書き込みの前に固定 422（F6）で拒否する。
// model.Pet は基本情報の列を持たないため、pet 引数からペットの基本情報を復元して登録することもできない。
// petWriter は writer を接続するときの差し込み口として保持する。
func (*repository) Create(_ context.Context, _ *model.Pet) error {
	return errPetRegistrationPending()
}

// Update は BUG-407 の fail-closed 化（lstepLifecycleService.HandlePetDeath/HandlePetRevival が
// status/deceased_at 更新と監査書込を同一 tx で原子化する）のため dbOrTx(ctx, r.db) を使う。
// ambient tx が無い呼び出し（大多数の既存経路）では r.db.WithContext(ctx) と等価（後方互換）。
func (r *repository) Update(ctx context.Context, clinicID, id uint64, cmd UpdatePetInput) error {
	return r.update(ctx, clinicID, id, buildPetUpdate(&cmd))
}

func (r *repository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// 基本情報・生死の列は pets から廃止された。revision の writer が接続されるまで、書き込みの前に固定 422 で拒否する
	// （生死は F5、基本情報は F4）。
	if err := rejectRetiredPetFields(fields); err != nil {
		return err
	}
	for key := range fields {
		if isDangerFieldKey(key) || isStructuralPetFieldKey(key) {
			return apperrors.WrapInvalidInput("protected pet fields require the typed pet update capability")
		}
	}
	db := persistence.DBOrTx(ctx, r.db)
	return updatePetFieldsWithDB(db, clinicID, id, fields)
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
	// 基本情報・生死の列は pets から廃止された。revision の writer が接続されるまで、行ロックや書き込みの前に
	// 固定 422 で拒否する（danger_level / danger_reason も基本情報）。
	if update.dangerLevel != nil || update.dangerReason != nil {
		return nil, errPetBasicRevisionPending()
	}
	if err := rejectRetiredPetFields(update.fields); err != nil {
		return nil, err
	}
	var loaded *model.Pet
	err := withPetUpdateTransaction(ctx, r.db, func(txCtx context.Context, tx *gorm.DB) error {
		var locked model.Pet
		if err := tx.WithContext(txCtx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).
			First(&locked).Error; err != nil {
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
		}

		fields := update.fields

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

// RecordDeath は pet の死亡記録 writer が未接続の間、書き込みの前に固定 422（F5）で拒否する。
// 生死は pets の列ではなく共通 basic revision の status / deceased_at で、その改訂は revision の writer の担当。
func (*repository) RecordDeath(_ context.Context, _, _ uint64, _ time.Time, _ string) error {
	return errPetLifecycleWritePending()
}

// ClearDeath は RecordDeath と同じ理由で固定 422（F5）。
func (*repository) ClearDeath(_ context.Context, _, _ uint64) error {
	return errPetLifecycleWritePending()
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.Pet{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	return nil
}

// FindOwnersByPetBirthday は指定月日と一致する誕生日の生存ペットを持つ飼い主IDリストを返す（FEAT-383）。
// EMR-491: 生存と誕生日は現行 basic revision の status='alive' かつ deceased_at IS NULL と birth_date で判定する。
// binding・現行 revision の無い pet は生存と確認できないため対象外。
func (r *repository) FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error) {
	type row struct{ OwnerID uint64 }
	var rows []row
	err := commonbasic.JoinCurrentPetBasic(
		r.db.WithContext(ctx).
			Table("pets AS p").
			Joins("JOIN owners AS o ON o.id = p.owner_id AND o.clinic_id = p.clinic_id AND o.deleted_at IS NULL"),
		"p", petBasicAlias, true,
	).
		Where("p.clinic_id = ? AND p.deleted_at IS NULL", clinicID).
		Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive).
		Where("EXTRACT(month FROM "+petBasicAlias+".birth_date) = ? AND EXTRACT(day FROM "+petBasicAlias+".birth_date) = ?", month, day).
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

// OwnerHasAlivePetBirthday は指定月日が誕生日の生存ペットを飼い主が持つかを
// owner 単位で返す。FindOwnersByPetBirthday と同一の生存・論理削除・clinic
// 契約（EMR-290/330）を維持し、配信トリガーの claim/write 直前にイベント
// ソースを再読する EMR-330 適格性チェック用の narrow な owner スコープの読み取り。
func (r *repository) OwnerHasAlivePetBirthday(ctx context.Context, clinicID, ownerID uint64, month, day int) (bool, error) {
	var count int64
	// ambient tx が存在すれば同一接続で再読する（delivery-owner lock 内の
	// 誕生日最終適格性チェックは第 2 接続を要求してはならない）。
	err := commonbasic.JoinCurrentPetBasic(
		persistence.DBOrTx(ctx, r.db).
			Table("pets AS p").
			Joins("JOIN owners AS o ON o.id = p.owner_id AND o.clinic_id = p.clinic_id AND o.deleted_at IS NULL"),
		"p", petBasicAlias, true,
	).
		Where("p.clinic_id = ? AND p.owner_id = ? AND p.deleted_at IS NULL", clinicID, ownerID).
		Where(petLivingPredicate(petBasicAlias), model.PetStatusAlive).
		Where("EXTRACT(month FROM "+petBasicAlias+".birth_date) = ? AND EXTRACT(day FROM "+petBasicAlias+".birth_date) = ?", month, day).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "pet", fmt.Sprintf("clinic=%d owner=%d birthday=%02d-%02d", clinicID, ownerID, month, day))
	}
	return count > 0, nil
}
