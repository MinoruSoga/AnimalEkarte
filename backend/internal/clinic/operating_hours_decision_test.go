package clinic

import (
	"errors"
	"fmt"
	"maps"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/operation"
)

// 共通医院時間の CAS・no-op・休診 set/unset 判定（emr-clinic-line-time-selected-contract-20261022 §2.2/§2.3）。

const ohMaxVersion uint64 = 1<<63 - 1

func ohConfigured(t *testing.T, version uint64, s OperatingHoursSettings) OperatingHoursState {
	t.Helper()
	st, err := ConfiguredOperatingHours(version, s)
	if err != nil {
		t.Fatalf("ConfiguredOperatingHours: %v", err)
	}
	return st
}

func TestConfiguredOperatingHoursVersionBounds(t *testing.T) {
	t.Parallel()
	s := ohMustSettings(t, ohFullFields())
	cases := []struct {
		name    string
		version uint64
		wantErr bool
	}{
		{name: "zero rejected", version: 0, wantErr: true},
		{name: "one accepted", version: 1},
		{name: "max bigint accepted", version: ohMaxVersion},
		{name: "max bigint + 1 rejected", version: ohMaxVersion + 1, wantErr: true},
		{name: "uint64 max rejected", version: ^uint64(0), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st, err := ConfiguredOperatingHours(tc.version, s)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %+v", st)
				}
				return
			}
			if err != nil || !st.Configured() || st.Version() != tc.version {
				t.Fatalf("state=%+v err=%v", st, err)
			}
		})
	}
	if _, err := ConfiguredOperatingHours(1, OperatingHoursSettings{}); err == nil {
		t.Fatal("a zero-value Settings must not become a configured state")
	}
}

func TestOperatingHoursStateAccessors(t *testing.T) {
	t.Parallel()
	un := UnconfiguredOperatingHours()
	if un.Configured() || un.Version() != 0 {
		t.Fatalf("unconfigured state: %+v", un)
	}
	if s, ok := un.Settings(); ok || s.Valid() {
		t.Fatal("unconfigured state must expose neither a Settings nor a valid zero value")
	}

	want := ohMustSettings(t, ohFullFields())
	st := ohConfigured(t, 7, want)
	got, ok := st.Settings()
	if !ok || !got.Equal(want) || st.Version() != 7 {
		t.Fatalf("configured state: ok=%v version=%d", ok, st.Version())
	}
	// 返却値は複製で、state を書き換えられない。
	cd := ohClosedDates(t, got)
	cd[0] = ClosedDate{}
	again, _ := st.Settings()
	if !again.Equal(want) {
		t.Fatal("state changed through a returned Settings")
	}
}

