package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepTriggerPriorityRepository ----

type mockTriggerPriorityRepo struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error)
	upsertBatchFn    func(ctx context.Context, clinicID uint64, items []model.LstepTriggerPriority) error
	getPriorityFn    func(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

func (m *mockTriggerPriorityRepo) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockTriggerPriorityRepo) UpsertBatch(ctx context.Context, clinicID uint64, items []model.LstepTriggerPriority) error {
	if m.upsertBatchFn != nil {
		return m.upsertBatchFn(ctx, clinicID, items)
	}
	return nil
}

func (m *mockTriggerPriorityRepo) FindPriorityByTriggerType(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	if m.getPriorityFn != nil {
		return m.getPriorityFn(ctx, clinicID, triggerType)
	}
	return 0, apperrors.WrapNotFound("lstep_trigger_priority", "0")
}

func TestGetPriorityFor_DBHit(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		getPriorityFn: func(_ context.Context, _ uint64, _ string) (int, error) {
			return 5, nil
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	p, err := svc.GetPriorityFor(context.Background(), 1, model.TriggerTypeDormantPrevention365)
	assert.NoError(t, err)
	assert.Equal(t, 5, p)
}

func TestGetPriorityFor_DefaultMapFallback(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		getPriorityFn: func(_ context.Context, _ uint64, _ string) (int, error) {
			return 0, apperrors.WrapNotFound("lstep_trigger_priority", "0")
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	p, err := svc.GetPriorityFor(context.Background(), 1, model.TriggerTypeDormantPrevention365)
	assert.NoError(t, err)
	assert.Equal(t, model.DefaultTriggerPriorities[model.TriggerTypeDormantPrevention365], p)
}

func TestGetPriorityFor_FinalFallback(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		getPriorityFn: func(_ context.Context, _ uint64, _ string) (int, error) {
			return 0, apperrors.WrapNotFound("lstep_trigger_priority", "0")
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	// "unknown_trigger" は DefaultTriggerPriorities に存在しない
	p, err := svc.GetPriorityFor(context.Background(), 1, "unknown_trigger")
	assert.NoError(t, err)
	assert.Equal(t, model.DefaultPriorityFallback, p)
}

func TestGetPriorityFor_RepoError(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		getPriorityFn: func(_ context.Context, _ uint64, _ string) (int, error) {
			return 0, errors.New("db error")
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	p, err := svc.GetPriorityFor(context.Background(), 1, model.TriggerTypeDormantPrevention365)
	assert.Error(t, err)
	assert.Equal(t, 0, p)
}

func TestGetByClinicID_RepoError(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		findByClinicIDFn: func(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	items, err := svc.GetByClinicID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, items)
}

func TestGetByClinicID_ComplementsAllTypes(t *testing.T) {
	// DB に 1 件だけ登録されている状態
	repo := &mockTriggerPriorityRepo{
		findByClinicIDFn: func(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
			return []model.LstepTriggerPriority{
				{ClinicID: 1, TriggerType: model.TriggerTypeDormantPrevention365, Priority: 99},
			}, nil
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	items, err := svc.GetByClinicID(context.Background(), 1)
	assert.NoError(t, err)
	// 全 AllTriggerTypes() 分補完される
	assert.Len(t, items, len(model.AllTriggerTypes()))

	// DB 登録分は DB の値を優先
	for _, it := range items {
		if it.TriggerType == model.TriggerTypeDormantPrevention365 {
			assert.Equal(t, 99, it.Priority)
		}
	}
}

func TestUpdatePriorities_UnknownTriggerType(t *testing.T) {
	svc := NewLstepTriggerPriorityService(&mockTriggerPriorityRepo{})
	err := svc.UpdatePriorities(context.Background(), 1, UpdateTriggerPrioritiesInput{
		Items: []TriggerPriorityItem{
			// typo vs vaccine_deadline_60d — must reject (LSA-13)
			{TriggerType: "vaccine_deadline_60", Priority: 1},
		},
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestUpdatePriorities_InvalidPriority(t *testing.T) {
	svc := NewLstepTriggerPriorityService(&mockTriggerPriorityRepo{})
	err := svc.UpdatePriorities(context.Background(), 1, UpdateTriggerPrioritiesInput{
		Items: []TriggerPriorityItem{
			{TriggerType: model.TriggerTypeDormantPrevention365, Priority: 0},
		},
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestUpdatePriorities_RepoError(t *testing.T) {
	repo := &mockTriggerPriorityRepo{
		upsertBatchFn: func(_ context.Context, _ uint64, _ []model.LstepTriggerPriority) error {
			return errors.New("db error")
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	err := svc.UpdatePriorities(context.Background(), 1, UpdateTriggerPrioritiesInput{
		Items: []TriggerPriorityItem{
			{TriggerType: model.TriggerTypeDormantPrevention365, Priority: 1},
		},
	})
	assert.Error(t, err)
}

func TestUpdatePriorities_OK(t *testing.T) {
	var capturedItems []model.LstepTriggerPriority
	repo := &mockTriggerPriorityRepo{
		upsertBatchFn: func(_ context.Context, _ uint64, items []model.LstepTriggerPriority) error {
			capturedItems = items
			return nil
		},
	}
	svc := NewLstepTriggerPriorityService(repo)
	err := svc.UpdatePriorities(context.Background(), 1, UpdateTriggerPrioritiesInput{
		Items: []TriggerPriorityItem{
			{TriggerType: model.TriggerTypeDormantPrevention365, Priority: 2},
			{TriggerType: model.TriggerTypeCheckupFollowUp, Priority: 3},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, capturedItems, 2)
	assert.Equal(t, uint64(1), capturedItems[0].ClinicID)
	assert.Equal(t, 2, capturedItems[0].Priority)
}
