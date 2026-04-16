package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ================================================================
// GetSettings テスト
// ================================================================

func TestLiffService_GetSettings(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: 設定を返す", func(t *testing.T) {
		want := liffDefaultSetting()
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
					assert.Equal(t, uint64(3), clinicID)
					return want, nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetSettings(ctx, 3)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("リポジトリエラー → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return nil, apperrors.ErrNotFound
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		_, err := svc.GetSettings(ctx, 999)
		require.Error(t, err)
	})
}

// ================================================================
// GetProfile テスト
// ================================================================

func TestLiffService_GetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: 顧客プロフィールを返す", func(t *testing.T) {
		want := &model.LineCustomer{ID: 1, ClinicID: 3, DisplayName: "テストユーザー"}
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), id)
					return want, nil
				},
			},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetProfile(ctx, 3, 1)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("顧客が存在しない → NotFound エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return nil, apperrors.ErrNotFound
				},
			},
			&mockLiffValidators{},
			nil,
		)

		_, err := svc.GetProfile(ctx, 3, 999)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	})
}

// ================================================================
// GetCourses テスト
// ================================================================

func TestLiffService_GetCourses(t *testing.T) {
	ctx := context.Background()

	allCourses := []model.ReservationType{
		{ID: 1, Name: "一般診察", IsInternal: false, ReservationVisible: true},
		{ID: 2, Name: "休憩枠", IsInternal: true, ReservationVisible: false}, // 内部メニュー → 除外
		{ID: 3, Name: "手術", IsInternal: false, ReservationVisible: true},
		{ID: 4, Name: "非公開コース", IsInternal: false, ReservationVisible: false}, // 非公開 → 除外
	}

	t.Run("is_internal=false && reservation_visible=true のみを返す", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return allCourses, nil
				},
			},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetCourses(ctx, 3)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, uint64(1), got[0].ID)
		assert.Equal(t, uint64(3), got[1].ID)
	})

	t.Run("全件が内部メニューのとき空スライスを返す", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return []model.ReservationType{{ID: 1, IsInternal: true, ReservationVisible: false}}, nil
				},
			},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetCourses(ctx, 3)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("リポジトリエラー → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return nil, errors.New("db error")
				},
			},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		_, err := svc.GetCourses(ctx, 3)
		require.Error(t, err)
	})
}

// ================================================================
// GetStaffs テスト
// ================================================================

func TestLiffService_GetStaffs(t *testing.T) {
	ctx := context.Background()

	t.Run("reservation_visible=false のスタッフは除外", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{
				findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 1, Name: "林先生", ReservationVisible: true},
						{ID: 2, Name: "スタッフ山田", ReservationVisible: false}, // 非公開 → 除外
					}, nil
				},
				// visible スタッフ(ID=1)のみバルク取得: 除外なし
				findExcludedReservationTypesByStaffIDsFn: func(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
					return nil, nil
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetStaffs(ctx, 3, 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, uint64(1), got[0].ID)
	})

	t.Run("指定コースが除外リストにあるスタッフは除外", func(t *testing.T) {
		const typeID = uint64(5) // 手術コース
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{
				findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 1, Name: "林先生", ReservationVisible: true},
						{ID: 2, Name: "トリマー田中", ReservationVisible: true},
					}, nil
				},
				// バルク取得: ID=2 が typeID=5 を除外している
				findExcludedReservationTypesByStaffIDsFn: func(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
					return []model.StaffReservationExclusion{
						{StaffID: 2, ReservationTypeID: typeID},
					}, nil
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetStaffs(ctx, 3, typeID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, uint64(1), got[0].ID)
	})

	t.Run("FindExcludedReservationTypesByStaffIDs がエラーを返す → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{
				findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
				},
				findExcludedReservationTypesByStaffIDsFn: func(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
					return nil, errors.New("db error")
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		_, err := svc.GetStaffs(ctx, 3, 1)
		require.Error(t, err)
	})
}

// ================================================================
// CreateReservation テスト
// ================================================================

