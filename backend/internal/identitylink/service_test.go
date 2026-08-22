package identitylink

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// --- doubles ---

type mockTxLogger struct {
	mu     sync.Mutex
	calls  []*audit.Entry
	fail   error
	called int
}

func (m *mockTxLogger) LogEntryTx(_ context.Context, input *audit.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called++
	if m.fail != nil {
		return m.fail
	}
	cp := *input
	m.calls = append(m.calls, &cp)
	return nil
}

type noopTransactor struct{}

func (noopTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	// Ambient tx marker for repository.requireAmbientTx in unit tests that use real repo
	// is not needed for pure mock repo tests; just run the callback.
	return fn(ctx)
}

type mockRepo struct {
	searchOwnersFn               func(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Owner, error)
	searchPetsFn                 func(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Pet, error)
	lockOwnersFn                 func(ctx context.Context, refs []OwnerMemberRef) ([]model.Owner, error)
	lockPetsFn                   func(ctx context.Context, refs []PetMemberRef) ([]model.Pet, error)
	findActiveOwnerMembershipFn  func(ctx context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error)
	findActivePetMembershipFn    func(ctx context.Context, clinicID, petID uint64) (*model.PetIdentityGroupMember, error)
	lockOwnerGroupByIDFn         func(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error)
	lockPetGroupByIDFn           func(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error)
	findOwnerGroupByIDFn         func(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error)
	findPetGroupByIDFn           func(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error)
	listActiveOwnerMembersFn     func(ctx context.Context, groupID uint64) ([]model.OwnerIdentityGroupMember, error)
	listActivePetMembersFn       func(ctx context.Context, groupID uint64) ([]model.PetIdentityGroupMember, error)
	listActiveOwnerByClinicsFn   func(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.OwnerIdentityGroupMember, error)
	listActivePetByClinicsFn     func(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.PetIdentityGroupMember, error)
	createOwnerGroupFn           func(ctx context.Context, group *model.OwnerIdentityGroup) error
	createOwnerMembersFn         func(ctx context.Context, members []model.OwnerIdentityGroupMember) error
	softDeleteOwnerMemberFn      func(ctx context.Context, memberID uint64) error
	softDeleteOwnerGroupFn       func(ctx context.Context, groupID uint64) error
	countActiveOwnerMembersFn    func(ctx context.Context, groupID uint64) (int64, error)
	createPetGroupFn             func(ctx context.Context, group *model.PetIdentityGroup) error
	createPetMembersFn           func(ctx context.Context, members []model.PetIdentityGroupMember) error
	softDeletePetMemberFn        func(ctx context.Context, memberID uint64) error
	softDeletePetGroupFn         func(ctx context.Context, groupID uint64) error
	countActivePetMembersFn      func(ctx context.Context, groupID uint64) (int64, error)
	isOwnerActiveInGroupFn       func(ctx context.Context, groupID, clinicID, ownerID uint64) (bool, error)
	resolveLinkedPetPairsFn      func(ctx context.Context, seedClinicID, seedPetID uint64, actorClinicIDs []uint64) ([]ClinicPetPair, error)
	listLinkedTreatmentHistoryFn func(ctx context.Context, pairs []ClinicPetPair, page, limit int) ([]LinkedTreatmentHistoryItem, int64, error)

	createOwnerGroupCalled   int
	createOwnerMembersCalled int
	createPetGroupCalled     int
	createPetMembersCalled   int
	softDeleteOwnerCalled    int
	softDeleteGroupCalled    int
}

