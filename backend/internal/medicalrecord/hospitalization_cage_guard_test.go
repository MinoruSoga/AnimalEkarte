package medicalrecord

// hospitalization_cage_guard_test.go — EMR-306 step3: 入院のケージ割当ガードの実 DB 統合テスト。
// - 構成ケージ（merged_into_cage_id IS NOT NULL）への割当 → ErrUnprocessable（API層 422）
// - 連結グループ代表への割当でグループ内に未削除の admitted/reserved 入院があれば → 409
// - 退院済み履歴のみのグループ・未連結ケージは従来挙動（追加チェックなし）
// - cageFinder のみを実装する repo（CageMergeRepository 面なし）の注入は 500 fail-closed
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

func TestHospitalizationCageGuard_Update_ClearCageOnReactivation(t *testing.T) {
	tests := []struct {
		name      string
		oldMember bool
		status    string
	}{
		{"former member", true, "admitted"},
		{"occupied representative", false, "reserved"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupHospitalizationRepoTestDB(t)
			svc := newCageGuardHospitalizationService(t, db)
			_, merger := newCageMergeService(t, db)
			ctx := context.Background()
			const clinicID = uint64(1)
			owner := makeTestOwner(t, db, clinicID, "clear再開飼主")
			pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "clear再開ペット")
			oldCage := makeCageMaster(t, db, clinicID, "退院時のケージ")
			otherCage := makeCageMaster(t, db, clinicID, "連結相手のケージ")
			hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &oldCage.ID, model.HospitalizationStatusDischarged)
			if tt.oldMember {
				require.NoError(t, mustMerge(t, merger, ctx, clinicID, otherCage.ID, []uint64{oldCage.ID}))
			} else {
				require.NoError(t, mustMerge(t, merger, ctx, clinicID, oldCage.ID, []uint64{otherCage.ID}))
				occupantPet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "現在の占有ペット")
				_, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, occupantPet.ID, oldCage.ID, model.HospitalizationStatusAdmitted))
				require.NoError(t, err)
			}
			input, err := (&updateHospitalizationRequest{Status: &tt.status, ClearCageID: true}).toServiceInput()
			require.NoError(t, err)

			got, err := svc.Update(ctx, clinicID, hosp.ID, &input)

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Nil(t, got.CageID)
			assert.Equal(t, model.HospitalizationStatus(tt.status), got.Status)
			var persisted model.Hospitalization
			require.NoError(t, db.First(&persisted, hosp.ID).Error)
			assert.Nil(t, persisted.CageID)
			assert.Equal(t, got.Status, persisted.Status)
		})
	}
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

// cageFinderOnlyRepo は cageFinder のみを実装する最小スタブ（Grok review LOW-4:
// CageMergeRepository 面を持たない注入が 500 fail-closed になることの検証用。
// mockCageRepository は merge 面を持つため、faceless 経路は別型で表現する）。
type cageFinderOnlyRepo struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
}

func (m *cageFinderOnlyRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

// TestHospitalizationCageGuard_FacelessRepoFailsClosed は CageMergeRepository 面を持たない
// cageFinder 注入の fail-closed 契約を固定する: ガードは no-op ではなく内部エラーを返す
// （fail-open だと構成ケージへの割当が素通りするため — Merge/Unmerge と同じ扱い）。
func TestHospitalizationCageGuard_FacelessRepoFailsClosed(t *testing.T) {
	const clinicID = uint64(1)

	newSvc := func(hosp *mockHospitalizationRepository) HospitalizationService {
		facelessCages := &cageFinderOnlyRepo{findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
			return &model.Cage{ID: id}, nil
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
			facelessCages,
			nil, nil, nil,
			&mockTransactor{},
		)
	}

	t.Run("create with cage fails closed without the merge face", func(t *testing.T) {
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
		require.Error(t, err, "merge 面なし注入は fail-closed（500）")
		assert.Nil(t, got)
		assert.False(t, created, "ガード不通過で永続化してはならない")
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("update cage change fails closed without the merge face", func(t *testing.T) {
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
		require.Error(t, err, "merge 面なし注入は fail-closed（500）")
		assert.Nil(t, got)
		assert.False(t, updated, "ガード不通過で更新してはならない")
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusInternalServerError, status)
	})
}

// EMR-306 review follow-up: 自己占有除外（spec c1）と非占有スキップ（spec c2）の契約。
// 構成ケージ上の既存入院を代表ケージへ移すグループ内移動は、自分自身の占有で
// 409 になってはならない（そうでないと連結前のレガシー割当を是正する経路が塞がれる）。
func TestHospitalizationCageGuard_Update_MemberToRepresentativeMoveAllowed(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 intra-move")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット intra-move")

	rep := makeCageMaster(t, db, clinicID, "代表 intra-move")
	member := makeCageMaster(t, db, clinicID, "構成 intra-move")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// 連結以前に構成ケージへ直接紐付いた占有中入院（makeCageHospitalization は
	// ガード経路を通らない直接 INSERT のため構築可能）。
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID})
	require.NoError(t, err, "自身の占有だけのグループ内移動（構成→代表）は許可する")
	require.NotNil(t, got)
	require.NotNil(t, got.CageID)
	assert.Equal(t, rep.ID, *got.CageID)
}

