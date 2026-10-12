package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestExpiredTentativeTagCursor_QueryBoundaries(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=disabled.invalid user=test dbname=test sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	repo := &reservationRepository{db: db}

	t.Run("joins ambient query session", func(t *testing.T) {
		ambient := db.Session(&gorm.Session{NewDB: true})
		ambient.Statement.Settings.Store("expiry-ambient-proof", true)
		query := repo.expiredTentativeTagOwnersQuery(persistence.WithTxValue(context.Background(), ambient), 7, time.Now(), 0, 1)
		v, ok := query.Statement.Settings.Load("expiry-ambient-proof")
		assert.True(t, ok)
		assert.Equal(t, true, v)
	})
	for _, tt := range []struct {
		name      string
		asOf      time.Time
		date      string
		limit     int
		wantLimit int
	}{
		{"before JST midnight", time.Date(2026, 10, 11, 14, 59, 59, 0, time.UTC), "2026-10-11", 1, 1},
		{"at JST midnight", time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC), "2026-10-12", 201, 200},
		{"invalid limit is bounded", time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC), "2026-10-12", 0, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var ids []uint64
			query := repo.expiredTentativeTagOwnersQuery(context.Background(), 7, tt.asOf, 42, tt.limit).Find(&ids)
			require.NoError(t, query.Error)
			sql := query.Statement.SQL.String()

			assert.Contains(t, sql, "DISTINCT expiry_tag.owner_id")
			assert.Contains(t, sql, "expiry_tag.clinic_id =")
			assert.Contains(t, sql, "owner.deleted_at IS NULL")
			assert.Contains(t, sql, "owner.clinic_id = expiry_tag.clinic_id")
			assert.Contains(t, sql, "substring(expiry_tag.tag_name, 12, 10) <")
			assert.NotContains(t, sql, "appointments")
			assert.NotContains(t, sql, "::date")
			assert.Contains(t, sql, "ORDER BY owner_id ASC")
			assert.Contains(t, query.Statement.Vars, tt.date)
			assert.Contains(t, query.Statement.Vars, uint64(42))
			assert.Contains(t, query.Statement.Vars, tt.wantLimit)
		})
	}
}

func TestExpiredTentativeTagCursor_DBStateAndCatchUp(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTagCache{}))
	setupEmptyNextVisitRetryQueueFixture(t, db)
	repo := NewReservationRepository(db)
	const clinicA, clinicB = uint64(1), uint64(2)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "期限超過A")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "期限超過B")
	rtA := makeReservationType(t, db, clinicA)
	rtB := makeReservationType(t, db, clinicB)
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	oldDate := jstMidnight(asOf.AddDate(0, 0, -30))
	for _, entry := range []struct{ clinic, owner, rt uint64 }{{clinicA, ownerA.ID, rtA.ID}, {clinicB, ownerB.ID, rtB.ID}} {
		makeNextVisitQueryReservation(t, db, entry.clinic, entry.rt, &entry.owner, model.ReservationStatusTentative, oldDate, &oldDate)
		makeNextVisitQueryReservation(t, db, entry.clinic, entry.rt, &entry.owner, model.ReservationStatusTentative, oldDate, &oldDate)
		require.NoError(t, db.Create(&model.LstepTagCache{ClinicID: entry.clinic, OwnerID: entry.owner, TagName: "next_visit_" + oldDate.Format(time.DateOnly), Category: "auto", SyncedAt: asOf}).Error)
	}
	ids, err := repo.FindOwnerIDsWithExpiredTentativeTagCursor(context.Background(), clinicA, oldDate, 0, 200)
	require.NoError(t, err)
	assert.Empty(t, ids, "intended day is still eligible, not expired")
	ids, err = repo.FindOwnerIDsWithExpiredTentativeTagCursor(context.Background(), clinicA, asOf, 0, 200)
	require.NoError(t, err)
	assert.Equal(t, []uint64{ownerA.ID}, ids, "catch up all past dates, dedupe and isolate clinic")
	ids, err = repo.FindOwnerIDsWithExpiredTentativeTagCursor(context.Background(), clinicA, asOf, ownerA.ID, 200)
	require.NoError(t, err)
	assert.Empty(t, ids)
	require.NoError(t, db.Where("clinic_id = ? AND owner_id = ?", clinicA, ownerA.ID).Delete(&model.LstepTagCache{}).Error)
	ids, err = repo.FindOwnerIDsWithExpiredTentativeTagCursor(context.Background(), clinicA, asOf, 0, 200)
	require.NoError(t, err)
	assert.Empty(t, ids, "successful tag removal avoids repeated repair")
}

