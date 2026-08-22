package identitylink

import (
	"context"
	"fmt"
	"slices"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Service is the identity-link use-case boundary.
type Service interface {
	SearchOwners(ctx context.Context, actor ActorContext, query string, limit int) ([]model.Owner, error)
	SearchPets(ctx context.Context, actor ActorContext, query string, limit int) ([]model.Pet, error)

	GetOwnerGroup(ctx context.Context, actor ActorContext, groupID uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	GetPetGroup(ctx context.Context, actor ActorContext, groupID uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)
	FindOwnerGroupByMember(ctx context.Context, actor ActorContext, clinicID, ownerID uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	FindPetGroupByMember(ctx context.Context, actor ActorContext, clinicID, petID uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)

	CreateOwnerGroup(ctx context.Context, actor ActorContext, members []OwnerMemberRef) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	AddOwnerMembers(ctx context.Context, actor ActorContext, groupID uint64, members []OwnerMemberRef) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	UnlinkOwnerMember(ctx context.Context, actor ActorContext, groupID uint64, member OwnerMemberRef) error

	CreatePetGroup(ctx context.Context, actor ActorContext, ownerGroupID uint64, members []PetMemberRef) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)
	AddPetMembers(ctx context.Context, actor ActorContext, groupID uint64, members []PetMemberRef) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)
	UnlinkPetMember(ctx context.Context, actor ActorContext, groupID uint64, member PetMemberRef) error

	ListLinkedTreatmentHistory(
		ctx context.Context,
		actor ActorContext,
		seedClinicID, seedPetID uint64,
		includeLinked bool,
		page, limit int,
	) ([]LinkedTreatmentHistoryItem, int64, error)
}

type service struct {
	repo       Repository
	transactor persistence.Transactor
	audit      audit.TxLogger
}

// NewService constructs the identity-link service. audit must be non-nil;
// nil audit is fail-closed at write time.
func NewService(repo Repository, transactor persistence.Transactor, auditLogger audit.TxLogger) Service {
	return &service{repo: repo, transactor: transactor, audit: auditLogger}
}

func (s *service) SearchOwners(ctx context.Context, actor ActorContext, query string, limit int) ([]model.Owner, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, err
	}
	owners, err := s.repo.SearchOwners(ctx, actor.VerifiedClinics, query, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to search owners for identity link")
	}
	return owners, nil
}

func (s *service) SearchPets(ctx context.Context, actor ActorContext, query string, limit int) ([]model.Pet, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, err
	}
	pets, err := s.repo.SearchPets(ctx, actor.VerifiedClinics, query, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to search pets for identity link")
	}
	return pets, nil
}

func (s *service) GetOwnerGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	// Read path: only members in actor clinics. Zero visible members → not found
	// (hidden/out-of-scope groups are not probed via unscoped member reads).
	members, err := s.repo.ListActiveOwnerMembersByClinicIDs(ctx, groupID, actor.VerifiedClinics)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to list owner identity members")
	}
	if len(members) == 0 {
		return nil, nil, apperrors.WrapNotFound("owner_identity_group", fmt.Sprintf("%d", groupID))
	}
	group, findErr := s.repo.FindOwnerGroupByID(ctx, groupID)
	if findErr != nil {
		return nil, nil, findErr
	}
	return group, members, nil
}

func (s *service) GetPetGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	members, err := s.repo.ListActivePetMembersByClinicIDs(ctx, groupID, actor.VerifiedClinics)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to list pet identity members")
	}
	if len(members) == 0 {
		return nil, nil, apperrors.WrapNotFound("pet_identity_group", fmt.Sprintf("%d", groupID))
	}
	group, findErr := s.repo.FindPetGroupByID(ctx, groupID)
	if findErr != nil {
		return nil, nil, findErr
	}
	return group, members, nil
}

func (s *service) FindOwnerGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, ownerID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	if !containsUint64(actor.VerifiedClinics, clinicID) {
		return nil, nil, apperrors.WrapForbidden("owner clinic outside actor scope")
	}
	m, err := s.repo.FindActiveOwnerMembership(ctx, clinicID, ownerID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to find owner membership")
	}
	if m == nil {
		return nil, nil, apperrors.WrapNotFound("owner_identity_group_member", fmt.Sprintf("%d/%d", clinicID, ownerID))
	}
	return s.GetOwnerGroup(ctx, actor, m.GroupID)
}

