package pet

// life_state_reader_test.go — EMR-491 D6 死亡 reader（LifeStateReader）のテスト。
//
// 検証する契約:
//   - ambient transaction 必須（tx 外は内部エラーで fail closed）
//   - pet 解決 → binding → subject FOR SHARE ロック → ロック後に現行 revision を再解決 →
//     revision.common_pet_id == subject.id → D6 真理値表
//   - pet 欠落 / 他院 / soft-delete は NotFound、binding・subject・現行 revision の欠落と別 subject の
//     revision は ErrPetLifeStateUnresolved（reader 失敗と区別でき、別 subject の死亡データは読まない）
//   - subject の FOR SHARE ロックは writer の FOR UPDATE と直列化され、ロック待ち中に commit された
//     ポインタ更新を再解決して拾う（stale な alive を返さない）
//
// 注: テスト DB は GORM AutoMigrate の double（SetupTestDB）。001_init.sql の複合 FK
// fk_common_pet_current_basic（DEFERRABLE INITIALLY DEFERRED）や publish trigger は存在しないため、
// 「別 subject の revision を指す」「subject / revision 欠落」の不整合状態を制約・trigger を無効化せずに
// 組み立てられる。本番 DDL ではその不整合は commit 時に FK で拒否されるため、reader のこの検査は
// defense in depth である。

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const lifeStateClinicA, lifeStateClinicB = uint64(1), uint64(2)

func lifeStatePtr(status model.PetStatus) *model.PetStatus { return &status }

func setupLifeStateDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{},
		&model.CommonPetSubject{}, &model.CommonPetClinicBinding{}, &model.CommonPetBasicRevision{},
	))
	testdb.Truncate(t, db,
		"common_pet_basic_revisions", "common_pet_clinic_bindings", "common_pet_subjects",
		"pets", "animal_species", "owners",
	)
	return db
}

func makeLifeStateSpecies(t *testing.T, db *gorm.DB) *model.AnimalSpecies {
	t.Helper()
	species := &model.AnimalSpecies{Name: "life-state-species"}
	require.NoError(t, db.Create(species).Error)
	return species
}

func makeLifeStatePet(t *testing.T, db *gorm.DB, clinicID uint64, tag string) *model.Pet {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, tag+"-owner")
	pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, PetNumber: tag}
	require.NoError(t, db.Create(pet).Error)
	return pet
}

func createLifeStateSubject(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&model.CommonPetSubject{ID: id}).Error)
	return id
}

func bindLifeStatePet(t *testing.T, db *gorm.DB, clinicID uint64, pet *model.Pet, commonID uuid.UUID) {
	t.Helper()
	require.NoError(t, db.Create(&model.CommonPetClinicBinding{
		ClinicID: clinicID, CommonPetID: commonID, LocalPetID: pet.ID,
		ConfirmedBy: 1, ConfirmedAt: time.Now(), OperationID: 1,
	}).Error)
}

func addLifeStateRevision(
	t *testing.T,
	db *gorm.DB,
	commonID uuid.UUID,
	clinicID uint64,
	localPetID uint64,
	speciesID uint64,
	revision int64,
	status *model.PetStatus,
	deceasedAt *time.Time,
) *model.CommonPetBasicRevision {
	t.Helper()
	rev := &model.CommonPetBasicRevision{
		CommonPetID: commonID, Revision: revision,
		SourceClinicID: clinicID, SourceLocalPetID: localPetID, OperationID: 1,
		Name: "life-state", AnimalSpeciesID: speciesID,
		Status: status, DeceasedAt: deceasedAt,
		NeuteredStatus: model.PetNeuteredStatusUnknown,
		Provenance:     json.RawMessage(`{}`),
		RecordedBy:     1, RecordedAt: time.Now(),
	}
	require.NoError(t, db.Create(rev).Error)
	return rev
}

func setLifeStatePointer(t *testing.T, db *gorm.DB, commonID uuid.UUID, revisionID uint64) {
	t.Helper()
	require.NoError(t, db.Model(&model.CommonPetSubject{}).
		Where("id = ?", commonID).
		Update("current_basic_revision_id", revisionID).Error)
}

// linkLifeStatePetWithRevision は subject + binding + 現行 revision を一式作る。
func linkLifeStatePetWithRevision(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	pet *model.Pet,
	speciesID uint64,
	status *model.PetStatus,
	deceasedAt *time.Time,
) (uuid.UUID, *model.CommonPetBasicRevision) {
	t.Helper()
	commonID := createLifeStateSubject(t, db)
	bindLifeStatePet(t, db, clinicID, pet, commonID)
	rev := addLifeStateRevision(t, db, commonID, clinicID, pet.ID, speciesID, 1, status, deceasedAt)
	setLifeStatePointer(t, db, commonID, rev.ID)
	return commonID, rev
}

