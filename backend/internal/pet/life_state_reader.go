package pet

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// LifeStateReader は EMR-491 D6 の死亡 reader。ローカル pet から共通 subject の現行 basic revision を
// 解決し、その revision だけを根拠に生死を返す（旧 pets.status / pets.deceased_at は読まない・
// フォールバックしない）。sharedkernel.PetLifeStateReader を実装し、利用側は constructor 注入で受け取る。
//
// 解決順序（ambient transaction 必須。tx が無ければ内部エラーで fail-closed）:
//  1. 認可済み clinic のローカル pet（未削除）と immutable な clinic binding を解決する。
//     pet が無い・他院・soft-delete は not found、binding 欠落は unresolved。
//  2. binding の subject 行を FOR SHARE でロックする。subject の現行 revision ポインタを動かす
//     writer（common_basic_revision_guard が subject を FOR UPDATE する）はこのロックと直列化される。
//  3. ロック取得後に別 statement で current_basic_revision_id と revision の同一性を再解決する
//     （ロック待ち中に commit されたポインタ更新を READ COMMITTED で確実に拾う）。
//  4. revision.common_pet_id == subject.id を確認してから、その revision の status / deceased_at を読む
//     （別 subject の死亡データには触れない）。
//  5. sharedkernel.ClassifyPetLifeState で D6 真理値表を適用する。
//
// tx を解放するまで subject ロックは保持されるため、呼び出し側の write と同じ tx で呼ぶこと。
type LifeStateReader struct{}

// NewLifeStateReader は死亡 reader を返す。DB は呼び出し側の ambient transaction を使う。
func NewLifeStateReader() *LifeStateReader {
	return &LifeStateReader{}
}

var _ sharedkernel.PetLifeStateReader = (*LifeStateReader)(nil)

// lifeStateBindingRow は pets と binding の解決結果（死亡データは含めない）。
type lifeStateBindingRow struct {
	PetID       uint64     `gorm:"column:pet_id"`
	CommonPetID *uuid.UUID `gorm:"column:common_pet_id"`
}

// lifeStateSubjectRow はロック対象 subject（id のみ）。
type lifeStateSubjectRow struct {
	ID uuid.UUID `gorm:"column:id"`
}

// lifeStateRevisionIdentityRow はロック後の現行ポインタと revision の同一性（死亡データは含めない）。
type lifeStateRevisionIdentityRow struct {
	SubjectID           uuid.UUID  `gorm:"column:subject_id"`
	CurrentRevisionID   *uint64    `gorm:"column:current_basic_revision_id"`
	RevisionID          *uint64    `gorm:"column:revision_id"`
	RevisionCommonPetID *uuid.UUID `gorm:"column:revision_common_pet_id"`
}

// lifeStateRevisionRow は同一 subject と確認済みの現行 revision の生死列。
type lifeStateRevisionRow struct {
	Status     *model.PetStatus `gorm:"column:status"`
	DeceasedAt *time.Time       `gorm:"column:deceased_at"`
}

// ReadPetLifeState は clinicID 配下のローカル petID の生死を返す。
// 返り値は alive / deceased、またはエラー（not found / ErrPetLifeStateUnresolved / reader 失敗）。
func (r *LifeStateReader) ReadPetLifeState(ctx context.Context, clinicID, petID uint64) (sharedkernel.PetLifeState, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return "", apperrors.WrapInternalServerError("pet life state read requires an ambient transaction")
	}
	tx = tx.WithContext(ctx)

	commonPetID, err := resolveLifeStateBinding(tx, clinicID, petID)
	if err != nil {
		return "", err
	}
	if err := lockLifeStateSubject(tx, commonPetID); err != nil {
		return "", err
	}
	revisionID, err := resolveLockedCurrentRevision(tx, commonPetID)
	if err != nil {
		return "", err
	}

	var row lifeStateRevisionRow
	if err := tx.Table("common_pet_basic_revisions").
		Select("status", "deceased_at").
		Where("id = ? AND common_pet_id = ?", revisionID, commonPetID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", unresolvedLifeState("current revision is not readable for the subject")
		}
		return "", apperrors.Wrap(err, "failed to read pet life state revision")
	}
	return sharedkernel.ClassifyPetLifeState(row.Status, row.DeceasedAt)
}