func (s *service) FindPetGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, petID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, nil, err
	}
	if !containsUint64(actor.VerifiedClinics, clinicID) {
		return nil, nil, apperrors.WrapForbidden("pet clinic outside actor scope")
	}
	m, err := s.repo.FindActivePetMembership(ctx, clinicID, petID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to find pet membership")
	}
	if m == nil {
		return nil, nil, apperrors.WrapNotFound("pet_identity_group_member", fmt.Sprintf("%d/%d", clinicID, petID))
	}
	return s.GetPetGroup(ctx, actor, m.GroupID)
}

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
		owners, lockErr := s.repo.LockOwners(txCtx, refs)
		if lockErr != nil {
			// hidden / nonexistent / cross-clinic missing → reject whole request, zero writes
			return lockErr
		}
		if len(owners) != len(refs) {
			return apperrors.WrapForbidden("mixed or hidden owner ids rejected")
		}

		// Idempotent retry: if all members already share one active group and
		// that group has exactly those members, return it without new writes.
		existingGroupIDs := map[uint64]struct{}{}
		for _, ref := range refs {
			m, findErr := s.repo.FindActiveOwnerMembership(txCtx, ref.ClinicID, ref.OwnerID)
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
			group, lockGErr := s.repo.LockOwnerGroupByID(txCtx, groupID)
			if lockGErr != nil {
				return lockGErr
			}
			active, listErr := s.repo.ListActiveOwnerMembers(txCtx, groupID)
			if listErr != nil {
				return listErr
			}
			if sameOwnerMemberSet(active, refs) {
				outGroup = group
				outMembers = filterOwnerMembersByClinics(active, actor.VerifiedClinics)
				return nil
			}
		}

		// Any member already in a different group → conflict (no partial merge in Phase 1).
		for _, ref := range refs {
			m, findErr := s.repo.FindActiveOwnerMembership(txCtx, ref.ClinicID, ref.OwnerID)
			if findErr != nil {
				return findErr
			}
			if m != nil {
				return apperrors.WrapConflict("owner already linked in another identity group")
			}
		}

		createdClinicID := pickCreatedClinicID(actor, refs[0].ClinicID)
		group := &model.OwnerIdentityGroup{
			CreatedClinicID: createdClinicID,
			Version:         1,
		}
		if createErr := s.repo.CreateOwnerGroup(txCtx, group); createErr != nil {
			return createErr
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
			return createErr
		}
		if auditErr := s.writeAudit(txCtx, actor, model.AuditActionOwnerIdentityLinkCreate, group.ID, nil, map[string]any{
			"group_id":          group.ID,
			"created_clinic_id": group.CreatedClinicID,
			"members":           ownerRefsAudit(refs),
		}); auditErr != nil {
			return auditErr
		}
		outGroup = group
		outMembers = memberRows
		return nil
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create owner identity group")
	}
	return outGroup, outMembers, nil
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

func (s *service) ListLinkedTreatmentHistory(
	ctx context.Context,
	actor ActorContext,
	seedClinicID, seedPetID uint64,
	includeLinked bool,
	page, limit int,
) ([]LinkedTreatmentHistoryItem, int64, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, 0, err
	}
	if !containsUint64(actor.VerifiedClinics, seedClinicID) {
		return nil, 0, apperrors.WrapForbidden("seed pet clinic outside actor scope")
	}

	var pairs []ClinicPetPair
	var err error
	if includeLinked {
		pairs, err = s.repo.ResolveLinkedPetPairs(ctx, seedClinicID, seedPetID, actor.VerifiedClinics)
	} else {
		// Default include_linked=false: seed pair only after existence check.
		if _, resolveErr := s.repo.ResolveLinkedPetPairs(ctx, seedClinicID, seedPetID, []uint64{seedClinicID}); resolveErr != nil {
			err = resolveErr
		} else {
			pairs = []ClinicPetPair{{ClinicID: seedClinicID, PetID: seedPetID}}
		}
	}
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to resolve linked pet pairs")
	}
	items, total, listErr := s.repo.ListLinkedTreatmentHistory(ctx, pairs, page, limit)
	if listErr != nil {
		return nil, 0, apperrors.Wrap(listErr, "failed to list linked treatment history")
	}
	return items, total, nil
}

