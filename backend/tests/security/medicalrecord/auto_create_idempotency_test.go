// Package medicalrecord — EMR-332 C15 evidence tests
// (docs/work/emr/emr-332-security/workflow-audit/c15-order-idempotency.md).
//
// Scope: deterministic, single-threaded, in-memory verification of
// MedicalRecordService.AutoCreateFromReservation order + idempotency guards.
// Every collaborator is an in-memory fake; no DB, network, notification, or
// persistent audit exists. The repo fake embeds the MedicalRecordRepository
// interface, so any repository call outside the asserted set panics (fail-loud)
// rather than silently passing.
//
// Advisory-lock contract under test (medical_record_auto_create.go):
// pg_try_advisory_xact_lock busy => acquired=false => immediate skip with no
// count, no create, no audit — NOT serialize-and-wait.
package medicalrecord

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ── in-memory fakes (all state scripted; call order recorded in seq) ─────────

// mrRepoFake embeds the interface: methods the auto-create path must not call
// stay nil and panic if reached — that is the side-effect tripwire.
type mrRepoFake struct {
	medicalrecord.MedicalRecordRepository

	seq        *[]string
	acquired   bool
	acquireErr error
	count      int64
	countErr   error
	existing   *model.MedicalRecord
	findErr    error
	createErr  error
	createID   uint64

	acquireCalls int
	countCalls   int
	findCalls    int
	createCalls  int
}

func (f *mrRepoFake) AcquireAutoCreateLock(_ context.Context, _, _ uint64, _ string) (bool, error) {
	*f.seq = append(*f.seq, "repo.AcquireAutoCreateLock")
	f.acquireCalls++
	return f.acquired, f.acquireErr
}

func (f *mrRepoFake) CountByPetAndDate(_ context.Context, _, _ uint64, _ string) (int64, error) {
	*f.seq = append(*f.seq, "repo.CountByPetAndDate")
	f.countCalls++
	return f.count, f.countErr
}

func (f *mrRepoFake) FindByAppointmentID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	*f.seq = append(*f.seq, "repo.FindByAppointmentID")
	f.findCalls++
	return f.existing, f.findErr
}

func (f *mrRepoFake) Create(_ context.Context, record *model.MedicalRecord) error {
	*f.seq = append(*f.seq, "repo.Create")
	f.createCalls++
	if f.createErr == nil && f.createID != 0 {
		record.ID = f.createID
	}
	return f.createErr
}

// resRepoFake implements the consumer-side reservation view (AssertOwnerInClinic,
// FindPetOwnerInClinic, FindByID, ReadPetLifeState,
// AssertMedicalRecordDoctorInClinic, BackfillForMedicalRecord,
// PrepareForMedicalRecordFinalization).
type resRepoFake struct {
	seq         *[]string
	ownerErr    error
	petOwnerID  uint64
	petErr      error
	appt        *model.Reservation
	apptErr     error
	lifeState   sharedkernel.PetLifeState
	lifeErr     error
	doctorErr   error
	backfillErr error
}

func (f *resRepoFake) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	*f.seq = append(*f.seq, "res.AssertOwnerInClinic")
	return f.ownerErr
}

func (f *resRepoFake) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	*f.seq = append(*f.seq, "res.FindPetOwnerInClinic")
	return f.petOwnerID, f.petErr
}

func (f *resRepoFake) FindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	*f.seq = append(*f.seq, "res.FindByID")
	return f.appt, f.apptErr
}

func (f *resRepoFake) ReadPetLifeState(_ context.Context, _, _ uint64) (sharedkernel.PetLifeState, error) {
	*f.seq = append(*f.seq, "res.ReadPetLifeState")
	return f.lifeState, f.lifeErr
}

func (f *resRepoFake) AssertMedicalRecordDoctorInClinic(_ context.Context, _, _ uint64) error {
	*f.seq = append(*f.seq, "res.AssertMedicalRecordDoctorInClinic")
	return f.doctorErr
}

