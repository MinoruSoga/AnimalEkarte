package identitylink

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *service) CreateOwnerGroup(
	ctx context.Context,
	actor ActorContext,
	members []OwnerMemberRef,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	refs, err := normalizeOwnerRefs(members)
	if err != nil {
		return nil, nil, err
	}
	if len(refs) < 2 {
		return nil, nil, apperrors.WrapInvalidInput("owner identity group requires at least 2 members")
	}
	if err := assertOwnerRefsInActorScope(actor, refs); err != nil {
		return nil, nil, err
	}
	if s.audit == nil {
		return nil, nil, apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	var outGroup *model.OwnerIdentityGroup
	var outMembers []model.OwnerIdentityGroupMember

	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		group, members, err := s.createOwnerGroupInTx(txCtx, actor, refs)
		if err != nil {
			return err
		}
		outGroup = group
		outMembers = members
		return nil
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create owner identity group")
	}
	return outGroup, outMembers, nil
}

func (s *service) createOwnerGroupInTx(
	txCtx context.Context,
	actor ActorContext,
	refs []OwnerMemberRef,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	owners, lockErr := s.repo.LockOwners(txCtx, refs)
	if lockErr != nil {
		// hidden / nonexistent / cross-clinic missing → reject whole request, zero writes
		return nil, nil, lockErr
	}
	if len(owners) != len(refs) {
		return nil, nil, apperrors.WrapForbidden("mixed or hidden owner ids rejected")
	}

	// Idempotent retry: if all members already share one active group and
	// that group has exactly those members, return it without new writes.
	existingGroupIDs := map[uint64]struct{}{}
	for _, ref := range refs {
		m, findErr := s.repo.FindActiveOwnerMembership(txCtx, ref.ClinicID, ref.OwnerID)
		if findErr != nil {
			return nil, nil, findErr
		}
		if m == nil {
			existingGroupIDs = nil
			break
		}
		existingGroupIDs[m.GroupID] = struct{}{}
	}
	if existingGroupIDs != nil && len(existingGroupIDs) == 1 {
		var groupID uint64
		for id := range existingGroupIDs {
			groupID = id
		}
		group, lockGErr := s.repo.LockOwnerGroupByID(txCtx, groupID)
		if lockGErr != nil {
			return nil, nil, lockGErr
		}
		active, listErr := s.repo.ListActiveOwnerMembers(txCtx, groupID)
		if listErr != nil {
			return nil, nil, listErr
		}
		if sameOwnerMemberSet(active, refs) {
			return group, filterOwnerMembersByClinics(active, actor.VerifiedClinics), nil
		}
	}

	// Any member already in a different group → conflict (no partial merge in Phase 1).
	for _, ref := range refs {
		m, findErr := s.repo.FindActiveOwnerMembership(txCtx, ref.ClinicID, ref.OwnerID)
		if findErr != nil {
			return nil, nil, findErr
		}
		if m != nil {
			return nil, nil, apperrors.WrapConflict("owner already linked in another identity group")
		}
	}

	createdClinicID := pickCreatedClinicID(actor, refs[0].ClinicID)
	group := &model.OwnerIdentityGroup{
		CreatedClinicID: createdClinicID,
		Version:         1,
	}
	if createErr := s.repo.CreateOwnerGroup(txCtx, group); createErr != nil {
		return nil, nil, createErr
	}
	memberRows := make([]model.OwnerIdentityGroupMember, 0, len(refs))
	for _, ref := range refs {
		memberRows = append(memberRows, model.OwnerIdentityGroupMember{
			GroupCreatedClinicID: group.CreatedClinicID,
			GroupID:              group.ID,
			ClinicID:             ref.ClinicID,
			OwnerID:              ref.OwnerID,
		})
	}
	if createErr := s.repo.CreateOwnerMembers(txCtx, memberRows); createErr != nil {
		return nil, nil, createErr
	}
	if auditErr := s.writeAudit(txCtx, actor, model.AuditActionOwnerIdentityLinkCreate, group.ID, nil, map[string]any{
		"group_id":          group.ID,
		"created_clinic_id": group.CreatedClinicID,
		"members":           ownerRefsAudit(refs),
	}); auditErr != nil {
		return nil, nil, auditErr
	}
	return group, memberRows, nil
}

