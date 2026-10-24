package medicalrecord

// hospitalization_cage_swap_test.go — EMR-306 review (Codex P1): ボード入れ替えを
// 単一トランザクションで行う SwapCages の実 DB 統合テスト。
// - 正常系: 2件の cage_id が交換され、レスポンスは入力順で返る（id 昇順ロックの内部
//   正規化は wire の順序を変えない）
// - バリデーション: 同一 id・同一ケージは 400、表示時の割当と不一致は 409
// - ガード: 構成ケージ宛は 422、占有グループ宛は 409 — どちらも「片側だけ残る」
//   半スワップ状態を作らない（両行が元の cage_id のまま）
// - テナント: 他医院の入院は NotFound
// 実 Postgres テスト DB（setupHospitalizationRepoTestDB）で検証する。

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func TestHospitalizationSwapCages_SwapsTwoHospitalizations(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元2")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusAdmitted)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.NoError(t, err)
	require.NotNil(t, gotA)
	require.NotNil(t, gotB)
	// レスポンスは入力順（idA, idB）で返る。
	assert.Equal(t, hospA.ID, gotA.ID)
	assert.Equal(t, hospB.ID, gotB.ID)
	require.NotNil(t, gotA.CageID)
	require.NotNil(t, gotB.CageID)
	assert.Equal(t, cage2.ID, *gotA.CageID, "idA は hospB のケージへ")
	assert.Equal(t, cage1.ID, *gotB.CageID, "idB は hospA のケージへ")

	// DB 上も交換済み。
	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, cage2.ID, *reloadedA.CageID)
	assert.Equal(t, cage1.ID, *reloadedB.CageID)
}

// id 昇順の内部正規化はロック順の話であり、呼出順（idA > idB）でも結果は入力順に返る。
func TestHospitalizationSwapCages_ReversedInputOrder(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 rev")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット rev")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 rev1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元 rev2")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusAdmitted)
	require.Greater(t, hospB.ID, hospA.ID)

	// idA > idB の順で呼んでも入れ替えと応答順は変わらない。
	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospB.ID, hospA.ID, [2]uint64{*hospB.CageID, *hospA.CageID})
	require.NoError(t, err)
	require.NotNil(t, gotA)
	require.NotNil(t, gotB)
	assert.Equal(t, hospB.ID, gotA.ID, "先頭要素は入力 idA")
	assert.Equal(t, hospA.ID, gotB.ID, "2番目要素は入力 idB")
	assert.Equal(t, cage1.ID, *gotA.CageID)
	assert.Equal(t, cage2.ID, *gotB.CageID)
}

func TestHospitalizationSwapCages_InvalidInputRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 invalid")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット invalid")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 inv1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元 inv2")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusAdmitted)
	hospNoCage := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, nil, model.HospitalizationStatusAdmitted)

	t.Run("same hospitalization id", func(t *testing.T) {
		gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospA.ID, [2]uint64{cage1.ID, cage1.ID})
		require.Error(t, err)
		assert.Nil(t, gotA)
		assert.Nil(t, gotB)
		assert.True(t, errors.Is(err, apperrors.ErrInvalidInput), "同一 id は INVALID_INPUT: %v", err)
	})

	t.Run("hospitalization without cage", func(t *testing.T) {
		gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospNoCage.ID, [2]uint64{cage1.ID, cage2.ID})
		require.Error(t, err)
		assert.Nil(t, gotA)
		assert.Nil(t, gotB)
		assert.True(t, apperrors.IsConflict(err), "表示時の割当からNULLへ変わった場合は CONFLICT: %v", err)
	})

	t.Run("both on same cage", func(t *testing.T) {
		hospC := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
		gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospC.ID, [2]uint64{*hospA.CageID, *hospC.CageID})
		require.Error(t, err)
		assert.Nil(t, gotA)
		assert.Nil(t, gotB)
		assert.True(t, errors.Is(err, apperrors.ErrInvalidInput), "同一ケージ内交換は INVALID_INPUT: %v", err)
	})

	// 失敗後も元の割当が残っている（行の再読込は都度新しい変数で — GORM は
	// dest struct の主キーを条件に含めるため、使い回すと id=2 AND id=1 になる）。
	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	assert.Equal(t, cage1.ID, *reloadedA.CageID)
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, cage2.ID, *reloadedB.CageID)
}

