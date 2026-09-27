package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ================================================================
// GenerateSlotCapacities（エンジン）
// ================================================================

func TestGenerateSlotCapacities_SingleStaff(t *testing.T) {
	input := &TimeSlotsInput{
		BusinessHours:  BusinessHours{Start: "0900", End: "1200"},
		CourseDuration: 60,
		Mode:           "minimize_gaps",
		Staffs: []StaffSlotInput{{
			StaffID: 1,
			ExistingResvs: []ExistingReservation{
				{StaffID: 1, StartTime: "1000", EndTime: "1100"},
			},
		}},
	}

	caps, err := GenerateSlotCapacities(input)
	require.NoError(t, err)
	require.Len(t, caps, 3)

	assert.Equal(t, "0900", caps[0].StartTime)
	assert.Equal(t, 1, caps[0].FreeStaffCount)
	assert.True(t, caps[0].StaffBound)

	// 10:00 は予約占有で全スタッフ不可 → 満員枠として残る
	assert.Equal(t, "1000", caps[1].StartTime)
	assert.Equal(t, 0, caps[1].FreeStaffCount)
	assert.True(t, caps[1].StaffBound)

	assert.Equal(t, "1100", caps[2].StartTime)
	assert.Equal(t, 1, caps[2].FreeStaffCount)
}

func TestGenerateSlotCapacities_MultiStaff(t *testing.T) {
	input := &TimeSlotsInput{
		BusinessHours:  BusinessHours{Start: "0900", End: "1200"},
		CourseDuration: 60,
		Mode:           "minimize_gaps",
		Staffs: []StaffSlotInput{
			{
				StaffID: 1,
				ExistingResvs: []ExistingReservation{
					{StaffID: 1, StartTime: "1000", EndTime: "1100"},
				},
			},
			{StaffID: 2},
		},
	}

	caps, err := GenerateSlotCapacities(input)
	require.NoError(t, err)
	require.Len(t, caps, 3)

	freeCounts := map[string]int{}
	for _, c := range caps {
		freeCounts[c.StartTime] = c.FreeStaffCount
	}
	// 10:00 はスタッフ1のみ埋まっている → スタッフ2が残る（別トリマー同時予約の根拠）
	assert.Equal(t, map[string]int{"0900": 2, "1000": 1, "1100": 2}, freeCounts)
}

func TestGenerateSlotCapacities_IncludesFreeOnlyStart(t *testing.T) {
	// minimize_gaps で予約が区間をずらすと、free 側だけに現れる開始時刻が生じる。
	// potential [09:00,10:00] → 枠 {0900,0930}
	// free [09:10,10:00] → 枠 {0910}
	input := &TimeSlotsInput{
		BusinessHours:  BusinessHours{Start: "0900", End: "1000"},
		CourseDuration: 30,
		Mode:           "minimize_gaps",
		Staffs: []StaffSlotInput{{
			StaffID: 1,
			ExistingResvs: []ExistingReservation{
				{StaffID: 1, StartTime: "0900", EndTime: "0910"},
			},
		}},
	}

	caps, err := GenerateSlotCapacities(input)
	require.NoError(t, err)
	require.Len(t, caps, 3)

	assert.Equal(t, "0900", caps[0].StartTime)
	assert.Equal(t, 0, caps[0].FreeStaffCount)
	assert.Equal(t, "0910", caps[1].StartTime)
	assert.Equal(t, 1, caps[1].FreeStaffCount)
	assert.Equal(t, "0930", caps[2].StartTime)
	assert.Equal(t, 0, caps[2].FreeStaffCount)
}

func TestGenerateSlotCapacities_AllStaffBookedKeepsFullSlot(t *testing.T) {
	input := &TimeSlotsInput{
		BusinessHours:  BusinessHours{Start: "0900", End: "1100"},
		CourseDuration: 60,
		Mode:           "minimize_gaps",
		Staffs: []StaffSlotInput{
			{StaffID: 1, ExistingResvs: []ExistingReservation{{StaffID: 1, StartTime: "0900", EndTime: "1000"}}},
			{StaffID: 2, ExistingResvs: []ExistingReservation{{StaffID: 2, StartTime: "0900", EndTime: "1000"}}},
		},
	}

	caps, err := GenerateSlotCapacities(input)
	require.NoError(t, err)
	require.Len(t, caps, 2)
	// 全スタッフが埋まった 0900 枠は GenerateTimeSlots では消えるが、ここでは残る
	assert.Equal(t, "0900", caps[0].StartTime)
	assert.Equal(t, 0, caps[0].FreeStaffCount)
	assert.Equal(t, "1000", caps[1].StartTime)
	assert.Equal(t, 2, caps[1].FreeStaffCount)
}

