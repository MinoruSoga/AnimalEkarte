package pet

// checkup_history_test.go — EMR-197-01: GET /v1/pets?checkup_history= の統合テスト。
//
// 列挙値契約:
//   within_1y|within_2y|within_3y … live 健診が JST 当日起点 N 年（包含境界）内に存在するペット
//   not_within_1y|not_within_2y|not_within_3y … 同 N 年窓内の live 健診が無いペット（履歴なしを含む）
//   none … live 健診履歴が一切無いペット（未指定＝フィルタ無しとは区別する）
// ペット解決は checkups.pet_id 直接参照と、live かつ clinic 一致 medical_records.pet_id
// 経由の両方を許容する。clinic_id parity・deleted_at 除外・include_deceased 既定・
// 拠点横断スコープとの両立を本ファイルで担保する。

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupPetCheckupHistoryTestDB は checkup_history フィルタテスト用に DB を整備する。
// core truncate（medical_records CASCADE）で checkups も連鎖クリアされるが、
// checkup_types / animal_species はコア集合に含まれないため個別に初期化する。
func setupPetCheckupHistoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.CheckupType{}, &model.MedicalRecord{}, &model.Checkup{},
	))
	db.Exec("TRUNCATE TABLE checkup_types, animal_species CASCADE")
	return db
}

func makeCheckupType(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.CheckupType {
	t.Helper()
	ct := &model.CheckupType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(ct).Error)
	return ct
}

// makeCheckupMedicalRecord は健診行の FK 参照先カルテを作成する。
// petID は nil 可（checkups.pet_id 直接参照経路の検証用）。
func makeCheckupMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, petID *uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Now(),
		PetID:    petID,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func makeCheckup(
	t *testing.T,
	db *gorm.DB,
	clinicID, checkupTypeID, medicalRecordID uint64,
	petID *uint64,
	date time.Time,
) *model.Checkup {
	t.Helper()
	c := &model.Checkup{
		ClinicID:        clinicID,
		MedicalRecordID: medicalRecordID,
		PetID:           petID,
		CheckupTypeID:   checkupTypeID,
		Date:            date,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

// jstToday は repository 実装と同じ「JST での暦日」を返す（窓境界の決定的な期待値計算用）。
func jstToday() time.Time {
	now := time.Now().In(config.JST)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func petIDs(pets []model.Pet) []uint64 {
	ids := make([]uint64, len(pets))
	for i := range pets {
		ids[i] = pets[i].ID
	}
	return ids
}

// TestPetRepositoryFindAllCheckupHistoryWithinYears は within_1y/2y/3y が
// 「JST 当日−N 年」を包含下限とする窓内 live 健診を持つペットだけを返すことを固定する。
func TestPetRepositoryFindAllCheckupHistoryWithinYears(t *testing.T) {
	withJSTLocal(t)
	db := setupPetCheckupHistoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "健診履歴飼主")
	checkupType := makeCheckupType(t, db, clinicA, "一般健診")
	today := jstToday()

	seed := func(name string, checkupDate *time.Time) *model.Pet {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, name)
		if checkupDate != nil {
			mr := makeCheckupMedicalRecord(t, db, clinicA, &pet.ID, fmt.Sprintf("MR-%s", name))
			makeCheckup(t, db, clinicA, checkupType.ID, mr.ID, nil, *checkupDate)
		}
		return pet
	}

	within6mo := seed("健診半年前", ptrTime(today.AddDate(0, -6, 0)))
	boundary1y := seed("健診ちょうど1年前", ptrTime(today.AddDate(-1, 0, 0)))
	within18mo := seed("健診1年半前", ptrTime(today.AddDate(-1, -6, 0)))
	within30mo := seed("健診2年半前", ptrTime(today.AddDate(-2, -6, 0)))
	seed("健診3年と1日前", ptrTime(today.AddDate(-3, 0, -1)))
	seed("健診4年前", ptrTime(today.AddDate(-4, 0, 0)))
	seed("健診なし", nil)

	cases := []struct {
		name    string
		filter  CheckupHistoryFilter
		wantIDs []uint64
	}{
		{
			"within_1y は1年以内＋ちょうど1年前（包含境界）のみ",
			CheckupHistoryWithin1Y,
			[]uint64{within6mo.ID, boundary1y.ID},
		},
		{
			"within_2y は2年以内を含み2年半前は除く",
			CheckupHistoryWithin2Y,
			[]uint64{within6mo.ID, boundary1y.ID, within18mo.ID},
		},
		{
			"within_3y は3年以内を含み3年と1日前/4年前/履歴なしは除く",
			CheckupHistoryWithin3Y,
			[]uint64{within6mo.ID, boundary1y.ID, within18mo.ID, within30mo.ID},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pets, total, err := repo.FindAll(
				ctx,
				[]uint64{clinicA},
				PetListFilters{CheckupHistory: tc.filter},
				1,
				100,
			)
			require.NoError(t, err)
			assert.Equal(t, int64(len(tc.wantIDs)), total, "count query と list query の述語は一致すること")
			assert.ElementsMatch(t, tc.wantIDs, petIDs(pets))
		})
	}
}