func (s *service) writeAudit(
	ctx context.Context,
	actor ActorContext,
	action string,
	resourceID uint64,
	oldValue, newValue any,
) error {
	if s.audit == nil {
		return apperrors.WrapInternalServerError("identity link audit logger is required")
	}
	clinicID := actor.HomeClinicID
	if clinicID == 0 && len(actor.VerifiedClinics) > 0 {
		clinicID = actor.VerifiedClinics[0]
	}
	rid := resourceID
	staffID := actor.StaffID
	return s.audit.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   &clinicID,
		ActorID:    &staffID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     action,
		Resource:   model.AuditResourceIdentityLink,
		ResourceID: &rid,
		// Non-PHI only: IDs and group metadata. Never names/phones.
		OldValue:  oldValue,
		NewValue:  newValue,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})
}

// --- helpers ---

func requireActorClinics(actor ActorContext) error {
	if len(actor.VerifiedClinics) == 0 {
		return apperrors.WrapUnauthorized("missing verified clinic scope")
	}
	if actor.StaffID == 0 {
		return apperrors.WrapUnauthorized("missing staff identity")
	}
	return nil
}

func assertOwnerRefsInActorScope(actor ActorContext, refs []OwnerMemberRef) error {
	for _, ref := range refs {
		if ref.ClinicID == 0 || ref.OwnerID == 0 {
			return apperrors.WrapInvalidInput("clinic_id and owner_id required")
		}
		if !containsUint64(actor.VerifiedClinics, ref.ClinicID) {
			return apperrors.WrapForbidden("mixed or cross-clinic owner ids rejected")
		}
	}
	return nil
}

func assertPetRefsInActorScope(actor ActorContext, refs []PetMemberRef) error {
	for _, ref := range refs {
		if ref.ClinicID == 0 || ref.PetID == 0 {
			return apperrors.WrapInvalidInput("clinic_id and pet_id required")
		}
		if !containsUint64(actor.VerifiedClinics, ref.ClinicID) {
			return apperrors.WrapForbidden("mixed or cross-clinic pet ids rejected")
		}
	}
	return nil
}

// assertActorCoversOwnerGroupClinics requires the actor to belong to the parent
// owner-group anchor clinic and every active owner-member clinic. Used by pet-group
// mutations that depend on a parent owner group (CreatePetGroup).
func assertActorCoversOwnerGroupClinics(
	actor ActorContext,
	group *model.OwnerIdentityGroup,
	members []model.OwnerIdentityGroupMember,
) error {
	needed := map[uint64]struct{}{group.CreatedClinicID: {}}
	for _, m := range members {
		needed[m.ClinicID] = struct{}{}
	}
	for clinicID := range needed {
		if !containsUint64(actor.VerifiedClinics, clinicID) {
			return apperrors.WrapForbidden("actor must belong to all clinics involved in the parent owner identity group")
		}
	}
	return nil
}

func assertCanManageOwnerGroup(
	actor ActorContext,
	group *model.OwnerIdentityGroup,
	existing []model.OwnerIdentityGroupMember,
	newRefs []OwnerMemberRef,
) error {
	// Actor must currently belong to every clinic involved (existing + new + anchor).
	needed := map[uint64]struct{}{group.CreatedClinicID: {}}
	for _, m := range existing {
		needed[m.ClinicID] = struct{}{}
	}
	for _, r := range newRefs {
		needed[r.ClinicID] = struct{}{}
	}
	for clinicID := range needed {
		if !containsUint64(actor.VerifiedClinics, clinicID) {
			return apperrors.WrapForbidden("actor must belong to all clinics involved in the identity group")
		}
	}
	return nil
}

func assertCanManagePetGroup(
	actor ActorContext,
	group *model.PetIdentityGroup,
	existing []model.PetIdentityGroupMember,
	newRefs []PetMemberRef,
) error {
	needed := map[uint64]struct{}{
		group.CreatedClinicID:           {},
		group.OwnerGroupCreatedClinicID: {},
	}
	for _, m := range existing {
		needed[m.ClinicID] = struct{}{}
	}
	for _, r := range newRefs {
		needed[r.ClinicID] = struct{}{}
	}
	for clinicID := range needed {
		if !containsUint64(actor.VerifiedClinics, clinicID) {
			return apperrors.WrapForbidden("actor must belong to all clinics involved in the identity group")
		}
	}
	return nil
}