// readLifeStateInTx は ambient transaction 内で reader を 1 回呼び、tx は常に rollback する。
func readLifeStateInTx(
	t *testing.T,
	db *gorm.DB,
	clinicID, petID uint64,
) (sharedkernel.PetLifeState, error) {
	t.Helper()
	var (
		state sharedkernel.PetLifeState
		err   error
	)
	rollback := errors.New("rollback after read")
	txErr := db.Transaction(func(tx *gorm.DB) error {
		state, err = NewLifeStateReader().ReadPetLifeState(persistence.WithTxValue(context.Background(), tx), clinicID, petID)
		return rollback
	})
	require.ErrorIs(t, txErr, rollback)
	return state, err
}

func TestLifeStateReader_RequiresAmbientTransaction(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	pet := makeLifeStatePet(t, db, lifeStateClinicA, "no-tx")
	linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, lifeStatePtr(model.PetStatusAlive), nil)

	state, err := NewLifeStateReader().ReadPetLifeState(context.Background(), lifeStateClinicA, pet.ID)

	require.Error(t, err)
	assert.Empty(t, state, "no life state may be returned without a transaction")
	assert.False(t, apperrors.IsNotFound(err), "missing tx is an internal error, not not-found: %v", err)
	assert.False(t, errors.Is(err, sharedkernel.ErrPetLifeStateUnresolved))
}

func TestLifeStateReader_D6TruthTable(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name       string
		status     *model.PetStatus
		deceasedAt *time.Time
		want       sharedkernel.PetLifeState
		wantErr    bool
	}{
		{name: "alive / nil date", status: lifeStatePtr(model.PetStatusAlive), want: sharedkernel.PetLifeStateAlive},
		{name: "deceased / nil date", status: lifeStatePtr(model.PetStatusDeceased), want: sharedkernel.PetLifeStateDeceased},
		{name: "deceased / dated", status: lifeStatePtr(model.PetStatusDeceased), deceasedAt: &deceasedAt, want: sharedkernel.PetLifeStateDeceased},
		{name: "alive + date is deceased", status: lifeStatePtr(model.PetStatusAlive), deceasedAt: &deceasedAt, want: sharedkernel.PetLifeStateDeceased},
		{name: "NULL status + date is deceased", deceasedAt: &deceasedAt, want: sharedkernel.PetLifeStateDeceased},
		{name: "NULL status + nil date is unresolved", wantErr: true},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pet := makeLifeStatePet(t, db, lifeStateClinicA, "truth-"+string(rune('a'+i)))
			linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, tc.status, tc.deceasedAt)

			state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)

			if tc.wantErr {
				require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
				assert.Empty(t, state)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, state)
		})
	}
}

func TestLifeStateReader_ResolutionFailures(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	alive := lifeStatePtr(model.PetStatusAlive)

	t.Run("missing pet is not found", func(t *testing.T) {
		state, err := readLifeStateInTx(t, db, lifeStateClinicA, 987654)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)
		assert.Empty(t, state)
	})

	t.Run("other clinic is not found and never reveals the pet", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "other-clinic")
		linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, alive, nil)

		state, err := readLifeStateInTx(t, db, lifeStateClinicB, pet.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)
		assert.False(t, errors.Is(err, sharedkernel.ErrPetLifeStateUnresolved), "cross-clinic must look like not found")
		assert.Empty(t, state)
	})

	t.Run("soft-deleted pet is not found", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "soft-deleted")
		linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, alive, nil)
		require.NoError(t, db.Delete(&model.Pet{}, pet.ID).Error)

		state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)
		assert.Empty(t, state)
	})

	t.Run("missing binding is unresolved", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "no-binding")

		state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)
		require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
		assert.False(t, apperrors.IsNotFound(err))
		assert.Empty(t, state)
	})

	t.Run("binding without subject is unresolved", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "no-subject")
		bindLifeStatePet(t, db, lifeStateClinicA, pet, uuid.New())

		state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)
		require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
		assert.Empty(t, state)
	})

	t.Run("subject without current revision is unresolved", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "no-current")
		commonID := createLifeStateSubject(t, db)
		bindLifeStatePet(t, db, lifeStateClinicA, pet, commonID)
		// 履歴 revision はあるがポインタが NULL: 最新の revision を推測して読まない。
		addLifeStateRevision(t, db, commonID, lifeStateClinicA, pet.ID, species.ID, 1, alive, nil)

		state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)
		require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
		assert.Empty(t, state)
	})

	t.Run("pointer to a missing revision is unresolved", func(t *testing.T) {
		pet := makeLifeStatePet(t, db, lifeStateClinicA, "dangling-pointer")
		commonID := createLifeStateSubject(t, db)
		bindLifeStatePet(t, db, lifeStateClinicA, pet, commonID)
		setLifeStatePointer(t, db, commonID, 999999)

		state, err := readLifeStateInTx(t, db, lifeStateClinicA, pet.ID)
		require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
		assert.Empty(t, state)
	})
}