// TestPetRepositoryFindAllCheckupHistoryNotWithinAndNone は否定系列挙値の意味を固定する:
// not_within_Ny は「N 年窓内の live 健診が無い」ペット（窓外のみ＋履歴なし）、
// none は「live 健診履歴が一切無い」ペット。いずれも未指定（フィルタ無し）とは別物。
func TestPetRepositoryFindAllCheckupHistoryNotWithinAndNone(t *testing.T) {
	withJSTLocal(t)
	db := setupPetCheckupHistoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "健診否定系飼主")
	checkupType := makeCheckupType(t, db, clinicA, "一般健診")
	today := jstToday()

	seed := func(name string, checkupDate *time.Time) *model.Pet {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, name)
		if checkupDate != nil {
			mr := makeCheckupMedicalRecord(t, db, clinicA, &pet.ID, fmt.Sprintf("MR-%s", name))
			makeCheckup(t, db, clinicA, checkupType.ID, mr.ID, nil, *checkupDate)
		}
		return pet
	}

	// 窓内 fixture（絞り込みから除外されるべき側）は変数化せず、exact-match 断言で存在のみ保証する。
	seed("健診半年前", ptrTime(today.AddDate(0, -6, 0)))
	seed("健診ちょうど1年前", ptrTime(today.AddDate(-1, 0, 0)))
	out1y1d := seed("健診1年と1日前", ptrTime(today.AddDate(-1, 0, -1)))
	within18mo := seed("健診1年半前", ptrTime(today.AddDate(-1, -6, 0)))
	boundary2y := seed("健診ちょうど2年前", ptrTime(today.AddDate(-2, 0, 0)))
	out2y1d := seed("健診2年と1日前", ptrTime(today.AddDate(-2, 0, -1)))
	within30mo := seed("健診2年半前", ptrTime(today.AddDate(-2, -6, 0)))
	boundary3y := seed("健診ちょうど3年前", ptrTime(today.AddDate(-3, 0, 0)))
	out3y1d := seed("健診3年と1日前", ptrTime(today.AddDate(-3, 0, -1)))
	old4y := seed("健診4年前", ptrTime(today.AddDate(-4, 0, 0)))
	never := seed("健診なし", nil)

	cases := []struct {
		name    string
		filter  CheckupHistoryFilter
		wantIDs []uint64
	}{
		{
			"not_within_1y は1年窓外の健診のみ/履歴なしを返す",
			CheckupHistoryNotWithin1Y,
			[]uint64{out1y1d.ID, within18mo.ID, boundary2y.ID, out2y1d.ID, within30mo.ID, boundary3y.ID, out3y1d.ID, old4y.ID, never.ID},
		},
		{
			"not_within_2y は2年窓外の健診のみ/履歴なしを返す",
			CheckupHistoryNotWithin2Y,
			[]uint64{out2y1d.ID, within30mo.ID, boundary3y.ID, out3y1d.ID, old4y.ID, never.ID},
		},
		{
			"not_within_3y は3年窓外の健診のみ/履歴なしを返す",
			CheckupHistoryNotWithin3Y,
			[]uint64{out3y1d.ID, old4y.ID, never.ID},
		},
		{
			"none は live 健診履歴を持たないペットのみを返す",
			CheckupHistoryNone,
			[]uint64{never.ID},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pets, total, err := repo.FindAll(
				ctx,
				[]uint64{clinicA},
				PetListFilters{CheckupHistory: tc.filter},
				1,
				100,
			)
			require.NoError(t, err)
			assert.Equal(t, int64(len(tc.wantIDs)), total)
			assert.ElementsMatch(t, tc.wantIDs, petIDs(pets))
		})
	}
}

