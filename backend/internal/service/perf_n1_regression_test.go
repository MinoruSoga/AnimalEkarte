package service

// perf_n1_regression_test.go — N+1 クエリ回帰テスト (PERF-1 / PERF-2 / PERF-3)
//
// RED フェーズ: 現行実装に対してアサーションが失敗することを確認してから GREEN 化する。
// 各テストは「呼び出し回数」だけを spy し、ビジネスロジックには触れない。

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ─────────────────────────────────────────────
// H-1 回帰テスト: settingsSvc == nil でも DetectDormantOwners がパニックしないこと
// (PERF-2 fix で GetDormantThresholds を直接呼ぶようにしたため、nil ガードが必須)
// ─────────────────────────────────────────────

func TestH1_DetectDormantOwners_NilSettingsSvc_DoesNotPanic(t *testing.T) {
	const clinicID = uint64(99)

	// settingsSvc=nil で構築（NewLstepBatchService の第6引数を nil に）
	withThresholdsCallCount := 0
	tagSyncSpy := &batchMockTagSyncSvc{
		syncDormantTagsWithThresholdsFn: func(_ context.Context, _, _ uint64, _ int, thresholds model.DormantThresholds) error {
			withThresholdsCallCount++
			// デフォルト閾値が渡されていることを確認する
			assert.Equal(t, 180, thresholds.Stage180, "nil settingsSvc は default Stage180=180 を使うべき")
			return nil
		},
	}
	entries := []repository.DormantOwnerEntry{
		{OwnerID: 1, DaysSince: 200},
		{OwnerID: 2, DaysSince: 250},
	}
	medRepo := &batchMockMedRecordRepo{
		findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
			return entries, nil
		},
	}

	svc := NewLstepBatchService(
		&batchMockReservationRepo{},
		tagSyncSpy,
		&mockClinicRepository{},
		medRepo,
		&batchMockAuditService{},
		nil, // settingsSvc = nil → H-1 パニック回帰
		nil,
	).(*lstepBatchService)

	// パニックしないことを assert.NotPanics で保証する
	assert.NotPanics(t, func() {
		count, errs := svc.detectDormantOwners(context.Background(), clinicID)
		assert.Equal(t, 2, count)
		assert.Empty(t, errs)
	}, "settingsSvc=nil でも detectDormantOwners はパニックしてはならない")

	assert.Equal(t, len(entries), withThresholdsCallCount,
		"nil settingsSvc でも全エントリに SyncDormantTagsWithThresholds を呼ぶべき")
}

// ─────────────────────────────────────────────
// PERF-2: DetectDormantOwners が
//   SyncDormantTagsWithThresholds を呼ぶこと (SyncDormantTags ではなく)
// ─────────────────────────────────────────────

func TestPERF2_DetectDormantOwners_CallsSyncWithThresholds(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	// --- spy: SyncDormantTagsWithThresholds の呼び出し回数 ---
	var withThresholdsCallCount int64
	tagSyncSpy := &batchMockTagSyncSvc{
		syncDormantTagsWithThresholdsFn: func(_ context.Context, _, _ uint64, _ int, _ model.DormantThresholds) error {
			atomic.AddInt64(&withThresholdsCallCount, 1)
			return nil
		},
	}

	// --- medRecord repo: ownerCount 件の dormant エントリを返す ---
	entries := make([]repository.DormantOwnerEntry, ownerCount)
	for i := range entries {
		entries[i] = repository.DormantOwnerEntry{OwnerID: uint64(i + 1), DaysSince: 200}
	}
	medRepo := &batchMockMedRecordRepo{
		findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
			return entries, nil
		},
	}

	batchSvc := newBatchService(
		&batchMockReservationRepo{},
		tagSyncSpy,
		&mockClinicRepository{},
		medRepo,
	)

	count, errs := batchSvc.detectDormantOwners(context.Background(), clinicID)
	_ = count
	_ = errs

	got := atomic.LoadInt64(&withThresholdsCallCount)
	// GREEN 条件: ownerCount 回 (各オーナーに SyncDormantTagsWithThresholds を呼ぶ)
	// RED  条件: 0 回 (現行は SyncDormantTags を呼ぶため SyncDormantTagsWithThresholds は未到達)
	assert.Equal(t, int64(ownerCount), got,
		"DetectDormantOwners は各オーナーに SyncDormantTagsWithThresholds を呼ぶべき (現在 %d 回)", got)
}

// ─────────────────────────────────────────────
// PERF-3: CreateCheckupSync が
//   FindByIDs を 1 回だけ呼ぶこと (FindByID を N 回でなく)
// ─────────────────────────────────────────────

// mockLstepClientForPerf3 は lstep.Client の最小スタブ（AddTag のみ有効）。
type mockLstepClientForPerf3 struct{}

func (m *mockLstepClientForPerf3) AddTag(_ context.Context, _, _ string) error    { return nil }
func (m *mockLstepClientForPerf3) RemoveTag(_ context.Context, _, _ string) error { return nil }
func (m *mockLstepClientForPerf3) GetUserTags(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLstepClientForPerf3) AddTagBulk(_ context.Context, _ []string, _ string) error {
	return nil
}
func (m *mockLstepClientForPerf3) GetUser(_ context.Context, _ string) (*lstep.UserInfo, error) {
	return nil, nil
}
func (m *mockLstepClientForPerf3) SetProperty(_ context.Context, _, _, _ string) error { return nil }

func TestPERF3_CreateCheckupSync_BatchFetchesOwners(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	// --- spy: FindByIDs の呼び出し回数 ---
	var findByIDsCallCount int64

	lineID := "line-u-test"
	ownerSlice := make([]*model.Owner, ownerCount)
	for i := range ownerSlice {
		o := &model.Owner{ID: uint64(i + 1), ClinicID: clinicID, LineUserID: &lineID}
		ownerSlice[i] = o
	}
	ownerRepoSpy := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
			atomic.AddInt64(&findByIDsCallCount, 1)
			return ownerSlice, nil
		},
		// RED フェーズ: 現行実装は FindByIDs でなく FindByID をループ内で呼ぶ
		findByIDFn: func(_ context.Context, _ uint64, id uint64) (*model.Owner, error) {
			o := &model.Owner{ID: id, ClinicID: clinicID, LineUserID: &lineID}
			return o, nil
		},
	}

	petRepoSpy := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ids []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ids))
			for _, id := range ids {
				m[id] = 1 // 全員 living pet あり → スキップしない
			}
			return m, nil
		},
	}

	svc := &checkupSyncService{
		ownerRepo:    ownerRepoSpy,
		petRepo:      petRepoSpy,
		tagCacheRepo: &mockLstepTagCacheRepository{},
		auditSvc:     &mockAuditService{},
		buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForPerf3{}, nil
		},
	}

	ownerIDs := make([]uint64, ownerCount)
	for i := range ownerIDs {
		ownerIDs[i] = uint64(i + 1)
	}
	input := CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    ownerIDs,
		TagName:     "checkup_perf_test", // system-managed プレフィックス外のタグ名
	}

	_, err := svc.CreateCheckupSync(context.Background(), clinicID, input, nil)
	assert.NoError(t, err)

	got := atomic.LoadInt64(&findByIDsCallCount)
	// GREEN 条件: 1 回 (バッチ化後)
	// RED  条件: 0 回 (現行は FindByID を N 回呼ぶため FindByIDs は未到達)
	assert.Equal(t, int64(1), got,
		"CreateCheckupSync は FindByIDs を 1 回だけ呼ぶべき (現在 %d 回)", got)
}
