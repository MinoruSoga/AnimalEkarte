package billing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type estimateUpdateRaceRepository struct {
	EstimateRepository
	beforeLock  func()
	afterLock   func()
	afterStale  func()
	beforeWrite func()
}

func (r *estimateUpdateRaceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	estimate, err := r.EstimateRepository.FindByID(ctx, clinicID, id)
	if err == nil && persistence.TxFromContext(ctx) == nil && r.afterStale != nil {
		r.afterStale()
	}
	return estimate, err
}

func (r *estimateUpdateRaceRepository) LockEditableByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	if r.beforeLock != nil {
		r.beforeLock()
	}
	estimate, err := r.EstimateRepository.LockEditableByID(ctx, clinicID, id)
	if err == nil && r.afterLock != nil {
		r.afterLock()
	}
	return estimate, err
}

func (r *estimateUpdateRaceRepository) UpdateIfNotLocked(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Estimate, error) {
	if r.beforeWrite != nil {
		r.beforeWrite()
	}
	return r.EstimateRepository.UpdateIfNotLocked(ctx, clinicID, id, fields)
}

func TestEstimateService_Update_ConcurrentHeaderAndItemReplacementStayConsistent(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	baseRepo := NewEstimateRepository(db)
	tx := testNewTransactor(db)
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "estimate update race owner")
	estimate := makeEstimate(t, db, clinicID, owner.ID, model.EstimateStatusDraft)
	initialItem := makeEstimateItem(t, db, estimate.ID, "initial item")
	require.NoError(t, db.Model(initialItem).Updates(map[string]any{
		"unit_price": 1000,
		"quantity":   1,
		"tax_type":   model.TaxTypeExcluded,
		"tax_rate":   0.10,
	}).Error)
	require.NoError(t, db.Model(estimate).Updates(map[string]any{
		"subtotal":     1000,
		"tax_total":    100,
		"total_amount": 1100,
	}).Error)

	staleRead := make(chan struct{})
	headerLocked := make(chan struct{})
	replacementAttempted := make(chan struct{})
	releaseHeader := make(chan struct{})
	var staleOnce, lockedOnce, attemptedOnce, releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseHeader) }) }
	t.Cleanup(release)

	headerRepo := &estimateUpdateRaceRepository{
		EstimateRepository: baseRepo,
		afterStale: func() {
			staleOnce.Do(func() { close(staleRead) })
			<-releaseHeader
		},
		afterLock: func() {
			lockedOnce.Do(func() { close(headerLocked) })
			<-releaseHeader
		},
	}
	replacementRepo := &estimateUpdateRaceRepository{
		EstimateRepository: baseRepo,
		beforeLock: func() {
			attemptedOnce.Do(func() { close(replacementAttempted) })
		},
		beforeWrite: func() {
			attemptedOnce.Do(func() { close(replacementAttempted) })
		},
	}

	headerSvc := NewEstimateService(headerRepo, nil, nil, nil, nil, tx)
	replacementSvc := NewEstimateService(replacementRepo, nil, nil, nil, nil, tx)
	headerTitle := "concurrent header update"
	replacementItems := []EstimateItemInput{{
		Name:      "replacement item",
		Category:  model.ItemCategoryOther,
		UnitPrice: 2000,
		Quantity:  2,
	}}

	headerErr := make(chan error, 1)
	go func() {
		_, err := headerSvc.Update(context.Background(), clinicID, estimate.ID, &UpdateEstimateInput{Title: &headerTitle})
		headerErr <- err
	}()

	var oldStaleRead bool
	select {
	case <-staleRead:
		oldStaleRead = true
	case <-headerLocked:
	case <-time.After(5 * time.Second):
		t.Fatal("header update neither read outside a transaction nor locked the estimate")
	}

	replacementErr := make(chan error, 1)
	go func() {
		_, err := replacementSvc.Update(context.Background(), clinicID, estimate.ID, &UpdateEstimateInput{Items: &replacementItems})
		replacementErr <- err
	}()

	if oldStaleRead {
		require.NoError(t, <-replacementErr)
		release()
		require.NoError(t, <-headerErr)
	} else {
		select {
		case <-replacementAttempted:
		case <-time.After(5 * time.Second):
			t.Fatal("item replacement did not attempt to acquire the estimate lock")
		}
		release()
		require.NoError(t, <-headerErr)
		require.NoError(t, <-replacementErr)
	}

	got, err := baseRepo.FindByID(context.Background(), clinicID, estimate.ID)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "replacement item", got.Items[0].Name)
	assert.Equal(t, int64(4000), got.Subtotal)
	assert.Equal(t, int64(400), got.TaxTotal)
	assert.Equal(t, int64(4400), got.TotalAmount)
}
