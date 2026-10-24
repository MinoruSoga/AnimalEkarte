package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ================================================================
// buildStaffSlotInputsFromWindow — 事前取得済みマップからの純関数構築 (G7-1)
// ================================================================

func TestBuildStaffSlotInputsFromWindow(t *testing.T) {
	startStr := "09:00:00"
	endStr := "12:00:00"
	staffs := []model.Staff{{ID: 1}, {ID: 2}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	dateStr := "2026-08-01"

	t.Run("シフトエントリがあるスタッフはScheduleOverrideを持つ", func(t *testing.T) {
		entry := &model.ShiftEntry{ID: 100, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, rsvShiftDate, startStr), EndsAt: rsvTS(t, rsvShiftDate, endStr)}
		entriesByDateStaff := map[string]map[uint64][]*model.ShiftEntry{
			dateStr: {1: {entry}},
		}
		breaksByEntry := map[uint64][]model.ShiftEntryBreak{
			100: {{StartsAt: rsvTS(t, rsvShiftDate, "10:00:00"), EndsAt: rsvTS(t, rsvShiftDate, "10:15:00")}},
		}

		inputs := buildStaffSlotInputsFromWindow(staffs, date, entriesByDateStaff, breaksByEntry, nil)

		require.Len(t, inputs, 2)
		require.NotNil(t, inputs[0].ScheduleOverride)
		assert.Equal(t, "full", inputs[0].ScheduleOverride.ShiftType)
		assert.Equal(t, "0900", *inputs[0].ScheduleOverride.WorkStart)
		assert.Equal(t, "1200", *inputs[0].ScheduleOverride.WorkEnd)
		require.Len(t, inputs[0].ScheduleOverride.Breaks, 1)
		assert.Equal(t, "1000", inputs[0].ScheduleOverride.Breaks[0].Start)
		assert.Equal(t, "1015", inputs[0].ScheduleOverride.Breaks[0].End)
		// staff 2 にはエントリが無い(該当日のmapにキーが無い) → override なし
		assert.Nil(t, inputs[1].ScheduleOverride)
	})

	t.Run("該当日のエントリが無いスタッフはScheduleOverrideがnil", func(t *testing.T) {
		inputs := buildStaffSlotInputsFromWindow(staffs, date, map[string]map[uint64][]*model.ShiftEntry{}, nil, nil)
		require.Len(t, inputs, 2)
		assert.Nil(t, inputs[0].ScheduleOverride)
		assert.Nil(t, inputs[1].ScheduleOverride)
	})

	t.Run("区間不明の休憩を持つスタッフはその日の候補から外す（休憩なしとみなさない）", func(t *testing.T) {
		entry := &model.ShiftEntry{ID: 300, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, rsvShiftDate, startStr), EndsAt: rsvTS(t, rsvShiftDate, endStr)}
		entriesByDateStaff := map[string]map[uint64][]*model.ShiftEntry{dateStr: {1: {entry}}}
		breaksByEntry := map[uint64][]model.ShiftEntryBreak{300: {{StartsAt: nil, EndsAt: nil}}}

		inputs := buildStaffSlotInputsFromWindow(staffs, date, entriesByDateStaff, breaksByEntry, nil)

		require.Len(t, inputs, 1, "staff 1 は区間不明の休憩のため候補外、staff 2 は残る")
		assert.Equal(t, uint64(2), inputs[0].StaffID)
	})

	t.Run("当日の既存予約はキャンセル済みを除外し担当スタッフでフィルタする", func(t *testing.T) {
		doctorID1 := uint64(1)
		start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
		end := time.Date(2026, 8, 1, 10, 30, 0, 0, time.UTC)
		reservationsByDate := map[string][]model.Reservation{
			dateStr: {
				{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
				{DoctorID: &doctorID1, Status: model.ReservationStatusCancelled, StartTime: start, EndTime: end},
			},
		}

		inputs := buildStaffSlotInputsFromWindow(staffs, date, nil, nil, reservationsByDate)

		require.Len(t, inputs, 2)
		require.Len(t, inputs[0].ExistingResvs, 1, "staff 1: 確定予約1件のみ(キャンセル済みは除外)")
		assert.Equal(t, "1000", inputs[0].ExistingResvs[0].StartTime)
		assert.Empty(t, inputs[1].ExistingResvs, "staff 2: 担当外なのでゼロ件")
	})

	t.Run("別日の予約・エントリは混入しない", func(t *testing.T) {
		otherDateStr := "2026-08-02"
		doctorID1 := uint64(1)
		otherStart := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
		entry := &model.ShiftEntry{ID: 200, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, rsvShiftDate, startStr), EndsAt: rsvTS(t, rsvShiftDate, endStr)}
		entriesByDateStaff := map[string]map[uint64][]*model.ShiftEntry{otherDateStr: {1: {entry}}}
		reservationsByDate := map[string][]model.Reservation{
			otherDateStr: {{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed, StartTime: otherStart, EndTime: otherStart}},
		}

		inputs := buildStaffSlotInputsFromWindow(staffs, date, entriesByDateStaff, nil, reservationsByDate)

		require.Len(t, inputs, 2)
		assert.Nil(t, inputs[0].ScheduleOverride)
		assert.Empty(t, inputs[0].ExistingResvs)
	})
}

