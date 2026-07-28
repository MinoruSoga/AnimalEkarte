package lstep

// line_send_log_repository_test.go — LineSendLogRepository 統合テスト。
//
// 保護する不変条件:
//   - FindByOwner は clinic_id + owner_id でスコープされる（他クリニック/他飼主の送信ログは混入しない）。
//   - FindByOwner は sent_at 降順で返す。
//   - FindByOwner の limit が有効に機能する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLineSendLogTestDB は line_send_logs テーブルを整備する。
func setupLineSendLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LineSendLog{}))
	db.Exec("TRUNCATE TABLE line_send_logs CASCADE")
	return db
}

func makeLineSendLog(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, messageType string, sentAt time.Time) *model.LineSendLog {
	t.Helper()
	l := &model.LineSendLog{
		ClinicID:       clinicID,
		OwnerID:        ownerID,
		SentByUserID:   1,
		MessageType:    messageType,
		ContentSummary: "summary:" + messageType,
		Status:         "sent",
		SentAt:         sentAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(l).Error)
	return l
}

func TestLineSendLogRepository_Create(t *testing.T) {
	db := setupLineSendLogTestDB(t)
	repo := NewLineSendLogRepository(db)
	ctx := context.Background()

	log := &model.LineSendLog{
		ClinicID:       1,
		OwnerID:        10,
		SentByUserID:   2,
		MessageType:    "reminder",
		ContentSummary: "来院リマインド",
		Status:         "sent",
		SentAt:         time.Now(),
	}
	require.NoError(t, repo.Create(ctx, log))
	assert.NotZero(t, log.ID)

	var count int64
	require.NoError(t, db.Model(&model.LineSendLog{}).Where("id = ?", log.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestLineSendLogRepository_FindByOwner(t *testing.T) {
	db := setupLineSendLogTestDB(t)
	repo := NewLineSendLogRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
		ownerA  = uint64(10)
		ownerA2 = uint64(11)
	)

	older := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	middle := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 3, 9, 0, 0, 0, time.UTC)

	makeLineSendLog(t, db, clinicA, ownerA, "reminder", older)
	makeLineSendLog(t, db, clinicA, ownerA, "reminder", newer)
	makeLineSendLog(t, db, clinicA, ownerA, "reminder", middle)
	// 別オーナー（混入してはならない）
	makeLineSendLog(t, db, clinicA, ownerA2, "reminder", newer)
	// 別クリニックの同一オーナーID（混入してはならない）
	makeLineSendLog(t, db, clinicB, ownerA, "reminder", newer)

	t.Run("returns only same clinic and owner logs ordered by sent_at DESC", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, ownerA, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.True(t, got[0].SentAt.After(got[1].SentAt) || got[0].SentAt.Equal(got[1].SentAt))
		assert.True(t, got[1].SentAt.After(got[2].SentAt) || got[1].SentAt.Equal(got[2].SentAt))
		assert.WithinDuration(t, newer, got[0].SentAt, time.Second)
		assert.WithinDuration(t, older, got[2].SentAt, time.Second)
	})

	t.Run("limit restricts the number of results", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, ownerA, 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.WithinDuration(t, newer, got[0].SentAt, time.Second)
	})

	t.Run("different owner does not see other owner's logs", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, ownerA2, 100)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("empty result for owner with no logs", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, uint64(999999), 100)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
