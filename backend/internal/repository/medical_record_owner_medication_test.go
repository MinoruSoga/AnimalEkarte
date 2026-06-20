package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupOwnerMedicationTestDB は #158 飼主投薬歴の横断クエリ用に
// owners / pets / medical_records / treatments / medicines / staffs を整備する。
func setupOwnerMedicationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.AnimalSpecies{}, &model.Pet{}, &model.Treatment{}, &model.Medicine{}, &model.Staff{},
	))
	// 共有テスト DB の他テスト残骸と混ざらないよう FK 連鎖ごと初期化する。
	db.Exec("TRUNCATE TABLE treatments, medical_records, pets, medicines, staffs, owners, animal_species CASCADE")
	return db
}

func makeMedicineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func makeOwnerMedRecord(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, doctorID *uint64, date time.Time, deleted bool) *model.MedicalRecord {
	t.Helper()
	oid, pid := ownerID, petID
	mr := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &oid, PetID: &pid, DoctorID: doctorID, Date: date}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	if deleted {
		require.NoError(t, db.WithContext(context.Background()).Delete(mr).Error)
	}
	return mr
}

func makeMedTreatment(t *testing.T, db *gorm.DB, mrID uint64, medicineID *uint64, content, adminRoute string, qty float64, deleted bool) {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: mrID,
		ItemType:        model.TreatmentItemTypeMedicine,
		MedicineID:      medicineID,
		Content:         content,
		AdminRoute:      adminRoute,
		Quantity:        qty,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
	if deleted {
		require.NoError(t, db.WithContext(context.Background()).Delete(tr).Error)
	}
}

// TestMedicalRecordRepository_FindOwnerMedicationHistory は #158 飼主横断投薬歴を検証する。
// 期待:
//   - 同一飼い主の全ペット横断で treatments(item_type=medicine) を date 降順で返す
//   - 別医院のカルテに紐づく投薬は medical_records.clinic_id 隔離で混入しない（同一 owner_id でも）
//   - 非 medicine treatment / 論理削除 treatment / 論理削除カルテは除外する
//   - 薬剤名は medicines.name 優先・空なら treatments.content にフォールバック
func TestMedicalRecordRepository_FindOwnerMedicationHistory(t *testing.T) {
	db := setupOwnerMedicationTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const clinic1 = uint64(1)
	const clinic2 = uint64(2)

	ownerA := makeOwner(t, db, clinic1, "横断テスト飼主A")
	petP1 := makeSpeciesAndPet(t, db, clinic1, ownerA.ID, "ポチ")
	petP2 := makeSpeciesAndPet(t, db, clinic1, ownerA.ID, "タマ")
	doctor := makeDoctor(t, db, clinic1, "山田先生")
	amoxi := makeMedicineMaster(t, db, clinic1, "アモキシシリン")

	// clinic1 / ownerA / petP1 — 薬剤マスタ参照 + 担当医あり（最古）
	mr1 := makeOwnerMedRecord(t, db, clinic1, ownerA.ID, petP1.ID, &doctor.ID, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), false)
	makeMedTreatment(t, db, mr1.ID, &amoxi.ID, "", "経口", 2, false)
	// 非 medicine（procedure）は集計対象外
	require.NoError(t, db.Create(&model.Treatment{MedicalRecordID: mr1.ID, ItemType: model.TreatmentItemTypeProcedure, Content: "処置X"}).Error)

	// clinic1 / ownerA / petP2 — マスタ無し → content フォールバック（最新）
	mr2 := makeOwnerMedRecord(t, db, clinic1, ownerA.ID, petP2.ID, nil, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), false)
	makeMedTreatment(t, db, mr2.ID, nil, "院内調剤薬", "皮下", 1, false)

	// 論理削除された treatment（除外されるべき）
	mr3 := makeOwnerMedRecord(t, db, clinic1, ownerA.ID, petP1.ID, nil, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC), false)
	makeMedTreatment(t, db, mr3.ID, nil, "削除済み投薬", "経口", 1, true)

	// 論理削除されたカルテに紐づく投薬（除外されるべき）
	mr4 := makeOwnerMedRecord(t, db, clinic1, ownerA.ID, petP1.ID, nil, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), true)
	makeMedTreatment(t, db, mr4.ID, nil, "削除カルテの投薬", "経口", 1, false)

	// clinic2 に「同一 owner_id 値」のカルテを作る（クロステナント漏洩検証）。
	// clinic1 で集計したとき混入してはならない。
	petP3 := makeSpeciesAndPet(t, db, clinic2, ownerA.ID, "別院ペット")
	mrLeak := makeOwnerMedRecord(t, db, clinic2, ownerA.ID, petP3.ID, nil, time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC), false)
	makeMedTreatment(t, db, mrLeak.ID, nil, "別院の投薬", "経口", 5, false)

	t.Run("cross-pet aggregation with clinic isolation", func(t *testing.T) {
		rows, total, err := repo.FindOwnerMedicationHistory(ctx, clinic1, ownerA.ID, 1, 10)
		require.NoError(t, err)
		// medicine 2件のみ（procedure / 削除 treatment / 削除カルテ / clinic2 を全除外）
		assert.Equal(t, int64(2), total)
		require.Len(t, rows, 2)

		// date 降順: mr2(6/15) → mr1(6/10)
		assert.Equal(t, "タマ", rows[0].PetName)
		assert.Equal(t, "院内調剤薬", rows[0].MedicineName) // content フォールバック
		assert.Equal(t, "皮下", rows[0].AdminRoute)
		assert.Equal(t, float64(1), rows[0].Quantity)
		assert.Equal(t, petP2.ID, rows[0].PetID)
		assert.Empty(t, rows[0].DoctorName)

		assert.Equal(t, "ポチ", rows[1].PetName)
		assert.Equal(t, "アモキシシリン", rows[1].MedicineName) // medicines.name 優先
		assert.Equal(t, "経口", rows[1].AdminRoute)
		assert.Equal(t, float64(2), rows[1].Quantity)
		assert.Equal(t, "山田先生", rows[1].DoctorName)
	})

	t.Run("clinic2 with same owner_id is independently scoped", func(t *testing.T) {
		rows, total, err := repo.FindOwnerMedicationHistory(ctx, clinic2, ownerA.ID, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, rows, 1)
		assert.Equal(t, "別院の投薬", rows[0].MedicineName)
	})

	t.Run("pagination keeps total but slices rows", func(t *testing.T) {
		rows, total, err := repo.FindOwnerMedicationHistory(ctx, clinic1, ownerA.ID, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, rows, 1)
		assert.Equal(t, "ポチ", rows[0].PetName) // 2ページ目 = 古い方
	})

	t.Run("owner with no medications returns empty without error", func(t *testing.T) {
		ownerEmpty := makeOwner(t, db, clinic1, "投薬なし飼主")
		rows, total, err := repo.FindOwnerMedicationHistory(ctx, clinic1, ownerEmpty.ID, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, rows)
	})
}
