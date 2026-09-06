package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeTenantRelationBilling(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	ownerID, petID *uint64,
	status model.BillingStatus,
	amount int64,
	completedAt *time.Time,
) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		OwnerID:       ownerID,
		PetID:         petID,
		TotalAmount:   amount,
		Status:        status,
		ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		CompletedAt:   completedAt,
	})
}

func makeTenantRelationSplit(
	t *testing.T,
	db *gorm.DB,
	clinicID, billingID uint64,
	method model.PaymentMethod,
	amount int64,
) model.PaymentSplit {
	t.Helper()
	split := model.PaymentSplit{
		ClinicID:  clinicID,
		BillingID: billingID,
		Method:    method,
		Amount:    amount,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&split).Error)
	return split
}

func makeTenantRelationRefund(
	t *testing.T,
	db *gorm.DB,
	clinicID, billingID uint64,
	amount int64,
) model.BillingRefund {
	t.Helper()
	refund := model.BillingRefund{
		ClinicID:  clinicID,
		BillingID: billingID,
		Amount:    amount,
		RefundedAt: time.Date(
			2026,
			6,
			11,
			0,
			0,
			0,
			0,
			time.UTC,
		),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&refund).Error)
	return refund
}

func makeTenantRelationStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
	active bool,
) *model.Staff {
	t.Helper()
	staff := makeDoctor(t, db, clinicID, name)
	require.NoError(t, db.Model(&model.Staff{}).
		Where("id = ?", staff.ID).
		Update("is_active", active).Error)
	staff.IsActive = active
	assignTenantRelationStaff(t, db, staff.ID, clinicID)
	return staff
}

func assignTenantRelationStaff(
	t *testing.T,
	db *gorm.DB,
	staffID, clinicID uint64,
) {
	t.Helper()
	seedBillingClinicForFK(t, db, clinicID)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staffID,
		ClinicID: clinicID,
	}).Error)
}

func assertBillingTenantChildren(
	t *testing.T,
	billing *model.Billing,
	clinicID uint64,
	wantSplitID, wantRefundID uint64,
	wantRefundTotal int64,
) {
	t.Helper()
	require.NotNil(t, billing)
	require.Len(t, billing.PaymentSplits, 1)
	assert.Equal(t, wantSplitID, billing.PaymentSplits[0].ID)
	assert.Equal(t, clinicID, billing.PaymentSplits[0].ClinicID)
	require.Len(t, billing.Refunds, 1)
	assert.Equal(t, wantRefundID, billing.Refunds[0].ID)
	assert.Equal(t, clinicID, billing.Refunds[0].ClinicID)
	assert.Equal(t, wantRefundTotal, billing.TotalRefundedAmount)
}

func TestAccountingRepository_ReadsExcludeForeignSplitAndRefundRows(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	billing := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		nil,
		nil,
		model.BillingStatusWaiting,
		1000,
		nil,
	)
	validSplit := makeTenantRelationSplit(
		t,
		db,
		clinicA,
		billing.ID,
		model.PaymentMethodCash,
		1000,
	)
	makeTenantRelationSplit(
		t,
		db,
		clinicB,
		billing.ID,
		model.PaymentMethodCreditCard,
		9000,
	)
	validRefund := makeTenantRelationRefund(t, db, clinicA, billing.ID, 100)
	makeTenantRelationRefund(t, db, clinicB, billing.ID, 900)

	t.Run("FindByID", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, billing.ID)
		require.NoError(t, err)
		assertBillingTenantChildren(
			t,
			got,
			clinicA,
			validSplit.ID,
			validRefund.ID,
			100,
		)
	})

	t.Run("FindByIDForClinics still requires child parent clinic equality", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			billing.ID,
		)
		require.NoError(t, err)
		assertBillingTenantChildren(
			t,
			got,
			clinicA,
			validSplit.ID,
			validRefund.ID,
			100,
		)
	})

	t.Run("FindAll", func(t *testing.T) {
		got, total, err := repo.FindAll(
			ctx,
			clinicA,
			AccountingListFilters{},
			1,
			100,
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		require.Len(t, got[0].PaymentSplits, 1)
		assert.Equal(t, validSplit.ID, got[0].PaymentSplits[0].ID)
		assert.Empty(t, got[0].Refunds, "list uses attachRefundTotals instead of Refunds preload")
		assert.Equal(t, int64(100), got[0].TotalRefundedAmount)
	})

	t.Run("FindAllForClinics still requires child parent clinic equality", func(t *testing.T) {
		got, total, err := repo.FindAllForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			AccountingListFilters{},
			1,
			100,
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		require.Len(t, got[0].PaymentSplits, 1)
		assert.Equal(t, validSplit.ID, got[0].PaymentSplits[0].ID)
		assert.Equal(t, int64(100), got[0].TotalRefundedAmount)
	})

	t.Run("Update refetch", func(t *testing.T) {
		got, err := repo.Update(
			ctx,
			clinicA,
			billing.ID,
			AccountingUpdate{Memo: strPtr("tenant child scope")},
		)
		require.NoError(t, err)
		assertBillingTenantChildren(
			t,
			got,
			clinicA,
			validSplit.ID,
			validRefund.ID,
			100,
		)
	})

	t.Run("LockAndFindByID", func(t *testing.T) {
		require.NoError(t, testNewTransactor(db).WithTx(
			ctx,
			func(txCtx context.Context) error {
				got, err := repo.LockAndFindByID(txCtx, clinicA, billing.ID)
				if err != nil {
					return err
				}
				assertBillingTenantChildren(
					t,
					got,
					clinicA,
					validSplit.ID,
					validRefund.ID,
					100,
				)
				return nil
			},
		))
	})
}

