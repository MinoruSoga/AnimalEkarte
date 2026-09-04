package reservation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// newSettingForValidation は業務ルール検証テスト用の設定を返す。
// デフォルト: 営業時間 09:00-19:00 / 休憩 12:00-13:00 / 予約窓 2〜30日
func newSettingForValidation() *model.LineReservationSetting {
	return &model.LineReservationSetting{
		Status:                "running",
		ClosedWeekdays:        []byte("[]"),
		ClosedDates:           []byte("[]"),
		NationalHolidayClosed: false,
		BusinessHours:         []byte(`{"start":"0900","end":"1900"}`),
		BreakHours:            []byte(`[{"start":"1200","end":"1300"}]`),
		BookingWindowMinDays:  2,
		BookingWindowMaxDays:  30,
	}
}

// dateInDays は今日から n 日後の 00:00 JST を返す。
func dateInDays(n int) time.Time {
	now := time.Now().In(config.JST)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
	return today.AddDate(0, 0, n)
}

func TestValidateBusinessRules(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*model.LineReservationSetting)
		date      time.Time
		startTime string
		endTime   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "正常系: 営業時間内・予約窓内・平日",
			date:      dateInDays(3),
			startTime: "1000",
			endTime:   "1015",
			wantErr:   false,
		},
		{
			name:      "過去日 → 400",
			date:      dateInDays(-5),
			startTime: "1000",
			endTime:   "1015",
			wantErr:   true,
			errSubstr: "予約可能日",
		},
		{
			name:      "予約窓より近い日（min_days=2、1日後） → 400",
			date:      dateInDays(1),
			startTime: "1000",
			endTime:   "1015",
			wantErr:   true,
			errSubstr: "予約可能日",
		},
		{
			name:      "予約窓より遠い日（max_days=30、90日後） → 400",
			date:      dateInDays(90),
			startTime: "1000",
			endTime:   "1015",
			wantErr:   true,
			errSubstr: "予約可能日",
		},
		{
			name:      "営業時間外（朝6:00）→ 400",
			date:      dateInDays(3),
			startTime: "0600",
			endTime:   "0615",
			wantErr:   true,
			errSubstr: "営業時間外",
		},
		{
			name:      "営業時間外（20:00）→ 400",
			date:      dateInDays(3),
			startTime: "2000",
			endTime:   "2015",
			wantErr:   true,
			errSubstr: "営業時間外",
		},
		{
			name:      "終了時刻が営業時間をはみ出す → 400",
			date:      dateInDays(3),
			startTime: "1845",
			endTime:   "1915",
			wantErr:   true,
			errSubstr: "営業時間外",
		},
		{
			name:      "休憩時間に重複（12:15-12:30）→ 400",
			date:      dateInDays(3),
			startTime: "1215",
			endTime:   "1230",
			wantErr:   true,
			errSubstr: "休憩時間",
		},
		{
			name:      "休憩時間の境界（11:45-12:00）→ OK（半開区間で重複なし）",
			date:      dateInDays(3),
			startTime: "1145",
			endTime:   "1200",
			wantErr:   false,
		},
		{
			name:      "休憩時間の終点直後（13:00-13:15）→ OK",
			date:      dateInDays(3),
			startTime: "1300",
			endTime:   "1315",
			wantErr:   false,
		},
		{
			name: "休業曜日（月曜）→ 400",
			modify: func(s *model.LineReservationSetting) {
				s.ClosedWeekdays = []byte("[1]")
			},
			date:      firstWeekday(1), // 今日から次の月曜を探す
			startTime: "1000",
			endTime:   "1015",
			wantErr:   true,
			errSubstr: "休業曜日",
		},
		{
			name: "休業日（明示日付）→ 400",
			modify: func(s *model.LineReservationSetting) {
				d := dateInDays(5).Format("2006-01-02")
				s.ClosedDates = []byte(fmt.Sprintf("[%q]", d))
			},
			date:      dateInDays(5),
			startTime: "1000",
			endTime:   "1015",
			wantErr:   true,
			errSubstr: "休業日",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := newSettingForValidation()
			if tt.modify != nil {
				tt.modify(settings)
			}
			err := validateBusinessRules(context.Background(), settings, tt.date, tt.startTime, tt.endTime)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input error, got %v", err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateBusinessRules_MalformedBreakHours_FailsClosed は BE-refactor.md D10/F-2 の
// fail-closed 化を検証する。break_hours の JSON が破損している場合、予約を拒否すべき。
//
// temp-revert RED: liff_service_availability_business.go の parseBusinessHoursForDate で
// `if err := json.Unmarshal(setting.BreakHours, &breaks); err != nil { breaks = nil }` と
// エラーを握りつぶす形に戻すと、休憩時間チェック（step 6）が空配列に対して行われ常に重複なしと
// 判定されるため、本テストは NoError になり RED になる（＝壊れた休憩時間設定下でも予約が通ってしまう）。
func TestValidateBusinessRules_MalformedBreakHours_FailsClosed(t *testing.T) {
	settings := newSettingForValidation()
	settings.BreakHours = []byte(`not-valid-json`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	require.Error(t, err,
		"break_hours の JSON が破損している場合、休憩時間との重複を検証できないため予約を拒否すべき（fail-closed）。"+
			"fail-open（現状バグ）だと breaks=nil にフォールバックし検証がスキップされ予約が通ってしまう")
}

// TestValidateBusinessRules_MalformedBreakEntry_FailsClosed は、break_hours の JSON 配列自体は
// パースできるが個々のエントリの HHMM 値が不正な場合も同様に拒否されることを検証する
// （検証ループ内の個別 continue による fail-open の是正）。
func TestValidateBusinessRules_MalformedBreakEntry_FailsClosed(t *testing.T) {
	settings := newSettingForValidation()
	settings.BreakHours = []byte(`[{"start":"not-a-time","end":"1300"}]`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	require.Error(t, err,
		"個別休憩エントリの時刻形式が不正な場合、そのエントリの重複判定をスキップせず拒否すべき（fail-closed）")
}

// TestValidateBusinessRules_NonDigitHHMMBreakEntry_FailsClosed は、長さ4文字の break_hours
// エントリに非ASCII数字が混入した場合も拒否されることを検証する（R1-3）。
//
// temp-revert RED: MinutesSinceMidnight の ASCII 数字チェックを外すと、byte 型の unsigned
// wraparound で ':' (0x3A) が "偶然" レンジ内の値に化けてしまい（'1'+':'+'0'+'0' → 20時00分と
// 誤変換）、エラーなく break 区間 [1200,780) として扱われる。この壊れた区間は通常の予約時刻
// （10:00-10:15）と重複判定されず、検証をすり抜けて予約が通ってしまう（＝本テストが NoError
// になり RED になる）。
func TestValidateBusinessRules_NonDigitHHMMBreakEntry_FailsClosed(t *testing.T) {
	settings := newSettingForValidation()
	settings.BreakHours = []byte(`[{"start":"1:00","end":"1300"}]`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	require.Error(t, err,
		"長さ4文字でも非ASCII数字混入のHHMMは拒否すべき（fail-closed）。"+
			"byteのunsigned wraparoundでレンジ内に化けると重複判定をすり抜け予約が通ってしまう")
}

// TestValidateBusinessRules_MalformedClosedWeekdays_FailsClosed は BE-refactor.md A-3 の
// fail-closed 化を検証する。closed_weekdays の JSON が破損している場合、休業曜日チェックを
// 検証できないため予約を拒否すべき（break_hours の D10/F-2 と対称）。
func TestValidateBusinessRules_MalformedClosedWeekdays_FailsClosed(t *testing.T) {
	settings := newSettingForValidation()
	settings.ClosedWeekdays = []byte(`not-valid-json`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	require.Error(t, err,
		"closed_weekdays の JSON が破損している場合、休業曜日チェックを検証できないため予約を"+
			"拒否すべき（fail-closed）。fail-open（現状バグ）だとチェックがスキップされ休業日に予約が通ってしまう")
}

// TestValidateBusinessRules_MalformedClosedDates_FailsClosed は closed_dates の JSON が
// 破損している場合も同様に拒否されることを検証する。
func TestValidateBusinessRules_MalformedClosedDates_FailsClosed(t *testing.T) {
	settings := newSettingForValidation()
	settings.ClosedDates = []byte(`not-valid-json`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	require.Error(t, err,
		"closed_dates の JSON が破損している場合、休業日チェックを検証できないため拒否すべき（fail-closed）")
}

// TestValidateBusinessRules_EmptyClosedWeekdaysAndDates_SkipsCheck は未設定（空 JSON 配列）の
// 場合はチェックをスキップし予約が通ることを検証する（fail-closed 化で「未設定」の現行挙動を
// 壊していないことのリグレッション防止）。
func TestValidateBusinessRules_EmptyClosedWeekdaysAndDates_SkipsCheck(t *testing.T) {
	settings := newSettingForValidation()
	settings.ClosedWeekdays = []byte(`[]`)
	settings.ClosedDates = []byte(`[]`)

	err := validateBusinessRules(context.Background(), settings, dateInDays(3), "1000", "1015")

	assert.NoError(t, err)
}

func TestReservationValidators_ValidateAndCreate_MissingTransactorFailsClosed(t *testing.T) {
	validators := NewReservationValidators(
		nil,
		&mockReservationRepository{},
		mockReservationTypeFinder{},
		&mockReservationStaffRepositoryForCapability{},
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		&mockTrimmingDetailRepository{},
		openDayHolidayFinder(),
	)

	result, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{
		ClinicID:          1,
		CustomerID:        2,
		ReservationTypeID: 9,
		Date:              dateInDays(3),
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestReservationValidators_ValidateAndCreate_RejectsHiddenStaffDirectPost(t *testing.T) {
	createCalled := false
	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 1, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true,
				Category: model.ReservationTypeCategoryGeneral,
			}, nil
		},
	}
	staffRepo := &mockReservationStaffRepositoryForCapability{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: false}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
	}
	validators := NewReservationValidators(
		&mockTransactor{}, repo, typeRepo, staffRepo,
		okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{},
		openDayHolidayFinder(),
	)

	result, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{
		ClinicID:          1,
		CustomerID:        2,
		ReservationTypeID: 9,
		StaffID:           3,
		Date:              dateInDays(3),
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationValidators_ValidateAndCreate_RejectsInactiveStaffDirectPost(t *testing.T) {
	createCalled := false
	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 1, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true,
				Category: model.ReservationTypeCategoryGeneral,
			}, nil
		},
	}
	staffRepo := &mockReservationStaffRepositoryForCapability{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: clinicID, IsActive: false, ReservationVisible: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
	}
	validators := NewReservationValidators(
		&mockTransactor{}, repo, typeRepo, staffRepo,
		okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{},
		openDayHolidayFinder(),
	)

	result, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{
		ClinicID:          1,
		CustomerID:        2,
		ReservationTypeID: 9,
		StaffID:           3,
		Date:              dateInDays(3),
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationValidators_ValidateAndCreate_MapsCapacityConflict(t *testing.T) {
	date := dateInDays(3)
	startDT, err := ToDateTime(date, "1000")
	assert.NoError(t, err)
	maxConcurrent := 2
	createCalled := false
	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 3, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			return 0, nil
		},
		countByTypeAndStartTimeFn: func(_ context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(9), reservationTypeID)
			assert.Equal(t, startDT, startTime)
			assert.Nil(t, excludeID)
			return 2, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                 id,
				ClinicID:           clinicID,
				IsActive:           true,
				ReservationVisible: true,
				MaxConcurrent:      &maxConcurrent,
			}, nil
		},
	}
	validators := NewReservationValidators(&mockTransactor{}, repo, typeRepo, &mockReservationStaffRepositoryForCapability{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{}, openDayHolidayFinder())

	result, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{
		ClinicID:          1,
		CustomerID:        2,
		ReservationTypeID: 9,
		Date:              date,
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	limitErr, ok := IsReservationLimitError(err)
	assert.True(t, ok, "expected ReservationLimitError but got: %v", err)
	assert.Equal(t, "SLOT_TAKEN", limitErr.Code)
	assert.Equal(t, "選択された時間枠は満員です。別の時間をお選びください。", limitErr.Message)
	assert.Equal(t, 5, limitErr.RedirectStep)
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationValidators_ValidateAndCreate_RejectsClinicHoliday(t *testing.T) {
	date := dateInDays(3)
	createCalled := false
	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 1, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                 id,
				ClinicID:           clinicID,
				IsActive:           true,
				ReservationVisible: true,
				Category:           model.ReservationTypeCategoryGeneral,
			}, nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, clinicID uint64, gotDate time.Time) (*model.ClinicHoliday, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, date.In(config.JST).Format(time.DateOnly), gotDate.In(config.JST).Format(time.DateOnly))
			return &model.ClinicHoliday{ID: 1, ClinicID: clinicID, Date: gotDate, Reason: "臨時休診"}, nil
		},
	}
	validators := NewReservationValidators(
		&mockTransactor{}, repo, typeRepo, &mockReservationStaffRepositoryForCapability{},
		okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{},
		holidayFinder,
	)

	result, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{
		ClinicID:          1,
		CustomerID:        2,
		ReservationTypeID: 9,
		Date:              date,
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "休診日")
	assert.Nil(t, result)
	assert.False(t, createCalled, "clinic holiday must not persist a LINE reservation")
}

// firstWeekday は今日から数えて最初に指定曜日になる日を返す（minDays=2 を満たすため最低 2 日は先にする）。
func firstWeekday(wd int) time.Time {
	for n := 2; n < 16; n++ {
		d := dateInDays(n)
		if int(d.Weekday()) == wd {
			return d
		}
	}
	return dateInDays(8)
}