// ================================================================
// mergeSlotCapacities（予約可能枠テーブル加算）
// ================================================================

func specificSlot(startTime string, date time.Time) model.ReservationTypeAvailableSlot {
	d := date
	return model.ReservationTypeAvailableSlot{
		AvailableType: model.AvailableSlotTypeSpecific,
		SpecificDate:  &d,
		StartTime:     startTime, // "HH:MM"
		IsActive:      true,
	}
}

func TestMergeSlotCapacities(t *testing.T) {
	date := time.Date(2026, 6, 2, 0, 0, 0, 0, config.JST)
	base := []SlotCapacity{
		{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 0, StaffBound: true},
		{TimeSlot: TimeSlot{StartTime: "1000", EndTime: "1100"}, FreeStaffCount: 2, StaffBound: true},
	}

	t.Run("adds whitelist-only slot as staff-unbound", func(t *testing.T) {
		merged := mergeSlotCapacities(base, []model.ReservationTypeAvailableSlot{
			specificSlot("20:30", date),
		}, date, 60)

		require.Len(t, merged, 3)
		assert.Equal(t, "2030", merged[2].StartTime)
		assert.Equal(t, "2130", merged[2].EndTime)
		assert.False(t, merged[2].StaffBound)
		assert.Equal(t, 0, merged[2].FreeStaffCount)
	})

	t.Run("whitelist start at staff-full slot revives it as unbound (legacy merge semantics)", func(t *testing.T) {
		merged := mergeSlotCapacities(base, []model.ReservationTypeAvailableSlot{
			specificSlot("09:00", date),
		}, date, 60)

		require.Len(t, merged, 2)
		assert.False(t, merged[0].StaffBound, "full grid slot claimed by whitelist becomes staff-unbound")
	})

	t.Run("whitelist start at free grid slot keeps staff binding", func(t *testing.T) {
		merged := mergeSlotCapacities(base, []model.ReservationTypeAvailableSlot{
			specificSlot("10:00", date),
		}, date, 60)

		require.Len(t, merged, 2)
		assert.True(t, merged[1].StaffBound)
		assert.Equal(t, 2, merged[1].FreeStaffCount)
	})

	t.Run("non-applicable or inactive slots are ignored", func(t *testing.T) {
		otherDate := time.Date(2026, 6, 3, 0, 0, 0, 0, config.JST)
		merged := mergeSlotCapacities(base, []model.ReservationTypeAvailableSlot{
			specificSlot("20:30", otherDate),
			{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &date, StartTime: "21:00", IsActive: false},
		}, date, 60)

		assert.Equal(t, base, merged)
	})
}

// ================================================================
// rateSlotCapacity（残数→ステータス判定）
// ================================================================

func TestRateSlotCapacity(t *testing.T) {
	max3 := 3
	cases := []struct {
		name          string
		slotCap       SlotCapacity
		maxConcurrent *int
		taken         map[string]int64
		wantStatus    SlotVacancyStatus
		wantRemaining *int
	}{
		{
			name:          "staff-bound slot with 3 free staffs is available",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 3, StaffBound: true},
			maxConcurrent: nil,
			wantStatus:    SlotVacancyAvailable,
			wantRemaining: ptrInt(3),
		},
		{
			name:          "1 free staff is low (△)",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 1, StaffBound: true},
			maxConcurrent: nil,
			wantStatus:    SlotVacancyLow,
			wantRemaining: ptrInt(1),
		},
		{
			name:          "0 free staff is full (✕)",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 0, StaffBound: true},
			maxConcurrent: nil,
			wantStatus:    SlotVacancyFull,
			wantRemaining: ptrInt(0),
		},
		{
			name:          "type capacity caps the staff dimension",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 3, StaffBound: true},
			maxConcurrent: &max3,
			taken:         map[string]int64{"0900": 2},
			wantStatus:    SlotVacancyLow,
			wantRemaining: ptrInt(1),
		},
		{
			name:          "type capacity reached makes the slot full",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 3, StaffBound: true},
			maxConcurrent: &max3,
			taken:         map[string]int64{"0900": 3},
			wantStatus:    SlotVacancyFull,
			wantRemaining: ptrInt(0),
		},
		{
			name:          "over-booked type capacity clamps to zero",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "0900", EndTime: "1000"}, FreeStaffCount: 3, StaffBound: true},
			maxConcurrent: &max3,
			taken:         map[string]int64{"0900": 5},
			wantStatus:    SlotVacancyFull,
			wantRemaining: ptrInt(0),
		},
		{
			name:          "staff-unbound slot without type capacity is unbounded (remaining null)",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "2030", EndTime: "2130"}, StaffBound: false},
			maxConcurrent: nil,
			wantStatus:    SlotVacancyAvailable,
			wantRemaining: nil,
		},
		{
			name:          "staff-unbound slot follows type capacity only",
			slotCap:       SlotCapacity{TimeSlot: TimeSlot{StartTime: "2030", EndTime: "2130"}, StaffBound: false},
			maxConcurrent: &max3,
			taken:         map[string]int64{"2030": 2},
			wantStatus:    SlotVacancyLow,
			wantRemaining: ptrInt(1),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := rateSlotCapacity(tt.slotCap, tt.maxConcurrent, tt.taken)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.slotCap.TimeSlot, got.TimeSlot)
			if tt.wantRemaining == nil {
				assert.Nil(t, got.Remaining)
			} else {
				require.NotNil(t, got.Remaining)
				assert.Equal(t, *tt.wantRemaining, *got.Remaining)
			}
		})
	}
}