func (m *mockRepo) SearchOwners(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Owner, error) {
	if m.searchOwnersFn != nil {
		return m.searchOwnersFn(ctx, clinicIDs, query, limit)
	}
	return nil, nil
}
func (m *mockRepo) SearchPets(ctx context.Context, clinicIDs []uint64, query string, limit int) ([]model.Pet, error) {
	if m.searchPetsFn != nil {
		return m.searchPetsFn(ctx, clinicIDs, query, limit)
	}
	return nil, nil
}
func (m *mockRepo) LockOwners(ctx context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
	if m.lockOwnersFn != nil {
		return m.lockOwnersFn(ctx, refs)
	}
	return nil, nil
}
func (m *mockRepo) LockPets(ctx context.Context, refs []PetMemberRef) ([]model.Pet, error) {
	if m.lockPetsFn != nil {
		return m.lockPetsFn(ctx, refs)
	}
	return nil, nil
}
func (m *mockRepo) FindActiveOwnerMembership(ctx context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error) {
	if m.findActiveOwnerMembershipFn != nil {
		return m.findActiveOwnerMembershipFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}
func (m *mockRepo) FindActivePetMembership(ctx context.Context, clinicID, petID uint64) (*model.PetIdentityGroupMember, error) {
	if m.findActivePetMembershipFn != nil {
		return m.findActivePetMembershipFn(ctx, clinicID, petID)
	}
	return nil, nil
}
func (m *mockRepo) LockOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
	if m.lockOwnerGroupByIDFn != nil {
		return m.lockOwnerGroupByIDFn(ctx, groupID)
	}
	return nil, apperrors.WrapNotFound("owner_identity_group", "0")
}
func (m *mockRepo) LockPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error) {
	if m.lockPetGroupByIDFn != nil {
		return m.lockPetGroupByIDFn(ctx, groupID)
	}
	return nil, apperrors.WrapNotFound("pet_identity_group", "0")
}
func (m *mockRepo) FindOwnerGroupByID(ctx context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
	if m.findOwnerGroupByIDFn != nil {
		return m.findOwnerGroupByIDFn(ctx, groupID)
	}
	return nil, apperrors.WrapNotFound("owner_identity_group", "0")
}
func (m *mockRepo) FindPetGroupByID(ctx context.Context, groupID uint64) (*model.PetIdentityGroup, error) {
	if m.findPetGroupByIDFn != nil {
		return m.findPetGroupByIDFn(ctx, groupID)
	}
	return nil, apperrors.WrapNotFound("pet_identity_group", "0")
}
func (m *mockRepo) ListActiveOwnerMembers(ctx context.Context, groupID uint64) ([]model.OwnerIdentityGroupMember, error) {
	if m.listActiveOwnerMembersFn != nil {
		return m.listActiveOwnerMembersFn(ctx, groupID)
	}
	return nil, nil
}
func (m *mockRepo) ListActivePetMembers(ctx context.Context, groupID uint64) ([]model.PetIdentityGroupMember, error) {
	if m.listActivePetMembersFn != nil {
		return m.listActivePetMembersFn(ctx, groupID)
	}
	return nil, nil
}
func (m *mockRepo) ListActiveOwnerMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.OwnerIdentityGroupMember, error) {
	if m.listActiveOwnerByClinicsFn != nil {
		return m.listActiveOwnerByClinicsFn(ctx, groupID, clinicIDs)
	}
	return nil, nil
}
func (m *mockRepo) ListActivePetMembersByClinicIDs(ctx context.Context, groupID uint64, clinicIDs []uint64) ([]model.PetIdentityGroupMember, error) {
	if m.listActivePetByClinicsFn != nil {
		return m.listActivePetByClinicsFn(ctx, groupID, clinicIDs)
	}
	return nil, nil
}
func (m *mockRepo) CreateOwnerGroup(ctx context.Context, group *model.OwnerIdentityGroup) error {
	m.createOwnerGroupCalled++
	if m.createOwnerGroupFn != nil {
		return m.createOwnerGroupFn(ctx, group)
	}
	group.ID = 99
	return nil
}
func (m *mockRepo) CreateOwnerMembers(ctx context.Context, members []model.OwnerIdentityGroupMember) error {
	m.createOwnerMembersCalled++
	if m.createOwnerMembersFn != nil {
		return m.createOwnerMembersFn(ctx, members)
	}
	return nil
}
func (m *mockRepo) SoftDeleteOwnerMember(ctx context.Context, memberID uint64) error {
	m.softDeleteOwnerCalled++
	if m.softDeleteOwnerMemberFn != nil {
		return m.softDeleteOwnerMemberFn(ctx, memberID)
	}
	return nil
}
func (m *mockRepo) SoftDeleteOwnerGroup(ctx context.Context, groupID uint64) error {
	m.softDeleteGroupCalled++
	if m.softDeleteOwnerGroupFn != nil {
		return m.softDeleteOwnerGroupFn(ctx, groupID)
	}
	return nil
}
func (m *mockRepo) CountActiveOwnerMembers(ctx context.Context, groupID uint64) (int64, error) {
	if m.countActiveOwnerMembersFn != nil {
		return m.countActiveOwnerMembersFn(ctx, groupID)
	}
	return 0, nil
}
func (m *mockRepo) CreatePetGroup(ctx context.Context, group *model.PetIdentityGroup) error {
	m.createPetGroupCalled++
	if m.createPetGroupFn != nil {
		return m.createPetGroupFn(ctx, group)
	}
	group.ID = 55
	return nil
}
func (m *mockRepo) CreatePetMembers(ctx context.Context, members []model.PetIdentityGroupMember) error {
	m.createPetMembersCalled++
	if m.createPetMembersFn != nil {
		return m.createPetMembersFn(ctx, members)
	}
	return nil
}
func (m *mockRepo) SoftDeletePetMember(ctx context.Context, memberID uint64) error {
	if m.softDeletePetMemberFn != nil {
		return m.softDeletePetMemberFn(ctx, memberID)
	}
	return nil
}
func (m *mockRepo) SoftDeletePetGroup(ctx context.Context, groupID uint64) error {
	if m.softDeletePetGroupFn != nil {
		return m.softDeletePetGroupFn(ctx, groupID)
	}
	return nil
}
func (m *mockRepo) CountActivePetMembers(ctx context.Context, groupID uint64) (int64, error) {
	if m.countActivePetMembersFn != nil {
		return m.countActivePetMembersFn(ctx, groupID)
	}
	return 0, nil
}
func (m *mockRepo) IsOwnerActiveInGroup(ctx context.Context, groupID, clinicID, ownerID uint64) (bool, error) {
	if m.isOwnerActiveInGroupFn != nil {
		return m.isOwnerActiveInGroupFn(ctx, groupID, clinicID, ownerID)
	}
	return true, nil
}
func (m *mockRepo) ResolveLinkedPetPairs(ctx context.Context, seedClinicID, seedPetID uint64, actorClinicIDs []uint64) ([]ClinicPetPair, error) {
	if m.resolveLinkedPetPairsFn != nil {
		return m.resolveLinkedPetPairsFn(ctx, seedClinicID, seedPetID, actorClinicIDs)
	}
	return []ClinicPetPair{{ClinicID: seedClinicID, PetID: seedPetID}}, nil
}
func (m *mockRepo) ListLinkedTreatmentHistory(ctx context.Context, pairs []ClinicPetPair, page, limit int) ([]LinkedTreatmentHistoryItem, int64, error) {
	if m.listLinkedTreatmentHistoryFn != nil {
		return m.listLinkedTreatmentHistoryFn(ctx, pairs, page, limit)
	}
	return nil, 0, nil
}