// 構成ケージ宛の移動は 422 で、交換は両側ロールバックする（半スワップを残さない）。
func TestHospitalizationSwapCages_MemberTargetRejectedAndRolledBack(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 member")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット member")
	rep := makeCageMaster(t, db, clinicID, "代表 swap-member")
	member := makeCageMaster(t, db, clinicID, "構成 swap-member")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	standalone := makeCageMaster(t, db, clinicID, "独立 swap-member")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusAdmitted)
	// 構成ケージ上の入院（連結前の旧割当を想定 — ガード経路を通らない直接 INSERT）。
	// Codex PR606 #26: discharged は臨床状態再検証で 400 になるため、
	// 422 構成ケージガードに到達する行は occupying status（admitted）で作る。
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.Error(t, err, "構成ケージへの移動を含む交換は拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	assert.True(t, errors.Is(err, apperrors.ErrUnprocessable), "構成ケージ宛は 422: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusUnprocessableEntity, status)

	// 片側だけ commit されない（両行が元のまま）。
	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	assert.Equal(t, standalone.ID, *reloadedA.CageID, "拒否された交換で hospA の cage_id が変わってはならない")
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, member.ID, *reloadedB.CageID, "拒否された交換で hospB の cage_id が変わってはならない")
}

// 占有グループの代表宛の移動は 409 で、交換は両側ロールバックする。
func TestHospitalizationSwapCages_OccupiedGroupRejectedAndRolledBack(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	_, merger := newCageMergeService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 occ")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット occ")
	rep := makeCageMaster(t, db, clinicID, "代表 swap-occ")
	member := makeCageMaster(t, db, clinicID, "構成 swap-occ")
	require.NoError(t, mustMerge(t, merger, ctx, clinicID, rep.ID, []uint64{member.ID}))

	// 構成ケージに占有中の旧入院（ガード経路を通らない直接 INSERT で占有状態を作る）。
	makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &member.ID, model.HospitalizationStatusAdmitted)
	standalone := makeCageMaster(t, db, clinicID, "独立 swap-occ")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &standalone.ID, model.HospitalizationStatusAdmitted)
	// 代表上の入院（admitted — #26 で discharged は先に 400 拒否されるため、
	// 409 占有グループガードの検証には occupying status を使う）。
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &rep.ID, model.HospitalizationStatusAdmitted)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.Error(t, err, "占有グループ代表への移動を含む交換は拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	assert.True(t, apperrors.IsConflict(err), "占有グループ宛は 409: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusConflict, status)

	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	assert.Equal(t, standalone.ID, *reloadedA.CageID)
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, rep.ID, *reloadedB.CageID)
}

// 交換相手が他医院の入院なら NotFound（LockByIDForUpdate は医院スコープ）。
func TestHospitalizationSwapCages_CrossClinicRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)
	const otherClinicID = uint64(2)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 xtenant")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット xtenant")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 xt1")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)

	otherOwner := makeTestOwner(t, db, otherClinicID, "swap飼主 other")
	otherPet := makeSpeciesAndPet(t, db, otherClinicID, otherOwner.ID, "swapペット other")
	otherCage := makeCageMaster(t, db, otherClinicID, "スワップ元 other")
	hospOther := makeCageHospitalization(t, db, otherClinicID, otherOwner.ID, otherPet.ID, &otherCage.ID, model.HospitalizationStatusAdmitted)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospOther.ID, [2]uint64{*hospA.CageID, *hospOther.CageID})
	require.Error(t, err, "他医院の入院を含む交換は拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusNotFound, status, "他医院 id は NotFound: %v", err)

	// 自医院側も変わっていない。
	var reloaded model.Hospitalization
	require.NoError(t, db.First(&reloaded, hospA.ID).Error)
	assert.Equal(t, cage1.ID, *reloaded.CageID)
}

