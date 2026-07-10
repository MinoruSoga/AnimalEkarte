package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// makeMedicalRecordForPet はテスト用カルテ 1 件を作成する。deleted=true なら論理削除する。
// pet_id は medical_records → pets の FK 制約があるため、呼び出し側で実在ペットの ID を渡すこと。
// 既存ヘルパー（makeOwner / makeSpeciesAndPet）に合わせ context は内部生成する。
func makeMedicalRecordForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64, date time.Time, deleted bool) {
	t.Helper()
	ctx := context.Background()
	pid := petID
	mr := &model.MedicalRecord{ClinicID: clinicID, PetID: &pid, Date: date}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	if deleted {
		// gorm.DeletedAt により deleted_at がセットされる（論理削除）。
		require.NoError(t, db.WithContext(ctx).Delete(mr).Error)
	}
}

// TestMedicalRecordRepository_FindFirstVisitDateByPetID は #158 飼主レポートの初診日導出を検証する。
// 期待: 同一ペット・同一医院の有効カルテのうち最古の date を返し、
// 論理削除カルテ・別医院カルテ・別ペットのカルテは集計から除外する。捏造はしない（無ければ nil）。
func TestMedicalRecordRepository_FindFirstVisitDateByPetID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}))
	// 共有テスト DB の他テスト残骸と混ざらないよう、FK 連鎖ごと初期化する。
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species CASCADE")

	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	d := func(s string) time.Time {
		tm, err := time.Parse("2006-01-02", s)
		require.NoError(t, err)
		return tm
	}

	owner := makeOwner(t, db, clinicA, "山田太郎")
	petA := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")
	// otherPet の所属医院（clinicB）はテスト対象外。pet スコープ除外の検証では
	// otherPet.ID を使った clinicA のカルテを作り、petA への問い合わせから除外されることを見る。
	otherPet := makeSpeciesAndPet(t, db, clinicB, owner.ID, "タマ")
	deletedOnlyPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ミケ")

	// clinicA / petA: 有効カルテの最古は 2023-03-01。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2024-01-15"), false)
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2023-03-01"), false) // 最古の有効カルテ
	// より古いが論理削除済み → 除外される。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2022-01-01"), true)
	// 同じ pet だが別医院 (clinicB) のより古いカルテ → clinic スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicB, petA.ID, d("2020-01-01"), false)
	// clinicA の別ペットのより古いカルテ → pet スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicA, otherPet.ID, d("2019-01-01"), false)
	// 全カルテ論理削除のペット。
	makeMedicalRecordForPet(t, db, clinicA, deletedOnlyPet.ID, d("2021-05-05"), true)

	t.Run("最古の有効カルテ date を返し、論理削除・別医院・別ペットを除外する", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2023-03-01", got.Format("2006-01-02"))
	})

	t.Run("カルテが無いペットは nil を返す（捏造しない）", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, uint64(99999))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("全カルテが論理削除されたペットは nil を返す", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, deletedOnlyPet.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

// ---- FindAll（B-1: server-side pagination / search / filter 拡張） ----

// setupMedicalRecordListTestDB は FindAll のテストに必要な関連テーブル（pets/animal_species/staffs/inquiries）
// を整備した上で、B-1 で追加した search/filter 用の JOIN 対象データが他テストと混ざらないよう TRUNCATE する。
func setupMedicalRecordListTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}, &model.Staff{}, &model.Inquiry{}))
	db.Exec("TRUNCATE TABLE inquiries, medical_records, pets, animal_species, staffs CASCADE")
	return db
}

