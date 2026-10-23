package medicalrecord

// cage_merge_test.go — EMR-306 ケージ一時連結のサービス統合テスト（実 Postgres テスト DB）。
// Merge / Unmerge と、連結中ケージの Update(is_active=false) / Delete ガードを対象とする。
// 実 DB 上で医院隔離・ロック直列化・状態遷移の実効性を検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// newCageMergeService は実リポジトリ+実トランザクタでサービスを組み立て、
// CageMerger 面を取り出す（型アサーション経路ごと検証する）。
func newCageMergeService(t *testing.T, db *gorm.DB) (CageService, CageMerger) {
	t.Helper()
	repo := NewCageRepository(db)
	svc := NewCageService(repo, persistence.NewTransactor(db))
	merger, ok := svc.(CageMerger)
	require.True(t, ok, "cage service must implement CageMerger")
	return svc, merger
}

// makeCageHospitalization は status を明示した入院レコードを作成する。
// makeHospitalizationRec は status 未指定（DB default 'reserved'）なので、
// admitted/discharged の検証にはこちらを使う。
// makeCageHospitalization creates a hospitalization with the stay history that yields the requested
// effective state through the EMR-499 core (hotel stay: no doctor/admission record needed):
// admitted = admit event, discharged = admit + discharge, reserved = a stay with NO event (the core
// cannot prove a reservation without acceptance proof, so it reads as unknown and blocks the cage).
func makeCageHospitalization(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, cageID *uint64, status model.HospitalizationStatus) *model.Hospitalization {
	t.Helper()
	ctx := context.Background()
	stay := &model.StayCase{ClinicID: clinicID, OwnerID: ownerID, PetID: petID, StayKind: "hotel"}
	require.NoError(t, db.WithContext(ctx).Create(stay).Error)
	h := &model.Hospitalization{
		ClinicID:   clinicID,
		OwnerID:    ownerID,
		PetID:      petID,
		CageID:     cageID,
		StayCaseID: &stay.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(h).Error)
	admittedAt := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	addEvent := func(revision int64, kind string, at time.Time) {
		require.NoError(t, db.WithContext(ctx).Create(&model.StayMovementEvent{
			ClinicID: clinicID, StayCaseID: stay.ID, HospitalizationID: &h.ID, Revision: revision,
			EventKind: kind, ActualAt: &at, OperationID: 1, RecordedBy: 1,
		}).Error)
	}
	switch status {
	case model.HospitalizationStatusAdmitted:
		addEvent(1, "admit", admittedAt)
	case model.HospitalizationStatusDischarged:
		addEvent(1, "admit", admittedAt)
		addEvent(2, "discharge", admittedAt.Add(24*time.Hour))
	}
	return h
}

func cageMergedInto(t *testing.T, db *gorm.DB, cageID uint64) *uint64 {
	t.Helper()
	var cage model.Cage
	require.NoError(t, db.WithContext(context.Background()).First(&cage, cageID).Error)
	return cage.MergedIntoCageID
}

func deactivateCageRow(t *testing.T, db *gorm.DB, cageID uint64) {
	t.Helper()
	require.NoError(t, db.Model(&model.Cage{}).Where("id = ?", cageID).Update("is_active", false).Error)
}

func TestCageService_Merge_SetsMergedIntoOnMembers(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表ケージ")
	m1 := makeCageMaster(t, db, clinicID, "構成ケージ1")
	m2 := makeCageMaster(t, db, clinicID, "構成ケージ2")

	got, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID, m2.ID})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, rep.ID, got.ID, "Merge は代表ケージを返す")

	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID))
	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m2.ID))
	assert.Nil(t, cageMergedInto(t, db, rep.ID), "代表ケージ自身は NULL のまま")
}

func TestCageService_Merge_RejectsInvalidMemberIDs(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表ケージ")
	m1 := makeCageMaster(t, db, clinicID, "構成ケージ1")

	// 上限超（maxCageMergeMembers+1）はドメイン検査前に InvalidInput で早期拒否される。
	overLimit := make([]uint64, maxCageMergeMembers+1)
	for i := range overLimit {
		overLimit[i] = uint64(i + 1)
	}

	tests := []struct {
		name      string
		memberIDs []uint64
	}{
		{name: "empty member list", memberIDs: nil},
		{name: "duplicate member ids", memberIDs: []uint64{m1.ID, m1.ID}},
		{name: "representative itself in members", memberIDs: []uint64{rep.ID}},
		{name: "member count exceeds cap", memberIDs: overLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := merger.Merge(ctx, clinicID, rep.ID, tt.memberIDs)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput, got %v", err)
			assert.Nil(t, cageMergedInto(t, db, m1.ID), "失敗時に連結状態が残ってはならない")
		})
	}
}

