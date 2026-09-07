package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestPickCredentialAuditClinicID_PrefersMainActiveAssignment(t *testing.T) {
	clinics := []model.Clinic{
		{ID: 1, IsActive: false},
		{ID: 2, IsActive: true},
		{ID: 3, IsActive: true},
	}
	activeIDs, activeSet := ActiveCredentialAuditClinics(clinics)
	assert.Equal(t, []uint64{2, 3}, activeIDs)

	id, err := PickCredentialAuditClinicID(
		&model.Account{},
		9,
		[]model.StaffClinicAssignment{
			{StaffID: 9, ClinicID: 2, IsMain: false},
			{StaffID: 9, ClinicID: 3, IsMain: true},
		},
		activeIDs,
		activeSet,
	)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), id)
}

func TestPickCredentialAuditClinicID_SystemAdminFallsBackToFirstActive(t *testing.T) {
	activeIDs, activeSet := ActiveCredentialAuditClinics([]model.Clinic{
		{ID: 4, IsActive: true},
	})

	id, err := PickCredentialAuditClinicID(
		&model.Account{IsSystemAdmin: true},
		9,
		nil,
		activeIDs,
		activeSet,
	)
	require.NoError(t, err)
	assert.Equal(t, uint64(4), id)
}

func TestPickCredentialAuditClinicID_ForbiddenWithoutActiveClinic(t *testing.T) {
	_, err := PickCredentialAuditClinicID(
		&model.Account{},
		9,
		nil,
		nil,
		map[uint64]struct{}{},
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}
