package clinic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// --- test doubles for UpdateStandard integrity ---

type passthroughTransactor struct{}

func (passthroughTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockClinicRowLocker struct {
	lockFn func(ctx context.Context, id uint64) (*model.Clinic, error)
	calls  int
}

func (m *mockClinicRowLocker) LockByIDForUpdate(ctx context.Context, id uint64) (*model.Clinic, error) {
	m.calls++
	if m.lockFn != nil {
		return m.lockFn(ctx, id)
	}
	return &model.Clinic{ID: id}, nil
}

type recordingAuditTxLogger struct {
	mu      sync.Mutex
	entries []*AuditEntry
	err     error
}

func (r *recordingAuditTxLogger) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry != nil {
		cp := *entry
		r.entries = append(r.entries, &cp)
	}
	return r.err
}

func (r *recordingAuditTxLogger) snapshot() []*AuditEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*AuditEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

func integrityDeps(locker ClinicRowLocker, audit AuditTxLogger) *ClosingSettingsServiceDeps {
	return &ClosingSettingsServiceDeps{
		Transactor:   passthroughTransactor{},
		ClinicLocker: locker,
		AuditTx:      audit,
	}
}

// TestClosingSettingsService_UpdateStandard_Validation rejects invalid time format,
// reversed time order after partial merge, and invalid closed_weekdays before persistence.
func TestClosingSettingsService_UpdateStandard_Validation(t *testing.T) {
	ctx := context.Background()
	const clinicID uint64 = 1
	const actorID uint64 = 99

	base := &model.ClinicSettings{
		ClinicID:            clinicID,
		ClosingAmPmBoundary: "14:00",
		ClosingWeekdayEnd:   "18:30",
		ClosingSundayEnd:    "17:30",
	}

	cases := []struct {
		name    string
		current *model.ClinicSettings
		input   UpdateClinicSettingsInput
		wantMsg string
	}{
		{
			name:    "invalid time format on boundary",
			current: base,
			input: UpdateClinicSettingsInput{
				ClosingAmPmBoundary: strPtr("not-a-time"),
			},
			wantMsg: "境界時刻",
		},
		{
			name:    "time order reversed: boundary after weekday end (partial final state)",
			current: base,
			input: UpdateClinicSettingsInput{
				// final: boundary 19:00, weekday_end 18:30 → invalid
				ClosingAmPmBoundary: strPtr("19:00"),
			},
			wantMsg: "より後",
		},
		{
			name:    "partial update leaves sunday_end before new boundary",
			current: base,
			input: UpdateClinicSettingsInput{
				// final: boundary 18:00, sunday_end 17:30 → invalid
				ClosingAmPmBoundary: strPtr("18:00"),
			},
			wantMsg: "より後",
		},
		{
			name:    "closed_weekdays out of range",
			current: base,
			input: UpdateClinicSettingsInput{
				ClosedWeekdays: []int64{7},
			},
			wantMsg: "closed_weekdays",
		},
		{
			name:    "closed_weekdays duplicate",
			current: base,
			input: UpdateClinicSettingsInput{
				ClosedWeekdays: []int64{1, 1},
			},
			wantMsg: "closed_weekdays",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var saved bool
			settingsRepo := &mockClinicSettingsRepository{
				findByClinicIDFn: func(_ context.Context, id uint64) (*model.ClinicSettings, error) {
					cp := *tc.current
					cp.ClinicID = id
					return &cp, nil
				},
				saveFn: func(_ context.Context, _ uint64, _ *model.ClinicSettings) (*model.ClinicSettings, error) {
					saved = true
					return nil, errors.New("save must not be called on validation failure")
				},
			}
			audit := &recordingAuditTxLogger{}
			svc := NewClosingSettingsService(settingsRepo, nil, nil, integrityDeps(&mockClinicRowLocker{}, audit))
			_, err := svc.UpdateStandard(ctx, clinicID, actorID, tc.input)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err) || containsAny(err.Error(), tc.wantMsg),
				"err=%v wantMsg=%q", err, tc.wantMsg)
			assert.False(t, saved, "must fail-fast before persistence")
			assert.Empty(t, audit.snapshot(), "must not audit rejected updates")
		})
	}
}

