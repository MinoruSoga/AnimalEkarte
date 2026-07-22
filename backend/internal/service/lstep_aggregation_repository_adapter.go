package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/repository"
)

// lstepAggregationRepositoryAdapter isolates the residual owner LTV
// persistence DTOs from the lstep domain contract until that repository is
// migrated in its owning domain batch.
type lstepAggregationRepositoryAdapter struct {
	inner repository.LtvRepository
}

func (a lstepAggregationRepositoryAdapter) FindOwnerLTV(ctx context.Context, params *lstep.FindOwnerLTVParams) ([]lstep.OwnerLTVRow, error) {
	rows, err := a.inner.FindOwnerLTV(ctx, &repository.FindOwnerLTVParams{
		ClinicID:        params.ClinicID,
		Sort:            params.Sort,
		MinTotalAmount:  params.MinTotalAmount,
		MaxTotalAmount:  params.MaxTotalAmount,
		MinVisitCount:   params.MinVisitCount,
		LineLinked:      params.LineLinked,
		Year:            params.Year,
		From:            params.From,
		To:              params.To,
		AmountBasis:     params.AmountBasis,
		IncludeZero:     params.IncludeZero,
		Search:          params.Search,
		PeriodPreset:    params.PeriodPreset,
		MaxVisitCount:   params.MaxVisitCount,
		LastVisitBucket: params.LastVisitBucket,
		IncludeNoVisit:  params.IncludeNoVisit,
		Order:           params.Order,
	})
	if err != nil {
		return nil, err
	}

	result := make([]lstep.OwnerLTVRow, len(rows))
	for i := range rows {
		row := &rows[i]
		result[i] = lstep.OwnerLTVRow{
			OwnerID:              row.OwnerID,
			OwnerName:            row.OwnerName,
			LineUserID:           row.LineUserID,
			LstepOptOut:          row.LstepOptOut,
			TotalAmount:          row.TotalAmount,
			TotalVisitCount:      row.TotalVisitCount,
			AnnualVisitCount:     row.AnnualVisitCount,
			LastVisitDate:        row.LastVisitDate,
			FirstVisitDate:       row.FirstVisitDate,
			AnnualAmount:         row.AnnualAmount,
			BillingCount:         row.BillingCount,
			PeriodVisitCount:     row.PeriodVisitCount,
			DaysSinceLastVisit:   row.DaysSinceLastVisit,
			LastVisitBucket:      row.LastVisitBucket,
			MaxSingleVisitAmount: row.MaxSingleVisitAmount,
		}
	}
	return result, nil
}
