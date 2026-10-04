package identitylink

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const permissionReadCanStillRace = "a permission change which commits after the in-transaction read can still race"

type scriptedAuthorizer struct {
	decision LinkWriteDecision
	err      error
	calls    int
	saw      []uint64
}

func (s *scriptedAuthorizer) AuthorizeLinkWrite(_ context.Context, _ uint64, clinicIDs []uint64) (LinkWriteDecision, error) {
	s.calls++
	s.saw = append([]uint64(nil), clinicIDs...)
	return s.decision, s.err
}

func denyEdit(clinicIDs ...uint64) *scriptedAuthorizer {
	return &scriptedAuthorizer{decision: LinkWriteDecision{EditDeniedClinicIDs: clinicIDs}}
}

type recordingTransactor struct {
	committed bool
}

func (t *recordingTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	t.committed = true
	return nil
}

type captureFailLogger struct {
	entry *audit.Entry
}

func (l *captureFailLogger) LogEntryTx(_ context.Context, input *audit.Entry) error {
	copied := *input
	l.entry = &copied
	return errors.New("audit write failed")
}

func ownersForRefs(_ context.Context, refs []OwnerMemberRef) ([]model.Owner, error) {
	out := make([]model.Owner, 0, len(refs))
	for _, ref := range refs {
		out = append(out, model.Owner{ID: ref.OwnerID, ClinicID: ref.ClinicID})
	}
	return out, nil
}

func petsForRefs(refs []PetMemberRef, ownerID func(uint64) uint64) ([]model.Pet, error) {
	out := make([]model.Pet, 0, len(refs))
	for _, ref := range refs {
		out = append(out, model.Pet{ID: ref.PetID, ClinicID: ref.ClinicID, OwnerID: ownerID(ref.ClinicID)})
	}
	return out, nil
}

func assertNoLinkWrites(t *testing.T, repo *mockRepo) {
	t.Helper()
	assert.Zero(t, repo.createOwnerGroupCalled)
	assert.Zero(t, repo.createOwnerMembersCalled)
	assert.Zero(t, repo.softDeleteOwnerCalled)
	assert.Zero(t, repo.softDeleteGroupCalled)
	assert.Zero(t, repo.createPetGroupCalled)
	assert.Zero(t, repo.createPetMembersCalled)
}

func TestClassifyLinkWrite_SeparatesAdminMembershipAndEdit(t *testing.T) {
	admin := classifyLinkWrite(true, []uint64{1, 2}, nil, nil)
	assert.True(t, admin.SystemAdmin)
	assert.Empty(t, admin.NonMemberClinicIDs)
	assert.Empty(t, admin.EditDeniedClinicIDs)

	nonMember := classifyLinkWrite(false, []uint64{1, 2}, []uint64{1}, []uint64{1})
	assert.Equal(t, []uint64{2}, nonMember.NonMemberClinicIDs)
	assert.Empty(t, nonMember.EditDeniedClinicIDs)

	editDenied := classifyLinkWrite(false, []uint64{1, 2}, []uint64{1, 2}, []uint64{1})
	assert.Empty(t, editDenied.NonMemberClinicIDs)
	assert.Equal(t, []uint64{2}, editDenied.EditDeniedClinicIDs)
}

func TestLinkWrite_PermissionChangeAfterReadCanStillRace(t *testing.T) {
	// a permission change which commits after the in-transaction read can still race.
	body, err := os.ReadFile("link_write_auth.go")
	require.NoError(t, err)
	source := string(body)
	require.Contains(t, source, permissionReadCanStillRace)
	require.NotContains(t, source, "FOR UPDATE")
	require.NotContains(t, source, "clause.Locking")
	require.NotContains(t, source, "LockPermissionPolicy")
}

func TestCreateOwnerGroup_MissingReaderAndFailedLookupLeaveRowsUnchanged(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn:                ownersForRefs,
		findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
	}
	refs := []OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}}
	actor := testActor(1, 2)

	_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, nil).CreateOwnerGroup(context.Background(), actor, refs)
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Contains(t, err.Error(), errLinkWriteReaderMissing)
	assertNoLinkWrites(t, repo)

	repo.createOwnerGroupCalled = 0
	auth := &scriptedAuthorizer{err: errors.New("membership lookup failed")}
	_, _, err = NewService(repo, noopTransactor{}, &mockTxLogger{}, auth).CreateOwnerGroup(context.Background(), actor, refs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "membership lookup failed")
	assertNoLinkWrites(t, repo)
	assert.Equal(t, 1, auth.calls)
}