// TestClosingSettingsService_UpdateStandard_ConcurrentSameClinic_NoLostUpdate proves
// concurrent partial PATCHes on one clinic both land (real DB + clinics row FOR UPDATE).
// Repeats under contention so a lock-free RMW is very likely to lose a field.
func TestClosingSettingsService_UpdateStandard_ConcurrentSameClinic_NoLostUpdate(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	clinic := makeClinicFixture(t, db, "concurrent-same-clinic")
	repo := NewClinicSettingsRepository(db)
	clinicRepo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)

	const rounds = 15
	for round := 0; round < rounds; round++ {
		_, err := repo.Save(context.Background(), clinic.ID, &model.ClinicSettings{
			ClinicID:            clinic.ID,
			ClosingAmPmBoundary: "14:00",
			ClosingWeekdayEnd:   "18:30",
			ClosingSundayEnd:    "17:30",
		})
		require.NoError(t, err)

		svc := NewClosingSettingsService(repo, nil, nil, &ClosingSettingsServiceDeps{
			Transactor:   transactor,
			ClinicLocker: clinicRepo,
			AuditTx:      &recordingAuditTxLogger{},
		})

		weekdayEnd := "20:00"
		sundayEnd := "16:00"
		var wg sync.WaitGroup
		errCh := make(chan error, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, e := svc.UpdateStandard(context.Background(), clinic.ID, 11, UpdateClinicSettingsInput{
				ClosingWeekdayEnd: &weekdayEnd,
			})
			errCh <- e
		}()
		go func() {
			defer wg.Done()
			_, e := svc.UpdateStandard(context.Background(), clinic.ID, 12, UpdateClinicSettingsInput{
				ClosingSundayEnd: &sundayEnd,
			})
			errCh <- e
		}()
		wg.Wait()
		close(errCh)
		for e := range errCh {
			require.NoError(t, e)
		}

		final, err := repo.FindByClinicID(context.Background(), clinic.ID)
		require.NoError(t, err)
		// Both partial updates must be visible — lost update would drop one of them.
		assert.Equal(t, "20:00", normalizeHHMM(final.ClosingWeekdayEnd), "closing_weekday_end from goroutine A (round %d)", round)
		assert.Equal(t, "16:00", normalizeHHMM(final.ClosingSundayEnd), "closing_sunday_end from goroutine B (round %d)", round)
		assert.Equal(t, "14:00", normalizeHHMM(final.ClosingAmPmBoundary), "untouched boundary preserved (round %d)", round)
	}
}

// TestClosingSettingsService_UpdateStandard_ConcurrentOtherClinic_NoContention proves
// concurrent PATCHes on different clinics both succeed without blocking each other.
func TestClosingSettingsService_UpdateStandard_ConcurrentOtherClinic_NoContention(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	clinicA := makeClinicFixture(t, db, "concurrent-clinic-a")
	clinicB := makeClinicFixture(t, db, "concurrent-clinic-b")
	repo := NewClinicSettingsRepository(db)
	clinicRepo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)

	svc := NewClosingSettingsService(repo, nil, nil, &ClosingSettingsServiceDeps{
		Transactor:   transactor,
		ClinicLocker: clinicRepo,
		AuditTx:      &recordingAuditTxLogger{},
	})

	endA := "19:00"
	endB := "15:30"
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := svc.UpdateStandard(context.Background(), clinicA.ID, 21, UpdateClinicSettingsInput{
			ClosingWeekdayEnd: &endA,
		})
		errCh <- e
	}()
	go func() {
		defer wg.Done()
		_, e := svc.UpdateStandard(context.Background(), clinicB.ID, 22, UpdateClinicSettingsInput{
			ClosingWeekdayEnd: &endB,
		})
		errCh <- e
	}()
	wg.Wait()
	close(errCh)
	for e := range errCh {
		require.NoError(t, e, "cross-clinic concurrent updates must both succeed")
	}

	gotA, err := repo.FindByClinicID(context.Background(), clinicA.ID)
	require.NoError(t, err)
	gotB, err := repo.FindByClinicID(context.Background(), clinicB.ID)
	require.NoError(t, err)
	assert.Equal(t, "19:00", normalizeHHMM(gotA.ClosingWeekdayEnd))
	assert.Equal(t, "15:30", normalizeHHMM(gotB.ClosingWeekdayEnd))
}

