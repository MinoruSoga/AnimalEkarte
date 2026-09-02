package identitylink

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// Repository owns identity-link persistence. All mutating methods require an ambient tx.
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

	// IsOwnerActiveInGroup reports whether (clinicID, ownerID) is an active member of groupID.
	IsOwnerActiveInGroup(ctx context.Context, groupID, clinicID, ownerID uint64) (bool, error)

	// ResolveLinkedPetPairs returns correlated (clinic_id, pet_id) pairs for a seed pet,
	// including the seed itself, scoped to actorClinicIDs. Bare IN expansion is forbidden.
	ResolveLinkedPetPairs(ctx context.Context, seedClinicID, seedPetID uint64, actorClinicIDs []uint64) ([]ClinicPetPair, error)

	// ListLinkedTreatmentHistory loads treatments for correlated pet pairs only.
	ListLinkedTreatmentHistory(ctx context.Context, pairs []ClinicPetPair, page, limit int) ([]LinkedTreatmentHistoryItem, int64, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository constructs the identity-link repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) conn(ctx context.Context) *gorm.DB {
	return persistence.DBOrTx(ctx, r.db)
}

func (r *repository) requireAmbientTx(ctx context.Context) (*gorm.DB, error) {
	if tx := persistence.TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx), nil
	}
	return nil, apperrors.WrapInternalServerError("identitylink write requires ambient transaction")
}

func (r *repository) SearchOwners(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Owner, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("clinicIDs required")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := textsearch.NormalizeQuerySpaces(query)
	if q == "" {
		return []model.Owner{}, nil
	}

	rawPattern := "%" + textsearch.EscapeLike(q) + "%"
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(q)) + "%"

	var owners []model.Owner
	err := r.conn(ctx).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).
		Where(
			`(translate(name, ?, ?) ILIKE ? ESCAPE '\'
			  OR name_kana ILIKE ? ESCAPE '\' OR phone ILIKE ? ESCAPE '\'
			  OR translate(name, ?, ?) ILIKE ? ESCAPE '\'
			  OR translate(name_kana, ?, ?) ILIKE ? ESCAPE '\')`,
			textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
			rawPattern, rawPattern,
			textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
			textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		).
		Order("clinic_id ASC, id ASC").
		Limit(limit).
		Find(&owners).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", q)
	}
	return owners, nil
}

func (r *repository) SearchPets(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Pet, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("clinicIDs required")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := textsearch.NormalizeQuerySpaces(query)
	if q == "" {
		return []model.Pet{}, nil
	}

	rawPattern := "%" + textsearch.EscapeLike(q) + "%"
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(q)) + "%"

	var pets []model.Pet
	err := r.conn(ctx).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).
		Where(
			`(translate(name, ?, ?) ILIKE ? ESCAPE '\'
			  OR name_kana ILIKE ? ESCAPE '\' OR pet_number ILIKE ? ESCAPE '\'
			  OR translate(name, ?, ?) ILIKE ? ESCAPE '\'
			  OR translate(name_kana, ?, ?) ILIKE ? ESCAPE '\')`,
			textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
			rawPattern, rawPattern,
			textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
			textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		).
		Order("clinic_id ASC, id ASC").
		Limit(limit).
		Find(&pets).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", q)
	}
	return pets, nil
}

func sortOwnerRefs(refs []OwnerMemberRef) []OwnerMemberRef {
	out := append([]OwnerMemberRef(nil), refs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ClinicID != out[j].ClinicID {
			return out[i].ClinicID < out[j].ClinicID
		}
		return out[i].OwnerID < out[j].OwnerID
	})
	return out
}

func sortPetRefs(refs []PetMemberRef) []PetMemberRef {
	out := append([]PetMemberRef(nil), refs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ClinicID != out[j].ClinicID {
			return out[i].ClinicID < out[j].ClinicID
		}
		return out[i].PetID < out[j].PetID
	})
	return out
}

// LockOwners locks and returns every requested owner in deterministic order.
// Any missing pair fails closed (no partial result).
func (r *repository) LockOwners(ctx context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return nil, apperrors.WrapInvalidInput("owner members required")
	}
	ordered := sortOwnerRefs(refs)
	owners := make([]model.Owner, 0, len(ordered))
	for _, ref := range ordered {
		var owner model.Owner
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("clinic_id = ? AND id = ?", ref.ClinicID, ref.OwnerID).
			First(&owner).Error
		if err != nil {
			return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d/%d", ref.ClinicID, ref.OwnerID))
		}
		owners = append(owners, owner)
	}
	return owners, nil
}