func normalizeOwnerRefs(members []OwnerMemberRef) ([]OwnerMemberRef, error) {
	seen := make(map[string]struct{}, len(members))
	out := make([]OwnerMemberRef, 0, len(members))
	for _, m := range members {
		if m.ClinicID == 0 || m.OwnerID == 0 {
			return nil, apperrors.WrapInvalidInput("clinic_id and owner_id required")
		}
		k := pairKey(m.ClinicID, m.OwnerID)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, m)
	}
	return out, nil
}

func normalizePetRefs(members []PetMemberRef) ([]PetMemberRef, error) {
	seen := make(map[string]struct{}, len(members))
	out := make([]PetMemberRef, 0, len(members))
	for _, m := range members {
		if m.ClinicID == 0 || m.PetID == 0 {
			return nil, apperrors.WrapInvalidInput("clinic_id and pet_id required")
		}
		k := pairKey(m.ClinicID, m.PetID)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, m)
	}
	return out, nil
}

func pickCreatedClinicID(actor ActorContext, fallback uint64) uint64 {
	if actor.HomeClinicID != 0 && containsUint64(actor.VerifiedClinics, actor.HomeClinicID) {
		return actor.HomeClinicID
	}
	if fallback != 0 {
		return fallback
	}
	return actor.VerifiedClinics[0]
}

func sameOwnerMemberSet(active []model.OwnerIdentityGroupMember, refs []OwnerMemberRef) bool {
	if len(active) != len(refs) {
		return false
	}
	want := make(map[string]struct{}, len(refs))
	for _, r := range refs {
		want[pairKey(r.ClinicID, r.OwnerID)] = struct{}{}
	}
	for _, m := range active {
		if _, ok := want[pairKey(m.ClinicID, m.OwnerID)]; !ok {
			return false
		}
	}
	return true
}

func samePetMemberSet(active []model.PetIdentityGroupMember, refs []PetMemberRef) bool {
	if len(active) != len(refs) {
		return false
	}
	want := make(map[string]struct{}, len(refs))
	for _, r := range refs {
		want[pairKey(r.ClinicID, r.PetID)] = struct{}{}
	}
	for _, m := range active {
		if _, ok := want[pairKey(m.ClinicID, m.PetID)]; !ok {
			return false
		}
	}
	return true
}

func filterOwnerMembersByClinics(members []model.OwnerIdentityGroupMember, clinicIDs []uint64) []model.OwnerIdentityGroupMember {
	out := make([]model.OwnerIdentityGroupMember, 0, len(members))
	for _, m := range members {
		if slices.Contains(clinicIDs, m.ClinicID) {
			out = append(out, m)
		}
	}
	return out
}

func filterPetMembersByClinics(members []model.PetIdentityGroupMember, clinicIDs []uint64) []model.PetIdentityGroupMember {
	out := make([]model.PetIdentityGroupMember, 0, len(members))
	for _, m := range members {
		if slices.Contains(clinicIDs, m.ClinicID) {
			out = append(out, m)
		}
	}
	return out
}

func ownerRefsAudit(refs []OwnerMemberRef) []map[string]uint64 {
	out := make([]map[string]uint64, 0, len(refs))
	for _, r := range refs {
		out = append(out, map[string]uint64{"clinic_id": r.ClinicID, "owner_id": r.OwnerID})
	}
	return out
}

func petRefsAudit(refs []PetMemberRef) []map[string]uint64 {
	out := make([]map[string]uint64, 0, len(refs))
	for _, r := range refs {
		out = append(out, map[string]uint64{"clinic_id": r.ClinicID, "pet_id": r.PetID})
	}
	return out
}

func refsFromOwnerMembers(members []model.OwnerIdentityGroupMember) []OwnerMemberRef {
	out := make([]OwnerMemberRef, 0, len(members))
	for _, m := range members {
		out = append(out, OwnerMemberRef{ClinicID: m.ClinicID, OwnerID: m.OwnerID})
	}
	return out
}

func refsFromPetMembers(members []model.PetIdentityGroupMember) []PetMemberRef {
	out := make([]PetMemberRef, 0, len(members))
	for _, m := range members {
		out = append(out, PetMemberRef{ClinicID: m.ClinicID, PetID: m.PetID})
	}
	return out
}
