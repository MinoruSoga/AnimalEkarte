package reservation

// count_clinic_scope_isolation_test.go — BE-refactor.md R2-5 (D12) の回帰防止。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupB10CountReservationIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

func TestReservationRepository_CountMedicalRecordsByReservationID_ClinicIsolation(t *testing.T) {
	db := setupB10CountReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	res := makeReservation(t, db, clinicA)
	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-A-2", Date: time.Now(), AppointmentID: &res.ID}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	t.Run("同一クリニックIDでは件数が見える", func(t *testing.T) {
		count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicA, res.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件を返す（漏洩しない）", func(t *testing.T) {
		count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicB, res.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)
	})
}