func (f *resRepoFake) BackfillForMedicalRecord(_ context.Context, _, _ uint64, _, _, _ *uint64) error {
	*f.seq = append(*f.seq, "res.BackfillForMedicalRecord")
	return f.backfillErr
}

func (f *resRepoFake) PrepareForMedicalRecordFinalization(_ context.Context, _, _ uint64) error {
	*f.seq = append(*f.seq, "res.PrepareForMedicalRecordFinalization")
	return nil
}

// txFake is the in-memory transaction boundary.
type txFake struct {
	seq     *[]string
	err     error
	runFn   bool // when false, WithTx fails without running fn (fail-closed model)
	txCalls int
}

func (f *txFake) WithTx(_ context.Context, fn func(context.Context) error) error {
	f.txCalls++
	*f.seq = append(*f.seq, "tx.WithTx.begin")
	if !f.runFn {
		*f.seq = append(*f.seq, "tx.WithTx.reject")
		if f.err != nil {
			return f.err
		}
		return errors.New("tx rejected")
	}
	if err := fn(context.Background()); err != nil {
		*f.seq = append(*f.seq, "tx.WithTx.rollback")
		return err
	}
	*f.seq = append(*f.seq, "tx.WithTx.commit")
	return f.err
}

// planRepoFake implements medicalrecord.ClinicalPlanRepository.
type planRepoFake struct {
	seq         *[]string
	plan        *model.ClinicalPlan
	findErr     error
	createCalls int
	updateCalls int
}

func (f *planRepoFake) FindByMedicalRecordID(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
	*f.seq = append(*f.seq, "plan.FindByMedicalRecordID")
	return f.plan, f.findErr
}

func (f *planRepoFake) Create(_ context.Context, _ *model.ClinicalPlan) error {
	*f.seq = append(*f.seq, "plan.Create")
	f.createCalls++
	return nil
}

func (f *planRepoFake) Update(_ context.Context, _, _ uint64, _ medicalrecord.UpdateClinicalPlanInput, _ *int) error {
	*f.seq = append(*f.seq, "plan.Update")
	f.updateCalls++
	return nil
}

func (f *planRepoFake) Delete(_ context.Context, _, _ uint64) error {
	*f.seq = append(*f.seq, "plan.Delete")
	return nil
}

// auditFake implements mrAuditLogger (LogMedicalRecordChange/LogAddendumCreate)
// plus AuditLogger (LogEntry) so a subrecord-failure audit would be countable.
type auditFake struct {
	seq           *[]string
	changeCalls   int
	addendumCalls int
	entryCalls    int
}

func (f *auditFake) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, action string, _ uint64, _, _ map[string]any) error {
	*f.seq = append(*f.seq, "audit.LogMedicalRecordChange:"+action)
	f.changeCalls++
	return nil
}

func (f *auditFake) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	*f.seq = append(*f.seq, "audit.LogAddendumCreate")
	f.addendumCalls++
	return nil
}

func (f *auditFake) LogEntry(_ context.Context, _ *medicalrecord.AuditEntry) error {
	*f.seq = append(*f.seq, "audit.LogEntry")
	f.entryCalls++
	return nil
}

// lcFake implements mrLineCustomerRepo.
type lcFake struct {
	seq      *[]string
	customer *model.LineCustomer
	err      error
	calls    int
}

func (f *lcFake) FindByID(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
	*f.seq = append(*f.seq, "lc.FindByID")
	f.calls++
	return f.customer, f.err
}

// captureSlog redirects the default slog logger into a buffer for the test.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

func newSeq() *[]string { s := []string{}; return &s }

func u64(v uint64) *uint64 { return &v }

// validReservation builds a general reservation that clears every early-return
// guard: owner+pet set, no LINE fallback needed, appointment link present.
func validReservation() *model.Reservation {
	return &model.Reservation{
		ID:        9001,
		ClinicID:  7,
		StartTime: time.Date(2026, 10, 9, 15, 30, 0, 0, time.UTC), // JST 2026-10-10 00:30
		OwnerID:   u64(70),
		PetID:     u64(700),
	}
}