// 別の占有入院がいる場合は、自身の占有を除外しても依然 409 で拒否する。
func TestHospitalizationCageGuard_Update_IntraGroupMoveBlockedByOtherOccupant(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 intra-occ")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット intra-occ")

	rep := makeCageMaster(t, db, clinicID, "代表 intra-occ")
	member := makeCageMaster(t, db, clinicID, "構成 intra-occ")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// 移動対象は構成ケージ上の入院。代表には別の占有入院がいる → 移動は 409。
	hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusReserved)

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err), "他者の占有があるグループ内移動は Conflict: %v", err)
}

// Grok review HIGH: status 省略の通常作成は Create がデフォルト（過去/当日 →
// admitted、未来日 → reserved）に解決する。ガードは解決後の hospitalization.Status
// で占有判定するため、省略時も占有中グループへの割当は 409 で拒否される。
func TestHospitalizationCageGuard_Create_OmittedStatusResolvesToOccupying(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 omitted-status")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット omitted-status")

	newOccupiedGroup := func(t *testing.T, name string) *model.Cage {
		t.Helper()
		rep := makeCageMaster(t, db, clinicID, "代表 "+name)
		member := makeCageMaster(t, db, clinicID, "構成 "+name)
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)
		return rep
	}
	omittedInput := func(cageID uint64, startDate time.Time) *CreateHospitalizationInput {
		return &CreateHospitalizationInput{
			OwnerID:             owner.ID,
			PetID:               pet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           startDate,
			EndDate:             startDate.AddDate(0, 0, 3),
			Status:              "", // 画面の作成 payload と同じく省略
			CageID:              &cageID,
		}
	}

	t.Run("omitted status resolves to admitted and conflicts on occupied group", func(t *testing.T) {
		rep := newOccupiedGroup(t, "omit-admitted")
		today := time.Now().UTC().Truncate(24 * time.Hour)
		got, err := svc.Create(ctx, clinicID, omittedInput(rep.ID, today))
		require.Error(t, err, "省略→admitted 解決の作成も占有検査を受ける")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err), "Conflict 期待: %v", err)
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusConflict, status)
	})

	t.Run("omitted status resolves to reserved and conflicts on occupied group", func(t *testing.T) {
		rep := newOccupiedGroup(t, "omit-reserved")
		future := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 7)
		got, err := svc.Create(ctx, clinicID, omittedInput(rep.ID, future))
		require.Error(t, err, "省略→reserved 解決の作成も占有検査を受ける")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err), "Conflict 期待: %v", err)
	})

	t.Run("omitted status still rejects member cage with 422", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表 omit-member")
		member := makeCageMaster(t, db, clinicID, "構成 omit-member")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		today := time.Now().UTC().Truncate(24 * time.Hour)
		got, err := svc.Create(ctx, clinicID, omittedInput(member.ID, today))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, apperrors.ErrUnprocessable), "構成ケージへの構造的 422 は維持: %v", err)
	})

	t.Run("omitted status allowed when group unoccupied", func(t *testing.T) {
		rep := makeCageMaster(t, db, clinicID, "代表 omit-empty")
		member := makeCageMaster(t, db, clinicID, "構成 omit-empty")
		require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
		today := time.Now().UTC().Truncate(24 * time.Hour)
		got, err := svc.Create(ctx, clinicID, omittedInput(rep.ID, today))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.HospitalizationStatusAdmitted, got.Status, "デフォルト解決が保存される")
	})
}

// discharged 結果の更新・作成は非占有であり、グループ占有検査をスキップする
// （構成ケージへの構造的 422 は維持 — 履歴投入を阻害しない範囲で占有のみ緩和）。
func TestHospitalizationCageGuard_NonOccupyingResultSkipsOccupancy(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "guard飼主 backfill")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "guardペット backfill")

	rep := makeCageMaster(t, db, clinicID, "代表 backfill")
	member := makeCageMaster(t, db, clinicID, "構成 backfill")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))
	// グループは占有中（構成ケージに admitted）。
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

	t.Run("create discharged onto occupied representative is allowed", func(t *testing.T) {
		got, err := svc.Create(ctx, clinicID, cageGuardCreateInput(owner.ID, pet.ID, rep.ID, model.HospitalizationStatusDischarged))
		require.NoError(t, err, "非占有結果の作成は占有検査をスキップする")
		require.NotNil(t, got)
	})

	t.Run("update cage change with discharged result is allowed", func(t *testing.T) {
		standalone := makeCageMaster(t, db, clinicID, "独立 backfill")
		hosp := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusDischarged)

		got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID})
		require.NoError(t, err, "非占有結果のケージ変更は占有検査をスキップする")
		require.NotNil(t, got)
		require.NotNil(t, got.CageID)
		assert.Equal(t, rep.ID, *got.CageID)
	})
}