func TestCageService_Merge_RejectsUnknownOrCrossClinic(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rep := makeCageMaster(t, db, clinicA, "代表ケージ")
	m1 := makeCageMaster(t, db, clinicA, "構成ケージ1")
	otherClinicCage := makeCageMaster(t, db, clinicB, "他院ケージ")

	t.Run("unknown member id → NotFound", func(t *testing.T) {
		_, err := merger.Merge(ctx, clinicA, rep.ID, []uint64{999999})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
	})

	t.Run("cross-clinic member id → NotFound", func(t *testing.T) {
		_, err := merger.Merge(ctx, clinicA, rep.ID, []uint64{otherClinicCage.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
	})

	t.Run("cross-clinic representative id → NotFound", func(t *testing.T) {
		_, err := merger.Merge(ctx, clinicA, otherClinicCage.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
	})

	assert.Nil(t, cageMergedInto(t, db, m1.ID))
}

func TestCageService_Merge_RejectsInactiveCage(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("inactive member → Conflict", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表ケージ inactive-member")
		m1 := makeCageMaster(t, db, clinicID, "無効構成ケージ")
		deactivateCageRow(t, db, m1.ID)

		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("inactive representative → Conflict", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "無効代表ケージ")
		m1 := makeCageMaster(t, db, clinicID, "構成ケージ inactive-rep")
		deactivateCageRow(t, db, rep.ID)

		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Nil(t, cageMergedInto(t, db, m1.ID))
	})
}

func TestCageService_Merge_RejectsAlreadyMergedOrNesting(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("member already linked to another group → Conflict", func(t *testing.T) {
		repA := makeCageMaster(t, db, clinicID, "代表A")
		m1 := makeCageMaster(t, db, clinicID, "構成1")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, repA.ID, []uint64{m1.ID}))

		repB := makeCageMaster(t, db, clinicID, "代表B")
		_, err := merger.Merge(ctx, clinicID, repB.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("representative that is itself a member → Conflict", func(t *testing.T) {
		repA := makeCageMaster(t, db, clinicID, "代表A2")
		m1 := makeCageMaster(t, db, clinicID, "構成A2")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, repA.ID, []uint64{m1.ID}))

		// m1 is now a member; using it as representative must fail.
		m2 := makeCageMaster(t, db, clinicID, "構成A2-2")
		_, err := merger.Merge(ctx, clinicID, m1.ID, []uint64{m2.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("representative already heading a group → Conflict", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表heads")
		m1 := makeCageMaster(t, db, clinicID, "構成heads-1")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))

		m2 := makeCageMaster(t, db, clinicID, "構成heads-2")
		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m2.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Nil(t, cageMergedInto(t, db, m2.ID))
	})

	t.Run("member candidate heading its own group → Conflict (nesting)", func(t *testing.T) {
		nestedRep := makeCageMaster(t, db, clinicID, "入れ子代表")
		nestedMember := makeCageMaster(t, db, clinicID, "入れ子構成")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, nestedRep.ID, []uint64{nestedMember.ID}))

		rep := makeCageMaster(t, db, clinicID, "代表nesting")
		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{nestedRep.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})
}

func mustMerge(t *testing.T, merger CageMerger, ctx context.Context, clinicID, representativeID uint64, memberIDs []uint64) error {
	t.Helper()
	_, err := merger.Merge(ctx, clinicID, representativeID, memberIDs)
	return err
}

func TestCageService_Merge_RejectsOccupiedCage(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 occupancy")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ポチ")

	t.Run("admitted hospitalization on representative → Conflict", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "入院中代表")
		m1 := makeCageMaster(t, db, clinicID, "構成-入院中代表")
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusAdmitted)

		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Nil(t, cageMergedInto(t, db, m1.ID))
	})

	t.Run("reserved hospitalization on member → Conflict", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表-予約構成")
		m1 := makeCageMaster(t, db, clinicID, "予約中構成")
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &m1.ID, model.HospitalizationStatusReserved)

		_, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})
}