// Codex PR606 #26 (EMR-306 spec 確認済み): discharged を含む交換は拒否する。
// 退院済み行の cage_id は退院時点の履歴であり、この経路では変更しない。
// ロック下再検証により、表示→PATCH 間で退院したケースも両側ロールバックで止める。
func TestHospitalizationSwapCages_DischargedSideRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationService(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 nonocc")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット nonocc")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 no1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元 no2")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusDischarged)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.Error(t, err, "退院済み入院を含む交換は拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput), "discharged は INVALID_INPUT: %v", err)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusBadRequest, status)

	// 拒否後も両行の cage_id は変わらない（半スワップを残さない）。
	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	assert.Equal(t, cage1.ID, *reloadedA.CageID)
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, cage2.ID, *reloadedB.CageID)
}

// Codex PR606 #26 / EMR-491 D6: 死亡ペットを含む交換は同一 tx の死亡 reader 判定で拒否する。
// 死亡は共通 subject の現行 revision が持つため、reader（test double）が deceased を返す状況を再現する。
func TestHospitalizationSwapCages_DeceasedPetRejected(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	svc := newCageGuardHospitalizationServiceWithLife(t, db, stubPetLifeState(sharedkernel.PetLifeStateDeceased, nil))
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 dead")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット dead")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 dead1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元 dead2")
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusAdmitted)

	gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.Error(t, err, "死亡ペットを含む交換は拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	assert.True(t, errors.Is(err, apperrors.ErrInvalidInput), "deceased pet は INVALID_INPUT: %v", err)

	var reloadedA, reloadedB model.Hospitalization
	require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
	assert.Equal(t, cage1.ID, *reloadedA.CageID)
	require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
	assert.Equal(t, cage2.ID, *reloadedB.CageID)
}

// EMR-491 D6: 生死を確定できない（unresolved）/ reader 失敗の pet は生存扱いにせず、
// 業務拒否（INVALID_INPUT）ではなくエラーとして交換を拒否し、半スワップを残さない。
func TestHospitalizationSwapCages_UnresolvedOrFailedLifeStateRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "unresolved", err: apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")},
		{name: "reader failure", err: errors.New("db down")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupHospitalizationRepoTestDB(t)
			svc := newCageGuardHospitalizationServiceWithLife(t, db, stubPetLifeState("", tc.err))
			ctx := context.Background()
			const clinicID = uint64(1)

			owner := makeTestOwner(t, db, clinicID, "swap飼主 unresolved")
			pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット unresolved")
			cage1 := makeCageMaster(t, db, clinicID, "スワップ元 unres1")
			cage2 := makeCageMaster(t, db, clinicID, "スワップ元 unres2")
			hospA := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
			hospB := makeCageHospitalization(t, db, clinicID, owner.ID, pet.ID, &cage2.ID, model.HospitalizationStatusAdmitted)

			gotA, gotB, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
			require.Error(t, err, "生存を確認できないペットを含む交換は拒否")
			assert.Nil(t, gotA)
			assert.Nil(t, gotB)
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, errors.Is(err, apperrors.ErrInvalidInput), "unknown は deceased の業務拒否ではない: %v", err)

			var reloadedA, reloadedB model.Hospitalization
			require.NoError(t, db.First(&reloadedA, hospA.ID).Error)
			assert.Equal(t, cage1.ID, *reloadedA.CageID)
			require.NoError(t, db.First(&reloadedB, hospB.ID).Error)
			assert.Equal(t, cage2.ID, *reloadedB.CageID)
		})
	}
}

// 死亡 reader 未配線の SwapCages は生存扱いにせず 500 fail-closed で拒否する。
func TestHospitalizationSwapCages_WithoutPetLifeStateReaderFailsClosed(t *testing.T) {
	svc := NewHospitalizationService(
		nil, nil, swapSnapshotPetRepository{}, nil,
		nil, nil, nil,
		&mockTransactor{},
	)

	gotA, gotB, err := svc.SwapCages(context.Background(), 1, 1, 2, [2]uint64{1, 2})
	require.Error(t, err)
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusInternalServerError, status, "fail-closed は 500: %v", err)
}