// TestPetRepositoryFindAllCheckupHistoryFutureDated はレビュー指摘(MEDIUM-1)の固定:
// 「N 年以内に受診」の窓は [JST 当日−N 年, JST 当日] の両端包含であり、
// 未来日付の健診は within_Ny を満たさない（受診予定は受診履歴ではない）。
// なお未来健診のみを持つペットは not_within_2y には含まれるが、none には含まれない
// （live 健診自体は存在するため「履歴なし」ではない）。
func TestPetRepositoryFindAllCheckupHistoryFutureDated(t *testing.T) {
	withJSTLocal(t)
	db := setupPetCheckupHistoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "未来健診飼主")
	checkupType := makeCheckupType(t, db, clinicA, "一般健診")
	today := jstToday()

	futureOnly := makeSpeciesAndPet(t, db, clinicA, owner.ID, "未来健診のみペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &futureOnly.ID, "MR-FUTURE")
		makeCheckup(t, db, clinicA, checkupType.ID, mr.ID, nil, today.AddDate(0, 1, 0))
	}
	pastInWindow := makeSpeciesAndPet(t, db, clinicA, owner.ID, "窓内健診ペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &pastInWindow.ID, "MR-PAST")
		makeCheckup(t, db, clinicA, checkupType.ID, mr.ID, nil, today.AddDate(0, -6, 0))
	}
	todayPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "当日健診ペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &todayPet.ID, "MR-TODAY")
		makeCheckup(t, db, clinicA, checkupType.ID, mr.ID, nil, today)
	}
	never := makeSpeciesAndPet(t, db, clinicA, owner.ID, "無健診ペット")

	t.Run("未来日付の健診は within_2y を満たさない（当日は包含）", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryWithin2Y},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []uint64{pastInWindow.ID, todayPet.ID}, petIDs(pets))
	})

	t.Run("未来健診のみのペットは not_within_2y に含まれる", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryNotWithin2Y},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []uint64{futureOnly.ID, never.ID}, petIDs(pets))
	})

	t.Run("未来健診のみのペットは none に含まれない（live 健診自体は存在）", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryNone},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.ElementsMatch(t, []uint64{never.ID}, petIDs(pets))
	})
}

// TestPetRepositoryFindAllCheckupHistoryIsolation は隔離不変条件を固定する:
//   - ペット解決は checkups.pet_id 直接参照と live+clinic一致 medical_records.pet_id の双方で有効
//   - 別医院の checkup 行 / clinic 不一致の medical_record 経由では一切絞り込めない
//   - soft-deleted な checkup / medical_record は存在しないものとして扱う
//   - include_deceased 既定（生存のみ）と拠点横断スコープの両立
func TestPetRepositoryFindAllCheckupHistoryIsolation(t *testing.T) {
	withJSTLocal(t)
	db := setupPetCheckupHistoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "健診隔離飼主A")
	ownerB := makeTestOwner(t, db, clinicB, "健診隔離飼主B")
	typeA := makeCheckupType(t, db, clinicA, "健診A")
	typeB := makeCheckupType(t, db, clinicB, "健診B")
	inWindow := jstToday().AddDate(0, -6, 0)

	// 直接参照経路: checkups.pet_id = pets.id（medical_records.pet_id は NULL）
	directPath := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "直接参照ペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, nil, "MR-DIRECT")
		makeCheckup(t, db, clinicA, typeA.ID, mr.ID, &directPath.ID, inWindow)
	}

	// カルテ経由フォールバック: checkups.pet_id NULL、medical_records.pet_id = pets.id
	mrPath := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "カルテ参照ペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &mrPath.ID, "MR-VIA-PET")
		makeCheckup(t, db, clinicA, typeA.ID, mr.ID, nil, inWindow)
	}

	// 別医院の checkup 行が自院ペットを指しても絞り込みには効かない（clinic_id parity）
	crossCheckup := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "他院健診ペット")
	{
		mrB := makeCheckupMedicalRecord(t, db, clinicB, nil, "MR-X1")
		makeCheckup(t, db, clinicB, typeB.ID, mrB.ID, &crossCheckup.ID, inWindow)
	}

	// clinic 不一致の medical_record 経由でも解決しない（m.clinic_id = c.clinic_id の二重条件）
	crossMR := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "他院カルテペット")
	{
		mrB := makeCheckupMedicalRecord(t, db, clinicB, &crossMR.ID, "MR-X2")
		makeCheckup(t, db, clinicA, typeA.ID, mrB.ID, nil, inWindow)
	}

	// soft-deleted 健診は履歴として数えない
	deletedCheckupPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "削除健診ペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &deletedCheckupPet.ID, "MR-DEL-C")
		c := makeCheckup(t, db, clinicA, typeA.ID, mr.ID, nil, inWindow)
		require.NoError(t, db.WithContext(ctx).Delete(c).Error)
	}

	// soft-deleted カルテ経由でも解決しない
	deletedMRPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "削除カルテペット")
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &deletedMRPet.ID, "MR-DEL-M")
		makeCheckup(t, db, clinicA, typeA.ID, mr.ID, nil, inWindow)
		require.NoError(t, db.WithContext(ctx).Delete(mr).Error)
	}

	// 死亡ペット＋live 窓内健診: 既定（include_deceased 未指定）は一覧から除外される
	deceasedPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "死亡健診ペット")
	deceasedAt := jstToday().AddDate(0, -1, 0)
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", deceasedPet.ID).
		Updates(map[string]any{"deceased_at": deceasedAt, "status": model.PetStatusDeceased}).Error)
	{
		mr := makeCheckupMedicalRecord(t, db, clinicA, &deceasedPet.ID, "MR-DEC")
		makeCheckup(t, db, clinicA, typeA.ID, mr.ID, nil, inWindow)
	}

	noCheckup := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "無健診ペット")

	// 別医院のペット＋別医院自身の健診: 拠点横断スコープで各自院の履歴のみ評価される
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "別院健診ペット")
	{
		mrB := makeCheckupMedicalRecord(t, db, clinicB, &petB.ID, "MR-B")
		makeCheckup(t, db, clinicB, typeB.ID, mrB.ID, nil, inWindow)
	}

	t.Run("clinic 一致の直接/カルテ経由のみ within_1y に乗る", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryWithin1Y},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []uint64{directPath.ID, mrPath.ID}, petIDs(pets))
	})

	t.Run("死亡ペットは既定で除外され include_deceased=true で within_1y に戻る", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryWithin1Y, IncludeDeceased: true},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.ElementsMatch(t, []uint64{directPath.ID, mrPath.ID, deceasedPet.ID}, petIDs(pets))
	})

	t.Run("none は live 健診を持たないペットのみ（越境行・削除済みは履歴に数えない）", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			PetListFilters{CheckupHistory: CheckupHistoryNone},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.ElementsMatch(t,
			[]uint64{crossCheckup.ID, crossMR.ID, deletedCheckupPet.ID, deletedMRPet.ID, noCheckup.ID},
			petIDs(pets),
		)
	})

	t.Run("拠点横断スコープでは各院の健診が自院ペットのみに効く", func(t *testing.T) {
		pets, total, err := repo.FindAll(
			ctx,
			[]uint64{clinicA, clinicB},
			PetListFilters{CheckupHistory: CheckupHistoryWithin1Y},
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.ElementsMatch(t, []uint64{directPath.ID, mrPath.ID, petB.ID}, petIDs(pets))
		assert.NotContains(t, petIDs(pets), crossCheckup.ID, "他院の健診行で自院ペットが乗ってはならない")
	})
}

