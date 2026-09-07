package auth

import (
	"github.com/animal-ekarte/backend/internal/model"
)

type logoutAuditIdentity struct {
	staffID  uint64
	clinicID uint64
	valid    bool
}

type refreshTokenRevocation struct {
	familyID string
}

const maxRefreshTokenCookieAggregateBytes = 32 << 10

func (h *HTTPHandler) authService() Service {
	if h.deps.Auth != nil {
		return h.deps.Auth
	}
	return NewService(h.deps.Accounts, h.deps.Staff, h.deps.EffectivePermissions)
}

// ResolveClinicInfo returns the main clinic and all assigned clinic IDs.
func ResolveClinicInfo(
	assignments []model.StaffClinicAssignment,
) (mainClinicID string, clinicIDs []uint64) {
	return NewService(nil, nil, nil).ResolveClinicInfo(assignments)
}

// ResolveSystemAdminMainClinicID keeps an active preferred clinic or falls
// back to the first active, nonzero clinic for a system administrator.
func ResolveSystemAdminMainClinicID(
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) string {
	return NewService(nil, nil, nil).
		ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
}

// AuditClinicIDFromAssignments chooses the main clinic, falling back to the first assignment.
func AuditClinicIDFromAssignments(assignments []model.StaffClinicAssignment) (uint64, bool) {
	if len(assignments) == 0 {
		return 0, false
	}
	fallback := assignments[0].ClinicID
	for i := range assignments {
		if assignments[i].IsMain {
			return assignments[i].ClinicID, true
		}
	}
	return fallback, true
}
