package medicalrecord

// EMR-447 A01: procedure の is_bill_discount_excluded 契約（consultation と同型）。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestProcedureBillDiscount_RequestBinding(t *testing.T) {
	var req createProcedureRequest
	if err := json.Unmarshal([]byte(`{"name":"a","tax_type":"excluded"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.IsBillDiscountExcluded.toCreateInput() != nil {
		t.Fatal("omitted must stay nil")
	}
	if err := json.Unmarshal([]byte(`{"name":"a","tax_type":"excluded","is_bill_discount_excluded":null}`), &req); !apperrors.IsInvalidInput(err) {
		t.Fatalf("null: %v", err)
	}
	var req2 updateProcedureRequest
	if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":false}`), &req2); err != nil {
		t.Fatal(err)
	}
	if v := req2.IsBillDiscountExcluded.toUpdateInput(); v == nil || *v {
		t.Fatalf("false lost: %v", v)
	}
	if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":"yes"}`), &req2); !apperrors.IsInvalidInput(err) {
		t.Fatalf("non-boolean: %v", err)
	}
}

func TestProcedureBillDiscount_CreateAndPatch(t *testing.T) {
	t.Run("create omitted saves explicit false and audits", func(t *testing.T) {
		var stored *model.Procedure
		audit := &billAuditRecorder{}
		repo := &mockProcedureRepository{createFn: func(_ context.Context, p *model.Procedure) error {
			p.ID = 5
			stored = p
			return nil
		}}
		_, err := NewProcedureServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
			Create(t.Context(), 7, &CreateProcedureInput{ActorID: masterTestActorID(), Name: "item", Anesthesia: "none", IsBillDiscountExcluded: nil})
		if err != nil {
			t.Fatal(err)
		}
		if stored == nil || stored.IsBillDiscountExcluded == nil || *stored.IsBillDiscountExcluded {
			t.Fatalf("stored=%+v want false", stored)
		}
		found := false
		for _, e := range audit.entries {
			if billAuditField(e) == colIsBillDiscountExcluded {
				found = true
				if e.NewValue.(map[string]any)[colIsBillDiscountExcluded] != false || e.OldValue != nil {
					t.Fatalf("wrong audit entry: %+v", e)
				}
			}
		}
		if !found {
			t.Fatal("create must audit is_bill_discount_excluded")
		}
	})
	t.Run("patch omitted preserves stored NULL/false/true without audit", func(t *testing.T) {
		for _, stored := range []*bool{nil, boolPtr(false), boolPtr(true)} {
			fields := buildProcedureUpdate(&UpdateProcedureInput{Version: 1, ActorID: masterTestActorID()})
			if _, ok := fields[colIsBillDiscountExcluded]; ok {
				t.Fatal("omitted must not enter the update map")
			}
			audit := &billAuditRecorder{}
			repo := &mockProcedureRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
					return &model.Procedure{ID: 5, ClinicID: 7, IsBillDiscountExcluded: stored}, nil
				},
				updateFieldsFn: func(context.Context, uint64, uint64, UpdateProcedureInput) (*model.Procedure, error) {
					return &model.Procedure{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: stored}, nil
				},
			}
			name := "changed"
			row, err := NewProcedureServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
				Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: 7, ActorID: masterTestActorID(), Name: &name})
			if err != nil {
				t.Fatal(err)
			}
			if row.IsBillDiscountExcluded != stored {
				t.Fatalf("stored %v rewritten to %v", stored, row.IsBillDiscountExcluded)
			}
			for _, e := range audit.entries {
				if billAuditField(e) == colIsBillDiscountExcluded {
					t.Fatal("patch without the field must not audit it")
				}
			}
		}
	})
	t.Run("patch true-to-false and false-to-true audited with before/after", func(t *testing.T) {
		for _, tc := range []struct {
			stored, input, wantBefore, wantAfter bool
		}{{true, false, true, false}, {false, true, false, true}} {
			audit := &billAuditRecorder{}
			repo := &mockProcedureRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
					return &model.Procedure{ID: 5, ClinicID: 7, IsBillDiscountExcluded: boolPtr(tc.stored)}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateProcedureInput) (*model.Procedure, error) {
					return &model.Procedure{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
				},
			}
			row, err := NewProcedureServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
				Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: 7, ActorID: masterTestActorID(), IsBillDiscountExcluded: boolPtr(tc.input)})
			if err != nil {
				t.Fatal(err)
			}
			if row.IsBillDiscountExcluded == nil || *row.IsBillDiscountExcluded != tc.wantAfter {
				t.Fatalf("response lost value: %+v", row)
			}
			var bill *AuditEntry
			for _, e := range audit.entries {
				if billAuditField(e) == colIsBillDiscountExcluded {
					bill = e
				}
			}
			if bill == nil {
				t.Fatal("patch with the field must audit it")
			}
			if got := bill.OldValue.(map[string]any)[colIsBillDiscountExcluded]; got != tc.wantBefore {
				t.Fatalf("before %v want %v", got, tc.wantBefore)
			}
			if got := bill.NewValue.(map[string]any)[colIsBillDiscountExcluded]; got != tc.wantAfter {
				t.Fatalf("after %v want %v", got, tc.wantAfter)
			}
		}
	})
}

func TestProcedureBillDiscount_FailClosedResponseAndReader(t *testing.T) {
	t.Run("constructor without dependencies rejects the field before mutation", func(t *testing.T) {
		called := false
		repo := &mockProcedureRepository{createFn: func(context.Context, *model.Procedure) error { called = true; return nil }}
		_, err := NewProcedureService(repo, masterTestTransactor{}).Create(t.Context(), 7, &CreateProcedureInput{ActorID: masterTestActorID(), Name: "item", Anesthesia: "none", IsBillDiscountExcluded: boolPtr(true)})
		if err == nil || called {
			t.Fatal("missing audit dependency must fail closed")
		}
	})
	t.Run("audit failure rolls back the whole write", func(t *testing.T) {
		rolledBack := false
		tx := masterServiceJournalTx{rollback: func() { rolledBack = true }}
		audit := &billAuditRecorder{err: errors.New("audit unavailable")}
		repo := &mockProcedureRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
				return &model.Procedure{ID: 5, ClinicID: 7}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateProcedureInput) (*model.Procedure, error) {
				return &model.Procedure{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
			},
		}
		row, err := NewProcedureServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}).
			Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: 7, ActorID: masterTestActorID(), IsBillDiscountExcluded: boolPtr(false)})
		if err == nil || row != nil || !rolledBack {
			t.Fatalf("audit failure must roll back: %v %v", row, err)
		}
	})
	t.Run("response preserves null", func(t *testing.T) {
		raw, err := json.Marshal(toProcedureResponse(&model.Procedure{Version: 17}))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"is_bill_discount_excluded":null`)) {
			t.Fatalf("lost null: %s", raw)
		}
	})
	t.Run("reader error is not converted to null", func(t *testing.T) {
		repo := &mockProcedureRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
			return nil, errors.New("column does not exist")
		}}
		if _, err := NewProcedureServiceWithAudit(repo, masterTestWriteDependencies()).GetByID(t.Context(), 7, 5); err == nil {
			t.Fatal("reader failure must surface as an error")
		}
	})
}
