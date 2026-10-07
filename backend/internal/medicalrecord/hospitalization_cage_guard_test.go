package medicalrecord

// hospitalization_cage_guard_test.go — EMR-306 step3: 入院のケージ割当ガードの実 DB 統合テスト。
// - 構成ケージ（merged_into_cage_id IS NOT NULL）への割当 → ErrUnprocessable（API層 422）
// - 連結グループ代表への割当でグループ内に未削除の admitted/reserved 入院があれば → 409
// - 退院済み履歴のみのグループ・未連結ケージは従来挙動（追加チェックなし）
// - cageFinder のみを実装する repo（CageMergeRepository 面なし）ではガード no-op
// 実 Postgres テスト DB（setupHospitalizationRepoTestDB）で検証する。

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// newCageGuardHospitalizationService は実リポジトリ+実トランザクタで入院サービスを組み立てる。
// cageRepo に実 NewCageRepository を渡すため CageMergeRepository 面が型アサーションで有効になる。
func newCageGuardHospitalizationService(t *testing.T, db *gorm.DB) HospitalizationService {
	t.Helper()
	return NewHospitalizationService(
		NewHospitalizationRepository(db),
		reservation.NewReservationRepository(db),
		petdomain.NewRepository(db),
		NewCageRepository(db),
		nil, nil, nil,
		persistence.NewTransactor(db),
	)
}

func cageGuardCreateInput(ownerID, petID, cageID uint64, status model.HospitalizationStatus) *CreateHospitalizationInput {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	return &CreateHospitalizationInput{
		OwnerID:             ownerID,
		PetID:               petID,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now.AddDate(0, 0, 3),
		Status:              status,
		CageID:              &cageID,
	}
}

func TestHospitalizationCageGuard_Create_MergedMemberRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 member-create")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット member-create")
	rep := makeCageMaster(t, db, clinicID, "代表 member-create")
	member := makeCageMaster(t, db, clinicID, "構成 member-create")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, member.ID, model.HospitalizationStatusAdmitted))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, apperrors.ErrUnprocessable), "構成ケージ割当は ErrUnprocessable: %v", err)

	// A1: ガードエラーが API 層の分類ロジックで 422 に解決されること。
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusUnprocessableEntity, status)

	var n int64
	require.NoError(t, db.Model(&model.Hospitalization{}).Where("cage_id = ?", member.ID).Count(&n).Error)
	assert.Equal(t, int64(0), n, "構成ケージへの入院は永続化されてはならない")
}

func TestHospitalizationCageGuard_Create_RepresentativeGroupOccupiedConflict(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 rep-conflict")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット rep-conflict")

	t.Run("admitted hospitalization on member blocks representative assignment", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表 occ-member")
		member := makeCageMaster(t, db, clinicID, "構成 occ-member")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		// 占有はガード経路ではなく直接 INSERT で作る（makeCageHospitalization は status 明示）。
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

		got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, rep.ID, model.HospitalizationStatusAdmitted))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err), "占有構成を持つ代表への割当は Conflict: %v", err)
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusConflict, status)
	})

	t.Run("admitted hospitalization on representative itself blocks", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表 occ-rep")
		member := makeCageMaster(t, db, clinicID, "構成 occ-rep")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusAdmitted)

		got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, rep.ID, model.HospitalizationStatusReserved))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err), "占有中の代表への割当は Conflict: %v", err)
	})

	t.Run("reserved hospitalization on member blocks representative assignment", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表 occ-member-reserved")
		member := makeCageMaster(t, db, clinicID, "構成 occ-member-reserved")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusReserved)

		_, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, rep.ID, model.HospitalizationStatusAdmitted))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "予約中構成を持つ代表への割当は Conflict: %v", err)
	})
}

func TestHospitalizationCageGuard_Create_DischargedOnlyHistoryAllowed(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 discharged")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット discharged")

	rep := makeCageMaster(t, db, clinicID, "代表 discharged-only")
	member := makeCageMaster(t, db, clinicID, "構成 discharged-only")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusDischarged)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusDischarged)

	got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, rep.ID, model.HospitalizationStatusAdmitted))
	require.NoError(t, err, "退院済み履歴のみのグループ代表は割当できる")
	require.NotNil(t, got)
	require.NotNil(t, got.CageID)
	assert.Equal(t, rep.ID, *got.CageID)
}

func TestHospitalizationCageGuard_Create_StandaloneCageUnchanged(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 standalone")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット standalone")

	// 未連結ケージは占有チェックの対象外（従来どおり複数入院を許容）。
	standalone := makeCageMaster(t, db, clinicID, "独立ケージ occupied")
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusAdmitted)

	got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, standalone.ID, model.HospitalizationStatusAdmitted))
	require.NoError(t, err, "未連結ケージの挙動は変えない（占有中でも割当可能）")
	require.NotNil(t, got)

	empty := makeCageMaster(t, db, clinicID, "独立ケージ empty")
	got2, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, empty.ID, model.HospitalizationStatusReserved))
	require.NoError(t, err)
	require.NotNil(t, got2)
}

func TestHospitalizationCageGuard_Update_MergedMemberRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 member-update")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット member-update")

	standalone := makeCageMaster(t, db, clinicID, "独立ケージ member-update")
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusAdmitted)

	rep := makeCageMaster(t, db, clinicID, "代表 member-update")
	member := makeCageMaster(t, db, clinicID, "構成 member-update")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &member.ID})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, apperrors.ErrUnprocessable), "構成ケージへの変更は ErrUnprocessable: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusUnprocessableEntity, status)

	var reloaded model.Hospitalization
	require.NoError(t, db.First(&reloaded, hosp.ID).Error)
	require.NotNil(t, reloaded.CageID)
	assert.Equal(t, standalone.ID, *reloaded.CageID, "失敗した更新で cage_id が変わってはならない")
}

