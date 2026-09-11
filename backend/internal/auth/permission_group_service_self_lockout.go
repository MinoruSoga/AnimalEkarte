package auth

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *permissionGroupService) guardActorKeepsPermissionAdministration(
	ctx context.Context,
	clinicID, actorStaffID uint64,
	actorIsSystemAdmin bool,
) error {
	if actorIsSystemAdmin {
		return nil
	}
	if actorStaffID == 0 {
		return apperrors.WrapInternalServerError("permission mutation actor is invalid")
	}
	rules, err := s.repo.FindAllEffectivePermissionsByStaffID(ctx, actorStaffID, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to resolve actor permission administration")
	}
	if !permissionRulesAllow(rules, string(model.ResourceMasterPermission), "view") ||
		!permissionRulesAllow(rules, string(model.ResourceMasterPermission), "edit") {
		return apperrors.WrapForbidden("cannot remove own permission administration")
	}
	return nil
}