// ================================================================
// buildAvailableDatesStaffInputsFn — 期間プリフェッチ (G7-1)
// ================================================================

func TestGetAvailableDates_BuildAvailableDatesStaffInputsFn(t *testing.T) {
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}}
	settings := AvailableDatesSettings{BookingWindowMinDays: 0, BookingWindowMaxDays: 1}

	t.Run("シフト・休憩・予約を1クエリずつプリフェッチし日付ごとに正しく振り分ける", func(t *testing.T) {
		minDate, maxDate := BookingWindowDates(settings)
		startStr := "09:00:00"
		endStr := "17:00:00"
		entry := model.ShiftEntry{ID: 55, StaffID: 1, Date: minDate, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, minDate, startStr), EndsAt: rsvTS(t, minDate, endStr)}

		var capturedStaffIDs []uint64
		var capturedFrom, capturedTo time.Time
		scheduleRepo := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error) {
				capturedStaffIDs = staffIDs
				capturedFrom, capturedTo = from, to
				return []model.ShiftEntry{entry}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				require.Equal(t, []uint64{55}, entryIDs)
				return map[uint64][]model.ShiftEntryBreak{55: {{StartsAt: rsvTS(t, minDate, "12:00:00"), EndsAt: rsvTS(t, minDate, "13:00:00")}}}, nil
			},
		}
		doctorID1 := uint64(1)
		resvStart := time.Date(minDate.Year(), minDate.Month(), minDate.Day(), 10, 0, 0, 0, config.JST)
		adminRepo := &mockLiffAdminRepository{
			findTimeRangesByDateRangeFn: func(_ context.Context, _ uint64, from, to time.Time) ([]model.Reservation, error) {
				assert.True(t, from.Equal(minDate))
				assert.True(t, to.Equal(maxDate.AddDate(0, 0, 1)))
				return []model.Reservation{{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed, StartTime: resvStart, EndTime: resvStart.Add(30 * time.Minute)}}, nil
			},
		}

		svc := &liffService{scheduleRepo: scheduleRepo, adminRepo: adminRepo}
		fn, err := svc.buildAvailableDatesStaffInputsFn(ctx, 1, staffs, settings)
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, capturedStaffIDs)
		assert.True(t, capturedFrom.Equal(minDate))
		assert.True(t, capturedTo.Equal(maxDate.AddDate(0, 0, 1)))

		inputs, err := fn(ctx, minDate, 0, 0)
		require.NoError(t, err)
		require.Len(t, inputs, 1)
		require.NotNil(t, inputs[0].ScheduleOverride)
		assert.Equal(t, "0900", *inputs[0].ScheduleOverride.WorkStart)
		require.Len(t, inputs[0].ScheduleOverride.Breaks, 1)
		require.Len(t, inputs[0].ExistingResvs, 1)

		// 翌日は entry も予約も無いのでゼロ値
		nextDayInputs, err := fn(ctx, minDate.AddDate(0, 0, 1), 0, 0)
		require.NoError(t, err)
		require.Len(t, nextDayInputs, 1)
		assert.Nil(t, nextDayInputs[0].ScheduleOverride)
		assert.Empty(t, nextDayInputs[0].ExistingResvs)
	})

	t.Run("シフトエントリ取得失敗は非致命的(override無しで続行)", func(t *testing.T) {
		scheduleRepo := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return nil, errors.New("db error")
			},
		}
		adminRepo := &mockLiffAdminRepository{}
		svc := &liffService{scheduleRepo: scheduleRepo, adminRepo: adminRepo}

		fn, err := svc.buildAvailableDatesStaffInputsFn(ctx, 1, staffs, settings)
		require.NoError(t, err, "シフト取得失敗はエラーを伝播しない")

		minDate, _ := BookingWindowDates(settings)
		inputs, err := fn(ctx, minDate, 0, 0)
		require.NoError(t, err)
		require.Len(t, inputs, 1)
		assert.Nil(t, inputs[0].ScheduleOverride)
	})

	t.Run("休憩取得失敗は非致命的(Breaksが空のまま続行)", func(t *testing.T) {
		minDate, _ := BookingWindowDates(settings)
		startStr := "09:00:00"
		endStr := "17:00:00"
		entry := model.ShiftEntry{ID: 77, StaffID: 1, Date: minDate, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, rsvShiftDate, startStr), EndsAt: rsvTS(t, rsvShiftDate, endStr)}
		scheduleRepo := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{entry}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				return nil, errors.New("db error")
			},
		}
		adminRepo := &mockLiffAdminRepository{}
		svc := &liffService{scheduleRepo: scheduleRepo, adminRepo: adminRepo}

		fn, err := svc.buildAvailableDatesStaffInputsFn(ctx, 1, staffs, settings)
		require.NoError(t, err)

		inputs, err := fn(ctx, minDate, 0, 0)
		require.NoError(t, err)
		require.Len(t, inputs, 1)
		require.NotNil(t, inputs[0].ScheduleOverride, "エントリ自体は取得できているのでoverrideは設定される")
		assert.Empty(t, inputs[0].ScheduleOverride.Breaks)
	})

	t.Run("当日予約プリフェッチ失敗は致命的(エラーを伝播する)", func(t *testing.T) {
		scheduleRepo := &mockLiffScheduleRepository{}
		adminRepo := &mockLiffAdminRepository{
			findTimeRangesByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]model.Reservation, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &liffService{scheduleRepo: scheduleRepo, adminRepo: adminRepo}

		fn, err := svc.buildAvailableDatesStaffInputsFn(ctx, 1, staffs, settings)
		require.Error(t, err)
		assert.Nil(t, fn)
	})
}