// TestListPetQueryCheckupHistoryValidation は checkup_history の
// request boundary での列挙値検証を固定する（invalid → 400 系、未指定 → フィルタ無し）。
func TestListPetQueryCheckupHistoryValidation(t *testing.T) {
	t.Run("newListPetQuery が checkup_history を読む", func(t *testing.T) {
		q := newListPetQuery(url.Values{"checkup_history": {"within_2y"}})
		assert.Equal(t, "within_2y", q.CheckupHistory)
	})

	t.Run("全 enum 値を受理してサービスフィルタへ写す", func(t *testing.T) {
		for value, want := range map[string]CheckupHistoryFilter{
			"within_1y":     CheckupHistoryWithin1Y,
			"within_2y":     CheckupHistoryWithin2Y,
			"within_3y":     CheckupHistoryWithin3Y,
			"not_within_1y": CheckupHistoryNotWithin1Y,
			"not_within_2y": CheckupHistoryNotWithin2Y,
			"not_within_3y": CheckupHistoryNotWithin3Y,
			"none":          CheckupHistoryNone,
		} {
			filters, err := (&listPetQuery{CheckupHistory: value}).toServiceFilters()
			require.NoError(t, err, value)
			assert.Equal(t, want, filters.CheckupHistory, value)
		}
	})

	t.Run("未指定/空文字はフィルタ無し（述語を張らない）", func(t *testing.T) {
		filters, err := (&listPetQuery{}).toServiceFilters()
		require.NoError(t, err)
		assert.Empty(t, filters.CheckupHistory)
	})

	t.Run("enum 外の値は invalid input（黙って無視しない）", func(t *testing.T) {
		for _, value := range []string{"within_4y", "recent", "WITHIN_1Y", "within_1y ", "none "} {
			filters, err := (&listPetQuery{CheckupHistory: value}).toServiceFilters()
			require.Error(t, err, value)
			assert.Equal(t, listPetFilters{}, filters, value)
			assert.True(t, apperrors.IsInvalidInput(err), "%q は invalid input 系エラーでなければならない", value)
		}
	})
}
