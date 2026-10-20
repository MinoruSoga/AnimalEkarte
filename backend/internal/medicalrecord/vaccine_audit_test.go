package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type vaccineTestTransactorFn func(context.Context, func(context.Context) error) error

func (fn vaccineTestTransactorFn) WithTx(ctx context.Context, body func(context.Context) error) error {
	return fn(ctx, body)
}

type vaccineTestAuditFn func(context.Context, *AuditEntry) error

func (fn vaccineTestAuditFn) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	return fn(ctx, entry)
}
func vaccineTestActorID() *uint64 { id := uint64(1); return &id }
func newVaccineServiceForTest(repo VaccineRepository) VaccineService {
	return NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{
		Transactor: vaccineTestTransactorFn(func(ctx context.Context, body func(context.Context) error) error { return body(ctx) }),
		Audit:      vaccineTestAuditFn(func(context.Context, *AuditEntry) error { return nil }),
	})
}

type vaccineAuditTXKey struct{}
type vaccineAuditRepo struct {
	VaccineRepository
	t      *testing.T
	row    *model.Vaccine
	events *[]string
}

func cloneVaccineAuditRow(row *model.Vaccine) *model.Vaccine {
	if row == nil {
		return nil
	}
	copy := *row
	if row.IsNonInsurance != nil {
		value := *row.IsNonInsurance
		copy.IsNonInsurance = &value
	}
	return &copy
}
func (r *vaccineAuditRepo) assertTX(ctx context.Context) {
	r.t.Helper()
	require.Equal(r.t, true, ctx.Value(vaccineAuditTXKey{}), "repository must use transaction context")
}
func (r *vaccineAuditRepo) Create(ctx context.Context, row *model.Vaccine) error {
	r.assertTX(ctx)
	*r.events = append(*r.events, "create")
	row.ID = 17
	r.row = cloneVaccineAuditRow(row)
	return nil
}
func (r *vaccineAuditRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	r.assertTX(ctx)
	*r.events = append(*r.events, "read")
	require.Equal(r.t, uint64(2), clinicID)
	require.Equal(r.t, uint64(17), id)
	return r.row, nil
}
func (r *vaccineAuditRepo) LockWriteRows(ctx context.Context, clinicID uint64, ids []uint64) error {
	r.assertTX(ctx)
	*r.events = append(*r.events, "lock-target")
	require.Equal(r.t, uint64(2), clinicID)
	require.Equal(r.t, []uint64{17}, ids)
	return nil
}

