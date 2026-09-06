package medicalrecord

// medical_record_owner_visit_repository_test.go — FindDormantOwnerEntriesCursor
// (PERF-FOLLOWUP-02: カーソルページネーション) のページ境界テスト。
// setupTestDB は medical_records / owners を TRUNCATE 済みの状態で返すため、
// 追加の TRUNCATE は不要。

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
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

	species := &model.AnimalSpecies{Name: fmt.Sprintf("休眠動物種-%d-%d", clinicID, oldDate.UnixNano())}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pets := make([]model.Pet, count)
	for i := range owners {
		pets[i] = model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owners[i].ID,
			AnimalSpeciesID: species.ID,
			Name:            fmt.Sprintf("休眠ペット%d", i),
		}
	}
	require.NoError(t, db.WithContext(ctx).CreateInBatches(&pets, 200).Error)

	records := make([]model.MedicalRecord, count)
	for i := range owners {
		oid := owners[i].ID
		pid := pets[i].ID
		records[i] = model.MedicalRecord{
			ClinicID: clinicID,
			RecordNo: fmt.Sprintf("R-%d-%d", clinicID, i),
			Date:     oldDate,
			OwnerID:  &oid,
			PetID:    &pid,
		}
	}
	require.NoError(t, db.WithContext(ctx).CreateInBatches(&records, 200).Error)

	ids := make([]uint64, count)
	for i := range owners {
		ids[i] = owners[i].ID
	}
	return ids
}

func TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_ExactlyOnePage(t *testing.T) {
	db := testdb.SetupTestDB(t)
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

func TestMedicalRecordRepository_OwnerVisitReads_CurrentOwnerAfterTransfer(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70105)

	visitDate := time.Now().In(time.Local).AddDate(0, 0, -200)
	fixture := makeCurrentOwnerTransferFixture(
		t,
		db,
		clinicID,
		"MR-VISIT-CURRENT-OWNER",
		visitDate,
	)

	latest, err := repo.FindLatestByOwner(ctx, clinicID, fixture.CurrentOwner.ID)
	require.NoError(t, err)
	require.NotNil(t, latest)
	require.NotNil(t, latest.OwnerID)
	assert.Equal(t, fixture.PreviousOwner.ID, *latest.OwnerID, "returned owner_id remains the historical snapshot")

	summary, err := repo.FindOwnerVisitSummary(ctx, clinicID, fixture.CurrentOwner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), summary.TotalCount)
	previousSummary, err := repo.FindOwnerVisitSummary(ctx, clinicID, fixture.PreviousOwner.ID)
	require.NoError(t, err)
	assert.Zero(t, previousSummary.TotalCount)

	var cursorRepo DormantOwnerEntriesAtRepository = repo
	entries, err := cursorRepo.FindDormantOwnerEntriesCursorAt(ctx, clinicID, 180, 0, 1, time.Now())
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, fixture.CurrentOwner.ID, entries[0].OwnerID)
}

func TestMedicalRecordRepository_FindDormantOwnerEntriesCursor_TwoPages(t *testing.T) {
	db := testdb.SetupTestDB(t)
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
	db := testdb.SetupTestDB(t)
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

func TestMedicalRecordRepository_FindDormantOwnerEntriesCursorAt_UsesSuppliedEvaluationTime(t *testing.T) {
	db := testdb.SetupTestDB(t)
	store := NewMedicalRecordRepository(db)
	var repo DormantOwnerEntriesAtRepository = store
	ctx := context.Background()
	const clinicID = uint64(1)
	const minDaysSince = 180

	evaluatedAt := time.Date(2030, 1, 2, 10, 0, 0, 0, time.Local)
	eligibleIDs := seedDormantOwners(
		t,
		db,
		clinicID,
		1,
		evaluatedAt.AddDate(0, 0, -(minDaysSince+1)),
	)
	recentIDs := seedDormantOwners(
		t,
		db,
		clinicID,
		1,
		evaluatedAt.AddDate(0, 0, -(minDaysSince-1)),
	)

	got, err := repo.FindDormantOwnerEntriesCursorAt(
		ctx,
		clinicID,
		minDaysSince,
		0,
		500,
		evaluatedAt,
	)

	require.NoError(t, err)
	ids := make([]uint64, 0, len(got))
	for _, entry := range got {
		ids = append(ids, entry.OwnerID)
		assert.Equal(t, minDaysSince+1, entry.DaysSince)
	}
	assert.Contains(t, ids, eligibleIDs[0])
	assert.NotContains(t, ids, recentIDs[0])
}

func TestMedicalRecordRepository_FindDormantOwnerEntries_RequiresActiveOwnerInSameClinic(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const minDaysSince = 180
	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)

	validOwner := makeTestOwner(t, db, clinicA, "有効な休眠飼主")
	crossClinicOwner := makeTestOwner(t, db, clinicB, "別医院の誤参照飼主")
	deletedOwner := makeTestOwner(t, db, clinicA, "削除済み飼主")
	require.NoError(t, db.Delete(deletedOwner).Error)

	for _, ownerID := range []*uint64{&validOwner.ID, &crossClinicOwner.ID, &deletedOwner.ID, nil} {
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA,
			OwnerID:  ownerID,
			Date:     oldDate,
		})
	}

	want := []DormantOwnerEntry{{OwnerID: validOwner.ID}}
	assertOnlyValidOwner := func(t *testing.T, got []DormantOwnerEntry) {
		t.Helper()
		require.Len(t, got, len(want))
		assert.Equal(t, want[0].OwnerID, got[0].OwnerID)
	}

	got, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
	require.NoError(t, err)
	assertOnlyValidOwner(t, got)

	cursorGot, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, 0, 500)
	require.NoError(t, err)
	assertOnlyValidOwner(t, cursorGot)
}

