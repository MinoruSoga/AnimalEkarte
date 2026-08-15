package lstep

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestCalculateCPMStage_DefaultThresholds pins the legacy V1 classification contract at
// the package boundary before the tag-sync implementation moves into internal/lstep.
func TestCalculateCPMStage_DefaultThresholds(t *testing.T) {
	tests := []struct {
		name string
		data CPMData
		want CPMStage
	}{
		{name: "no visits", data: CPMData{DaysSinceVisit: -1, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageDormant},
		{name: "dormant boundary", data: CPMData{TotalVisitCount: 4, DaysSinceVisit: 240, FirstVisitDaysSince: 500, LTVAmount: 40_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageDormant},
		{name: "noah", data: CPMData{TotalVisitCount: 3, AnnualVisitCount: 3, DaysSinceVisit: 30, FirstVisitDaysSince: 365, LTVAmount: 80_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageNoah},
		{name: "core", data: CPMData{TotalVisitCount: 2, AnnualVisitCount: 2, DaysSinceVisit: 30, FirstVisitDaysSince: 180, LTVAmount: 50_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageCore},
		{name: "spot", data: CPMData{TotalVisitCount: 1, DaysSinceVisit: 100, FirstVisitDaysSince: 200, MaxSingleVisitAmount: 30_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageSpot},
		{name: "growing", data: CPMData{TotalVisitCount: 2, DaysSinceVisit: 10, FirstVisitDaysSince: 45, LTVAmount: 25_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageGrowing},
		{name: "encounter", data: CPMData{TotalVisitCount: 1, DaysSinceVisit: 5, FirstVisitDaysSince: 5, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageEncounter},
		{name: "encounter upper LTV boundary", data: CPMData{TotalVisitCount: 1, AnnualVisitCount: 1, DaysSinceVisit: 10, FirstVisitDaysSince: 10, LTVAmount: 19_999, MaxSingleVisitAmount: 19_999, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageEncounter},
		{name: "single visit at LTV break is unclassified", data: CPMData{TotalVisitCount: 1, AnnualVisitCount: 1, DaysSinceVisit: 10, FirstVisitDaysSince: 10, LTVAmount: 20_000, MaxSingleVisitAmount: 20_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageUnclassified},
		{name: "unclassified", data: CPMData{TotalVisitCount: 4, AnnualVisitCount: 2, DaysSinceVisit: 10, FirstVisitDaysSince: 120, LTVAmount: 15_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageUnclassified},
		{name: "zero visits but not dormant is unclassified", data: CPMData{DaysSinceVisit: 50, FirstVisitDaysSince: 50, LTVAmount: 100_000, Thresholds: model.CPMV1Thresholds{}}, want: CPMStageUnclassified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CalculateCPMStage(tt.data))
		})
	}
}

func TestCalculateCPMStage_CustomThresholds(t *testing.T) {
	thresholds := model.CPMV1Thresholds{
		DormantDays:      30,
		NoahDays:         60,
		NoahAnnualVisits: 2,
		NoahLTV:          10_000,
		CoreDays:         45,
		CoreAnnualVisits: 1,
		CoreLTV:          5_000,
		SpotMinAmount:    4_000,
		SpotInactiveDays: 10,
		GrowingMaxDays:   20,
		GrowingMinVisits: 2,
		GrowingMaxVisits: 4,
		LTVBreakLow:      1_000,
	}

	got := CalculateCPMStage(CPMData{
		TotalVisitCount:     2,
		AnnualVisitCount:    2,
		DaysSinceVisit:      5,
		FirstVisitDaysSince: 60,
		LTVAmount:           10_000,
		Thresholds:          thresholds,
	})

	assert.Equal(t, CPMStageNoah, got)
}