func (r *vaccineAuditRepo) Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
	r.assertTX(ctx)
	*r.events = append(*r.events, "update")
	if input.Name != nil {
		r.row.Name = *input.Name
	}
	if input.IsNonInsurance != nil {
		if r.row.IsNonInsurance != nil {
			*r.row.IsNonInsurance = *input.IsNonInsurance
		} else {
			value := *input.IsNonInsurance
			r.row.IsNonInsurance = &value
		}
	}
	return r.FindByID(ctx, clinicID, id)
}
func TestVaccineService_AtomicInsuranceAudit(t *testing.T) {
	for _, operation := range []string{"create-default", "create-explicit-false", "update-legacy-null", "update-omit-setting", "update-existing-true"} {
		for _, failure := range []string{"none", "audit", "commit"} {
			t.Run(operation+"/"+failure, func(t *testing.T) {
				var events []string
				repo := &vaccineAuditRepo{t: t, events: &events}
				updating := operation == "update-legacy-null" || operation == "update-omit-setting" || operation == "update-existing-true"
				if updating {
					repo.row = &model.Vaccine{ID: 17, ClinicID: 2, Name: "before", IsNonInsurance: nil}
				}
				if operation == "update-existing-true" {
					value := true
					repo.row.IsNonInsurance = &value
				}
				before := cloneVaccineAuditRow(repo.row)
				var audit *AuditEntry
				deps := MasterInsuranceWriteDependencies{
					Transactor: vaccineTestTransactorFn(func(ctx context.Context, body func(context.Context) error) error {
						events = append(events, "begin")
						if err := body(context.WithValue(ctx, vaccineAuditTXKey{}, true)); err != nil {
							repo.row = before
							return err
						}
						if failure == "commit" {
							repo.row = before
							return errors.New("commit failed")
						}
						events = append(events, "commit")
						return nil
					}),
					Audit: vaccineTestAuditFn(func(ctx context.Context, entry *AuditEntry) error {
						require.Equal(t, true, ctx.Value(vaccineAuditTXKey{}))
						events = append(events, "audit")
						audit = entry
						if failure == "audit" {
							return errors.New("audit unavailable")
						}
						return nil
					}),
				}
				svc := NewVaccineServiceWithAudit(repo, deps)
				var got *model.Vaccine
				var err error
				setting := false
				if updating {
					name := "after"
					input := &UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), Name: &name}
					if operation == "update-legacy-null" || operation == "update-existing-true" {
						input.IsNonInsurance = &setting
					}
					got, err = svc.Update(t.Context(), 2, 17, input)
				} else {
					input := &CreateVaccineInput{ActorID: vaccineTestActorID(), Name: "created"}
					if operation == "create-explicit-false" {
						input.IsNonInsurance = &setting
					}
					got, err = svc.Create(t.Context(), 2, input)
				}
				if operation == "update-omit-setting" {
					require.Nil(t, audit, "name-only PATCH must not emit insurance audit")
					if failure == "commit" {
						require.Error(t, err)
						require.Nil(t, got)
						require.Equal(t, before, repo.row)
					} else {
						require.NoError(t, err)
						require.NotNil(t, got)
						require.Nil(t, got.IsNonInsurance, "legacy NULL must be preserved")
						require.Equal(t, "after", got.Name)
						require.Equal(t, []string{"begin", "lock-target", "read", "update", "read", "commit"}, events)
					}
					return
				}
				if failure != "none" {
					require.Error(t, err)
					require.Nil(t, got)
					require.Equal(t, before, repo.row)
				} else {
					require.NoError(t, err)
					require.NotNil(t, got)
				}
				require.NotNil(t, audit)
				require.Equal(t, uint64(2), *audit.ClinicID)
				require.Equal(t, uint64(1), *audit.ActorID)
				require.Equal(t, uint64(17), *audit.ResourceID)
				require.Equal(t, "vaccine", audit.Resource)
				if updating {
					require.Equal(t, "vaccine.update", audit.Action)
					var oldSetting any
					if operation == "update-existing-true" {
						oldSetting = true
					}
					require.Equal(t, map[string]any{"is_non_insurance": oldSetting}, audit.OldValue)
				} else {
					require.Equal(t, "vaccine.create", audit.Action)
					require.Nil(t, audit.OldValue)
				}
				var wantSetting any = false
				if operation == "create-default" {
					wantSetting = true
				}
				if operation == "update-omit-setting" {
					wantSetting = nil
				}
				require.Equal(t, map[string]any{"is_non_insurance": wantSetting}, audit.NewValue)
				wantEvents := []string{"begin", "create", "read", "audit"}
				if updating {
					wantEvents = []string{"begin", "lock-target", "read", "update", "read", "audit"}
				}
				if failure == "none" {
					wantEvents = append(wantEvents, "commit")
				}
				require.Equal(t, wantEvents, events)
			})
		}
	}
}
func TestVaccineService_MissingAuditDependenciesOrActorStopsBeforeMutation(t *testing.T) {
	for _, missing := range []string{"all-dependencies", "actor", "clinic"} {
		t.Run(missing, func(t *testing.T) {
			called := false
			deps := MasterInsuranceWriteDependencies{
				Transactor: vaccineTestTransactorFn(func(context.Context, func(context.Context) error) error { called = true; return nil }),
				Audit:      vaccineTestAuditFn(func(context.Context, *AuditEntry) error { called = true; return nil }),
			}
			actor, clinicID := vaccineTestActorID(), uint64(2)
			if missing == "all-dependencies" {
				deps = MasterInsuranceWriteDependencies{}
			}
			if missing == "actor" {
				actor = nil
			}
			if missing == "clinic" {
				clinicID = 0
			}
			svc := NewVaccineServiceWithAudit(&vaccineAuditRepo{t: t}, deps)
			got, err := svc.Create(t.Context(), clinicID, &CreateVaccineInput{ActorID: actor, Name: "valid"})
			require.Error(t, err)
			require.Nil(t, got)
			require.False(t, called)
			updated, err := svc.Update(t.Context(), clinicID, 17, &UpdateVaccineInput{Version: 1, ActorID: actor})
			require.Error(t, err)
			require.Nil(t, updated)
			require.False(t, called)
		})
	}
}

func TestVaccineService_ParentValidationAfterWriteLocks(t *testing.T) {
	var events []string
	name, parentID := "after", uint64(10)
	repo := &mockVaccineRepository{
		lockWriteRowsFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
			require.Equal(t, uint64(2), clinicID)
			require.Equal(t, []uint64{10, 17}, ids)
			events = append(events, "write-locks")
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Vaccine, error) {
			require.NotEmpty(t, events)
			require.Equal(t, "write-locks", events[0])
			events = append(events, "read")
			return &model.Vaccine{ID: id, ClinicID: clinicID}, nil
		},
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateVaccineInput) (*model.Vaccine, error) {
			events = append(events, "update")
			return &model.Vaccine{ID: 17, ClinicID: 2}, nil
		},
	}
	_, err := newVaccineServiceForTest(repo).Update(t.Context(), 2, 17, &UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), Name: &name, ParentID: &parentID})
	require.NoError(t, err)
	require.Equal(t, []string{"write-locks", "read", "read", "update"}, events)
}
