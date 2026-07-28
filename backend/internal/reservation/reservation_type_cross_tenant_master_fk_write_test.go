package reservation

// reservation_type_cross_tenant_master_fk_write_test.go — BE9-2C R①:
// service/cross_tenant_master_fk_write_test.go から reservationTypeService 節を同名移動。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationTypeService_Create_RejectsCrossClinicGroupID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedGroupID = uint64(20)
	const foreignGroupID = uint64(999)

	groupRepoFor := func() ReservationTypeGroupRepository {
		return &mockReservationTypeGroupRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationTypeGroup, error) {
				if gotClinicID != clinicID || id != ownedGroupID {
					return nil, apperrors.WrapNotFound("reservation_type_group", "foreign")
				}
				return &model.ReservationTypeGroup{ID: id, ClinicID: clinicID}, nil
			},
		}
	}

	newSvc := func(created *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, _ *model.ReservationType) error {
				*created = true
				return nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, groupRepoFor())
	}

	t.Run("rejects cross-clinic group_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignGroupID
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationTypeInput{Name: "診察", GroupID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "reservation type must NOT be persisted referencing another clinic's group")
	})

	t.Run("accepts same-clinic group_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedGroupID
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationTypeInput{Name: "診察", GroupID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationTypeService_Update_RejectsCrossClinicGroupID(t *testing.T) {
	const clinicID = uint64(1)
	const reservationTypeID = uint64(5)
	const ownedGroupID = uint64(20)
	const foreignGroupID = uint64(999)

	groupRepoFor := func() ReservationTypeGroupRepository {
		return &mockReservationTypeGroupRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationTypeGroup, error) {
				if gotClinicID != clinicID || id != ownedGroupID {
					return nil, apperrors.WrapNotFound("reservation_type_group", "foreign")
				}
				return &model.ReservationTypeGroup{ID: id, ClinicID: clinicID}, nil
			},
		}
	}

	newSvc := func(updated *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
			updateFn: func(_ context.Context, gotClinicID, id uint64, _ map[string]any) (*model.ReservationType, error) {
				*updated = true
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, groupRepoFor())
	}

	t.Run("rejects cross-clinic group_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignGroupID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{GroupID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "reservation type must NOT be updated to reference another clinic's group")
	})

	t.Run("accepts same-clinic group_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedGroupID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{GroupID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// TestReservationTypeService_Update_RejectsCrossClinicParentID は ParentID の
// 所有権ガード(validateReservationTypeParent、既存実装)にクロステナント isolation の
// 実証テストを追加する(X-14: 「実装済みだが証拠不足」パターン。GroupID と対称)。
func TestReservationTypeService_Update_RejectsCrossClinicParentID(t *testing.T) {
	const clinicID = uint64(1)
	const reservationTypeID = uint64(5)
	const ownedParentID = uint64(30)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
				if id == reservationTypeID {
					return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
				}
				if gotClinicID != clinicID || id != ownedParentID {
					return nil, apperrors.WrapNotFound("reservation_type", "foreign")
				}
				return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, gotClinicID, id uint64, _ map[string]any) (*model.ReservationType, error) {
				*updated = true
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "reservation type must NOT be updated to reference another clinic's parent")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── staffService (X-14 batch U7): OccupationID ──
//
// staffService.Create/CreateWithAccount/Update persisted a request-derived OccupationID
// without verifying it belongs to the caller's clinic. Guard: occupationRepo.FindByID
// (ctx, clinicID, OccupationID) before persist, mirroring medicineService's
// validateInventoryOwnership (X-14 batch U2). occupationRepo is now a mandatory
// NewStaffService dependency (see staff_service_core.go validateOccupationOwnership).
