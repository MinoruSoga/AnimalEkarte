package medicalrecord

// repository_test.go — Repository の統合テスト。
//
// 移動元: consultation_repository_test.go（BE8-4 batch25）。makeSpeciesAndPet/
// makeHistoryMedicalRecord はフラット package の同名ヘルパーの複製（BE8-4: import cycle を
// 避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。makeTestOwner は
// フラット側と同様 testdb.MakeTestOwner に直接委譲する。
//
// 保護する不変条件:
//   - FindAll/FindByID/Update/Delete/Reorder は clinic_id で正しく分離される。
//   - FindAll はソフトデリート済みを除外し sort_order/name で並ぶ。
//   - CountUsageByConsultationID は medical_records を JOIN してクリニック分離しつつ
//     ソフトデリート済み treatment を除外する（P2）。
//   - CountChildrenByParentID はソフトデリート済み子・別クリニックの子を除外する（P2）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupConsultationTestDB は consultations と、CountUsageByConsultationID の JOIN 先
// treatments/medical_records/pets/animal_species を整備する。
func setupConsultationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Consultation{}, &model.AnimalSpecies{}, &model.Pet{}, &model.Treatment{}))
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE consultations CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeConsultation(t *testing.T, db *gorm.DB, clinicID uint64, name string, parentID *uint64) *model.Consultation {
	t.Helper()
	c := &model.Consultation{ClinicID: clinicID, Name: name, ParentID: parentID}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func makeConsultationTreatment(t *testing.T, db *gorm.DB, medicalRecordID, consultationID uint64) *model.Treatment {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ItemType:        model.TreatmentItemTypeConsultation,
		ConsultationID:  &consultationID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
	return tr
}

// makeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func TestConsultationRepository_FindAll(t *testing.T) {
	// G2F-11: FindAll applies persistence.MaxMasterListRows (same as vaccine/procedure).

	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := makeConsultation(t, db, clinicA, "内科診察", nil)
	c2 := makeConsultation(t, db, clinicA, "外科診察", nil)
	deleted := makeConsultation(t, db, clinicA, "廃止診察", nil)
	makeConsultation(t, db, clinicB, "別クリニック診察", nil)

	require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "ソフトデリート済み・別クリニックは除外されるべき")
	ids := []uint64{got[0].ID, got[1].ID}
	assert.Contains(t, ids, c1.ID)
	assert.Contains(t, ids, c2.ID)
}

func TestConsultationRepository_FindByID(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := makeConsultation(t, db, clinicA, "問診", nil)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, got.ID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(9999999))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestConsultationRepository_Create(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	price := int64(3000)
	c := &model.Consultation{ClinicID: clinicA, Name: "新規診察", Price: &price}
	require.NoError(t, repo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	var stored model.Consultation
	require.NoError(t, db.First(&stored, c.ID).Error)
	assert.Equal(t, "新規診察", stored.Name)
	require.NotNil(t, stored.Price)
	assert.Equal(t, price, *stored.Price)
	assert.Equal(t, model.TaxTypeExcluded, stored.TaxType, "DBデフォルトが適用されるべき")
}

func TestConsultationRepository_Update(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの更新は成功する", func(t *testing.T) {
		c := makeConsultation(t, db, clinicA, "更新前", nil)
		got, err := repo.Update(ctx, clinicA, c.ID, map[string]any{"name": "更新後"})
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Name)
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, uint64(9999999), map[string]any{"name": "無効"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックIDからの更新はNotFoundを返し値は変わらない", func(t *testing.T) {
		c := makeConsultation(t, db, clinicA, "越境前", nil)
		_, err := repo.Update(ctx, clinicB, c.ID, map[string]any{"name": "越境後"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stored model.Consultation
		require.NoError(t, db.First(&stored, c.ID).Error)
		assert.Equal(t, "越境前", stored.Name)
	})
}

func TestConsultationRepository_Delete(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの削除はソフトデリートされる", func(t *testing.T) {
		c := makeConsultation(t, db, clinicA, "削除対象", nil)
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))

		var normal model.Consultation
		err := db.First(&normal, c.ID).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		var withDeleted model.Consultation
		require.NoError(t, db.Unscoped().First(&withDeleted, c.ID).Error)
		assert.True(t, withDeleted.DeletedAt.Valid)
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, uint64(9999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの削除はNotFoundを返し削除されない", func(t *testing.T) {
		c := makeConsultation(t, db, clinicA, "越境削除対象", nil)
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stillThere model.Consultation
		require.NoError(t, db.First(&stillThere, c.ID).Error)
	})
}