func TestHospitalizationCageGuard_Update_RepresentativeGroupOccupiedConflict(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 rep-update")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット rep-update")

	rep := makeCageMaster(t, db, clinicID, "代表 rep-update")
	member := makeCageMaster(t, db, clinicID, "構成 rep-update")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

	standalone := makeCageMaster(t, db, clinicID, "独立ケージ rep-update")
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusAdmitted)

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err), "占有グループ代表への変更は Conflict: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusConflict, status)
}

func TestHospitalizationCageGuard_Update_EchoOnOccupiedRepresentativeAllowed(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 echo")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット echo")

	rep := makeCageMaster(t, db, clinicID, "代表 echo")
	member := makeCageMaster(t, db, clinicID, "構成 echo")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// 占有中の代表に乗っている入院そのものの更新は自己競合してはならない（D2）。
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusAdmitted)

	t.Run("explicit same-cage echo update is allowed", func(t *testing.T) {
		got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID})
		require.NoError(t, err, "ケージ未変更のエコー更新はガードを発火させない")
		require.NotNil(t, got)
	})

	t.Run("memo-only update on occupying hospitalization is allowed", func(t *testing.T) {
		memo := "経過メモ"
		got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{Memo: &memo})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, memo, got.Memo)
	})
}

func TestHospitalizationCageGuard_Update_ReactivationOnMemberRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 react-member")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット react-member")

	rep := makeCageMaster(t, db, clinicID, "代表 react-member")
	member := makeCageMaster(t, db, clinicID, "構成 react-member")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// 構成ケージに紐付く退院済み入院（例: 旧データ/経路外の残骸）の再有効化は拒否する。
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusDischarged)

	admitted := model.HospitalizationStatusAdmitted
	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{Status: &admitted})
	require.Error(t, err, "非占有→占有遷移はケージ未指定でもガードを発火させる")
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, apperrors.ErrUnprocessable), "構成ケージ上の再有効化は ErrUnprocessable: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusUnprocessableEntity, status)
}

func TestHospitalizationCageGuard_Update_ReactivationAllowedWhenGroupUnoccupied(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 react-ok")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット react-ok")

	rep := makeCageMaster(t, db, clinicID, "代表 react-ok")
	member := makeCageMaster(t, db, clinicID, "構成 react-ok")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// 代表上の退院済み入院を再有効化: 自身の discharged 行は CountOccupiedCages に
	// 数えられないため自己競合しない（D2）。
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusDischarged)

	admitted := model.HospitalizationStatusAdmitted
	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{Status: &admitted})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, model.HospitalizationStatusAdmitted, got.Status)
}

func TestHospitalizationCageGuard_Update_ReactivationGroupOccupiedConflict(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 react-occ")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット react-occ")

	rep := makeCageMaster(t, db, clinicID, "代表 react-occ")
	member := makeCageMaster(t, db, clinicID, "構成 react-occ")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusDischarged)
	// 別ケージ（構成）に占有中の入院があると再有効化は 409。
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

	admitted := model.HospitalizationStatusAdmitted
	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{Status: &admitted})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err), "グループ占有中の再有効化は Conflict: %v", err)
}

// TestHospitalizationCageGuard_LegacyCageFinderSkipsGuard は CageMergeRepository 面を持たない
// cageFinder 注入サービスのレガシー経路回帰: FindByID が構成ケージ相当の行を返しても
// ガードは no-op で、従来の validateOwnedMasterFK のみが走る。
func TestHospitalizationCageGuard_LegacyCageFinderSkipsGuard(t *testing.T) {
	const clinicID = uint64(1)

	newSvc := func(hosp *mockHospitalizationRepository) HospitalizationService {
		mergedInto := uint64(77) // FindByID が返すケージは「構成ケージ」に見える
		legacyCages := &mockCageRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
			return &model.Cage{ID: id, MergedIntoCageID: &mergedInto}, nil
		}}
		return NewHospitalizationService(
			hosp,
			&mockReservationRepository{
				assertOwnerInClinicFn:  func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil },
			},
			&mockPetRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, Status: model.PetStatusAlive}, nil
			}},
			legacyCages,
			nil, nil, nil,
			&mockTransactor{},
		)
	}

	t.Run("create on member-looking cage succeeds without the merge face", func(t *testing.T) {
		created := false
		svc := newSvc(&mockHospitalizationRepository{
			createFn: func(_ context.Context, _ *model.Hospitalization) error { created = true; return nil },
		})
		cageID := uint64(10)
		now := time.Now().UTC().Truncate(24 * time.Hour)
		got, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID:             2,
			PetID:               5,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, 3),
			Status:              model.HospitalizationStatusAdmitted,
			CageID:              &cageID,
		})
		require.NoError(t, err, "merge 面なしでは構成ケージ相当でも従来どおり作成できる")
		assert.NotNil(t, got)
		assert.True(t, created)
	})

	t.Run("update cage change succeeds without the merge face", func(t *testing.T) {
		updated := false
		svc := newSvc(&mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: id, ClinicID: clinicID, Status: model.HospitalizationStatusAdmitted}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				updated = true
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		})
		cageID := uint64(99)
		got, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &cageID})
		require.NoError(t, err, "merge 面なしではケージ変更ガードは発火しない")
		assert.NotNil(t, got)
		assert.True(t, updated)
	})
}
