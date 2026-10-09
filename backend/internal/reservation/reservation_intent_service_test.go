package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---------------------------------------------------------------------------
// Codex PR #606 Session A: ConfirmTentative service 層の回帰テスト
//   #4  intent advisory lock → 行ロックの順序強制
//   #15 無効化された予約区分の拒否
//   #6  非 general placeholder 区分は明示指定なしに確定させない
//   #16 譲渡後の現飼い主追従（新旧 owner のタグ再同期）
//   #3  確定時の死亡ペット拒否（行ロック保持中の検証）
// ---------------------------------------------------------------------------

// buildConfirmTentativeSvc は ConfirmTentative の service テスト用の組み立て口。
// tentative 行・区分・ロック呼出を差し替えられるようにする。
func buildConfirmTentativeSvc(
	repo *mockReservationRepository,
	typeFinder reservationTypeFinder,
	syncer NextVisitTagSyncer,
) ReservationService {
	return NewReservationServiceWithClinicHolidays(
		repo,
		typeFinder,
		&mockTransactor{},
		nil,
		&mockUnavailableTimeRepository{},
		nil,
		nil,
		&mockClinicHolidayFinder{},
		syncer,
	)
}

func confirmTentativeFixture(petID, ownerID, typeID uint64) *model.Reservation {
	sentinelStart := time.Date(2030, 6, 15, 0, 0, 0, 0, config.JST)
	return &model.Reservation{
		ID:                42,
		ClinicID:          1,
		PetID:             &petID,
		OwnerID:           &ownerID,
		Status:            model.ReservationStatusTentative,
		StartTime:         sentinelStart,
		EndTime:           sentinelStart.Add(nextVisitIntentDurationMinutes * time.Minute),
		ReservationTypeID: typeID,
		Source:            model.ReservationSourceManual,
	}
}

func alivePetFinder(petID, ownerID uint64) func(ctx context.Context, clinicID, pid uint64) (*model.Pet, error) {
	return func(_ context.Context, _, _ uint64) (*model.Pet, error) {
		return &model.Pet{ID: petID, OwnerID: ownerID}, nil
	}
}

func TestReservationService_ConfirmTentative_AcquiresIntentLockBeforeRowLock(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)

	var order []string
	repo := &mockReservationRepository{
		acquireBookingLockFn: func(_ context.Context, _ uint64) error {
			order = append(order, "booking_lock")
			return nil
		},
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) {
			order = append(order, "read")
			return &petID, nil
		},
		acquireNextVisitIntentLockFn: func(_ context.Context, _, _ uint64) error {
			order = append(order, "intent_lock")
			return nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			order = append(order, "row_lock")
			return tentative, nil
		},
		findPetByIDInClinicFn: alivePetFinder(petID, ownerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			order = append(order, "confirm")
			return tentative, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	intentIdx, rowIdx := -1, -1
	for i, step := range order {
		switch step {
		case "intent_lock":
			if intentIdx == -1 {
				intentIdx = i
			}
		case "row_lock":
			if rowIdx == -1 {
				rowIdx = i
			}
		}
	}
	require.NotEqual(t, -1, intentIdx, "(clinic,pet) intent lock が取得されていない")
	require.NotEqual(t, -1, rowIdx, "行ロックが取得されていない")
	assert.Less(t, intentIdx, rowIdx,
		"Codex #4: lock 順序は intent advisory lock → 行ロック（逆順だと set/clear との AB-BA デッドロックになる）: %v", order)
}

func TestReservationService_ConfirmTentative_RejectsInactiveReservationType(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		findPetByIDInClinicFn:    alivePetFinder(petID, ownerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			t.Fatal("inactive 区分の confirm は repo 書込へ到達してはならない")
			return nil, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: false}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "Codex #15: 無効化区分の確定は InvalidInput, 実際: %v", err)
	assert.Nil(t, got)
}