// 現行ポインタが別 subject の revision を指す場合は、その revision の生死が alive でも deceased でも
// 採用せず unresolved にする（別 subject の死亡データを読まない・流用しない）。
func TestLifeStateReader_OtherSubjectRevisionIsNeverUsed(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name         string
		ownStatus    model.PetStatus
		ownDeceased  *time.Time
		otherStatus  model.PetStatus
		otherDeceasd *time.Time
	}{
		{name: "own deceased, pointer to other subject alive revision", ownStatus: model.PetStatusDeceased, ownDeceased: &deceasedAt, otherStatus: model.PetStatusAlive},
		{name: "own alive, pointer to other subject deceased revision", ownStatus: model.PetStatusAlive, otherStatus: model.PetStatusDeceased, otherDeceasd: &deceasedAt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			own := makeLifeStatePet(t, db, lifeStateClinicA, "own-"+tc.name)
			commonID, ownRev := linkLifeStatePetWithRevision(t, db, lifeStateClinicA, own, species.ID, lifeStatePtr(tc.ownStatus), tc.ownDeceased)
			require.NotZero(t, ownRev.ID)

			other := makeLifeStatePet(t, db, lifeStateClinicA, "other-"+tc.name)
			_, otherRev := linkLifeStatePetWithRevision(t, db, lifeStateClinicA, other, species.ID, lifeStatePtr(tc.otherStatus), tc.otherDeceasd)
			require.NotEqual(t, ownRev.ID, otherRev.ID)

			// own subject のポインタを他 subject の revision へ向ける（不整合状態）。
			setLifeStatePointer(t, db, commonID, otherRev.ID)

			state, err := readLifeStateInTx(t, db, lifeStateClinicA, own.ID)
			require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
			assert.Empty(t, state, "other subject's life state must not be adopted")
		})
	}
}

func TestLifeStateReader_ReaderFailureIsNotAResolutionError(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	pet := makeLifeStatePet(t, db, lifeStateClinicA, "reader-failure")
	linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, lifeStatePtr(model.PetStatusAlive), nil)

	// 終了済み transaction へのクエリは I/O 失敗になる。not found / unresolved とは区別され、alive にもならない。
	tx := db.Begin()
	require.NoError(t, tx.Error)
	require.NoError(t, tx.Rollback().Error)

	state, err := NewLifeStateReader().ReadPetLifeState(persistence.WithTxValue(context.Background(), tx), lifeStateClinicA, pet.ID)

	require.Error(t, err)
	assert.Empty(t, state)
	assert.False(t, apperrors.IsNotFound(err), "reader failure must not look like not found: %v", err)
	assert.False(t, errors.Is(err, sharedkernel.ErrPetLifeStateUnresolved), "reader failure must stay distinguishable from resolution errors: %v", err)

	// ValidatePetNotDeceased もこのエラーを通さず伝播する。
	guardErr := sharedkernel.ValidatePetNotDeceased(persistence.WithTxValue(context.Background(), tx), NewLifeStateReader(), lifeStateClinicA, pet.ID, "死亡")
	require.Error(t, guardErr)
	assert.False(t, apperrors.IsInvalidInput(guardErr))
}

func TestRequireSameSubjectRevision(t *testing.T) {
	subjectID := uuid.New()
	otherID := uuid.New()
	revisionID := uint64(7)

	t.Run("same subject returns the revision id", func(t *testing.T) {
		got, err := requireSameSubjectRevision(lifeStateRevisionIdentityRow{
			SubjectID: subjectID, CurrentRevisionID: &revisionID, RevisionID: &revisionID, RevisionCommonPetID: &subjectID,
		})
		require.NoError(t, err)
		assert.Equal(t, revisionID, got)
	})

	for name, row := range map[string]lifeStateRevisionIdentityRow{
		"no pointer":          {SubjectID: subjectID},
		"pointer, no row":     {SubjectID: subjectID, CurrentRevisionID: &revisionID},
		"revision of another": {SubjectID: subjectID, CurrentRevisionID: &revisionID, RevisionID: &revisionID, RevisionCommonPetID: &otherID},
	} {
		t.Run(name+" is unresolved", func(t *testing.T) {
			got, err := requireSameSubjectRevision(row)
			require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
			assert.Zero(t, got)
		})
	}
}

