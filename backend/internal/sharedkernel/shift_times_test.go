package sharedkernel

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHHMM(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{name: "HH:MM", input: "09:05", wantH: 9, wantM: 5},
		{name: "HH:MM:SS strips seconds", input: "18:30:00", wantH: 18, wantM: 30},
		{name: "empty", input: "", wantErr: true},
		{name: "bad separator", input: "09-05", wantErr: true},
		{name: "no zero pad", input: "9:05", wantErr: true},
		{name: "non-numeric", input: "ab:cd", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m, err := ParseHHMM(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantH, h)
			assert.Equal(t, tt.wantM, m)
		})
	}
}

// BUG-036: 勤務種別では start/end 必須。off/paid_leave は免除。
func TestValidateShiftTimes_RequiredForWorkingShifts(t *testing.T) {
	start := "09:00:00"
	end := "18:00:00"

	t.Run("full rejects both nil", func(t *testing.T) {
		err := ValidateShiftTimes(model.ShiftTypeFull, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("full accepts valid range", func(t *testing.T) {
		assert.NoError(t, ValidateShiftTimes(model.ShiftTypeFull, &start, &end))
	})

	t.Run("off allows nil times", func(t *testing.T) {
		assert.NoError(t, ValidateShiftTimes(model.ShiftTypeOff, nil, nil))
	})

	t.Run("paid_leave allows nil times", func(t *testing.T) {
		assert.NoError(t, ValidateShiftTimes(model.ShiftTypePaidLeave, nil, nil))
	})
}
