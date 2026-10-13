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

// seedDormantOwners は clinicID に休眠飼主（最終来院が oldDate）を count 件一括作成し、
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
// 別医院の休眠飼主が混入しないことを検証する（clinicScope の回帰防止）。
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
	require.Len(t, got, 1, "自医院の休眠飼主のみ返る")
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
// 以下、FindOwnerVisitSummary / FindOwnersByFirstVisitDate /
// FindOwnersByLastVisitDays / FindDormantOwnerEntries のテスト（#212 カバレッジ向上）。
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

	t.Run("clinic_id隔離: 別医院のカルテは集計に含まれない", func(t *testing.T) {
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

	t.Run("該当なし: カルテが存在しない飼主はゼロ集計", func(t *testing.T) {
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

	t.Run("正常系: MIN(date) が targetDate と一致する飼主を返す", func(t *testing.T) {
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

	t.Run("clinic_id隔離: 別医院の飼主は混入しない", func(t *testing.T) {
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

	t.Run("clinic_id隔離: 医院Aのカルテが医院Bの飼主を誤参照しても混入しない", func(t *testing.T) {
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

	t.Run("正常系: MAX(date) が asOf-exactDays と一致する飼主を返す", func(t *testing.T) {
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

	t.Run("clinic_id隔離: 別医院の飼主は混入しない", func(t *testing.T) {
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

	t.Run("clinic_id隔離: 医院Aのカルテが医院Bの飼主を誤参照しても混入しない", func(t *testing.T) {
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

func TestMedicalRecordRepository_FindDormantOwnerEntries(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const minDaysSince = 180
	oldDate := time.Now().In(time.Local).AddDate(0, 0, -200)

	t.Run("正常系: minDaysSince 以上経過した飼主を返す", func(t *testing.T) {
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
		require.NotNil(t, found, "休眠飼主が結果に含まれる")
		assert.GreaterOrEqual(t, found.DaysSince, minDaysSince)
	})

	t.Run("clinic_id隔離: 別医院の休眠飼主は混入しない", func(t *testing.T) {
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

	t.Run("該当なし: 直近来院のみの飼主は含まれない", func(t *testing.T) {
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

// ---------------------------------------------------------------------------
// EMR-330: 通知候補抽出は alive 由来カルテのみを対象とする回帰テスト。
// alive 契約: pet.status <> 'deceased' AND pet.deceased_at IS NULL
// （EMR-290/EMR-328 の OR 契約。status=deceased 単独、deceased_at 非NULL 単独の
// 半死亡不整合行もどちらも死亡扱い）。死亡ペット由来カルテは初診・最終来院・
// 休眠の通知候補から除外する一方、FindOwnerVisitSummary の履歴集計は
// 死亡行を保持しなければならない（LTV/売上/来院履歴の事実は変えない）。
// 現行実装は生死を見ないため本テストは RED になる。
// ---------------------------------------------------------------------------

// ownerVisitAlivePetSpec は 1 ペット分の生死 sentinel と来院日リストを表す入力。
type ownerVisitAlivePetSpec struct {
	status     model.PetStatus
	deceasedAt *time.Time
	visits     []time.Time
}

// makeVisitTestPet は指定した生死 sentinel のペットを1件作成する。
// status='alive' かつ deceased_at 非NULL の半死亡不整合行も表現できる。
func makeVisitTestPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, name string, status model.PetStatus, deceasedAt *time.Time) *model.Pet {
	t.Helper()
	pet := makeSpeciesAndPet(t, db, clinicID, ownerID, name)
	require.NoError(t, db.Model(pet).Updates(map[string]any{
		"status":      status,
		"deceased_at": deceasedAt,
	}).Error)
	return pet
}

func TestMedicalRecordRepository_NotificationCandidates_ExcludeDeceasedPetVisits(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	repo := NewMedicalRecordRepository(db)
	var dormantAtRepo DormantOwnerEntriesAtRepository = repo
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const minDaysSince = 180
	const exactDays = 30

	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	deadAt := now.Add(-48 * time.Hour)

	// 通知対象日と休眠閾値はすべて today 相対で組み立てる（固定カレンダー日は
	// 実行日の経過で休眠判定がずれるため）。
	targetDate := today.AddDate(0, 0, -exactDays) // 初診・最終来院の共通対象日
	dormantVisitDate := today.AddDate(0, 0, -(minDaysSince + 20))
	recentVisitDate := today.AddDate(0, 0, -10)

	alive := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusAlive, visits: dates}
	}
	deadBoth := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusDeceased, deceasedAt: &deadAt, visits: dates}
	}
	deadStatusOnly := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusDeceased, visits: dates}
	}
	deadTimestampOnly := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusAlive, deceasedAt: &deadAt, visits: dates}
	}

	tests := []struct {
		name           string
		clinicID       uint64
		pets           []ownerVisitAlivePetSpec
		wantFirstVisit bool // FindOwnersByFirstVisitDate(clinicA, targetDate) に含まれるべきか
		wantLastVisit  bool // FindOwnersByLastVisitDays(clinicA, exactDays, today) に含まれるべきか
		wantDormant    bool // 休眠 finder 系（非カーソル/カーソル/評価時刻指定）に含まれるべきか
		// alsoFirstVisitAt は生存ペット由来の初診日で追加 membership を確認する日。
		alsoFirstVisitAt []time.Time
	}{
		{
			name:     "全死亡(両sentinel): 対象日の来院があっても初診・最終来院・休眠すべて除外",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadBoth(targetDate)},
		},
		{
			name:     "全死亡(両sentinel): 休眠対象の古い来院だけでは休眠候補にならない",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadBoth(dormantVisitDate)},
		},
		{
			name:     "status のみ deceased (deceased_at IS NULL) の半死亡行も除外",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadStatusOnly(targetDate)},
		},
		{
			name:     "status のみ deceased: 休眠対象の古い来院も除外",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadStatusOnly(dormantVisitDate)},
		},
		{
			name:     "deceased_at のみ非NULL (status=alive) の半死亡行も除外",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadTimestampOnly(targetDate)},
		},
		{
			name:     "deceased_at のみ非NULL: 休眠対象の古い来院も除外",
			clinicID: clinicA,
			pets:     []ownerVisitAlivePetSpec{deadTimestampOnly(dormantVisitDate)},
		},
		{
			name:     "混在: 死亡ペットの対象日来院は通知候補を作らず、生存ペット由来日は候補になる",
			clinicID: clinicA,
			pets: []ownerVisitAlivePetSpec{
				deadBoth(targetDate),
				alive(targetDate.AddDate(0, 0, 10)),
			},
			wantFirstVisit:   false, // 生存由来 MIN = target+10
			wantLastVisit:    false,
			wantDormant:      false,
			alsoFirstVisitAt: []time.Time{targetDate.AddDate(0, 0, 10)},
		},
		{
			name:     "混在: 死亡ペットの直近来院が最終来院日を奪わない（生存由来 MAX=対象日で候補）",
			clinicID: clinicA,
			pets: []ownerVisitAlivePetSpec{
				deadBoth(recentVisitDate),
				alive(targetDate),
			},
			wantFirstVisit: true, // 生存由来 MIN = targetDate
			wantLastVisit:  true, // 生存由来 MAX = targetDate
			wantDormant:    false,
		},
		{
			name:     "混在: 死亡ペットの早期来院が初診日を奪わない（生存由来 MIN=対象日で候補）",
			clinicID: clinicA,
			pets: []ownerVisitAlivePetSpec{
				deadBoth(targetDate.AddDate(0, 0, -30), recentVisitDate),
				alive(targetDate),
			},
			wantFirstVisit: true, // 全履歴 MIN=-60d でも生存由来 MIN=targetDate
			wantLastVisit:  true, // 全履歴 MAX=-10d でも生存由来 MAX=targetDate
			wantDormant:    false,
		},
		{
			name:     "混在: 死亡ペットの直近来院で休眠判定が隠されない（生存由来 MAX が古ければ休眠候補）",
			clinicID: clinicA,
			pets: []ownerVisitAlivePetSpec{
				deadBoth(today.AddDate(0, 0, -5)),
				alive(dormantVisitDate),
			},
			wantDormant: true,
		},
		{
			name:     "別医院: clinicB の生死混在データは clinicA の結果に混入しない",
			clinicID: clinicB,
			pets: []ownerVisitAlivePetSpec{
				deadBoth(targetDate.AddDate(0, 0, -10)),
				alive(dormantVisitDate),
			},
			wantFirstVisit: false,
			wantLastVisit:  false,
			wantDormant:    false,
		},
		{
			name:           "健全性: 生存ペットのみの飼主は従来通り候補になる",
			clinicID:       clinicA,
			pets:           []ownerVisitAlivePetSpec{alive(targetDate)},
			wantFirstVisit: true,
			wantLastVisit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := makeTestOwner(t, db, tt.clinicID, "EMR330-"+tt.name)
			var wantSummaryCount int64
			var wantSummaryFirst time.Time
			for i, spec := range tt.pets {
				pet := makeVisitTestPet(t, db, tt.clinicID, owner.ID, fmt.Sprintf("emr330-pet-%d", i), spec.status, spec.deceasedAt)
				for _, d := range spec.visits {
					ownerID := owner.ID
					petID := pet.ID
					makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
						ClinicID: tt.clinicID,
						OwnerID:  &ownerID,
						PetID:    &petID,
						Date:     d,
					})
					wantSummaryCount++
					if wantSummaryFirst.IsZero() || d.Before(wantSummaryFirst) {
						wantSummaryFirst = d
					}
				}
			}

			firstVisitOwners, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, targetDate)
			require.NoError(t, err)
			if tt.wantFirstVisit {
				assert.Contains(t, firstVisitOwners, owner.ID, "生存ペット由来の初診日が対象日なら通知候補")
			} else {
				assert.NotContains(t, firstVisitOwners, owner.ID, "死亡ペット由来の初診日では通知候補にしない")
			}
			for _, d := range tt.alsoFirstVisitAt {
				got, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, d)
				require.NoError(t, err)
				assert.Contains(t, got, owner.ID, "生存ペット由来の初診日では通知候補になる")
			}

			lastVisitOwners, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, today)
			require.NoError(t, err)
			if tt.wantLastVisit {
				assert.Contains(t, lastVisitOwners, owner.ID, "生存ペット由来の最終来院日が対象日なら通知候補")
			} else {
				assert.NotContains(t, lastVisitOwners, owner.ID, "死亡ペット由来の最終来院日では通知候補にしない")
			}

			dormantIDs := func(entries []DormantOwnerEntry) []uint64 {
				ids := make([]uint64, 0, len(entries))
				for _, e := range entries {
					ids = append(ids, e.OwnerID)
				}
				return ids
			}
			dormant, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
			require.NoError(t, err)
			dormantCursor, err := repo.FindDormantOwnerEntriesCursor(ctx, clinicA, minDaysSince, 0, 500)
			require.NoError(t, err)
			dormantAt, err := dormantAtRepo.FindDormantOwnerEntriesCursorAt(ctx, clinicA, minDaysSince, 0, 500, now)
			require.NoError(t, err)
			for label, ids := range map[string][]uint64{
				"FindDormantOwnerEntries":         dormantIDs(dormant),
				"FindDormantOwnerEntriesCursor":   dormantIDs(dormantCursor),
				"FindDormantOwnerEntriesCursorAt": dormantIDs(dormantAt),
			} {
				if tt.wantDormant {
					assert.Contains(t, ids, owner.ID, "%s: 生存由来の最終来院が古ければ休眠候補", label)
				} else {
					assert.NotContains(t, ids, owner.ID, "%s: 死亡ペット由来の来院では休眠候補にしない", label)
				}
			}

			if tt.clinicID == clinicB {
				// 隔離の健全性: 対象医院側では alive 由来抽出が成立する。
				bDormant, err := repo.FindDormantOwnerEntries(ctx, clinicB, minDaysSince)
				require.NoError(t, err)
				assert.Contains(t, dormantIDs(bDormant), owner.ID, "clinicB 側では生存由来の休眠候補として抽出される")
			}

			// 履歴集計は通知候補フィルタと独立: 死亡ペット由来の行も事実として保持する。
			summary, err := repo.FindOwnerVisitSummary(ctx, tt.clinicID, owner.ID)
			require.NoError(t, err)
			require.NotNil(t, summary)
			assert.EqualValues(t, wantSummaryCount, summary.TotalCount,
				"FindOwnerVisitSummary は死亡ペット由来の来院も履歴として保持する")
			if wantSummaryCount > 0 {
				require.NotNil(t, summary.FirstVisitAt)
				// date 型由来の TZ 丸めに耐性を持たせるため暦日で比較する。
				assert.Equal(t, wantSummaryFirst.Format("2006-01-02"),
					summary.FirstVisitAt.In(time.Local).Format("2006-01-02"),
					"初診日は死亡ペット由来行を含む全履歴の MIN を返す")
			}
		})
	}

	t.Run("論理削除カルテは生存ペット由来でも通知候補にしない", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "EMR330-論理削除飼主")
		pet := makeVisitTestPet(t, db, clinicA, owner.ID, "emr330-softdelete-pet", model.PetStatusAlive, nil)
		ownerID := owner.ID
		rec := makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
			ClinicID: clinicA,
			OwnerID:  &ownerID,
			PetID:    &pet.ID,
			Date:     targetDate,
		})
		require.NoError(t, db.Delete(rec).Error)

		firstVisitOwners, err := repo.FindOwnersByFirstVisitDate(ctx, clinicA, targetDate)
		require.NoError(t, err)
		assert.NotContains(t, firstVisitOwners, owner.ID, "論理削除カルテは初診候補にしない")

		lastVisitOwners, err := repo.FindOwnersByLastVisitDays(ctx, clinicA, exactDays, today)
		require.NoError(t, err)
		assert.NotContains(t, lastVisitOwners, owner.ID, "論理削除カルテは最終来院候補にしない")

		dormant, err := repo.FindDormantOwnerEntries(ctx, clinicA, minDaysSince)
		require.NoError(t, err)
		for _, e := range dormant {
			assert.NotEqual(t, owner.ID, e.OwnerID, "論理削除カルテは休眠候補にしない")
		}
	})
}

