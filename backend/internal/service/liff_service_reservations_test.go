package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// ================================================================
// テストヘルパー: liff_service_reservations.go 専用
// ================================================================

// reservationSvcOpt は newReservationTestSvc が返す *liffService をカスタマイズするためのオプション。
type reservationSvcOpt func(*liffService)

// newReservationTestSvc は CreateReservation の成功パスに必要な最小限の依存を
// デフォルト構成した *liffService を返す。個々のテストは opts で必要なフィールドのみ上書きする。
func newReservationTestSvc(opts ...reservationSvcOpt) *liffService {
	svc := &liffService{
		settingRepo: &mockLiffSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return liffDefaultSetting(), nil
			},
		},
		customerRepo: &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{ID: 1}, nil
			},
		},
		ownerRepo:       &mockLiffOwnerRepository{},
		adminRepo:       &mockLiffAdminRepository{},
		reservationRepo: &mockLiffReservationRepository{},
		validators:      &mockLiffValidators{},
		notifier:        &mockLiffNotifier{},
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func reservationBaseInput() *CreateReservationInput {
	return &CreateReservationInput{
		ReservationTypeID: 1,
		StaffID:           10,
		Date:              time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
		StartTime:         "1000",
		EndTime:           "1015",
	}
}

// ================================================================
// CreateReservation: liff_service_test.go でカバーされていない分岐
// ================================================================

func TestLiffService_CreateReservation_SettingLookupError(t *testing.T) {
	svc := newReservationTestSvc(func(s *liffService) {
		s.settingRepo = &mockLiffSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		}
	})

	_, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.Error(t, err)
}

func TestLiffService_CreateReservation_ValidateAndCreateGenericError(t *testing.T) {
	svc := newReservationTestSvc(func(s *liffService) {
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, _ *CreateReservationInput) (*model.Reservation, error) {
				return nil, errors.New("generic failure")
			},
		}
	})

	_, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.Error(t, err)
	_, ok := reservation.IsReservationLimitError(err)
	assert.False(t, ok, "通常のエラーは ReservationLimitError としてラップされない")
}

func TestLiffService_CreateReservation_DateParseErrorSkipsAutoDelegation(t *testing.T) {
	var capturedStaffID uint64
	svc := newReservationTestSvc(func(s *liffService) {
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
				capturedStaffID = input.StaffID
				return &model.Reservation{ID: 1, ClinicID: input.ClinicID}, nil
			},
		}
	})

	input := reservationBaseInput()
	input.StaffID = 0
	input.StartTime = "invalid" // reservation.ToDateTime のパースを失敗させる

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), capturedStaffID, "日時パース失敗時は自動割当をスキップし staff_id=0 のまま渡る")
}

func TestLiffService_CreateReservation_TrimmingDetail_Success(t *testing.T) {
	var createdDetail *model.AppointmentTrimmingDetail
	var setOptionsCalled bool
	var setOptionsIDs []uint64
	courseID := uint64(7)

	svc := newReservationTestSvc(func(s *liffService) {
		s.trimmingDetailRepo = &mockTrimmingDetailRepository{
			createFn: func(_ context.Context, detail *model.AppointmentTrimmingDetail) error {
				createdDetail = detail
				return nil
			},
			setOptionsFn: func(_ context.Context, _, _ uint64, optionIDs []uint64) error {
				setOptionsCalled = true
				setOptionsIDs = optionIDs
				return nil
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, _ *CreateReservationInput) (*model.Reservation, error) {
				return &model.Reservation{ID: 42, ClinicID: 3}, nil
			},
		}
	})

	input := reservationBaseInput()
	input.TrimmingCourseID = &courseID
	input.TrimmingOptionIDs = []uint64{1, 2}

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.NoError(t, err)
	require.NotNil(t, createdDetail)
	assert.Equal(t, uint64(42), createdDetail.AppointmentID)
	if assert.NotNil(t, createdDetail.CourseID) {
		assert.Equal(t, courseID, *createdDetail.CourseID)
	}
	assert.True(t, setOptionsCalled)
	assert.Equal(t, []uint64{1, 2}, setOptionsIDs)
}

func TestLiffService_CreateReservation_TrimmingDetail_CreateError_BestEffort(t *testing.T) {
	var setOptionsCalled bool
	courseID := uint64(7)

	svc := newReservationTestSvc(func(s *liffService) {
		s.trimmingDetailRepo = &mockTrimmingDetailRepository{
			createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				return errors.New("db error")
			},
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				setOptionsCalled = true
				return nil
			},
		}
	})

	input := reservationBaseInput()
	input.TrimmingCourseID = &courseID
	input.TrimmingOptionIDs = []uint64{1}

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.NoError(t, err, "トリミング詳細作成失敗は best-effort であり予約自体は成功する")
	assert.False(t, setOptionsCalled, "詳細作成に失敗した場合はオプション設定を試みない")
}