// makeFullMedicalRecord は search/filter テスト用にフィールドを一通り指定してカルテを1件作成する。
func makeFullMedicalRecord(t *testing.T, db *gorm.DB, mr *model.MedicalRecord) *model.MedicalRecord {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

// makeInquiryForRecord は指定カルテに主訴（chief_complaint）付きの問診を1件作成する。
func makeInquiryForRecord(t *testing.T, db *gorm.DB, medicalRecordID uint64, chiefComplaint string) {
	t.Helper()
	inq := &model.Inquiry{MedicalRecordID: medicalRecordID, ChiefComplaint: chiefComplaint}
	require.NoError(t, db.WithContext(context.Background()).Create(inq).Error)
}

// TestMedicalRecordRepository_FindAll_ClinicIsolation は clinic_id 隔離が
// search/filter/pagination の全経路で維持されることを検証する（B-1）。
func TestMedicalRecordRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeOwner(t, db, clinicA, "医院A飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "医院Aペット")
	ownerB := makeOwner(t, db, clinicB, "医院B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bペット")

	recA := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "A-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicB, RecordNo: "B-001", Date: time.Now(), OwnerID: &ownerB.ID, PetID: &petB.ID})

	t.Run("指定した clinicIDs 以外のカルテは含まれない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, recA.ID, got[0].ID)
	})

	t.Run("clinicIDs が空はフェイルセーフで空配列を返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})

	t.Run("検索語が他院データにマッチしても混入しない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "医院B"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

// TestMedicalRecordRepository_FindAll_Search は search が飼主名・ペット名・record_no・主訴を
// 部分一致で横断検索できることを検証する（B-1 AC-2）。
func TestMedicalRecordRepository_FindAll_Search(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerTarget := makeOwner(t, db, clinicA, "検索対象飼主タロウ")
	ownerOther := makeOwner(t, db, clinicA, "別の飼主ハナコ")
	// makeOwner/makeSpeciesAndPet は NameKana を設定しないため、カタカナを含む語で検索すると
	// NormalizeKana によりひらがな化されたパターンが生カタカナ列と文字種不一致でマッチしない
	// （owner_repository_test.go の同種コメント参照）。カナ正規化の影響を受けない漢字で検証する。
	petTarget := makeSpeciesAndPet(t, db, clinicA, ownerTarget.ID, "検索対象犬")
	petOther := makeSpeciesAndPet(t, db, clinicA, ownerOther.ID, "別のペット")

	byOwnerName := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-OWNER", Date: time.Now(), OwnerID: &ownerTarget.ID, PetID: &petOther.ID})
	byPetName := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-PET", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petTarget.ID})
	byRecordNo := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-検索対象-999", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})
	byChiefComplaint := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-INQUIRY", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})
	makeInquiryForRecord(t, db, byChiefComplaint.ID, "検索対象の主訴：嘔吐")
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-NOMATCH", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})

	t.Run("飼主名で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象飼主"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byOwnerName.ID, got[0].ID)
	})

	t.Run("ペット名で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象犬"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byPetName.ID, got[0].ID)
	})

	t.Run("record_no で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象-999"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byRecordNo.ID, got[0].ID)
	})

	t.Run("主訴（chief_complaint）で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象の主訴"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byChiefComplaint.ID, got[0].ID)
	})

	t.Run("該当しない検索語は空を返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "存在しない検索語XYZ"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

// TestMedicalRecordRepository_FindAll_Filters は status / doctor_id / animal_species_id /
// start_date・end_date フィルタが独立して機能することを検証する（B-1 AC-2）。
func TestMedicalRecordRepository_FindAll_Filters(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "フィルタ検証飼主")
	dogPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "犬ペット")
	catSpecies := &model.AnimalSpecies{Name: "猫"}
	require.NoError(t, db.WithContext(ctx).Create(catSpecies).Error)
	catPet := &model.Pet{ClinicID: clinicA, OwnerID: owner.ID, AnimalSpeciesID: catSpecies.ID, Name: "猫ペット"}
	require.NoError(t, db.WithContext(ctx).Create(catPet).Error)

	doctorA := makeDoctor(t, db, clinicA, "担当医A")
	doctorB := makeDoctor(t, db, clinicA, "担当医B")

	draft := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "F-DRAFT", Date: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &owner.ID, PetID: &dogPet.ID, DoctorID: &doctorA.ID, Status: model.MedicalRecordStatusDraft,
	})
	finalized := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "F-FINAL", Date: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &owner.ID, PetID: &catPet.ID, DoctorID: &doctorB.ID, Status: model.MedicalRecordStatusFinalized,
	})

	t.Run("status で絞り込める", func(t *testing.T) {
		statusDraft := model.MedicalRecordStatusDraft
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Status: &statusDraft}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, draft.ID, got[0].ID)
	})

	t.Run("doctor_id で絞り込める", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{DoctorID: &doctorB.ID}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("animal_species_id で絞り込める", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{AnimalSpeciesID: &catSpecies.ID}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("start_date/end_date で絞り込める（既存挙動の維持）", func(t *testing.T) {
		start := "2026-02-01"
		end := "2026-02-28"
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{StartDate: &start, EndDate: &end}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("フィルタ未指定時は全件返す", func(t *testing.T) {
		_, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
	})
}

// TestMedicalRecordRepository_FindAll_Sort は B-1 follow-up の列ソート server 化を検証する。
// 許可キー（date/owner_name/pet_name/status）ごとに ORDER BY が反映されること、
// 未許可キーは既定順（date DESC, created_at DESC）にフォールバックすることを確認する。
func TestMedicalRecordRepository_FindAll_Sort(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeOwner(t, db, clinicA, "Aソート飼主")
	ownerB := makeOwner(t, db, clinicA, "Bソート飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Aソートペット")
	petB := makeSpeciesAndPet(t, db, clinicA, ownerB.ID, "Bソートペット")

	// 日付の昇順と record_no の対応が食い違うように意図的に作成順序をずらす。
	recNewer := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "S-NEW", Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerB.ID, PetID: &petB.ID, Status: model.MedicalRecordStatusFinalized,
	})
	recOlder := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "S-OLD", Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID, Status: model.MedicalRecordStatusDraft,
	})

	t.Run("sort=date & order=asc は日付昇順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "date", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recOlder.ID, got[0].ID)
		assert.Equal(t, recNewer.ID, got[1].ID)
	})

	t.Run("sort=date & order=desc は日付降順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "date", Order: "desc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID)
		assert.Equal(t, recOlder.ID, got[1].ID)
	})

	t.Run("sort=owner_name & order=asc は飼主名昇順になる（JOIN経由）", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "owner_name", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recOlder.ID, got[0].ID, "Aソート飼主 が先頭")
		assert.Equal(t, recNewer.ID, got[1].ID, "Bソート飼主 が次")
	})

	t.Run("sort=pet_name & order=desc はペット名降順になる（JOIN経由）", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "pet_name", Order: "desc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID, "Bソートペット が先頭")
		assert.Equal(t, recOlder.ID, got[1].ID, "Aソートペット が次")
	})

	t.Run("sort=status & order=asc はステータス昇順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "status", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, string(model.MedicalRecordStatusDraft), string(got[0].Status))
	})

	t.Run("未許可の sort キーは既定順（date DESC）にフォールバックする", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "unknown_column", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID, "既定順(date DESC)を維持する")
		assert.Equal(t, recOlder.ID, got[1].ID)
	})
}