func TestAccountingRepository_SavePaymentSplitsRejectsForeignOrMixedRows(
	t *testing.T,
) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	billingA := makeBillingRet(t, db, clinicA)
	billingA2 := makeBillingRet(t, db, clinicA)
	original := makeTenantRelationSplit(
		t,
		db,
		clinicA,
		billingA.ID,
		model.PaymentMethodCash,
		500,
	)

	assertOriginalPreserved := func(t *testing.T) {
		t.Helper()
		var splits []model.PaymentSplit
		require.NoError(t, db.WithContext(ctx).
			Where("billing_id = ?", billingA.ID).
			Order("id ASC").
			Find(&splits).Error)
		require.Len(t, splits, 1)
		assert.Equal(t, original.ID, splits[0].ID)
		assert.Equal(t, clinicA, splits[0].ClinicID)
		assert.Equal(t, int64(500), splits[0].Amount)
	}

	t.Run("foreign clinic cannot target another clinic billing", func(t *testing.T) {
		err := repo.SavePaymentSplits(ctx, []model.PaymentSplit{{
			ClinicID:  clinicB,
			BillingID: billingA.ID,
			Method:    model.PaymentMethodCreditCard,
			Amount:    9000,
		}})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)
		assertOriginalPreserved(t)
	})

	t.Run("every split must use the same clinic", func(t *testing.T) {
		err := repo.SavePaymentSplits(ctx, []model.PaymentSplit{
			{
				ClinicID:  clinicA,
				BillingID: billingA.ID,
				Method:    model.PaymentMethodCash,
				Amount:    200,
			},
			{
				ClinicID:  clinicB,
				BillingID: billingA.ID,
				Method:    model.PaymentMethodCreditCard,
				Amount:    300,
			},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "got %v", err)
		assertOriginalPreserved(t)
	})

	t.Run("every split must use the same billing", func(t *testing.T) {
		err := repo.SavePaymentSplits(ctx, []model.PaymentSplit{
			{
				ClinicID:  clinicA,
				BillingID: billingA.ID,
				Method:    model.PaymentMethodCash,
				Amount:    200,
			},
			{
				ClinicID:  clinicA,
				BillingID: billingA2.ID,
				Method:    model.PaymentMethodCreditCard,
				Amount:    300,
			},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "got %v", err)
		assertOriginalPreserved(t)
	})

	t.Run("valid replacement is preserved", func(t *testing.T) {
		err := repo.SavePaymentSplits(ctx, []model.PaymentSplit{
			{
				ClinicID:  clinicA,
				BillingID: billingA.ID,
				Method:    model.PaymentMethodCash,
				Amount:    200,
			},
			{
				ClinicID:  clinicA,
				BillingID: billingA.ID,
				Method:    model.PaymentMethodCreditCard,
				Amount:    300,
			},
		})
		require.NoError(t, err)

		var splits []model.PaymentSplit
		require.NoError(t, db.WithContext(ctx).
			Where("billing_id = ?", billingA.ID).
			Order("amount ASC").
			Find(&splits).Error)
		require.Len(t, splits, 2)
		assert.Equal(t, int64(200), splits[0].Amount)
		assert.Equal(t, int64(300), splits[1].Amount)
		for i := range splits {
			assert.Equal(t, clinicA, splits[i].ClinicID)
			assert.Equal(t, billingA.ID, splits[i].BillingID)
		}
	})
}

func TestAccountingRepository_ReportsExcludeForeignSplitAndRefundRows(
	t *testing.T,
) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	billing := makeMonthlyBilling(
		t,
		db,
		clinicA,
		midJune(10),
		model.ItemCategoryExamination,
		1000,
		0.10,
		model.PaymentMethodCash,
		1000,
	)
	makeTenantRelationSplit(
		t,
		db,
		clinicB,
		billing.ID,
		model.PaymentMethodCreditCard,
		9000,
	)
	makeTenantRelationRefund(t, db, clinicA, billing.ID, 100)
	makeTenantRelationRefund(t, db, clinicB, billing.ID, 900)

	t.Run("daily", func(t *testing.T) {
		got, err := repo.GetDailySummary(ctx, clinicA, midJune(10))
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.BillingCount)
		assert.Equal(t, int64(1000), got.GrandTotal)
		require.Len(t, got.PaymentTotals, 1)
		assert.Equal(t, string(model.PaymentMethodCash), got.PaymentTotals[0].Method)
		assert.Equal(t, int64(1000), got.PaymentTotals[0].Total)
	})

	t.Run("monthly", func(t *testing.T) {
		got, err := repo.GetMonthlyReport(ctx, clinicA, 2026, 6)
		require.NoError(t, err)
		var paymentTotal int64
		for _, row := range got.PaymentRows {
			paymentTotal += row.Amount
		}
		assert.Equal(t, int64(1000), paymentTotal)
		assert.Equal(t, int64(100), got.TotalRefund)
		assert.Equal(t, int64(900), got.GrandTotal)
	})

	t.Run("close", func(t *testing.T) {
		got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
			ClinicID: clinicA,
			PeriodStart: time.Date(
				2026,
				6,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			PeriodEnd: time.Date(
				2026,
				7,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
		})
		require.NoError(t, err)
		var paymentTotal int64
		for _, row := range got.PaymentRows {
			paymentTotal += row.Amount
		}
		assert.Equal(t, int64(1000), paymentTotal)
		assert.Equal(t, int64(100), got.TotalRefund)
		require.Len(t, got.BillingDetails, 1)
		assert.Equal(t, int64(1000), got.BillingDetails[0].BillingAmount)
		assert.Equal(t, int64(100), got.BillingDetails[0].RefundAmount)
		assert.Equal(t, int64(900), got.BillingDetails[0].NetAmount)
	})
}