// ---------------------------------------------------------------------------
// 以下、FindLatestByOwner / FindOwnerVisitSummary / FindOwnersByFirstVisitDate /
// FindOwnersByLastVisitDays / FindOwnersByNextVisitRecommended /
// FindDormantOwnerEntries のテスト（#212 カバレッジ向上）。
// ---------------------------------------------------------------------------

var ownerVisitRecordSeq int64

// makeVisitRecordForOwnerVisitTest は本ファイル専用のミニマル model.MedicalRecord 作成ヘルパー。
// rec に ClinicID / OwnerID / Date 等を設定して渡す。PetID が空で、同一医院の有効な
// OwnerID が存在する正常系 fixture では current-owner read 用のペットを明示的に補う。
// 別医院 owner・削除 owner・nil owner の corrupt fixture には補わず、負例を維持する。
// RecordNo が空なら自動採番する
// （medical_records は clinic_id + record_no の複合 UNIQUE INDEX を持つ）。
func makeVisitRecordForOwnerVisitTest(t *testing.T, db *gorm.DB, rec *model.MedicalRecord) *model.MedicalRecord {
	t.Helper()
	if rec.PetID == nil && rec.OwnerID != nil {
		var owner model.Owner
		err := db.WithContext(context.Background()).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *rec.OwnerID, rec.ClinicID).
			First(&owner).Error
		if err == nil {
			pet := makeSpeciesAndPet(
				t,
				db,
				rec.ClinicID,
				owner.ID,
				fmt.Sprintf("来院ペット-%d", atomic.AddInt64(&ownerVisitRecordSeq, 1)),
			)
			rec.PetID = &pet.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			require.NoError(t, err)
		}
	}
	if rec.RecordNo == "" {
		seq := atomic.AddInt64(&ownerVisitRecordSeq, 1)
		rec.RecordNo = fmt.Sprintf("R-OVT-%d-%d", rec.ClinicID, seq)
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rec).Error)
	return rec
}

