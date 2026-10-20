package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// These scenarios use the actual services and a controlled repository. They do
// not prove PostgreSQL locking, trigger increments, or database rollback.
func TestMasterInsuranceCAS_EditingVersionAndAtomicAudit(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		for _, stale := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stale=%v", entity, stale), func(t *testing.T) {
				actor := uint64(12)
				setting := false
				audit := &insuranceAuditRecorder{}
				rolledBack := false
				tx := masterServiceJournalTx{rollback: func() { rolledBack = true }}
				deps := MasterInsuranceWriteDependencies{Transactor: tx, Audit: audit}
				beforeVersion := int64(7)
				if stale {
					beforeVersion = 9
				}
				seen := false
				check := func(ctx context.Context, version int64) error {
					if ctx.Value(insuranceAuditContextKey{}) != true || version != 7 {
						t.Fatal("editing version or TX context changed")
					}
					seen = true
					if stale {
						return apperrors.WrapConflict("stale master")
					}
					return nil
				}
				var version int64
				var err error
				if entity == "consultation" {
					repo := &mockConsultationRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Consultation, error) {
						return &model.Consultation{ID: 5, ClinicID: 7, Version: beforeVersion}, nil
					}, updateFieldsFn: func(ctx context.Context, cid, id uint64, in UpdateConsultationInput) (*model.Consultation, error) {
						if cid != 7 || id != 5 {
							t.Fatal("wrong scope")
						}
						if err := check(ctx, in.Version); err != nil {
							return nil, err
						}
						return &model.Consultation{ID: 5, ClinicID: 7, Version: 8, IsNonInsurance: in.IsNonInsurance}, nil
					}}
					row, e := NewConsultationServiceWithAudit(repo, deps).Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: 7, ActorID: &actor, IsNonInsurance: &setting})
					err = e
					if row != nil {
						version = toConsultationResponse(row).Version
					}
				} else {
					repo := &mockProcedureRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Procedure, error) {
						return &model.Procedure{ID: 5, ClinicID: 7, Version: beforeVersion}, nil
					}, updateFieldsFn: func(ctx context.Context, cid, id uint64, in UpdateProcedureInput) (*model.Procedure, error) {
						if cid != 7 || id != 5 {
							t.Fatal("wrong scope")
						}
						if err := check(ctx, in.Version); err != nil {
							return nil, err
						}
						return &model.Procedure{ID: 5, ClinicID: 7, Version: 8, IsNonInsurance: in.IsNonInsurance}, nil
					}}
					row, e := NewProcedureServiceWithAudit(repo, deps).Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: 7, ActorID: &actor, IsNonInsurance: &setting})
					err = e
					if row != nil {
						version = toProcedureResponse(row).Version
					}
				}
				if !seen {
					t.Fatal("repository not called")
				}
				if stale {
					if err == nil || !rolledBack || audit.entry != nil || version != 0 {
						t.Fatal("conflict reported/audited as success")
					}
				} else {
					if err != nil || rolledBack || audit.entry == nil || !audit.inTx || version != 8 {
						t.Fatalf("success did not preserve atomic version/audit: %v", err)
					}
				}
			})
		}
	}
}

func TestMasterInsuranceCAS_InvalidDirectServiceVersion(t *testing.T) {
	for _, version := range []int64{0, -1} {
		if _, err := NewConsultationServiceWithAudit(&mockConsultationRepository{}, masterTestWriteDependencies()).Update(t.Context(), 7, 5, &UpdateConsultationInput{Version: version}); err == nil {
			t.Fatal("invalid consultation version reached mutation")
		}
		if _, err := NewProcedureServiceWithAudit(&mockProcedureRepository{}, masterTestWriteDependencies()).Update(t.Context(), 7, 5, &UpdateProcedureInput{Version: version}); err == nil {
			t.Fatal("invalid procedure version reached mutation")
		}
	}
}

func TestMasterInsuranceCAS_PatchBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entity := range []string{"consultation", "procedure"} {
		for _, version := range []string{"", `,"version":null`, `,"version":0`, `,"version":-1`, `,"version":"7"`, `,"version":1.5`} {
			t.Run(entity+version, func(t *testing.T) {
				r := gin.New()
				r.Use(func(c *gin.Context) { setClinicID(c); setStaffID(c) })
				if entity == "consultation" {
					r.PATCH("/:id", NewConsultationHandler(&mockConsultationService{}).UpdateConsultation)
				} else {
					r.PATCH("/:id", NewProcedureHandler(&mockProcedureService{}).UpdateProcedure)
				}
				req := httptest.NewRequest(http.MethodPatch, "/5", strings.NewReader(`{"name":"changed"`+version+`}`))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	}
}

// DryRun and callbacks exercise actual SQL/error mapping without a DB connection.
func TestMasterInsuranceCAS_RepositorySQLAndMisses(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		for _, missing := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/missing=%v", entity, missing), func(t *testing.T) {
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
				if missing {
					if err := db.Callback().Query().After("gorm:query").Register("cas-test-missing", func(tx *gorm.DB) { _ = tx.AddError(gorm.ErrRecordNotFound) }); err != nil {
						t.Fatal(err)
					}
				}
				ctx := persistence.WithTxValue(t.Context(), db)
				value := false
				if entity == "consultation" {
					_, err = NewConsultationRepository(nil).Update(ctx, 7, 5, UpdateConsultationInput{Version: 13, IsNonInsurance: &value})
				} else {
					_, err = NewProcedureRepository(nil).Update(ctx, 7, 5, UpdateProcedureInput{Version: 13, IsNonInsurance: &value})
				}
				if err == nil {
					t.Fatal("zero affected rows reported success")
				}
				// The HTTP handlers use the same error chain to select 404 versus 409.
				want := apperrors.ErrConflict
				if missing {
					want = apperrors.ErrNotFound
				}
				if !errors.Is(err, want) {
					t.Fatalf("got %v want %v", err, want)
				}
				sql := output.String()
				if !strings.Contains(sql, "clinic_id = 7") || !strings.Contains(sql, "id = 5 AND version = 13") || !strings.Contains(sql, `"is_non_insurance"=false`) || strings.Contains(sql, `"version"=`) {
					t.Fatalf("unsafe CAS SQL: %s", sql)
				}
			})
		}
	}
}

