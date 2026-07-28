package billing

// estimate_fk_clinic_isolation_test.go — AUD-005
// Create の medical_record_id / owner_id clinic 所有確認・MR-Owner 整合・created_by 所属検証・拒否時非副作用を検証する。
// Callers: go test ./internal/service/ -run 'TestEstimateService_.*(CrossClinic|CreatedBy|Owner|MedicalRecord)'
// Existing equivalent: none (Glob *estimate*isolation* = 0). No external data files; mock-only.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const defaultEstimateCreatedByStaffID = uint64(42)

type mockStaffClinicMembershipCounter struct {
	countFn func(ctx context.Context, staffID, clinicID uint64) (int64, error)
}

func (m *mockStaffClinicMembershipCounter) CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, staffID, clinicID)
	}
	return 1, nil
}

func estimateCreateBaseInput() *CreateEstimateInput {
	return &CreateEstimateInput{
		Title:       "見積",
		Subtotal:    1000,
		TaxTotal:    100,
		TotalAmount: 1100,
		CreatedBy:   ptrU64(defaultEstimateCreatedByStaffID),
	}
}

func TestEstimateService_Create_RejectsNilCreatedBy(t *testing.T) {
	const clinicID = uint64(1)
	created := false
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, _ *model.Estimate) error {
			created = true
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, &mockStaffClinicMembershipCounter{}, nil, noopTransactor{})
	in := estimateCreateBaseInput()
	in.CreatedBy = nil
	out, err := svc.Create(context.Background(), clinicID, in)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "want InvalidInput, got %v", err)
	assert.Nil(t, out)
	assert.False(t, created)
}

func TestEstimateService_Create_RejectsForeignClinicCreatedBy(t *testing.T) {
	const clinicID = uint64(1)
	const foreignStaffID = uint64(99)
	created := false
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, _ *model.Estimate) error {
			created = true
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, &mockStaffClinicMembershipCounter{
		countFn: func(_ context.Context, staffID, gotClinicID uint64) (int64, error) {
			if staffID == foreignStaffID && gotClinicID == clinicID {
				return 0, nil
			}
			return 1, nil
		},
	}, nil, noopTransactor{})
	in := estimateCreateBaseInput()
	in.CreatedBy = ptrU64(foreignStaffID)
	out, err := svc.Create(context.Background(), clinicID, in)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
	assert.Nil(t, out)
	assert.False(t, created)
}

func TestEstimateService_Create_RejectsCrossClinicMedicalRecordAndOwner(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMRID = uint64(100)
	const foreignMRID = uint64(901)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const otherOwnerID = uint64(11)

	linkAware := func(created *bool) EstimateService {
		repo := &mockEstimateRepository{
			createFn: func(_ context.Context, e *model.Estimate) error {
				*created = true
				e.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
				return &model.Estimate{ID: id, ClinicID: clinicID, Title: "見積"}, nil
			},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.MedicalRecord, error) {
				if gotClinicID != clinicID || id == foreignMRID {
					return nil, apperrors.WrapNotFound("medical_record", "foreign")
				}
				return &model.MedicalRecord{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ptrU64(ownedOwnerID),
				}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
		}
		return NewEstimateService(repo, mrRepo, resRepo, &mockStaffClinicMembershipCounter{}, nil, noopTransactor{})
	}

	base := estimateCreateBaseInput

	t.Run("rejects cross-clinic medical_record_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.MedicalRecordID = ptrU64(foreignMRID)
		out, err := svc.Create(context.Background(), clinicID, in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic owner_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.OwnerID = ptrU64(foreignOwnerID)
		out, err := svc.Create(context.Background(), clinicID, in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects medical_record owner mismatch and does not persist", func(t *testing.T) {
		created := false
		repo := &mockEstimateRepository{
			createFn: func(_ context.Context, _ *model.Estimate) error {
				created = true
				return nil
			},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ptrU64(ownedOwnerID),
				}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID != ownedOwnerID && ownerID != otherOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
		}
		svc := NewEstimateService(repo, mrRepo, resRepo, &mockStaffClinicMembershipCounter{}, nil, noopTransactor{})
		in := base()
		in.MedicalRecordID = ptrU64(ownedMRID)
		in.OwnerID = ptrU64(otherOwnerID)
		out, err := svc.Create(context.Background(), clinicID, in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic matching related FKs", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.MedicalRecordID = ptrU64(ownedMRID)
		in.OwnerID = ptrU64(ownedOwnerID)
		out, err := svc.Create(context.Background(), clinicID, in)
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestEstimateService_Create_PersistsCreatedByFromInput(t *testing.T) {
	const clinicID = uint64(1)
	const staffID = uint64(42)

	var persisted *uint64
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			persisted = e.CreatedBy
			e.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, ClinicID: clinicID, Title: "見積", CreatedBy: ptrU64(staffID)}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, &mockStaffClinicMembershipCounter{
		countFn: func(_ context.Context, gotStaffID, gotClinicID uint64) (int64, error) {
			if gotStaffID == staffID && gotClinicID == clinicID {
				return 1, nil
			}
			return 0, nil
		},
	}, nil, noopTransactor{})
	out, err := svc.Create(context.Background(), clinicID, &CreateEstimateInput{
		Title:       "見積",
		Subtotal:    1000,
		TaxTotal:    100,
		TotalAmount: 1100,
		CreatedBy:   ptrU64(staffID),
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, persisted)
	assert.Equal(t, staffID, *persisted)
}