func (s *service) AddOwnerMembers(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
	members []OwnerMemberRef,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	refs, err := normalizeOwnerRefs(members)
	if err != nil {
		return nil, nil, err
	}
	if len(refs) == 0 {
		return nil, nil, apperrors.WrapInvalidInput("members required")
	}
	if err := assertOwnerRefsInActorScope(actor, refs); err != nil {
		return nil, nil, err
	}
	if s.audit == nil {
		return nil, nil, apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	var outGroup *model.OwnerIdentityGroup
	var outMembers []model.OwnerIdentityGroupMember

	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		group, lockErr := s.repo.LockOwnerGroupByID(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		// Actor must belong to created_clinic_id or any existing member clinic + new member clinics.
		existing, listErr := s.repo.ListActiveOwnerMembers(txCtx, groupID)
		if listErr != nil {
			return listErr
		}
		if err := assertCanManageOwnerGroup(actor, group, existing, refs); err != nil {
			return err
		}

		if _, lockOwnersErr := s.repo.LockOwners(txCtx, refs); lockOwnersErr != nil {
			return lockOwnersErr
		}

		toCreate := make([]model.OwnerIdentityGroupMember, 0, len(refs))
		for _, ref := range refs {
			m, findErr := s.repo.FindActiveOwnerMembership(txCtx, ref.ClinicID, ref.OwnerID)
			if findErr != nil {
				return findErr
			}
			if m != nil {
				if m.GroupID == groupID {
					// idempotent retry — already a member
					continue
				}
				return apperrors.WrapConflict("owner already linked in another identity group")
			}
			toCreate = append(toCreate, model.OwnerIdentityGroupMember{
				GroupCreatedClinicID: group.CreatedClinicID,
				GroupID:              group.ID,
				ClinicID:             ref.ClinicID,
				OwnerID:              ref.OwnerID,
			})
		}
		if len(toCreate) > 0 {
			if createErr := s.repo.CreateOwnerMembers(txCtx, toCreate); createErr != nil {
				return createErr
			}
			if auditErr := s.writeAudit(txCtx, actor, model.AuditActionOwnerIdentityLinkAdd, group.ID, nil, map[string]any{
				"group_id": group.ID,
				"members":  ownerRefsAudit(refsFromOwnerMembers(toCreate)),
			}); auditErr != nil {
				return auditErr
			}
		}
		active, listErr := s.repo.ListActiveOwnerMembersByClinicIDs(txCtx, groupID, actor.VerifiedClinics)
		if listErr != nil {
			return listErr
		}
		outGroup = group
		outMembers = active
		return nil
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to add owner identity members")
	}
	return outGroup, outMembers, nil
}

func (s *service) UnlinkOwnerMember(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
	member OwnerMemberRef,
) error {
	if err := requireActorClinics(actor); err != nil {
		return err
	}
	if member.ClinicID == 0 || member.OwnerID == 0 {
		return apperrors.WrapInvalidInput("clinic_id and owner_id required")
	}
	if !containsUint64(actor.VerifiedClinics, member.ClinicID) {
		return apperrors.WrapForbidden("member clinic outside actor scope")
	}
	if s.audit == nil {
		return apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		group, lockErr := s.repo.LockOwnerGroupByID(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		existing, listErr := s.repo.ListActiveOwnerMembers(txCtx, groupID)
		if listErr != nil {
			return listErr
		}
		if err := assertCanManageOwnerGroup(actor, group, existing, []OwnerMemberRef{member}); err != nil {
			return err
		}

		var target *model.OwnerIdentityGroupMember
		for i := range existing {
			if existing[i].ClinicID == member.ClinicID && existing[i].OwnerID == member.OwnerID {
				target = &existing[i]
				break
			}
		}
		if target == nil {
			return apperrors.WrapNotFound("owner_identity_group_member", fmt.Sprintf("%d/%d", member.ClinicID, member.OwnerID))
		}
		if delErr := s.repo.SoftDeleteOwnerMember(txCtx, target.ID); delErr != nil {
			return delErr
		}
		remaining, countErr := s.repo.CountActiveOwnerMembers(txCtx, groupID)
		if countErr != nil {
			return countErr
		}
		groupSoftDeleted := false
		if remaining == 0 {
			if delGroupErr := s.repo.SoftDeleteOwnerGroup(txCtx, groupID); delGroupErr != nil {
				return delGroupErr
			}
			groupSoftDeleted = true
		}
		return s.writeAudit(txCtx, actor, model.AuditActionOwnerIdentityLinkUnlink, groupID, map[string]any{
			"clinic_id": member.ClinicID,
			"owner_id":  member.OwnerID,
		}, map[string]any{
			"group_id":           groupID,
			"group_soft_deleted": groupSoftDeleted,
			"clinic_id":          member.ClinicID,
			"owner_id":           member.OwnerID,
		})
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to unlink owner identity member")
	}
	return nil
}
