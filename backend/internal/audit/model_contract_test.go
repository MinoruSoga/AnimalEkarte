package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

type modelContractRepository struct {
	saved *model.AuditLog
	tx    bool
	err   error
}

func (r *modelContractRepository) Create(_ context.Context, value *model.AuditLog) error {
	r.saved = value
	return r.err
}
func (r *modelContractRepository) CreateTx(ctx context.Context, value *model.AuditLog) error {
	r.tx = true
	return r.Create(ctx, value)
}
func TestAuditModelContract_TxClinicAndNullableInsuranceJSON(t *testing.T) {
	clinic, actor := uint64(7), uint64(12)
	for _, setting := range []any{nil, false, true} {
		repo := &modelContractRepository{}
		err := NewService(repo).LogEntryTx(t.Context(), &Entry{ClinicID: &clinic, ActorID: &actor, ActorType: model.AuditActorTypeStaff, Action: "consultation.update", Resource: "consultation", OldValue: map[string]any{"is_non_insurance": setting}, NewValue: map[string]any{"is_non_insurance": true}})
		if err != nil {
			t.Fatal(err)
		}
		if !repo.tx || repo.saved.ClinicID != clinic || repo.saved.ActorID == nil || *repo.saved.ActorID != actor {
			t.Fatal("audit identity/TX lost")
		}
		var before map[string]any
		if repo.saved.OldValue == nil {
			t.Fatal("missing prior setting")
		}
		if err := json.Unmarshal(*repo.saved.OldValue, &before); err != nil {
			t.Fatal(err)
		}
		if got, exists := before["is_non_insurance"]; !exists || got != setting {
			t.Fatalf("setting %#v became %#v", setting, before)
		}
	}
}
func TestAuditModelContract_RequiredClinicAndWriterFailure(t *testing.T) {
	actor := uint64(12)
	repo := &modelContractRepository{}
	if err := NewService(repo).LogEntryTx(t.Context(), &Entry{ActorID: &actor, ActorType: model.AuditActorTypeStaff, Action: "procedure.create", Resource: "procedure"}); err == nil || repo.saved != nil {
		t.Fatal("nil clinic reached storage")
	}
	clinic := uint64(7)
	repo.err = errors.New("audit unavailable")
	if err := NewService(repo).LogEntryTx(t.Context(), &Entry{ClinicID: &clinic, ActorID: &actor, ActorType: model.AuditActorTypeStaff, Action: "procedure.create", Resource: "procedure"}); err == nil {
		t.Fatal("audit failure was swallowed")
	}
}
