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
		locked, lockErr := s.lockPetGroupForUpdate(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		group := locked.petGroup
		existing := locked.petMembers
		ownerGroup := locked.ownerGroup
		ownerMembers := locked.ownerMembers
		if err := assertCanManagePetGroup(actor, group, existing, refs); err != nil {
			return err
		}
		decision, authErr := s.authorizeAffectedClinics(
			txCtx,
			actor,
			petAffectedClinicIDs(group.CreatedClinicID, existing, refs, ownerGroup, ownerMembers),
		)
		if authErr != nil {
			return authErr
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
			}, decision); auditErr != nil {
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
		locked, lockErr := s.lockPetGroupForUpdate(txCtx, groupID)
		if lockErr != nil {
			return lockErr
		}
		group := locked.petGroup
		existing := locked.petMembers
		ownerGroup := locked.ownerGroup
		ownerMembers := locked.ownerMembers
		if err := assertCanManagePetGroup(actor, group, existing, []PetMemberRef{member}); err != nil {
			return err
		}
		decision, authErr := s.authorizeAffectedClinics(
			txCtx,
			actor,
			petAffectedClinicIDs(group.CreatedClinicID, existing, []PetMemberRef{member}, ownerGroup, ownerMembers),
		)
		if authErr != nil {
			return authErr
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
		}, decision)
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to unlink pet identity member")
	}
	return nil
}

// petGroupLockContext carries the rows that AddPetMembers and UnlinkPetMember
// read under the shared lock order: owner group first, then pet group.
type petGroupLockContext struct {
	petGroup     *model.PetIdentityGroup
	ownerGroup   *model.OwnerIdentityGroup
	ownerMembers []model.OwnerIdentityGroupMember
	petMembers   []model.PetIdentityGroupMember
}

// lockPetGroupForUpdate locks the owner identity group before the pet identity
// group so AddPetMembers / UnlinkPetMember share the same lock order as
// createPetGroupInTx (owner group -> pet group -> pet rows). The pet group's
// OwnerGroupID is resolved by a non-locking pre-read; the locked pet group row
// is authoritative and a divergence fails closed with a conflict. The owner
// group lock still precedes authorizeAffectedClinics, keeping the FOR UPDATE
// read ahead of the pet-write authorization TOCTOU window.
func (s *service) lockPetGroupForUpdate(
	ctx context.Context,
	groupID uint64,
) (*petGroupLockContext, error) {
	seed, findErr := s.repo.FindPetGroupByID(ctx, groupID)
	if findErr != nil {
		return nil, findErr
	}
	// 認可境界となる owner group は FOR UPDATE で読み、かつ pet group のロック
	// より先に取得する。非ロック読みだと並行する owner 側の変更
	// （AddOwnerMembers 等）が割り込み、stale な affected-clinic 集合で認可が
	// 通る TOCTOU になり、ロック順序を逆転させると createPetGroupInTx との間で
	// デッドロックを起こし得る。
	ownerGroup, ownerErr := s.repo.LockOwnerGroupByID(ctx, seed.OwnerGroupID)
	if ownerErr != nil {
		return nil, ownerErr
	}
	ownerMembers, ownerListErr := s.repo.ListActiveOwnerMembers(ctx, seed.OwnerGroupID)
	if ownerListErr != nil {
		return nil, ownerListErr
	}
	group, lockErr := s.repo.LockPetGroupByID(ctx, groupID)
	if lockErr != nil {
		return nil, lockErr
	}
	if group.OwnerGroupID != ownerGroup.ID {
		return nil, apperrors.WrapConflict("pet identity group changed owner identity group")
	}
	petMembers, listErr := s.repo.ListActivePetMembers(ctx, groupID)
	if listErr != nil {
		return nil, listErr
	}
	return &petGroupLockContext{
		petGroup:     group,
		ownerGroup:   ownerGroup,
		ownerMembers: ownerMembers,
		petMembers:   petMembers,
	}, nil
}
