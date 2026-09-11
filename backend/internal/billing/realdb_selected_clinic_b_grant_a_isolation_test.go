package billing

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 billing
// clinic-fixed 22 + cross-clinic 3 through real repo/service/HTTP.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBBillInsA    = "D3bill-realdb-insurance-A"
	realDBBillInsB    = "D3bill-realdb-insurance-B"
	realDBBillCampA   = "D3bill-realdb-campaign-A"
	realDBBillCampB   = "D3bill-realdb-campaign-B"
	realDBBillPayA    = "D3bill-realdb-pay-A"
	realDBBillPayB    = "D3bill-realdb-pay-B"
	realDBBillEstA    = "D3bill-realdb-estimate-A"
	realDBBillEstB    = "D3bill-realdb-estimate-B"
	realDBBillMemoA   = "D3bill-realdb-billing-memo-A"
	realDBBillMemoB   = "D3bill-realdb-billing-memo-B"
	realDBBillRefundB = "D3bill-realdb-refund-B"
	realDBBillCloseA  = "D3bill-realdb-close-A"
	realDBBillCloseB  = "D3bill-realdb-close-B"
	realDBBillOwnerA  = "D3bill-realdb-owner-A"
	realDBBillOwnerB  = "D3bill-realdb-owner-B"
	realDBBillPetA    = "D3bill-realdb-pet-A"
	realDBBillPetB    = "D3bill-realdb-pet-B"
	realDBBillItemB   = "D3bill-realdb-item-B"
	realDBBillTreatB  = "D3bill-realdb-unbilled-B"
	realDBBillMRA     = "D3bill-MR-A"
	realDBBillMRB     = "D3bill-MR-B"
	realDBBillDate    = "2026-09-15"
	realDBBillAmountB = int64(7777)
)

type nameClinicDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Name     string `json:"name"`
}
type estimateDTO struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}
type billingMemoDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Memo     string `json:"memo"`
}
type refundDTO struct {
	Reason string `json:"reason"`
}
type closeDTO struct {
	ID   uint64 `json:"id"`
	Memo string `json:"memo"`
}
type dailySummaryDTO struct {
	BillingCount int64 `json:"billing_count"`
	GrandTotal   int64 `json:"grand_total"`
}
type unpaidBalanceDTO struct {
	UnpaidTotal int64 `json:"unpaid_total"`
	UnpaidCount int64 `json:"unpaid_count"`
}
type discountDTO struct {
	Suggestions []struct {
		Name string `json:"name"`
	} `json:"suggestions"`
}
type unbilledItemDTO struct {
	Name string `json:"name"`
}
type ungroupedDTO struct {
	MedicalRecordCount int64 `json:"medical_record_count"`
	HasUngrouped       bool  `json:"has_ungrouped"`
}
type confirmationDTO struct {
	MedicalRecordID string `json:"medical_record_id"`
	Status          string `json:"status"`
}
type monthlyReportDTO struct {
	Summary struct {
		TotalAmount int64 `json:"total_amount"`
	} `json:"summary"`
}

type clinicReaderStub struct{}

func (clinicReaderStub) FindByID(_ context.Context, id uint64) (*model.Clinic, error) {
	return &model.Clinic{ID: id, StandardTaxRate: 0.10, ReducedTaxRate: 0.08, IsActive: true}, nil
}

type holidayReaderStub struct{}

func (holidayReaderStub) FindAllByYearMonth(context.Context, uint64, string) ([]model.ClinicHoliday, error) {
	return []model.ClinicHoliday{}, nil
}

type closingScheduleStub struct{}

func (closingScheduleStub) ResolveSchedule(context.Context, uint64, time.Time) (*sharedkernel.DaySchedule, error) {
	return &sharedkernel.DaySchedule{AmStart: "09:00", AmPmBoundary: "14:00", PmEnd: "18:30"}, nil
}

type medRecStub struct{ db *gorm.DB }

func (s medRecStub) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	var mr model.MedicalRecord
	err := s.db.WithContext(ctx).Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).First(&mr).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", id))
	}
	return &mr, nil
}
func (s medRecStub) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return s.FindByID(ctx, clinicID, id)
}

type treatmentStub struct {
	t                 *testing.T
	forbiddenClinicID uint64
	petB              uint64
}

func (s *treatmentStub) FindUnbilledByPetID(_ context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
	if s.forbiddenClinicID != 0 && clinicID == s.forbiddenClinicID {
		s.t.Fatalf("FindUnbilledByPetID must not query clinic %d without grant", clinicID)
	}
	if petID == s.petB {
		return []model.Treatment{{ID: 9001, Content: realDBBillTreatB, UnitPrice: 1000, Quantity: 1}}, nil
	}
	return []model.Treatment{}, nil
}
func (s *treatmentStub) CountFinalizedUnconfirmedByPetAndDate(_ context.Context, clinicID, petID uint64, _ time.Time) (int64, error) {
	if s.forbiddenClinicID != 0 && clinicID == s.forbiddenClinicID {
		s.t.Fatalf("CountFinalizedUnconfirmed must not query clinic %d without grant", clinicID)
	}
	if petID == s.petB {
		return 2, nil
	}
	return 0, nil
}

