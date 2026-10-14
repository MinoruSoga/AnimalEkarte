package medicalrecord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMedicalRecordRepository_DormantCursors_UseCalendarDays(t *testing.T) {
	for _, hour := range []int{0, 8, 9, 23} {
		t.Run(fmt.Sprintf("local_hour_%02d", hour), func(t *testing.T) {
			db := testdb.SetupTestDB(t)
			repo := NewMedicalRecordRepository(db)
			evaluatedAt := time.Date(2026, time.October, 15, hour, 30, 0, 0, time.Local)
			seedDormantOwners(t, db, 1, 1, evaluatedAt.AddDate(0, 0, -250))
			cursors := []struct {
				name string
				read func(context.Context, uint64, int, uint64, int, time.Time) ([]DormantOwnerEntry, error)
			}{
				{"notification", repo.FindDormantOwnerEntriesCursorAt},
				{"history", repo.FindDormantOwnerHistoryEntriesCursorAt},
			}
			for _, cursor := range cursors {
				t.Run(cursor.name, func(t *testing.T) {
					entries, err := cursor.read(context.Background(), 1, 180, 0, 10, evaluatedAt)
					require.NoError(t, err)
					require.Len(t, entries, 1)
					assert.Equal(t, 250, entries[0].DaysSince)
				})
			}
		})
	}
	t.Run("non_cursor_current_calendar_day", func(t *testing.T) {
		db := testdb.SetupTestDB(t)
		repo := NewMedicalRecordRepository(db)
		evaluatedAt := time.Now().In(time.Local)
		seedDormantOwners(t, db, 1, 1, evaluatedAt.AddDate(0, 0, -250))
		entries, err := repo.FindDormantOwnerEntries(context.Background(), 1, 180)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, 250, entries[0].DaysSince)
	})
}

func TestDormantCalendarDaysSince(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	visitDate := time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -250)
	for _, hour := range []int{0, 8, 9, 23} {
		t.Run(fmt.Sprintf("JST_hour_%02d", hour), func(t *testing.T) {
			evaluatedAt := time.Date(2026, time.October, 15, hour, 30, 0, 0, jst)
			assert.Equal(t, 250, dormantCalendarDaysSince(evaluatedAt, visitDate))
		})
	}
	cases := []struct {
		name        string
		evaluatedAt time.Time
		visitDate   time.Time
		want        int
	}{
		{"leap_day", time.Date(2024, time.March, 1, 0, 30, 0, 0, jst), time.Date(2024, time.February, 28, 0, 0, 0, 0, time.UTC), 2},
		{"year_boundary", time.Date(2026, time.January, 1, 0, 30, 0, 0, jst), time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC), 1},
		{"same_date", time.Date(2026, time.October, 15, 0, 30, 0, 0, jst), time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC), 0},
		{"future_date", time.Date(2026, time.October, 15, 23, 30, 0, 0, jst), time.Date(2026, time.October, 16, 0, 0, 0, 0, time.UTC), -1},
		{"offset_change", time.Date(2026, time.March, 9, 0, 30, 0, 0, time.FixedZone("EDT", -4*60*60)), time.Date(2026, time.March, 7, 0, 0, 0, 0, time.FixedZone("EST", -5*60*60)), 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, dormantCalendarDaysSince(tc.evaluatedAt, tc.visitDate))
		})
	}
}