func TestLiffService_CreateReservation(t *testing.T) {
	ctx := context.Background()

	baseInput := func() *CreateReservationInput {
		return &CreateReservationInput{
			ReservationTypeID: 1,
			StaffID:           10,
			Date:              time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
			StartTime:         "1000",
			EndTime:           "1015",
		}
	}

	t.Run("staffID指定あり: バリデーターにそのまま渡る", func(t *testing.T) {
		var capturedStaffID uint64
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Appointment, error) {
					capturedStaffID = input.StaffID
					return &model.Appointment{ID: 99, ClinicID: 3}, nil
				},
			},
			&mockLiffNotifier{},
		)

		got, err := svc.CreateReservation(ctx, 3, 1, baseInput())
		require.NoError(t, err)
		assert.Equal(t, uint64(99), got.ID)
		assert.Equal(t, uint64(10), capturedStaffID)
	})

	t.Run("staffID=0 + top_priority: 先頭スタッフに固定割当", func(t *testing.T) {
		var assignedStaffID uint64
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					s := liffDefaultSetting()
					s.NoStaffMode = "top_priority"
					return s, nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{
				findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 5, Name: "林先生", ReservationVisible: true},
						{ID: 6, Name: "三井先生", ReservationVisible: true},
					}, nil
				},
				findExcludedReservationTypesByStaffIDsFn: func(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
					return nil, nil
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Appointment, error) {
					assignedStaffID = input.StaffID
					return &model.Appointment{ID: 1}, nil
				},
			},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.StaffID = 0
		_, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.Equal(t, uint64(5), assignedStaffID, "top_priority は先頭スタッフに割当")
	})

	t.Run("staffID=0 + first_available: 先頭が埋まりのとき2番目に割当", func(t *testing.T) {
		date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
		var assignedStaffID uint64
		doctorID5 := uint64(5)

		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					s := liffDefaultSetting()
					s.NoStaffMode = "first_available"
					return s, nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{
				findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 5, Name: "林先生", ReservationVisible: true},
						{ID: 6, Name: "三井先生", ReservationVisible: true},
					}, nil
				},
				findExcludedReservationTypesByStaffIDsFn: func(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
					return nil, nil
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Appointment, error) {
					// ID=5（林先生）は 10:00-10:15 に既存予約あり
					return []model.Appointment{
						{
							DoctorID:  &doctorID5,
							StartTime: date.Add(10 * time.Hour),
							EndTime:   date.Add(10*time.Hour + 15*time.Minute),
							Status:    model.ReservationStatusConfirmed,
						},
					}, nil
				},
			},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Appointment, error) {
					assignedStaffID = input.StaffID
					return &model.Appointment{ID: 1}, nil
				},
			},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.StaffID = 0
		input.Date = date
		_, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.Equal(t, uint64(6), assignedStaffID, "林先生が埋まり → 三井先生に割当")
	})

	t.Run("CustomerFieldsが非空のとき UpdateAdditionalFields が呼ばれる", func(t *testing.T) {
		var updateCalled bool
		fields := json.RawMessage(`{"phone":"090-1234-5678"}`)
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
				updateAdditionalFieldsFn: func(_ context.Context, clinicID, id uint64, f []byte) error {
					updateCalled = true
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), id)
					assert.Equal(t, []byte(fields), f)
					return nil
				},
			},
			&mockLiffValidators{},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.CustomerFields = fields
		_, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.True(t, updateCalled, "UpdateAdditionalFields が呼ばれること")
	})

	t.Run("CustomerFieldsが'{}'のとき UpdateAdditionalFields は呼ばれない", func(t *testing.T) {
		var updateCalled bool
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
				updateAdditionalFieldsFn: func(_ context.Context, _, _ uint64, _ []byte) error {
					updateCalled = true
					return nil
				},
			},
			&mockLiffValidators{},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.CustomerFields = json.RawMessage(`{}`)
		_, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.False(t, updateCalled)
	})

	t.Run("自動オーナー紐付け後に予約へ owner_id / pet_id を反映する", func(t *testing.T) {
		svc := newLiffSvcWithDeps(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(77), id)
					assert.Equal(t, uint64(200), fields["owner_id"])
					assert.Equal(t, uint64(300), fields["pet_id"])
					return &model.Appointment{ID: id, ClinicID: clinicID, OwnerID: ptrUint64(200), PetID: ptrUint64(300)}, nil
				},
			},
			&mockLiffCustomerRepository{
				findByIDFn: func() func(context.Context, uint64, uint64) (*model.LineCustomer, error) {
					calls := 0
					return func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
						calls++
						if calls == 1 {
							return &model.LineCustomer{ID: 1}, nil
						}
						return &model.LineCustomer{
							ID:      1,
							OwnerID: ptrUint64(200),
							Owner: &model.Owner{
								ID: 200,
								Pets: []model.Pet{
									{ID: 300, Name: "ポチ"},
									{ID: 301, Name: "タマ"},
								},
							},
						}, nil
					}
				}(),
				updateOwnerLinkFn: func(_ context.Context, clinicID, id uint64, ownerID *uint64) error {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, ownerID)
					assert.Equal(t, uint64(200), *ownerID)
					return nil
				},
			},
			&mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, clinicID uint64, name, phone string) (*model.Owner, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, "田中太郎", name)
					assert.Equal(t, "090-1234-5678", phone)
					return &model.Owner{ID: 200}, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Appointment, error) {
					return &model.Appointment{ID: 77, ClinicID: input.ClinicID}, nil
				},
			},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.CustomerFields = json.RawMessage(`{"owner_name":"田中太郎","phone":"090-1234-5678","pets":[{"name":"ポチ"}]}`)
		got, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		require.NotNil(t, got.OwnerID)
		require.NotNil(t, got.PetID)
		assert.Equal(t, uint64(200), *got.OwnerID)
		assert.Equal(t, uint64(300), *got.PetID)
	})

	t.Run("NotifyCreated が fire-and-forget で呼ばれる", func(t *testing.T) {
		notifyCh := make(chan struct{}, 1)
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1, LineUserID: "U001"}, nil
				},
			},
			&mockLiffValidators{},
			&mockLiffNotifier{
				notifyCreatedFn: func(_ context.Context, _ *model.Appointment, _ *model.LineCustomer) {
					notifyCh <- struct{}{}
				},
			},
		)

		_, err := svc.CreateReservation(ctx, 3, 1, baseInput())
		require.NoError(t, err)

		select {
		case <-notifyCh:
			// OK
		case <-time.After(500 * time.Millisecond):
			t.Fatal("NotifyCreated が呼ばれなかった")
		}
	})

	t.Run("ReservationLimitError はそのまま伝播する", func(t *testing.T) {
		limitErr := &ReservationLimitError{Code: "SLOT_TAKEN", Message: "予約済み", RedirectStep: 5}
		svc := newLiffSvc(
			&mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return liffDefaultSetting(), nil
				},
			},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, _ *CreateReservationInput) (*model.Appointment, error) {
					return nil, limitErr
				},
			},
			nil,
		)

		_, err := svc.CreateReservation(ctx, 3, 1, baseInput())
		require.Error(t, err)
		limErr, ok := IsReservationLimitError(err)
		require.True(t, ok)
		assert.Equal(t, "SLOT_TAKEN", limErr.Code)
		assert.Equal(t, 5, limErr.RedirectStep)
	})
}