// TestMedicalRecordRepository_FindAll_Pagination は total > limit のとき
// page1 と page2 が重複しないことを固定する回帰テスト（B-1 旧 failure mode 再発防止 / harness P1）。
func TestMedicalRecordRepository_FindAll_Pagination(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "ページング検証飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ページング検証ペット")

	const totalRecords = 25
	const limit = 20
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ids := make([]uint64, 0, totalRecords)
	for i := 0; i < totalRecords; i++ {
		mr := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA,
			RecordNo: fmt.Sprintf("P-%03d", i),
			// 日付をずらして採番する（BE デフォルトソート: date DESC, created_at DESC）
			Date:    base.AddDate(0, 0, i),
			OwnerID: &owner.ID,
			PetID:   &pet.ID,
		})
		ids = append(ids, mr.ID)
	}

	t.Run("total は limit を超えても正しい件数を返す（旧 failure mode: 常に20件表示の再発防止）", func(t *testing.T) {
		_, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, limit)
		require.NoError(t, err)
		assert.Equal(t, int64(totalRecords), total, "total は limit=20 を超えた実件数(25件)を返すべき")
	})

	t.Run("page1 は limit 件、page2 は残り件数を返し重複しない", func(t *testing.T) {
		page1, total1, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, limit)
		require.NoError(t, err)
		page2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 2, limit)
		require.NoError(t, err)

		assert.Equal(t, int64(totalRecords), total1)
		assert.Equal(t, int64(totalRecords), total2)
		assert.Len(t, page1, limit, "page1 は limit 件（20件）返すべき")
		assert.Len(t, page2, totalRecords-limit, "page2 は残り件数（5件）返すべき")

		seen := map[uint64]bool{}
		for _, r := range page1 {
			seen[r.ID] = true
		}
		for _, r := range page2 {
			assert.False(t, seen[r.ID], "page2 のレコードが page1 と重複してはならない（B-1 旧 failure mode: 常に先頭20件のみ表示の再発防止）")
			seen[r.ID] = true
		}
		for _, id := range ids {
			assert.True(t, seen[id], "作成した全レコード(id=%d)が page1+page2 のいずれかに含まれるべき（欠落防止）", id)
		}
		assert.Len(t, seen, len(ids), "page1+page2 の合計件数は作成件数と一致すべき（重複/欠落なし）")
	})
}