func TestCageService_Merge_AllowsDischargedOnlyHistory(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 discharged")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "タマ")

	rep := makeCageMaster(t, db, clinicID, "代表-退院履歴")
	m1 := makeCageMaster(t, db, clinicID, "構成-退院履歴")
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusDischarged)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &m1.ID, model.HospitalizationStatusDischarged)

	got, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID})
	require.NoError(t, err, "退院済み履歴のみのケージは連結できる")
	assert.Equal(t, rep.ID, got.ID)
	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID))
}

func TestCageService_Unmerge_ClearsMemberLinks(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-unmerge")
	m1 := makeCageMaster(t, db, clinicID, "構成-unmerge-1")
	m2 := makeCageMaster(t, db, clinicID, "構成-unmerge-2")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID, m2.ID}))

	require.NoError(t, merger.Unmerge(ctx, clinicID, rep.ID))
	assert.Nil(t, cageMergedInto(t, db, m1.ID))
	assert.Nil(t, cageMergedInto(t, db, m2.ID))
	assert.Nil(t, cageMergedInto(t, db, rep.ID))
}

func TestCageService_Unmerge_RejectsNonRepresentative(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-unmerge-ng")
	m1 := makeCageMaster(t, db, clinicID, "構成-unmerge-ng")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))

	t.Run("member cage → Conflict", func(t *testing.T) {
		err := merger.Unmerge(ctx, clinicID, m1.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("standalone cage heading no group → Conflict", func(t *testing.T) {
		solo := makeCageMaster(t, db, clinicID, "独立ケージ")
		err := merger.Unmerge(ctx, clinicID, solo.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("unknown cage → NotFound", func(t *testing.T) {
		err := merger.Unmerge(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
	})

	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID), "失敗時に連結が解除されてはならない")
}

func TestCageService_Unmerge_RejectsOccupiedGroup(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 unmerge-occupied")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "クロ")

	t.Run("admitted on member blocks unmerge", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表-um-occ1")
		m1 := makeCageMaster(t, db, clinicID, "構成-um-occ1")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &m1.ID, model.HospitalizationStatusAdmitted)

		err := merger.Unmerge(ctx, clinicID, rep.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID), "解除失敗時に連結が残るべき")
	})

	t.Run("reserved on representative blocks unmerge", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表-um-occ2")
		m1 := makeCageMaster(t, db, clinicID, "構成-um-occ2")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusReserved)

		err := merger.Unmerge(ctx, clinicID, rep.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})
}

func TestCageService_Unmerge_AllowsDischargedOnlyHistory(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 unmerge-discharged")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ミケ")

	rep := makeCageMaster(t, db, clinicID, "代表-um-dis")
	m1 := makeCageMaster(t, db, clinicID, "構成-um-dis")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &m1.ID, model.HospitalizationStatusDischarged)

	require.NoError(t, merger.Unmerge(ctx, clinicID, rep.ID), "退院済み履歴のみなら解除できる")
	assert.Nil(t, cageMergedInto(t, db, m1.ID))
}

func TestCageService_Merge_RemergeAfterUnmerge(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-remerge")
	m1 := makeCageMaster(t, db, clinicID, "構成-remerge-1")
	m2 := makeCageMaster(t, db, clinicID, "構成-remerge-2")

	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{m1.ID}))
	require.NoError(t, merger.Unmerge(ctx, clinicID, rep.ID))

	got, err := merger.Merge(ctx, clinicID, rep.ID, []uint64{m1.ID, m2.ID})
	require.NoError(t, err, "解除後の再連結は可能")
	assert.Equal(t, rep.ID, got.ID)
	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID))
	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m2.ID))
}

