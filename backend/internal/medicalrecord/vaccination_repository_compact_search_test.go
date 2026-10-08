package medicalrecord

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestVaccinationRepository_FindAll_OwnerNameCompactSpace(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "stored ascii space query unspaced", storedName: "接種空白差姓 接種空白差名", query: "接種空白差姓接種空白差名"},
		{name: "stored unspaced query ascii space", storedName: "接種逆空白姓接種逆空白名", query: "接種逆空白姓 接種逆空白名"},
		{name: "stored ideographic space query unspaced", storedName: "接種全角姓　接種全角名", query: "接種全角姓接種全角名"},
		{name: "stored unspaced query ideographic space", storedName: "接種逆全角姓接種逆全角名", query: "接種逆全角姓　接種逆全角名"},
		{name: "stored repeated ascii spaces query unspaced", storedName: "接種連続姓   接種連続名", query: "接種連続姓接種連続名"},
		{name: "stored mixed spaces query unspaced", storedName: "接種混在姓 　接種混在名", query: "接種混在姓接種混在名"},
		{name: "stored spaced katakana query hiragana unspaced", storedName: "スズキ ハナコ", query: "すずきはなこ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rec := makeVaccinationForOwnerName(t, db, clinicA, tt.storedName, "pet-"+tt.name, "vaccine-"+tt.name)
			_, foreign := makeVaccinationForOwnerName(t, db, clinicB, tt.storedName, "foreign-pet-"+tt.name, "foreign-vaccine-"+tt.name)
			got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, tt.query, 1, 100)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, rec.ID, got[0].ID)
			require.Equal(t, clinicA, got[0].ClinicID)
			require.NotEqual(t, foreign.ID, got[0].ID)
		})
	}
}

func TestVaccinationRepository_FindAll_CompactLikePagingAndExistingFields(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	t.Run("percent is a literal", func(t *testing.T) {
		_, match := makeVaccinationForOwnerName(t, db, clinicA, "接種率%姓 接種率名", "pet-percent", "vaccine-percent")
		_, decoy := makeVaccinationForOwnerName(t, db, clinicA, "接種率X姓 接種率名", "pet-percent-decoy", "vaccine-percent-decoy")
		_, foreign := makeVaccinationForOwnerName(t, db, clinicB, "接種率%姓 接種率名", "pet-percent-foreign", "vaccine-percent-foreign")
		requireVaccinationSearchHit(t, repo, ctx, clinicA, "接種率%姓接種率名", match.ID)
		got, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, "接種率%姓接種率名", 1, 100)
		require.NoError(t, err)
		require.NotEqual(t, decoy.ID, got[0].ID)
		require.NotEqual(t, foreign.ID, got[0].ID)
	})

	t.Run("underscore is a literal", func(t *testing.T) {
		_, match := makeVaccinationForOwnerName(t, db, clinicA, "接種下線_姓 接種下線名", "pet-under", "vaccine-under")
		_, decoy := makeVaccinationForOwnerName(t, db, clinicA, "接種下線X姓 接種下線名", "pet-under-decoy", "vaccine-under-decoy")
		requireVaccinationSearchHit(t, repo, ctx, clinicA, "接種下線_姓接種下線名", match.ID)
		got, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, "接種下線_姓接種下線名", 1, 100)
		require.NoError(t, err)
		require.NotEqual(t, decoy.ID, got[0].ID)
	})

	t.Run("count matches paging", func(t *testing.T) {
		_, first := makeVaccinationForOwnerName(t, db, clinicA, "接種頁姓 接種頁名", "pet-page-1", "vaccine-page-1")
		_, second := makeVaccinationForOwnerName(t, db, clinicA, "接種頁姓接種頁名", "pet-page-2", "vaccine-page-2")
		page1, total1, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, "接種頁姓接種頁名", 1, 1)
		require.NoError(t, err)
		page2, total2, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, "接種頁姓接種頁名", 2, 1)
		require.NoError(t, err)
		require.Equal(t, int64(2), total1)
		require.Equal(t, total1, total2)
		require.Len(t, page1, 1)
		require.Len(t, page2, 1)
		require.NotEqual(t, page1[0].ID, page2[0].ID)
		require.ElementsMatch(t, []uint64{first.ID, second.ID}, []uint64{page1[0].ID, page2[0].ID})
	})

	t.Run("pet name still hits", func(t *testing.T) {
		_, rec := makeVaccinationForOwnerName(t, db, clinicA, "ペット名照合飼主", "固有予防ペット", "固有予防ワクチン")
		requireVaccinationSearchHit(t, repo, ctx, clinicA, "固有予防ペット", rec.ID)
	})

	t.Run("kana folded owner name still hits", func(t *testing.T) {
		_, rec := makeVaccinationForOwnerName(t, db, clinicA, "タナカハナコ", "kana-pet", "kana-vaccine")
		requireVaccinationSearchHit(t, repo, ctx, clinicA, "たなかはなこ", rec.ID)
	})

	t.Run("phone query stays inside the caller clinic", func(t *testing.T) {
		owner, _ := makeVaccinationForOwnerName(t, db, clinicA, "電話非対象飼主", "phone-pet", "phone-vaccine")
		foreignOwner, foreignRec := makeVaccinationForOwnerName(t, db, clinicB, "電話非対象飼主", "phone-foreign-pet", "phone-foreign-vaccine")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", owner.ID).Update("phone", "080-2732-1000").Error)
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", foreignOwner.ID).Update("phone", "080-2732-1000").Error)
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, "080-2732-1000", 1, 100)
		require.NoError(t, err)
		require.Equal(t, clinicB, foreignRec.ClinicID)
		require.Equal(t, int64(0), total)
		require.Empty(t, got)
	})
}