// ================================================================
// GetAvailableDates — 職種ガードのバッチ化 (G7-1)
// ================================================================

func TestGetAvailableDates_OccupationGuardUsesBatchedCounts(t *testing.T) {
	ctx := context.Background()
	staff := model.Staff{ID: 7, IsActive: true, ReservationVisible: true}

	settingRepo := &mockLiffSettingRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				BookingWindowMinDays:    0,
				BookingWindowMaxDays:    1,
				TimeSlotIntervalMinutes: 60,
				TimeSlotMode:            "minimize_gaps",
			}, nil
		},
	}
	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                   1,
				IsActive:             true,
				ReservationVisible:   true,
				DurationMinutes:      60,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}
	staffRepo := &mockLiffStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) { return &staff, nil },
	}

	var occupationGuardCallCount int
	var capturedDates []time.Time
	occupationRepo := &mockReservationTypeOccupationRepository{
		findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeOccupation, error) {
			return []model.ReservationTypeOccupation{{ID: 1}}, nil
		},
		countByStaffIDsFn: func(_ context.Context, _, _ uint64, dates []time.Time) (map[string]int64, error) {
			occupationGuardCallCount++
			capturedDates = dates
			result := make(map[string]int64, len(dates))
			for _, d := range dates {
				result[d.Format("2006-01-02")] = 0 // 誰も出勤していない扱い
			}
			return result, nil
		},
	}

	svc := &liffService{
		settingRepo:    settingRepo,
		typeLiffRepo:   typeRepo,
		staffRepo:      staffRepo,
		scheduleRepo:   &mockLiffScheduleRepository{},
		adminRepo:      &mockLiffAdminRepository{},
		occupationRepo: occupationRepo,
	}

	results, _, err := svc.GetAvailableDates(ctx, 1, 1, staff.ID)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	assert.Equal(t, 1, occupationGuardCallCount, "職種ガードは日毎ではなく1回のバッチ呼出であるべき(G7-1)")
	assert.NotEmpty(t, capturedDates)
	for _, r := range results {
		assert.False(t, r.Available, "全日 staff_off として除外される(count=0)")
		assert.Equal(t, "staff_off", r.Reason)
	}
}