// matchingAppointment is the reservation-owned row returned by res.FindByID:
// owner/pet consistent, same JST day, revisit.
func matchingAppointment(r *model.Reservation) *model.Reservation {
	return &model.Reservation{
		ID:        r.ID,
		ClinicID:  r.ClinicID,
		StartTime: r.StartTime,
		OwnerID:   r.OwnerID,
		PetID:     r.PetID,
		VisitType: model.VisitTypeRevisit,
	}
}

// aliveLifeState is the explicit test-double life state for a confirmed-alive pet (EMR-491 D6).
const aliveLifeState = sharedkernel.PetLifeStateAlive

func newService(repo *mrRepoFake, res *resRepoFake, plan *planRepoFake, audit *auditFake, lc *lcFake, tx *txFake) medicalrecord.MedicalRecordService {
	return medicalrecord.NewMedicalRecordServiceWithTxAudit(
		repo, nil, plan, nil, nil, nil, lc, res, nil, audit, nil, tx,
	)
}

// ── C15: early-return order guards (no repo/tx contact at all) ───────────────

func TestC15_AutoCreateEarlyReturnGuards(t *testing.T) {
	ownerID := uint64(70)
	petID := uint64(700)
	lcID := uint64(88)
	start := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		reservation *model.Reservation
		lc          *lcFake
		wantLCCalls int
	}{
		{"nil reservation", nil, &lcFake{seq: newSeq()}, 0},
		{
			"trimming type skipped",
			&model.Reservation{ID: 1, StartTime: start, OwnerID: &ownerID, PetID: &petID,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryTrimming}},
			&lcFake{seq: newSeq()}, 0,
		},
		{
			"hospitalization name skipped",
			&model.Reservation{ID: 2, StartTime: start, OwnerID: &ownerID, PetID: &petID,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryGeneral, Name: "入院"}},
			&lcFake{seq: newSeq()}, 0,
		},
		{
			"hotel name skipped",
			&model.Reservation{ID: 3, StartTime: start, OwnerID: &ownerID, PetID: &petID,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryGeneral, Name: "ペットホテル"}},
			&lcFake{seq: newSeq()}, 0,
		},
		{
			"missing owner+pet without LINE link skipped",
			&model.Reservation{ID: 4, StartTime: start},
			&lcFake{seq: newSeq()}, 0,
		},
		{
			"LINE link present but customer has no owner -> still skipped",
			&model.Reservation{ID: 5, StartTime: start, LineCustomerID: &lcID},
			&lcFake{customer: &model.LineCustomer{ID: lcID, OwnerID: nil}}, 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seq := newSeq()
			repo := &mrRepoFake{seq: seq, acquired: true}
			res := &resRepoFake{seq: seq}
			plan := &planRepoFake{seq: seq}
			audit := &auditFake{seq: seq}
			tx := &txFake{seq: seq, runFn: true}
			tc.lc.seq = seq

			svc := newService(repo, res, plan, audit, tc.lc, tx)
			assert.NotPanics(t, func() {
				svc.AutoCreateFromReservation(context.Background(), 7, tc.reservation)
			})

			assert.Equal(t, 0, tx.txCalls, "tx must not open for skipped reservations")
			assert.Equal(t, 0, repo.acquireCalls+repo.countCalls+repo.findCalls+repo.createCalls,
				"no repository call may run for skipped reservations")
			assert.Equal(t, 0, audit.changeCalls+audit.entryCalls, "no audit write may run for skipped reservations")
			assert.Equal(t, tc.wantLCCalls, tc.lc.calls)
		})
	}
}

// ── C15: advisory-lock busy => non-blocking skip (no create, no audit) ───────

