package identitylink

import (
	"slices"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

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
