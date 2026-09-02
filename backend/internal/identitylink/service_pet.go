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
		group, members, err := s.createPetGroupInTx(txCtx, actor, ownerGroupID, refs)
		if err != nil {
			return err
		}
		outGroup = group
		outMembers = members
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
