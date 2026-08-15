package lstep

// lstep_tag_code_mapping_repository_test.go — LstepTagCodeMappingRepository 統合テスト。
//
// 保護する不変条件:
//   - FindAllByClinicID / FindByClinicIDAndTagName は clinic_id で分離され、
//     deleted_at IS NULL のソフトデリート済み行を除外する。
//   - SoftDelete / SoftDeleteByClinicIDAndTagName は deleted_at を設定するのみで物理削除はしない
//     （行は DB に残るが FindAll* からは除外される）。

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepTagCodeMappingTestDB は lstep_tag_code_mappings テーブルを用意する。
func setupLstepTagCodeMappingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTagCodeMapping{}))
	db.Exec("TRUNCATE TABLE lstep_tag_code_mappings CASCADE")
	return db
}

func TestLstepTagCodeMappingRepository_Create(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	repo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	m := &model.LstepTagCodeMapping{
		ClinicID: 1,
		TagName:  "健診対象",
		CodeType: model.CodeTypeCheckupType,
		Codes:    pq.StringArray{"CT001", "CT002"},
	}
	require.NoError(t, repo.Create(ctx, m))
	assert.NotZero(t, m.ID)

	var stored model.LstepTagCodeMapping
	require.NoError(t, db.Where("id = ?", m.ID).First(&stored).Error)
	assert.Equal(t, "健診対象", stored.TagName)
	assert.ElementsMatch(t, []string{"CT001", "CT002"}, []string(stored.Codes))
}

func TestLstepTagCodeMappingRepository_FindAllByClinicID(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	repo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	require.NoError(t, repo.Create(ctx, &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "tag-1", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C1"}}))
	toSoftDelete := &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "tag-2", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C2"}}
	require.NoError(t, repo.Create(ctx, toSoftDelete))
	require.NoError(t, repo.Create(ctx, &model.LstepTagCodeMapping{ClinicID: clinicB, TagName: "tag-3", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C3"}}))

	require.NoError(t, repo.SoftDelete(ctx, clinicA, toSoftDelete.ID))

	t.Run("clinic_id分離・ソフトデリート除外の両方を満たす", func(t *testing.T) {
		mappings, err := repo.FindAllByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, mappings, 1)
		assert.Equal(t, "tag-1", mappings[0].TagName)
	})

	t.Run("ソフトデリート済み行はDBに残る（物理削除ではない）", func(t *testing.T) {
		var stored model.LstepTagCodeMapping
		require.NoError(t, db.Where("id = ?", toSoftDelete.ID).First(&stored).Error, "行はDBに残っていること")
		assert.NotNil(t, stored.DeletedAt)
	})

	t.Run("別クリニックの行は含まれない", func(t *testing.T) {
		mappings, err := repo.FindAllByClinicID(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, mappings, 1)
		assert.Equal(t, "tag-3", mappings[0].TagName)
	})
}

func TestLstepTagCodeMappingRepository_FindByClinicIDAndTagName(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	repo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	require.NoError(t, repo.Create(ctx, &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "shared-tag", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C1"}}))
	require.NoError(t, repo.Create(ctx, &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "shared-tag", CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"P1"}}))
	require.NoError(t, repo.Create(ctx, &model.LstepTagCodeMapping{ClinicID: clinicB, TagName: "shared-tag", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C9"}}))

	t.Run("同一タグ名で複数code_typeが存在すれば全て返す", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicA, "shared-tag")
		require.NoError(t, err)
		assert.Len(t, mappings, 2)
	})

	t.Run("別クリニックの同名タグは含まれない", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicB, "shared-tag")
		require.NoError(t, err)
		require.Len(t, mappings, 1)
		assert.Equal(t, model.CodeTypeCheckupType, mappings[0].CodeType)
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicA, "nonexistent-tag")
		require.NoError(t, err)
		assert.Empty(t, mappings)
	})
}

func TestLstepTagCodeMappingRepository_SoftDelete(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	repo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	t.Run("対象行のdeleted_atのみ設定される", func(t *testing.T) {
		m := &model.LstepTagCodeMapping{ClinicID: 1, TagName: "to-delete", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C1"}}
		require.NoError(t, repo.Create(ctx, m))

		require.NoError(t, repo.SoftDelete(ctx, 1, m.ID))

		var stored model.LstepTagCodeMapping
		require.NoError(t, db.Where("id = ?", m.ID).First(&stored).Error)
		assert.NotNil(t, stored.DeletedAt)
	})

	t.Run("別clinic_idからのSoftDeleteはNotFoundになり対象行を変更しない", func(t *testing.T) {
		m := &model.LstepTagCodeMapping{ClinicID: 5, TagName: "protected", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C1"}}
		require.NoError(t, repo.Create(ctx, m))

		err := repo.SoftDelete(ctx, 999, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stored model.LstepTagCodeMapping
		require.NoError(t, db.Where("id = ?", m.ID).First(&stored).Error)
		assert.Nil(t, stored.DeletedAt, "別clinic_idからの削除は反映されない")
	})
}

func TestLstepTagCodeMappingRepository_SoftDeleteByClinicIDAndTagName(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	repo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	m1 := &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "bulk-tag", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C1"}}
	m2 := &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "bulk-tag", CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"P1"}}
	m3 := &model.LstepTagCodeMapping{ClinicID: clinicA, TagName: "other-tag", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C2"}}
	otherClinic := &model.LstepTagCodeMapping{ClinicID: clinicB, TagName: "bulk-tag", CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"C3"}}
	require.NoError(t, repo.Create(ctx, m1))
	require.NoError(t, repo.Create(ctx, m2))
	require.NoError(t, repo.Create(ctx, m3))
	require.NoError(t, repo.Create(ctx, otherClinic))

	require.NoError(t, repo.SoftDeleteByClinicIDAndTagName(ctx, clinicA, "bulk-tag"))

	t.Run("同一タグ名の全 code_type がソフトデリートされる", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicA, "bulk-tag")
		require.NoError(t, err)
		assert.Empty(t, mappings)
	})

	t.Run("別タグ名の行は影響を受けない", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicA, "other-tag")
		require.NoError(t, err)
		require.Len(t, mappings, 1)
	})

	t.Run("別クリニックの同名タグは影響を受けない", func(t *testing.T) {
		mappings, err := repo.FindByClinicIDAndTagName(ctx, clinicB, "bulk-tag")
		require.NoError(t, err)
		require.Len(t, mappings, 1)
	})
}

