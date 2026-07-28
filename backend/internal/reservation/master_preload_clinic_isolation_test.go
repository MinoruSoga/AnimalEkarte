package reservation

// master_preload_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — single-clinic master Preload 隔離回帰テスト。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// --- (b5) appointment_admin: ReservationType ---

func TestReservationAdminRepository_FindByIDForNotify_ReservationTypePreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE appointments CASCADE").Error)
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtB := makeReservationType(t, db, clinicB)
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicA,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: rtB.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(ctx).Create(res).Error)

	got, err := repo.FindByIDForNotify(ctx, clinicA, res.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.ReservationType, "別クリニックの診療区分マスタが Preload で混入してはならない")
}
