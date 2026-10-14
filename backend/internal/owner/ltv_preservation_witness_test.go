package owner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// TestEMR330_LTVRevenueAndVisitsSurviveDeath は EMR-330 の保存契約の証人:
// ペットが死亡しても LTV・売上・来院履歴の集計事実は変わらない。
// 死亡 sentinel は status='deceased' / deceased_at 非 NULL のどちら片方でも
// 成立し、全ペット死亡でも結果を消してはならない。
func TestEMR330_LTVRevenueAndVisitsSurviveDeath(t *testing.T) {
	cases := []struct {
		name    string
		status  model.PetStatus
		dated   bool
		allPets bool
	}{
		{"deceased status only", model.PetStatusDeceased, false, false},
		{"deceased timestamp only", model.PetStatusAlive, true, false},
		{"mixed live and dead pets", model.PetStatusDeceased, true, false},
		{"all pets deceased", model.PetStatusDeceased, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupLTVTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(73330)
			o := testdb.MakeTestOwner(t, db, clinicID, "EMR330 synthetic owner")
			species := &model.AnimalSpecies{Name: fmt.Sprintf("EMR330 synthetic species %d", time.Now().UnixNano())}
			require.NoError(t, db.WithContext(ctx).Create(species).Error)
			var pets []*model.Pet
			now := time.Now()
			for i, amount := range []int64{1000, 9000} {
				p := &model.Pet{ClinicID: clinicID, OwnerID: o.ID, AnimalSpeciesID: species.ID, Name: fmt.Sprintf("Synthetic pet %d", i), Status: model.PetStatusAlive}
				require.NoError(t, db.WithContext(ctx).Create(p).Error)
				pets = append(pets, p)
				record := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &o.ID, PetID: &p.ID, RecordNo: fmt.Sprintf("EMR330-LTV-%d-%s", i, tc.name), Date: now.AddDate(0, 0, -(i+1)*10)}
				require.NoError(t, db.WithContext(ctx).Create(record).Error)
				bill := &model.Billing{ClinicID: clinicID, OwnerID: &o.ID, MedicalRecordID: &record.ID, TotalAmount: amount, Status: model.BillingStatusCompleted}
				require.NoError(t, db.WithContext(ctx).Create(bill).Error)
			}
			repo := NewLtvRepository(db)
			params := &FindOwnerLTVParams{ClinicID: clinicID, IncludeZero: true, IncludeNoVisit: true, PeriodPreset: "last_12_months"}
			beforeRows, err := repo.FindOwnerLTV(ctx, params)
			require.NoError(t, err)
			require.Len(t, beforeRows, 1)
			before := beforeRows[0]
			require.Equal(t, int64(10000), before.TotalAmount)
			require.Equal(t, int64(2), before.TotalVisitCount)
			require.Equal(t, int64(9000), before.MaxSingleVisitAmount)
			require.NotNil(t, before.FirstVisitDate)
			require.NotNil(t, before.LastVisitDate)
			targetPets := pets[:1]
			if tc.allPets {
				targetPets = pets
			}
			for _, p := range targetPets {
				var deceasedAt *time.Time
				if tc.dated {
					deceasedAt = &now
				}
				patch := &model.Pet{Status: tc.status, DeceasedAt: deceasedAt}
				require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).Where("clinic_id = ? AND id = ?", clinicID, p.ID).Select("status", "deceased_at").Updates(patch).Error)
			}
			afterRows, err := repo.FindOwnerLTV(ctx, params)
			require.NoError(t, err)
			require.Len(t, afterRows, 1)
			after := afterRows[0]
			assert.Equal(t, before.TotalAmount, after.TotalAmount)
			assert.Equal(t, before.TotalVisitCount, after.TotalVisitCount)
			assert.Equal(t, before.AnnualVisitCount, after.AnnualVisitCount)
			assert.Equal(t, before.MaxSingleVisitAmount, after.MaxSingleVisitAmount)
			assert.Equal(t, before.AnnualAmount, after.AnnualAmount)
			assert.Equal(t, before.BillingCount, after.BillingCount)
			assert.Equal(t, before.PeriodVisitCount, after.PeriodVisitCount)
			assert.Equal(t, before.FirstVisitDate, after.FirstVisitDate)
			assert.Equal(t, before.LastVisitDate, after.LastVisitDate)
		})
	}
}