func TestAccountingRepository_UnpaidAggregatesExcludeForeignOwnerPetRelations(
	t *testing.T,
) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "valid owner")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "foreign owner")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "valid pet")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "foreign pet")

	valid := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerA.ID,
		&petA.ID,
		model.BillingStatusWaiting,
		1000,
		nil,
	)
	makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerB.ID,
		&petB.ID,
		model.BillingStatusWaiting,
		9000,
		nil,
	)
	makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerA.ID,
		&petB.ID,
		model.BillingStatusWaiting,
		8000,
		nil,
	)

	t.Run("billing detail list", func(t *testing.T) {
		got, total, err := repo.FindUnpaidByBilling(
			ctx,
			clinicA,
			firstDay2026June,
			lastDay2026June,
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, valid.ID, got[0].ID)
		require.NotNil(t, got[0].Owner)
		require.NotNil(t, got[0].Pet)
		assert.Equal(t, ownerA.ID, got[0].Owner.ID)
		assert.Equal(t, petA.ID, got[0].Pet.ID)
	})

	t.Run("owner aggregate and summary", func(t *testing.T) {
		got, total, summary, err := repo.FindUnpaidByOwner(
			ctx,
			clinicA,
			firstDay2026June,
			lastDay2026June,
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int64(1000), summary.TotalAmount)
		assert.Equal(t, int64(1), summary.BillingCount)
		assert.Equal(t, int64(1), summary.OwnerCount)
		require.Len(t, got, 1)
		assert.Equal(t, ownerA.ID, got[0].OwnerID)
		assert.Equal(t, ownerA.Name, got[0].OwnerName)
		assert.Equal(t, int64(1000), got[0].TotalAmount)
	})

	t.Run("single owner balance", func(t *testing.T) {
		got, err := repo.SumUnpaidByOwner(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1000), got.TotalAmount)
		assert.Equal(t, int64(1), got.Count)

		foreign, err := repo.SumUnpaidByOwner(ctx, clinicA, ownerB.ID)
		require.NoError(t, err)
		assert.Equal(t, OwnerUnpaidBalance{}, foreign)
	})

	t.Run("monthly owner pet aggregate", func(t *testing.T) {
		got, total, summary, err := repo.FindMonthlyUnpaidCarryover(
			ctx,
			clinicA,
			firstDay2026June,
			lastDay2026June,
			1,
			100,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int64(1000), summary.CurrentMonthUnpaid)
		assert.Equal(t, int64(1000), summary.NextMonthCarryover)
		require.Len(t, got, 1)
		assert.Equal(t, ownerA.ID, got[0].OwnerID)
		require.NotNil(t, got[0].PetID)
		assert.Equal(t, petA.ID, *got[0].PetID)
		assert.Equal(t, petA.Name, got[0].PetName)
		assert.Equal(t, int64(1000), got[0].CurrentMonthUnpaid)
	})
}