func TestMasterInsuranceCAS_ResponsePreservesVersionAndNull(t *testing.T) {
	for _, v := range []any{toConsultationResponse(&model.Consultation{Version: 17}), toProcedureResponse(&model.Procedure{Version: 17})} {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"version":17`)) || !bytes.Contains(raw, []byte(`"is_non_insurance":null`)) {
			t.Fatalf("lost version or null: %s", raw)
		}
	}
}

func TestMasterInsuranceCAS_HTTPConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entity := range []string{"consultation", "procedure"} {
		t.Run(entity, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) { setClinicID(c); setStaffID(c) })
			if entity == "consultation" {
				r.PATCH("/:id", NewConsultationHandler(&mockConsultationService{updateFn: func(_ context.Context, _, _ uint64, in *UpdateConsultationInput) (*model.Consultation, error) {
					if in.Version != 7 {
						t.Fatal("expected version changed")
					}
					return nil, apperrors.WrapConflict("stale master")
				}}).UpdateConsultation)
			} else {
				r.PATCH("/:id", NewProcedureHandler(&mockProcedureService{updateFn: func(_ context.Context, _, _ uint64, in *UpdateProcedureInput) (*model.Procedure, error) {
					if in.Version != 7 {
						t.Fatal("expected version changed")
					}
					return nil, apperrors.WrapConflict("stale master")
				}}).UpdateProcedure)
			}
			req := httptest.NewRequest(http.MethodPatch, "/5", strings.NewReader(`{"name":"changed","version":7}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusConflict {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// Simulate storage's second write (inactive default override) independently of
// the inserted Go value, proving the response uses the same-TX stored reread.
type casConsultationRereadRepo struct{ mockConsultationRepository }

func (r *casConsultationRereadRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	if ctx.Value(insuranceAuditContextKey{}) != true {
		return nil, errors.New("read outside TX")
	}
	return &model.Consultation{ID: id, ClinicID: clinicID, Version: 2, IsActive: false}, nil
}

type casProcedureRereadRepo struct{ mockProcedureRepository }

func (r *casProcedureRereadRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	if ctx.Value(insuranceAuditContextKey{}) != true {
		return nil, errors.New("read outside TX")
	}
	return &model.Procedure{ID: id, ClinicID: clinicID, Version: 2, IsActive: false}, nil
}
func TestMasterInsuranceCAS_InactiveCreateReturnsStoredVersion(t *testing.T) {
	actor := uint64(12)
	deps := MasterInsuranceWriteDependencies{Transactor: masterServiceJournalTx{rollback: func() {}}, Audit: &insuranceAuditRecorder{}}
	cr := &casConsultationRereadRepo{mockConsultationRepository{createFn: func(_ context.Context, c *model.Consultation) error { c.ID = 5; c.Version = 1; return nil }}}
	c, err := NewConsultationServiceWithAudit(cr, deps).Create(t.Context(), 7, &CreateConsultationInput{ActorID: &actor, Name: "inactive", IsActive: false})
	if err != nil || c == nil || c.Version != 2 || c.IsActive {
		t.Fatalf("stale create response: %v %v", c, err)
	}
	pr := &casProcedureRereadRepo{mockProcedureRepository{createFn: func(_ context.Context, p *model.Procedure) error { p.ID = 5; p.Version = 1; return nil }}}
	p, err := NewProcedureServiceWithAudit(pr, deps).Create(t.Context(), 7, &CreateProcedureInput{ActorID: &actor, Name: "inactive", Anesthesia: "none", IsActive: false})
	if err != nil || p == nil || p.Version != 2 || p.IsActive {
		t.Fatalf("stale create response: %v %v", p, err)
	}
}

func TestMasterInsuranceCAS_RepositoryRejectsInvalidWithoutDB(t *testing.T) {
	for _, version := range []int64{0, -1} {
		if _, err := NewConsultationRepository(nil).Update(t.Context(), 7, 5, UpdateConsultationInput{Version: version}); !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("invalid consultation version: %v", err)
		}
		if _, err := NewProcedureRepository(nil).Update(t.Context(), 7, 5, UpdateProcedureInput{Version: version}); !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("invalid procedure version: %v", err)
		}
	}
	if _, err := NewConsultationRepository(nil).Update(t.Context(), 7, 5, UpdateConsultationInput{Version: 1}); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("version-only consultation: %v", err)
	}
	if _, err := NewProcedureRepository(nil).Update(t.Context(), 7, 5, UpdateProcedureInput{Version: 1}); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("version-only procedure: %v", err)
	}
}