// ---- H-7: CRUD 系メソッドの正常系・clinic_id 隔離カバレッジ拡充 ----
// FindByID/FindByIDForClinics/Create/Update/Delete/CountByPetID/CountByOwnerID は
// FindAll 系テストでは経路が通らず未カバーだったため、正常系と越境防止を中心に追加する。

// TestMedicalRecordRepository_FindByID は正常取得と他院からのアクセス拒否を検証する。
func TestMedicalRecordRepository_FindByID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeOwner(t, db, clinicA, "FindByID飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FindByIDペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "FBI-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
	})

	t.Run("同一医院からは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec.ID, got.ID)
		assert.Equal(t, "FBI-001", got.RecordNo)
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_LockDraftByID は X-11 finalize-child-write-race fix の
// LockDraftByID の基本挙動（clinic_id 隔離・存在確認・status に関わらずロック取得できること）を
// 検証する（並行ロック挙動そのものの実証は medical_record_finalize_lock_concurrency_test.go）。
func TestMedicalRecordRepository_LockDraftByID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeOwner(t, db, clinicA, "LockDraftByID飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "LockDraftByIDペット")
	draftRec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "LDB-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
		Status: model.MedicalRecordStatusDraft,
	})
	finalizedRec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "LDB-002", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
		Status: model.MedicalRecordStatusFinalized,
	})

	t.Run("同一医院からは draft レコードを取得できる", func(t *testing.T) {
		got, err := repo.LockDraftByID(ctx, clinicA, draftRec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, draftRec.ID, got.ID)
		assert.Equal(t, model.MedicalRecordStatusDraft, got.Status)
	})

	t.Run("status に関わらずロック取得できる（finalized も返す。判定は呼び出し元）", func(t *testing.T) {
		got, err := repo.LockDraftByID(ctx, clinicA, finalizedRec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, finalizedRec.ID, got.ID)
		assert.Equal(t, model.MedicalRecordStatusFinalized, got.Status)
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		_, err := repo.LockDraftByID(ctx, clinicB, draftRec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.LockDraftByID(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_FindByIDForClinics はマルチクリニック横断取得の
// clinic_id 隔離（許可リスト外は拒否）を検証する。
func TestMedicalRecordRepository_FindByIDForClinics(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)

	ownerA := makeOwner(t, db, clinicA, "FBIFC飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FBIFCペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "FBIFC-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
	})

	t.Run("許可リストに所属医院が含まれれば取得できる", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, rec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec.ID, got.ID)
	})

	t.Run("許可リストに所属医院が含まれなければ NotFound", func(t *testing.T) {
		_, err := repo.FindByIDForClinics(ctx, []uint64{clinicB, clinicC}, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("許可リストが空はフェイルセーフで NotFound", func(t *testing.T) {
		_, err := repo.FindByIDForClinics(ctx, []uint64{}, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_Create は作成の正常系と record_no 一意制約違反を検証する。
func TestMedicalRecordRepository_Create(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "Create飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Createペット")

	t.Run("正常系: Create 後 FindByID で取得できる", func(t *testing.T) {
		rec := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CR-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID}
		require.NoError(t, repo.Create(ctx, rec))
		require.NotZero(t, rec.ID)

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MedicalRecordStatusDraft, got.Status, "既定ステータスは draft であるべき")
	})
}

// TestMedicalRecordRepository_Update は draft のみ更新可能な業務ルールと
// clinic_id 越境更新の拒否を検証する。
func TestMedicalRecordRepository_Update(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "Update飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Updateペット")

	t.Run("draft カルテは更新できる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-DRAFT", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		updated, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-DRAFT-EDITED"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "UP-DRAFT-EDITED", updated.RecordNo)
	})

	t.Run("finalized カルテは Conflict で更新できない", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "UP-FIN", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			Status: model.MedicalRecordStatusFinalized,
		})
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-FIN-EDITED"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテの更新は Conflict であるべき: %v", err)
	})

	t.Run("他院からの更新は対象0件で Conflict（越境更新の拒否）", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-XCLINIC", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		_, err := repo.Update(ctx, clinicB, rec.ID, map[string]any{"record_no": "HACKED"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))

		// 実データが書き換わっていないことも確認する。
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "UP-XCLINIC", got.RecordNo, "他院からの更新でレコードが変更されてはならない")
	})

	// ─── BE-refactor.md X-10 (mr-version-check-not-atomic): version 述語の原子性 ───

	t.Run("expectedVersion 一致時は更新でき version がインクリメントされる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-VER-OK", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		require.Equal(t, 1, rec.Version, "makeFullMedicalRecord 直後の既定 version は 1 のはず")
		expected := rec.Version
		updated, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-VER-OK-EDITED", "version": rec.Version + 1}, &expected)
		require.NoError(t, err)
		assert.Equal(t, "UP-VER-OK-EDITED", updated.RecordNo)
		assert.Equal(t, rec.Version+1, updated.Version)
	})

	t.Run("expectedVersion 不一致時は「他のユーザーが変更した」Conflict になり書き込まれない", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-VER-STALE", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		staleVersion := rec.Version - 1 // 既に他ユーザーが version を進めた状態を模す
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "SHOULD-NOT-APPLY", "version": rec.Version + 1}, &staleVersion)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "他のユーザーがこのカルテを変更しました", "version不一致は not-draft とは異なる文言で区別されるべき")

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "UP-VER-STALE", got.RecordNo, "version不一致時はフィールドが書き換わってはならない")
	})

	t.Run("finalized カルテへの expectedVersion 付き更新は version不一致メッセージではなく not-draft を返す", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "UP-VER-FIN", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			Status: model.MedicalRecordStatusFinalized,
		})
		expected := rec.Version
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "SHOULD-NOT-APPLY"}, &expected)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "not in draft status", "確定済みは version不一致ではなく従来の not-draft 文言を維持するべき")
	})
}