func TestLiffService_CreateReservation_NotifierNil_NoPanic(t *testing.T) {
	svc := newReservationTestSvc(func(s *liffService) {
		s.notifier = nil
	})

	got, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestLiffService_CreateReservation_NotifyEnrichmentFallback(t *testing.T) {
	var notifiedAppt *model.Reservation
	var notifiedCustomer *model.LineCustomer

	svc := newReservationTestSvc(func(s *liffService) {
		s.adminRepo = &mockLiffAdminRepository{
			findByIDForNotifyFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return nil, errors.New("not found")
			},
		}
		s.customerRepo = &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return nil, errors.New("customer lookup failed")
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
				return &model.Reservation{ID: 55, ClinicID: input.ClinicID}, nil
			},
		}
		s.notifier = &mockLiffNotifier{
			notifyCreatedFn: func(_ context.Context, appt *model.Reservation, customer *model.LineCustomer) {
				notifiedAppt = appt
				notifiedCustomer = customer
			},
		}
	})

	got, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.NoError(t, err)
	require.NotNil(t, notifiedAppt)
	assert.Equal(t, got.ID, notifiedAppt.ID, "enrichment に失敗した場合は元の appt で通知する")
	assert.Nil(t, notifiedCustomer, "顧客取得失敗時は nil のまま通知される（best-effort）")
}

// ================================================================
// resolveReservationPetID: 直接単体テスト
// ================================================================

func TestResolveReservationPetID(t *testing.T) {
	tests := []struct {
		name           string
		customer       *model.LineCustomer
		customerFields []byte
		wantNil        bool
		wantID         uint64
	}{
		{
			name:    "customer が nil",
			wantNil: true,
		},
		{
			name:     "customer.Owner が nil",
			customer: &model.LineCustomer{ID: 1},
			wantNil:  true,
		},
		{
			name:     "オーナーにペットがいない",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{}}},
			wantNil:  true,
		},
		{
			name:     "ペットが1頭のみ -> customerFields に関わらずそのIDを返す",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{{ID: 300, Name: "ポチ"}}}},
			wantID:   300,
		},
		{
			name:           "複数頭・customerFields が空 -> nil",
			customer:       &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{{ID: 300}, {ID: 301}}}},
			customerFields: nil,
			wantNil:        true,
		},
		{
			name:           "複数頭・customerFields が不正JSON -> nil",
			customer:       &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{{ID: 300}, {ID: 301}}}},
			customerFields: []byte(`{invalid`),
			wantNil:        true,
		},
		{
			name:           "複数頭・pets配列が空 -> nil",
			customer:       &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{{ID: 300}, {ID: 301}}}},
			customerFields: []byte(`{"pets":[]}`),
			wantNil:        true,
		},
		{
			name:           "複数頭・ペット名が空白のみ -> nil",
			customer:       &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{{ID: 300}, {ID: 301}}}},
			customerFields: []byte(`{"pets":[{"name":"  "}]}`),
			wantNil:        true,
		},
		{
			name: "複数頭・名前一致 -> 一致したIDを返す",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{
				{ID: 300, Name: "ポチ"}, {ID: 301, Name: "タマ"},
			}}},
			customerFields: []byte(`{"pets":[{"name":"タマ"}]}`),
			wantID:         301,
		},
		{
			name: "複数頭・名前不一致 -> nil",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{
				{ID: 300, Name: "ポチ"}, {ID: 301, Name: "タマ"},
			}}},
			customerFields: []byte(`{"pets":[{"name":"ハチ"}]}`),
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveReservationPetID(tt.customer, tt.customerFields)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.wantID, *got)
			}
		})
	}
}

// ================================================================
// tryAutoLinkOwner: 直接単体テスト
// ================================================================