func TestReservationService_ConfirmTentative_PlaceholderTypeRequiresExplicitOverride(t *testing.T) {
	petID, ownerID := uint64(22), uint64(10)
	start := time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)

	newSvc := func(tentativeType *model.ReservationType, onConfirm func(ConfirmTentativeUpdate)) ReservationService {
		tentative := confirmTentativeFixture(petID, ownerID, tentativeType.ID)
		repo := &mockReservationRepository{
			findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
			lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
			findPetByIDInClinicFn:    alivePetFinder(petID, ownerID),
			confirmTentativeFn: func(_ context.Context, _, _ uint64, input ConfirmTentativeUpdate) (*model.Reservation, error) {
				if onConfirm != nil {
					onConfirm(input)
				}
				return tentative, nil
			},
		}
		typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			if id == tentativeType.ID {
				return tentativeType, nil
			}
			return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
		}}
		return buildConfirmTentativeSvc(repo, typeFinder, nil)
	}

	t.Run("rejects confirm when the stored placeholder is not general and no type is given", func(t *testing.T) {
		trimmingPlaceholder := &model.ReservationType{
			ID: 99, ClinicID: 1, Name: "migration placeholder",
			Category: model.ReservationTypeCategoryTrimming, IsActive: true,
		}
		svc := newSvc(trimmingPlaceholder, func(ConfirmTentativeUpdate) {
			t.Fatal("placeholder 維持の confirm は repo 書込へ到達してはならない")
		})

		got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
			StartTime: start,
			EndTime:   end,
			// ReservationTypeID 未指定 = stored placeholder 維持
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "Codex #6: 非 general placeholder の黙認昇格は InvalidInput, 実際: %v", err)
		assert.Nil(t, got)
	})

	t.Run("confirms when an explicit general type overrides the placeholder", func(t *testing.T) {
		trimmingPlaceholder := &model.ReservationType{
			ID: 99, ClinicID: 1, Name: "migration placeholder",
			Category: model.ReservationTypeCategoryTrimming, IsActive: true,
		}
		var captured ConfirmTentativeUpdate
		svc := newSvc(trimmingPlaceholder, func(input ConfirmTentativeUpdate) {
			captured = input
		})

		explicitTypeID := uint64(7)
		got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
			StartTime:         start,
			EndTime:           end,
			ReservationTypeID: &explicitTypeID,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, explicitTypeID, captured.ReservationTypeID,
			"明示された区分が repo 更新へ渡る")
	})
}

func TestReservationService_ConfirmTentative_RejectsDeceasedPetPostLock(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)
	deceasedAt := time.Now().Add(-time.Hour)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		// Codex #3: 死亡は行ロック保持中（intent lock 後）の検証で拒否される。
		findPetByIDInClinicFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: petID, OwnerID: ownerID, Status: model.PetStatusDeceased, DeceasedAt: &deceasedAt}, nil
		},
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			t.Fatal("死亡ペットの confirm は repo 書込へ到達してはならない")
			return nil, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "死亡ペットの確定は InvalidInput, 実際: %v", err)
	assert.Nil(t, got)
}

func TestReservationService_ConfirmTentative_SyncsPreviousAndCurrentOwnerAfterTransfer(t *testing.T) {
	petID, prevOwnerID, newOwnerID, typeID := uint64(22), uint64(10), uint64(20), uint64(7)
	tentative := confirmTentativeFixture(petID, prevOwnerID, typeID)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		findPetByIDInClinicFn:    alivePetFinder(petID, newOwnerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			// repo 側（ロック保持中）が owner_id を現飼い主へ追従した結果を再現する。
			confirmed := *tentative
			confirmed.Status = model.ReservationStatusConfirmed
			confirmed.OwnerID = &newOwnerID
			return &confirmed, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	syncer := &recordingNextVisitTagSyncer{}
	svc := buildConfirmTentativeSvc(repo, typeFinder, syncer)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.OwnerID)
	assert.Equal(t, newOwnerID, *got.OwnerID)

	synced := map[uint64]bool{}
	for _, ownerID := range syncer.calls {
		synced[ownerID] = true
	}
	assert.True(t, synced[newOwnerID], "Codex #16: 現飼い主の next_visit タグが再同期される")
	assert.True(t, synced[prevOwnerID], "Codex #16: 譲渡元（旧飼い主）の stale タグも除去される")
}
