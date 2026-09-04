package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestAccountingRepository_FindAll_OwnerNameIdeographicSpaceFourWay(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "DB fullwidth × query fullwidth", storedName: "会計全角全角姓　会計全角全角名", query: "会計全角全角姓　会計全角全角名"},
		{name: "DB fullwidth × query halfwidth", storedName: "会計全角半角姓　会計全角半角名", query: "会計全角半角姓 会計全角半角名"},
		{name: "DB halfwidth × query fullwidth", storedName: "会計半角全角姓 会計半角全角名", query: "会計半角全角姓　会計半角全角名"},
		{name: "DB halfwidth × query halfwidth", storedName: "会計半角半角姓 会計半角半角名", query: "会計半角半角姓 会計半角半角名"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := testdb.MakeTestOwner(t, db, clinicA, tt.storedName)
			foreign := testdb.MakeTestOwner(t, db, clinicB, tt.storedName)
			billing := makeBillingWith(t, db, billingFixtureOpts{
				ClinicID:      clinicA,
				OwnerID:       &owner.ID,
				TotalAmount:   1000,
				Status:        model.BillingStatusWaiting,
				ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			})
			_ = makeBillingWith(t, db, billingFixtureOpts{
				ClinicID:      clinicB,
				OwnerID:       &foreign.ID,
				TotalAmount:   1000,
				Status:        model.BillingStatusWaiting,
				ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			})

			got, total, err := repo.FindAll(ctx, clinicA, AccountingListFilters{Search: tt.query}, 1, 100)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, billing.ID, got[0].ID)
		})
	}
}

func TestAccountingRepository_FindAll_SpaceOnlySearchReturnsNoRows(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "会計空白のみ姓 会計空白のみ名")
	_ = makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicA,
		OwnerID:       &owner.ID,
		TotalAmount:   1000,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})

	got, total, err := repo.FindAll(ctx, clinicA, AccountingListFilters{Search: "　　"}, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, got)
}
