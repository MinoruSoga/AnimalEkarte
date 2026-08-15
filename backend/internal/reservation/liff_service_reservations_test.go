package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
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
	_, ok := IsReservationLimitError(err)
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
	input.StartTime = "invalid" // ToDateTime のパースを失敗させる

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), capturedStaffID, "日時パース失敗時は自動割当をスキップし staff_id=0 のまま渡る")
}

func TestLiffService_CreateReservation_TrimmingDetail_Success(t *testing.T) {
	var delegatedInput *CreateReservationInput
	var legacyDetailWriteCalled bool
	courseID := uint64(7)

	svc := newReservationTestSvc(func(s *liffService) {
		s.trimmingDetailRepo = &mockTrimmingDetailRepository{
			createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				legacyDetailWriteCalled = true
				return nil
			},
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				legacyDetailWriteCalled = true
				return nil
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
				delegatedInput = input
				return &model.Reservation{ID: 42, ClinicID: 3}, nil
			},
		}
	})

	input := reservationBaseInput()
	input.TrimmingCourseID = &courseID
	input.TrimmingOptionIDs = []uint64{1, 2}

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.NoError(t, err)
	require.NotNil(t, delegatedInput)
	if assert.NotNil(t, delegatedInput.TrimmingCourseID) {
		assert.Equal(t, courseID, *delegatedInput.TrimmingCourseID)
	}
	assert.Equal(t, []uint64{1, 2}, delegatedInput.TrimmingOptionIDs)
	assert.False(t, legacyDetailWriteCalled, "trimming writes belong to the validator transaction")
}

func TestLiffService_CreateReservation_TrimmingDetail_CreateErrorFailsReservation(t *testing.T) {
	var legacyDetailWriteCalled bool
	courseID := uint64(7)

	svc := newReservationTestSvc(func(s *liffService) {
		s.trimmingDetailRepo = &mockTrimmingDetailRepository{
			createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				legacyDetailWriteCalled = true
				return nil
			},
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				legacyDetailWriteCalled = true
				return nil
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, _ *CreateReservationInput) (*model.Reservation, error) {
				return nil, errors.New("detail insert failed")
			},
		}
	})

	input := reservationBaseInput()
	input.TrimmingCourseID = &courseID
	input.TrimmingOptionIDs = []uint64{1}

	_, err := svc.CreateReservation(context.Background(), 3, 1, input)
	require.Error(t, err, "trimming detail persistence failure must abort the reservation")
	assert.False(t, legacyDetailWriteCalled, "the service must not perform a second best-effort write")
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
	deceasedAtForResolve := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
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
		{
			name: "単頭だが死亡 -> nil",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{
				{ID: 300, Name: "ポチ", DeceasedAt: &deceasedAtForResolve},
			}}},
			wantNil: true,
		},
		{
			name: "生存1頭+死亡1頭 -> 生存ID",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{
				{ID: 300, Name: "ポチ", DeceasedAt: &deceasedAtForResolve},
				{ID: 301, Name: "タマ"},
			}}},
			wantID: 301,
		},
		{
			// 生存が2頭以上のときだけ名前解決する。死亡名への一致は skip して nil。
			name: "複数頭生存・名前一致だが死亡 -> nil",
			customer: &model.LineCustomer{ID: 1, Owner: &model.Owner{ID: 10, Pets: []model.Pet{
				{ID: 300, Name: "ポチ"},
				{ID: 301, Name: "タマ", DeceasedAt: &deceasedAtForResolve},
				{ID: 302, Name: "ハチ"},
			}}},
			customerFields: []byte(`{"pets":[{"name":"タマ"}]}`),
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
// SEC-CS2-F02: name+phone 自動オーナー紐付けの回帰テスト
// CreateReservation は攻撃者が入力した氏名+電話番号で
// line_customers.owner_id を書き込んではならない。
// ================================================================

func TestLiffService_tryAutoLinkOwner(t *testing.T) {
	t.Run("一致する氏名+電話番号でも owner-link 書き込みはゼロ", func(t *testing.T) {
		var (
			ownerLookupCalls  int
			updateOwnerCalls  int
			reservationUpdate int
		)
		svc := newReservationTestSvc(func(s *liffService) {
			s.customerRepo = &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					// 未紐付け顧客
					return &model.LineCustomer{ID: 1}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					updateOwnerCalls++
					return nil
				},
			}
			s.ownerRepo = &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					ownerLookupCalls++
					return &model.Owner{ID: 200}, nil
				},
			}
			s.reservationRepo = &mockLiffReservationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					reservationUpdate++
					return &model.Reservation{ID: 77}, nil
				},
			}
			s.validators = &mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					return &model.Reservation{ID: 77, ClinicID: input.ClinicID}, nil
				},
			}
		})

		input := reservationBaseInput()
		// 攻撃者が被害者の氏名+電話番号を customer_fields に載せて予約するシナリオ
		input.CustomerFields = []byte(`{"owner_name":"田中太郎","phone":"090-0000-0000","pets":[{"name":"ポチ"}]}`)

		got, err := svc.CreateReservation(context.Background(), 3, 1, input)
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, 0, ownerLookupCalls, "FindByNameAndPhone は呼ばれてはならない")
		assert.Equal(t, 0, updateOwnerCalls, "UpdateOwnerLink は呼ばれてはならない")
		assert.Equal(t, 0, reservationUpdate, "未紐付け顧客では予約への owner/pet 反映も行わない")
		assert.Nil(t, got.OwnerID)
		assert.Nil(t, got.PetID)
	})

	t.Run("customer_name フォールバックでも owner-link 書き込みはゼロ", func(t *testing.T) {
		var updateOwnerCalls int
		svc := newReservationTestSvc(func(s *liffService) {
			s.customerRepo = &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: 1}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					updateOwnerCalls++
					return nil
				},
			}
			s.ownerRepo = &mockLiffOwnerRepository{
				findByNameAndPhoneFn: func(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
					t.Fatal("FindByNameAndPhone は呼ばれてはならない")
					return &model.Owner{ID: 200}, nil
				},
			}
			s.validators = &mockLiffValidators{
				validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
					return &model.Reservation{ID: 78, ClinicID: input.ClinicID}, nil
				},
			}
		})

		input := reservationBaseInput()
		input.CustomerFields = []byte(`{"customer_name":"田中太郎","phone":"090-0000-0000"}`)

		_, err := svc.CreateReservation(context.Background(), 3, 1, input)
		require.NoError(t, err)
		assert.Equal(t, 0, updateOwnerCalls)
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
