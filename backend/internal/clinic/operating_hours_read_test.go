package clinic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type fakeOperatingHoursRepository struct {
	row    model.ClinicOperatingHours
	found  bool
	err    error
	calls  int
	gotIDs []uint64
}

func (f *fakeOperatingHoursRepository) FindByClinicID(_ context.Context, clinicID uint64) (model.ClinicOperatingHours, bool, error) {
	f.calls++
	f.gotIDs = append(f.gotIDs, clinicID)
	return f.row, f.found, f.err
}

func TestOperatingHoursReadService_Get(t *testing.T) {
	t.Run("absent row is unconfigured, not an error and not complemented", func(t *testing.T) {
		repo := &fakeOperatingHoursRepository{}
		got, err := NewOperatingHoursReadService(repo).Get(context.Background(), 7)
		require.NoError(t, err)
		assert.False(t, got.Configured)
		assert.Zero(t, got.Version)
		assert.False(t, got.Settings.Valid(), "unconfigured must not carry a Settings value (no all-open/all-closed default)")
		assert.Equal(t, []uint64{7}, repo.gotIDs)
	})

	t.Run("stored row becomes configured with the stored version", func(t *testing.T) {
		repo := &fakeOperatingHoursRepository{found: true, row: operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)}
		got, err := NewOperatingHoursReadService(repo).Get(context.Background(), 7)
		require.NoError(t, err)
		assert.True(t, got.Configured)
		assert.EqualValues(t, 3, got.Version)
		assert.True(t, got.Settings.Valid())
	})

	t.Run("repository error propagates as unavailable and never as unconfigured", func(t *testing.T) {
		boom := errors.New("connection reset")
		got, err := NewOperatingHoursReadService(&fakeOperatingHoursRepository{err: boom}).Get(context.Background(), 7)
		require.ErrorIs(t, err, ErrOperatingHoursUnavailable)
		require.ErrorIs(t, err, boom)
		assert.False(t, got.Configured)
	})

	t.Run("corrupt stored row is a distinct sentinel", func(t *testing.T) {
		repo := &fakeOperatingHoursRepository{found: true, row: operatingHoursStoredRow(`[]`, `[]`)}
		_, err := NewOperatingHoursReadService(repo).Get(context.Background(), 7)
		requireCorrupt(t, err)
		assert.NotErrorIs(t, err, ErrOperatingHoursUnavailable)
	})

	t.Run("row of another clinic is corrupt", func(t *testing.T) {
		row := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)
		row.ClinicID = 8
		_, err := NewOperatingHoursReadService(&fakeOperatingHoursRepository{found: true, row: row}).Get(context.Background(), 7)
		requireCorrupt(t, err)
	})

	t.Run("missing repository is unavailable", func(t *testing.T) {
		_, err := NewOperatingHoursReadService(nil).Get(context.Background(), 7)
		require.ErrorIs(t, err, ErrOperatingHoursUnavailable)
	})
}

func TestParseYearMonth(t *testing.T) {
	for _, valid := range []string{"2026-01", "2026-11", "2026-12", "0000-01", "9999-12"} {
		t.Run("valid "+valid, func(t *testing.T) {
			got, err := ParseYearMonth(valid)
			require.NoError(t, err)
			assert.Equal(t, valid, got.String())
		})
	}
	for _, invalid := range []string{
		"", "2026-1", "2026-00", "2026-13", "2026-11-01", "26-11", "2026/11", "2026-1a", "abcd-11",
		" 2026-11", "2026-11 ", "２０２６-11", "2026-１１", "+026-11", "2026-110",
	} {
		t.Run("invalid "+invalid, func(t *testing.T) {
			_, err := ParseYearMonth(invalid)
			require.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
		})
	}
}

func TestCommonOperatingHours_ClosedDatesIn(t *testing.T) {
	closed := `[{"date":"2026-10-31","reason":"前月末"},{"date":"2026-11-01","reason":"月初"},{"date":"2026-11-30","reason":"月末"},{"date":"2026-12-01","reason":"翌月初"},{"date":"2027-11-05","reason":"翌年同月"}]`
	repo := &fakeOperatingHoursRepository{found: true, row: operatingHoursStoredRow(operatingHoursSevenClosedDays, closed)}
	current, err := NewOperatingHoursReadService(repo).Get(context.Background(), 7)
	require.NoError(t, err)

	month, err := ParseYearMonth("2026-11")
	require.NoError(t, err)
	got, err := current.ClosedDatesIn(month)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "2026-11-01", got[0].Date.String())
	assert.Equal(t, "2026-11-30", got[1].Date.String())

	empty, err := ParseYearMonth("2026-02")
	require.NoError(t, err)
	none, err := current.ClosedDatesIn(empty)
	require.NoError(t, err)
	assert.NotNil(t, none)
	assert.Empty(t, none)

	unconfigured, err := CommonOperatingHours{}.ClosedDatesIn(month)
	require.NoError(t, err)
	assert.NotNil(t, unconfigured)
	assert.Empty(t, unconfigured)

	_, err = current.ClosedDatesIn(YearMonth{})
	require.Error(t, err, "a month not built by ParseYearMonth is rejected")
	_, err = CommonOperatingHours{Configured: true, Version: 1}.ClosedDatesIn(month)
	require.Error(t, err, "configured with invalid settings is a misuse, not an empty month")
}

func TestOperatingHoursResponses_RejectInconsistentConfiguredState(t *testing.T) {
	settingsOnly := CommonOperatingHours{Configured: true, Version: 0}
	_, err := NewClinicOperatingHoursResponse(7, settingsOnly)
	require.Error(t, err, "configured with a null/zero version must never be produced")
	month, parseErr := ParseYearMonth("2026-11")
	require.NoError(t, parseErr)
	_, err = NewClinicCommonHolidayMonthResponse(7, month, settingsOnly)
	require.Error(t, err)
	_, err = NewClinicOperatingHoursResponse(7, CommonOperatingHours{Configured: true, Version: 1})
	require.Error(t, err, "configured with null settings must never be produced")
}