func TestC15_AutoCreateLockBusySkipsCreate(t *testing.T) {
	logs := captureSlog(t)
	seq := newSeq()
	reservation := validReservation()
	repo := &mrRepoFake{seq: seq, acquired: false} // pg_try_advisory_xact_lock => false
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq}
	audit := &auditFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	svc := newService(repo, res, plan, audit, nil, tx)
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)

	// Busy lock returns an empty result: the ONLY repo interaction is the
	// non-blocking lock attempt. No count, no create, no audit, no plan write.
	assert.Equal(t, []string{"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "tx.WithTx.commit"}, *seq)
	assert.Equal(t, 0, repo.createCalls)
	assert.Equal(t, 0, audit.changeCalls+audit.entryCalls)
	assert.Contains(t, logs.String(), "auto-create lock is busy")
}

func TestC15_AutoCreateLockErrorSkipsCreate(t *testing.T) {
	logs := captureSlog(t)
	seq := newSeq()
	reservation := validReservation()
	repo := &mrRepoFake{seq: seq, acquired: false, acquireErr: errors.New("lock unavailable")}
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq}
	audit := &auditFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	svc := newService(repo, res, plan, audit, nil, tx)
	assert.NotPanics(t, func() {
		svc.AutoCreateFromReservation(context.Background(), 7, reservation)
	})

	assert.Equal(t, []string{"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "tx.WithTx.rollback"}, *seq)
	assert.Equal(t, 0, repo.createCalls)
	assert.Equal(t, 0, audit.changeCalls+audit.entryCalls)
	assert.Contains(t, logs.String(), "failed to acquire auto-create lock")
	assert.Contains(t, logs.String(), "auto-create transaction failed")
}

// ── C15: transaction dependency absent => fail-closed, zero repo calls ──────

func TestC15_AutoCreateWithoutTransactionDependencyIsFailClosed(t *testing.T) {
	seq := newSeq()
	reservation := validReservation()
	repo := &mrRepoFake{seq: seq, acquired: true}
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq}
	audit := &auditFake{seq: seq}

	// tx = nil -> withTx returns InternalServerError before any lock/count/write.
	svc := medicalrecord.NewMedicalRecordServiceWithTxAudit(repo, nil, plan, nil, nil, nil, nil, res, nil, audit, nil, nil)
	assert.NotPanics(t, func() {
		svc.AutoCreateFromReservation(context.Background(), 7, reservation)
	})
	assert.Empty(t, *seq)
	assert.Equal(t, 0, repo.acquireCalls+repo.countCalls+repo.findCalls+repo.createCalls)
}

// ── C15: same-day existing record suppresses a duplicate create ──────────────

func TestC15_AutoCreateSameDayExistingSkipsCreate(t *testing.T) {
	logs := captureSlog(t)
	seq := newSeq()
	reservation := validReservation()
	repo := &mrRepoFake{seq: seq, acquired: true, count: 1}
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq}
	audit := &auditFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	svc := newService(repo, res, plan, audit, nil, tx)
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)

	assert.Equal(t, []string{
		"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate", "tx.WithTx.commit",
	}, *seq)
	assert.Equal(t, 0, repo.createCalls)
	assert.Equal(t, 0, audit.changeCalls+audit.entryCalls)
	assert.Contains(t, logs.String(), "same-day record already exists")
}

// ── C15: appointment-linked existing record repairs without create side effects ─