func testActor(clinics ...uint64) ActorContext {
	return ActorContext{
		StaffID:         7,
		HomeClinicID:    clinics[0],
		VerifiedClinics: clinics,
		IPAddress:       "127.0.0.1",
		UserAgent:       "test",
	}
}

// --- tests ---

func TestCreateOwnerGroup_RejectsMixedCrossClinic_NoPartialWrite(t *testing.T) {
	repo := &mockRepo{}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	// Actor only belongs to clinic 1; member includes clinic 2.
	_, _, err := svc.CreateOwnerGroup(context.Background(), testActor(1), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 20},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "mixed/cross-clinic must be forbidden")
	assert.Contains(t, err.Error(), "mixed or cross-clinic")
	assert.Equal(t, 0, repo.createOwnerGroupCalled, "no partial write on mixed/cross-clinic reject")
	assert.Equal(t, 0, repo.createOwnerMembersCalled)
	assert.Equal(t, 0, auditLog.called)
}

func TestCreatePetGroup_RejectsMixedCrossClinic_NoPartialWrite(t *testing.T) {
	repo := &mockRepo{}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	// Actor only belongs to clinic 1; pet members include clinic 2 → reject-all before any write.
	_, _, err := svc.CreatePetGroup(context.Background(), testActor(1), 1, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 2, PetID: 200},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "mixed/cross-clinic pet ids must be forbidden")
	assert.Contains(t, err.Error(), "mixed or cross-clinic")
	assert.Equal(t, 0, repo.createPetGroupCalled, "no partial write on mixed/cross-clinic reject")
	assert.Equal(t, 0, repo.createPetMembersCalled)
	assert.Equal(t, 0, auditLog.called)
}

