package medicalrecord

// EMR-447 A01: checkup_type の is_bill_discount_excluded 契約。
// checkup_type は専有 transactor/auditTx を使うため、bill-discount 監査も
// logBillDiscountAuditTx で同一 tx に載せる（保険監査は変更しない）。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCheckupTypeBillDiscount_RequestBinding(t *testing.T) {
	var req createCheckupTypeRequest
	if err := json.Unmarshal([]byte(`{"name":"a"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.IsBillDiscountExcluded.toCreateInput() != nil {
		t.Fatal("omitted must stay nil")
	}
	if err := json.Unmarshal([]byte(`{"name":"a","is_bill_discount_excluded":null}`), &req); !apperrors.IsInvalidInput(err) {
		t.Fatalf("null: %v", err)
	}
	var req2 updateCheckupTypeRequest
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

func TestCheckupTypeBillDiscount_CreateSavesExplicitValueAndAudits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    *bool
		wantSave bool
	}{
		{name: "omitted saves explicit false", input: nil, wantSave: false},
		{name: "explicit true preserved", input: boolPtr(true), wantSave: true},
		{name: "explicit false preserved", input: boolPtr(false), wantSave: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stored *model.CheckupType
			audit := &recordingCheckupTypeAudit{}
			repo := &mockCheckupTypeRepository{createFn: func(_ context.Context, ct *model.CheckupType) error {
				ct.ID = 5
				ct.Version = 1
				stored = ct
				return nil
			}}
			svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, audit)
			got, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "定期健診", IsBillDiscountExcluded: tc.input, ActorID: &testCheckupTypeActorID})
			if err != nil {
				t.Fatal(err)
			}
			if stored == nil || stored.IsBillDiscountExcluded == nil || *stored.IsBillDiscountExcluded != tc.wantSave {
				t.Fatalf("stored=%+v want %v", stored, tc.wantSave)
			}
			if got.IsBillDiscountExcluded == nil || *got.IsBillDiscountExcluded != tc.wantSave {
				t.Fatalf("response lost value: %+v", got)
			}
			var bill *AuditEntry
			for _, e := range audit.entries {
				if billAuditField(e) == colIsBillDiscountExcluded {
					bill = e
				}
			}
			if bill == nil {
				t.Fatal("create must audit is_bill_discount_excluded")
			}
			if bill.OldValue != nil || bill.NewValue.(map[string]any)[colIsBillDiscountExcluded] != tc.wantSave {
				t.Fatalf("wrong audit entry: %+v", bill)
			}
			if bill.Resource != checkupTypeAuditResource || bill.Action != checkupTypeAuditActionCreate {
				t.Fatalf("wrong audit identity: %+v", bill)
			}
		})
	}
}

func TestCheckupTypeBillDiscount_PatchPreservesAndUpdates(t *testing.T) {
	t.Run("omitted writes nothing and audits nothing", func(t *testing.T) {
		fields := buildCheckupTypeUpdate(&UpdateCheckupTypeInput{Version: int64Ptr(3)})
		if _, ok := fields[colIsBillDiscountExcluded]; ok {
			t.Fatal("omitted must not enter the update map")
		}
	})
	for _, tc := range []struct {
		name       string
		stored     *bool
		input      *bool
		wantBefore any
		wantAfter  any
	}{
		{name: "NULL to false audited", stored: nil, input: boolPtr(false), wantBefore: nil, wantAfter: false},
		{name: "true to false audited", stored: boolPtr(true), input: boolPtr(false), wantBefore: true, wantAfter: false},
		{name: "false to true audited", stored: boolPtr(false), input: boolPtr(true), wantBefore: false, wantAfter: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stored := &model.CheckupType{ID: 9, ClinicID: 1, IsBillDiscountExcluded: tc.stored, Version: 3}
			audit := &recordingCheckupTypeAudit{}
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					cp := *stored
					return &cp, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error) {
					after := *stored
					if v, ok := buildCheckupTypeUpdate(&cmd)[colIsBillDiscountExcluded]; ok {
						b := v.(bool)
						after.IsBillDiscountExcluded = &b
					}
					after.Version = 4
					return &after, nil
				},
			}
			svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, audit)
			got, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{IsBillDiscountExcluded: tc.input, Version: int64Ptr(3), ActorID: &testCheckupTypeActorID})
			if err != nil {
				t.Fatal(err)
			}
			if got.IsBillDiscountExcluded == nil || *got.IsBillDiscountExcluded != tc.wantAfter {
				t.Fatalf("response lost value: %+v", got)
			}
			var bill *AuditEntry
			for _, e := range audit.entries {
				if billAuditField(e) == colIsBillDiscountExcluded {
					bill = e
				}
			}
			if bill == nil || bill.Action != checkupTypeAuditActionUpdate {
				t.Fatalf("missing bill audit: %+v", audit.entries)
			}
			if got := bill.OldValue.(map[string]any)[colIsBillDiscountExcluded]; got != tc.wantBefore {
				t.Fatalf("before %v want %v", got, tc.wantBefore)
			}
			if got := bill.NewValue.(map[string]any)[colIsBillDiscountExcluded]; got != tc.wantAfter {
				t.Fatalf("after %v want %v", got, tc.wantAfter)
			}
		})
	}
	t.Run("name-only patch emits no bill discount audit", func(t *testing.T) {
		name := "改名"
		audit := &recordingCheckupTypeAudit{}
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 3}, nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateCheckupTypeInput) (*model.CheckupType, error) {
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 4}, nil
			},
		}
		svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, audit)
		if _, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{Name: &name, Version: int64Ptr(3), ActorID: &testCheckupTypeActorID}); err != nil {
			t.Fatal(err)
		}
		for _, e := range audit.entries {
			if billAuditField(e) == colIsBillDiscountExcluded {
				t.Fatal("name-only patch must not audit bill discount")
			}
		}
	})
}

func TestCheckupTypeBillDiscount_FailClosedResponseAndReader(t *testing.T) {
	t.Run("nil audit dependency is a server error before write", func(t *testing.T) {
		created := false
		repo := &mockCheckupTypeRepository{createFn: func(_ context.Context, _ *model.CheckupType) error { created = true; return nil }}
		svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, nil)
		_, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "x", ActorID: &testCheckupTypeActorID, IsBillDiscountExcluded: boolPtr(true)})
		assertCheckupTypeInternalError(t, err)
		if created {
			t.Fatal("write escaped fail-closed audit dependency")
		}
	})
	t.Run("audit failure rolls back the whole write", func(t *testing.T) {
		updated := false
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 3}, nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateCheckupTypeInput) (*model.CheckupType, error) {
				updated = true
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 4}, nil
			},
		}
		svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{err: errors.New("audit down")})
		_, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{IsBillDiscountExcluded: boolPtr(false), Version: int64Ptr(3), ActorID: &testCheckupTypeActorID})
		if err == nil || updated {
			t.Fatal("audit failure must fail the update")
		}
	})
	t.Run("response preserves null", func(t *testing.T) {
		raw, err := json.Marshal(toCheckupTypeResponse(&model.CheckupType{ID: 1, Version: 7}))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"is_bill_discount_excluded":null`)) {
			t.Fatalf("lost null: %s", raw)
		}
	})
	t.Run("reader error is not converted to null", func(t *testing.T) {
		repo := &mockCheckupTypeRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.CheckupType, error) {
			return nil, errors.New("column does not exist")
		}}
		if _, err := newTestCheckupTypeService(repo).GetByID(context.Background(), 1, 9); err == nil {
			t.Fatal("reader failure must surface as an error")
		}
	})
}