type ownerReaderStub struct{ db *gorm.DB }

func (s ownerReaderStub) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	var o model.Owner
	err := s.db.WithContext(ctx).Where("clinic_id = ? AND id = ?", clinicID, id).First(&o).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
	}
	return &o, nil
}

// billingItemRepoNoTrimmingSQL wraps the real repo but returns empty unbilled
// vaccination/exam/trimming sets. AutoMigrate doubles lack migration columns
// used by those SQL paths; treatment stub still provides nonempty B observation.
type billingItemRepoNoTrimmingSQL struct {
	BillingItemRepository
}

func (r billingItemRepoNoTrimmingSQL) FindUnbilledTrimmingItemsByPetID(context.Context, uint64, uint64) ([]model.BillingItem, error) {
	return []model.BillingItem{}, nil
}
func (r billingItemRepoNoTrimmingSQL) FindUnbilledVaccinationItemsByPetID(context.Context, uint64, uint64) ([]model.BillingItem, int, error) {
	return []model.BillingItem{}, 0, nil
}
func (r billingItemRepoNoTrimmingSQL) FindUnbilledExamItemsByPetID(context.Context, uint64, uint64) ([]model.BillingItem, int, error) {
	return []model.BillingItem{}, 0, nil
}
func (r billingItemRepoNoTrimmingSQL) CountNonAccountingTrimmingByPetAndDate(context.Context, uint64, uint64, time.Time) (int64, error) {
	return 0, nil
}

type clinicScopedGuard struct {
	t                 *testing.T
	forbiddenClinicID uint64
}

func (g clinicScopedGuard) reject(clinicID uint64, op string) {
	if g.forbiddenClinicID != 0 && clinicID == g.forbiddenClinicID {
		g.t.Fatalf("%s must not query clinic %d without selected-clinic grant", op, clinicID)
	}
}

type accountingQueryGuard struct {
	AccountingService
	clinicScopedGuard
}

func (s *accountingQueryGuard) List(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	s.reject(clinicID, "accounting List")
	return s.AccountingService.List(ctx, clinicID, filters, page, limit)
}
func (s *accountingQueryGuard) ListForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	for _, id := range clinicIDs {
		s.reject(id, "ListForClinics")
	}
	return s.AccountingService.ListForClinics(ctx, clinicIDs, filters, page, limit)
}
func (s *accountingQueryGuard) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	for _, clinicID := range clinicIDs {
		s.reject(clinicID, "GetByIDForClinics")
	}
	return s.AccountingService.GetByIDForClinics(ctx, clinicIDs, id)
}
func (s *accountingQueryGuard) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
	s.reject(clinicID, "GetDailySummary")
	return s.AccountingService.GetDailySummary(ctx, clinicID, dateStr)
}
func (s *accountingQueryGuard) ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	s.reject(clinicID, "ListUnpaidByBilling")
	return s.AccountingService.ListUnpaidByBilling(ctx, clinicID, startDate, endDate, page, limit)
}
func (s *accountingQueryGuard) GetOwnerUnpaidBalance(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	s.reject(clinicID, "GetOwnerUnpaidBalance")
	return s.AccountingService.GetOwnerUnpaidBalance(ctx, clinicID, ownerID)
}
func (s *accountingQueryGuard) GetMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, year, month, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
	s.reject(clinicID, "GetMonthlyUnpaidCarryover")
	return s.AccountingService.GetMonthlyUnpaidCarryover(ctx, clinicID, year, month, page, limit)
}

type insuranceQueryGuard struct {
	InsuranceService
	clinicScopedGuard
}

func (s *insuranceQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	s.reject(clinicID, "insurance List")
	return s.InsuranceService.List(ctx, clinicID)
}
func (s *insuranceQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	s.reject(clinicID, "insurance GetByID")
	return s.InsuranceService.GetByID(ctx, clinicID, id)
}

type campaignQueryGuard struct {
	CampaignService
	clinicScopedGuard
}

func (s *campaignQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	s.reject(clinicID, "campaign List")
	return s.CampaignService.List(ctx, clinicID)
}
func (s *campaignQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	s.reject(clinicID, "campaign GetByID")
	return s.CampaignService.GetByID(ctx, clinicID, id)
}

type paymentQueryGuard struct {
	PaymentMethodMasterService
	clinicScopedGuard
}

func (s *paymentQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	s.reject(clinicID, "payment List")
	return s.PaymentMethodMasterService.List(ctx, clinicID)
}
func (s *paymentQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	s.reject(clinicID, "payment GetByID")
	return s.PaymentMethodMasterService.GetByID(ctx, clinicID, id)
}

