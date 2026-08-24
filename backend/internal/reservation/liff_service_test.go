package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
		{ID: 1, Name: "一般診察", IsActive: true, IsInternal: false, ReservationVisible: true},
		{ID: 2, Name: "休憩枠", IsActive: true, IsInternal: true, ReservationVisible: false}, // 内部メニュー → 除外
		{ID: 3, Name: "手術", IsActive: true, IsInternal: false, ReservationVisible: true},
		{ID: 4, Name: "非公開コース", IsActive: true, IsInternal: false, ReservationVisible: false}, // 非公開 → 除外
		{ID: 5, Name: "無効コース", IsActive: false, IsInternal: false, ReservationVisible: true},  // 無効 → 除外
	}

	t.Run("is_active=true && is_internal=false && reservation_visible=true のみを返す", func(t *testing.T) {
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

	t.Run("子を持つ親ノードは除外し葉ノードを返す", func(t *testing.T) {
		parentID := uint64(10)
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return []model.ReservationType{
						{ID: parentID, Name: "トリミング", IsActive: true, IsInternal: false, ReservationVisible: true},
						{ID: 11, ParentID: &parentID, Name: "シャンプー", IsActive: true, IsInternal: false, ReservationVisible: true},
					}, nil
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
		require.Len(t, got, 1)
		assert.Equal(t, uint64(11), got[0].ID)
	})

	t.Run("非表示の子でも親ノードは構造上の葉として扱わない", func(t *testing.T) {
		parentID := uint64(20)
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return []model.ReservationType{
						{ID: parentID, Name: "トリミング", IsActive: true, IsInternal: false, ReservationVisible: true},
						{ID: 21, ParentID: &parentID, Name: "非公開シャンプー", IsActive: true, IsInternal: false, ReservationVisible: false},
					}, nil
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

	t.Run("全件が内部メニューのとき空スライスを返す", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return []model.ReservationType{{ID: 1, IsActive: true, IsInternal: true, ReservationVisible: false}}, nil
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
			&mockLiffTypeRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
					return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true}, nil
				},
			},
			&mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 1, Name: "林先生", ReservationVisible: true},
						{ID: 2, Name: "スタッフ山田", ReservationVisible: false}, // 非公開 → 除外
					}, nil
				},
				findCapabilitiesByStaffIDsFn: func(_ context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, []uint64{1}, staffIDs)
					return []model.StaffReservationCapability{{ClinicID: 3, StaffID: 1, ReservationTypeID: 1}}, nil
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

	t.Run("指定コースが対応可能リストにないスタッフは除外", func(t *testing.T) {
		const typeID = uint64(5) // 手術コース
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
					return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true}, nil
				},
			},
			&mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 1, Name: "林先生", ReservationVisible: true},
						{ID: 2, Name: "トリマー田中", ReservationVisible: true},
					}, nil
				},
				findCapabilitiesByStaffIDsFn: func(_ context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, []uint64{1, 2}, staffIDs)
					return []model.StaffReservationCapability{
						{ClinicID: 3, StaffID: 1, ReservationTypeID: typeID},
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

	t.Run("FindAllReservationCapabilitiesByStaffIDs がエラーを返す → エラー伝播", func(t *testing.T) {
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
					return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true}, nil
				},
			},
			&mockLiffStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
				},
				findCapabilitiesByStaffIDsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
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
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					capturedStaffID = input.StaffID
					return &model.Reservation{ID: 99, ClinicID: 3}, nil
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
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 5, Name: "林先生", ReservationVisible: true},
						{ID: 6, Name: "三井先生", ReservationVisible: true},
					}, nil
				},
				findCapabilitiesByStaffIDsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
					return []model.StaffReservationCapability{
						{ClinicID: 3, StaffID: 5, ReservationTypeID: 1},
						{ClinicID: 3, StaffID: 6, ReservationTypeID: 1},
					}, nil
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
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					assignedStaffID = input.StaffID
					return &model.Reservation{ID: 1}, nil
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
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{
						{ID: 5, Name: "林先生", ReservationVisible: true},
						{ID: 6, Name: "三井先生", ReservationVisible: true},
					}, nil
				},
				findCapabilitiesByStaffIDsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
					return []model.StaffReservationCapability{
						{ClinicID: 3, StaffID: 5, ReservationTypeID: 1},
						{ClinicID: 3, StaffID: 6, ReservationTypeID: 1},
					}, nil
				},
			},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
					// ID=5（林先生）は 10:00-10:15 に既存予約あり
					return []model.Reservation{
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
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					assignedStaffID = input.StaffID
					return &model.Reservation{ID: 1}, nil
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

	t.Run("SEC-CS2-F02: 未紐付け顧客は氏名+電話一致でも owner-link しない", func(t *testing.T) {
		var (
			updateOwnerCalls  int
			ownerLookupCalls  int
			reservationUpdate int
		)
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
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					reservationUpdate++
					return &model.Reservation{ID: 77}, nil
				},
			},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					// 未紐付けのまま
					return &model.LineCustomer{ID: 1}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					updateOwnerCalls++
					return nil
				},
			},
			&mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					ownerLookupCalls++
					return &model.Owner{ID: 200}, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					return &model.Reservation{ID: 77, ClinicID: input.ClinicID}, nil
				},
			},
			&mockLiffNotifier{},
		)

		input := baseInput()
		// 攻撃者が被害者の氏名+電話番号を入力しても owner 自動紐付けは発生しない
		input.CustomerFields = json.RawMessage(`{"owner_name":"田中太郎","phone":"090-1234-5678","pets":[{"name":"ポチ"}]}`)
		got, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.Equal(t, 0, updateOwnerCalls, "UpdateOwnerLink は呼ばれてはならない")
		assert.Equal(t, 0, ownerLookupCalls, "FindByNameAndPhone は呼ばれてはならない")
		assert.Equal(t, 0, reservationUpdate, "未紐付けでは予約への owner/pet 反映もしない")
		assert.Nil(t, got.OwnerID)
		assert.Nil(t, got.PetID)
	})

	t.Run("既に紐付け済み顧客は予約へ owner_id / pet_id を反映する", func(t *testing.T) {
		var updateOwnerCalls int
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
				updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(77), id)
					assert.Equal(t, uint64(200), fields["owner_id"])
					assert.Equal(t, uint64(300), fields["pet_id"])
					return &model.Reservation{ID: id, ClinicID: clinicID, OwnerID: ptrUint64(200), PetID: ptrUint64(300)}, nil
				},
			},
			&mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					// トークン/スタッフ経路で既に紐付け済み
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
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					updateOwnerCalls++
					return nil
				},
			},
			&mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					t.Fatal("既に紐付け済みなら FindByNameAndPhone は不要")
					return nil, nil
				},
			},
			&mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					return &model.Reservation{ID: 77, ClinicID: input.ClinicID}, nil
				},
			},
			&mockLiffNotifier{},
		)

		input := baseInput()
		input.CustomerFields = json.RawMessage(`{"owner_name":"田中太郎","phone":"090-1234-5678","pets":[{"name":"ポチ"}]}`)
		got, err := svc.CreateReservation(ctx, 3, 1, input)
		require.NoError(t, err)
		assert.Equal(t, 0, updateOwnerCalls, "既紐付け顧客でも UpdateOwnerLink は呼ばない")
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
				notifyCreatedFn: func(_ context.Context, _ *model.Reservation, _ *model.LineCustomer) {
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
				validateAndCreateFn: func(_ context.Context, _ *CreateReservationInput) (*model.Reservation, error) {
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
		want := []model.Reservation{{ID: 1, ClinicID: 3}, {ID: 2, ClinicID: 3}}
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByCustomerIDFn: func(_ context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
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
				findByCustomerIDFn: func(_ context.Context, _, _ uint64) ([]model.Reservation, error) {
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

	t.Run("正常系: キャンセル後に同じ clinic と reservation で draft cleanup を呼ぶ", func(t *testing.T) {
		var cleanupCalls int
		medicalRecord := &mockMedicalRecordService{
			deleteDraftFromReservationFn: func(gotCtx context.Context, clinicID, reservationID uint64) {
				cleanupCalls++
				assert.Equal(t, ctx, gotCtx)
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(10), reservationID)
			},
		}
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				cancelByIDFn: func(_ context.Context, _, _, _ uint64) error { return nil },
			},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)
		svc.medicalRecord = medicalRecord

		err := svc.CancelReservation(ctx, 3, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, 1, cleanupCalls)
	})

	t.Run("キャンセル失敗時は draft cleanup を呼ばない", func(t *testing.T) {
		var cleanupCalls int
		medicalRecord := &mockMedicalRecordService{
			deleteDraftFromReservationFn: func(_ context.Context, _, _ uint64) {
				cleanupCalls++
			},
		}
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				cancelByIDFn: func(_ context.Context, _, _, _ uint64) error {
					return errors.New("cancel failed")
				},
			},
			&mockLiffCustomerRepository{},
			&mockLiffValidators{},
			nil,
		)
		svc.medicalRecord = medicalRecord

		err := svc.CancelReservation(ctx, 3, 1, 10)

		require.Error(t, err)
		assert.Zero(t, cleanupCalls)
	})

	t.Run("正常系: キャンセル成功", func(t *testing.T) {
		var cancelCalled bool
		svc := newLiffSvc(
			&mockLiffSettingRepository{},
			&mockLiffTypeRepository{},
			&mockLiffStaffRepository{},
			&mockLiffScheduleRepository{},
			&mockLiffAdminRepository{
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 10, ClinicID: 3}, nil
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
				notifyCancelledFn: func(_ context.Context, _ *model.Reservation, _ *model.LineCustomer) {},
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
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
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
				findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 10}, nil
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
				notifyCancelledFn: func(_ context.Context, _ *model.Reservation, _ *model.LineCustomer) {
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
		dayResv   []model.Reservation
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
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: false,
		},
		{
			name:    "キャンセル済みの予約は無視",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusCancelled},
			},
			wantAvail: true,
		},
		{
			name:    "別スタッフの予約は無視",
			staffID: 1, startMin: 600, endMin: 615,
			dayResv: []model.Reservation{
				{DoctorID: &otherDoctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: true,
		},
		{
			name:    "直前に終わる予約は重複しない",
			staffID: 1, startMin: 615, endMin: 630,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, StartTime: base.Add(10 * time.Hour), EndTime: base.Add(10*time.Hour + 15*time.Minute), Status: model.ReservationStatusConfirmed},
			},
			wantAvail: true,
		},
		{
			name:    "直後から始まる予約は重複しない",
			staffID: 1, startMin: 575, endMin: 600,
			dayResv: []model.Reservation{
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
// isCapable ヘルパーテスト
// ================================================================

func TestIsCapable(t *testing.T) {
	capabilities := []model.StaffReservationCapability{
		{StaffID: 1, ReservationTypeID: 5},
		{StaffID: 1, ReservationTypeID: 8},
	}

	t.Run("対応可能リストにあるコースIDはtrue", func(t *testing.T) {
		assert.True(t, isCapable(capabilities, 5))
		assert.True(t, isCapable(capabilities, 8))
	})

	t.Run("対応可能リストにないコースIDはfalse", func(t *testing.T) {
		assert.False(t, isCapable(capabilities, 1))
		assert.False(t, isCapable(capabilities, 99))
	})

	t.Run("対応可能リストが空のときは常にfalse", func(t *testing.T) {
		assert.False(t, isCapable(nil, 5))
		assert.False(t, isCapable([]model.StaffReservationCapability{}, 5))
	})
}

// ================================================================
// NewLiffService / NewLiffServiceWithType テスト
// ================================================================

// TestNewLiffService は削除された委譲コンストラクタ NewLiffService(typeRepo に nil を渡す
// バリアント)が担っていた検証を、その本体だった NewLiffServiceWithType(<既存引数>, nil) への
// 直接呼出として維持する（BE-refactor.md D-6）。
func TestNewLiffService(t *testing.T) {
	svc := NewLiffServiceWithType(
		&mockLiffSettingRepository{},
		&mockLiffTypeRepository{},
		nil,
		&mockLiffStaffRepository{},
		&mockLiffScheduleRepository{},
		&mockLiffAdminRepository{},
		&mockLiffCustomerRepository{},
		&mockLiffOwnerRepository{},
		&mockTransactor{},
		&mockLiffReservationRepository{},
		&mockLiffNotifier{},
		&mockLiffUnavailableTimeRepository{},
		&mockAvailableSlotRepository{},
		&mockReservationTypeOccupationRepository{},
		&mockTrimmingCourseRepository{},
		&mockTrimmingOptionRepository{},
		&mockTrimmingDetailRepository{},
		&mockVaccinationRepository{},
		openDayHolidayFinder(),
	)
	require.NotNil(t, svc)

	impl, ok := svc.(*liffService)
	require.True(t, ok, "戻り値は具象型 *liffService であるべき")
	assert.Nil(t, impl.typeRepo, "typeRepo に nil を渡した場合 nil のままであること")
	assert.NotNil(t, impl.validators, "validators が初期化されていること")
}

func TestNewLiffServiceWithType(t *testing.T) {
	typeRepo := &mockLiffTypeRepository{}
	medicalRecord := &mockMedicalRecordService{}
	svc := NewLiffServiceWithType(
		&mockLiffSettingRepository{},
		&mockLiffTypeRepository{},
		typeRepo,
		&mockLiffStaffRepository{},
		&mockLiffScheduleRepository{},
		&mockLiffAdminRepository{},
		&mockLiffCustomerRepository{},
		&mockLiffOwnerRepository{},
		&mockTransactor{},
		&mockLiffReservationRepository{},
		&mockLiffNotifier{},
		&mockLiffUnavailableTimeRepository{},
		&mockAvailableSlotRepository{},
		&mockReservationTypeOccupationRepository{},
		&mockTrimmingCourseRepository{},
		&mockTrimmingOptionRepository{},
		&mockTrimmingDetailRepository{},
		&mockVaccinationRepository{},
		openDayHolidayFinder(),
		medicalRecord,
	)
	require.NotNil(t, svc)

	impl, ok := svc.(*liffService)
	require.True(t, ok, "戻り値は具象型 *liffService であるべき")
	assert.Same(t, typeRepo, impl.typeRepo, "typeRepo が明示的に配線されること")
	assert.Same(t, medicalRecord, impl.medicalRecord, "medicalRecord cleanup view が明示的に配線されること")
	assert.NotNil(t, impl.validators, "validators が初期化されていること")
}