func TestVaccinationRepository_FindAll_OwnerNumberStaysInCallerClinic(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	// Vaccination search matches pet, owner, and vaccine names. It has no owners.id
	// arm. An id-shaped query must not return another clinic.
	localOwner, _ := makeVaccinationForOwnerName(t, db, clinicA, "番号隔離接種姓 番号隔離接種名", "number-pet", "number-vaccine")
	foreignOwner, foreignRec := makeVaccinationForOwnerName(t, db, clinicB, "番号隔離接種姓 番号隔離接種名", "number-foreign-pet", "number-foreign-vaccine")

	for _, id := range []uint64{localOwner.ID, foreignOwner.ID} {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, fmt.Sprintf("%d", id), 1, 100)
		require.NoError(t, err)
		require.Equal(t, int64(0), total, "owner number=%d", id)
		for _, row := range got {
			require.NotEqual(t, foreignRec.ID, row.ID)
			require.Equal(t, clinicA, row.ClinicID)
		}
	}
}

func makeVaccinationForOwnerName(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	ownerName, petName, vaccineName string,
) (*model.Owner, *model.Vaccination) {
	t.Helper()
	owner := makeTestOwner(t, db, clinicID, ownerName)
	pet := makeVaccinationRepoTestPet(t, db, clinicID, owner.ID, petName)
	vaccine := makeVaccineMaster(t, db, clinicID, vaccineName)
	return owner, makeVaccinationRecord(t, db, clinicID, pet.ID, vaccine.ID)
}

func requireVaccinationSearchHit(
	t *testing.T,
	repo VaccinationRepository,
	ctx context.Context,
	clinicID uint64,
	query string,
	wantID uint64,
) {
	t.Helper()
	got, total, err := repo.FindAll(ctx, clinicID, nil, nil, nil, nil, query, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(1), total, "query=%s", query)
	require.Len(t, got, 1, "query=%s", query)
	require.Equal(t, wantID, got[0].ID, "query=%s", query)
	require.Equal(t, clinicID, got[0].ClinicID)
}