func TestAccountingRepository_CloseDetailsDoNotExposeForeignOwnerPetNames(
	t *testing.T,
) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "valid close owner")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "foreign close owner")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "valid close pet")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "foreign close pet")
	completedAt := midJune(10)

	valid := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerA.ID,
		&petA.ID,
		model.BillingStatusCompleted,
		1000,
		&completedAt,
	)
	foreignOwnerPet := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerB.ID,
		&petB.ID,
		model.BillingStatusCompleted,
		2000,
		&completedAt,
	)
	foreignPet := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerA.ID,
		&petB.ID,
		model.BillingStatusCompleted,
		3000,
		&completedAt,
	)

	for _, billing := range []*model.Billing{valid, foreignOwnerPet, foreignPet} {
		item := model.BillingItem{
			BillingID: billing.ID,
			Category:  model.ItemCategoryExamination,
			Name:      "close detail",
			UnitPrice: billing.TotalAmount,
			Quantity:  1,
			TaxType:   model.TaxTypeExcluded,
			TaxRate:   0.10,
		}
		require.NoError(t, db.WithContext(ctx).Create(&item).Error)
		makeTenantRelationSplit(
			t,
			db,
			clinicA,
			billing.ID,
			model.PaymentMethodCash,
			billing.TotalAmount,
		)
	}

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID: clinicA,
		PeriodStart: time.Date(
			2026,
			6,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		PeriodEnd: time.Date(
			2026,
			7,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
	})
	require.NoError(t, err)

	byBillingID := make(map[uint64]CloseBillingDetailRow, len(got.BillingDetails))
	for _, row := range got.BillingDetails {
		byBillingID[row.BillingID] = row
	}
	require.Contains(t, byBillingID, valid.ID)
	assert.Equal(t, ownerA.Name, byBillingID[valid.ID].OwnerName)
	assert.Equal(t, petA.Name, byBillingID[valid.ID].PetName)

	// SEC-SWEEP-02-BILL-B1a-FIX: own-clinic billings stay in close aggregates;
	// foreign-clinic owner/pet names are blanked via clinic-correlated LEFT JOIN.
	// Amounts: valid 1000 + foreignOwnerPet 2000 + foreignPet 3000 = 6000.
	require.Contains(t, byBillingID, foreignOwnerPet.ID)
	assert.Empty(t, byBillingID[foreignOwnerPet.ID].OwnerName)
	assert.Empty(t, byBillingID[foreignOwnerPet.ID].PetName)

	require.Contains(t, byBillingID, foreignPet.ID)
	assert.Equal(t, ownerA.Name, byBillingID[foreignPet.ID].OwnerName)
	assert.Empty(t, byBillingID[foreignPet.ID].PetName)

	var paymentTotal int64
	for _, p := range got.PaymentRows {
		paymentTotal += p.Amount
	}
	assert.Equal(t, int64(6000), paymentTotal, "own-clinic billings with foreign pet_id remain in payment totals")
}

