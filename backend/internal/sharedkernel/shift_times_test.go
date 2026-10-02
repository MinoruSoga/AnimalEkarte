package sharedkernel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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

// EMR-241: 時刻が不要な区分（off / paid_leave）でも「片方だけ」の時刻指定は
// 半分欠けたレコードが永続化されるため拒否する。両方 nil のみ許容し、
// 両方指定された場合は従来どおり時刻フィールドを無視して受理する。
func TestValidateShiftTimes_OneSidedRejectedForNonWorkingShifts(t *testing.T) {
	start := "09:00:00"
	end := "18:00:00"

	nonWorkingTypes := []struct {
		name      string
		shiftType model.ShiftType
	}{
		{name: "off", shiftType: model.ShiftTypeOff},
		{name: "paid_leave", shiftType: model.ShiftTypePaidLeave},
	}

	for _, nt := range nonWorkingTypes {
		t.Run(nt.name+" rejects one-sided times (start only)", func(t *testing.T) {
			err := ValidateShiftTimes(nt.shiftType, &start, nil)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(nt.name+" rejects one-sided times (end only)", func(t *testing.T) {
			err := ValidateShiftTimes(nt.shiftType, nil, &end)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(nt.name+" allows both-absent times", func(t *testing.T) {
			assert.NoError(t, ValidateShiftTimes(nt.shiftType, nil, nil))
		})
	}
}

// EMR-241: RequiresTimeSlot の区分判定をピンする。off / paid_leave だけが
// 時刻任意で、それ以外の勤務区分（full / morning / afternoon）は両時刻必須。
func TestRequiresTimeSlot(t *testing.T) {
	tests := []struct {
		name      string
		shiftType model.ShiftType
		want      bool
	}{
		{name: "full requires times", shiftType: model.ShiftTypeFull, want: true},
		{name: "morning requires times", shiftType: model.ShiftTypeMorning, want: true},
		{name: "afternoon requires times", shiftType: model.ShiftTypeAfternoon, want: true},
		{name: "off does not require times", shiftType: model.ShiftTypeOff, want: false},
		{name: "paid_leave does not require times", shiftType: model.ShiftTypePaidLeave, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RequiresTimeSlot(tt.shiftType))
		})
	}
}

// EMR-241: 勤務区分では「両方空」と「片方だけ」の時刻を拒否し、start<end を強制する
// ルールを全 RequiresTimeSlot 区分でピンする（staff の shift entry と reservation の
// reservation schedule が共有する契約。EMR-37 A-12 裁定により両時刻必須を維持）。
func TestValidateShiftTimes_WorkingCategoryTimeRules(t *testing.T) {
	start := "09:00:00"
	end := "18:00:00"
	earlier := "08:00:00"

	workingTypes := []struct {
		name      string
		shiftType model.ShiftType
	}{
		{name: "full", shiftType: model.ShiftTypeFull},
		{name: "morning", shiftType: model.ShiftTypeMorning},
		{name: "afternoon", shiftType: model.ShiftTypeAfternoon},
	}

	for _, wt := range workingTypes {
		t.Run(wt.name+" rejects both-empty times", func(t *testing.T) {
			err := ValidateShiftTimes(wt.shiftType, nil, nil)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(wt.name+" rejects one-sided times (start only)", func(t *testing.T) {
			err := ValidateShiftTimes(wt.shiftType, &start, nil)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(wt.name+" rejects one-sided times (end only)", func(t *testing.T) {
			err := ValidateShiftTimes(wt.shiftType, nil, &end)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(wt.name+" rejects end equal to start", func(t *testing.T) {
			err := ValidateShiftTimes(wt.shiftType, &start, &start)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(wt.name+" rejects end before start", func(t *testing.T) {
			err := ValidateShiftTimes(wt.shiftType, &end, &earlier)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})

		t.Run(wt.name+" accepts valid start<end range", func(t *testing.T) {
			assert.NoError(t, ValidateShiftTimes(wt.shiftType, &start, &end))
		})
	}
}