func TestAddOwnerMembers_RejectsMixedCrossClinic_NoPartialWrite(t *testing.T) {
	repo := &mockRepo{}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.AddOwnerMembers(context.Background(), testActor(1), 3, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 9, OwnerID: 99},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "mixed/cross-clinic must be forbidden")
	assert.Contains(t, err.Error(), "mixed or cross-clinic")
	assert.Equal(t, 0, repo.createOwnerMembersCalled, "no partial write on mixed/cross-clinic reject")
	assert.Equal(t, 0, auditLog.called)
}

func TestAddPetMembers_RejectsMixedCrossClinic_NoPartialWrite(t *testing.T) {
	repo := &mockRepo{}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.AddPetMembers(context.Background(), testActor(1), 7, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 9, PetID: 999},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "mixed/cross-clinic pet ids must be forbidden")
	assert.Contains(t, err.Error(), "mixed or cross-clinic")
	assert.Equal(t, 0, repo.createPetMembersCalled, "no partial write on mixed/cross-clinic reject")
	assert.Equal(t, 0, auditLog.called)
}

func TestCreateOwnerGroup_RejectsHiddenOwner_NoPartialWrite(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn: func(_ context.Context, _ []OwnerMemberRef) ([]model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "1/999")
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.CreateOwnerGroup(context.Background(), testActor(1, 2), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 999},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, 0, repo.createOwnerGroupCalled)
	assert.Equal(t, 0, auditLog.called)
}

func TestCreateOwnerGroup_AuditFailureRollsBack(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn: func(_ context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
			out := make([]model.Owner, 0, len(refs))
			for _, r := range refs {
				out = append(out, model.Owner{ID: r.OwnerID, ClinicID: r.ClinicID, Name: "x"})
			}
			return out, nil
		},
		findActiveOwnerMembershipFn: func(_ context.Context, _, _ uint64) (*model.OwnerIdentityGroupMember, error) {
			return nil, nil
		},
		createOwnerGroupFn: func(_ context.Context, group *model.OwnerIdentityGroup) error {
			group.ID = 42
			return nil
		},
	}
	auditLog := &mockTxLogger{fail: errors.New("audit write failed")}
	// Use a transactor that surfaces callback errors (simulates rollback).
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.CreateOwnerGroup(context.Background(), testActor(1, 2), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 20},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
	assert.Equal(t, 1, repo.createOwnerGroupCalled, "create attempted before audit")
	// In real WithTx the whole tx rolls back; unit test proves audit failure aborts the callback.
}

func TestCreateOwnerGroup_NilAuditFailClosed(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, noopTransactor{}, nil)
	_, _, err := svc.CreateOwnerGroup(context.Background(), testActor(1, 2), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 20},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
	assert.Equal(t, 0, repo.createOwnerGroupCalled)
}

func TestCreateOwnerGroup_SuccessWritesAuditWithoutPHI(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn: func(_ context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
			out := make([]model.Owner, 0, len(refs))
			for _, r := range refs {
				out = append(out, model.Owner{ID: r.OwnerID, ClinicID: r.ClinicID, Name: "Secret Name", Phone: "090"})
			}
			return out, nil
		},
		findActiveOwnerMembershipFn: func(_ context.Context, _, _ uint64) (*model.OwnerIdentityGroupMember, error) {
			return nil, nil
		},
		createOwnerGroupFn: func(_ context.Context, group *model.OwnerIdentityGroup) error {
			group.ID = 7
			return nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	group, members, err := svc.CreateOwnerGroup(context.Background(), testActor(1, 2), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 20},
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, uint64(7), group.ID)
	assert.Len(t, members, 2)
	require.Equal(t, 1, auditLog.called)
	require.Len(t, auditLog.calls, 1)
	// Non-PHI: action/resource only IDs
	assert.Equal(t, model.AuditActionOwnerIdentityLinkCreate, auditLog.calls[0].Action)
	assert.Equal(t, model.AuditResourceIdentityLink, auditLog.calls[0].Resource)
	// NewValue must not carry names/phones — only IDs
	nv, ok := auditLog.calls[0].NewValue.(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, nv, "name")
	assert.NotContains(t, nv, "phone")
}