func TestCreateOwnerGroup_MembershipDenialIsDistinctFromEditDenial(t *testing.T) {
	repo := func() *mockRepo {
		return &mockRepo{
			lockOwnersFn:                ownersForRefs,
			findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
		}
	}
	refs := []OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}}
	actor := testActor(1, 2)

	nonMember := &scriptedAuthorizer{decision: LinkWriteDecision{NonMemberClinicIDs: []uint64{2}}}
	store := repo()
	_, _, err := NewService(store, noopTransactor{}, &mockTxLogger{}, nonMember).CreateOwnerGroup(context.Background(), actor, refs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a current member")
	assert.NotContains(t, err.Error(), "identity-links edit is required")
	assertNoLinkWrites(t, store)

	editDenied := denyEdit(2)
	store = repo()
	_, _, err = NewService(store, noopTransactor{}, &mockTxLogger{}, editDenied).CreateOwnerGroup(context.Background(), actor, refs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identity-links edit is required")
	assert.NotContains(t, err.Error(), "not a current member")
	assertNoLinkWrites(t, store)
}

func TestOwnerWrites_EditOnAOnlyRejectsClinicBWithoutMovingAnchor(t *testing.T) {
	t.Run("create added member B", func(t *testing.T) {
		repo := &mockRepo{
			lockOwnersFn:                ownersForRefs,
			findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
		}
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).CreateOwnerGroup(
			context.Background(), testActor(1, 2),
			[]OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assertNoLinkWrites(t, repo)
	})

	t.Run("create anchor only B", func(t *testing.T) {
		repo := &mockRepo{
			lockOwnersFn:                ownersForRefs,
			findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
		}
		actor := testActor(1, 2)
		actor.HomeClinicID = 2
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).CreateOwnerGroup(
			context.Background(), actor,
			[]OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 1, OwnerID: 11}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Contains(t, err.Error(), "identity-links edit is required")
		assertNoLinkWrites(t, repo)
	})

	t.Run("create existing member B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 5, CreatedClinicID: 2, Version: 1}
		repo := &mockRepo{
			lockOwnersFn: ownersForRefs,
			findActiveOwnerMembershipFn: func(_ context.Context, clinicID, ownerID uint64) (*model.OwnerIdentityGroupMember, error) {
				return &model.OwnerIdentityGroupMember{GroupID: group.ID, ClinicID: clinicID, OwnerID: ownerID}, nil
			},
			lockOwnerGroupByIDFn: func(context.Context, uint64) (*model.OwnerIdentityGroup, error) { return group, nil },
			listActiveOwnerMembersFn: func(context.Context, uint64) ([]model.OwnerIdentityGroupMember, error) {
				return []model.OwnerIdentityGroupMember{
					{GroupID: group.ID, ClinicID: 1, OwnerID: 10},
					{GroupID: group.ID, ClinicID: 2, OwnerID: 20},
				}, nil
			},
		}
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).CreateOwnerGroup(
			context.Background(), testActor(1, 2),
			[]OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(2), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("add existing member B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		repo := ownerManageRepo(group, []uint64{1, 2})
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).AddOwnerMembers(
			context.Background(), testActor(1, 2), group.ID, []OwnerMemberRef{{ClinicID: 1, OwnerID: 99}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("add added member B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		repo := ownerManageRepo(group, []uint64{1})
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).AddOwnerMembers(
			context.Background(), testActor(1, 2), group.ID, []OwnerMemberRef{{ClinicID: 2, OwnerID: 20}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("add anchor only B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 2, Version: 1}
		repo := ownerManageRepo(group, []uint64{1})
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).AddOwnerMembers(
			context.Background(), testActor(1, 2), group.ID, []OwnerMemberRef{{ClinicID: 1, OwnerID: 99}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(2), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("unlink target B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		repo := ownerManageRepo(group, []uint64{1, 2})
		err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).UnlinkOwnerMember(
			context.Background(), testActor(1, 2), group.ID, OwnerMemberRef{ClinicID: 2, OwnerID: 101},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("unlink while existing member B is not the target", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		repo := ownerManageRepo(group, []uint64{1, 2})
		err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).UnlinkOwnerMember(
			context.Background(), testActor(1, 2), group.ID, OwnerMemberRef{ClinicID: 1, OwnerID: 100},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("unlink anchor only B", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 2, Version: 1}
		repo := ownerManageRepo(group, []uint64{1})
		err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).UnlinkOwnerMember(
			context.Background(), testActor(1, 2), group.ID, OwnerMemberRef{ClinicID: 1, OwnerID: 100},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(2), group.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})
}

func ownerManageRepo(group *model.OwnerIdentityGroup, memberClinics []uint64) *mockRepo {
	members := make([]model.OwnerIdentityGroupMember, 0, len(memberClinics))
	for i, clinicID := range memberClinics {
		members = append(members, model.OwnerIdentityGroupMember{
			ID: uint64(i + 1), GroupID: group.ID, ClinicID: clinicID, OwnerID: uint64(100 + i), GroupCreatedClinicID: group.CreatedClinicID,
		})
	}
	return &mockRepo{
		lockOwnerGroupByIDFn: func(context.Context, uint64) (*model.OwnerIdentityGroup, error) { return group, nil },
		listActiveOwnerMembersFn: func(context.Context, uint64) ([]model.OwnerIdentityGroupMember, error) {
			return members, nil
		},
		lockOwnersFn: ownersForRefs,
		findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) {
			return nil, nil
		},
	}
}

func TestPetWrites_EditOnAOnlyRejectsParentOrPetClinicBWithoutMovingAnchor(t *testing.T) {
	t.Run("create added pet clinic B", func(t *testing.T) {
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petCreateRepo(owner, []uint64{1, 2})
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).CreatePetGroup(
			context.Background(), testActor(1, 2), owner.ID,
			[]PetMemberRef{{ClinicID: 1, PetID: 10}, {ClinicID: 2, PetID: 20}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), owner.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("create parent clinic is not a pet target", func(t *testing.T) {
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petCreateRepo(owner, []uint64{1, 4})
		auth := denyEdit(4)
		actor := testActor(1, 2, 4)
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, auth).CreatePetGroup(
			context.Background(), actor, owner.ID,
			[]PetMemberRef{{ClinicID: 1, PetID: 10}, {ClinicID: 2, PetID: 20}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Contains(t, err.Error(), "identity-links edit is required")
		assert.Contains(t, auth.saw, uint64(4))
		assert.Equal(t, uint64(1), owner.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("add parent clinic is not a pet target", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 1, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petManageRepo(petGroup, owner, []uint64{1}, []uint64{1, 4})
		auth := denyEdit(4)
		errGroup, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, auth).AddPetMembers(
			context.Background(), testActor(1, 2, 4), petGroup.ID, []PetMemberRef{{ClinicID: 1, PetID: 99}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, errGroup)
		assert.Contains(t, auth.saw, uint64(4))
		assert.Equal(t, uint64(1), petGroup.CreatedClinicID)
		assert.Equal(t, uint64(1), owner.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("unlink parent clinic is not the unlink target", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 1, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petManageRepo(petGroup, owner, []uint64{1}, []uint64{1, 4})
		auth := denyEdit(4)
		err := NewService(repo, noopTransactor{}, &mockTxLogger{}, auth).UnlinkPetMember(
			context.Background(), testActor(1, 4), petGroup.ID, PetMemberRef{ClinicID: 1, PetID: 100},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Contains(t, auth.saw, uint64(4))
		assert.Equal(t, uint64(1), petGroup.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("add anchor only B", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 2, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petManageRepo(petGroup, owner, []uint64{1}, []uint64{1})
		_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).AddPetMembers(
			context.Background(), testActor(1, 2), petGroup.ID, []PetMemberRef{{ClinicID: 1, PetID: 99}},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(2), petGroup.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})

	t.Run("unlink target B", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 1, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petManageRepo(petGroup, owner, []uint64{1, 2}, []uint64{1, 2})
		err := NewService(repo, noopTransactor{}, &mockTxLogger{}, denyEdit(2)).UnlinkPetMember(
			context.Background(), testActor(1, 2), petGroup.ID, PetMemberRef{ClinicID: 2, PetID: 101},
		)
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, uint64(1), petGroup.CreatedClinicID)
		assertNoLinkWrites(t, repo)
	})
}

func petCreateRepo(owner *model.OwnerIdentityGroup, ownerClinics []uint64) *mockRepo {
	members := ownerMembers(owner, ownerClinics)
	return &mockRepo{
		lockOwnerGroupByIDFn: func(context.Context, uint64) (*model.OwnerIdentityGroup, error) { return owner, nil },
		listActiveOwnerMembersFn: func(context.Context, uint64) ([]model.OwnerIdentityGroupMember, error) {
			return members, nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			return petsForRefs(refs, func(clinicID uint64) uint64 { return clinicID })
		},
		isOwnerActiveInGroupFn: func(context.Context, uint64, uint64, uint64) (bool, error) { return true, nil },
		findActivePetMembershipFn: func(context.Context, uint64, uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
	}
}

func petManageRepo(group *model.PetIdentityGroup, owner *model.OwnerIdentityGroup, petClinics, ownerClinics []uint64) *mockRepo {
	petMembers := make([]model.PetIdentityGroupMember, 0, len(petClinics))
	for i, clinicID := range petClinics {
		petMembers = append(petMembers, model.PetIdentityGroupMember{
			ID: uint64(i + 1), GroupID: group.ID, ClinicID: clinicID, PetID: uint64(100 + i), GroupCreatedClinicID: group.CreatedClinicID,
		})
	}
	return &mockRepo{
		lockPetGroupByIDFn: func(context.Context, uint64) (*model.PetIdentityGroup, error) { return group, nil },
		listActivePetMembersFn: func(context.Context, uint64) ([]model.PetIdentityGroupMember, error) {
			return petMembers, nil
		},
		findOwnerGroupByIDFn: func(context.Context, uint64) (*model.OwnerIdentityGroup, error) { return owner, nil },
		listActiveOwnerMembersFn: func(context.Context, uint64) ([]model.OwnerIdentityGroupMember, error) {
			return ownerMembers(owner, ownerClinics), nil
		},
		lockPetsFn: func(_ context.Context, refs []PetMemberRef) ([]model.Pet, error) {
			return petsForRefs(refs, func(uint64) uint64 { return 1 })
		},
		isOwnerActiveInGroupFn: func(context.Context, uint64, uint64, uint64) (bool, error) { return true, nil },
		findActivePetMembershipFn: func(context.Context, uint64, uint64) (*model.PetIdentityGroupMember, error) {
			return nil, nil
		},
	}
}

func ownerMembers(owner *model.OwnerIdentityGroup, clinicIDs []uint64) []model.OwnerIdentityGroupMember {
	out := make([]model.OwnerIdentityGroupMember, 0, len(clinicIDs))
	for i, clinicID := range clinicIDs {
		out = append(out, model.OwnerIdentityGroupMember{
			ID: uint64(i + 1), GroupID: owner.ID, ClinicID: clinicID, OwnerID: uint64(10 + i), GroupCreatedClinicID: owner.CreatedClinicID,
		})
	}
	return out
}

func TestSixWrites_SucceedWhenEveryAffectedClinicHasEdit(t *testing.T) {
	auth := allowLinkWrites{}
	logger := &mockTxLogger{}

	t.Run("create owner keeps home clinic", func(t *testing.T) {
		repo := &mockRepo{
			lockOwnersFn:                ownersForRefs,
			findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
			createOwnerGroupFn: func(_ context.Context, group *model.OwnerIdentityGroup) error {
				group.ID = 7
				return nil
			},
		}
		actor := testActor(1, 2)
		actor.HomeClinicID = 1
		group, members, err := NewService(repo, noopTransactor{}, logger, auth).CreateOwnerGroup(
			context.Background(), actor,
			[]OwnerMemberRef{{ClinicID: 2, OwnerID: 20}, {ClinicID: 1, OwnerID: 10}},
		)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assert.Len(t, members, 2)
		require.NotEmpty(t, logger.calls)
		facts := logger.calls[len(logger.calls)-1].NewValue.(map[string]any)
		assert.Equal(t, false, facts["actor_is_system_admin"])
		assert.Contains(t, facts["affected_clinic_ids"].([]uint64), uint64(1))
	})

	t.Run("add owner response follows verified clinics", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		var listed []uint64
		repo := ownerManageRepo(group, []uint64{1, 3})
		repo.listActiveOwnerByClinicsFn = func(_ context.Context, _ uint64, clinicIDs []uint64) ([]model.OwnerIdentityGroupMember, error) {
			listed = append([]uint64(nil), clinicIDs...)
			return []model.OwnerIdentityGroupMember{{ClinicID: 1, OwnerID: 100}, {ClinicID: 3, OwnerID: 102}}, nil
		}
		actor := testActor(1, 2, 3)
		_, members, err := NewService(repo, noopTransactor{}, logger, auth).AddOwnerMembers(
			context.Background(), actor, group.ID, []OwnerMemberRef{{ClinicID: 1, OwnerID: 99}},
		)
		require.NoError(t, err)
		assert.Equal(t, []uint64{1, 2, 3}, listed)
		assert.Equal(t, uint64(3), members[1].ClinicID)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
	})

	t.Run("unlink owner", func(t *testing.T) {
		group := &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}
		repo := ownerManageRepo(group, []uint64{1})
		err := NewService(repo, noopTransactor{}, logger, auth).UnlinkOwnerMember(
			context.Background(), testActor(1), group.ID, OwnerMemberRef{ClinicID: 1, OwnerID: 100},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, repo.softDeleteOwnerCalled)
	})

	t.Run("create pet keeps home clinic", func(t *testing.T) {
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		repo := petCreateRepo(owner, []uint64{1, 2})
		actor := testActor(1, 2)
		group, _, err := NewService(repo, noopTransactor{}, logger, auth).CreatePetGroup(
			context.Background(), actor, owner.ID,
			[]PetMemberRef{{ClinicID: 2, PetID: 20}, {ClinicID: 1, PetID: 10}},
		)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), group.CreatedClinicID)
		assert.Equal(t, 1, repo.createPetGroupCalled)
	})

	t.Run("add pet response follows verified clinics", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 1, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		var listed []uint64
		repo := petManageRepo(petGroup, owner, []uint64{1}, []uint64{1, 2})
		repo.listActivePetByClinicsFn = func(_ context.Context, _ uint64, clinicIDs []uint64) ([]model.PetIdentityGroupMember, error) {
			listed = append([]uint64(nil), clinicIDs...)
			return []model.PetIdentityGroupMember{{ClinicID: 1, PetID: 100}, {ClinicID: 2, PetID: 200}}, nil
		}
		_, members, err := NewService(repo, noopTransactor{}, logger, auth).AddPetMembers(
			context.Background(), testActor(1, 2), petGroup.ID, []PetMemberRef{{ClinicID: 1, PetID: 99}},
		)
		require.NoError(t, err)
		assert.Equal(t, []uint64{1, 2}, listed)
		assert.Len(t, members, 2)
		assert.Equal(t, uint64(1), petGroup.CreatedClinicID)
	})

	t.Run("unlink pet", func(t *testing.T) {
		petGroup := &model.PetIdentityGroup{ID: 8, CreatedClinicID: 1, OwnerGroupID: 4, OwnerGroupCreatedClinicID: 1, Version: 1}
		owner := &model.OwnerIdentityGroup{ID: 4, CreatedClinicID: 1, Version: 1}
		deleted := false
		repo := petManageRepo(petGroup, owner, []uint64{1}, []uint64{1})
		repo.softDeletePetMemberFn = func(context.Context, uint64) error {
			deleted = true
			return nil
		}
		err := NewService(repo, noopTransactor{}, logger, auth).UnlinkPetMember(
			context.Background(), testActor(1), petGroup.ID, PetMemberRef{ClinicID: 1, PetID: 100},
		)
		require.NoError(t, err)
		assert.True(t, deleted)
	})
}

func TestAuthorizeRunsAfterGroupLockAndBeforeFirstWrite(t *testing.T) {
	locked := false
	repo := &mockRepo{
		lockOwnerGroupByIDFn: func(context.Context, uint64) (*model.OwnerIdentityGroup, error) {
			locked = true
			return &model.OwnerIdentityGroup{ID: 9, CreatedClinicID: 1, Version: 1}, nil
		},
		listActiveOwnerMembersFn: func(context.Context, uint64) ([]model.OwnerIdentityGroupMember, error) {
			return []model.OwnerIdentityGroupMember{{GroupID: 9, ClinicID: 1, OwnerID: 10, GroupCreatedClinicID: 1}}, nil
		},
		lockOwnersFn: ownersForRefs,
		findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) {
			return nil, nil
		},
	}
	checking := &orderAuthorizer{repo: repo, locked: &locked}
	_, _, err := NewService(repo, noopTransactor{}, &mockTxLogger{}, checking).AddOwnerMembers(
		context.Background(), testActor(1), 9, []OwnerMemberRef{{ClinicID: 1, OwnerID: 99}},
	)
	require.NoError(t, err)
	assert.True(t, checking.sawLockedBeforeWrite)
	assert.Equal(t, 1, repo.createOwnerMembersCalled)
}

type orderAuthorizer struct {
	repo                 *mockRepo
	locked               *bool
	sawLockedBeforeWrite bool
}

func (o *orderAuthorizer) AuthorizeLinkWrite(context.Context, uint64, []uint64) (LinkWriteDecision, error) {
	if o.locked != nil && *o.locked && o.repo.createOwnerMembersCalled == 0 {
		o.sawLockedBeforeWrite = true
	}
	return LinkWriteDecision{}, nil
}

func TestSystemAdminWriteKeepsAllActiveClinicsAndAuditsRollback(t *testing.T) {
	repo := &mockRepo{
		lockOwnersFn:                ownersForRefs,
		findActiveOwnerMembershipFn: func(context.Context, uint64, uint64) (*model.OwnerIdentityGroupMember, error) { return nil, nil },
		createOwnerGroupFn: func(_ context.Context, group *model.OwnerIdentityGroup) error {
			group.ID = 3
			return nil
		},
	}
	auth := &scriptedAuthorizer{decision: LinkWriteDecision{SystemAdmin: true, EditDeniedClinicIDs: []uint64{1, 2}}}
	logger := &mockTxLogger{}
	group, _, err := NewService(repo, noopTransactor{}, logger, auth).CreateOwnerGroup(
		context.Background(), testActor(1, 2),
		[]OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}},
	)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), group.CreatedClinicID)
	require.NotEmpty(t, logger.calls)
	facts := logger.calls[0].NewValue.(map[string]any)
	assert.Equal(t, true, facts["actor_is_system_admin"])
	assert.ElementsMatch(t, []uint64{1, 2}, facts["affected_clinic_ids"].([]uint64))

	failLog := &captureFailLogger{}
	tx := &recordingTransactor{}
	repo.createOwnerGroupCalled = 0
	_, _, err = NewService(repo, tx, failLog, auth).CreateOwnerGroup(
		context.Background(), testActor(1, 2),
		[]OwnerMemberRef{{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit")
	assert.False(t, tx.committed)
	require.NotNil(t, failLog.entry)
	failedFacts := failLog.entry.NewValue.(map[string]any)
	assert.Equal(t, true, failedFacts["actor_is_system_admin"])
	assert.ElementsMatch(t, []uint64{1, 2}, failedFacts["affected_clinic_ids"].([]uint64))
}

type handlerWriteStub struct {
	calls      int
	verified   []uint64
	homeClinic uint64
}

func (s *handlerWriteStub) note(actor ActorContext) {
	s.calls++
	s.verified = append([]uint64(nil), actor.VerifiedClinics...)
	s.homeClinic = actor.HomeClinicID
}

func (s *handlerWriteStub) SearchOwners(context.Context, ActorContext, string, int) ([]model.Owner, error) {
	return nil, nil
}
func (s *handlerWriteStub) SearchPets(context.Context, ActorContext, string, int) ([]model.Pet, error) {
	return nil, nil
}
func (s *handlerWriteStub) GetOwnerGroup(context.Context, ActorContext, uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	return nil, nil, nil
}
func (s *handlerWriteStub) GetPetGroup(context.Context, ActorContext, uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	return nil, nil, nil
}
func (s *handlerWriteStub) FindOwnerGroupByMember(context.Context, ActorContext, uint64, uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	return nil, nil, nil
}
func (s *handlerWriteStub) FindPetGroupByMember(context.Context, ActorContext, uint64, uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	return nil, nil, nil
}
func (s *handlerWriteStub) CreateOwnerGroup(_ context.Context, actor ActorContext, _ []OwnerMemberRef) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	s.note(actor)
	return &model.OwnerIdentityGroup{ID: 1, CreatedClinicID: actor.HomeClinicID, Version: 1}, []model.OwnerIdentityGroupMember{
		{ClinicID: 1, OwnerID: 10}, {ClinicID: 2, OwnerID: 20},
	}, nil
}
func (s *handlerWriteStub) AddOwnerMembers(_ context.Context, actor ActorContext, _ uint64, _ []OwnerMemberRef) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	s.note(actor)
	return &model.OwnerIdentityGroup{ID: 1, CreatedClinicID: 1}, nil, nil
}
func (s *handlerWriteStub) UnlinkOwnerMember(_ context.Context, actor ActorContext, _ uint64, _ OwnerMemberRef) error {
	s.note(actor)
	return nil
}
func (s *handlerWriteStub) CreatePetGroup(_ context.Context, actor ActorContext, _ uint64, _ []PetMemberRef) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	s.note(actor)
	return &model.PetIdentityGroup{ID: 1, CreatedClinicID: actor.HomeClinicID}, nil, nil
}
func (s *handlerWriteStub) AddPetMembers(_ context.Context, actor ActorContext, _ uint64, _ []PetMemberRef) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	s.note(actor)
	return &model.PetIdentityGroup{ID: 1, CreatedClinicID: 1}, nil, nil
}
func (s *handlerWriteStub) UnlinkPetMember(_ context.Context, actor ActorContext, _ uint64, _ PetMemberRef) error {
	s.note(actor)
	return nil
}
func (s *handlerWriteStub) ListLinkedTreatmentHistory(context.Context, ActorContext, uint64, uint64, bool, int, int) ([]LinkedTreatmentHistoryItem, int64, error) {
	return nil, 0, nil
}

func TestHandlerWrite_EmptyEditSetIsSingleFilterDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &handlerWriteStub{}
	h := NewHandler(stub, nil)
	calls := []struct {
		name string
		fn   func(*gin.Context)
	}{
		{"create owner", h.CreateOwnerGroup},
		{"add owner", h.AddOwnerMembers},
		{"unlink owner", h.UnlinkOwnerMember},
		{"create pet", h.CreatePetGroup},
		{"add pet", h.AddPetMembers},
		{"unlink pet", h.UnlinkPetMember},
	}
	for _, call := range calls {
		t.Run(call.name, func(t *testing.T) {
			before := stub.calls
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "9"}}
			c.Set("user_id", "17")
			c.Set("clinic_id", "1")
			c.Set("clinic_ids", []uint64{1, 2})
			c.Set("is_system_admin", false)
			httpapi.SetClinicPermissionChecker(c, func(*gin.Context, uint64, string, string) bool { return false })
			call.fn(c)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Equal(t, 1, strings.Count(w.Body.String(), `"error"`))
			assert.NotContains(t, w.Body.String(), errLinkWriteReaderMissing)
			assert.Equal(t, before, stub.calls)
		})
	}
}

func TestHandlerCreateOwner_DoesNotReplaceVerifiedClinicsWithEditSubset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &handlerWriteStub{}
	h := NewHandler(stub, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"members":[{"clinic_id":1,"owner_id":10},{"clinic_id":2,"owner_id":20}]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "17")
	c.Set("clinic_id", "1")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == 1 && resource == string(model.ResourceIdentityLinks) && action == "edit"
	})
	h.CreateOwnerGroup(c)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, []uint64{1, 2}, stub.verified)
	assert.Equal(t, uint64(1), stub.homeClinic)
	assert.Contains(t, w.Body.String(), `"created_clinic_id":1`)
	assert.Contains(t, w.Body.String(), `"clinic_id":2`)
}