func TestGetAvailableDates_OccupationGuardSkipsCountWhenNoOccupations(t *testing.T) {
	ctx := context.Background()
	staff := model.Staff{ID: 7, IsActive: true, ReservationVisible: true}

	settingRepo := &mockLiffSettingRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				BookingWindowMinDays:    0,
				BookingWindowMaxDays:    1,
				TimeSlotIntervalMinutes: 60,
				TimeSlotMode:            "minimize_gaps",
			}, nil
		},
	}
	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                   1,
				IsActive:             true,
				ReservationVisible:   true,
				DurationMinutes:      60,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}
	staffRepo := &mockLiffStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) { return &staff, nil },
	}

	var occupationGuardCallCount int
	occupationRepo := &mockReservationTypeOccupationRepository{
		findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeOccupation, error) {
			return []model.ReservationTypeOccupation{}, nil
		},
		countByStaffIDsFn: func(_ context.Context, _, _ uint64, _ []time.Time) (map[string]int64, error) {
			occupationGuardCallCount++
			return map[string]int64{}, nil
		},
	}

	svc := &liffService{
		settingRepo:    settingRepo,
		typeLiffRepo:   typeRepo,
		staffRepo:      staffRepo,
		scheduleRepo:   &mockLiffScheduleRepository{},
		adminRepo:      &mockLiffAdminRepository{},
		occupationRepo: occupationRepo,
	}

	results, _, err := svc.GetAvailableDates(ctx, 1, 1, staff.ID)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Zero(t, occupationGuardCallCount, "職種紐付け0件なら Count を呼ばない")

	hasAvailable := false
	for _, r := range results {
		if r.Available {
			hasAvailable = true
			break
		}
	}
	assert.True(t, hasAvailable, "職種紐付けなしはガードをスキップし少なくとも1日は Available")
}