func ptrInt(v int) *int { return &v }

// ================================================================
// GetStaffTimeSlotAvailabilities（サービス結合）
// ================================================================

type stubAvailableSlotReader struct {
	slots []model.ReservationTypeAvailableSlot
}

func (m stubAvailableSlotReader) FindAll(_ context.Context, _, _ uint64) ([]model.ReservationTypeAvailableSlot, error) {
	return m.slots, nil
}

func newAvailabilityTestSvc(
	course *model.ReservationType,
	dayReservations []model.Reservation,
	countByStart map[string]int64,
) *liffService {
	setting := &mockLiffSettingRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return liffDefaultSetting(), nil
		},
	}
	typeRepo := &mockLiffTypeRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			return course, nil
		},
	}
	staffRepo := &mockLiffStaffRepository{
		findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
			return []model.Staff{
				{ID: 1, IsActive: true, ReservationVisible: true},
				{ID: 2, IsActive: true, ReservationVisible: true},
			}, nil
		},
		findCapabilitiesByStaffIDsFn: func(_ context.Context, _ uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
			caps := make([]model.StaffReservationCapability, 0, len(staffIDs))
			for _, staffID := range staffIDs {
				caps = append(caps, model.StaffReservationCapability{StaffID: staffID, ReservationTypeID: course.ID})
			}
			return caps, nil
		},
	}
	adminRepo := &mockLiffAdminRepository{
		findTimeRangesByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]model.Reservation, error) {
			return dayReservations, nil
		},
	}
	reservationRepo := &mockLiffReservationRepository{
		countByTypeAndStartTimeFn: func(_ context.Context, _, _ uint64, startTime time.Time, _ *uint64) (int64, error) {
			return countByStart[startTime.In(config.JST).Format("1504")], nil
		},
	}
	svc := newLiffSvcWithDeps(
		setting,
		typeRepo,
		staffRepo,
		&mockLiffScheduleRepository{},
		adminRepo,
		reservationRepo,
		&mockLiffCustomerRepository{},
		&mockLiffOwnerRepository{},
		&mockLiffValidators{},
		&mockLiffNotifier{},
	)
	return svc
}

