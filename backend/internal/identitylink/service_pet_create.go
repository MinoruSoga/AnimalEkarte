package identitylink

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *service) createPetGroupInTx(
	txCtx context.Context,
	actor ActorContext,
	ownerGroupID uint64,
	refs []PetMemberRef,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	ownerGroup, lockOGErr := s.repo.LockOwnerGroupByID(txCtx, ownerGroupID)
	if lockOGErr != nil {
		return nil, nil, lockOGErr
	}
	ownerMembers, listErr := s.repo.ListActiveOwnerMembers(txCtx, ownerGroupID)
	if listErr != nil {
		return nil, nil, listErr
	}
	if len(ownerMembers) == 0 {
		return nil, nil, apperrors.WrapNotFound("owner_identity_group", fmt.Sprintf("%d", ownerGroupID))
	}
	// Actor must cover every parent-owner clinic (anchor + all active members)
	// and all pet clinics (pet refs already checked via assertPetRefsInActorScope).
	// No any-member fallback: missing one parent-owner clinic is Forbidden, zero writes.
	if err := assertActorCoversOwnerGroupClinics(actor, ownerGroup, ownerMembers); err != nil {
		return nil, nil, err
	}

	pets, lockPetsErr := s.repo.LockPets(txCtx, refs)
	if lockPetsErr != nil {
		return nil, nil, lockPetsErr
	}
	if len(pets) != len(refs) {
		return nil, nil, apperrors.WrapForbidden("mixed or hidden pet ids rejected")
	}

	for _, pet := range pets {
		ok, ownErr := s.repo.IsOwnerActiveInGroup(txCtx, ownerGroupID, pet.ClinicID, pet.OwnerID)
		if ownErr != nil {
			return nil, nil, ownErr
		}
		if !ok {
			return nil, nil, apperrors.WrapConflict("pet owner is not in the specified owner identity group")
		}
	}

	existing, err := s.existingMatchingPetGroup(txCtx, actor, ownerGroupID, refs)
	if err != nil {
		return nil, nil, err
	}
	if existing.group != nil {
		return existing.group, existing.members, nil
	}

	for _, ref := range refs {
		m, findErr := s.repo.FindActivePetMembership(txCtx, ref.ClinicID, ref.PetID)
		if findErr != nil {
			return nil, nil, findErr
		}
		if m != nil {
			return nil, nil, apperrors.WrapConflict("pet already linked in another identity group")
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
		return nil, nil, createErr
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
		return nil, nil, createErr
	}
	if auditErr := s.writeAudit(txCtx, actor, model.AuditActionPetIdentityLinkCreate, group.ID, nil, map[string]any{
		"group_id":       group.ID,
		"owner_group_id": ownerGroupID,
		"members":        petRefsAudit(refs),
	}); auditErr != nil {
		return nil, nil, auditErr
	}
	return group, memberRows, nil
}

type matchingPetGroup struct {
	group   *model.PetIdentityGroup
	members []model.PetIdentityGroupMember
}

func (s *service) existingMatchingPetGroup(
	txCtx context.Context,
	actor ActorContext,
	ownerGroupID uint64,
	refs []PetMemberRef,
) (matchingPetGroup, error) {
	existingGroupIDs := map[uint64]struct{}{}
	for _, ref := range refs {
		m, findErr := s.repo.FindActivePetMembership(txCtx, ref.ClinicID, ref.PetID)
		if findErr != nil {
			return matchingPetGroup{}, findErr
		}
		if m == nil {
			return matchingPetGroup{}, nil
		}
		existingGroupIDs[m.GroupID] = struct{}{}
	}
	if len(existingGroupIDs) != 1 {
		return matchingPetGroup{}, nil
	}
	var groupID uint64
	for id := range existingGroupIDs {
		groupID = id
	}
	group, lockGErr := s.repo.LockPetGroupByID(txCtx, groupID)
	if lockGErr != nil {
		return matchingPetGroup{}, lockGErr
	}
	if group.OwnerGroupID != ownerGroupID {
		return matchingPetGroup{}, apperrors.WrapConflict("pets already linked under a different owner identity group")
	}
	active, listErr := s.repo.ListActivePetMembers(txCtx, groupID)
	if listErr != nil {
		return matchingPetGroup{}, listErr
	}
	if !samePetMemberSet(active, refs) {
		return matchingPetGroup{}, nil
	}
	return matchingPetGroup{
		group:   group,
		members: filterPetMembersByClinics(active, actor.VerifiedClinics),
	}, nil
}
