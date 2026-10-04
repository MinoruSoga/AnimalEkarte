package identitylink

import (
	"context"
	"slices"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	errLinkWriteReaderMissing = "identity link permission reader is required"
	errLinkWriteNotMember     = "actor is not a current member of an affected clinic"
	errLinkWriteEditRequired  = "identity-links edit is required on every affected clinic"
)

// LinkWriteDecision is the current membership and identity-links edit result.
// SystemAdmin keeps the existing all-active-clinic access and is not narrowed
// by NonMemberClinicIDs or EditDeniedClinicIDs.
type LinkWriteDecision struct {
	SystemAdmin         bool
	NonMemberClinicIDs  []uint64
	EditDeniedClinicIDs []uint64
	AffectedClinicIDs   []uint64
}

// LinkWriteAuthorizer re-checks current membership and per-clinic identity-links
// edit inside the caller's transaction. Implementations must not lock permission
// rows. a permission change which commits after the in-transaction read can still race.
type LinkWriteAuthorizer interface {
	AuthorizeLinkWrite(ctx context.Context, staffID uint64, clinicIDs []uint64) (LinkWriteDecision, error)
}

type dbLinkWriteAuthorizer struct {
	db *gorm.DB
}

// NewDBLinkWriteAuthorizer reads membership and identity-links edit without
// locking permission rows. a permission change which commits after the in-transaction read can still race.
func NewDBLinkWriteAuthorizer(db *gorm.DB) LinkWriteAuthorizer {
	return dbLinkWriteAuthorizer{db: db}
}

func (a dbLinkWriteAuthorizer) AuthorizeLinkWrite(
	ctx context.Context,
	staffID uint64,
	clinicIDs []uint64,
) (LinkWriteDecision, error) {
	requested := compactClinicIDs(clinicIDs)
	if a.db == nil || staffID == 0 || len(requested) == 0 {
		return LinkWriteDecision{}, apperrors.WrapForbidden(errLinkWriteReaderMissing)
	}
	db := persistence.DBOrTx(ctx, a.db)
	admin, found, err := loadCurrentSystemAdmin(db, staffID)
	if err != nil {
		return LinkWriteDecision{}, err
	}
	if !found {
		return classifyLinkWrite(false, requested, nil, nil), nil
	}
	if admin {
		return classifyLinkWrite(true, requested, nil, nil), nil
	}
	members, err := loadCurrentMemberClinics(db, staffID, requested)
	if err != nil {
		return LinkWriteDecision{}, err
	}
	editClinics, err := loadIdentityLinkEditClinics(db, staffID, members)
	if err != nil {
		return LinkWriteDecision{}, err
	}
	return classifyLinkWrite(false, requested, members, editClinics), nil
}

func loadCurrentSystemAdmin(db *gorm.DB, staffID uint64) (bool, bool, error) {
	var rows []struct {
		IsSystemAdmin bool `gorm:"column:is_system_admin"`
	}
	err := db.Raw(`
		SELECT a.is_system_admin AS is_system_admin
		FROM staffs s
		JOIN accounts a
			ON a.id = s.account_id
			AND a.deleted_at IS NULL
			AND a.is_active = true
		WHERE s.id = ?
			AND s.deleted_at IS NULL
			AND s.is_active = true
	`, staffID).Scan(&rows).Error
	if err != nil {
		return false, false, apperrors.Wrap(err, "failed to read current system admin")
	}
	if len(rows) == 0 {
		return false, false, nil
	}
	return rows[0].IsSystemAdmin, true, nil
}

func loadCurrentMemberClinics(db *gorm.DB, staffID uint64, clinicIDs []uint64) ([]uint64, error) {
	var rows []struct {
		ClinicID uint64 `gorm:"column:clinic_id"`
	}
	err := db.Raw(`
		SELECT clinic_id
		FROM staff_clinic_assignments
		WHERE staff_id = ?
			AND clinic_id IN ?
			AND deleted_at IS NULL
	`, staffID, clinicIDs).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read current clinic membership")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ClinicID)
	}
	return ids, nil
}