func TestUnlinkOwnerMember_LastMemberSoftDeletesGroup(t *testing.T) {
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(_ context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: groupID, CreatedClinicID: 1, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{ID: 5, GroupID: 3, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1},
			}, nil
		},
		countActiveOwnerMembersFn: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	err := svc.UnlinkOwnerMember(context.Background(), testActor(1), 3, OwnerMemberRef{ClinicID: 1, OwnerID: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.softDeleteOwnerCalled)
	assert.Equal(t, 1, repo.softDeleteGroupCalled, "last-member unlink must soft-delete group")
	require.Equal(t, 1, auditLog.called)
}

func TestUnlinkOwnerMember_RejectsCrossClinicMember(t *testing.T) {
	repo := &mockRepo{}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	err := svc.UnlinkOwnerMember(context.Background(), testActor(1), 3, OwnerMemberRef{ClinicID: 9, OwnerID: 10})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, 0, repo.softDeleteOwnerCalled)
	assert.Equal(t, 0, auditLog.called)
}

func TestCreatePetGroup_RequiresLinkedOwners(t *testing.T) {
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(_ context.Context, id uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: id, CreatedClinicID: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{GroupID: 1, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1},
				{GroupID: 1, ClinicID: 2, OwnerID: 20, GroupCreatedClinicID: 1},
			}, nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			return []model.Pet{
				{ID: refs[0].PetID, ClinicID: refs[0].ClinicID, OwnerID: 10},
				{ID: refs[1].PetID, ClinicID: refs[1].ClinicID, OwnerID: 99}, // owner not in group
			}, nil
		},
		isOwnerActiveInGroupFn: func(_ context.Context, _, clinicID, ownerID uint64) (bool, error) {
			return ownerID == 10 && clinicID == 1, nil
		},
		findActivePetMembershipFn: func(_ context.Context, _, _ uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.CreatePetGroup(context.Background(), testActor(1, 2), 1, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 2, PetID: 200},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 0, auditLog.called)
}

// Regression: actor missing parent owner-group CreatedClinicID (anchor) but holding
// one owner-member clinic must NOT create a pet group via the old any-member fallback.
func TestCreatePetGroup_RejectsMissingParentOwnerAnchorClinic_NoPartialWrite(t *testing.T) {
	// Owner group anchor clinic=3; active members at clinics 1 and 2.
	// Actor has clinics 1+2 (and pets there) but NOT anchor 3.
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(_ context.Context, id uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: id, CreatedClinicID: 3, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{GroupID: 1, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 3},
				{GroupID: 1, ClinicID: 2, OwnerID: 20, GroupCreatedClinicID: 3},
			}, nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			out := make([]model.Pet, 0, len(refs))
			for _, r := range refs {
				ownerID := uint64(10)
				if r.ClinicID == 2 {
					ownerID = 20
				}
				out = append(out, model.Pet{ID: r.PetID, ClinicID: r.ClinicID, OwnerID: ownerID})
			}
			return out, nil
		},
		isOwnerActiveInGroupFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
		findActivePetMembershipFn: func(_ context.Context, _, _ uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.CreatePetGroup(context.Background(), testActor(1, 2), 1, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 2, PetID: 200},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "missing parent-owner anchor clinic must be forbidden")
	assert.Equal(t, 0, repo.createPetGroupCalled, "zero CreatePetGroup on auth reject")
	assert.Equal(t, 0, repo.createPetMembersCalled, "zero CreatePetMembers on auth reject")
	assert.Equal(t, 0, auditLog.called, "zero audit on auth reject")
}

// Regression: actor has parent-owner anchor but is missing an active owner-member clinic
// must still be Forbidden (full parent owner clinic set required, not just anchor).
func TestCreatePetGroup_RejectsMissingParentOwnerMemberClinic_NoPartialWrite(t *testing.T) {
	// Owner group anchor clinic=1; members at clinics 1, 2, and 3.
	// Actor has 1+2 (pets at 1+2) but not member clinic 3.
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(_ context.Context, id uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: id, CreatedClinicID: 1, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{GroupID: 1, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1},
				{GroupID: 1, ClinicID: 2, OwnerID: 20, GroupCreatedClinicID: 1},
				{GroupID: 1, ClinicID: 3, OwnerID: 30, GroupCreatedClinicID: 1},
			}, nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			out := make([]model.Pet, 0, len(refs))
			for _, r := range refs {
				ownerID := uint64(10)
				if r.ClinicID == 2 {
					ownerID = 20
				}
				out = append(out, model.Pet{ID: r.PetID, ClinicID: r.ClinicID, OwnerID: ownerID})
			}
			return out, nil
		},
		isOwnerActiveInGroupFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
		findActivePetMembershipFn: func(_ context.Context, _, _ uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	_, _, err := svc.CreatePetGroup(context.Background(), testActor(1, 2), 1, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 2, PetID: 200},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "missing parent-owner member clinic must be forbidden")
	assert.Equal(t, 0, repo.createPetGroupCalled)
	assert.Equal(t, 0, repo.createPetMembersCalled)
	assert.Equal(t, 0, auditLog.called)
}