func TestLiffService_TypeScopedPublicReads_RejectInactiveReservationType(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	ctx := context.Background()
	date := time.Date(2026, 8, 3, 0, 0, 0, 0, config.JST)
	downstreamErr := errors.New("downstream reached")

	typeRepo := func(isActive bool) *mockLiffTypeRepository {
		return &mockLiffTypeRepository{
			findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, typeID, gotTypeID)
				return &model.ReservationType{
					ID:                   gotTypeID,
					ClinicID:             gotClinicID,
					IsActive:             isActive,
					ReservationVisible:   true,
					DurationMinutes:      30,
					ReservationDayOption: model.DayOptionAnyday,
				}, nil
			},
		}
	}
	settingRepo := func() *mockLiffSettingRepository {
		return &mockLiffSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return liffDefaultSetting(), nil
			},
		}
	}

	t.Run("GetStaffs", func(t *testing.T) {
		downstreamCalls := 0
		staffRepo := &mockLiffStaffRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				downstreamCalls++
				return nil, downstreamErr
			},
		}

		inactiveSvc := &liffService{typeLiffRepo: typeRepo(false), staffRepo: staffRepo}
		_, err := inactiveSvc.GetStaffs(ctx, clinicID, typeID)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), "reservation type is not available for LINE reservation")
		assert.Zero(t, downstreamCalls, "inactive type must short-circuit staff lookup")

		activeSvc := &liffService{typeLiffRepo: typeRepo(true), staffRepo: staffRepo}
		_, err = activeSvc.GetStaffs(ctx, clinicID, typeID)
		require.ErrorIs(t, err, downstreamErr)
		assert.Equal(t, 1, downstreamCalls, "active public type must reach staff lookup")
	})

	t.Run("GetAvailableDates", func(t *testing.T) {
		downstreamCalls := 0
		staffRepo := &mockLiffStaffRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				downstreamCalls++
				return nil, downstreamErr
			},
		}

		inactiveSvc := &liffService{
			settingRepo:  settingRepo(),
			typeLiffRepo: typeRepo(false),
			staffRepo:    staffRepo,
		}
		_, _, err := inactiveSvc.GetAvailableDates(ctx, clinicID, typeID, 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), "reservation type is not available for LINE reservation")
		assert.Zero(t, downstreamCalls, "inactive type must short-circuit date availability dependencies")

		activeSvc := &liffService{
			settingRepo:  settingRepo(),
			typeLiffRepo: typeRepo(true),
			staffRepo:    staffRepo,
		}
		_, _, err = activeSvc.GetAvailableDates(ctx, clinicID, typeID, 0)
		require.ErrorIs(t, err, downstreamErr)
		assert.Equal(t, 1, downstreamCalls, "active public type must reach date availability dependencies")
	})

	t.Run("GetAvailableTimes", func(t *testing.T) {
		downstreamCalls := 0
		staffRepo := &mockLiffStaffRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				downstreamCalls++
				return nil, downstreamErr
			},
		}

		inactiveSvc := &liffService{
			settingRepo:  settingRepo(),
			typeLiffRepo: typeRepo(false),
			staffRepo:    staffRepo,
		}
		_, err := inactiveSvc.GetAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), "reservation type is not available for LINE reservation")
		assert.Zero(t, downstreamCalls, "inactive type must short-circuit time availability dependencies")

		activeSvc := &liffService{
			settingRepo:  settingRepo(),
			typeLiffRepo: typeRepo(true),
			staffRepo:    staffRepo,
		}
		_, err = activeSvc.GetAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.ErrorIs(t, err, downstreamErr)
		assert.Equal(t, 1, downstreamCalls, "active public type must reach time availability dependencies")
	})
}

// BUG-015: staff available-times may proceed for inactive types; LIFF GetAvailableTimes stays reject.
func TestLiffService_GetStaffAvailableTimes_AllowsInactiveReservationType(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	ctx := context.Background()
	date := time.Date(2026, 8, 3, 0, 0, 0, 0, config.JST)
	downstreamErr := errors.New("downstream reached")

	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, typeID, gotTypeID)
			return &model.ReservationType{
				ID:                   gotTypeID,
				ClinicID:             gotClinicID,
				IsActive:             false,
				ReservationVisible:   true,
				DurationMinutes:      30,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}
	settingRepo := &mockLiffSettingRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return liffDefaultSetting(), nil
		},
	}
	downstreamCalls := 0
	staffRepo := &mockLiffStaffRepository{
		findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
			downstreamCalls++
			return nil, downstreamErr
		},
	}
	svc := &liffService{
		settingRepo:  settingRepo,
		typeLiffRepo: typeRepo,
		staffRepo:    staffRepo,
	}

	_, err := svc.GetAvailableTimes(ctx, clinicID, typeID, 0, date)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "LIFF GetAvailableTimes must keep inactive short-circuit")
	assert.Contains(t, err.Error(), "reservation type is not available for LINE reservation")
	assert.Zero(t, downstreamCalls)

	_, err = svc.GetStaffAvailableTimes(ctx, clinicID, typeID, 0, date)
	require.Error(t, err)
	assert.False(t, apperrors.IsInvalidInput(err), "staff available-times must not reject inactive reservation type")
	assert.ErrorIs(t, err, downstreamErr)
	assert.Equal(t, 1, downstreamCalls, "inactive staff path must reach slot dependencies")
}