func TestMedicalRecordRepository_FindLatestByOwner(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("正常系: created_at DESC で最新カルテを返す", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "最新カルテ飼主")
		ownerID := owner.ID

		older := makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID:  clinicA,
			OwnerID:   &ownerID,
			Date:      time.Now().In(time.Local).AddDate(0, 0, -30),
			CreatedAt: time.Now().In(time.Local).AddDate(0, 0, -30),
		})
		newer := makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID:  clinicA,
			OwnerID:   &ownerID,
			Date:      time.Now().In(time.Local),
			CreatedAt: time.Now().In(time.Local),
		})

		got, err := repo.FindLatestByOwner(ctx, clinicA, ownerID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, newer.ID, got.ID)
		assert.NotEqual(t, older.ID, got.ID)
	})

	// #236 BUG#4 (2026-07-13): FindLatestByOwner は修正済み。
	// apperrors.FromGORM(err, "medical_record", ...) で先にラップしてから
	// apperrors.IsNotFound(wrapped) を判定するため、「該当なし」「他院のみ存在」の
	// いずれのケースでも正しく nil, nil を返す。
	t.Run("clinic_id隔離: 別クリニックのカルテは対象外", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "隔離飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA,
			OwnerID:  &ownerID,
			Date:     time.Now().In(time.Local),
		})

		got, err := repo.FindLatestByOwner(ctx, clinicB, ownerID)
		require.NoError(t, err)
		assert.Nil(t, got, "別クリニックからは nil を返す")
	})

	t.Run("該当なし: カルテが存在しない飼い主は nil, nil", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "カルテなし飼主")

		got, err := repo.FindLatestByOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestMedicalRecordRepository_FindOwnerVisitSummary(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("正常系: 初回/最終/合計/年間集計", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "集計飼主")
		ownerID := owner.ID
		firstDate := time.Now().In(time.Local).AddDate(-2, 0, 0)
		lastDate := time.Now().In(time.Local).AddDate(0, -3, 0)

		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: firstDate,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: lastDate,
		})

		got, err := repo.FindOwnerVisitSummary(ctx, clinicA, ownerID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.EqualValues(t, 2, got.TotalCount)
		assert.EqualValues(t, 1, got.AnnualCount, "1年以内の来院は lastDate の1件のみ")
		require.NotNil(t, got.FirstVisitAt)
		require.NotNil(t, got.LastVisitAt)
		assert.True(t, got.FirstVisitAt.Before(*got.LastVisitAt) || got.FirstVisitAt.Equal(*got.LastVisitAt))
	})

	t.Run("clinic_id隔離: 別クリニックのカルテは集計に含まれない", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "隔離集計飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: time.Now().In(time.Local),
		})

		got, err := repo.FindOwnerVisitSummary(ctx, clinicB, ownerID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.EqualValues(t, 0, got.TotalCount)
		assert.Nil(t, got.FirstVisitAt)
		assert.Nil(t, got.LastVisitAt)
	})

	t.Run("該当なし: カルテが存在しない飼い主はゼロ集計", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "集計なし飼主")

		got, err := repo.FindOwnerVisitSummary(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.EqualValues(t, 0, got.TotalCount)
		assert.EqualValues(t, 0, got.AnnualCount)
		assert.Nil(t, got.FirstVisitAt)
		assert.Nil(t, got.LastVisitAt)
	})
}

func TestMedicalRecordRepository_FindOwnersByFirstVisitDate(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	target := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)

	t.Run("正常系: MIN(date) が targetDate と一致する飼い主を返す", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "初回来院飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: target,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: target.AddDate(0, 0, 10),
		})

		got, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.Contains(t, got, ownerID)
	})

	t.Run("clinic_id隔離: 別クリニックの飼い主は混入しない", func(t *testing.T) {
		ownerA := makeTestOwner(t, db, clinicA, "隔離初回A")
		ownerAID := ownerA.ID
		ownerB := makeTestOwner(t, db, clinicB, "隔離初回B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerAID, Date: target,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicB, OwnerID: &ownerBID, Date: target,
		})

		got, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.Contains(t, got, ownerAID)
		assert.NotContains(t, got, ownerBID)
	})

	t.Run("clinic_id隔離: 医院Aのカルテが医院Bの飼い主を誤参照しても混入しない", func(t *testing.T) {
		ownerB := makeTestOwner(t, db, clinicB, "不整合初回B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerBID, Date: target,
		})

		got, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerBID)
	})

	t.Run("該当なし: targetDate に一致するカルテが無ければ空配列", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "初回不一致飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: target.AddDate(0, 0, 1),
		})

		got, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerID)
	})
}

func TestMedicalRecordRepository_FindOwnersByLastVisitDays(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const exactDays = 30
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	lastVisitDate := asOf.AddDate(0, 0, -exactDays)

	t.Run("正常系: MAX(date) が asOf-exactDays と一致する飼い主を返す", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "最終来院飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: lastVisitDate.AddDate(0, 0, -20),
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: lastVisitDate,
		})

		got, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, asOf)
		require.NoError(t, err)
		assert.Contains(t, got, ownerID)
	})

	t.Run("clinic_id隔離: 別クリニックの飼い主は混入しない", func(t *testing.T) {
		ownerA := makeTestOwner(t, db, clinicA, "隔離最終A")
		ownerAID := ownerA.ID
		ownerB := makeTestOwner(t, db, clinicB, "隔離最終B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerAID, Date: lastVisitDate,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicB, OwnerID: &ownerBID, Date: lastVisitDate,
		})

		got, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, asOf)
		require.NoError(t, err)
		assert.Contains(t, got, ownerAID)
		assert.NotContains(t, got, ownerBID)
	})

	t.Run("clinic_id隔離: 医院Aのカルテが医院Bの飼い主を誤参照しても混入しない", func(t *testing.T) {
		ownerB := makeTestOwner(t, db, clinicB, "不整合最終B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerBID, Date: lastVisitDate,
		})

		got, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, asOf)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerBID)
	})

	t.Run("該当なし: exactDays に一致するカルテが無ければ空配列", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "最終不一致飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: lastVisitDate.AddDate(0, 0, -5),
		})

		got, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, asOf)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerID)
	})
}