func TestC15_AutoCreateExistingAppointmentRecordRepairsWithoutCreate(t *testing.T) {
	seq := newSeq()
	reservation := validReservation()
	existing := &model.MedicalRecord{
		ID:            555,
		ClinicID:      7,
		Date:          time.Date(2026, 10, 10, 0, 0, 0, 0, config_JST()),
		AppointmentID: u64(reservation.ID),
	}
	repo := &mrRepoFake{seq: seq, acquired: true, count: 0, existing: existing}
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq, plan: &model.ClinicalPlan{ID: 42, MedicalRecordID: 555}}
	audit := &auditFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	svc := newService(repo, res, plan, audit, nil, tx)
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)

	// Lock -> count -> context validation -> appointment row lock (Backfill)
	// -> FindByID -> re-validation -> deceased check -> FindByAppointmentID hit.
	want := []string{
		"tx.WithTx.begin",
		"repo.AcquireAutoCreateLock",
		"repo.CountByPetAndDate",
		"res.AssertOwnerInClinic",
		"res.FindPetOwnerInClinic",
		"res.BackfillForMedicalRecord",
		"res.FindByID",
		"res.AssertOwnerInClinic",
		"res.FindPetOwnerInClinic",
		"res.ReadPetLifeState",
		"repo.FindByAppointmentID",
		"tx.WithTx.commit",
		"plan.FindByMedicalRecordID",
	}
	assert.Equal(t, want, *seq)
	assert.Equal(t, 0, repo.createCalls, "existing appointment record must not be re-created")
	assert.Equal(t, 0, audit.changeCalls, "create audit must not repeat for the repair path")
	assert.Equal(t, 0, audit.entryCalls, "no failure audit when repair succeeds")
	assert.Equal(t, 0, plan.createCalls+plan.updateCalls, "existing plan row must not be rewritten by empty input")
}

// ── C15: first visit creates exactly once; replay is suppressed ──────────────

func TestC15_AutoCreateCreatesOnceThenReplaySuppresses(t *testing.T) {
	seq := newSeq()
	reservation := validReservation()
	repo := &mrRepoFake{seq: seq, acquired: true, count: 0, createID: 6001}
	res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
	plan := &planRepoFake{seq: seq, findErr: apperrors.WrapNotFound("clinical_plan", "")}
	audit := &auditFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	svc := newService(repo, res, plan, audit, nil, tx)
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)

	want := []string{
		"tx.WithTx.begin",
		"repo.AcquireAutoCreateLock",
		"repo.CountByPetAndDate",
		"res.AssertOwnerInClinic",
		"res.FindPetOwnerInClinic",
		"res.BackfillForMedicalRecord",
		"res.FindByID",
		"res.AssertOwnerInClinic",
		"res.FindPetOwnerInClinic",
		"res.ReadPetLifeState",
		"repo.FindByAppointmentID",
		"repo.Create",
		"tx.WithTx.commit",
		"audit.LogMedicalRecordChange:create",
		"plan.FindByMedicalRecordID",
		"plan.Create",
	}
	assert.Equal(t, want, *seq)
	assert.Equal(t, 1, repo.createCalls)
	assert.Equal(t, 1, audit.changeCalls)
	assert.Equal(t, 0, audit.entryCalls)

	// Replay: the same reservation arrives again (e.g. status re-assertion or
	// duplicate event). CountByPetAndDate now reports the committed row, so the
	// second pass must stop after Acquire+Count with zero additional writes.
	*seq = []string{}
	repo.count = 1
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)

	assert.Equal(t, []string{
		"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate", "tx.WithTx.commit",
	}, *seq)
	assert.Equal(t, 1, repo.createCalls, "replayed auto-create must not create a second record")
	assert.Equal(t, 1, audit.changeCalls, "create audit must fire exactly once across replay")
}

// ── C15: cross-link / deceased / doctor violations reject before Create ──────