func loadIdentityLinkEditClinics(db *gorm.DB, staffID uint64, clinicIDs []uint64) ([]uint64, error) {
	if len(clinicIDs) == 0 {
		return nil, nil
	}
	var rows []struct {
		ClinicID uint64 `gorm:"column:clinic_id"`
	}
	// Same effective-permission join as auth.FindAllEffectivePermissionsByStaffID.
	// Plain SELECT only: do not lock permission rows.
	err := db.Raw(`
		SELECT pg.clinic_id AS clinic_id
		FROM staff_permission_groups spg
		JOIN permission_groups pg
			ON pg.id = spg.group_id
			AND pg.deleted_at IS NULL
			AND pg.is_active = true
			AND pg.clinic_id IN ?
		JOIN permission_group_rules pgr
			ON pgr.group_id = pg.id
			AND pgr.deleted_at IS NULL
			AND pgr.resource = ?
		WHERE spg.staff_id = ?
		GROUP BY pg.clinic_id
		HAVING bool_or(pgr.can_edit)
	`, clinicIDs, string(model.ResourceIdentityLinks), staffID).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read identity-links edit")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ClinicID)
	}
	return ids, nil
}

func classifyLinkWrite(systemAdmin bool, requested, members, editClinics []uint64) LinkWriteDecision {
	if systemAdmin {
		return LinkWriteDecision{SystemAdmin: true}
	}
	memberSet := clinicSet(members)
	editSet := clinicSet(editClinics)
	decision := LinkWriteDecision{}
	for _, clinicID := range requested {
		if _, ok := memberSet[clinicID]; !ok {
			decision.NonMemberClinicIDs = append(decision.NonMemberClinicIDs, clinicID)
			continue
		}
		if _, ok := editSet[clinicID]; !ok {
			decision.EditDeniedClinicIDs = append(decision.EditDeniedClinicIDs, clinicID)
		}
	}
	return decision
}

func (s *service) authorizeAffectedClinics(
	ctx context.Context,
	actor ActorContext,
	clinicIDs []uint64,
) (LinkWriteDecision, error) {
	affected := compactClinicIDs(clinicIDs)
	if s.authorizer == nil || len(affected) == 0 {
		return LinkWriteDecision{}, apperrors.WrapForbidden(errLinkWriteReaderMissing)
	}
	decision, err := s.authorizer.AuthorizeLinkWrite(ctx, actor.StaffID, affected)
	if err != nil {
		return LinkWriteDecision{}, err
	}
	decision.AffectedClinicIDs = affected
	if decision.SystemAdmin {
		return decision, nil
	}
	if len(decision.NonMemberClinicIDs) > 0 {
		return LinkWriteDecision{}, apperrors.WrapForbidden(errLinkWriteNotMember)
	}
	if len(decision.EditDeniedClinicIDs) > 0 {
		return LinkWriteDecision{}, apperrors.WrapForbidden(errLinkWriteEditRequired)
	}
	return decision, nil
}

func ownerAffectedClinicIDs(
	anchor uint64,
	existing []model.OwnerIdentityGroupMember,
	refs []OwnerMemberRef,
) []uint64 {
	ids := []uint64{anchor}
	for _, member := range existing {
		ids = append(ids, member.ClinicID)
	}
	for _, ref := range refs {
		ids = append(ids, ref.ClinicID)
	}
	return ids
}

func petAffectedClinicIDs(
	petAnchor uint64,
	petMembers []model.PetIdentityGroupMember,
	refs []PetMemberRef,
	ownerGroup *model.OwnerIdentityGroup,
	ownerMembers []model.OwnerIdentityGroupMember,
) []uint64 {
	ids := []uint64{petAnchor}
	if ownerGroup != nil {
		ids = append(ids, ownerGroup.CreatedClinicID)
	}
	for _, member := range petMembers {
		ids = append(ids, member.ClinicID)
	}
	for _, ref := range refs {
		ids = append(ids, ref.ClinicID)
	}
	for _, member := range ownerMembers {
		ids = append(ids, member.ClinicID)
	}
	return ids
}

func compactClinicIDs(ids []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(ids))
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

func clinicSet(ids []uint64) map[uint64]struct{} {
	set := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id != 0 {
			set[id] = struct{}{}
		}
	}
	return set
}
