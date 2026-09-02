package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const clinicPermissionCheckerContextKey = "clinic_permission_checker"

// ClinicPermissionChecker evaluates one resource/action pair for a specific clinic.
// It must not write an HTTP response; callers own the single error write.
type ClinicPermissionChecker func(c *gin.Context, clinicID uint64, resource, action string) bool

// SetClinicPermissionChecker stores the per-clinic RBAC evaluator on the request.
// RequirePermission attaches auth.HasPermissionInClinic so list/detail/write
// expansion can re-check every destination clinic instead of reusing the
// selected clinic's grant.
func SetClinicPermissionChecker(c *gin.Context, check ClinicPermissionChecker) {
	if check == nil {
		return
	}
	c.Set(clinicPermissionCheckerContextKey, check)
}

// PeekClinicPermissionChecker reads the per-clinic evaluator without writing a response.
func PeekClinicPermissionChecker(c *gin.Context) (ClinicPermissionChecker, bool) {
	val, exists := c.Get(clinicPermissionCheckerContextKey)
	if !exists {
		return nil, false
	}
	check, ok := val.(ClinicPermissionChecker)
	if !ok || check == nil {
		return nil, false
	}
	return check, true
}

func requireClinicPermissionChecker(c *gin.Context) (ClinicPermissionChecker, bool) {
	check, ok := PeekClinicPermissionChecker(c)
	if !ok {
		RespondError(c, apperrors.WrapForbidden("forbidden"))
		return nil, false
	}
	return check, true
}

func peekedSystemAdmin(c *gin.Context) bool {
	isAdmin, ok := PeekIsSystemAdmin(c)
	return ok && isAdmin
}

func compactPositiveClinicIDs(clinicIDs []uint64) []uint64 {
	compacted := make([]uint64, 0, len(clinicIDs))
	seen := make(map[uint64]struct{}, len(clinicIDs))
	for _, id := range clinicIDs {
		if id == 0 {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		compacted = append(compacted, id)
	}
	return compacted
}

// AuthorizeClinicIDsForPermission requires membership and the resource/action
// grant on every requested clinic. One miss is 403 (fail-closed, no partial write).
// system_admin is an explicit bypass after the membership/active-clinic subset check.
func AuthorizeClinicIDsForPermission(
	c *gin.Context,
	requested []uint64,
	resource, action string,
) bool {
	if !AuthorizeClinicIDs(c, requested) {
		return false
	}
	if peekedSystemAdmin(c) {
		return true
	}
	check, ok := requireClinicPermissionChecker(c)
	if !ok {
		return false
	}
	for _, id := range requested {
		if id == 0 || !check(c, id, resource, action) {
			RespondError(c, apperrors.WrapForbidden("forbidden"))
			return false
		}
	}
	return true
}

// FilterClinicIDsForPermission returns the subset of clinic IDs that have the
// required grant. Missing checker or an empty result is 403.
// system_admin keeps the compact membership set (explicit grant).
func FilterClinicIDsForPermission(
	c *gin.Context,
	clinicIDs []uint64,
	resource, action string,
) ([]uint64, bool) {
	compacted := compactPositiveClinicIDs(clinicIDs)
	if peekedSystemAdmin(c) {
		if len(compacted) == 0 {
			RespondError(c, apperrors.WrapForbidden("forbidden"))
			return nil, false
		}
		return compacted, true
	}
	check, ok := requireClinicPermissionChecker(c)
	if !ok {
		return nil, false
	}
	filtered := make([]uint64, 0, len(compacted))
	for _, id := range compacted {
		if check(c, id, resource, action) {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		RespondError(c, apperrors.WrapForbidden("forbidden"))
		return nil, false
	}
	return filtered, true
}

// ResolveListClinicIDsForPermission resolves #86 list scope then keeps only
// clinics that hold the required grant. Explicit unassigned IDs remain 403.
func ResolveListClinicIDsForPermission(
	c *gin.Context,
	resource, action string,
) ([]uint64, bool) {
	clinicIDs, ok := ResolveListClinicIDs(c)
	if !ok {
		return nil, false
	}
	return FilterClinicIDsForPermission(c, clinicIDs, resource, action)
}

// ResolveAllClinicIDsForPermission resolves #86 detail scope then keeps only
// clinics that hold the required grant. system_admin still returns the selected
// clinic only (full-scan prevention in ResolveAllClinicIDs).
func ResolveAllClinicIDsForPermission(
	c *gin.Context,
	resource, action string,
) ([]uint64, bool) {
	clinicIDs, ok := ResolveAllClinicIDs(c)
	if !ok {
		return nil, false
	}
	return FilterClinicIDsForPermission(c, clinicIDs, resource, action)
}
