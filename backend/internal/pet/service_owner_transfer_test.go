package pet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

type petOwnerTransferCall struct {
	clinicID, petID, previousOwnerID, newOwnerID uint64
	from                                         time.Time
}

type petOwnerTransferSpy struct {
	calls  []petOwnerTransferCall
	ids    []uint64
	err    error
	sawCtx context.Context
}

func (s *petOwnerTransferSpy) TransferUpcomingPetReservations(
	ctx context.Context,
	clinicID, petID, previousOwnerID, newOwnerID uint64,
	from time.Time,
) ([]uint64, error) {
	s.sawCtx = ctx
	s.calls = append(s.calls, petOwnerTransferCall{clinicID, petID, previousOwnerID, newOwnerID, from})
	if s.err != nil {
		return nil, s.err
	}
	return s.ids, nil
}

func ownerChangeRepo(t *testing.T, tx *petCommitObserver, petID, clinicID, initialOwnerID, lockedOwnerID uint64) *mockPetRepository {
	t.Helper()
	current := &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: initialOwnerID}
	return &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return current, nil
		},
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			require.True(t, tx.inside)
			current = &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: lockedOwnerID}
			return current, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
			require.True(t, tx.inside)
			if v, ok := fields[colPetOwnerID].(uint64); ok {
				current = &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: v}
			}
			return nil
		},
	}
}

func TestPetOwnerTransfer_PassesLockedOwnersJSTBoundaryAndAudits(t *testing.T) {
	tx := &petCommitObserver{}
	const clinicID, petID = uint64(1), uint64(2)
	newOwnerID := uint64(30)
	transfer := &petOwnerTransferSpy{ids: []uint64{42, 41}}
	auditLog := &petOwnerServiceAuditDouble{}
	actorID := uint64(7)

	repo := ownerChangeRepo(t, tx, petID, clinicID, 10, 20)
	var synced []uint64
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(_ context.Context, _ uint64, ownerID uint64) error {
		require.True(t, tx.committed)
		synced = append(synced, ownerID)
		return nil
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
		&petOwnerServiceRepositoryDouble{}, tx,
		PetOwnerTransferDependencies{Reservations: transfer, Audit: auditLog})

	updated, err := svc.Update(context.Background(), clinicID, petID, &UpdatePetInput{OwnerID: &newOwnerID, ActorID: &actorID})
	require.NoError(t, err)
	require.NotNil(t, updated)

	require.Len(t, transfer.calls, 1)
	call := transfer.calls[0]
	assert.Equal(t, clinicID, call.clinicID)
	assert.Equal(t, petID, call.petID)
	assert.Equal(t, uint64(20), call.previousOwnerID, "use the locked owner, not the stale pre-lock read")
	assert.Equal(t, newOwnerID, call.newOwnerID)
	wantFrom := time.Now().In(config.JST)
	wantFrom = time.Date(wantFrom.Year(), wantFrom.Month(), wantFrom.Day(), 0, 0, 0, 0, config.JST)
	assert.Equal(t, wantFrom, call.from, "boundary is the JST start of today fixed at the transaction stage")

	require.Len(t, auditLog.entries, 1)
	entry := auditLog.entries[0]
	assert.Equal(t, model.AuditActionPetOwnerTransfer, entry.Action)
	assert.Equal(t, "pet", entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, petID, *entry.ResourceID)
	assert.Equal(t, map[string]any{"owner_id": uint64(20)}, entry.OldValue)
	assert.Equal(t, map[string]any{"owner_id": newOwnerID}, entry.NewValue)
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, actorID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	metadata, ok := entry.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []uint64{42, 41}, metadata["appointment_ids"])
	assert.Equal(t, wantFrom.Format("2006-01-02"), metadata["from_date"])

	assert.ElementsMatch(t, []uint64{20, newOwnerID}, synced, "post-commit tag sync covers both replaced owners")
}

func TestPetOwnerTransfer_DirectServiceCallAuditsSystemActor(t *testing.T) {
	tx := &petCommitObserver{}
	newOwnerID := uint64(30)
	transfer := &petOwnerTransferSpy{}
	auditLog := &petOwnerServiceAuditDouble{}
	repo := ownerChangeRepo(t, tx, 2, 1, 10, 20)
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), nil,
		&petOwnerServiceRepositoryDouble{}, tx,
		PetOwnerTransferDependencies{Reservations: transfer, Audit: auditLog})

	_, err := svc.Update(context.Background(), 1, 2, &UpdatePetInput{OwnerID: &newOwnerID})
	require.NoError(t, err)
	require.Len(t, auditLog.entries, 1)
	assert.Nil(t, auditLog.entries[0].ActorID)
	assert.Equal(t, model.AuditActorTypeSystem, auditLog.entries[0].ActorType)
}

func TestPetOwnerTransfer_FailuresDoNotCommitOrSync(t *testing.T) {
	newOwnerID := uint64(30)
	for _, tc := range []struct {
		name      string
		transfer  *petOwnerTransferSpy
		audit     *petOwnerServiceAuditDouble
		withDeps  bool
		wantCalls int
		wantAudit int
	}{
		{name: "transfer failure", transfer: &petOwnerTransferSpy{err: errors.New("transfer failed")}, audit: &petOwnerServiceAuditDouble{}, withDeps: true, wantCalls: 1, wantAudit: 0},
		{name: "audit failure", transfer: &petOwnerTransferSpy{}, audit: &petOwnerServiceAuditDouble{err: errors.New("audit failed")}, withDeps: true, wantCalls: 1, wantAudit: 1},
		{name: "missing dependencies", withDeps: false, wantCalls: 0, wantAudit: 0},
		{name: "nil reservations", transfer: nil, audit: &petOwnerServiceAuditDouble{}, withDeps: true, wantCalls: 0, wantAudit: 0},
		{name: "nil audit", transfer: &petOwnerTransferSpy{}, audit: nil, withDeps: true, wantCalls: 0, wantAudit: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &petCommitObserver{}
			repo := ownerChangeRepo(t, tx, 2, 1, 10, 20)
			synced := 0
			tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error {
				synced++
				return nil
			}}
			var deps []PetOwnerTransferDependencies
			if tc.withDeps {
				var reservations UpcomingReservationOwnerTransfer
				if tc.transfer != nil {
					reservations = tc.transfer
				}
				var audit PetOwnerAuditLogger
				if tc.audit != nil {
					audit = tc.audit
				}
				deps = []PetOwnerTransferDependencies{{Reservations: reservations, Audit: audit}}
			}
			svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
				&petOwnerServiceRepositoryDouble{}, tx, deps...)

			_, err := svc.Update(context.Background(), 1, 2, &UpdatePetInput{OwnerID: &newOwnerID})
			require.Error(t, err)
			assert.False(t, tx.committed, "failed transfer/audit must roll back the pet update")
			assert.Zero(t, synced, "no post-commit tag sync after a failed transfer")
			if tc.transfer != nil {
				assert.Len(t, tc.transfer.calls, tc.wantCalls)
			}
			if tc.audit != nil {
				assert.Len(t, tc.audit.entries, tc.wantAudit)
			}
		})
	}
}
