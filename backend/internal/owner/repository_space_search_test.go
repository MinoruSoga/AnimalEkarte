package owner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerRepository_FindAll_IdeographicSpaceFourWay(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "DB fullwidth × query fullwidth", storedName: "飼主全角全角姓　飼主全角全角名", query: "飼主全角全角姓　飼主全角全角名"},
		{name: "DB fullwidth × query halfwidth", storedName: "飼主全角半角姓　飼主全角半角名", query: "飼主全角半角姓 飼主全角半角名"},
		{name: "DB halfwidth × query fullwidth", storedName: "飼主半角全角姓 飼主半角全角名", query: "飼主半角全角姓　飼主半角全角名"},
		{name: "DB halfwidth × query halfwidth", storedName: "飼主半角半角姓 飼主半角半角名", query: "飼主半角半角姓 飼主半角半角名"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := makeTestOwner(t, db, clinicA, tt.storedName)
			foreign := makeTestOwner(t, db, clinicB, tt.storedName)

			got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, tt.query)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, fixture.ID, got[0].ID)
			require.NotEqual(t, foreign.ID, got[0].ID)
		})
	}

	t.Run("ideographic-space-only query returns no rows", func(t *testing.T) {
		_ = makeTestOwner(t, db, clinicA, "空白除外飼主 名")
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "　　")
		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		require.Empty(t, got)
	})
}

// representativeOwnerNameRows is the owner-name volume this test judges.
// ownerCompactSearchBudget is the FindAll response budget for that volume.
// Seq Scan is not a failure. The compact predicate is the migration 015
// expression idx_owners_name_compact_fold_trgm; this revision adds no migration.
const representativeOwnerNameRows = 1000
const ownerCompactSearchBudget = 2 * time.Second

func TestOwnerRepository_FindAll_CompactFamilyGivenName(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "stored ascii space query unspaced", storedName: "空白差姓 空白差名", query: "空白差姓空白差名"},
		{name: "stored unspaced query ascii space", storedName: "逆空白姓逆空白名", query: "逆空白姓 逆空白名"},
		{name: "stored ideographic space query unspaced", storedName: "全角空白姓　全角空白名", query: "全角空白姓全角空白名"},
		{name: "stored unspaced query ideographic space", storedName: "逆全角姓逆全角名", query: "逆全角姓　逆全角名"},
		{name: "stored repeated ascii spaces query unspaced", storedName: "連続空白姓   連続空白名", query: "連続空白姓連続空白名"},
		{name: "stored mixed spaces query unspaced", storedName: "混在空白姓 　混在空白名", query: "混在空白姓混在空白名"},
		{name: "stored spaced katakana query hiragana unspaced", storedName: "ヤマダ ハナコ", query: "やまだはなこ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := makeTestOwner(t, db, clinicA, tt.storedName)
			foreign := makeTestOwner(t, db, clinicB, tt.storedName)
			got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, tt.query)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, fixture.ID, got[0].ID)
			require.Equal(t, clinicA, got[0].ClinicID)
			require.NotEqual(t, foreign.ID, got[0].ID)
		})
	}
}

func TestOwnerRepository_FindAll_CompactLikePagingAndExistingFields(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("percent is a literal", func(t *testing.T) {
		match := makeTestOwner(t, db, clinicA, "百分率%姓 百分率名")
		decoy := makeTestOwner(t, db, clinicA, "百分率X姓 百分率名")
		foreign := makeTestOwner(t, db, clinicB, "百分率%姓 百分率名")
		requireOwnerSearchHit(t, repo, ctx, clinicA, "百分率%姓百分率名", match.ID)
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "百分率%姓百分率名")
		require.NoError(t, err)
		require.NotEqual(t, decoy.ID, got[0].ID)
		require.NotEqual(t, foreign.ID, got[0].ID)
	})

	t.Run("underscore is a literal", func(t *testing.T) {
		match := makeTestOwner(t, db, clinicA, "下線_姓 下線名")
		decoy := makeTestOwner(t, db, clinicA, "下線X姓 下線名")
		requireOwnerSearchHit(t, repo, ctx, clinicA, "下線_姓下線名", match.ID)
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "下線_姓下線名")
		require.NoError(t, err)
		require.NotEqual(t, decoy.ID, got[0].ID)
	})

	t.Run("count matches paging", func(t *testing.T) {
		first := makeTestOwner(t, db, clinicA, "頁送り姓 頁送り名")
		second := makeTestOwner(t, db, clinicA, "頁送り姓頁送り名")
		page1, total1, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 1, "頁送り姓頁送り名")
		require.NoError(t, err)
		page2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, 2, 1, "頁送り姓頁送り名")
		require.NoError(t, err)
		require.Equal(t, int64(2), total1)
		require.Equal(t, total1, total2)
		require.Len(t, page1, 1)
		require.Len(t, page2, 1)
		require.NotEqual(t, page1[0].ID, page2[0].ID)
		require.ElementsMatch(t, []uint64{first.ID, second.ID}, []uint64{page1[0].ID, page2[0].ID})
	})

	t.Run("kana query still hits name_kana", func(t *testing.T) {
		match := makeTestOwner(t, db, clinicA, "漢字のみ飼主")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", match.ID).Update("name_kana", "けんさくよみがな").Error)
		decoy := makeTestOwner(t, db, clinicA, "別漢字飼主")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", decoy.ID).Update("name_kana", "べつかな").Error)
		requireOwnerSearchHit(t, repo, ctx, clinicA, "ケンサクヨミガナ", match.ID)
	})

	t.Run("phone query still hits", func(t *testing.T) {
		match := makeTestOwner(t, db, clinicA, "電話照合飼主")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", match.ID).Update("phone", "080-2731-4455").Error)
		_ = makeTestOwner(t, db, clinicA, "別電話飼主")
		requireOwnerSearchHit(t, repo, ctx, clinicA, "080-2731-4455", match.ID)
	})
}