// Codex PR606 #26 再指摘: 注入された petRepo が行ロック面を持たない場合は
// 通常 read へ退避せず fail-closed（500）で拒否する。
func TestHospitalizationSwapCages_PetRepoWithoutRowLockFailsClosed(t *testing.T) {
	// mockPetRepository は petFinder（FindByID）のみ — petRowLockFinder 面を持たない。
	svc := NewHospitalizationService(
		nil, nil, &mockPetRepository{}, nil,
		nil, nil, nil,
		&mockTransactor{},
		WithPetLifeStateReader(stubPetLifeReader{}),
	)

	gotA, gotB, err := svc.SwapCages(context.Background(), 1, 1, 2, [2]uint64{1, 2})
	require.Error(t, err, "行ロック面を持たない petRepo では拒否")
	assert.Nil(t, gotA)
	assert.Nil(t, gotB)
	status, _, _ := httpapi.ResolveErrorResponse(err)
	assert.Equal(t, http.StatusInternalServerError, status, "fail-closed は 500: %v", err)
}

// swapEventLog は SwapCages 内の呼び出し順序を記録する。
type swapEventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *swapEventLog) add(event string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
}

func (l *swapEventLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.events...)
}

// loggingPetLocker は実 pet リポジトリのローカル行ロックを包み、ロック取得を記録する。
type loggingPetLocker struct {
	petdomain.CompleteRepository
	log *swapEventLog
}

func (l *loggingPetLocker) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	p, err := l.CompleteRepository.LockByIDForUpdate(ctx, clinicID, id)
	if err == nil {
		l.log.add(fmt.Sprintf("lock:%d", id))
	}
	return p, err
}

// loggingLifeReader は死亡 reader 呼び出しを、ambient tx の有無つきで記録する（常に alive）。
type loggingLifeReader struct {
	log *swapEventLog
}

func (r loggingLifeReader) ReadPetLifeState(ctx context.Context, _ uint64, petID uint64) (sharedkernel.PetLifeState, error) {
	r.log.add(fmt.Sprintf("life:%d tx=%t", petID, persistence.TxFromContext(ctx) != nil))
	return sharedkernel.PetLifeStateAlive, nil
}

// EMR-491 D6: スワップは対象ペットを pet_id 昇順に、ローカル行ロックの直後に、同じ ambient tx の中で
// 死亡 reader 判定する（ローカル行ロックだけでは死亡判定は成立しない）。
func TestHospitalizationSwapCages_LifeStateReadAfterLocalLockInsideTx(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	log := &swapEventLog{}
	svc := NewHospitalizationService(
		NewHospitalizationRepository(db),
		reservation.NewReservationRepository(db),
		&loggingPetLocker{CompleteRepository: petdomain.NewRepository(db), log: log},
		NewCageRepository(db),
		nil, nil, nil,
		persistence.NewTransactor(db),
		WithPetLifeStateReader(loggingLifeReader{log: log}),
	)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "swap飼主 order")
	petLow := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット order low")
	petHigh := makeSpeciesAndPet(t, db, clinicID, owner.ID, "swapペット order high")
	cage1 := makeCageMaster(t, db, clinicID, "スワップ元 ord1")
	cage2 := makeCageMaster(t, db, clinicID, "スワップ元 ord2")
	// 入院 A は pet が大きい方、B は小さい方: 入力順に関わらず pet_id 昇順で判定される。
	hospA := makeCageHospitalization(t, db, clinicID, owner.ID, petHigh.ID, &cage1.ID, model.HospitalizationStatusAdmitted)
	hospB := makeCageHospitalization(t, db, clinicID, owner.ID, petLow.ID, &cage2.ID, model.HospitalizationStatusAdmitted)

	_, _, err := svc.SwapCages(ctx, clinicID, hospA.ID, hospB.ID, [2]uint64{*hospA.CageID, *hospB.CageID})
	require.NoError(t, err)

	require.Less(t, petLow.ID, petHigh.ID)
	assert.Equal(t, []string{
		fmt.Sprintf("lock:%d", petLow.ID),
		fmt.Sprintf("life:%d tx=true", petLow.ID),
		fmt.Sprintf("lock:%d", petHigh.ID),
		fmt.Sprintf("life:%d tx=true", petHigh.ID),
	}, log.snapshot(), "ローカル行ロック直後に同一 tx で死亡 reader を pet_id 昇順に呼ぶ")
}