// ================================================================
// GetMyReservations テスト
// ================================================================

func TestLiffService_GetMyReservations(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: 顧客の予約一覧を返す", func(t *testing.T) {
		want := []model.Appointment{{ID: 1, ClinicID: 3}, {ID: 2, ClinicID: 3}}
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByCustomerIDFn: func(_ context.Context, clinicID, customerID uint64) ([]model.Appointment, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), customerID)
					return want, nil
				},
			},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		got, err := svc.GetMyReservations(ctx, 3, 1)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("リポジトリエラー → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByCustomerIDFn: func(_ context.Context, _, _ uint64) ([]model.Appointment, error) {
					return nil, errors.New("db error")
				},
			},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)

		_, err := svc.GetMyReservations(ctx, 3, 1)
		require.Error(t, err)
	})
}

// ================================================================
// CancelReservation テスト
// ================================================================

func TestLiffService_CancelReservation(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: キャンセル成功", func(t *testing.T) {
		var cancelCalled bool
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
					return &model.Appointment{ID: 10, ClinicID: 3}, nil
				},
				cancelByIDFn: func(_ context.Context, clinicID, customerID, id uint64) error {
					cancelCalled = true
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), customerID)
					assert.Equal(t, uint64(10), id)
					return nil
				},
			},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			&mockLiffValidators{},
			&mockLiffNotifier{
				notifyCancelledFn: func(_ context.Context, _ *model.Appointment, _ *model.LineCustomer) {},
			},
		)

		err := svc.CancelReservation(ctx, 3, 1, 10)
		require.NoError(t, err)
		assert.True(t, cancelCalled)
	})

	t.Run("CancelByID がエラーを返す → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
					return nil, apperrors.ErrNotFound
				},
				cancelByIDFn: func(_ context.Context, _, _, _ uint64) error {
					return apperrors.ErrNotFound
				},
			},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			&mockLiffNotifier{},
		)

		err := svc.CancelReservation(ctx, 3, 1, 999)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	})

	t.Run("NotifyCancelled が fire-and-forget で呼ばれる", func(t *testing.T) {
		notifyCh := make(chan struct{}, 1)
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
					return &model.Appointment{ID: 10}, nil
				},
				cancelByIDFn: func(_ context.Context, _, _, _ uint64) error { return nil },
			},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			&mockLiffValidators{},
			&mockLiffNotifier{
				notifyCancelledFn: func(_ context.Context, _ *model.Appointment, _ *model.LineCustomer) {
					notifyCh <- struct{}{}
				},
			},
		)

		err := svc.CancelReservation(ctx, 3, 1, 10)
		require.NoError(t, err)

		select {
		case <-notifyCh:
			// OK
		case <-time.After(500 * time.Millisecond):
			t.Fatal("NotifyCancelled が呼ばれなかった")
		}
	})
}