// reader の FOR SHARE ロックは、subject の現行 revision を動かす writer（common_basic_revision_guard が
// subject を FOR UPDATE する）と直列化される: reader tx が生きている間は writer のロックは取れず、
// reader tx 終了後に取れる。
func TestLifeStateReader_HoldsSubjectShareLockUntilTransactionEnds(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	pet := makeLifeStatePet(t, db, lifeStateClinicA, "share-lock")
	commonID, _ := linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, lifeStatePtr(model.PetStatusAlive), nil)

	lockSubjectForUpdate := func(tx *gorm.DB) error {
		var row lifeStateSubjectRow
		return tx.Table("common_pet_subjects").
			Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			Select("id").Where("id = ?", commonID).Take(&row).Error
	}

	require.NoError(t, db.Transaction(func(readerTx *gorm.DB) error {
		ctx := persistence.WithTxValue(context.Background(), readerTx)
		state, err := NewLifeStateReader().ReadPetLifeState(ctx, lifeStateClinicA, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, sharedkernel.PetLifeStateAlive, state)

		competing := db.Begin()
		require.NoError(t, competing.Error)
		defer competing.Rollback()
		assert.Error(t, lockSubjectForUpdate(competing),
			"a writer must not obtain FOR UPDATE on the subject while the reader's FOR SHARE is held")
		return nil
	}))

	// reader tx 終了後は writer がロックできる。
	writer := db.Begin()
	require.NoError(t, writer.Error)
	defer writer.Rollback()
	assert.NoError(t, lockSubjectForUpdate(writer))
}

// ロック待ち中に別 tx が現行ポインタを死亡 revision へ進めて commit した場合、reader は待機後に
// ポインタを再解決し、古い alive を返さず deceased を返す（ロック前に読んだ値を使い回さない）。
func TestLifeStateReader_ReResolvesCurrentRevisionAfterLockWait(t *testing.T) {
	db := setupLifeStateDB(t)
	species := makeLifeStateSpecies(t, db)
	pet := makeLifeStatePet(t, db, lifeStateClinicA, "relock")
	commonID, _ := linkLifeStatePetWithRevision(t, db, lifeStateClinicA, pet, species.ID, lifeStatePtr(model.PetStatusAlive), nil)

	// writer: subject を FOR UPDATE で保持したまま、死亡 revision を追加してポインタを進める。
	writer := db.Begin()
	require.NoError(t, writer.Error)
	committed := false
	defer func() {
		if !committed {
			writer.Rollback()
		}
	}()
	var locked lifeStateSubjectRow
	require.NoError(t, writer.Table("common_pet_subjects").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").Where("id = ?", commonID).Take(&locked).Error)

	type readResult struct {
		state sharedkernel.PetLifeState
		err   error
	}
	started := make(chan struct{})
	results := make(chan readResult, 1)
	go func() {
		var res readResult
		res.err = db.Transaction(func(readerTx *gorm.DB) error {
			close(started)
			ctx := persistence.WithTxValue(context.Background(), readerTx)
			res.state, res.err = NewLifeStateReader().ReadPetLifeState(ctx, lifeStateClinicA, pet.ID)
			return res.err
		})
		results <- res
	}()
	<-started

	// reader は writer の FOR UPDATE 解放を待つ（commit 前には返らない）。
	select {
	case res := <-results:
		t.Fatalf("reader returned before the writer committed: state=%q err=%v", res.state, res.err)
	case <-time.After(300 * time.Millisecond):
	}

	deceased := model.PetStatusDeceased
	deceasedAt := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	newRev := &model.CommonPetBasicRevision{
		CommonPetID: commonID, Revision: 2,
		SourceClinicID: lifeStateClinicA, SourceLocalPetID: pet.ID, OperationID: 2,
		Name: "life-state", AnimalSpeciesID: species.ID,
		Status: &deceased, DeceasedAt: &deceasedAt,
		NeuteredStatus: model.PetNeuteredStatusUnknown,
		Provenance:     json.RawMessage(`{}`),
		RecordedBy:     1, RecordedAt: time.Now(),
	}
	require.NoError(t, writer.Create(newRev).Error)
	require.NoError(t, writer.Model(&model.CommonPetSubject{}).
		Where("id = ?", commonID).
		Update("current_basic_revision_id", newRev.ID).Error)
	require.NoError(t, writer.Commit().Error)
	committed = true

	select {
	case res := <-results:
		require.NoError(t, res.err)
		assert.Equal(t, sharedkernel.PetLifeStateDeceased, res.state,
			"the pointer must be re-resolved after the lock wait; a stale alive would let a write through")
	case <-time.After(5 * time.Second):
		t.Fatal("reader did not return after the writer committed")
	}
}