type estimateQueryGuard struct {
	EstimateService
	clinicScopedGuard
}

func (s *estimateQueryGuard) List(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
	s.reject(clinicID, "estimate List")
	return s.EstimateService.List(ctx, clinicID, ownerID, medicalRecordID, status, page, limit)
}
func (s *estimateQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	s.reject(clinicID, "estimate GetByID")
	return s.EstimateService.GetByID(ctx, clinicID, id)
}

type refundQueryGuard struct {
	RefundService
	clinicScopedGuard
}

func (s *refundQueryGuard) ListByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	s.reject(clinicID, "refund List")
	return s.RefundService.ListByBillingID(ctx, clinicID, billingID)
}

type cashRegisterQueryGuard struct {
	CashRegisterService
	clinicScopedGuard
}

func (s *cashRegisterQueryGuard) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	s.reject(clinicID, "cash List")
	return s.CashRegisterService.List(ctx, clinicID, startDate, endDate, page, limit)
}
func (s *cashRegisterQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	s.reject(clinicID, "cash GetByID")
	return s.CashRegisterService.GetByID(ctx, clinicID, id)
}
func (s *cashRegisterQueryGuard) GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
	s.reject(clinicID, "cash GetPreview")
	return s.CashRegisterService.GetPreview(ctx, clinicID, dateStr, period)
}

type reportQueryGuard struct {
	AccountingReportService
	clinicScopedGuard
}

func (s *reportQueryGuard) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error) {
	s.reject(clinicID, "report GetMonthly")
	return s.AccountingReportService.GetMonthly(ctx, clinicID, year, month)
}
func (s *reportQueryGuard) ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error) {
	s.reject(clinicID, "report ExportCSV")
	return s.AccountingReportService.ExportMonthlyCSV(ctx, clinicID, year, month)
}

type billingItemQueryGuard struct {
	BillingItemService
	clinicScopedGuard
}

func (s *billingItemQueryGuard) GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	s.reject(clinicID, "GetUnbilledItems")
	return s.BillingItemService.GetUnbilledItems(ctx, clinicID, petID)
}
func (s *billingItemQueryGuard) GetUnbilledItemDetails(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
	s.reject(clinicID, "GetUnbilledItemDetails")
	return s.BillingItemService.GetUnbilledItemDetails(ctx, clinicID, petID)
}
func (s *billingItemQueryGuard) GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
	s.reject(clinicID, "GetUngroupedSameDaySummary")
	return s.BillingItemService.GetUngroupedSameDaySummary(ctx, clinicID, petID, date)
}
func (s *billingItemQueryGuard) GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
	s.reject(clinicID, "GetDiscountSuggestions")
	return s.BillingItemService.GetDiscountSuggestions(ctx, clinicID, itemID)
}

type confirmationQueryGuard struct {
	BillingConfirmationService
	clinicScopedGuard
}

func (s *confirmationQueryGuard) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	s.reject(clinicID, "GetOrCreate")
	return s.BillingConfirmationService.GetOrCreate(ctx, clinicID, medicalRecordID)
}

type realDBBillingHandlers struct {
	insurance    *InsuranceHandler
	campaign     *CampaignHandler
	payment      *PaymentMethodMasterHandler
	estimate     *EstimateHandler
	accounting   *AccountingHandler
	refund       *RefundHandler
	cash         *CashRegisterHandler
	report       *AccountingReportHandler
	billingItem  *BillingItemHandler
	confirmation *BillingConfirmationHandler
}

type realDBBillingFixture struct {
	fx             testdb.ClinicGrantFixture
	insA, insB     *model.Insurance
	campA, campB   *model.Campaign
	payA, payB     *model.PaymentMethodMaster
	cashPayB       *model.PaymentMethodMaster
	estA, estB     *model.Estimate
	billA, billB   *model.Billing
	itemB          *model.BillingItem
	closeA, closeB *model.CashRegisterClose
	ownerA, ownerB *model.Owner
	petA, petB     *model.Pet
	mrA, mrB       *model.MedicalRecord
	handlers       realDBBillingHandlers
}

func setupRealDBBillingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{}, &model.MedicalRecord{},
		&model.Insurance{}, &model.Campaign{}, &model.CampaignTargetCategory{}, &model.CampaignTargetItem{},
		&model.PaymentMethodMaster{}, &model.Estimate{}, &model.Billing{}, &model.BillingItem{},
		&model.BillingRefund{}, &model.PaymentSplit{}, &model.CashRegisterClose{},
		&model.BillingConfirmation{},
	))
	testdb.Truncate(t, db,
		"billing_confirmations", "billing_refunds", "payment_splits", "billing_items", "billings",
		"cash_register_closes", "estimates", "campaign_target_categories", "campaigns",
		"insurances", "payment_methods", "medical_records", "pets", "animal_species", "owners",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

func seedSpeciesPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, name string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: species.ID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func newRealDBBillingHandlers(t *testing.T, db *gorm.DB, fx realDBBillingFixture, forbiddenClinicID uint64) realDBBillingHandlers {
	t.Helper()
	accRepo := NewAccountingRepository(db)
	payRepo := NewPaymentMethodMasterRepository(db)
	campRepo := NewCampaignRepository(db)
	itemRepo := BillingItemRepository(billingItemRepoNoTrimmingSQL{BillingItemRepository: NewBillingItemRepository(db)})
	tx := testNewTransactor(db)
	treat := &treatmentStub{t: t, petB: fx.petB.ID}

	insSvc := InsuranceService(NewInsuranceService(NewInsuranceRepository(db)))
	campSvc := CampaignService(NewCampaignService(campRepo, nil, tx))
	paySvc := PaymentMethodMasterService(NewPaymentMethodMasterService(payRepo))
	estSvc := EstimateService(NewEstimateService(NewEstimateRepository(db), nil, nil, nil, nil, tx))
	accSvc := AccountingService(NewAccountingService(accRepo, nil, nil, nil, nil, tx, nil, payRepo))
	refundSvc := RefundService(NewRefundService(NewRefundRepository(db), accRepo, nil, tx))
	cashSvc := CashRegisterService(NewCashRegisterService(NewCashRegisterCloseRepository(db), accRepo, closingScheduleStub{}, payRepo, clinicReaderStub{}, tx))
	reportSvc := AccountingReportService(NewAccountingReportService(accRepo, payRepo, holidayReaderStub{}, clinicReaderStub{}))
	itemSvc := BillingItemService(NewBillingItemServiceWithCampaign(itemRepo, accRepo, treat, tx, nil, nil, campRepo, ownerReaderStub{db: db}))
	confSvc := BillingConfirmationService(NewBillingConfirmationService(NewBillingConfirmationRepository(db), medRecStub{db: db}, tx))

	if forbiddenClinicID != 0 {
		g := clinicScopedGuard{t: t, forbiddenClinicID: forbiddenClinicID}
		insSvc = &insuranceQueryGuard{InsuranceService: insSvc, clinicScopedGuard: g}
		campSvc = &campaignQueryGuard{CampaignService: campSvc, clinicScopedGuard: g}
		paySvc = &paymentQueryGuard{PaymentMethodMasterService: paySvc, clinicScopedGuard: g}
		estSvc = &estimateQueryGuard{EstimateService: estSvc, clinicScopedGuard: g}
		accSvc = &accountingQueryGuard{AccountingService: accSvc, clinicScopedGuard: g}
		refundSvc = &refundQueryGuard{RefundService: refundSvc, clinicScopedGuard: g}
		cashSvc = &cashRegisterQueryGuard{CashRegisterService: cashSvc, clinicScopedGuard: g}
		reportSvc = &reportQueryGuard{AccountingReportService: reportSvc, clinicScopedGuard: g}
		itemSvc = &billingItemQueryGuard{BillingItemService: itemSvc, clinicScopedGuard: g}
		confSvc = &confirmationQueryGuard{BillingConfirmationService: confSvc, clinicScopedGuard: g}
		treat.forbiddenClinicID = forbiddenClinicID
	}

	noopPerm := func(_, _ string) gin.HandlerFunc { return func(*gin.Context) {} }
	return realDBBillingHandlers{
		insurance:    NewInsuranceHandler(insSvc),
		campaign:     NewCampaignHandler(campSvc),
		payment:      NewPaymentMethodMasterHandler(paySvc),
		estimate:     NewEstimateHandler(estSvc, nil),
		accounting:   NewAccountingHandler(accSvc, cashSvc, nil),
		refund:       NewRefundHandler(refundSvc, noopPerm),
		cash:         NewCashRegisterHandler(cashSvc, noopPerm),
		report:       NewAccountingReportHandler(reportSvc, noopPerm),
		billingItem:  NewBillingItemHandler(itemSvc, cashSvc, nil, noopPerm),
		confirmation: NewBillingConfirmationHandler(confSvc, noopPerm),
	}
}

func seedRealDBBillingFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBBillingFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3bill realDB")
	ctx := context.Background()
	day, err := time.ParseInLocation(time.DateOnly, realDBBillDate, config.JST)
	require.NoError(t, err)
	start := day.AddDate(0, -1, 0)
	end := day.AddDate(0, 1, 0)

	insA := &model.Insurance{ClinicID: fx.ClinicA, Name: realDBBillInsA, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(insA).Error)
	insB := &model.Insurance{ClinicID: fx.ClinicB, Name: realDBBillInsB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(insB).Error)

	campA := &model.Campaign{ClinicID: fx.ClinicA, Name: realDBBillCampA, StartDate: start, EndDate: end, DiscountType: model.CampaignDiscountTypeRate, DiscountValue: 10, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(campA).Error)
	campB := &model.Campaign{ClinicID: fx.ClinicB, Name: realDBBillCampB, StartDate: start, EndDate: end, DiscountType: model.CampaignDiscountTypeRate, DiscountValue: 15, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(campB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.CampaignTargetCategory{CampaignID: campB.ID, Category: model.ItemCategoryExamination}).Error)

	cashKey := "cash"
	payA := &model.PaymentMethodMaster{ClinicID: fx.ClinicA, Name: realDBBillPayA, SystemKey: &cashKey, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(payA).Error)
	payB := &model.PaymentMethodMaster{ClinicID: fx.ClinicB, Name: realDBBillPayB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(payB).Error)
	cashPayB := &model.PaymentMethodMaster{ClinicID: fx.ClinicB, Name: "現金", SystemKey: &cashKey, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(cashPayB).Error)

	estA := &model.Estimate{ClinicID: fx.ClinicA, EstimateNo: "EA-1", Title: realDBBillEstA, Status: model.EstimateStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(estA).Error)
	estB := &model.Estimate{ClinicID: fx.ClinicB, EstimateNo: "EB-1", Title: realDBBillEstB, Status: model.EstimateStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(estB).Error)

	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBBillOwnerA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBBillOwnerB)
	petA := seedSpeciesPet(t, db, fx.ClinicA, ownerA.ID, realDBBillPetA)
	petB := seedSpeciesPet(t, db, fx.ClinicB, ownerB.ID, realDBBillPetB)

	mrA := &model.MedicalRecord{ClinicID: fx.ClinicA, RecordNo: realDBBillMRA, Date: day, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrA).Error)
	mrB := &model.MedicalRecord{ClinicID: fx.ClinicB, RecordNo: realDBBillMRB, Date: day, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingConfirmation{MedicalRecordID: mrB.ID, Status: model.ConfirmationStatusPending}).Error)

	completedAt := time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, config.JST)
	billA := &model.Billing{ClinicID: fx.ClinicA, OwnerID: &ownerA.ID, PetID: &petA.ID, Status: model.BillingStatusWaiting, ScheduledDate: day, TotalAmount: 1000, Memo: realDBBillMemoA}
	require.NoError(t, db.WithContext(ctx).Create(billA).Error)
	billB := &model.Billing{ClinicID: fx.ClinicB, OwnerID: &ownerB.ID, PetID: &petB.ID, Status: model.BillingStatusCompleted, ScheduledDate: day, CompletedAt: &completedAt, TotalAmount: realDBBillAmountB, Memo: realDBBillMemoB}
	require.NoError(t, db.WithContext(ctx).Create(billB).Error)
	billBWaiting := &model.Billing{ClinicID: fx.ClinicB, OwnerID: &ownerB.ID, PetID: &petB.ID, Status: model.BillingStatusWaiting, ScheduledDate: day, TotalAmount: 3333, Memo: "D3bill-waiting-B"}
	require.NoError(t, db.WithContext(ctx).Create(billBWaiting).Error)

	itemB := &model.BillingItem{BillingID: billB.ID, ClinicID: &fx.ClinicB, Category: model.ItemCategoryExamination, Name: realDBBillItemB, UnitPrice: realDBBillAmountB, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10}
	require.NoError(t, db.WithContext(ctx).Create(itemB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PaymentSplit{ClinicID: fx.ClinicB, BillingID: billB.ID, Method: model.PaymentMethodCash, Amount: realDBBillAmountB, PaymentMethodID: &cashPayB.ID}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{ClinicID: fx.ClinicB, BillingID: billB.ID, Amount: 100, Reason: realDBBillRefundB, RefundedAt: completedAt}).Error)

	closeA := &model.CashRegisterClose{ClinicID: fx.ClinicA, CloseDate: day, Period: "am", CategoryBreakdown: json.RawMessage(`{}`), Memo: realDBBillCloseA}
	require.NoError(t, db.WithContext(ctx).Create(closeA).Error)
	closeB := &model.CashRegisterClose{ClinicID: fx.ClinicB, CloseDate: day, Period: "am", CategoryBreakdown: json.RawMessage(`{}`), Memo: realDBBillCloseB}
	require.NoError(t, db.WithContext(ctx).Create(closeB).Error)

	out := realDBBillingFixture{
		fx: fx, insA: insA, insB: insB, campA: campA, campB: campB, payA: payA, payB: payB, cashPayB: cashPayB,
		estA: estA, estB: estB, billA: billA, billB: billB, itemB: itemB, closeA: closeA, closeB: closeB,
		ownerA: ownerA, ownerB: ownerB, petA: petA, petB: petB, mrA: mrA, mrB: mrB,
	}
	out.handlers = newRealDBBillingHandlers(t, db, out, forbiddenClinicID)
	return out
}