// TestMedicalRecordRepository_Update_VersionPredicate_ConcurrentUpdates_OnlyOneSucceeds は
// BE-refactor.md X-10 の受け入れ条件: 同一 version=N を起点にした同一カルテへの同時2件の
// Update で、正確に1件のみ成功し他方は「他のユーザーがこのカルテを変更しました」Conflict を
// 受け取り、lost update が発生しないことを実DBの2 goroutineで検証する。
func TestMedicalRecordRepository_Update_VersionPredicate_ConcurrentUpdates_OnlyOneSucceeds(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeOwner(t, db, clinicID, "並行更新飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "並行更新ペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicID, RecordNo: "UP-CONCURRENT", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	startVersion := rec.Version

	const attempts = 2
	notes := []string{"スタッフAによる更新", "スタッフBによる更新"}
	results := make([]error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func(idx int) {
			defer wg.Done()
			expected := startVersion
			_, err := repo.Update(ctx, clinicID, rec.ID, map[string]any{
				"record_no": notes[idx],
				"version":   startVersion + 1,
			}, &expected)
			results[idx] = err
		}(i)
	}
	wg.Wait()

	successCount, conflictCount := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successCount++
		case apperrors.IsConflict(err):
			conflictCount++
		default:
			t.Fatalf("unexpected error from concurrent version-checked update: %v", err)
		}
	}
	assert.Equal(t, 1, successCount, "同一versionを起点にした同時更新はちょうど1件だけ成功するべき（lost update防止）")
	assert.Equal(t, 1, conflictCount, "もう1件は version 不一致 Conflict になるべき")

	got, err := repo.FindByID(ctx, clinicID, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, startVersion+1, got.Version, "version はちょうど1回だけインクリメントされるべき")
	assert.Contains(t, notes, got.RecordNo, "書き込まれた内容は勝者側のnotesのいずれかであるべき")
}