// TestMedicalRecordRepository_FindOwnersByNextVisitRecommended は Raw SQL による
// clinic_id 二重指定（横テナント漏洩防止・削除禁止コメント対象）が実際に機能することを検証する。
func TestMedicalRecordRepository_FindOwnersByNextVisitRecommended(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	target := time.Date(2026, 8, 20, 0, 0, 0, 0, time.Local)

	t.Run("正常系: 最新カルテの次回来院推奨日が targetDate の飼い主を返す", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "推奨日飼主")
		ownerID := owner.ID
		otherDate := target.AddDate(0, 0, 5)
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: time.Now().In(time.Local).AddDate(0, 0, -10),
			NextVisitRecommendedDate: &otherDate,
		})
		// 最新カルテ（後から作成 = 大きい id）に target を設定
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: time.Now().In(time.Local),
			NextVisitRecommendedDate: &target,
		})

		got, err := repo.FindOwnersByNextVisitRecommended(ctx, clinicA, target)
		require.NoError(t, err)
		assert.Contains(t, got, ownerID)
	})

	t.Run("clinic_id隔離(二重指定WHERE): 同一targetDateでも別クリニックの飼い主は混入しない", func(t *testing.T) {
		ownerA := makeTestOwner(t, db, clinicA, "推奨日隔離A")
		ownerAID := ownerA.ID
		ownerB := makeTestOwner(t, db, clinicB, "推奨日隔離B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerAID, Date: time.Now().In(time.Local),
			NextVisitRecommendedDate: &target,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicB, OwnerID: &ownerBID, Date: time.Now().In(time.Local),
			NextVisitRecommendedDate: &target,
		})

		gotA, err := repo.FindOwnersByNextVisitRecommended(ctx, clinicA, target)
		require.NoError(t, err)
		assert.Contains(t, gotA, ownerAID)
		assert.NotContains(t, gotA, ownerBID, "clinic_id 二重WHERE が機能し他医院の飼い主が混入しない")

		gotB, err := repo.FindOwnersByNextVisitRecommended(ctx, clinicB, target)
		require.NoError(t, err)
		assert.Contains(t, gotB, ownerBID)
		assert.NotContains(t, gotB, ownerAID)
	})

	t.Run("clinic_id隔離: 医院Aのカルテが医院Bの飼い主を誤参照しても混入しない", func(t *testing.T) {
		ownerB := makeTestOwner(t, db, clinicB, "不整合推奨日B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerBID, Date: time.Now().In(time.Local),
			NextVisitRecommendedDate: &target,
		})

		got, err := repo.FindOwnersByNextVisitRecommended(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerBID)
	})

	t.Run("該当なし: targetDate に一致する最新カルテが無ければ空配列", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "推奨日不一致飼主")
		ownerID := owner.ID
		otherDate := target.AddDate(0, 0, 1)
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: time.Now().In(time.Local),
			NextVisitRecommendedDate: &otherDate,
		})

		got, err := repo.FindOwnersByNextVisitRecommended(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, got, ownerID)
	})
}

func TestMedicalRecordRepository_FindDormantOwnerEntries(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const minDaysSince = 180
	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)

	t.Run("正常系: minDaysSince 以上経過した飼い主を返す", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "休眠飼主単体")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: oldDate,
		})

		got, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
		require.NoError(t, err)

		var found *DormantOwnerEntry
		for i := range got {
			if got[i].OwnerID == ownerID {
				found = &got[i]
			}
		}
		require.NotNil(t, found, "休眠飼い主が結果に含まれる")
		assert.GreaterOrEqual(t, found.DaysSince, minDaysSince)
	})

	t.Run("clinic_id隔離: 別クリニックの休眠飼い主は混入しない", func(t *testing.T) {
		ownerA := makeTestOwner(t, db, clinicA, "休眠隔離A")
		ownerAID := ownerA.ID
		ownerB := makeTestOwner(t, db, clinicB, "休眠隔離B")
		ownerBID := ownerB.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerAID, Date: oldDate,
		})
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicB, OwnerID: &ownerBID, Date: oldDate,
		})

		got, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
		require.NoError(t, err)

		ids := make([]uint64, len(got))
		for i, e := range got {
			ids[i] = e.OwnerID
		}
		assert.Contains(t, ids, ownerAID)
		assert.NotContains(t, ids, ownerBID)
	})

	t.Run("該当なし: 直近来院のみの飼い主は含まれない", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "直近来院飼主")
		ownerID := owner.ID
		makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA, OwnerID: &ownerID, Date: time.Now().In(time.Local),
		})

		got, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
		require.NoError(t, err)

		ids := make([]uint64, len(got))
		for i, e := range got {
			ids[i] = e.OwnerID
		}
		assert.NotContains(t, ids, ownerID)
	})
}