func TestAccountingRepository_NestedStaffPreloadsRequireExactBillingClinic(
	t *testing.T,
) {
	db := setupAccountingIsolationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Staff{}))
	repo := NewAccountingRepository(db)
	refundRepo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	validStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"valid secondary-clinic billing staff",
		true,
	)
	assignTenantRelationStaff(t, db, validStaff.ID, clinicA)
	foreignStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"foreign billing staff",
		true,
	)
	inactiveAssignedStaff := makeTenantRelationStaff(
		t,
		db,
		clinicA,
		"inactive historical billing staff",
		false,
	)
	validBilling := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		nil,
		nil,
		model.BillingStatusWaiting,
		1000,
		nil,
	)
	contaminatedBilling := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		nil,
		nil,
		model.BillingStatusWaiting,
		2000,
		nil,
	)
	inactiveStaffBilling := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		nil,
		nil,
		model.BillingStatusWaiting,
		3000,
		nil,
	)
	validSecondaryClinicBilling := makeTenantRelationBilling(
		t,
		db,
		clinicB,
		nil,
		nil,
		model.BillingStatusWaiting,
		4000,
		nil,
	)
	require.NoError(t, db.Create(&model.Payment{
		BillingID:     validBilling.ID,
		BillingAmount: validBilling.TotalAmount,
		PaidBy:        &validStaff.ID,
	}).Error)
	require.NoError(t, db.Create(&model.Payment{
		BillingID:     contaminatedBilling.ID,
		BillingAmount: contaminatedBilling.TotalAmount,
		PaidBy:        &foreignStaff.ID,
	}).Error)
	require.NoError(t, db.Create(&model.Payment{
		BillingID:     inactiveStaffBilling.ID,
		BillingAmount: inactiveStaffBilling.TotalAmount,
		PaidBy:        &inactiveAssignedStaff.ID,
	}).Error)
	require.NoError(t, db.Create(&model.Payment{
		BillingID:     validSecondaryClinicBilling.ID,
		BillingAmount: validSecondaryClinicBilling.TotalAmount,
		PaidBy:        &validStaff.ID,
	}).Error)
	validRefund := makeTenantRelationRefund(
		t,
		db,
		clinicA,
		validBilling.ID,
		100,
	)
	require.NoError(t, db.Model(&model.BillingRefund{}).
		Where("id = ?", validRefund.ID).
		Update("refunded_by", validStaff.ID).Error)
	foreignRefund := makeTenantRelationRefund(
		t,
		db,
		clinicA,
		contaminatedBilling.ID,
		200,
	)
	require.NoError(t, db.Model(&model.BillingRefund{}).
		Where("id = ?", foreignRefund.ID).
		Update("refunded_by", foreignStaff.ID).Error)

	t.Run("single-clinic detail loads only exact-clinic staff", func(t *testing.T) {
		valid, err := repo.FindByID(ctx, clinicA, validBilling.ID)
		require.NoError(t, err)
		require.Len(t, valid.Payments, 1)
		require.NotNil(t, valid.Payments[0].PaidByStaff)
		assert.Equal(t, validStaff.ID, valid.Payments[0].PaidByStaff.ID)
		require.Len(t, valid.Refunds, 1)
		require.NotNil(t, valid.Refunds[0].RefundedByStaff)
		assert.Equal(t, validStaff.ID, valid.Refunds[0].RefundedByStaff.ID)

		contaminated, err := repo.FindByID(
			ctx,
			clinicA,
			contaminatedBilling.ID,
		)
		require.NoError(t, err)
		require.Len(t, contaminated.Payments, 1)
		assert.Nil(t, contaminated.Payments[0].PaidByStaff)
		require.Len(t, contaminated.Refunds, 1)
		assert.Nil(t, contaminated.Refunds[0].RefundedByStaff)
	})

	t.Run("multi-clinic detail still enforces each billing parent", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			contaminatedBilling.ID,
		)
		require.NoError(t, err)
		require.Len(t, got.Payments, 1)
		assert.Nil(t, got.Payments[0].PaidByStaff)
		require.Len(t, got.Refunds, 1)
		assert.Nil(t, got.Refunds[0].RefundedByStaff)
	})

	t.Run("inactive assigned historical staff remains visible", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, inactiveStaffBilling.ID)
		require.NoError(t, err)
		require.Len(t, got.Payments, 1)
		require.NotNil(t, got.Payments[0].PaidByStaff)
		assert.Equal(t, inactiveAssignedStaff.ID, got.Payments[0].PaidByStaff.ID)
	})

	t.Run("multi-clinic list omits payment staff preload", func(t *testing.T) {
		items, _, err := repo.FindAllForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			AccountingListFilters{},
			1,
			100,
		)
		require.NoError(t, err)
		seen := make(map[uint64]*model.Billing, len(items))
		for i := range items {
			seen[items[i].ID] = &items[i]
		}
		for _, billingID := range []uint64{
			validBilling.ID,
			validSecondaryClinicBilling.ID,
		} {
			got := seen[billingID]
			require.NotNil(t, got)
			require.Len(t, got.Payments, 1)
			assert.Nil(t, got.Payments[0].PaidByStaff)
		}
	})

	t.Run("list and update refetch do not expose foreign payment staff", func(t *testing.T) {
		items, _, err := repo.FindAll(
			ctx,
			clinicA,
			AccountingListFilters{},
			1,
			100,
		)
		require.NoError(t, err)
		for i := range items {
			if items[i].ID == contaminatedBilling.ID {
				require.Len(t, items[i].Payments, 1)
				assert.Nil(t, items[i].Payments[0].PaidByStaff)
			}
		}

		got, err := repo.Update(
			ctx,
			clinicA,
			contaminatedBilling.ID,
			AccountingUpdate{Memo: strPtr("staff preload scope")},
		)
		require.NoError(t, err)
		require.Len(t, got.Payments, 1)
		assert.Nil(t, got.Payments[0].PaidByStaff)
		require.Len(t, got.Refunds, 1)
		assert.Nil(t, got.Refunds[0].RefundedByStaff)
	})

	t.Run("refund list does not expose foreign staff", func(t *testing.T) {
		refunds, err := refundRepo.FindByBillingID(
			ctx,
			clinicA,
			contaminatedBilling.ID,
		)
		require.NoError(t, err)
		require.Len(t, refunds, 1)
		assert.Nil(t, refunds[0].RefundedByStaff)
	})
}