func (r *repository) LockPets(ctx context.Context, refs []PetMemberRef) ([]model.Pet, error) {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return nil, apperrors.WrapInvalidInput("pet members required")
	}
	ordered := sortPetRefs(refs)
	pets := make([]model.Pet, 0, len(ordered))
	for _, ref := range ordered {
		var pet model.Pet
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("clinic_id = ? AND id = ?", ref.ClinicID, ref.PetID).
			First(&pet).Error
		if err != nil {
			return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d/%d", ref.ClinicID, ref.PetID))
		}
		pets = append(pets, pet)
	}
	return pets, nil
}

func (r *repository) FindActiveOwnerMembership(ctx context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error) {
	var m model.OwnerIdentityGroupMember
	err := r.conn(ctx).
		Where("clinic_id = ? AND owner_id = ? AND deleted_at IS NULL", clinicID, ownerID).
		First(&m).Error
	if err != nil {
		if apperrors.IsNotFound(apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d/%d", clinicID, ownerID))) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d/%d", clinicID, ownerID))
	}
	return &m, nil
}

func (r *repository) FindActivePetMembership(ctx context.Context, clinicID, petID uint64) (*model.PetIdentityGroupMember, error) {
	var m model.PetIdentityGroupMember
	err := r.conn(ctx).
		Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", clinicID, petID).
		First(&m).Error
	if err != nil {
		if apperrors.IsNotFound(apperrors.FromGORM(err, "pet_identity_group_member", fmt.Sprintf("%d/%d", clinicID, petID))) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "pet_identity_group_member", fmt.Sprintf("%d/%d", clinicID, petID))
	}
	return &m, nil
}

func (r *repository) FindOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
	var g model.OwnerIdentityGroup
	err := r.conn(ctx).
		Where("id = ? AND deleted_at IS NULL", groupID).
		First(&g).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner_identity_group", fmt.Sprintf("%d", groupID))
	}
	return &g, nil
}

func (r *repository) LockOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return nil, err
	}
	var g model.OwnerIdentityGroup
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", groupID).
		First(&g).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner_identity_group", fmt.Sprintf("%d", groupID))
	}
	return &g, nil
}

func (r *repository) LockPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error) {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return nil, err
	}
	var g model.PetIdentityGroup
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", groupID).
		First(&g).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_identity_group", fmt.Sprintf("%d", groupID))
	}
	return &g, nil
}

func (r *repository) FindPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error) {
	var g model.PetIdentityGroup
	err := r.conn(ctx).
		Where("id = ? AND deleted_at IS NULL", groupID).
		First(&g).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_identity_group", fmt.Sprintf("%d", groupID))
	}
	return &g, nil
}

func (r *repository) ListActiveOwnerMembers(ctx context.Context, groupID uint64) ([]model.OwnerIdentityGroupMember, error) {
	var members []model.OwnerIdentityGroupMember
	err := r.conn(ctx).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("clinic_id ASC, owner_id ASC").
		Find(&members).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return members, nil
}

func (r *repository) ListActivePetMembers(ctx context.Context, groupID uint64) ([]model.PetIdentityGroupMember, error) {
	var members []model.PetIdentityGroupMember
	err := r.conn(ctx).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("clinic_id ASC, pet_id ASC").
		Find(&members).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return members, nil
}

func (r *repository) ListActiveOwnerMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.OwnerIdentityGroupMember, error) {
	if len(clinicIDs) == 0 {
		return []model.OwnerIdentityGroupMember{}, nil
	}
	var members []model.OwnerIdentityGroupMember
	err := r.conn(ctx).
		Where("group_id = ? AND deleted_at IS NULL AND clinic_id IN ?", groupID, clinicIDs).
		Order("clinic_id ASC, owner_id ASC").
		Find(&members).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return members, nil
}