// Happy path: actor covers parent-owner anchor + all owner-member clinics + pet clinics.
func TestCreatePetGroup_AllowsWhenActorCoversAllParentOwnerAndPetClinics(t *testing.T) {
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(_ context.Context, id uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: id, CreatedClinicID: 1, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{GroupID: 1, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1},
				{GroupID: 1, ClinicID: 2, OwnerID: 20, GroupCreatedClinicID: 1},
			}, nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			out := make([]model.Pet, 0, len(refs))
			for _, r := range refs {
				ownerID := uint64(10)
				if r.ClinicID == 2 {
					ownerID = 20
				}
				out = append(out, model.Pet{ID: r.PetID, ClinicID: r.ClinicID, OwnerID: ownerID})
			}
			return out, nil
		},
		isOwnerActiveInGroupFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
		findActivePetMembershipFn: func(_ context.Context, _, _ uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
		createPetGroupFn: func(_ context.Context, group *model.PetIdentityGroup) error {
			group.ID = 55
			return nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	group, members, err := svc.CreatePetGroup(context.Background(), testActor(1, 2), 1, []PetMemberRef{
		{ClinicID: 1, PetID: 100},
		{ClinicID: 2, PetID: 200},
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, uint64(55), group.ID)
	assert.Len(t, members, 2)
	assert.Equal(t, 1, repo.createPetGroupCalled)
	assert.Equal(t, 1, repo.createPetMembersCalled)
	assert.Equal(t, 1, auditLog.called)
}

func TestGetOwnerGroup_HiddenOutsideScope_NotFound(t *testing.T) {
	repo := &mockRepo{
		listActiveOwnerByClinicsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.OwnerIdentityGroupMember, error) {
			return nil, nil // no visible members
		},
	}
	svc := NewService(repo, noopTransactor{}, &mockTxLogger{})
	_, _, err := svc.GetOwnerGroup(context.Background(), testActor(1), 99)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestGetOwnerGroup_ReturnsPersistedVersion(t *testing.T) {
	repo := &mockRepo{
		listActiveOwnerByClinicsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{{
				GroupID:              7,
				GroupCreatedClinicID: 1,
				ClinicID:             1,
				OwnerID:              100,
			}}, nil
		},
		findOwnerGroupByIDFn: func(_ context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: groupID, CreatedClinicID: 3, Version: 4}, nil
		},
	}
	svc := NewService(repo, noopTransactor{}, &mockTxLogger{})
	group, members, err := svc.GetOwnerGroup(context.Background(), testActor(1), 7)
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, uint64(7), group.ID)
	assert.Equal(t, uint64(3), group.CreatedClinicID)
	assert.Equal(t, int64(4), group.Version)
	assert.Len(t, members, 1)
}

func TestGetPetGroup_ReturnsPersistedOwnerGroupAndVersion(t *testing.T) {
	repo := &mockRepo{
		listActivePetByClinicsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.PetIdentityGroupMember, error) {
			return []model.PetIdentityGroupMember{{
				GroupID:              9,
				GroupCreatedClinicID: 1,
				ClinicID:             1,
				PetID:                200,
			}}, nil
		},
		findPetGroupByIDFn: func(_ context.Context, groupID uint64) (*model.PetIdentityGroup, error) {
			return &model.PetIdentityGroup{
				ID:                        groupID,
				CreatedClinicID:           1,
				OwnerGroupCreatedClinicID: 2,
				OwnerGroupID:              42,
				Version:                   5,
			}, nil
		},
	}
	svc := NewService(repo, noopTransactor{}, &mockTxLogger{})
	group, members, err := svc.GetPetGroup(context.Background(), testActor(1), 9)
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, uint64(42), group.OwnerGroupID)
	assert.Equal(t, uint64(2), group.OwnerGroupCreatedClinicID)
	assert.Equal(t, int64(5), group.Version)
	assert.Len(t, members, 1)
}