func TestCageService_Update_DeactivateMergedConflict(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	svc, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-deact")
	member := makeCageMaster(t, db, clinicID, "構成-deact")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	isActiveFalse := false
	isActiveTrue := true

	t.Run("member cage: is_active=false → Conflict", func(t *testing.T) {
		_, err := svc.Update(ctx, clinicID, member.ID, &UpdateCageInput{IsActive: &isActiveFalse})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		var row model.Cage
		require.NoError(t, db.First(&row, member.ID).Error)
		assert.True(t, row.IsActive, "無効化がロールバックされていること")
	})

	t.Run("representative cage: is_active=false → Conflict", func(t *testing.T) {
		_, err := svc.Update(ctx, clinicID, rep.ID, &UpdateCageInput{IsActive: &isActiveFalse})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("member cage: other fields still update while merged", func(t *testing.T) {
		name := "連結中でも名称更新は可"
		got, err := svc.Update(ctx, clinicID, member.ID, &UpdateCageInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, name, got.Name)
		assert.Equal(t, rep.ID, *got.MergedIntoCageID, "通常更新は連結状態を変えない")
	})

	t.Run("member cage: is_active=true update is unaffected", func(t *testing.T) {
		got, err := svc.Update(ctx, clinicID, member.ID, &UpdateCageInput{IsActive: &isActiveTrue})
		require.NoError(t, err)
		assert.True(t, got.IsActive)
	})

	t.Run("after unmerge both cages deactivate normally", func(t *testing.T) {
		require.NoError(t, merger.Unmerge(ctx, clinicID, rep.ID))
		for _, id := range []uint64{rep.ID, member.ID} {
			got, err := svc.Update(ctx, clinicID, id, &UpdateCageInput{IsActive: &isActiveFalse})
			require.NoError(t, err)
			assert.False(t, got.IsActive)
		}
	})
}

func TestCageService_Delete_MergedConflict(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	svc, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-del")
	member := makeCageMaster(t, db, clinicID, "構成-del")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	t.Run("member cage delete → Conflict", func(t *testing.T) {
		err := svc.Delete(ctx, clinicID, member.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("representative cage delete → Conflict", func(t *testing.T) {
		err := svc.Delete(ctx, clinicID, rep.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	})

	t.Run("after unmerge both cages delete normally", func(t *testing.T) {
		require.NoError(t, merger.Unmerge(ctx, clinicID, rep.ID))
		require.NoError(t, svc.Delete(ctx, clinicID, member.ID))
		require.NoError(t, svc.Delete(ctx, clinicID, rep.ID))
	})
}

// TestCageService_OccupiedCageIDs は連結ピッカー占有エンドポイントの読み取り面を
// 実 DB で検証する（EMR-306 Codex review: hospitalization:view 非依存の
// master 権限エンドポイント。admitted/reserved のみ・医院隔離・退院/未割当を除外）。
func TestCageService_OccupiedCageIDs(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	svc, _ := newCageMergeService(t, db)
	occupancy, ok := svc.(CageOccupancyReader)
	require.True(t, ok, "cage service must implement CageOccupancyReader")
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 occupancy-list")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ハク")
	otherClinicOwner := makeTestOwner(t, db, 2, "他院飼主")
	otherClinicPet := makeSpeciesAndPet(t, db, 2, otherClinicOwner.ID, "他院ペット")

	admittedCage := makeCageMaster(t, db, clinicID, "占有-入院中")
	reservedCage := makeCageMaster(t, db, clinicID, "占有-予約")
	dischargedCage := makeCageMaster(t, db, clinicID, "退院履歴のみ")
	unassignedOnly := makeCageMaster(t, db, clinicID, "ケージ未割当入院のみ")
	emptyCage := makeCageMaster(t, db, clinicID, "空ケージ")
	otherClinicCage := makeCageMaster(t, db, 2, "他院占有ケージ")

	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &admittedCage.ID, model.HospitalizationStatusAdmitted)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &reservedCage.ID, model.HospitalizationStatusReserved)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &dischargedCage.ID, model.HospitalizationStatusDischarged)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, nil, model.HospitalizationStatusAdmitted)
	makeCageHospitalization(t, db, 2, otherClinicOwner.ID, otherClinicPet.ID, &otherClinicCage.ID, model.HospitalizationStatusAdmitted)

	got, err := occupancy.OccupiedCageIDs(ctx, clinicID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uint64{admittedCage.ID, reservedCage.ID}, got,
		"admitted/reserved の占有ケージのみ（退院・未割当・他医院・空は除外）")

	// 非占有ケージが混ざらないことの明示的な確認
	assert.NotContains(t, got, dischargedCage.ID)
	assert.NotContains(t, got, unassignedOnly.ID)
	assert.NotContains(t, got, emptyCage.ID)
	assert.NotContains(t, got, otherClinicCage.ID)
}