func TestGetStaffTimeSlotAvailabilities_StatusTransitions(t *testing.T) {
	// 営業 09:00-19:00 / 休憩 12:00-13:00 / dur=60 → 枠 {0900,1000,1100,1300..1800}
	// スタッフ1に 10:00-11:00 の予約 → 1000 の空きスタッフ数=1（△）。
	// 区分定員 max_concurrent=3 かつ 1100 に3件の予約 → 1100 は定員満員（✕）。
	maxConcurrent := 3
	doctorID1 := uint64(1)
	day := time.Date(2026, 6, 2, 0, 0, 0, 0, config.JST)
	svc := newAvailabilityTestSvc(
		&model.ReservationType{ID: 10, ClinicID: 3, DurationMinutes: 60, IsActive: true, ReservationVisible: true, MaxConcurrent: &maxConcurrent},
		[]model.Reservation{
			{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed,
				StartTime: time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST),
				EndTime:   time.Date(2026, 6, 2, 11, 0, 0, 0, config.JST)},
		},
		map[string]int64{"1100": 3},
	)

	got, err := svc.GetStaffTimeSlotAvailabilities(context.Background(), 3, 10, 0, day)
	require.NoError(t, err)
	require.Len(t, got, 9)

	byStart := map[string]TimeSlotAvailability{}
	for _, a := range got {
		byStart[a.StartTime] = a
	}
	// 0900: スタッフ2名空き・定員残3 → remaining=2 → 〇
	assert.Equal(t, SlotVacancyAvailable, byStart["0900"].Status)
	require.NotNil(t, byStart["0900"].Remaining)
	assert.Equal(t, 2, *byStart["0900"].Remaining)
	// 1000: スタッフ1のみ空き・定員残3 → remaining=1 → △
	assert.Equal(t, SlotVacancyLow, byStart["1000"].Status)
	require.NotNil(t, byStart["1000"].Remaining)
	assert.Equal(t, 1, *byStart["1000"].Remaining)
	// 1100: スタッフ2名空き・定員残0 → ✕
	assert.Equal(t, SlotVacancyFull, byStart["1100"].Status)
	require.NotNil(t, byStart["1100"].Remaining)
	assert.Equal(t, 0, *byStart["1100"].Remaining)
}

func TestGetStaffTimeSlotAvailabilities_AvailableSlotMerge(t *testing.T) {
	// 全スタッフが 10:00 に予約済み → グリッド上 1000 は満員。
	// 予約可能枠テーブルに 10:00 と 20:30 があれば、従来の加算モードと同じく
	// 1000 はスタッフ制約なしで復活し、20:30 は新規枠として追加される。
	day := time.Date(2026, 6, 2, 0, 0, 0, 0, config.JST)
	doctorID1, doctorID2 := uint64(1), uint64(2)
	svc := newAvailabilityTestSvc(
		&model.ReservationType{ID: 10, ClinicID: 3, DurationMinutes: 60, IsActive: true, ReservationVisible: true},
		[]model.Reservation{
			{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed,
				StartTime: time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST),
				EndTime:   time.Date(2026, 6, 2, 11, 0, 0, 0, config.JST)},
			{DoctorID: &doctorID2, Status: model.ReservationStatusConfirmed,
				StartTime: time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST),
				EndTime:   time.Date(2026, 6, 2, 11, 0, 0, 0, config.JST)},
		},
		nil,
	)
	svc.availableSlotRepo = stubAvailableSlotReader{slots: []model.ReservationTypeAvailableSlot{
		specificSlot("10:00", day),
		specificSlot("20:30", day),
	}}

	got, err := svc.GetStaffTimeSlotAvailabilities(context.Background(), 3, 10, 0, day)
	require.NoError(t, err)
	require.Len(t, got, 10)

	byStart := map[string]TimeSlotAvailability{}
	for _, a := range got {
		byStart[a.StartTime] = a
	}
	// 10:00: グリッドは満員だがホワイトリストが復活 → スタッフ非拘束・定員なし → 〇上限なし
	assert.Equal(t, SlotVacancyAvailable, byStart["1000"].Status)
	assert.Nil(t, byStart["1000"].Remaining)
	// 20:30: ホワイトリストのみの枠 → 〇上限なし
	assert.Equal(t, SlotVacancyAvailable, byStart["2030"].Status)
	assert.Nil(t, byStart["2030"].Remaining)
	// 09:00: スタッフ2名空き → 〇 remaining=2
	assert.Equal(t, SlotVacancyAvailable, byStart["0900"].Status)
	require.NotNil(t, byStart["0900"].Remaining)
	assert.Equal(t, 2, *byStart["0900"].Remaining)
}

func TestGetStaffTimeSlotAvailabilities_ClosedDateReturnsEmpty(t *testing.T) {
	day := time.Date(2026, 6, 2, 0, 0, 0, 0, config.JST) // Tuesday
	setting := liffDefaultSetting()
	setting.ClosedDates = []byte(`["2026-06-02"]`)
	svc := newAvailabilityTestSvc(
		&model.ReservationType{ID: 10, ClinicID: 3, DurationMinutes: 60, IsActive: true, ReservationVisible: true},
		nil,
		nil,
	)
	svc.settingRepo = &mockLiffSettingRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return setting, nil
		},
	}

	got, err := svc.GetStaffTimeSlotAvailabilities(context.Background(), 3, 10, 0, day)
	require.NoError(t, err)
	assert.Empty(t, got)
}