func TestExpiredTentativeTagCursor_FormerOwnersAreCacheCandidates(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=disabled.invalid user=test dbname=test sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	repo := &reservationRepository{db: db}
	var ids []uint64
	query := repo.expiredTentativeTagOwnersQuery(context.Background(), 7, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), 42, 201).Find(&ids)
	require.NoError(t, query.Error)
	sql := query.Statement.SQL.String()
	assert.Contains(t, sql, "lstep_tag_cache AS expiry_tag")
	assert.Contains(t, sql, "DISTINCT expiry_tag.owner_id")
	assert.NotContains(t, sql, "appointments")
	assert.NotContains(t, sql, "current_pet")
	assert.Contains(t, sql, "owner.clinic_id = expiry_tag.clinic_id")
	assert.Contains(t, sql, "owner.deleted_at IS NULL")
	assert.NotContains(t, sql, "::date")
	assert.Contains(t, sql, "CASE WHEN")
	assert.Contains(t, sql, "% 400")
	assert.Contains(t, sql, "ORDER BY owner_id ASC")
	assert.Contains(t, query.Statement.Vars, 200)
}

func TestExpiredTentativeTagCursor_DBTransferSoftDeleteAndMalformedDates(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTagCache{}))
	setupEmptyNextVisitRetryQueueFixture(t, db)
	repo := NewReservationRepository(db)
	a := testdb.MakeTestOwner(t, db, 1, "retry former A")
	b := testdb.MakeTestOwner(t, db, 1, "retry former B")
	c := testdb.MakeTestOwner(t, db, 1, "retry current C")
	foreign := testdb.MakeTestOwner(t, db, 2, "retry foreign")
	pet := makeNextVisitQueryPet(t, db, 1, a.ID, "retry transferred", model.PetStatusAlive, nil)
	rt := makeReservationType(t, db, 1)
	old := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	appointment := makeNextVisitQueryPetReservation(t, db, 1, rt.ID, &a.ID, &pet.ID, model.ReservationStatusTentative, old, &old)
	for _, owner := range []uint64{a.ID, b.ID} {
		require.NoError(t, db.Create(&model.LstepTagCache{ClinicID: 1, OwnerID: owner, TagName: "next_visit_2024-02-29", Category: "auto", SyncedAt: old}).Error)
	}
	require.NoError(t, db.Create(&model.LstepTagCache{ClinicID: 2, OwnerID: foreign.ID, TagName: "next_visit_2024-02-29", Category: "auto", SyncedAt: old}).Error)
	for _, tag := range []string{"next_visit_2023-02-29", "next_visit_1900-02-29", "next_visit_2024-04-31", "next_visit_0000-01-01", "next_visit_2024-13-01", "next_visit_broken", "cpm_core", "next_visit_2026-10-12", "next_visit_2026-10-13"} {
		require.NoError(t, db.Create(&model.LstepTagCache{ClinicID: 1, OwnerID: c.ID, TagName: tag, Category: "auto", SyncedAt: old}).Error)
	}
	require.NoError(t, db.Model(pet).Update("owner_id", b.ID).Error)
	require.NoError(t, db.Model(pet).Update("owner_id", c.ID).Error)
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	for _, phase := range []string{"transferred twice", "soft deleted", "appointment terminal"} {
		if phase == "soft deleted" {
			require.NoError(t, db.Delete(pet).Error)
		}
		if phase == "appointment terminal" {
			require.NoError(t, db.Model(appointment).Update("status", model.ReservationStatusCancelled).Error)
		}
		ids, err := repo.FindOwnerIDsWithExpiredTentativeTagCursor(context.Background(), 1, asOf, 0, 200)
		require.NoError(t, err)
		assert.Equal(t, []uint64{a.ID, b.ID}, ids, phase)
	}
	var stored model.Reservation
	require.NoError(t, db.Unscoped().First(&stored, appointment.ID).Error)
	assert.Equal(t, &a.ID, stored.OwnerID)
	assert.Equal(t, model.ReservationStatusCancelled, stored.Status)
}
