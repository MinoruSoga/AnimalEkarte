package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

type masterServiceJournalTx struct {
	rollback  func()
	commitErr error
}

func (tx masterServiceJournalTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(context.WithValue(ctx, insuranceAuditContextKey{}, true)); err != nil {
		tx.rollback()
		return err
	}
	if tx.commitErr != nil {
		tx.rollback()
		return tx.commitErr
	}
	return nil
}
func TestMasterInsuranceServices_AuditAndCommitFailureReturnNoSuccess(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		for _, update := range []bool{false, true} {
			for _, failure := range []string{"audit", "commit"} {
				t.Run(fmt.Sprintf("%s/update=%v/%s", entity, update, failure), func(t *testing.T) {
					var stored *bool
					tx := masterServiceJournalTx{rollback: func() { stored = nil }}
					audit := &insuranceAuditRecorder{}
					if failure == "audit" {
						audit.err = errors.New("audit unavailable")
					} else {
						tx.commitErr = errors.New("commit unavailable")
					}
					deps := MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}
					setting, actor := true, uint64(12)
					var err error
					if entity == "consultation" {
						repo := &mockConsultationRepository{
							findByIDFn: func(_ context.Context, cid, id uint64) (*model.Consultation, error) {
								if cid != 7 || id != 5 {
									t.Fatal("scope changed")
								}
								return &model.Consultation{ID: 5, ClinicID: 7, IsNonInsurance: stored}, nil
							},
							createFn: func(_ context.Context, c *model.Consultation) error { c.ID = 5; stored = c.IsNonInsurance; return nil },
							updateFieldsFn: func(_ context.Context, cid, id uint64, in UpdateConsultationInput) (*model.Consultation, error) {
								stored = in.IsNonInsurance
								return &model.Consultation{ID: 5, ClinicID: 7, IsNonInsurance: stored}, nil
							},
						}
						svc := NewConsultationServiceWithAudit(repo, deps)
						var result *model.Consultation
						if update {
							result, err = svc.Update(t.Context(), 7, 5, &UpdateConsultationInput{ActorID: &actor, IsNonInsurance: &setting})
						} else {
							result, err = svc.Create(t.Context(), 7, &CreateConsultationInput{ActorID: &actor, Name: "item", IsNonInsurance: &setting})
						}
						if result != nil {
							t.Fatal("returned success after transaction failure")
						}
					} else {
						repo := &mockProcedureRepository{
							findByIDFn: func(_ context.Context, cid, id uint64) (*model.Procedure, error) {
								if cid != 7 || id != 5 {
									t.Fatal("scope changed")
								}
								return &model.Procedure{ID: 5, ClinicID: 7, IsNonInsurance: stored}, nil
							},
							createFn: func(_ context.Context, p *model.Procedure) error { p.ID = 5; stored = p.IsNonInsurance; return nil },
							updateFieldsFn: func(_ context.Context, cid, id uint64, in UpdateProcedureInput) (*model.Procedure, error) {
								stored = in.IsNonInsurance
								return &model.Procedure{ID: 5, ClinicID: 7, IsNonInsurance: stored}, nil
							},
						}
						svc := NewProcedureServiceWithAudit(repo, deps)
						var result *model.Procedure
						if update {
							result, err = svc.Update(t.Context(), 7, 5, &UpdateProcedureInput{ActorID: &actor, IsNonInsurance: &setting})
						} else {
							result, err = svc.Create(t.Context(), 7, &CreateProcedureInput{ActorID: &actor, Name: "item", Anesthesia: "none", IsNonInsurance: &setting})
						}
						if result != nil {
							t.Fatal("returned success after transaction failure")
						}
					}
					if err == nil || stored != nil {
						t.Fatal("transaction failure was not propagated/rolled back")
					}
					if !audit.inTx || audit.entry == nil || *audit.entry.ActorID != 12 {
						t.Fatal("audit did not share authenticated transaction")
					}
				})
			}
		}
	}
}