func TestListLinkedTreatmentHistory_DefaultExcludesLinked(t *testing.T) {
	var sawPairs []ClinicPetPair
	repo := &mockRepo{
		resolveLinkedPetPairsFn: func(_ context.Context, seedClinicID, seedPetID uint64, _ []uint64) ([]ClinicPetPair, error) {
			return []ClinicPetPair{
				{ClinicID: seedClinicID, PetID: seedPetID},
				{ClinicID: 2, PetID: 200},
			}, nil
		},
		listLinkedTreatmentHistoryFn: func(_ context.Context, pairs []ClinicPetPair, _, _ int) ([]LinkedTreatmentHistoryItem, int64, error) {
			sawPairs = pairs
			return nil, 0, nil
		},
	}
	svc := NewService(repo, noopTransactor{}, &mockTxLogger{})
	_, _, err := svc.ListLinkedTreatmentHistory(context.Background(), testActor(1, 2), 1, 100, false, 1, 20)
	require.NoError(t, err)
	require.Len(t, sawPairs, 1)
	assert.Equal(t, ClinicPetPair{ClinicID: 1, PetID: 100}, sawPairs[0])
}

func TestListLinkedTreatmentHistory_IncludeLinkedUsesCorrelatedPairs(t *testing.T) {
	var sawPairs []ClinicPetPair
	repo := &mockRepo{
		resolveLinkedPetPairsFn: func(_ context.Context, seedClinicID, seedPetID uint64, clinics []uint64) ([]ClinicPetPair, error) {
			assert.Equal(t, []uint64{1, 2}, clinics)
			return []ClinicPetPair{
				{ClinicID: seedClinicID, PetID: seedPetID},
				{ClinicID: 2, PetID: 200},
			}, nil
		},
		listLinkedTreatmentHistoryFn: func(_ context.Context, pairs []ClinicPetPair, _, _ int) ([]LinkedTreatmentHistoryItem, int64, error) {
			sawPairs = pairs
			return []LinkedTreatmentHistoryItem{{TreatmentID: 1}}, 1, nil
		},
	}
	svc := NewService(repo, noopTransactor{}, &mockTxLogger{})
	items, total, err := svc.ListLinkedTreatmentHistory(context.Background(), testActor(1, 2), 1, 100, true, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Len(t, sawPairs, 2)
}

func TestCreateOwnerGroup_IdempotentRetry(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn: func(_ context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
			out := make([]model.Owner, 0, len(refs))
			for _, r := range refs {
				out = append(out, model.Owner{ID: r.OwnerID, ClinicID: r.ClinicID})
			}
			return out, nil
		},
		findActiveOwnerMembershipFn: func(_ context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error) {
			return &model.OwnerIdentityGroupMember{
				GroupID: 5, ClinicID: clinicID, OwnerID: ownerID, GroupCreatedClinicID: 1,
			}, nil
		},
		lockOwnerGroupByIDFn: func(_ context.Context, groupID uint64) (*model.OwnerIdentityGroup, error) {
			return &model.OwnerIdentityGroup{ID: groupID, CreatedClinicID: 1, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(_ context.Context, _ uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{
				{GroupID: 5, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1},
				{GroupID: 5, ClinicID: 2, OwnerID: 20, GroupCreatedClinicID: 1},
			}, nil
		},
	}
	auditLog := &mockTxLogger{}
	svc := NewService(repo, noopTransactor{}, auditLog)

	group, members, err := svc.CreateOwnerGroup(context.Background(), testActor(1, 2), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: 10},
		{ClinicID: 2, OwnerID: 20},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(5), group.ID)
	assert.Len(t, members, 2)
	assert.Equal(t, 0, repo.createOwnerGroupCalled, "idempotent retry must not create new group")
	assert.Equal(t, 0, auditLog.called)
}

// Ensure Transactor interface is satisfied by persistence implementation used in wiring.
var _ persistence.Transactor = noopTransactor{}
