package clinic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type clinicTxMarker struct{}

type recordingClinicTransactor struct {
	called bool
}

func (t *recordingClinicTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	t.called = true
	return fn(context.WithValue(ctx, clinicTxMarker{}, true))
}

type atomicClinicStore struct {
	createdInTx bool
}

func (r *atomicClinicStore) FindAll(context.Context) ([]model.Clinic, error) {
	return nil, nil
}

func (r *atomicClinicStore) FindByStaffID(context.Context, uint64) ([]model.Clinic, error) {
	return nil, nil
}

func (r *atomicClinicStore) FindByIDs(context.Context, []uint64) ([]model.Clinic, error) {
	return nil, nil
}

func (r *atomicClinicStore) FindActiveIDs(context.Context, []uint64) ([]uint64, error) {
	return nil, nil
}

func (r *atomicClinicStore) FindByID(context.Context, uint64) (*model.Clinic, error) {
	return nil, nil
}

func (r *atomicClinicStore) LockByIDForUpdate(context.Context, uint64) (*model.Clinic, error) {
	return nil, nil
}

func (r *atomicClinicStore) FindCompany(context.Context) (*model.Company, error) {
	return &model.Company{ID: 7}, nil
}

func (r *atomicClinicStore) Create(ctx context.Context, clinic *model.Clinic) error {
	r.createdInTx, _ = ctx.Value(clinicTxMarker{}).(bool)
	clinic.ID = 42
	return nil
}

func (r *atomicClinicStore) UpdateClinic(context.Context, uint64, *UpdateClinicInput) error {
	return nil
}

func (r *atomicClinicStore) Delete(context.Context, uint64) error {
	return nil
}

func (r *atomicClinicStore) CountOwnersByClinicID(context.Context, uint64) (int64, error) {
	return 0, nil
}

func (r *atomicClinicStore) CountStaffByClinicID(context.Context, uint64) (int64, error) {
	return 0, nil
}

func (r *atomicClinicStore) CountBlockingReferencesByClinicID(context.Context, uint64) ([]ClinicDependencyCount, error) {
	return nil, nil
}

type atomicPermissionGroupWriter struct {
	createInTx      []bool
	updateRulesInTx []bool
	ruleErr         error
}

func (w *atomicPermissionGroupWriter) Create(ctx context.Context, group *model.PermissionGroup) error {
	inTx, _ := ctx.Value(clinicTxMarker{}).(bool)
	w.createInTx = append(w.createInTx, inTx)
	group.ID = uint64(len(w.createInTx))
	return nil
}

func (w *atomicPermissionGroupWriter) UpdateRules(
	ctx context.Context,
	_ uint64,
	_ uint64,
	rules []model.PermissionGroupRule,
) error {
	inTx, _ := ctx.Value(clinicTxMarker{}).(bool)
	w.updateRulesInTx = append(w.updateRulesInTx, inTx)
	if len(rules) == 0 {
		return errors.New("default permission rules must not be empty")
	}
	return w.ruleErr
}

func (w *atomicPermissionGroupWriter) DeleteSoftDeletedByClinicID(context.Context, uint64) error {
	return nil
}

func TestService_CreateClinic_UsesOneTransactionForClinicAndDefaultPermissions(t *testing.T) {
	store := &atomicClinicStore{}
	writer := &atomicPermissionGroupWriter{}
	tx := &recordingClinicTransactor{}
	svc := NewService(store, writer, tx)

	created, err := svc.CreateClinic(context.Background(), &CreateClinicInput{Name: "新医院"})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, uint64(42), created.ID)
	assert.True(t, tx.called)
	assert.True(t, store.createdInTx)
	assert.Equal(t, []bool{true, true}, writer.createInTx)
	assert.Equal(t, []bool{true, true}, writer.updateRulesInTx)
}

func TestService_CreateClinic_PropagatesPermissionRuleFailure(t *testing.T) {
	ruleErr := errors.New("rule write failed")
	store := &atomicClinicStore{}
	writer := &atomicPermissionGroupWriter{ruleErr: ruleErr}
	tx := &recordingClinicTransactor{}
	svc := NewService(store, writer, tx)

	created, err := svc.CreateClinic(context.Background(), &CreateClinicInput{Name: "新医院"})

	assert.Nil(t, created)
	require.Error(t, err)
	assert.ErrorIs(t, err, ruleErr)
	assert.True(t, tx.called)
	assert.True(t, store.createdInTx)
	assert.Equal(t, []bool{true}, writer.createInTx)
	assert.Equal(t, []bool{true}, writer.updateRulesInTx)
}
