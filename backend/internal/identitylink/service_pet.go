package identitylink

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *service) CreatePetGroup(
	ctx context.Context,
	actor ActorContext,
	ownerGroupID uint64,
	members []PetMemberRef,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	refs, err := normalizePetRefs(members)
	if err != nil {
		return nil, nil, err
	}
	if len(refs) < 2 {
		return nil, nil, apperrors.WrapInvalidInput("pet identity group requires at least 2 members")
	}
	if err := assertPetRefsInActorScope(actor, refs); err != nil {
		return nil, nil, err
	}
	if s.audit == nil {
		return nil, nil, apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	var outGroup *model.PetIdentityGroup
	var outMembers []model.PetIdentityGroupMember

	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		ownerGroup, lockOGErr := s.repo.LockOwnerGroupByID(txCtx, ownerGroupID)
		if lockOGErr != nil {
			return lockOGErr
		}
		ownerMembers, listErr := s.repo.ListActiveOwnerMembers(txCtx, ownerGroupID)
		if listErr != nil {
			return listErr
		}
		if len(ownerMembers) == 0 {
			return apperrors.WrapNotFound("owner_identity_group", fmt.Sprintf("%d", ownerGroupID))
		}
		// Actor must cover every parent-owner clinic (anchor + all active members)
		// and all pet clinics (pet refs already checked via assertPetRefsInActorScope).
		// No any-member fallback: missing one parent-owner clinic is Forbidden, zero writes.
		if err := assertActorCoversOwnerGroupClinics(actor, ownerGroup, ownerMembers); err != nil {
			return err
		}

		pets, lockPetsErr := s.repo.LockPets(txCtx, refs)
		if lockPetsErr != nil {
			return lockPetsErr
		}
		if len(pets) != len(refs) {
			return apperrors.WrapForbidden("mixed or hidden pet ids rejected")
		}

		// Each pet's owner must be an active member of the owner group.
		for _, pet := range pets {
			ok, ownErr := s.repo.IsOwnerActiveInGroup(txCtx, ownerGroupID, pet.ClinicID, pet.OwnerID)
			if ownErr != nil {
				return ownErr
			}
			if !ok {
				return apperrors.WrapConflict("pet owner is not in the specified owner identity group")
			}
		}

		// Idempotent: all already in same pet group under this owner group.
		existingGroupIDs := map[uint64]struct{}{}
		for _, ref := range refs {
			m, findErr := s.repo.FindActivePetMembership(txCtx, ref.ClinicID, ref.PetID)
			if findErr != nil {
				return findErr
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
			group, lockGErr := s.repo.LockPetGroupByID(txCtx, groupID)
			if lockGErr != nil {
				return lockGErr
			}
			if group.OwnerGroupID != ownerGroupID {
				return apperrors.WrapConflict("pets already linked under a different owner identity group")
			}
			active, listErr := s.repo.ListActivePetMembers(txCtx, groupID)
			if listErr != nil {
				return listErr
			}
			if samePetMemberSet(active, refs) {
				outGroup = group
				outMembers = filterPetMembersByClinics(active, actor.VerifiedClinics)
				return nil
			}
		}

		for _, ref := range refs {
			m, findErr := s.repo.FindActivePetMembership(txCtx, ref.ClinicID, ref.PetID)
			if findErr != nil {
				return findErr
			}
			if m != nil {
				return apperrors.WrapConflict("pet already linked in another identity group")
			}
		}

		createdClinicID := pickCreatedClinicID(actor, refs[0].ClinicID)
		group := &model.PetIdentityGroup{
			CreatedClinicID:           createdClinicID,
			OwnerGroupCreatedClinicID: ownerGroup.CreatedClinicID,
			OwnerGroupID:              ownerGroup.ID,
			Version:                   1,
		}
		if createErr := s.repo.CreatePetGroup(txCtx, group); createErr != nil {
			return createErr
		}
		memberRows := make([]model.PetIdentityGroupMember, 0, len(refs))
		for _, ref := range refs {
			memberRows = append(memberRows, model.PetIdentityGroupMember{
				GroupCreatedClinicID: group.CreatedClinicID,
				GroupID:              group.ID,
				ClinicID:             ref.ClinicID,
				PetID:                ref.PetID,
			})
		}
		if createErr := s.repo.CreatePetMembers(txCtx, memberRows); createErr != nil {
			return createErr
		}
		if auditErr := s.writeAudit(txCtx, actor, model.AuditActionPetIdentityLinkCreate, group.ID, nil, map[string]any{
			"group_id":       group.ID,
			"owner_group_id": ownerGroupID,
			"members":        petRefsAudit(refs),
		}); auditErr != nil {
			return auditErr
		}
		outGroup = group
		outMembers = memberRows
		return nil
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create pet identity group")
	}
	return outGroup, outMembers, nil
}

func (s *service) AddPetMembers(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
	members []PetMemberRef,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	refs, err := normalizePetRefs(members)
	if err != nil {
		return nil, nil, err
	}
	if len(refs) == 0 {
		return nil, nil, apperrors.WrapInvalidInput("members required")
	}
	if err := assertPetRefsInActorScope(actor, refs); err != nil {
		return nil, nil, err
	}
	if s.audit == nil {
		return nil, nil, apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	var outGroup *model.PetIdentityGroup
	var outMembers []model.PetIdentityGroupMember

	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		group, lockErr := s.repo.LockPetGroupByID(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		existing, listErr := s.repo.ListActivePetMembers(txCtx, groupID)
		if listErr != nil {
			return listErr
		}
		if err := assertCanManagePetGroup(actor, group, existing, refs); err != nil {
			return err
		}

		pets, lockPetsErr := s.repo.LockPets(txCtx, refs)
		if lockPetsErr != nil {
			return lockPetsErr
		}
		for _, pet := range pets {
			ok, ownErr := s.repo.IsOwnerActiveInGroup(txCtx, group.OwnerGroupID, pet.ClinicID, pet.OwnerID)
			if ownErr != nil {
				return ownErr
			}
			if !ok {
				return apperrors.WrapConflict("pet owner is not in the parent owner identity group")
			}
		}

		toCreate := make([]model.PetIdentityGroupMember, 0, len(refs))
		for _, ref := range refs {
			m, findErr := s.repo.FindActivePetMembership(txCtx, ref.ClinicID, ref.PetID)
			if findErr != nil {
				return findErr
			}
			if m != nil {
				if m.GroupID == groupID {
					continue
				}
				return apperrors.WrapConflict("pet already linked in another identity group")
			}
			toCreate = append(toCreate, model.PetIdentityGroupMember{
				GroupCreatedClinicID: group.CreatedClinicID,
				GroupID:              group.ID,
				ClinicID:             ref.ClinicID,
				PetID:                ref.PetID,
			})
		}
		if len(toCreate) > 0 {
			if createErr := s.repo.CreatePetMembers(txCtx, toCreate); createErr != nil {
				return createErr
			}
			if auditErr := s.writeAudit(txCtx, actor, model.AuditActionPetIdentityLinkAdd, group.ID, nil, map[string]any{
				"group_id": group.ID,
				"members":  petRefsAudit(refsFromPetMembers(toCreate)),
			}); auditErr != nil {
				return auditErr
			}
		}
		active, listErr := s.repo.ListActivePetMembersByClinicIDs(txCtx, groupID, actor.VerifiedClinics)
		if listErr != nil {
			return listErr
		}
		outGroup = group
		outMembers = active
		return nil
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to add pet identity members")
	}
	return outGroup, outMembers, nil
}

func (s *service) UnlinkPetMember(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
	member PetMemberRef,
) error {
	if err := requireActorClinics(actor); err != nil {
		return err
	}
	if member.ClinicID == 0 || member.PetID == 0 {
		return apperrors.WrapInvalidInput("clinic_id and pet_id required")
	}
	if !containsUint64(actor.VerifiedClinics, member.ClinicID) {
		return apperrors.WrapForbidden("member clinic outside actor scope")
	}
	if s.audit == nil {
		return apperrors.WrapInternalServerError("identity link audit logger is required")
	}

	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		group, lockErr := s.repo.LockPetGroupByID(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		existing, listErr := s.repo.ListActivePetMembers(txCtx, groupID)
		if listErr != nil {
			return listErr
		}
		if err := assertCanManagePetGroup(actor, group, existing, []PetMemberRef{member}); err != nil {
			return err
		}

		var target *model.PetIdentityGroupMember
		for i := range existing {
			if existing[i].ClinicID == member.ClinicID && existing[i].PetID == member.PetID {
				target = &existing[i]
				break
			}
		}
		if target == nil {
			return apperrors.WrapNotFound("pet_identity_group_member", fmt.Sprintf("%d/%d", member.ClinicID, member.PetID))
		}
		if delErr := s.repo.SoftDeletePetMember(txCtx, target.ID); delErr != nil {
			return delErr
		}
		remaining, countErr := s.repo.CountActivePetMembers(txCtx, groupID)
		if countErr != nil {
			return countErr
		}
		groupSoftDeleted := false
		if remaining == 0 {
			if delGroupErr := s.repo.SoftDeletePetGroup(txCtx, groupID); delGroupErr != nil {
				return delGroupErr
			}
			groupSoftDeleted = true
		}
		return s.writeAudit(txCtx, actor, model.AuditActionPetIdentityLinkUnlink, groupID, map[string]any{
			"clinic_id": member.ClinicID,
			"pet_id":    member.PetID,
		}, map[string]any{
			"group_id":           groupID,
			"group_soft_deleted": groupSoftDeleted,
			"clinic_id":          member.ClinicID,
			"pet_id":             member.PetID,
		})
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to unlink pet identity member")
	}
	return nil
}