func TestDecideOperatingHoursInitialize(t *testing.T) {
	t.Parallel()
	existing := ohConfigured(t, 1, ohMustSettings(t, ohFullFields()))

	t.Run("on unconfigured succeeds with Changed=true", func(t *testing.T) {
		t.Parallel()
		w, err := DecideOperatingHoursInitialize(UnconfiguredOperatingHours(), ohFullFields())
		if err != nil || !w.Changed || !w.Settings.Valid() {
			t.Fatalf("write=%+v err=%v", w, err)
		}
		if !w.Settings.Equal(ohMustSettings(t, ohFullFields())) {
			t.Fatal("initialized Settings differ from the input")
		}
	})

	t.Run("on configured is already-configured", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursInitialize(existing, ohFullFields())
		if !errors.Is(err, ErrOperatingHoursAlreadyConfigured) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("already-configured wins over invalid input", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursInitialize(existing, OperatingHoursFields{})
		if !errors.Is(err, ErrOperatingHoursAlreadyConfigured) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("incomplete fields rejected", func(t *testing.T) {
		t.Parallel()
		f := ohFullFields()
		f.ClosedDates = nil
		_, err := DecideOperatingHoursInitialize(UnconfiguredOperatingHours(), f)
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestDecideOperatingHoursUpdate(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields())
	current := ohConfigured(t, 5, base)
	sameWeek := OperatingHoursFields{WeeklyHours: ohWeeklyPtr(ohWeek())}

	t.Run("unconfigured", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursUpdate(UnconfiguredOperatingHours(), 1, sameWeek)
		if !errors.Is(err, ErrOperatingHoursUnconfigured) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursUpdate(current, 4, sameWeek)
		if !errors.Is(err, ErrOperatingHoursVersionConflict) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("version is checked before the changes are validated", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursUpdate(current, 4, OperatingHoursFields{})
		if !errors.Is(err, ErrOperatingHoursVersionConflict) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("empty changes with a matching version are invalid input", func(t *testing.T) {
		t.Parallel()
		_, err := DecideOperatingHoursUpdate(current, 5, OperatingHoursFields{})
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("same values are a no-op after CAS", func(t *testing.T) {
		t.Parallel()
		w, err := DecideOperatingHoursUpdate(current, 5, OperatingHoursFields{
			WeeklyHours:           ohWeeklyPtr(ohWeek()),
			ClosedDates:           ohClosedPtr(ohClosedList()),
			NationalHolidayClosed: ohBoolPtr(true),
		})
		if err != nil || w.Changed || !w.Settings.Equal(base) {
			t.Fatalf("write=%+v err=%v", w, err)
		}
	})

	t.Run("a real change reports Changed=true", func(t *testing.T) {
		t.Parallel()
		w, err := DecideOperatingHoursUpdate(current, 5, OperatingHoursFields{NationalHolidayClosed: ohBoolPtr(false)})
		if err != nil || !w.Changed || ohNationalHolidayClosed(t, w.Settings) {
			t.Fatalf("write=%+v err=%v", w, err)
		}
		if again, _ := current.Settings(); !again.Equal(base) {
			t.Fatal("current state mutated")
		}
	})

	t.Run("max version is accepted", func(t *testing.T) {
		t.Parallel()
		maxState := ohConfigured(t, ohMaxVersion, base)
		if _, err := DecideOperatingHoursUpdate(maxState, ohMaxVersion, sameWeek); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDecideClosedDateSet(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields()) // 2026-01-12 棚卸し, 2026-03-01 ""
	current := ohConfigured(t, 3, base)

	t.Run("unconfigured does not generate other fields", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateSet(UnconfiguredOperatingHours(), 1, ohMustDate(t, "2026-02-01"), "x")
		if !errors.Is(err, ErrOperatingHoursUnconfigured) || w.Changed || w.Settings.Valid() {
			t.Fatalf("write=%+v err=%v", w, err)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := DecideClosedDateSet(current, 2, ohMustDate(t, "2026-02-01"), "x")
		if !errors.Is(err, ErrOperatingHoursVersionConflict) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("new date changes only closed_dates", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateSet(current, 3, ohMustDate(t, "2026-02-01"), "臨時")
		if err != nil || !w.Changed || len(ohClosedDates(t, w.Settings)) != 3 {
			t.Fatalf("write=%+v err=%v", w, err)
		}
		if len(ohWeeklyIntervals(t, w.Settings, 1)) != 2 || !ohNationalHolidayClosed(t, w.Settings) {
			t.Fatal("weekly_hours and national_holiday_closed must be kept")
		}
	})

	t.Run("same date same reason is a no-op after CAS", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateSet(current, 3, ohMustDate(t, "2026-01-12"), "棚卸し")
		if err != nil || w.Changed || !w.Settings.Equal(base) {
			t.Fatalf("write=%+v err=%v", w, err)
		}
	})

	t.Run("same date different reason is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := DecideClosedDateSet(current, 3, ohMustDate(t, "2026-01-12"), "別理由")
		if !errors.Is(err, ErrClosedDateReasonConflict) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestDecideClosedDateUnset(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields())
	current := ohConfigured(t, 3, base)

	t.Run("unconfigured", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateUnset(UnconfiguredOperatingHours(), 1, ohMustDate(t, "2026-01-12"))
		if !errors.Is(err, ErrOperatingHoursUnconfigured) || w.Changed || w.Settings.Valid() {
			t.Fatalf("write=%+v err=%v", w, err)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := DecideClosedDateUnset(current, 9, ohMustDate(t, "2026-01-12"))
		if !errors.Is(err, ErrOperatingHoursVersionConflict) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("missing date is a no-op after CAS", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateUnset(current, 3, ohMustDate(t, "2026-02-01"))
		if err != nil || w.Changed || !w.Settings.Equal(base) {
			t.Fatalf("write=%+v err=%v", w, err)
		}
	})

	t.Run("existing date is removed", func(t *testing.T) {
		t.Parallel()
		w, err := DecideClosedDateUnset(current, 3, ohMustDate(t, "2026-01-12"))
		if err != nil || !w.Changed {
			t.Fatalf("write=%+v err=%v", w, err)
		}
		if _, ok := ohClosedDateOn(t, w.Settings, ohMustDate(t, "2026-01-12")); ok {
			t.Fatal("date still closed")
		}
		if again, _ := current.Settings(); !again.Equal(base) {
			t.Fatal("current state mutated")
		}
	})
}

func TestDecideClosedDateRejectsZeroDateAsMisuse(t *testing.T) {
	t.Parallel()
	current := ohConfigured(t, 3, ohMustSettings(t, ohFullFields()))
	_, err := DecideClosedDateSet(current, 3, CalendarDate{}, "x")
	ohAssertMisuse(t, "set", err, errZeroCalendarDate)
	_, err = DecideClosedDateUnset(current, 3, CalendarDate{})
	ohAssertMisuse(t, "unset", err, errZeroCalendarDate)
}

func TestOperatingHoursRejection(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want operation.Status
		ok   bool
	}{
		{name: "already configured", err: ErrOperatingHoursAlreadyConfigured, want: operation.StatusVersionConflict, ok: true},
		{name: "version conflict", err: ErrOperatingHoursVersionConflict, want: operation.StatusVersionConflict, ok: true},
		{name: "unconfigured is a missing owner target (validation)", err: ErrOperatingHoursUnconfigured, want: operation.StatusValidation, ok: true},
		{name: "closed date reason conflict", err: ErrClosedDateReasonConflict, want: operation.StatusValidation, ok: true},
		{name: "invalid input", err: apperrors.WrapInvalidInput("bad"), want: operation.StatusValidation, ok: true},
		{name: "wrapped already configured", err: fmt.Errorf("apply: %w", ErrOperatingHoursAlreadyConfigured), want: operation.StatusVersionConflict, ok: true},
		{name: "wrapped version conflict", err: fmt.Errorf("apply: %w", ErrOperatingHoursVersionConflict), want: operation.StatusVersionConflict, ok: true},
		{name: "wrapped unconfigured", err: fmt.Errorf("apply: %w", ErrOperatingHoursUnconfigured), want: operation.StatusValidation, ok: true},
		{name: "wrapped reason conflict", err: fmt.Errorf("apply: %w", ErrClosedDateReasonConflict), want: operation.StatusValidation, ok: true},
		{name: "wrapped invalid input", err: fmt.Errorf("apply: %w", apperrors.WrapInvalidInput("bad")), want: operation.StatusValidation, ok: true},
		{name: "unknown error is not a definite rejection", err: errors.New("db down")},
		{name: "invalid-settings misuse is not a rejection", err: errInvalidOperatingHoursSettings},
		{name: "wrapped invalid-settings misuse is not a rejection", err: fmt.Errorf("apply: %w", errInvalidOperatingHoursSettings)},
		{name: "zero calendar date misuse is not a rejection", err: errZeroCalendarDate},
		{name: "invalid weekday misuse is not a rejection", err: errInvalidWeekday},
		{name: "nil", err: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := OperatingHoursRejection(tc.err)
			if ok != tc.ok {
				t.Fatalf("ok=%v, want %v", ok, tc.ok)
			}
			if !tc.ok {
				if got != nil {
					t.Fatalf("got %+v with ok=false", got)
				}
				return
			}
			if got == nil || got.Status != tc.want {
				t.Fatalf("got %+v, want status %q", got, tc.want)
			}
		})
	}
}

func TestOperatingHoursReceiptMaps(t *testing.T) {
	t.Parallel()
	refs := []struct {
		clinicID uint64
		want     map[string]string
	}{
		{clinicID: 1, want: map[string]string{"clinic_id": "1"}},
		{clinicID: ohMaxVersion, want: map[string]string{"clinic_id": "9223372036854775807"}},
	}
	for _, tc := range refs {
		if got := OperatingHoursResultRefs(tc.clinicID); !maps.Equal(got, tc.want) {
			t.Errorf("refs(%d) = %v, want %v", tc.clinicID, got, tc.want)
		}
	}
	versions := []struct {
		version uint64
		want    map[string]string
	}{
		{version: 1, want: map[string]string{"clinic_operating_hours": "1"}},
		{version: 10, want: map[string]string{"clinic_operating_hours": "10"}},
		{version: ohMaxVersion, want: map[string]string{"clinic_operating_hours": "9223372036854775807"}},
	}
	for _, tc := range versions {
		if got := OperatingHoursCommittedVersions(tc.version); !maps.Equal(got, tc.want) {
			t.Errorf("versions(%d) = %v, want %v", tc.version, got, tc.want)
		}
	}
}