// resolveLifeStateBinding は clinic スコープで未削除のローカル pet を引き、immutable binding の
// common_pet_id を返す。pet が無い（他院・soft-delete を含む）は not found、binding 欠落は unresolved。
func resolveLifeStateBinding(tx *gorm.DB, clinicID, petID uint64) (uuid.UUID, error) {
	var row lifeStateBindingRow
	err := tx.Table("pets AS p").
		Select("p.id AS pet_id", "b.common_pet_id AS common_pet_id").
		Joins("LEFT JOIN common_pet_clinic_bindings b ON b.clinic_id = p.clinic_id AND b.local_pet_id = p.id").
		Where("p.id = ? AND p.clinic_id = ? AND p.deleted_at IS NULL", petID, clinicID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.WrapNotFound("pet", strconv.FormatUint(petID, 10))
		}
		return uuid.Nil, apperrors.Wrap(err, "failed to resolve pet clinic binding")
	}
	if row.CommonPetID == nil {
		return uuid.Nil, unresolvedLifeState("pet has no common clinic binding")
	}
	return *row.CommonPetID, nil
}

// lockLifeStateSubject は subject 行を FOR SHARE でロックする（ambient tx の終了まで保持）。
func lockLifeStateSubject(tx *gorm.DB, commonPetID uuid.UUID) error {
	var row lifeStateSubjectRow
	err := tx.Table("common_pet_subjects").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("id").
		Where("id = ?", commonPetID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return unresolvedLifeState("common pet subject is missing")
		}
		return apperrors.Wrap(err, "failed to lock common pet subject")
	}
	return nil
}

// resolveLockedCurrentRevision はロック取得後の別 statement で現行ポインタと revision の同一性を再解決し、
// 同一 subject と確認できた revision ID を返す。WHERE は subject だけで絞り、revision が別 subject の
// ものだった場合に「行なし」へ潰れず不一致として検出できるようにする。
func resolveLockedCurrentRevision(tx *gorm.DB, commonPetID uuid.UUID) (uint64, error) {
	var row lifeStateRevisionIdentityRow
	err := tx.Table("common_pet_subjects AS s").
		Select(
			"s.id AS subject_id",
			"s.current_basic_revision_id AS current_basic_revision_id",
			"r.id AS revision_id",
			"r.common_pet_id AS revision_common_pet_id",
		).
		Joins("LEFT JOIN common_pet_basic_revisions r ON r.id = s.current_basic_revision_id").
		Where("s.id = ?", commonPetID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, unresolvedLifeState("common pet subject is missing after lock")
		}
		return 0, apperrors.Wrap(err, "failed to resolve current pet revision")
	}
	return requireSameSubjectRevision(row)
}

// requireSameSubjectRevision は現行ポインタが存在し、その revision が subject 自身のものであることを
// 検証して revision ID を返す純粋関数。
func requireSameSubjectRevision(row lifeStateRevisionIdentityRow) (uint64, error) {
	if row.CurrentRevisionID == nil {
		return 0, unresolvedLifeState("common pet subject has no current revision")
	}
	if row.RevisionID == nil || row.RevisionCommonPetID == nil {
		return 0, unresolvedLifeState("current revision row is missing")
	}
	if *row.RevisionCommonPetID != row.SubjectID {
		return 0, unresolvedLifeState("current revision belongs to another subject")
	}
	return *row.RevisionID, nil
}

// unresolvedLifeState は ErrPetLifeStateUnresolved を単一 %w で包んだ解決エラーを返す
// （reader 失敗の apperrors.Wrap とは errors.Is で区別できる）。detail に患者データは含めない。
func unresolvedLifeState(detail string) error {
	return fmt.Errorf("%s: %w", detail, sharedkernel.ErrPetLifeStateUnresolved)
}