func configureBillGrant(fx realDBBillingFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}
func withID(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}
}

func TestRealDB_SelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("insurances_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/insurances", configureBillGrant(fx, string(model.ResourceMasterInsurance), fx.fx.ClinicA))
		fx.handlers.insurance.ListInsurances(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBBillInsA, realDBBillInsB)
	})
	t.Run("insurances_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/insurances", configureBillGrant(fx, string(model.ResourceMasterInsurance), fx.fx.ClinicB))
		fx.handlers.insurance.ListInsurances(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []nameClinicDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBBillInsA, item.Name)
			if item.ID == fx.insB.ID {
				found = true
			}
		}
		require.True(t, found)
	})
	t.Run("insurances_get_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceMasterInsurance), fx.fx.ClinicA), fx.insB.ID))
		fx.handlers.insurance.GetInsurance(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("insurances_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceMasterInsurance), fx.fx.ClinicB), fx.insB.ID))
		fx.handlers.insurance.GetInsurance(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillInsB)
		assert.NotContains(t, w.Body.String(), realDBBillInsA)
	})
	t.Run("insurances_get_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceMasterInsurance), fx.fx.ClinicB), fx.insA.ID))
		fx.handlers.insurance.GetInsurance(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBBillInsA)
	})

	t.Run("campaigns_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.campaign.ListCampaigns(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("campaigns_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.campaign.ListCampaigns(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []nameClinicDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBBillCampA, item.Name)
			if item.ID == fx.campB.ID {
				found = true
			}
		}
		require.True(t, found)
	})
	t.Run("campaigns_get_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA), fx.campB.ID))
		fx.handlers.campaign.GetCampaign(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("campaigns_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.campB.ID))
		fx.handlers.campaign.GetCampaign(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillCampB)
		assert.NotContains(t, w.Body.String(), realDBBillCampA)
	})
	t.Run("campaigns_get_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.campA.ID))
		fx.handlers.campaign.GetCampaign(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBBillCampA)
	})

	t.Run("payment_methods_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", configureBillGrant(fx, string(model.ResourcePaymentMethod), fx.fx.ClinicA))
		fx.handlers.payment.ListPaymentMethods(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("payment_methods_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", configureBillGrant(fx, string(model.ResourcePaymentMethod), fx.fx.ClinicB))
		fx.handlers.payment.ListPaymentMethods(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []nameClinicDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBBillPayA, item.Name)
			if item.ID == fx.payB.ID {
				found = true
			}
		}
		require.True(t, found)
	})
	t.Run("payment_methods_get_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourcePaymentMethod), fx.fx.ClinicA), fx.payB.ID))
		fx.handlers.payment.GetPaymentMethod(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("payment_methods_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourcePaymentMethod), fx.fx.ClinicB), fx.payB.ID))
		fx.handlers.payment.GetPaymentMethod(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillPayB)
	})
	t.Run("payment_methods_get_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourcePaymentMethod), fx.fx.ClinicB), fx.payA.ID))
		fx.handlers.payment.GetPaymentMethod(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("estimates_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/estimates?page=1&limit=50", configureBillGrant(fx, string(model.ResourceEstimates), fx.fx.ClinicA))
		fx.handlers.estimate.ListEstimates(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("estimates_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/estimates?page=1&limit=50", configureBillGrant(fx, string(model.ResourceEstimates), fx.fx.ClinicB))
		fx.handlers.estimate.ListEstimates(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]estimateDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		found := false
		for _, item := range listed.Data {
			assert.NotEqual(t, realDBBillEstA, item.Title)
			if item.ID == fx.estB.ID {
				found = true
				assert.Equal(t, realDBBillEstB, item.Title)
			}
		}
		require.True(t, found)
	})
	t.Run("estimates_get_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceEstimates), fx.fx.ClinicA), fx.estB.ID))
		fx.handlers.estimate.GetEstimate(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("estimates_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceEstimates), fx.fx.ClinicB), fx.estB.ID))
		fx.handlers.estimate.GetEstimate(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillEstB)
		assert.NotContains(t, w.Body.String(), realDBBillEstA)
	})
	t.Run("estimates_get_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceEstimates), fx.fx.ClinicB), fx.estA.ID))
		fx.handlers.estimate.GetEstimate(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}

func TestRealDB_SelectedClinicBGrantAIsolation_Remaining(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("refunds_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA), fx.billB.ID))
		fx.handlers.refund.ListRefunds(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("refunds_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.billB.ID))
		fx.handlers.refund.ListRefunds(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []refundDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		require.Equal(t, realDBBillRefundB, listed[0].Reason)
	})
	t.Run("refunds_A_billing_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.billA.ID))
		fx.handlers.refund.ListRefunds(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBBillRefundB)
	})

	t.Run("cash_closes_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/cash-register/closes?start_date=2026-09-01&end_date=2026-09-30&page=1&limit=50", configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicA))
		fx.handlers.cash.ListCashRegisterCloses(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("cash_closes_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/cash-register/closes?start_date=2026-09-01&end_date=2026-09-30&page=1&limit=50", configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicB))
		fx.handlers.cash.ListCashRegisterCloses(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]closeDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		found := false
		for _, item := range listed.Data {
			assert.NotEqual(t, realDBBillCloseA, item.Memo)
			if item.ID == fx.closeB.ID {
				found = true
				assert.Equal(t, realDBBillCloseB, item.Memo)
			}
		}
		require.True(t, found)
	})
	t.Run("cash_closes_get_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicA), fx.closeB.ID))
		fx.handlers.cash.GetCashRegisterClose(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("cash_closes_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicB), fx.closeB.ID))
		fx.handlers.cash.GetCashRegisterClose(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillCloseB)
		assert.NotContains(t, w.Body.String(), realDBBillCloseA)
	})
	t.Run("cash_closes_get_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicB), fx.closeA.ID))
		fx.handlers.cash.GetCashRegisterClose(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
	t.Run("cash_preview_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/cash-register/preview?date="+realDBBillDate+"&period=am", configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicA))
		fx.handlers.cash.GetCashRegisterPreview(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("cash_preview_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/cash-register/preview?date="+realDBBillDate+"&period=am", configureBillGrant(fx, string(model.ResourceCashRegisterClose), fx.fx.ClinicB))
		fx.handlers.cash.GetCashRegisterPreview(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), `"theoretical_cash"`)
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
	})

	t.Run("unpaid_list_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/unpaid?start_date=2026-09-01&end_date=2026-09-30&group_by=billing&page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.ListUnpaidBillings(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unpaid_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/unpaid?start_date=2026-09-01&end_date=2026-09-30&group_by=billing&page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.ListUnpaidBillings(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "D3bill-waiting-B")
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
	})
	t.Run("unpaid_balance_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/accountings/unpaid-balance?owner_id=%d", fx.ownerB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.GetOwnerUnpaidBalance(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unpaid_balance_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/accountings/unpaid-balance?owner_id=%d", fx.ownerB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.GetOwnerUnpaidBalance(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got unpaidBalanceDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Greater(t, got.UnpaidTotal, int64(0))
		require.Greater(t, got.UnpaidCount, int64(0))
	})
	t.Run("unpaid_balance_A_owner_zero_no_leak", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/accountings/unpaid-balance?owner_id=%d", fx.ownerA.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.GetOwnerUnpaidBalance(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got unpaidBalanceDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, int64(0), got.UnpaidTotal)
		assert.NotContains(t, w.Body.String(), realDBBillOwnerA)
	})
	t.Run("unpaid_monthly_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/unpaid-monthly?year=2026&month=9&page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.GetUnpaidMonthlySummary(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unpaid_monthly_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/unpaid-monthly?year=2026&month=9&page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.GetUnpaidMonthlySummary(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBBillOwnerA)
		assert.Contains(t, w.Body.String(), realDBBillOwnerB)
	})

	t.Run("reports_monthly_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reports/monthly?year=2026&month=9", configureBillGrant(fx, string(model.ResourceAccountingReports), fx.fx.ClinicA))
		fx.handlers.report.GetMonthlyReport(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("reports_monthly_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reports/monthly?year=2026&month=9", configureBillGrant(fx, string(model.ResourceAccountingReports), fx.fx.ClinicB))
		fx.handlers.report.GetMonthlyReport(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got monthlyReportDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Greater(t, got.Summary.TotalAmount, int64(0))
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
	})
	t.Run("reports_csv_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reports/monthly/csv?year=2026&month=9", configureBillGrant(fx, string(model.ResourceAccountingReports), fx.fx.ClinicA))
		fx.handlers.report.ExportMonthlyCSV(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("reports_csv_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reports/monthly/csv?year=2026&month=9", configureBillGrant(fx, string(model.ResourceAccountingReports), fx.fx.ClinicB))
		fx.handlers.report.ExportMonthlyCSV(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NotEmpty(t, w.Body.Bytes())
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
	})

	t.Run("unbilled_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/unbilled?pet_id=%d", fx.petB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.billingItem.GetUnbilledItems(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unbilled_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/unbilled?pet_id=%d", fx.petB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.billingItem.GetUnbilledItems(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []unbilledItemDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		assert.Equal(t, realDBBillTreatB, listed[0].Name)
	})
	t.Run("unbilled_details_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/unbilled-details?pet_id=%d", fx.petB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.billingItem.GetUnbilledItemDetails(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unbilled_details_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/unbilled-details?pet_id=%d", fx.petB.ID), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.billingItem.GetUnbilledItemDetails(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillTreatB)
	})
	t.Run("ungrouped_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/ungrouped-same-day?pet_id=%d&date=%s", fx.petB.ID, realDBBillDate), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.billingItem.GetUngroupedSameDay(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("ungrouped_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/billing-items/ungrouped-same-day?pet_id=%d&date=%s", fx.petB.ID, realDBBillDate), configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.billingItem.GetUngroupedSameDay(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got ungroupedDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Greater(t, got.MedicalRecordCount, int64(0))
		require.True(t, got.HasUngrouped)
	})
	t.Run("discount_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA), fx.itemB.ID))
		fx.handlers.billingItem.GetBillingItemDiscountSuggestions(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("discount_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.itemB.ID))
		fx.handlers.billingItem.GetBillingItemDiscountSuggestions(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got discountDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.NotEmpty(t, got.Suggestions)
		assert.Equal(t, realDBBillCampB, got.Suggestions[0].Name)
		assert.NotContains(t, w.Body.String(), realDBBillCampA)
	})
	t.Run("discount_A_item_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		// no A item seeded with distinctive path; use nonexistent id under B scope
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.itemB.ID+999999))
		fx.handlers.billingItem.GetBillingItemDiscountSuggestions(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("billing_confirmation_grantA_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA), fx.mrB.ID))
		fx.handlers.confirmation.GetBillingConfirmation(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("billing_confirmation_grantB_ok", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.mrB.ID))
		fx.handlers.confirmation.GetBillingConfirmation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got confirmationDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, fmt.Sprintf("%d", fx.mrB.ID), got.MedicalRecordID)
		assert.NotContains(t, w.Body.String(), realDBBillMRA)
	})
	t.Run("billing_confirmation_A_mr_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.mrA.ID))
		fx.handlers.confirmation.GetBillingConfirmation(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBBillMRA)
	})
}

