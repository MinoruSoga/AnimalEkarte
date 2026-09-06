package identitylink

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the identity-link use-case view of persistence.
// The GORM store is concrete; mutating methods still require an ambient tx.
type Repository interface {
	SearchOwners(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Owner, error)
	SearchPets(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Pet, error)

	LockOwners(ctx context.Context, refs []OwnerMemberRef) ([]model.Owner, error)
	LockPets(ctx context.Context, refs []PetMemberRef) ([]model.Pet, error)

	FindActiveOwnerMembership(ctx context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error)
	FindActivePetMembership(ctx context.Context, clinicID, petID uint64) (*model.PetIdentityGroupMember, error)

	LockOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error)
	LockPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error)
	FindOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error)
	FindPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error)

	ListActiveOwnerMembers(ctx context.Context, groupID uint64) ([]model.OwnerIdentityGroupMember, error)
	ListActivePetMembers(ctx context.Context, groupID uint64) ([]model.PetIdentityGroupMember, error)
	ListActiveOwnerMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.OwnerIdentityGroupMember, error)
	ListActivePetMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.PetIdentityGroupMember, error)

	CreateOwnerGroup(ctx context.Context, group *model.OwnerIdentityGroup) error
	CreateOwnerMembers(ctx context.Context, members []model.OwnerIdentityGroupMember) error
	SoftDeleteOwnerMember(ctx context.Context, memberID uint64) error
	SoftDeleteOwnerGroup(ctx context.Context, groupID uint64) error
	CountActiveOwnerMembers(ctx context.Context, groupID uint64) (int64, error)

	CreatePetGroup(ctx context.Context, group *model.PetIdentityGroup) error
	CreatePetMembers(ctx context.Context, members []model.PetIdentityGroupMember) error
	SoftDeletePetMember(ctx context.Context, memberID uint64) error
	SoftDeletePetGroup(ctx context.Context, groupID uint64) error
	CountActivePetMembers(ctx context.Context, groupID uint64) (int64, error)
	CountActivePetGroupsByOwnerGroupID(ctx context.Context, ownerGroupID uint64) (int64, error)

	IsOwnerActiveInGroup(ctx context.Context, groupID, clinicID, ownerID uint64) (bool, error)
	ResolveLinkedPetPairs(ctx context.Context, seedClinicID, seedPetID uint64, actorClinicIDs []uint64) ([]ClinicPetPair, error)
	ListLinkedTreatmentHistory(ctx context.Context, pairs []ClinicPetPair, page, limit int) ([]LinkedTreatmentHistoryItem, int64, error)
}

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
