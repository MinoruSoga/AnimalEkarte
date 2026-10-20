package medicalrecord

import (
	"context"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

type patchAuditTransactor struct {
	run func(context.Context, func(context.Context) error) error
}

func (tx patchAuditTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return tx.run(ctx, fn)
}

func TestMasterInsurancePatchAudit_OnlyExplicitInsuranceInput(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		for _, explicit := range []bool{false, true} {
			t.Run(entity+map[bool]string{false: "/name only", true: "/explicit unchanged"}[explicit], func(t *testing.T) {
				setting, actor, name := true, uint64(12), "renamed"
				var inputSetting *bool
				if explicit {
					inputSetting = &setting
				}
				audit := &insuranceAuditRecorder{}
				txCalls := 0
				tx := patchAuditTransactor{run: func(ctx context.Context, fn func(context.Context) error) error {
					txCalls++
					return fn(context.WithValue(ctx, insuranceAuditContextKey{}, true))
				}}
				deps := MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}
				wrote := false
				verify := func(ctx context.Context) {
					if active, _ := ctx.Value(insuranceAuditContextKey{}).(bool); !active {
						t.Fatal("update outside transaction")
					}
					wrote = true
				}
				if entity == "consultation" {
					repo := &mockConsultationRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
						return &model.Consultation{ID: 5, IsNonInsurance: &setting}, nil
					}, updateFieldsFn: func(ctx context.Context, _, _ uint64, in UpdateConsultationInput) (*model.Consultation, error) {
						verify(ctx)
						return &model.Consultation{ID: 5, IsNonInsurance: &setting}, nil
					}}
					if _, err := NewConsultationServiceWithAudit(repo, deps).Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 1, ActorID: &actor, Name: &name, IsNonInsurance: inputSetting}); err != nil {
						t.Fatal(err)
					}
				} else {
					repo := &mockProcedureRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
						return &model.Procedure{ID: 5, IsNonInsurance: &setting}, nil
					}, updateFieldsFn: func(ctx context.Context, _, _ uint64, in UpdateProcedureInput) (*model.Procedure, error) {
						verify(ctx)
						return &model.Procedure{ID: 5, IsNonInsurance: &setting}, nil
					}}
					if _, err := NewProcedureServiceWithAudit(repo, deps).Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: 1, ActorID: &actor, Name: &name, IsNonInsurance: inputSetting}); err != nil {
						t.Fatal(err)
					}
				}
				if !wrote || txCalls != 1 {
					t.Fatal("mutation did not share owner transaction")
				}
				if (audit.entry != nil) != explicit {
					t.Fatalf("insurance audit presence=%v explicit=%v", audit.entry != nil, explicit)
				}
				if explicit && !audit.inTx {
					t.Fatal("audit escaped transaction")
				}
			})
		}
	}
}