func TestOwnerRepository_FindAll_WildcardOnlyDoesNotListClinic(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	_ = makeTestOwner(t, db, clinicA, "記号除外姓 記号除外名")

	for _, query := range []string{"%", "_"} {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, query)
		require.NoError(t, err)
		require.Equal(t, int64(0), total, "query=%s", query)
		require.Empty(t, got, "query=%s", query)
	}
}

func TestOwnerRepository_FindAll_OwnerNumberStaysInCallerClinic(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	// 飼主No is owners.id. Owner FindAll has no id arm; pet search owns that exact
	// match. An id-shaped query must not return another clinic or every row.
	local := makeTestOwner(t, db, clinicA, "番号隔離姓 番号隔離名")
	foreign := makeTestOwner(t, db, clinicB, "番号隔離姓 番号隔離名")
	require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", local.ID).Updates(map[string]any{
		"phone": "090-8000-8000", "name_kana": "ばんごうかくり", "email": "local-owner-number@example.test",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", foreign.ID).Updates(map[string]any{
		"phone": "090-7000-7000", "name_kana": "べつばんごう", "email": "foreign-owner-number@example.test",
	}).Error)

	for _, id := range []uint64{local.ID, foreign.ID} {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, fmt.Sprintf("%d", id))
		require.NoError(t, err)
		require.Equal(t, int64(0), total, "owner number=%d", id)
		require.Empty(t, got, "owner number=%d", id)
	}
}

func TestOwnerRepository_FindAll_CompactSearchRepresentativeBudget(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const stored = "予算対象姓 予算対象名"
	const query = "予算対象姓予算対象名"

	fillers := make([]model.Owner, 0, representativeOwnerNameRows-1)
	for i := 0; i < representativeOwnerNameRows-1; i++ {
		fillers = append(fillers, model.Owner{
			ClinicID: clinicA,
			Name:     fmt.Sprintf("無関係予算飼主%04d", i),
		})
	}
	require.NoError(t, db.WithContext(ctx).CreateInBatches(fillers, 200).Error)
	target := makeTestOwner(t, db, clinicA, stored)
	foreign := makeTestOwner(t, db, clinicB, stored)

	started := time.Now()
	got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 20, query)
	elapsed := time.Since(started)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	require.Equal(t, target.ID, got[0].ID)
	require.NotEqual(t, foreign.ID, got[0].ID)
	require.Less(t, elapsed, ownerCompactSearchBudget,
		"representativeOwnerNameRows=%d budget=%s elapsed=%s",
		representativeOwnerNameRows, ownerCompactSearchBudget, elapsed)
}

func requireOwnerSearchHit(t *testing.T, repo Repository, ctx context.Context, clinicID uint64, query string, wantID uint64) {
	t.Helper()
	got, total, err := repo.FindAll(ctx, []uint64{clinicID}, 1, 100, query)
	require.NoError(t, err)
	require.Equal(t, int64(1), total, "query=%s", query)
	require.Len(t, got, 1, "query=%s", query)
	require.Equal(t, wantID, got[0].ID, "query=%s", query)
	require.Equal(t, clinicID, got[0].ClinicID)
}
