package medicalrecord

// EMR-447 A01: consultation の is_bill_discount_excluded 契約。
// POST 省略=false 明示保存 / 明示 true/false 保持 / null・非 boolean=400。
// PATCH 省略=変更なし（NULL/false/true 保持）/ 明示 true/false=SET+監査。
// 監査は保険監査と別 entry（field=is_bill_discount_excluded）で同一 tx、失敗で全体 rollback。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// billAuditRecorder は全監査 entry を記録する（保険+割引の2 entry を検証するため）。
type billAuditRecorder struct {
	entries []*AuditEntry
	err     error
}

func (a *billAuditRecorder) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	if a.err != nil {
		return a.err
	}
	a.entries = append(a.entries, entry)
	return nil
}

func billAuditField(entry *AuditEntry) string {
	m, _ := entry.Metadata.(map[string]any)
	f, _ := m["field"].(string)
	return f
}

func TestConsultationBillDiscount_RequestBinding(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		wantNil bool
		wantVal bool
		wantErr bool
	}{
		{name: "create omitted stays nil for service default", payload: `{"name":"a"}`, wantNil: true},
		{name: "create true carried", payload: `{"name":"a","is_bill_discount_excluded":true}`, wantVal: true},
		{name: "create false carried", payload: `{"name":"a","is_bill_discount_excluded":false}`, wantVal: false},
		{name: "create null rejected", payload: `{"name":"a","is_bill_discount_excluded":null}`, wantErr: true},
		{name: "create non-boolean rejected", payload: `{"name":"a","is_bill_discount_excluded":"yes"}`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var req createConsultationRequest
			err := json.Unmarshal([]byte(tc.payload), &req)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected bind error")
				}
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("want invalid input, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := req.IsBillDiscountExcluded.toCreateInput()
			if tc.wantNil != (got == nil) {
				t.Fatalf("nil=%v want nil=%v", got, tc.wantNil)
			}
			if got != nil && *got != tc.wantVal {
				t.Fatalf("got %v want %v", *got, tc.wantVal)
			}
		})
	}
	t.Run("update omitted means unchanged", func(t *testing.T) {
		var req updateConsultationRequest
		if err := json.Unmarshal([]byte(`{"version":2}`), &req); err != nil {
			t.Fatal(err)
		}
		if req.IsBillDiscountExcluded.toUpdateInput() != nil {
			t.Fatal("omitted must stay nil")
		}
	})
	t.Run("update null and non-boolean rejected", func(t *testing.T) {
		var req updateConsultationRequest
		if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":null}`), &req); !apperrors.IsInvalidInput(err) {
			t.Fatalf("null: %v", err)
		}
		var req2 updateConsultationRequest
		if err := json.Unmarshal([]byte(`{"version":2,"is_bill_discount_excluded":1}`), &req2); !apperrors.IsInvalidInput(err) {
			t.Fatalf("non-boolean: %v", err)
		}
	})
}

func TestConsultationBillDiscount_CreateSavesExplicitValueAndAudits(t *testing.T) {
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
			var stored *model.Consultation
			audit := &billAuditRecorder{}
			repo := &mockConsultationRepository{createFn: func(_ context.Context, c *model.Consultation) error {
				c.ID = 5
				stored = c
				return nil
			}}
			deps := MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}
			_, err := NewConsultationServiceWithAudit(repo, deps).Create(t.Context(), 7, &CreateConsultationInput{ActorID: masterTestActorID(), Name: "item", IsBillDiscountExcluded: tc.input})
			if err != nil {
				t.Fatal(err)
			}
			if stored == nil || stored.IsBillDiscountExcluded == nil || *stored.IsBillDiscountExcluded != tc.wantSave {
				t.Fatalf("stored=%+v want %v", stored, tc.wantSave)
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
			if bill.OldValue != nil {
				t.Fatal("create has no before row")
			}
			got := bill.NewValue.(map[string]any)[colIsBillDiscountExcluded]
			if got != tc.wantSave {
				t.Fatalf("audit new value %v want %v", got, tc.wantSave)
			}
			if *bill.ClinicID != 7 || *bill.ActorID != 12 || *bill.ResourceID != 5 || bill.Resource != "consultation" || bill.Action != "consultation.create" {
				t.Fatalf("wrong audit identity: %+v", bill)
			}
		})
	}
}