// Hidden LIFF types (internal / reservation-invisible) share LINE create's invalid-input
// so existence is not confirmed by a distinct error. GetStaffs / GetAvailableDates /
// GetAvailableTimes must short-circuit before staff/date/time work.
func TestLiffService_TypeScopedPublicReads_RejectHiddenReservationType(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	ctx := context.Background()
	date := time.Date(2026, 8, 3, 0, 0, 0, 0, config.JST)
	const lineUnavailableMsg = "reservation type is not available for LINE reservation"
	downstreamErr := errors.New("downstream reached")

	tests := []struct {
		name               string
		isInternal         bool
		reservationVisible bool
		call               func(svc *liffService) error
	}{
		{
			name:               "GetStaffs: active internal type",
			isInternal:         true,
			reservationVisible: true,
			call: func(svc *liffService) error {
				_, err := svc.GetStaffs(ctx, clinicID, typeID)
				return err
			},
		},
		{
			name:               "GetStaffs: active reservation-invisible type",
			isInternal:         false,
			reservationVisible: false,
			call: func(svc *liffService) error {
				_, err := svc.GetStaffs(ctx, clinicID, typeID)
				return err
			},
		},
		{
			name:               "GetAvailableDates: internal type",
			isInternal:         true,
			reservationVisible: true,
			call: func(svc *liffService) error {
				_, _, err := svc.GetAvailableDates(ctx, clinicID, typeID, 0)
				return err
			},
		},
		{
			name:               "GetAvailableTimes: reservation-invisible active type",
			isInternal:         false,
			reservationVisible: false,
			call: func(svc *liffService) error {
				_, err := svc.GetAvailableTimes(ctx, clinicID, typeID, 0, date)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downstreamCalls := 0
			staffRepo := &mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					downstreamCalls++
					return nil, downstreamErr
				},
			}
			svc := &liffService{
				settingRepo: &mockLiffSettingRepository{
					findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
						return liffDefaultSetting(), nil
					},
				},
				typeLiffRepo: &mockLiffTypeRepository{
					findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
						assert.Equal(t, clinicID, gotClinicID)
						assert.Equal(t, typeID, gotTypeID)
						return &model.ReservationType{
							ID:                   gotTypeID,
							ClinicID:             gotClinicID,
							IsActive:             true,
							IsInternal:           tt.isInternal,
							ReservationVisible:   tt.reservationVisible,
							DurationMinutes:      30,
							ReservationDayOption: model.DayOptionAnyday,
						}, nil
					},
				},
				staffRepo: staffRepo,
			}

			err := tt.call(svc)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Contains(t, err.Error(), lineUnavailableMsg)
			assert.NotContains(t, err.Error(), "reservation type is inactive")
			assert.Zero(t, downstreamCalls, "hidden type must not reach staff/date/time dependencies")
		})
	}
}

