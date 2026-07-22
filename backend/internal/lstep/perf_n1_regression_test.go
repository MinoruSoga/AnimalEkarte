package lstep

// perf_n1_regression_test.go — N+1 クエリ回帰テスト (PERF-2)
//
// RED フェーズ: 現行実装に対してアサーションが失敗することを確認してから GREEN 化する。
// 各テストは「呼び出し回数」だけを spy し、ビジネスロジックには触れない。

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
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
	entries := []medicalrecord.DormantOwnerEntry{
		{OwnerID: 1, DaysSince: 200},
		{OwnerID: 2, DaysSince: 250},
	}
	medRepo := &batchMockMedRecordRepo{
		findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
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
	entries := make([]medicalrecord.DormantOwnerEntry, ownerCount)
	for i := range entries {
		entries[i] = medicalrecord.DormantOwnerEntry{OwnerID: uint64(i + 1), DaysSince: 200}
	}
	medRepo := &batchMockMedRecordRepo{
		findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
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