func (r *repository) ListActivePetMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.PetIdentityGroupMember, error) {
	if len(clinicIDs) == 0 {
		return []model.PetIdentityGroupMember{}, nil
	}
	var members []model.PetIdentityGroupMember
	err := r.conn(ctx).
		Where("group_id = ? AND deleted_at IS NULL AND clinic_id IN ?", groupID, clinicIDs).
		Order("clinic_id ASC, pet_id ASC").
		Find(&members).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return members, nil
}

func (r *repository) CreateOwnerGroup(ctx context.Context, group *model.OwnerIdentityGroup) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	if err := tx.Create(group).Error; err != nil {
		return apperrors.FromGORM(err, "owner_identity_group", "")
	}
	return nil
}

func (r *repository) CreateOwnerMembers(ctx context.Context, members []model.OwnerIdentityGroupMember) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	if err := tx.Create(&members).Error; err != nil {
		return apperrors.FromGORM(err, "owner_identity_group_member", "")
	}
	return nil
}

func (r *repository) SoftDeleteOwnerMember(ctx context.Context, memberID uint64) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	res := tx.Where("id = ? AND deleted_at IS NULL", memberID).Delete(&model.OwnerIdentityGroupMember{})
	if res.Error != nil {
		return apperrors.FromGORM(res.Error, "owner_identity_group_member", fmt.Sprintf("%d", memberID))
	}
	if res.RowsAffected == 0 {
		return apperrors.WrapNotFound("owner_identity_group_member", fmt.Sprintf("%d", memberID))
	}
	return nil
}

func (r *repository) SoftDeleteOwnerGroup(ctx context.Context, groupID uint64) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	res := tx.Where("id = ? AND deleted_at IS NULL", groupID).Delete(&model.OwnerIdentityGroup{})
	if res.Error != nil {
		return apperrors.FromGORM(res.Error, "owner_identity_group", fmt.Sprintf("%d", groupID))
	}
	if res.RowsAffected == 0 {
		return apperrors.WrapNotFound("owner_identity_group", fmt.Sprintf("%d", groupID))
	}
	return nil
}

func (r *repository) CountActiveOwnerMembers(ctx context.Context, groupID uint64) (int64, error) {
	var n int64
	err := r.conn(ctx).Model(&model.OwnerIdentityGroupMember{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&n).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return n, nil
}

func (r *repository) CreatePetGroup(ctx context.Context, group *model.PetIdentityGroup) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	if err := tx.Create(group).Error; err != nil {
		return apperrors.FromGORM(err, "pet_identity_group", "")
	}
	return nil
}

func (r *repository) CreatePetMembers(ctx context.Context, members []model.PetIdentityGroupMember) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	if err := tx.Create(&members).Error; err != nil {
		return apperrors.FromGORM(err, "pet_identity_group_member", "")
	}
	return nil
}

func (r *repository) SoftDeletePetMember(ctx context.Context, memberID uint64) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	res := tx.Where("id = ? AND deleted_at IS NULL", memberID).Delete(&model.PetIdentityGroupMember{})
	if res.Error != nil {
		return apperrors.FromGORM(res.Error, "pet_identity_group_member", fmt.Sprintf("%d", memberID))
	}
	if res.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet_identity_group_member", fmt.Sprintf("%d", memberID))
	}
	return nil
}

func (r *repository) SoftDeletePetGroup(ctx context.Context, groupID uint64) error {
	tx, err := r.requireAmbientTx(ctx)
	if err != nil {
		return err
	}
	res := tx.Where("id = ? AND deleted_at IS NULL", groupID).Delete(&model.PetIdentityGroup{})
	if res.Error != nil {
		return apperrors.FromGORM(res.Error, "pet_identity_group", fmt.Sprintf("%d", groupID))
	}
	if res.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet_identity_group", fmt.Sprintf("%d", groupID))
	}
	return nil
}

func (r *repository) CountActivePetMembers(ctx context.Context, groupID uint64) (int64, error) {
	var n int64
	err := r.conn(ctx).Model(&model.PetIdentityGroupMember{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&n).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return n, nil
}

func (r *repository) IsOwnerActiveInGroup(ctx context.Context, groupID, clinicID, ownerID uint64) (bool, error) {
	var n int64
	err := r.conn(ctx).Model(&model.OwnerIdentityGroupMember{}).
		Where("group_id = ? AND clinic_id = ? AND owner_id = ? AND deleted_at IS NULL", groupID, clinicID, ownerID).
		Count(&n).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "owner_identity_group_member", fmt.Sprintf("%d", groupID))
	}
	return n > 0, nil
}