// TestMedicalRecordRepository_OwnerHasAliveVisitBoundary は配信直前再読用の
// owner-scoped boundary クエリが、通知候補ファインダーと同じ生存ペット
// predicate・clinic・現飼主境界で評価することを検証する（EMR-330）。
func TestMedicalRecordRepository_OwnerHasAliveVisitBoundary(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	targetDate := today.AddDate(0, 0, -30)
	deadAt := now.Add(-48 * time.Hour)

	seed := func(t *testing.T, clinicID uint64, ownerName string, specs []ownerVisitAlivePetSpec) *model.Owner {
		t.Helper()
		owner := makeTestOwner(t, db, clinicID, ownerName)
		for i, spec := range specs {
			pet := makeVisitTestPet(t, db, clinicID, owner.ID, fmt.Sprintf("boundary-pet-%d", i), spec.status, spec.deceasedAt)
			for _, d := range spec.visits {
				ownerID := owner.ID
				petID := pet.ID
				makeVisitRecordForOwnerVisitTest(t, db, &model.MedicalRecord{
					ClinicID: clinicID,
					OwnerID:  &ownerID,
					PetID:    &petID,
					Date:     d,
				})
			}
		}
		return owner
	}

	aliveSpec := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusAlive, visits: dates}
	}
	deadBoth := func(dates ...time.Time) ownerVisitAlivePetSpec {
		return ownerVisitAlivePetSpec{status: model.PetStatusDeceased, deceasedAt: &deadAt, visits: dates}
	}

	ownerAlive := seed(t, clinicA, "boundary-alive", []ownerVisitAlivePetSpec{aliveSpec(targetDate)})
	ownerDeadOnly := seed(t, clinicA, "boundary-deadonly", []ownerVisitAlivePetSpec{deadBoth(targetDate)})
	ownerStatusOnly := seed(t, clinicA, "boundary-statusonly", []ownerVisitAlivePetSpec{
		{status: model.PetStatusDeceased, visits: []time.Time{targetDate}},
	})
	ownerTimestampOnly := seed(t, clinicA, "boundary-timestamponly", []ownerVisitAlivePetSpec{
		{status: model.PetStatusAlive, deceasedAt: &deadAt, visits: []time.Time{targetDate}},
	})
	ownerMixed := seed(t, clinicA, "boundary-mixed", []ownerVisitAlivePetSpec{
		deadBoth(targetDate),
		aliveSpec(today.AddDate(0, 0, -200)),
	})
	ownerOtherClinic := seed(t, clinicB, "boundary-clinicB", []ownerVisitAlivePetSpec{aliveSpec(targetDate)})

	cases := []struct {
		name      string
		ownerID   uint64
		wantFirst bool
		wantLast  bool
	}{
		{"生存ペット由来の対象日来院は一致", ownerAlive.ID, true, true},
		{"両 sentinel 死亡ペット由来のみなら不一致", ownerDeadOnly.ID, false, false},
		{"status=deceased のみの半死亡由来は不一致", ownerStatusOnly.ID, false, false},
		{"deceased_at のみ非 NULL の半死亡由来は不一致", ownerTimestampOnly.ID, false, false},
		{"死亡ペット由来日+別日の生存由来では対象日不一致", ownerMixed.ID, false, false},
		{"別医院の対象日来院は医院Aでは不一致", ownerOtherClinic.ID, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFirst, err := repo.OwnerHasAliveFirstVisitOnDate(ctx, clinicA, tc.ownerID, targetDate)
			require.NoError(t, err)
			assert.Equal(t, tc.wantFirst, gotFirst)

			gotLast, err := repo.OwnerHasAliveLastVisitOnDate(ctx, clinicA, tc.ownerID, targetDate)
			require.NoError(t, err)
			assert.Equal(t, tc.wantLast, gotLast)
		})
	}

	// 生存兄弟ペットがいるだけではイベント由来行の生存を満たさない。
	t.Run("対象日の生存由来行がなくても他生存ペットでは true にならない", func(t *testing.T) {
		owner := seed(t, clinicA, "boundary-sibling-only", []ownerVisitAlivePetSpec{
			deadBoth(targetDate),
			aliveSpec(today.AddDate(0, 0, -60)),
		})
		got, err := repo.OwnerHasAliveFirstVisitOnDate(ctx, clinicA, owner.ID, targetDate)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestMedicalRecordRepository_DormantHistoryCursor_RetainsDeceasedVisitsAndClinicScope(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewMedicalRecordRepository(db)
	evaluatedAt := time.Now().In(time.Local)
	ids := seedDormantOwners(t, db, 1, 2, evaluatedAt.AddDate(0, 0, -250))
	seedDormantOwners(t, db, 2, 1, evaluatedAt.AddDate(0, 0, -300))
	require.NoError(t, db.Model(&model.Pet{}).Where("clinic_id = ?", 1).Updates(map[string]any{"status": model.PetStatusDeceased, "deceased_at": evaluatedAt}).Error)
	page1, err := repo.FindDormantOwnerHistoryEntriesCursorAt(context.Background(), 1, 180, 0, 1, evaluatedAt)
	require.NoError(t, err)
	require.Len(t, page1, 1)
	assert.Equal(t, ids[0], page1[0].OwnerID)
	assert.Equal(t, 250, page1[0].DaysSince)
	page2, err := repo.FindDormantOwnerHistoryEntriesCursorAt(context.Background(), 1, 180, page1[0].OwnerID, 1, evaluatedAt)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	assert.Equal(t, ids[1], page2[0].OwnerID)
	page3, err := repo.FindDormantOwnerHistoryEntriesCursorAt(context.Background(), 1, 180, page2[0].OwnerID, 1, evaluatedAt)
	require.NoError(t, err)
	assert.Empty(t, page3, "other-clinic history cannot enter this cursor")
	notifications, err := repo.FindDormantOwnerEntriesCursorAt(context.Background(), 1, 180, 0, 10, evaluatedAt)
	require.NoError(t, err)
	assert.Empty(t, notifications, "deceased history is retained for tags, excluded from notifications")
}