// TestClosingSettingsService_UpdateStandard_Audit records clinic_id, actor, and
// non-sensitive before/after metadata in the same successful update path.
func TestClosingSettingsService_UpdateStandard_Audit(t *testing.T) {
	ctx := context.Background()
	const clinicID uint64 = 7
	const actorID uint64 = 42

	current := &model.ClinicSettings{
		ClinicID:            clinicID,
		ClosingAmPmBoundary: "14:00",
		ClosingWeekdayEnd:   "18:30",
		ClosingSundayEnd:    "17:30",
	}
	settingsRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, id uint64) (*model.ClinicSettings, error) {
			cp := *current
			cp.ClinicID = id
			return &cp, nil
		},
		saveFn: func(_ context.Context, id uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
			s.ClinicID = id
			return s, nil
		},
	}
	audit := &recordingAuditTxLogger{}
	locker := &mockClinicRowLocker{}
	svc := NewClosingSettingsService(settingsRepo, nil, nil, integrityDeps(locker, audit))

	weekdayEnd := "19:45"
	res, err := svc.UpdateStandard(ctx, clinicID, actorID, UpdateClinicSettingsInput{
		ClosingWeekdayEnd: &weekdayEnd,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, locker.calls, "must lock clinic row")

	entries := audit.snapshot()
	require.Len(t, entries, 1)
	e := entries[0]
	require.NotNil(t, e.ClinicID)
	assert.Equal(t, clinicID, *e.ClinicID)
	require.NotNil(t, e.ActorID)
	assert.Equal(t, actorID, *e.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, e.ActorType)
	assert.Equal(t, auditActionClosingSettingsUpdateStandard, e.Action)
	assert.Equal(t, auditResourceClinicSettings, e.Resource)

	// Non-sensitive metadata: field change markers only — never raw closing time values.
	oldMap, ok := e.OldValue.(map[string]any)
	require.True(t, ok, "OldValue must be map metadata")
	newMap, ok := e.NewValue.(map[string]any)
	require.True(t, ok, "NewValue must be map metadata")
	assert.NotContains(t, stringifyAny(oldMap), "18:30")
	assert.NotContains(t, stringifyAny(oldMap), "14:00")
	assert.NotContains(t, stringifyAny(newMap), "19:45")
	assert.NotContains(t, stringifyAny(newMap), "18:30")

	changed, ok := newMap["changed_fields"].([]string)
	if !ok {
		// JSON-like any slice
		raw, ok2 := newMap["changed_fields"].([]any)
		require.True(t, ok2, "changed_fields required in NewValue")
		changed = make([]string, 0, len(raw))
		for _, v := range raw {
			changed = append(changed, v.(string))
		}
	}
	assert.Contains(t, changed, "closing_weekday_end")
	assert.Contains(t, oldMap, "fields")
	assert.Contains(t, newMap, "fields")
}

// TestClosingSettingsService_UpdateStandard_Rollback rolls back settings when audit fails
// (real DB transaction).
func TestClosingSettingsService_UpdateStandard_Rollback(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	clinic := makeClinicFixture(t, db, "audit-rollback-clinic")
	repo := NewClinicSettingsRepository(db)
	clinicRepo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)

	seed := &model.ClinicSettings{
		ClinicID:            clinic.ID,
		ClosingAmPmBoundary: "14:00",
		ClosingWeekdayEnd:   "18:30",
		ClosingSundayEnd:    "17:30",
	}
	_, err := repo.Save(context.Background(), clinic.ID, seed)
	require.NoError(t, err)

	failingAudit := &recordingAuditTxLogger{err: errors.New("audit write failed")}
	svc := NewClosingSettingsService(repo, nil, nil, &ClosingSettingsServiceDeps{
		Transactor:   transactor,
		ClinicLocker: clinicRepo,
		AuditTx:      failingAudit,
	})

	newEnd := "21:00"
	_, err = svc.UpdateStandard(context.Background(), clinic.ID, 55, UpdateClinicSettingsInput{
		ClosingWeekdayEnd: &newEnd,
	})
	require.Error(t, err)

	// Re-read: must still be pre-update value (transaction rolled back).
	got, err := repo.FindByClinicID(context.Background(), clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, "18:30", normalizeHHMM(got.ClosingWeekdayEnd),
		"settings must remain unchanged when audit fails")
	assert.NotEqual(t, "21:00", normalizeHHMM(got.ClosingWeekdayEnd))
}

func strPtr(s string) *string { return &s }

func containsAny(s, sub string) bool {
	return strings.Contains(s, sub)
}

func stringifyAny(v any) string {
	return fmt.Sprintf("%v", v)
}

func normalizeHHMM(s string) string {
	// GORM/pg may return "18:30:00"; compare on HH:MM.
	if len(s) >= 5 {
		return s[:5]
	}
	return s
}
