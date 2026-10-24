package medicalrecord

// hospitalization_cage_guard_test.go — EMR-306 step3: 入院のケージ割当ガードの実 DB 統合テスト。
// EMR-499 U1: the Create variants of these guards and the PATCH status "reactivation" variants are gone:
// Create refuses every non-NULL cage as occupancy_unknown (covered in hospitalization_cage_write_guard_test.go)
// and status is no longer writable. The remaining cases drive occupancy through real stay movement history.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// newCageGuardHospitalizationService は実リポジトリ+実トランザクタで入院サービスを組み立てる。
// cageRepo に実 NewCageRepository を渡すため CageMergeRepository 面が型アサーションで有効になる。
func newCageGuardHospitalizationService(t *testing.T, db *gorm.DB) HospitalizationService {
	t.Helper()
	return newCageGuardHospitalizationServiceWithLife(t, db, stubPetLifeReader{})
}

// newCageGuardHospitalizationServiceWithLife は EMR-491 D6 の死亡 reader を差し替えた入院サービスを組み立てる
// （ケージ/ローカル pet 行ロックは実 DB、生死判定だけ test double）。
func newCageGuardHospitalizationServiceWithLife(t *testing.T, db *gorm.DB, life sharedkernel.PetLifeStateReader) HospitalizationService {
	t.Helper()
	return NewHospitalizationService(
		NewHospitalizationRepository(db),
		reservation.NewReservationRepository(db),
		petdomain.NewRepository(db),
		NewCageRepositoryWithStayOccupancyReader(db, NewStayCageEffectiveStateReader(db)),
		nil, nil, nil,
		persistence.NewTransactor(db),
		WithPetLifeStateReader(life),
	)
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

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &member.ID, ExpectedCageID: &standalone.ID})
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

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID, ExpectedCageID: &standalone.ID})
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
		got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID, ExpectedCageID: &rep.ID})
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
// Create は merge 面に到達する前に occupancy_unknown（409）で拒否される（hospitalization_service_create.go）ため、ここでは Update のみを対象にする。
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
				return &model.Pet{ID: id}, nil
			}},
			facelessCages,
			nil, nil, nil,
			&mockTransactor{},
			WithPetLifeStateReader(stubPetLifeReader{}),
		)
	}

	t.Run("update cage change fails closed without the merge face", func(t *testing.T) {
		updated := false
		svc := newSvc(&mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				updated = true
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		})
		cageID := uint64(99)
		got, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &cageID, ExpectedCageIsNull: true})
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

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID, ExpectedCageID: &member.ID})
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

	got, err := svc.Update(ctx, clinicID, hosp.ID, &UpdateHospitalizationInput{CageID: &rep.ID, ExpectedCageID: &member.ID})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err), "他者の占有があるグループ内移動は Conflict: %v", err)
}