func TestLiffService_GetAvailableDates_LegacyClosedWeekdays(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	ctx := context.Background()
	setting := liffDefaultSetting()
	setting.BookingWindowMinDays = 0
	setting.BookingWindowMaxDays = 0
	setting.ClosedWeekdays = []byte(`["0","1","2","3","4","5","6"]`)
	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                   gotTypeID,
				ClinicID:             gotClinicID,
				IsActive:             true,
				ReservationVisible:   true,
				DurationMinutes:      30,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}
	svc := newLiffSvc(
		&mockLiffSettingRepository{
			findByClinicIDFn: func(_ context.Context, gotClinicID uint64) (*model.LineReservationSetting, error) {
				assert.Equal(t, clinicID, gotClinicID)
				return setting, nil
			},
		},
		typeRepo,
		&mockLiffStaffRepository{},
		&mockLiffScheduleRepository{},
		&mockLiffAdminRepository{},
		&mockLiffCustomerRepository{},
		&mockLiffValidators{},
		&mockLiffNotifier{},
	)

	results, window, err := svc.GetAvailableDates(ctx, clinicID, typeID, 0)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, window.Start, results[0].Date)
	assert.False(t, results[0].Available)
	assert.Equal(t, "closed", results[0].Reason)
}

func TestLiffService_GetStaffAvailableTimes_LegacyWeekdaysAndBusinessHoursFallback(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	const staffID = uint64(11)
	ctx := context.Background()
	openDate := time.Date(2026, 9, 13, 0, 0, 0, 0, config.JST)
	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                   gotTypeID,
				ClinicID:             gotClinicID,
				IsActive:             true,
				ReservationVisible:   true,
				DurationMinutes:      30,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}
	staffRepo := &mockLiffStaffRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotStaffID uint64) (*model.Staff, error) {
			return &model.Staff{ID: gotStaffID, ClinicID: gotClinicID, IsActive: true, ReservationVisible: true}, nil
		},
	}
	tests := []struct {
		name          string
		businessHours []byte
		want          []TimeSlot
		wantTimeError bool
	}{
		{
			name:          "valid business hours produce expected slots",
			businessHours: []byte(`{"start":"0900","end":"1000"}`),
			want:          []TimeSlot{{StartTime: "0900", EndTime: "0930"}, {StartTime: "0930", EndTime: "1000"}},
		},
		{name: "empty business hours keep default hours", businessHours: []byte{}},
		{name: "malformed business hours keep default hours", businessHours: []byte(`not-json`)},
		{name: "empty start uses existing time validation", businessHours: []byte(`{"start":"","end":"1000"}`), wantTimeError: true},
		{name: "empty end uses existing time validation", businessHours: []byte(`{"start":"0900","end":""}`), wantTimeError: true},
		{name: "malformed start uses existing time validation", businessHours: []byte(`{"start":"09:00","end":"1000"}`), wantTimeError: true},
		{name: "malformed end uses existing time validation", businessHours: []byte(`{"start":"0900","end":"09x0"}`), wantTimeError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := liffDefaultSetting()
			setting.ClosedWeekdays = []byte(`["1"]`)
			setting.BusinessHours = tt.businessHours
			setting.BreakHours = []byte(`[]`)
			setting.TimeSlotIntervalMinutes = 30
			setting.TimeSlotMode = "allow_gaps"
			svc := newLiffSvc(
				&mockLiffSettingRepository{
					findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
						return setting, nil
					},
				},
				typeRepo,
				staffRepo,
				&mockLiffScheduleRepository{},
				&mockLiffAdminRepository{},
				&mockLiffCustomerRepository{},
				&mockLiffValidators{},
				&mockLiffNotifier{},
			)

			slots, err := svc.GetStaffAvailableTimes(ctx, clinicID, typeID, staffID, openDate)

			if tt.wantTimeError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid time format")
				assert.NotContains(t, err.Error(), "休診設定の形式が正しくありません")
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, slots)
			assert.Equal(t, "0900", slots[0].StartTime)
			if tt.want != nil {
				assert.Equal(t, tt.want, slots)
			}
		})
	}
}