func TestLiffService_tryAutoLinkOwner(t *testing.T) {
	t.Run("ownerRepo が nil なら no-op", func(t *testing.T) {
		svc := &liffService{ownerRepo: nil}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"a","phone":"b"}`))
	})

	t.Run("customerRepo.FindByID がエラー -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					called = true
					return nil, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return nil, errors.New("db error")
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"a","phone":"b"}`))
		assert.False(t, called)
	})

	t.Run("customer が nil -> スキップ", func(t *testing.T) {
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return nil, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"a","phone":"b"}`))
	})

	t.Run("既に owner_id 紐付け済み -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1, OwnerID: ptrUint64(5)}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					called = true
					return nil, nil
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"a","phone":"b"}`))
		assert.False(t, called)
	})

	t.Run("customerFields が空 -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					called = true
					return nil, nil
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, nil)
		assert.False(t, called)
	})

	t.Run("customerFields が不正JSON -> スキップ", func(t *testing.T) {
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{invalid`))
	})

	t.Run("customer_name フォールバック使用でも phone 空 -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					called = true
					return nil, nil
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"customer_name":"田中太郎"}`))
		assert.False(t, called, "電話番号がなければ検索しない")
	})

	t.Run("owner_name が customer_name より優先される", func(t *testing.T) {
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, name, phone string) (*model.Owner, error) {
					assert.Equal(t, "田中太郎", name)
					assert.Equal(t, "090-0000-0000", phone)
					return nil, nil // 0件 or 複数件 -> 紐付けなし
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"田中太郎","customer_name":"別名","phone":"090-0000-0000"}`))
	})

	t.Run("FindByNameAndPhone エラー -> best-effortでスキップ", func(t *testing.T) {
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					return nil, errors.New("db error")
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"田中太郎","phone":"090-0000-0000"}`))
	})

	t.Run("UpdateOwnerLink エラー -> best-effortでパニックしない", func(t *testing.T) {
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return errors.New("db error")
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					return &model.Owner{ID: 200}, nil
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"田中太郎","phone":"090-0000-0000"}`))
	})

	t.Run("成功時: owner_id が紐付けられる", func(t *testing.T) {
		var linkedOwnerID *uint64
		svc := &liffService{
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, clinicID, id uint64, ownerID *uint64) error {
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), id)
					linkedOwnerID = ownerID
					return nil
				},
			},
			ownerRepo: &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					return &model.Owner{ID: 200}, nil
				},
			},
		}
		svc.tryAutoLinkOwner(context.Background(), 3, 1, []byte(`{"owner_name":"田中太郎","phone":"090-0000-0000"}`))
		if assert.NotNil(t, linkedOwnerID) {
			assert.Equal(t, uint64(200), *linkedOwnerID)
		}
	})
}

// ================================================================
// tryAttachReservationOwnerPet: 直接単体テスト
// ================================================================

func TestLiffService_tryAttachReservationOwnerPet(t *testing.T) {
	t.Run("reservationRepo が nil なら no-op", func(t *testing.T) {
		svc := &liffService{reservationRepo: nil}
		appt := &model.Reservation{ID: 1}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		assert.Equal(t, uint64(1), appt.ID, "変更されないこと")
	})

	t.Run("appt が nil なら no-op", func(t *testing.T) {
		svc := &liffService{reservationRepo: &mockLiffReservationRepository{}}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, nil, nil)
	})

	t.Run("customerRepo.FindByID がエラー -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					called = true
					return nil, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return nil, errors.New("db error")
				},
			},
		}
		appt := &model.Reservation{ID: 1}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		assert.False(t, called)
	})

	t.Run("customer.OwnerID が nil -> スキップ", func(t *testing.T) {
		called := false
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					called = true
					return nil, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
			},
		}
		appt := &model.Reservation{ID: 1}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		assert.False(t, called)
	})

	t.Run("Update がエラー -> best-effortで appt は変更されない", func(t *testing.T) {
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					return nil, errors.New("db error")
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1, OwnerID: ptrUint64(200), Owner: &model.Owner{ID: 200}}, nil
				},
			},
		}
		appt := &model.Reservation{ID: 1}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		assert.Nil(t, appt.OwnerID, "更新失敗時は appt が変更されないこと")
	})

	t.Run("汚染された異院OwnerIDがscopeでOwner nilになった場合は予約へ再付与しない", func(t *testing.T) {
		called := false
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					called = true
					return nil, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1, OwnerID: ptrUint64(999), Owner: nil}, nil
				},
			},
		}
		appt := &model.Reservation{ID: 77}

		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)

		assert.False(t, called)
		assert.Nil(t, appt.OwnerID)
	})

	t.Run("成功: pet_id が解決できない場合でも owner_id のみ反映", func(t *testing.T) {
		var capturedFields map[string]any
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
					capturedFields = fields
					return &model.Reservation{ID: id, ClinicID: clinicID, OwnerID: ptrUint64(200)}, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1, OwnerID: ptrUint64(200), Owner: &model.Owner{ID: 200}}, nil
				},
			},
		}
		appt := &model.Reservation{ID: 77}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		_, hasPetID := capturedFields["pet_id"]
		assert.False(t, hasPetID)
		assert.Equal(t, uint64(200), capturedFields["owner_id"])
		if assert.NotNil(t, appt.OwnerID) {
			assert.Equal(t, uint64(200), *appt.OwnerID)
		}
	})

	t.Run("成功: pet_id も解決され反映される", func(t *testing.T) {
		svc := &liffService{
			reservationRepo: &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Reservation, error) {
					return &model.Reservation{ID: id, ClinicID: clinicID, OwnerID: ptrUint64(200), PetID: ptrUint64(300)}, nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{
						ID: 1, OwnerID: ptrUint64(200),
						Owner: &model.Owner{ID: 200, Pets: []model.Pet{{ID: 300, Name: "ポチ"}}},
					}, nil
				},
			},
		}
		appt := &model.Reservation{ID: 77}
		svc.tryAttachReservationOwnerPet(context.Background(), 3, 1, appt, nil)
		if assert.NotNil(t, appt.PetID) {
			assert.Equal(t, uint64(300), *appt.PetID)
		}
	})
}
