package medicalrecord

// EMR-447 A01: vaccine の is_bill_discount_excluded 契約（consultation/procedure と同型）。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestVaccineBillDiscount_RequestBinding(t *testing.T) {
	var req createVaccineRequest
	if err := json.Unmarshal([]byte(`{"name":"a"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.IsBillDiscountExcluded.toCreateInput() != nil {
		t.Fatal("omitted must stay nil")
	}
	if err := json.Unmarshal([]byte(`{"name":"a","is_bill_discount_excluded":null}`), &req); !apperrors.IsInvalidInput(err) {
		t.Fatalf("null: %v", err)
	}
	var req2 updateVaccineRequest
	if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":true}`), &req2); err != nil {
		t.Fatal(err)
	}
	if v := req2.IsBillDiscountExcluded.toUpdateInput(); v == nil || !*v {
		t.Fatalf("true lost: %v", v)
	}
	if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":1}`), &req2); !apperrors.IsInvalidInput(err) {
		t.Fatalf("non-boolean: %v", err)
	}
}

func TestVaccineBillDiscount_CreateAndPatch(t *testing.T) {
	t.Run("create omitted saves explicit false and audits", func(t *testing.T) {
		var stored *model.Vaccine
		audit := &billAuditRecorder{}
		repo := &mockVaccineRepository{createFn: func(_ context.Context, v *model.Vaccine) error {
			v.ID = 17
			stored = v
			return nil
		}}
		_, err := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
			Create(t.Context(), 2, &CreateVaccineInput{ActorID: vaccineTestActorID(), Name: "item"})
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
			fields := buildVaccineUpdate(&UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID()})
			if _, ok := fields[colIsBillDiscountExcluded]; ok {
				t.Fatal("omitted must not enter the update map")
			}
			audit := &billAuditRecorder{}
			repo := &mockVaccineRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
					return &model.Vaccine{ID: 17, ClinicID: 2, IsBillDiscountExcluded: stored}, nil
				},
				updateFieldsFn: func(context.Context, uint64, uint64, UpdateVaccineInput) (*model.Vaccine, error) {
					return &model.Vaccine{ID: 17, ClinicID: 2, Version: 8, IsBillDiscountExcluded: stored}, nil
				},
			}
			name := "changed"
			row, err := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
				Update(t.Context(), 2, 17, &UpdateVaccineInput{Version: 7, ActorID: vaccineTestActorID(), Name: &name})
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
			repo := &mockVaccineRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
					return &model.Vaccine{ID: 17, ClinicID: 2, IsBillDiscountExcluded: boolPtr(tc.stored)}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateVaccineInput) (*model.Vaccine, error) {
					return &model.Vaccine{ID: 17, ClinicID: 2, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
				},
			}
			row, err := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
				Update(t.Context(), 2, 17, &UpdateVaccineInput{Version: 7, ActorID: vaccineTestActorID(), IsBillDiscountExcluded: boolPtr(tc.input)})
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

func TestVaccineBillDiscount_FailClosedResponseAndReader(t *testing.T) {
	t.Run("constructor without dependencies rejects the field before mutation", func(t *testing.T) {
		called := false
		repo := &mockVaccineRepository{createFn: func(context.Context, *model.Vaccine) error { called = true; return nil }}
		_, err := NewVaccineService(repo).Create(t.Context(), 2, &CreateVaccineInput{ActorID: vaccineTestActorID(), Name: "item", IsBillDiscountExcluded: boolPtr(true)})
		if err == nil || called {
			t.Fatal("missing audit dependency must fail closed")
		}
	})
	t.Run("audit failure rolls back the whole write", func(t *testing.T) {
		rolledBack := false
		tx := masterServiceJournalTx{rollback: func() { rolledBack = true }}
		audit := &billAuditRecorder{err: errors.New("audit unavailable")}
		repo := &mockVaccineRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
				return &model.Vaccine{ID: 17, ClinicID: 2}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateVaccineInput) (*model.Vaccine, error) {
				return &model.Vaccine{ID: 17, ClinicID: 2, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
			},
		}
		row, err := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}).
			Update(t.Context(), 2, 17, &UpdateVaccineInput{Version: 7, ActorID: vaccineTestActorID(), IsBillDiscountExcluded: boolPtr(false)})
		if err == nil || row != nil || !rolledBack {
			t.Fatalf("audit failure must roll back: %v %v", row, err)
		}
	})
	t.Run("response preserves null", func(t *testing.T) {
		raw, err := json.Marshal(toVaccineResponse(&model.Vaccine{Version: 17}))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"is_bill_discount_excluded":null`)) {
			t.Fatalf("lost null: %s", raw)
		}
	})
	t.Run("reader error is not converted to null", func(t *testing.T) {
		repo := &mockVaccineRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
			return nil, errors.New("column does not exist")
		}}
		if _, err := NewVaccineServiceWithAudit(repo, masterTestWriteDependencies()).GetByID(t.Context(), 2, 17); err == nil {
			t.Fatal("reader failure must surface as an error")
		}
	})
}
