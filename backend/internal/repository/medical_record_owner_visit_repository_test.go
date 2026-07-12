package repository

// medical_record_owner_visit_repository_test.go — FindDormantOwnerEntriesCursor
// (PERF-FOLLOWUP-02: カーソルページネーション) のページ境界テスト。
// setupTestDB は medical_records / owners を TRUNCATE 済みの状態で返すため、
// 追加の TRUNCATE は不要。

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// seedDormantOwners は clinicID に休眠飼い主（最終来院が oldDate）を count 件一括作成し、
// 作成した owner_id の一覧を返す。
func seedDormantOwners(t *testing.T, db *gorm.DB, clinicID uint64, count int, oldDate time.Time) []uint64 {
	t.Helper()
	ctx := context.Background()

	owners := make([]model.Owner, count)
	for i := range owners {
		owners[i] = model.Owner{ClinicID: clinicID, Name: fmt.Sprintf("休眠飼主%d", i)}
	}
	require.NoError(t, db.WithContext(ctx).CreateInBatches(&owners, 200).Error)

	records := make([]model.MedicalRecord, count)
	for i := range owners {
		oid := owners[i].ID
		records[i] = model.MedicalRecord{
			ClinicID: clinicID,
			RecordNo: fmt.Sprintf("R-%d-%d", clinicID, i),
			Date:     oldDate,
			OwnerID:  &oid,
		}
	}
	require.NoError(t, db.WithContext(ctx).CreateInBatches(&records, 200).Error)

	ids := make([]uint64, count)
	for i, o := range owners {
		ids[i] = o.ID
	}
	return ids
}

func TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_ExactlyOnePage(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	const pageSize = 500
	const minDaysSince = 180

	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)
	seedDormantOwners(t, db, clinicA, pageSize, oldDate)

	page1, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, 0, pageSize)
	require.NoError(t, err)
	require.Len(t, page1, pageSize, "ちょうど pageSize 件は1ページに収まる")

	afterOwnerID := page1[len(page1)-1].OwnerID
	page2, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, afterOwnerID, pageSize)
	require.NoError(t, err)
	assert.Empty(t, page2, "全件消化後の次ページ取得は空を返す")

	seen := make(map[uint64]bool, pageSize)
	prev := uint64(0)
	for _, e := range page1 {
		assert.False(t, seen[e.OwnerID], "重複 owner_id が含まれてはいけない")
		seen[e.OwnerID] = true
		assert.Greater(t, e.OwnerID, prev, "owner_id は昇順でなければならない")
		prev = e.OwnerID
	}
}

func TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_TwoPages(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	const pageSize = 500
	const total = pageSize + 1
	const minDaysSince = 180

	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)
	seedDormantOwners(t, db, clinicA, total, oldDate)

	page1, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, 0, pageSize)
	require.NoError(t, err)
	require.Len(t, page1, pageSize)

	afterOwnerID := page1[len(page1)-1].OwnerID
	page2, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, afterOwnerID, pageSize)
	require.NoError(t, err)
	require.Len(t, page2, 1, "501 件目は2ページ目に1件だけ現れる")

	seen := make(map[uint64]bool, total)
	for _, e := range append(page1, page2...) {
		assert.False(t, seen[e.OwnerID], "重複 owner_id が含まれてはいけない")
		seen[e.OwnerID] = true
	}
	assert.Len(t, seen, total)

	page3, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, page2[len(page2)-1].OwnerID, pageSize)
	require.NoError(t, err)
	assert.Empty(t, page3, "全件消化後の次ページ取得は空を返す")
}

// TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_ClinicIsolation は
// 別クリニックの休眠飼い主が混入しないことを検証する（clinicScope の回帰防止）。
func TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_ClinicIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const minDaysSince = 180

	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)
	seedDormantOwners(t, db, clinicA, 1, oldDate)
	seedDormantOwners(t, db, clinicB, 1, oldDate)

	got, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, 0, 500)
	require.NoError(t, err)
	require.Len(t, got, 1, "自医院の休眠飼い主のみ返る")
}
