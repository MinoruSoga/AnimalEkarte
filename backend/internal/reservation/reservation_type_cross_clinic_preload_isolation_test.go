package reservation

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

func TestReservationRepository_ReservationType_CrossClinicPreloadIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	db.Exec("TRUNCATE TABLE appointments CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationTypeGroup{}, &model.ReservationType{}, &model.Reservation{}))
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationType(t, db, clinicA)
	groupB := &model.ReservationTypeGroup{ClinicID: clinicB, Name: "医院Bのグループ"}
	require.NoError(t, db.WithContext(ctx).Create(groupB).Error)
	rtB := &model.ReservationType{ClinicID: clinicB, Name: "医院Bの診療区分", Category: model.ReservationTypeCategoryGeneral, GroupID: &groupB.ID}
	require.NoError(t, db.WithContext(ctx).Create(rtB).Error)

	resLegit := makeReservationWithType(t, db, clinicA, rtA.ID)
	resCross := makeReservationWithType(t, db, clinicA, rtB.ID) // 別clinicの診療区分FKを植え付け

	// (legit) 正規の診療区分は Preload される
	gotLegit, err := repo.FindByID(ctx, clinicA, resLegit.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.ReservationType)
	assert.Equal(t, rtA.ID, gotLegit.ReservationType.ID)

	// (i) 別テナント診療区分を指す予約は FindByID(単一)で fail-closed になる
	gotCross, err := repo.FindByID(ctx, clinicA, resCross.ID)
	require.Error(t, err)
	assert.Nil(t, gotCross)
	assert.True(t, apperrors.IsNotFound(err), "別クリニックの診療区分を指す予約は NotFound にする: %v", err)

	// (ii) #86 [A,B] でも、clinic A の予約と clinic B の診療区分は相関しない
	gotBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, resCross.ID)
	require.Error(t, err)
	assert.Nil(t, gotBoth)
	assert.True(t, apperrors.IsNotFound(err), "複数医院認可でも関連は予約自身の clinic と相関させる: %v", err)

	// (iii) #86 [A] のみでも同じ fail-closed 契約を維持する
	gotA, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, resCross.ID)
	require.Error(t, err)
	assert.Nil(t, gotA)
	assert.True(t, apperrors.IsNotFound(err), "認可外 clinic の診療区分を指す予約は NotFound にする: %v", err)
}

func makeReservationWithType(t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64) *model.Reservation {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}