func TestConsultationRepository_Reorder(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := makeConsultation(t, db, clinicA, "並び1", nil)
	c2 := makeConsultation(t, db, clinicA, "並び2", nil)
	c3 := makeConsultation(t, db, clinicA, "並び3", nil)

	t.Run("指定順にsort_orderが振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		var got3, got1, got2 model.Consultation
		require.NoError(t, db.First(&got3, c3.ID).Error)
		require.NoError(t, db.First(&got1, c1.ID).Error)
		require.NoError(t, db.First(&got2, c2.ID).Error)
		assert.Equal(t, 1, got3.SortOrder)
		assert.Equal(t, 2, got1.SortOrder)
		assert.Equal(t, 3, got2.SortOrder)
	})

	t.Run("別クリニックのIDを含む場合はエラーになる", func(t *testing.T) {
		cOther := makeConsultation(t, db, clinicB, "他院", nil)
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, cOther.ID})
		require.Error(t, err, "clinic_id 隔離により対象外IDでエラーになるべき")
	})
}

func TestConsultationRepository_CountChildrenByParentID(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := makeConsultation(t, db, clinicA, "親診察", nil)
	makeConsultation(t, db, clinicA, "子1", &parent.ID)
	makeConsultation(t, db, clinicA, "子2", &parent.ID)
	deletedChild := makeConsultation(t, db, clinicA, "削除済み子", &parent.ID)
	require.NoError(t, repo.Delete(ctx, clinicA, deletedChild.ID))
	// clinicB 側に「parent_id が clinicA の親を指す」行を作成する。CountChildrenByParentID の
	// クエリは常に対象行自身の clinic_id でスコープする（parent_id の値がどのクリニックの親を
	// 指すかは信頼しない）ため、この行は clinicB 視点では自院の正当な行として1件カウントされる
	// のが正しい挙動。同時に、下の「同一クリニックの有効な子のみカウントする」subtest が
	// clinicA からのカウントに clinicB のこの行が混入しない（2件のまま）ことも検証する。
	makeConsultation(t, db, clinicB, "別クリニックの子", &parent.ID)

	t.Run("同一クリニックの有効な子のみカウントする", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "別クリニックの同名parent_id行が混入してはならない")
	})

	t.Run("子が無い場合は0を返す", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, uint64(9999999))
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("clinic_id隔離: カウント対象は常に問い合わせ元クリニック自身の行のみ", func(t *testing.T) {
		// clinicB は自院に parent_id=parent.ID の行を1件持つため、その1件が正しくカウントされる
		// （parent.ID が clinicA 所属という事実は clinic_id スコープの判定に影響しない）。
		count, err := repo.CountChildrenByParentID(ctx, clinicB, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestConsultationRepository_CountUsageByConsultationID(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "診察使用飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "診察使用犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "CONS-USAGE-1", time.Now())
	consultation := makeConsultation(t, db, clinicA, "使用中診察", nil)
	unused := makeConsultation(t, db, clinicA, "未使用診察", nil)

	makeConsultationTreatment(t, db, mr.ID, consultation.ID)
	deletedTreatment := makeConsultationTreatment(t, db, mr.ID, consultation.ID)
	require.NoError(t, db.Delete(deletedTreatment).Error) // ソフトデリート

	t.Run("使用中のtreatmentのみカウントする（ソフトデリート除外）", func(t *testing.T) {
		count, err := repo.CountUsageByConsultationID(ctx, clinicA, consultation.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("未使用の診察マスタは0を返す", func(t *testing.T) {
		count, err := repo.CountUsageByConsultationID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("別クリニックからは使用実績が見えない（clinic_id隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByConsultationID(ctx, clinicB, consultation.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
