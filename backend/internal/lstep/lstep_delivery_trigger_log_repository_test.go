package lstep

// lstep_delivery_trigger_log_repository_test.go — LstepDeliveryTriggerLogRepository 統合テスト。
//
// 保護する不変条件:
//   - 当日・同一飼主・同一トリガー種別の二重発火防止 (ExistsTodayByOwnerAndType)。
//   - UpdateStatus / UpdateSuppressed は clinic_id 不一致・存在しない ID で NotFound を返す（IDOR 防止）。
//   - 期間集計系 (CountByStatusAndDateRange 等) は clinic_id で正しく分離される。
//   - CountVisitConversionsByType は fired かつ非抑制ログのみを分母にし、期間内の有効カルテ存在を分子にする。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

// setupLstepDeliveryTriggerLogTestDB は lstep_delivery_trigger_log テーブルを用意する。
// setupTestDB がすでに owners / medical_records を migrate・truncate 済みのため、
// FindByDateRangeWithFilters の owner JOIN や CountVisitConversionsByType の
// medical_records EXISTS サブクエリもそのまま利用できる。
func setupLstepDeliveryTriggerLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := repotest.SetupTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.LstepDeliveryTriggerLog{}))
	db.Exec("TRUNCATE TABLE lstep_delivery_trigger_log CASCADE")
	return db
}

// makeDeliveryTriggerLog はテスト用トリガーログを作成して返す。
func makeDeliveryTriggerLog(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, triggerType, status string, scheduledAt time.Time) *model.LstepDeliveryTriggerLog {
	t.Helper()
	log := &model.LstepDeliveryTriggerLog{
		ClinicID:    clinicID,
		OwnerID:     ownerID,
		TriggerType: triggerType,
		Status:      status,
		ScheduledAt: scheduledAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(log).Error)
	return log
}

func TestLstepDeliveryTriggerLogRepository_Create(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	scheduledAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	log := &model.LstepDeliveryTriggerLog{
		ClinicID:    1,
		OwnerID:     100,
		TriggerType: model.TriggerTypeBirthdayMessage,
		Status:      model.TriggerStatusScheduled,
		ScheduledAt: scheduledAt,
	}
	require.NoError(t, repo.Create(ctx, log))
	assert.NotZero(t, log.ID, "作成後は ID が採番されるべき")

	var stored model.LstepDeliveryTriggerLog
	require.NoError(t, db.First(&stored, log.ID).Error)
	assert.Equal(t, uint64(1), stored.ClinicID)
	assert.Equal(t, uint64(100), stored.OwnerID)
	assert.Equal(t, model.TriggerTypeBirthdayMessage, stored.TriggerType)
	assert.Equal(t, model.TriggerStatusScheduled, stored.Status)
}