// ================================================================
// isStaffAvailable ヘルパーテスト
// ================================================================

func TestIsStaffAvailable(t *testing.T) {
	base := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	doctorID := uint64(1)
	otherDoctorID := uint64(2)

	tests := []struct {
		name      string
		staffID   uint64
		startMin  int
		endMin    int
		dayResv   []model.Appointment
		wantAvail bool
	}{
		{
			name:    "予約なし → 空き",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv:   nil,
			wantAvail: true,
		},
		{
			name:    "完全に重複する予約あり → 埋まり",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv: []model.Appointment{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: false,
		},
		{
			name:    "キャンセル済みの予約は無視",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv: []model.Appointment{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusCancelled},
			},
			wantAvail: true,
		},
		{
			name:    "別スタッフの予約は無視",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv: []model.Appointment{
				{DoctorID: &otherDoctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: true,
		},
		{
			name:    "直前に終わる予約は重複しない",
			staffID: 1, startMin: 615, endMin: 630,
			dayResv: []model.Appointment{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: true,
		},
		{
			name:    "直後から始まる予約は重複しない",
			staffID: 1, startMin: 575, endMin: 600,
			dayResv: []model.Appointment{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaffAvailable(tt.staffID, tt.startMin, tt.endMin, tt.dayResv)
			assert.Equal(t, tt.wantAvail, got)
		})
	}
}

// ================================================================
// isExcluded ヘルパーテスト
// ================================================================

func TestIsExcluded(t *testing.T) {
	excluded := []model.StaffReservationExclusion{
		{StaffID: 1, ReservationTypeID: 5},
		{StaffID: 1, ReservationTypeID: 8},
	}

	t.Run("除外リストにあるコースIDはtrue", func(t *testing.T) {
		assert.True(t, isExcluded(excluded, 5))
		assert.True(t, isExcluded(excluded, 8))
	})

	t.Run("除外リストにないコースIDはfalse", func(t *testing.T) {
		assert.False(t, isExcluded(excluded, 1))
		assert.False(t, isExcluded(excluded, 99))
	})

	t.Run("除外リストが空のときは常にfalse", func(t *testing.T) {
		assert.False(t, isExcluded(nil, 5))
		assert.False(t, isExcluded([]model.StaffReservationExclusion{}, 5))
	})
}