func TestConsultationBillDiscount_PatchPreservesAndUpdates(t *testing.T) {
	t.Run("omitted writes nothing and audits nothing", func(t *testing.T) {
		fields := buildConsultationUpdate(&UpdateConsultationInput{Version: 1, ActorID: masterTestActorID()})
		if _, ok := fields[colIsBillDiscountExcluded]; ok {
			t.Fatal("omitted must not enter the update map")
		}
	})
	t.Run("explicit false is written through the map", func(t *testing.T) {
		fields := buildConsultationUpdate(&UpdateConsultationInput{Version: 1, ActorID: masterTestActorID(), IsBillDiscountExcluded: boolPtr(false)})
		if v, ok := fields[colIsBillDiscountExcluded]; !ok || v != false {
			t.Fatalf("explicit false lost: %v", fields)
		}
	})
	for _, tc := range []struct {
		name        string
		stored      *bool
		input       *bool
		wantBefore  any
		wantAfter   any
	}{
		{name: "NULL to false audited", stored: nil, input: boolPtr(false), wantBefore: nil, wantAfter: false},
		{name: "true to false audited", stored: boolPtr(true), input: boolPtr(false), wantBefore: true, wantAfter: false},
		{name: "false to true audited", stored: boolPtr(false), input: boolPtr(true), wantBefore: false, wantAfter: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			audit := &billAuditRecorder{}
			repo := &mockConsultationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
					return &model.Consultation{ID: 5, ClinicID: 7, IsBillDiscountExcluded: tc.stored}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateConsultationInput) (*model.Consultation, error) {
					return &model.Consultation{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
				},
			}
			row, err := NewConsultationServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
				Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 7, ActorID: masterTestActorID(), IsBillDiscountExcluded: tc.input})
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
			if bill == nil || bill.Action != "consultation.update" {
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
		name := "changed"
		audit := &billAuditRecorder{}
		repo := &mockConsultationRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
				return &model.Consultation{ID: 5, ClinicID: 7}, nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateConsultationInput) (*model.Consultation, error) {
				return &model.Consultation{ID: 5, ClinicID: 7, Version: 8}, nil
			},
		}
		if _, err := NewConsultationServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: audit}).
			Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 7, ActorID: masterTestActorID(), Name: &name}); err != nil {
			t.Fatal(err)
		}
		for _, e := range audit.entries {
			if billAuditField(e) == colIsBillDiscountExcluded {
				t.Fatal("name-only patch must not audit bill discount")
			}
		}
	})
}

func TestConsultationBillDiscount_FailClosedAndRollback(t *testing.T) {
	t.Run("constructor without dependencies rejects the field before mutation", func(t *testing.T) {
		called := false
		repo := &mockConsultationRepository{createFn: func(context.Context, *model.Consultation) error { called = true; return nil }}
		_, err := NewConsultationService(repo).Create(t.Context(), 7, &CreateConsultationInput{ActorID: masterTestActorID(), Name: "item", IsBillDiscountExcluded: boolPtr(true)})
		if err == nil || called {
			t.Fatal("missing audit dependency must fail closed")
		}
	})
	t.Run("audit failure rolls back the whole write", func(t *testing.T) {
		rolledBack := false
		tx := masterServiceJournalTx{rollback: func() { rolledBack = true }}
		audit := &billAuditRecorder{err: errors.New("audit unavailable")}
		repo := &mockConsultationRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
				return &model.Consultation{ID: 5, ClinicID: 7}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, in UpdateConsultationInput) (*model.Consultation, error) {
				return &model.Consultation{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: in.IsBillDiscountExcluded}, nil
			},
		}
		row, err := NewConsultationServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}).
			Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 7, ActorID: masterTestActorID(), IsBillDiscountExcluded: boolPtr(false)})
		if err == nil || row != nil || !rolledBack {
			t.Fatalf("audit failure must roll back: %v %v", row, err)
		}
	})
	t.Run("version conflict audits nothing", func(t *testing.T) {
		audit := &billAuditRecorder{}
		repo := &mockConsultationRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
				return &model.Consultation{ID: 5, ClinicID: 7}, nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateConsultationInput) (*model.Consultation, error) {
				return nil, apperrors.WrapConflict("stale master")
			},
		}
		_, err := NewConsultationServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: masterServiceJournalTx{rollback: func() {}}, Audit: audit}).
			Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 7, ActorID: masterTestActorID(), IsBillDiscountExcluded: boolPtr(false)})
		if !apperrors.IsConflict(err) || len(audit.entries) != 0 {
			t.Fatalf("conflict must not audit: %v %+v", err, audit.entries)
		}
	})
}

func TestConsultationBillDiscount_ResponseAndReader(t *testing.T) {
	t.Run("stored NULL stays null in the response", func(t *testing.T) {
		raw, err := json.Marshal(toConsultationResponse(&model.Consultation{Version: 17}))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"is_bill_discount_excluded":null`)) {
			t.Fatalf("lost null: %s", raw)
		}
	})
	t.Run("readers select the column explicitly", func(t *testing.T) {
		var output bytes.Buffer
		db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = sqlDB.Close() })
		repo := NewConsultationRepository(db)
		ctx := persistence.WithTxValue(t.Context(), db)
		if _, err := repo.FindByID(ctx, 7, 5); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.FindAll(ctx, 7); err != nil {
			t.Fatal(err)
		}
		if sql := output.String(); strings.Count(sql, `"is_bill_discount_excluded"`) < 2 {
			t.Fatalf("reader does not select the column: %s", sql)
		}
	})
	t.Run("reader error is not converted to null", func(t *testing.T) {
		repo := &mockConsultationRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
			return nil, errors.New("column does not exist")
		}}
		if _, err := NewConsultationServiceWithAudit(repo, masterTestWriteDependencies()).GetByID(t.Context(), 7, 5); err == nil {
			t.Fatal("reader failure must surface as an error")
		}
	})
}