func TestLstepDeliveryTriggerLogRepository_ExistsTodayByOwnerAndType(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	const ownerA, ownerB = uint64(10), uint64(20)
	target := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)

	makeDeliveryTriggerLog(t, db, clinicA, ownerA, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target)

	t.Run("同日・同飼主・同種別は既存扱い", func(t *testing.T) {
		exists, err := repo.ExistsTodayByOwnerAndType(ctx, clinicA, ownerA, model.TriggerTypeBirthdayMessage, target)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("別のトリガー種別は既存扱いしない", func(t *testing.T) {
		exists, err := repo.ExistsTodayByOwnerAndType(ctx, clinicA, ownerA, model.TriggerTypeVaccineDeadline30, target)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("別の飼主は既存扱いしない", func(t *testing.T) {
		exists, err := repo.ExistsTodayByOwnerAndType(ctx, clinicA, ownerB, model.TriggerTypeBirthdayMessage, target)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("日付境界外（前日）は既存扱いしない", func(t *testing.T) {
		previousDay := target.AddDate(0, 0, -1)
		exists, err := repo.ExistsTodayByOwnerAndType(ctx, clinicA, ownerA, model.TriggerTypeBirthdayMessage, previousDay)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("別クリニックの同条件ログは既存扱いしない（clinic_id 分離）", func(t *testing.T) {
		exists, err := repo.ExistsTodayByOwnerAndType(ctx, clinicB, ownerA, model.TriggerTypeBirthdayMessage, target)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestLstepDeliveryTriggerLogRepository_UpdateStatus(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	scheduledAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)

	t.Run("ステータス・fired_at・excluded_reason を更新できる", func(t *testing.T) {
		log := makeDeliveryTriggerLog(t, db, clinicA, 10, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, scheduledAt)
		firedAt := scheduledAt.Add(time.Hour)
		reason := "opted_out"

		require.NoError(t, repo.UpdateStatus(ctx, clinicA, log.ID, model.TriggerStatusExcluded, &firedAt, &reason))

		var stored model.LstepDeliveryTriggerLog
		require.NoError(t, db.First(&stored, log.ID).Error)
		assert.Equal(t, model.TriggerStatusExcluded, stored.Status)
		require.NotNil(t, stored.FiredAt)
		assert.WithinDuration(t, firedAt, *stored.FiredAt, time.Second)
		require.NotNil(t, stored.ExcludedReason)
		assert.Equal(t, reason, *stored.ExcludedReason)
	})

	t.Run("firedAt/excludedReason が nil の場合は status のみ更新される", func(t *testing.T) {
		log := makeDeliveryTriggerLog(t, db, clinicA, 12, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, scheduledAt)

		require.NoError(t, repo.UpdateStatus(ctx, clinicA, log.ID, model.TriggerStatusFired, nil, nil))

		var stored model.LstepDeliveryTriggerLog
		require.NoError(t, db.First(&stored, log.ID).Error)
		assert.Equal(t, model.TriggerStatusFired, stored.Status)
		assert.Nil(t, stored.FiredAt, "firedAt を渡さなければ更新されないべき")
		assert.Nil(t, stored.ExcludedReason, "excludedReason を渡さなければ更新されないべき")
	})

	t.Run("存在しない ID は NotFound を返す", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, clinicA, 999999, model.TriggerStatusFired, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックの ID 指定は NotFound を返す（clinic_id 分離）", func(t *testing.T) {
		log := makeDeliveryTriggerLog(t, db, clinicA, 11, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, scheduledAt)
		err := repo.UpdateStatus(ctx, clinicB, log.ID, model.TriggerStatusFired, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからの更新は NotFound であるべき: %v", err)

		// 実データは変更されていないことを確認
		var stored model.LstepDeliveryTriggerLog
		require.NoError(t, db.First(&stored, log.ID).Error)
		assert.Equal(t, model.TriggerStatusScheduled, stored.Status)
	})
}

func TestLstepDeliveryTriggerLogRepository_CountByStatusAndDateRange(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	within := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	ownerA := repotest.MakeTestOwner(t, db, clinicA, "集計対象飼主")
	ownerB := repotest.MakeTestOwner(t, db, clinicB, "別医院飼主")

	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeVaccineDeadline30, model.TriggerStatusExcluded, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, from.AddDate(0, -1, 0)) // 範囲外
	makeDeliveryTriggerLog(t, db, clinicB, ownerB.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)                 // 別クリニック
	makeDeliveryTriggerLog(t, db, clinicA, ownerB.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)                 // 不整合owner参照

	t.Run("トリガー種別指定なしは全種別を集計する", func(t *testing.T) {
		counts, err := repo.CountByStatusAndDateRange(ctx, clinicA, from, to, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), counts[model.TriggerStatusFired])
		assert.Equal(t, int64(1), counts[model.TriggerStatusExcluded])
	})

	t.Run("トリガー種別を指定すると絞り込まれる", func(t *testing.T) {
		counts, err := repo.CountByStatusAndDateRange(ctx, clinicA, from, to, model.TriggerTypeVaccineDeadline30)
		require.NoError(t, err)
		assert.Equal(t, int64(0), counts[model.TriggerStatusFired])
		assert.Equal(t, int64(1), counts[model.TriggerStatusExcluded])
	})

	t.Run("該当データなしは空マップを返す", func(t *testing.T) {
		counts, err := repo.CountByStatusAndDateRange(ctx, clinicB, from, from.AddDate(0, 0, 1), "")
		require.NoError(t, err)
		assert.Empty(t, counts)
	})
}

func TestLstepDeliveryTriggerLogRepository_CountExcludedReasonByDateRange(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	within := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	ownerA := repotest.MakeTestOwner(t, db, clinicA, "除外集計対象飼主")
	ownerB := repotest.MakeTestOwner(t, db, 2, "別医院除外飼主")

	reasonOptOut := "opted_out"
	log1 := makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusExcluded, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", log1.ID).
		Update("excluded_reason", reasonOptOut).Error)

	log2 := makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusExcluded, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", log2.ID).
		Update("excluded_reason", reasonOptOut).Error)

	// excluded_reason が NULL のまま除外されたログ（COALESCE で '' 集計になる）
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusExcluded, within)

	// 除外ではないログは集計対象外
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	crossOwnerLog := makeDeliveryTriggerLog(t, db, clinicA, ownerB.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusExcluded, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", crossOwnerLog.ID).
		Update("excluded_reason", reasonOptOut).Error)

	counts, err := repo.CountExcludedReasonByDateRange(ctx, clinicA, from, to, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts[reasonOptOut])
	assert.Equal(t, int64(1), counts[""])
}

func TestLstepDeliveryTriggerLogRepository_CountSuppressedByPriorityDateRange(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	within := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	ownerA := repotest.MakeTestOwner(t, db, clinicA, "抑制集計対象飼主")
	ownerB := repotest.MakeTestOwner(t, db, 2, "別医院抑制飼主")

	suppressed1 := makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", suppressed1.ID).
		Update("suppressed_by_priority", true).Error)
	suppressed2 := makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeVaccineDeadline30, model.TriggerStatusFired, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", suppressed2.ID).
		Update("suppressed_by_priority", true).Error)
	// 通常配信ログ（suppressed_by_priority=false）は件数に含まれない
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	crossOwnerLog := makeDeliveryTriggerLog(t, db, clinicA, ownerB.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", crossOwnerLog.ID).
		Update("suppressed_by_priority", true).Error)

	t.Run("トリガー種別指定なしは全種別を集計する（非ゼロ）", func(t *testing.T) {
		count, err := repo.CountSuppressedByPriorityDateRange(ctx, clinicA, from, to, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("トリガー種別を指定すると絞り込まれる", func(t *testing.T) {
		count, err := repo.CountSuppressedByPriorityDateRange(ctx, clinicA, from, to, model.TriggerTypeVaccineDeadline30)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("該当なしはゼロを返す", func(t *testing.T) {
		count, err := repo.CountSuppressedByPriorityDateRange(ctx, clinicA, from, to, model.TriggerTypeFilariaAlert)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestLstepDeliveryTriggerLogRepository_FindByDateRangeWithFilters(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	owner := repotest.MakeTestOwner(t, db, clinicA, "山田花子")

	// 新しい順に3件（DESC で返る想定）。owner が同じ clinic に属するログだけが返る。
	l1 := makeDeliveryTriggerLog(t, db, clinicA, owner.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC))
	l2 := makeDeliveryTriggerLog(t, db, clinicA, owner.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC))
	makeDeliveryTriggerLog(t, db, clinicA, 99999, model.TriggerTypeVaccineDeadline30, model.TriggerStatusExcluded, time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)) // 存在しない owner_id

	t.Run("owner 名が JOIN される（存在しない owner のログは除外）", func(t *testing.T) {
		rows, total, err := repo.FindByDateRangeWithFilters(ctx, clinicA, from, to, "", "", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, rows, 2)
		// DESC 順: l2, l1（owner relation が不正な l3 は除外）
		assert.Equal(t, l2.ID, rows[0].ID)
		assert.Equal(t, owner.Name, rows[0].OwnerName)
		assert.Equal(t, l1.ID, rows[1].ID)
		assert.Equal(t, owner.Name, rows[1].OwnerName)
	})

	t.Run("trigger_type と status で絞り込める", func(t *testing.T) {
		rows, total, err := repo.FindByDateRangeWithFilters(ctx, clinicA, from, to, model.TriggerTypeVaccineDeadline30, model.TriggerStatusExcluded, 10, 0)
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.Empty(t, rows)
	})

	t.Run("limit/offset でページングされる（total はページング前の総数）", func(t *testing.T) {
		rows, total, err := repo.FindByDateRangeWithFilters(ctx, clinicA, from, to, "", "", 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "total は所有関係・フィルタ適用後かつページング前の件数であるべき")
		require.Len(t, rows, 1)
		assert.Equal(t, l1.ID, rows[0].ID, "offset=1 の2番目の有効行が返るべき")
	})

	t.Run("別クリニック owner を指す不整合ログから owner 名を露出しない", func(t *testing.T) {
		const clinicB = uint64(2)
		otherClinicOwner := repotest.MakeTestOwner(t, db, clinicB, "別クリニック飼主")
		log := makeDeliveryTriggerLog(t, db, clinicA, otherClinicOwner.ID, "cross_clinic_owner", model.TriggerStatusScheduled, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC))

		rows, total, err := repo.FindByDateRangeWithFilters(ctx, clinicA, from, to, "cross_clinic_owner", model.TriggerStatusScheduled, 10, 0)
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.Empty(t, rows)
		assert.NotZero(t, log.ID, "fixture must persist the malformed relation")
	})
}

func TestLstepDeliveryTriggerLogRepository_CountByTypeAndStatus(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	within := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	ownerA := repotest.MakeTestOwner(t, db, clinicA, "種別集計対象飼主")
	ownerB := repotest.MakeTestOwner(t, db, 2, "別医院種別飼主")

	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusFired, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeBirthdayMessage, model.TriggerStatusExcluded, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerA.ID, model.TriggerTypeVaccineDeadline30, model.TriggerStatusFired, within)
	makeDeliveryTriggerLog(t, db, clinicA, ownerB.ID, "cross_owner", model.TriggerStatusFired, within)

	rows, err := repo.CountByTypeAndStatus(ctx, clinicA, from, to)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	found := make(map[string]int64)
	for _, r := range rows {
		found[r.TriggerType+"|"+r.Status] = r.Count
	}
	assert.Equal(t, int64(2), found[model.TriggerTypeBirthdayMessage+"|"+model.TriggerStatusFired])
	assert.Equal(t, int64(1), found[model.TriggerTypeBirthdayMessage+"|"+model.TriggerStatusExcluded])
	assert.Equal(t, int64(1), found[model.TriggerTypeVaccineDeadline30+"|"+model.TriggerStatusFired])
}

// TestLstepDeliveryTriggerLogRepository_CountVisitConversionsByType は
// fired かつ非抑制ログを分母、期間内の有効カルテ存在を分子として来院転換率の元データを集計する不変条件を検証する。
func TestLstepDeliveryTriggerLogRepository_CountVisitConversionsByType(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.Owner{}, &model.MedicalRecord{}))
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	// 19:00 JST（UTC+9）になるよう UTC 10:00 に設定し、日付境界の揺れを避ける。
	firedAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	const conversionWindowDays = 3

	// T1: 来院した飼主 — fired ログ + 期間内の有効カルテ → visited
	visitedOwner := repotest.MakeTestOwner(t, db, clinicA, "来院した飼主")
	makeDeliveryTriggerLog(t, db, clinicA, visitedOwner.ID, "T1", model.TriggerStatusFired, firedAt)
	visitedOwnerID := visitedOwner.ID
	require.NoError(t, db.Create(&model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "R-VISITED",
		Date:     time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC), // fired_at + 2日（3日以内）
		OwnerID:  &visitedOwnerID,
	}).Error)

	// T1: 来院しなかった飼主 — fired ログのみ、カルテなし → not visited
	notVisitedOwner := repotest.MakeTestOwner(t, db, clinicA, "来院しなかった飼主")
	makeDeliveryTriggerLog(t, db, clinicA, notVisitedOwner.ID, "T1", model.TriggerStatusFired, firedAt)

	// T2: カルテはあるがソフトデリート済み → not visited（deleted_at IS NULL 条件）
	deletedRecordOwner := repotest.MakeTestOwner(t, db, clinicA, "削除カルテの飼主")
	makeDeliveryTriggerLog(t, db, clinicA, deletedRecordOwner.ID, "T2", model.TriggerStatusFired, firedAt)
	deletedOwnerID := deletedRecordOwner.ID
	deletedRecord := &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "R-DELETED",
		Date:     time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		OwnerID:  &deletedOwnerID,
	}
	require.NoError(t, db.Create(deletedRecord).Error)
	require.NoError(t, db.Delete(deletedRecord).Error)

	// T3: 医院Aの配信対象に対し、医院Bの不整合カルテが同じ owner_id を参照しても来院扱いにしない。
	crossClinicRecordOwner := repotest.MakeTestOwner(t, db, clinicA, "別医院カルテ誤参照の飼主")
	makeDeliveryTriggerLog(t, db, clinicA, crossClinicRecordOwner.ID, "T3", model.TriggerStatusFired, firedAt)
	crossClinicOwnerID := crossClinicRecordOwner.ID
	require.NoError(t, db.Create(&model.MedicalRecord{
		ClinicID: clinicB,
		RecordNo: "R-CROSS-CLINIC",
		Date:     time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		OwnerID:  &crossClinicOwnerID,
	}).Error)

	// T1: 抑制されたログは delivered からも除外される
	suppressedOwner := repotest.MakeTestOwner(t, db, clinicA, "抑制された飼主")
	suppressedLog := makeDeliveryTriggerLog(t, db, clinicA, suppressedOwner.ID, "T1", model.TriggerStatusFired, firedAt)
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).Where("id = ?", suppressedLog.ID).
		Update("suppressed_by_priority", true).Error)

	// T1: scheduled のまま（未配信）は delivered から除外される
	scheduledOwner := repotest.MakeTestOwner(t, db, clinicA, "未配信の飼主")
	makeDeliveryTriggerLog(t, db, clinicA, scheduledOwner.ID, "T1", model.TriggerStatusScheduled, firedAt)

	// 別クリニックのログは集計対象外
	otherClinicOwner := repotest.MakeTestOwner(t, db, clinicB, "別クリニック飼主")
	makeDeliveryTriggerLog(t, db, clinicB, otherClinicOwner.ID, "T1", model.TriggerStatusFired, firedAt)
	makeDeliveryTriggerLog(t, db, clinicA, otherClinicOwner.ID, "CROSS_OWNER", model.TriggerStatusFired, firedAt)

	// makeDeliveryTriggerLog は scheduled_at のみ設定し fired_at は NULL のまま作成する
	// （他の多数のテストがこの挙動に依存しているため共通ヘルパー自体は変更しない）。
	// CountVisitConversionsByType は l.fired_at IS NOT NULL かつ範囲内であることを要求するため、
	// このテストでは status=fired の全行に対して fired_at を明示的に設定する。
	require.NoError(t, db.Model(&model.LstepDeliveryTriggerLog{}).
		Where("clinic_id IN ? AND status = ?", []uint64{clinicA, clinicB}, model.TriggerStatusFired).
		Update("fired_at", firedAt).Error)

	rows, err := repo.CountVisitConversionsByType(ctx, clinicA, from, to, conversionWindowDays)
	require.NoError(t, err)

	byType := make(map[string]VisitConversionRow)
	for _, r := range rows {
		byType[r.TriggerType] = r
	}

	require.Contains(t, byType, "T1")
	assert.Equal(t, int64(2), byType["T1"].DeliveredCount, "抑制・未配信ログを除いた fired ログ数であるべき")
	assert.Equal(t, int64(1), byType["T1"].VisitedCount, "来院した飼主のみカウントされるべき")

	require.Contains(t, byType, "T2")
	assert.Equal(t, int64(1), byType["T2"].DeliveredCount)
	assert.Equal(t, int64(0), byType["T2"].VisitedCount, "ソフトデリート済みカルテは来院とみなさないべき")

	require.Contains(t, byType, "T3")
	assert.Equal(t, int64(1), byType["T3"].DeliveredCount)
	assert.Equal(t, int64(0), byType["T3"].VisitedCount, "別医院の不整合カルテは来院とみなさないべき")
	assert.NotContains(t, byType, "CROSS_OWNER", "別医院ownerを参照する不整合ログは分母から除外すべき")
}