func TestRealDB_CrossClinic_Accountings_ConditionSeparated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list_grantA_selectedB_no_clinic_ids_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings?page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.ListAccountings(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
		assert.NotContains(t, w.Body.String(), realDBBillMemoB)
	})
	t.Run("list_grantA_selectedB_clinic_ids_AB_returns_A_only", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		path := fmt.Sprintf("/api/v1/accountings?page=1&limit=50&clinic_ids=%d,%d", fx.fx.ClinicA, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.ListAccountings(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]billingMemoDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		foundA := false
		for _, item := range listed.Data {
			assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
			assert.NotEqual(t, realDBBillMemoB, item.Memo)
			if item.ID == fx.billA.ID {
				foundA = true
				assert.Equal(t, realDBBillMemoA, item.Memo)
			}
		}
		require.True(t, foundA)
	})
	t.Run("list_grantB_selectedB_returns_B_only", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings?page=1&limit=50", configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.ListAccountings(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]billingMemoDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		foundB := false
		for _, item := range listed.Data {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBBillMemoA, item.Memo)
			if item.ID == fx.billB.ID {
				foundB = true
				assert.Equal(t, realDBBillMemoB, item.Memo)
			}
		}
		require.True(t, foundB)
	})

	t.Run("get_grantA_selectedB_returns_A", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA), fx.billA.ID))
		fx.handlers.accounting.GetAccounting(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillMemoA)
		assert.NotContains(t, w.Body.String(), realDBBillMemoB)
	})
	t.Run("get_grantB_selectedB_A_id_404", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.billA.ID))
		fx.handlers.accounting.GetAccounting(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBBillMemoA)
	})
	t.Run("get_grantB_selectedB_returns_B", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/", withID(configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB), fx.billB.ID))
		fx.handlers.accounting.GetAccounting(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBBillMemoB)
		assert.NotContains(t, w.Body.String(), realDBBillMemoA)
	})

	t.Run("daily_summary_grantA_selectedB_no_clinic_ids_403", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		fx.handlers = newRealDBBillingHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/daily-summary?date="+realDBBillDate, configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.GetDailySummary(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("daily_summary_grantA_selectedB_clinic_ids_AB_returns_A", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		path := fmt.Sprintf("/api/v1/accountings/daily-summary?date=%s&clinic_ids=%d,%d", realDBBillDate, fx.fx.ClinicA, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicA))
		fx.handlers.accounting.GetDailySummary(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got dailySummaryDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		// A has waiting only; completed B amount must not appear as A grand total
		assert.NotContains(t, w.Body.String(), fmt.Sprintf("%d", realDBBillAmountB))
	})
	t.Run("daily_summary_grantB_selectedB_returns_B", func(t *testing.T) {
		db := setupRealDBBillingIsolationTestDB(t)
		fx := seedRealDBBillingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/accountings/daily-summary?date="+realDBBillDate, configureBillGrant(fx, string(model.ResourceAccounting), fx.fx.ClinicB))
		fx.handlers.accounting.GetDailySummary(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got dailySummaryDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Greater(t, got.BillingCount, int64(0))
		require.Greater(t, got.GrandTotal, int64(0))
	})
}
