package reservation

// preload_followup_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — ReservationType JOIN clinic 隔離回帰テスト。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// --- (a3) reservation: FindAllByCategory ReservationType ---

func TestReservationRepository_FindAllByCategory_ReservationTypeJoinClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	// TRUNCATE first: 他テストが残した orphan 行を除去してから AutoMigrate（FK 検証を通すため）。
	db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE")
	db.Exec("TRUNCATE TABLE appointments CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{},
	))
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtB := makeReservationType(t, db, clinicB) // category general
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

	got, _, err := repo.FindAllByCategory(ctx, clinicA, model.ReservationTypeCategoryGeneral, nil, nil, nil, nil, 1, 100)
	require.NoError(t, err)

	var found *model.Reservation
	for i := range got {
		if got[i].ID == res.ID {
			found = &got[i]
			break
		}
	}
	assert.Nil(t, found, "別クリニックの診療区分マスタで clinic A の予約をカテゴリ分類してはならない")
}