func (r *repository) ResolveLinkedPetPairs(
	ctx context.Context,
	seedClinicID, seedPetID uint64,
	actorClinicIDs []uint64,
) ([]ClinicPetPair, error) {
	if seedClinicID == 0 || seedPetID == 0 {
		return nil, apperrors.WrapInvalidInput("seed pet required")
	}
	if len(actorClinicIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("clinicIDs required")
	}
	if !containsUint64(actorClinicIDs, seedClinicID) {
		return nil, apperrors.WrapForbidden("seed pet clinic outside actor scope")
	}

	// Seed pair is always included when the pet exists in actor scope.
	var seed model.Pet
	err := r.conn(ctx).
		Where("clinic_id = ? AND id = ?", seedClinicID, seedPetID).
		First(&seed).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d/%d", seedClinicID, seedPetID))
	}

	pairs := []ClinicPetPair{{ClinicID: seed.ClinicID, PetID: seed.ID}}

	membership, err := r.FindActivePetMembership(ctx, seedClinicID, seedPetID)
	if err != nil {
		return nil, err
	}
	if membership == nil {
		return pairs, nil
	}

	members, err := r.ListActivePetMembersByClinicIDs(ctx, membership.GroupID, actorClinicIDs)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{
		pairKey(seed.ClinicID, seed.ID): {},
	}
	for _, m := range members {
		k := pairKey(m.ClinicID, m.PetID)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		pairs = append(pairs, ClinicPetPair{ClinicID: m.ClinicID, PetID: m.PetID})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].ClinicID != pairs[j].ClinicID {
			return pairs[i].ClinicID < pairs[j].ClinicID
		}
		return pairs[i].PetID < pairs[j].PetID
	})
	return pairs, nil
}

func (r *repository) ListLinkedTreatmentHistory(
	ctx context.Context,
	pairs []ClinicPetPair,
	page, limit int,
) ([]LinkedTreatmentHistoryItem, int64, error) {
	if len(pairs) == 0 {
		return []LinkedTreatmentHistoryItem{}, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Correlated pair predicate — never bare IN(clinic_ids) AND IN(pet_ids).
	orParts := make([]string, 0, len(pairs))
	args := make([]any, 0, len(pairs)*2)
	for _, p := range pairs {
		orParts = append(orParts, "(mr.clinic_id = ? AND mr.pet_id = ?)")
		args = append(args, p.ClinicID, p.PetID)
	}
	pairSQL := strings.Join(orParts, " OR ")

	countQ := r.conn(ctx).
		Table("treatments t").
		Joins("JOIN medical_records mr ON mr.id = t.medical_record_id AND mr.deleted_at IS NULL").
		Where("t.deleted_at IS NULL").
		Where(pairSQL, args...)

	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "treatment", "linked_history")
	}

	type row struct {
		ClinicID        uint64
		PetID           uint64
		MedicalRecordID uint64
		RecordNo        string
		RecordDate      string
		TreatmentID     uint64
		ItemType        string
		Content         string
		UnitPrice       int64
		Quantity        float64
	}
	var rows []row
	err := r.conn(ctx).
		Table("treatments t").
		Select(`mr.clinic_id AS clinic_id,
			mr.pet_id AS pet_id,
			mr.id AS medical_record_id,
			mr.record_no AS record_no,
			mr.date::text AS record_date,
			t.id AS treatment_id,
			t.item_type::text AS item_type,
			t.content AS content,
			t.unit_price AS unit_price,
			t.quantity AS quantity`).
		Joins("JOIN medical_records mr ON mr.id = t.medical_record_id AND mr.deleted_at IS NULL").
		Where("t.deleted_at IS NULL").
		Where(pairSQL, args...).
		Order("mr.date DESC, t.id DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, apperrors.FromGORM(err, "treatment", "linked_history")
	}

	out := make([]LinkedTreatmentHistoryItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, LinkedTreatmentHistoryItem(row))
	}
	return out, total, nil
}

func pairKey(clinicID, entityID uint64) string {
	return fmt.Sprintf("%d:%d", clinicID, entityID)
}

func containsUint64(ids []uint64, target uint64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
