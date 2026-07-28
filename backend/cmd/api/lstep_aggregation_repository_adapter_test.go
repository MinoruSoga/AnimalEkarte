package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/owner"
)

type aggregationAdapterRepository struct {
	params *owner.FindOwnerLTVParams
	rows   []owner.OwnerLTVRow
	err    error
}

func (r *aggregationAdapterRepository) FindOwnerLTV(_ context.Context, params *owner.FindOwnerLTVParams) ([]owner.OwnerLTVRow, error) {
	r.params = params
	return r.rows, r.err
}

func TestLstepAggregationRepositoryAdapter_MapsContract(t *testing.T) {
	lineUserID := "U-adapter"
	annualAmount := int64(1200)
	billingCount := int64(3)
	periodVisitCount := int64(2)
	daysSinceLastVisit := 10
	lastVisitBucket := "within_3m"
	lastVisitDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	firstVisitDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &aggregationAdapterRepository{rows: []owner.OwnerLTVRow{{
		OwnerID:              9,
		OwnerName:            "owner",
		LineUserID:           &lineUserID,
		LstepOptOut:          true,
		TotalAmount:          5000,
		TotalVisitCount:      4,
		AnnualVisitCount:     3,
		LastVisitDate:        &lastVisitDate,
		FirstVisitDate:       &firstVisitDate,
		AnnualAmount:         &annualAmount,
		BillingCount:         &billingCount,
		PeriodVisitCount:     &periodVisitCount,
		DaysSinceLastVisit:   &daysSinceLastVisit,
		LastVisitBucket:      &lastVisitBucket,
		MaxSingleVisitAmount: 3000,
	}}}
	adapter := lstepAggregationRepositoryAdapter{inner: repo}

	from := "2026-01-01"
	to := "2026-12-31"
	rows, err := adapter.FindOwnerLTV(context.Background(), &lstep.FindOwnerLTVParams{
		ClinicID: 7, Sort: "total_amount", From: &from, To: &to, IncludeZero: true, Order: "desc",
	})

	require.NoError(t, err)
	require.NotNil(t, repo.params)
	assert.Equal(t, uint64(7), repo.params.ClinicID)
	assert.Equal(t, "total_amount", repo.params.Sort)
	assert.Equal(t, &from, repo.params.From)
	assert.Equal(t, &to, repo.params.To)
	assert.True(t, repo.params.IncludeZero)
	assert.Equal(t, "desc", repo.params.Order)
	require.Len(t, rows, 1)
	assert.Equal(t, lstep.OwnerLTVRow{
		OwnerID:              9,
		OwnerName:            "owner",
		LineUserID:           &lineUserID,
		LstepOptOut:          true,
		TotalAmount:          5000,
		TotalVisitCount:      4,
		AnnualVisitCount:     3,
		LastVisitDate:        &lastVisitDate,
		FirstVisitDate:       &firstVisitDate,
		AnnualAmount:         &annualAmount,
		BillingCount:         &billingCount,
		PeriodVisitCount:     &periodVisitCount,
		DaysSinceLastVisit:   &daysSinceLastVisit,
		LastVisitBucket:      &lastVisitBucket,
		MaxSingleVisitAmount: 3000,
	}, rows[0])
}