func TestAccountingRepository_WriteActorsRequireActiveSameClinicStaff(
	t *testing.T,
) {
	db := setupAccountingIsolationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Staff{}))
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	billing := makeBillingRet(t, db, clinicA)
	validStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"valid secondary-clinic actor",
		true,
	)
	assignTenantRelationStaff(t, db, validStaff.ID, clinicA)
	inactiveStaff := makeTenantRelationStaff(
		t,
		db,
		clinicA,
		"inactive actor",
		false,
	)
	foreignStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"foreign actor",
		true,
	)
	unassignedStaff := makeDoctor(t, db, clinicA, "unassigned actor")

	for _, staff := range []*model.Staff{
		foreignStaff,
		inactiveStaff,
		unassignedStaff,
	} {
		err := repo.SavePayment(ctx, &model.Payment{
			BillingID:     billing.ID,
			BillingAmount: 9000,
			PaidBy:        &staff.ID,
		})
		require.Error(t, err)

		var count int64
		require.NoError(t, db.Model(&model.Payment{}).
			Where("billing_id = ?", billing.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	}

	require.NoError(t, repo.SavePayment(ctx, &model.Payment{
		BillingID:     billing.ID,
		BillingAmount: 1000,
		PaidBy:        &validStaff.ID,
	}))
	err := repo.SavePayment(ctx, &model.Payment{
		BillingID:     billing.ID,
		BillingAmount: 9000,
		PaidBy:        &foreignStaff.ID,
	})
	require.Error(t, err)
	var savedPayment model.Payment
	require.NoError(t, db.Where("billing_id = ?", billing.ID).
		First(&savedPayment).Error)
	assert.Equal(t, int64(1000), savedPayment.BillingAmount)
	require.NotNil(t, savedPayment.PaidBy)
	assert.Equal(t, validStaff.ID, *savedPayment.PaidBy)

	originalSplit := makeTenantRelationSplit(
		t,
		db,
		clinicA,
		billing.ID,
		model.PaymentMethodCash,
		1000,
	)
	for _, staff := range []*model.Staff{foreignStaff, inactiveStaff} {
		err := repo.SavePaymentSplits(ctx, []model.PaymentSplit{{
			ClinicID:  clinicA,
			BillingID: billing.ID,
			Method:    model.PaymentMethodCreditCard,
			Amount:    9000,
			PaidBy:    &staff.ID,
		}})
		require.Error(t, err)

		var splits []model.PaymentSplit
		require.NoError(t, db.Where("billing_id = ?", billing.ID).
			Find(&splits).Error)
		require.Len(t, splits, 1)
		assert.Equal(t, originalSplit.ID, splits[0].ID)
	}
	require.NoError(t, repo.SavePaymentSplits(ctx, []model.PaymentSplit{{
		ClinicID:  clinicA,
		BillingID: billing.ID,
		Method:    model.PaymentMethodCash,
		Amount:    1000,
		PaidBy:    &validStaff.ID,
	}}))
}

func TestRefundRepository_WriteActorRequiresActiveSameClinicStaff(
	t *testing.T,
) {
	db := setupRefundRepoTestDB(t)
	repo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	billing := makeBillingForRefund(t, db, clinicA)
	validStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"valid secondary-clinic refund actor",
		true,
	)
	assignTenantRelationStaff(t, db, validStaff.ID, clinicA)
	inactiveStaff := makeTenantRelationStaff(
		t,
		db,
		clinicA,
		"inactive refund actor",
		false,
	)
	foreignStaff := makeTenantRelationStaff(
		t,
		db,
		clinicB,
		"foreign refund actor",
		true,
	)
	unassignedStaff := makeDoctor(t, db, clinicA, "unassigned refund actor")

	for _, staff := range []*model.Staff{
		foreignStaff,
		inactiveStaff,
		unassignedStaff,
	} {
		err := repo.Create(ctx, &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     100,
			RefundedBy: &staff.ID,
			RefundedAt: time.Now(),
		})
		require.Error(t, err)

		var count int64
		require.NoError(t, db.Model(&model.BillingRefund{}).
			Where("billing_id = ?", billing.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	}

	require.NoError(t, repo.Create(ctx, &model.BillingRefund{
		ClinicID:   clinicA,
		BillingID:  billing.ID,
		Amount:     100,
		RefundedBy: &validStaff.ID,
		RefundedAt: time.Now(),
	}))
}

// TestBillingItemRepository_ValidateCreateReferences_CorrelatesAppointmentClinic
// SEC-SWEEP-02-BILL-B1a: trimming course 参照は appointments 親 clinic 相関付きで解決する。
func TestBillingItemRepository_ValidateCreateReferences_CorrelatesAppointmentClinic(t *testing.T) {
	db := setupBillingItemTrimmingTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "validate-course-owner-a")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "validate-course-pet-a")
	rtA := makeTrimmingReservationType(t, db, clinicA)
	apptA := makeTrimmingAppointment(t, db, clinicA, petA.ID, rtA.ID, model.ReservationStatusAccounting)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", apptA.ID).Update("owner_id", ownerA.ID).Error)
	billingA := &model.Billing{
		ClinicID: clinicA, OwnerID: &ownerA.ID, PetID: &petA.ID,
		Status: model.BillingStatusWaiting, ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.Create(billingA).Error)
	courseA := makeTrimmingCourse(t, db, clinicA, "course-a", priceOf(1000))
	attachTrimmingCourse(t, db, clinicA, apptA.ID, courseA.ID)

	t.Run("accepts same-clinic course with parent appointment correlation", func(t *testing.T) {
		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.ValidateCreateReferences(
				txCtx, clinicA, billingA.ID, nil, nil, &apptA.ID, &courseA.ID, nil,
			)
			return err
		})
		require.NoError(t, err)
	})

	t.Run("rejects cross-tenant trimming course detail", func(t *testing.T) {
		ownerB := testdb.MakeTestOwner(t, db, clinicB, "validate-course-owner-b")
		petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "validate-course-pet-b")
		rtB := makeTrimmingReservationType(t, db, clinicB)
		apptB := makeTrimmingAppointment(t, db, clinicB, petB.ID, rtB.ID, model.ReservationStatusAccounting)
		require.NoError(t, db.Where("appointment_id = ?", apptA.ID).Delete(&model.AppointmentTrimmingDetail{}).Error)
		corrupt := &model.AppointmentTrimmingDetail{
			ClinicID: clinicA, AppointmentID: apptB.ID, CourseID: &courseA.ID,
		}
		require.NoError(t, db.Create(corrupt).Error)
		billingB := &model.Billing{
			ClinicID: clinicB, OwnerID: &ownerB.ID, PetID: &petB.ID,
			Status: model.BillingStatusWaiting, ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, db.Create(billingB).Error)
		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.ValidateCreateReferences(
				txCtx, clinicB, billingB.ID, nil, nil, &apptB.ID, &courseA.ID, nil,
			)
			return err
		})
		require.Error(t, err, "cross-tenant trimming course detail must not validate")
	})
}