func TestC15_AutoCreateOrderViolationsRejectBeforeCreate(t *testing.T) {
	reservation := validReservation()
	cases := []struct {
		name    string
		mutate  func(res *resRepoFake)
		wantSeq []string
	}{
		{
			name: "pet bound to a different owner is rejected",
			mutate: func(res *resRepoFake) {
				res.petOwnerID = 999 // pet belongs to another owner
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic", "tx.WithTx.rollback",
			},
		},
		{
			name: "owner outside clinic is rejected",
			mutate: func(res *resRepoFake) {
				res.ownerErr = apperrors.WrapNotFound("owner", "70")
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "tx.WithTx.rollback",
			},
		},
		{
			name: "appointment row lock failure is rejected",
			mutate: func(res *resRepoFake) {
				res.backfillErr = apperrors.WrapConflict("appointment locked")
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.BackfillForMedicalRecord", "tx.WithTx.rollback",
			},
		},
		{
			name: "appointment binding conflict (different pet) is rejected",
			mutate: func(res *resRepoFake) {
				res.appt.PetID = u64(424242) // appointment is for a different pet
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.BackfillForMedicalRecord", "res.FindByID", "tx.WithTx.rollback",
			},
		},
		{
			name: "deceased pet write is rejected after context resolution",
			mutate: func(res *resRepoFake) {
				res.lifeState = sharedkernel.PetLifeStateDeceased
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.BackfillForMedicalRecord", "res.FindByID",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.ReadPetLifeState", "tx.WithTx.rollback",
			},
		},
		{
			name: "unresolved pet life state is rejected, never treated as alive",
			mutate: func(res *resRepoFake) {
				res.lifeState = ""
				res.lifeErr = apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.BackfillForMedicalRecord", "res.FindByID",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.ReadPetLifeState", "tx.WithTx.rollback",
			},
		},
		{
			name: "pet life state reader failure is rejected, never treated as alive",
			mutate: func(res *resRepoFake) {
				res.lifeState = ""
				res.lifeErr = errors.New("SECCHK20261009READERDOWN")
			},
			wantSeq: []string{
				"tx.WithTx.begin", "repo.AcquireAutoCreateLock", "repo.CountByPetAndDate",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.BackfillForMedicalRecord", "res.FindByID",
				"res.AssertOwnerInClinic", "res.FindPetOwnerInClinic",
				"res.ReadPetLifeState", "tx.WithTx.rollback",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seq := newSeq()
			repo := &mrRepoFake{seq: seq, acquired: true, count: 0, createID: 6001}
			res := &resRepoFake{seq: seq, petOwnerID: 70, appt: matchingAppointment(reservation), lifeState: aliveLifeState}
			tc.mutate(res)
			plan := &planRepoFake{seq: seq}
			audit := &auditFake{seq: seq}
			tx := &txFake{seq: seq, runFn: true}

			svc := newService(repo, res, plan, audit, nil, tx)
			assert.NotPanics(t, func() {
				svc.AutoCreateFromReservation(context.Background(), 7, reservation)
			})

			assert.Equal(t, tc.wantSeq, *seq)
			assert.Equal(t, 0, repo.createCalls, "refused path must not write a medical record")
			assert.Equal(t, 0, audit.changeCalls+audit.entryCalls, "refused path must not write audit rows")
		})
	}
}

// config_JST returns the JST location used by the service for date keys.
func config_JST() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}

// date-key contract: the advisory lock and same-day count use the JST calendar
// day of reservation.StartTime — a 15:30 UTC reservation is 2026-10-10 in JST.
func TestC15_AutoCreateDateKeyIsJSTDerived(t *testing.T) {
	seq := newSeq()
	reservation := validReservation() // 2026-10-09T15:30Z => JST 2026-10-10
	repo := &mrRepoFake{seq: seq, acquired: false}
	res := &resRepoFake{seq: seq}
	tx := &txFake{seq: seq, runFn: true}

	var lockDate string
	capture := &dateCapturingRepo{mrRepoFake: repo, onAcquire: func(d string) { lockDate = d }}

	svc := medicalrecord.NewMedicalRecordServiceWithTxAudit(
		capture, nil, &planRepoFake{seq: seq}, nil, nil, nil, nil, res, nil, &auditFake{seq: seq}, nil, tx,
	)
	svc.AutoCreateFromReservation(context.Background(), 7, reservation)
	assert.Equal(t, "2026-10-10", lockDate)
}

type dateCapturingRepo struct {
	*mrRepoFake
	onAcquire func(date string)
}

func (r *dateCapturingRepo) AcquireAutoCreateLock(ctx context.Context, clinicID, petID uint64, date string) (bool, error) {
	r.onAcquire(date)
	return r.mrRepoFake.AcquireAutoCreateLock(ctx, clinicID, petID, date)
}