func TestLstepDeliveryTriggerLogRepository_FindByOwnerAndDate(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	const ownerA, ownerB = uint64(10), uint64(20)
	target := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	match := makeDeliveryTriggerLog(t, db, clinicA, ownerA, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target.Add(9*time.Hour))
	makeDeliveryTriggerLog(t, db, clinicA, ownerB, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target.Add(9*time.Hour))                   // 別飼主
	makeDeliveryTriggerLog(t, db, clinicB, ownerA, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target.Add(9*time.Hour))                   // 別クリニック
	makeDeliveryTriggerLog(t, db, clinicA, ownerA, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target.AddDate(0, 0, -1).Add(9*time.Hour)) // 前日
	makeDeliveryTriggerLog(t, db, clinicA, ownerA, model.TriggerTypeBirthdayMessage, model.TriggerStatusScheduled, target.AddDate(0, 0, 1).Add(1*time.Hour))  // 翌日

	logs, err := repo.FindByOwnerAndDate(ctx, clinicA, ownerA, target)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, match.ID, logs[0].ID)
}

func TestLstepDeliveryTriggerLogRepository_UpdateSuppressed(t *testing.T) {
	db := setupLstepDeliveryTriggerLogTestDB(t)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	scheduledAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)

	t.Run("suppressed_by_priority と理由が更新される", func(t *testing.T) {
		log := makeDeliveryTriggerLog(t, db, clinicA, 10, model.TriggerTypeDormantPrevention365, model.TriggerStatusFired, scheduledAt)
		reason := "owner_id=10 already triggered by dormant_prevention_180d"

		require.NoError(t, repo.UpdateSuppressed(ctx, clinicA, log.ID, reason))

		var stored model.LstepDeliveryTriggerLog
		require.NoError(t, db.First(&stored, log.ID).Error)
		assert.True(t, stored.SuppressedByPriority)
		require.NotNil(t, stored.SuppressionReason)
		assert.Equal(t, reason, *stored.SuppressionReason)
	})

	t.Run("存在しない ID は NotFound を返す", func(t *testing.T) {
		err := repo.UpdateSuppressed(ctx, clinicA, 999999, "reason")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックの ID 指定は NotFound を返す（clinic_id 分離）", func(t *testing.T) {
		log := makeDeliveryTriggerLog(t, db, clinicA, 11, model.TriggerTypeDormantPrevention365, model.TriggerStatusFired, scheduledAt)
		err := repo.UpdateSuppressed(ctx, clinicB, log.ID, "reason")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからの更新は NotFound であるべき: %v", err)
	})
}
