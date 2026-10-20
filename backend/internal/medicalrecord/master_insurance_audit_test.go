package medicalrecord

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type insuranceAuditContextKey struct{}
type insuranceAuditTx struct {
	value *bool
	calls int
}

func (tx *insuranceAuditTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx.calls++
	previous := *tx.value
	if err := fn(context.WithValue(ctx, insuranceAuditContextKey{}, true)); err != nil {
		*tx.value = previous
		return err
	}
	return nil
}

type insuranceAuditRecorder struct {
	entry *AuditEntry
	err   error
	inTx  bool
}

func (a *insuranceAuditRecorder) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	a.entry = entry
	a.inTx, _ = ctx.Value(insuranceAuditContextKey{}).(bool)
	return a.err
}
func TestMasterInsuranceAudit_AtomicActorAndNullableBefore(t *testing.T) {
	for _, auditFails := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "audit rollback"}[auditFails], func(t *testing.T) {
			setting := false
			tx := &insuranceAuditTx{value: &setting}
			audit := &insuranceAuditRecorder{}
			if auditFails {
				audit.err = errors.New("audit unavailable")
			}
			actor := uint64(12)
			err := (MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}).write(t.Context(), 7, &actor, "consultation", "consultation.update", func(ctx context.Context) (masterInsuranceChange, error) {
				if active, _ := ctx.Value(insuranceAuditContextKey{}).(bool); !active {
					t.Fatal("write escaped TX")
				}
				setting = true
				return masterInsuranceChange{ID: 5, BeforeExists: true, Before: nil, After: &setting}, nil
			})
			if (err != nil) != auditFails {
				t.Fatalf("error=%v", err)
			}
			if setting == auditFails {
				t.Fatalf("setting=%v, audit failure=%v", setting, auditFails)
			}
			if !audit.inTx || tx.calls != 1 {
				t.Fatal("audit and write did not share transaction")
			}
			entry := audit.entry
			if entry == nil || *entry.ActorID != 12 || *entry.ClinicID != 7 || *entry.ResourceID != 5 {
				t.Fatal("wrong audit identity")
			}
			if !reflect.DeepEqual(entry.OldValue, map[string]any{"is_non_insurance": nil}) || !reflect.DeepEqual(entry.NewValue, map[string]any{"is_non_insurance": true}) {
				t.Fatalf("lost setting meaning: %#v", entry)
			}
		})
	}
}
func TestMasterInsuranceAudit_RejectsMissingDependencyOrActorBeforeWrite(t *testing.T) {
	setting := false
	tx := &insuranceAuditTx{value: &setting}
	audit := &insuranceAuditRecorder{}
	actor := uint64(12)
	for _, tc := range []struct {
		deps  MasterInsuranceWriteDependencies
		actor *uint64
	}{
		{MasterInsuranceWriteDependencies{}, &actor},
		{MasterInsuranceWriteDependencies{Transactor: tx}, &actor},
		{MasterInsuranceWriteDependencies{Audit: audit}, &actor},
		{MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}, nil},
	} {
		called := false
		err := tc.deps.write(t.Context(), 7, tc.actor, "procedure", "procedure.create", func(context.Context) (masterInsuranceChange, error) {
			called = true
			return masterInsuranceChange{}, nil
		})
		if err == nil || called {
			t.Fatal("invalid dependencies reached mutation")
		}
	}
}