// BUG-RES-AVAILABLE-TIMES-404: in-clinic settings-missing is identifiable unset;
// LIFF keeps settings-required; empty holiday stays []; type not-found and DB errors stay errors.
func TestGetAvailableTimes_SettingsMissing_Contracts(t *testing.T) {
	const clinicID = uint64(3)
	const typeID = uint64(7)
	ctx := context.Background()
	date := time.Date(2026, 9, 13, 0, 0, 0, 0, config.JST)

	typeRepoOK := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotTypeID uint64) (*model.ReservationType, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, typeID, gotTypeID)
			return &model.ReservationType{
				ID:                   gotTypeID,
				ClinicID:             gotClinicID,
				IsActive:             true,
				ReservationVisible:   true,
				DurationMinutes:      30,
				ReservationDayOption: model.DayOptionAnyday,
			}, nil
		},
	}

	t.Run("staff path: settings not-found is identifiable unset", func(t *testing.T) {
		svc := &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return nil, apperrors.FromGORM(gorm.ErrRecordNotFound, "line_reservation_setting", "clinic")
				},
			},
			typeLiffRepo: typeRepoOK,
			staffRepo: &mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					t.Fatal("staff lookup must not run when settings are unset")
					return nil, nil
				},
			},
		}
		slots, err := svc.GetStaffAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.Error(t, err)
		assert.Nil(t, slots)
		assert.True(t, IsLineReservationSettingsUnset(err), "want unset error, got %v", err)
		assert.False(t, apperrors.IsNotFound(err), "unset must not be generic not-found")
	})

	t.Run("LIFF path: settings not-found stays settings-required not-found", func(t *testing.T) {
		svc := &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return nil, apperrors.FromGORM(gorm.ErrRecordNotFound, "line_reservation_setting", "clinic")
				},
			},
			typeLiffRepo: typeRepoOK,
		}
		slots, err := svc.GetAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.Error(t, err)
		assert.Nil(t, slots)
		assert.False(t, IsLineReservationSettingsUnset(err), "LIFF must not get in-clinic unset exception")
		assert.True(t, apperrors.IsNotFound(err), "LIFF settings missing stays not-found")
	})

	t.Run("staff path: settings present closed date returns empty slots", func(t *testing.T) {
		closedDate := time.Date(2026, 9, 14, 0, 0, 0, 0, config.JST) // Monday
		svc := &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					s := liffDefaultSetting()
					s.ClosedWeekdays = []byte(`["1"]`)
					return s, nil
				},
			},
			typeLiffRepo: typeRepoOK,
			staffRepo: &mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					t.Fatal("closed date must return before staff resolution")
					return nil, nil
				},
			},
		}
		slots, err := svc.GetStaffAvailableTimes(ctx, clinicID, typeID, 0, closedDate)
		require.NoError(t, err)
		assert.Empty(t, slots)
		assert.False(t, IsLineReservationSettingsUnset(err))
	})

	t.Run("staff path: type not-found stays not-found not unset", func(t *testing.T) {
		svc := &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			typeLiffRepo: &mockLiffTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
					return nil, apperrors.FromGORM(gorm.ErrRecordNotFound, "reservation_type", "7")
				},
			},
		}
		slots, err := svc.GetStaffAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.Error(t, err)
		assert.Nil(t, slots)
		assert.False(t, IsLineReservationSettingsUnset(err))
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("staff path: settings DB error stays non-unset error", func(t *testing.T) {
		dbErr := errors.New("db connection lost")
		svc := &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return nil, dbErr
				},
			},
			typeLiffRepo: typeRepoOK,
		}
		slots, err := svc.GetStaffAvailableTimes(ctx, clinicID, typeID, 0, date)
		require.Error(t, err)
		assert.Nil(t, slots)
		assert.False(t, IsLineReservationSettingsUnset(err))
		assert.ErrorIs(t, err, dbErr)
		assert.False(t, apperrors.IsNotFound(err))
	})
}