type failNthCreateLstepTagCodeMappingRepository struct {
	LstepTagCodeMappingRepository
	failAt      int
	createCalls int
}

func (r *failNthCreateLstepTagCodeMappingRepository) Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error {
	r.createCalls++
	if r.createCalls == r.failAt {
		return errors.New("injected create failure")
	}
	return r.LstepTagCodeMappingRepository.Create(ctx, mapping)
}

func TestLstepTagCodeMappingService_PutMappingsForTag_RollsBackPartialReplacement(t *testing.T) {
	db := setupLstepTagCodeMappingTestDB(t)
	baseRepo := NewLstepTagCodeMappingRepository(db)
	ctx := context.Background()

	const clinicID = uint64(1)
	original := &model.LstepTagCodeMapping{
		ClinicID: clinicID,
		TagName:  HlthHealthcheckDoneTag,
		CodeType: model.CodeTypeCheckupType,
		Codes:    pq.StringArray{"ORIGINAL"},
	}
	require.NoError(t, baseRepo.Create(ctx, original))

	failingRepo := &failNthCreateLstepTagCodeMappingRepository{
		LstepTagCodeMappingRepository: baseRepo,
		failAt:                        2,
	}
	svc := NewLstepTagCodeMappingService(failingRepo, persistence.NewTransactor(db))

	got, err := svc.PutMappingsForTag(ctx, clinicID, HlthHealthcheckDoneTag, []PutMappingEntry{
		{CodeType: model.CodeTypePrescription, Codes: []string{"RX_NEW"}},
		{CodeType: model.CodeTypeMerchandiseItem, Codes: []string{"FOOD_NEW"}},
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, 2, failingRepo.createCalls)

	active, findErr := baseRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
	require.NoError(t, findErr)
	if assert.Len(t, active, 1) {
		assert.Equal(t, original.ID, active[0].ID)
		assert.Equal(t, pq.StringArray{"ORIGINAL"}, active[0].Codes)
		assert.Nil(t, active[0].DeletedAt)
	}

	var allRows []*model.LstepTagCodeMapping
	require.NoError(t, db.Where("clinic_id = ? AND tag_name = ?", clinicID, HlthHealthcheckDoneTag).Find(&allRows).Error)
	assert.Len(t, allRows, 1, "soft delete and successful creates before the failure must roll back together")
}
