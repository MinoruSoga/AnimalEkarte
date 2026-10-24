package sharedkernel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRequireStored(t *testing.T) {
	t.Run("returns the stored value", func(t *testing.T) {
		status := model.ReservationStatusConfirmed
		got, err := RequireStored(&status, "appointments.status")
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusConfirmed, got)
	})

	t.Run("nil is an error and never a default value", func(t *testing.T) {
		got, err := RequireStored[model.ReservationStatus](nil, "appointments.status")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrStoredValueMissing))
		assert.Empty(t, got)
	})
}

func TestPtrUnlessZero(t *testing.T) {
	assert.Nil(t, PtrUnlessZero(model.ReservationStatus("")))
	got := PtrUnlessZero(model.ReservationStatusPending)
	require.NotNil(t, got)
	assert.Equal(t, model.ReservationStatusPending, *got)
}

func TestPtrEquals(t *testing.T) {
	cancelled := model.ReservationStatusCancelled
	assert.True(t, PtrEquals(&cancelled, model.ReservationStatusCancelled))
	assert.False(t, PtrEquals(&cancelled, model.ReservationStatusPending))
	assert.False(t, PtrEquals(nil, model.ReservationStatusCancelled), "nil (unknown) never matches")
}

func TestTextOrEmpty(t *testing.T) {
	status := model.ReservationStatusCheckedIn
	assert.Equal(t, "checked_in", TextOrEmpty(&status))
	assert.Equal(t, "", TextOrEmpty[model.ReservationStatus](nil))
}

func TestSortOrderValue(t *testing.T) {
	order := 3
	assert.Equal(t, 3, SortOrderValue(&order))
	assert.Equal(t, 0, SortOrderValue(nil))
}

func TestPtrCopiesValue(t *testing.T) {
	src := 5
	got := Ptr(src)
	src = 9
	assert.Equal(t, 5, *got)
}