// TestMedicalRecordRepository_Delete は論理削除の正常系と clinic_id 越境削除の拒否を検証する。
func TestMedicalRecordRepository_Delete(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "Delete飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Deleteペット")

	t.Run("同一医院からの削除は成功し以降 FindByID で NotFound になる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "DEL-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		require.NoError(t, repo.Delete(ctx, clinicA, rec.ID))

		_, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院からの削除は NotFound（越境削除の拒否）", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "DEL-XCLINIC", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		err := repo.Delete(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		// 実データが削除されていないことも確認する。
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "DEL-XCLINIC", got.RecordNo)
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_CountByPetID_CountByOwnerID は集計系メソッドの
// clinic_id 隔離・論理削除除外を検証する。
func TestMedicalRecordRepository_CountByPetID_CountByOwnerID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "Count飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Countペット")
	otherOwner := makeOwner(t, db, clinicB, "他院Count飼主")
	otherPet := makeSpeciesAndPet(t, db, clinicB, otherOwner.ID, "他院Countペット")

	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CNT-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CNT-002", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	makeMedicalRecordForPet(t, db, clinicA, pet.ID, time.Now(), true) // 論理削除は除外
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicB, RecordNo: "CNT-OTHER", Date: time.Now(), OwnerID: &otherOwner.ID, PetID: &otherPet.ID})

	t.Run("CountByPetID は同一医院・有効カルテのみ数える", func(t *testing.T) {
		count, err := repo.CountByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CountByPetID は他院からは0を返す（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountByPetID(ctx, clinicB, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountByOwnerID は同一医院・有効カルテのみ数える", func(t *testing.T) {
		count, err := repo.CountByOwnerID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CountByOwnerID は他院からは0を返す（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountByOwnerID(ctx, clinicB, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// TestMedicalRecordRepository_DeleteDraftByAppointmentID は予約紐づけ draft カルテの
// 自動削除ロジック（#83 Q10）を検証する。draft 以外・存在しない場合はエラーにしない。
func TestMedicalRecordRepository_DeleteDraftByAppointmentID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "DDBA飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "DDBAペット")
	appointmentID := uint64(4242)

	t.Run("draft カルテは削除される", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DDBA-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			AppointmentID: &appointmentID,
		})
		require.NoError(t, repo.DeleteDraftByAppointmentID(ctx, clinicA, appointmentID))

		_, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("finalized カルテは削除されない", func(t *testing.T) {
		finalizedAppointmentID := uint64(4243)
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DDBA-002", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			AppointmentID: &finalizedAppointmentID, Status: model.MedicalRecordStatusFinalized,
		})
		require.NoError(t, repo.DeleteDraftByAppointmentID(ctx, clinicA, finalizedAppointmentID))

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, rec.ID, got.ID, "finalized カルテは削除されず残るべき")
	})

	t.Run("対象なしはエラーにしない", func(t *testing.T) {
		require.NoError(t, repo.DeleteDraftByAppointmentID(ctx, clinicA, uint64(999999)))
	})
}
