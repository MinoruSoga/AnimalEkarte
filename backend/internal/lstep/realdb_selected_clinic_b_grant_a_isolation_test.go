package lstep

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 lstep package (28 clinic-fixed routes)
//
// Proves selected-clinic grant isolation through real repository + service + HTTP
// handler paths. Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBLstepOwnerNameA        = "D3lstep-realdb-owner-A"
	realDBLstepOwnerNameB        = "D3lstep-realdb-owner-B"
	realDBLstepLineCustomerNameA = "D3lstep-realdb-line-customer-A"
	realDBLstepLineCustomerNameB = "D3lstep-realdb-line-customer-B"
	realDBLstepTagNameA          = "D3lstep-realdb-tag-A"
	realDBLstepTagNameB          = "D3lstep-realdb-tag-B"
	realDBLstepTagCodeNameA      = "D3lstep-realdb-tagcode-A"
	realDBLstepTagCodeNameB      = "D3lstep-realdb-tagcode-B"
	realDBLstepCsvFileA          = "D3lstep-realdb-csv-A.csv"
	realDBLstepCsvFileB          = "D3lstep-realdb-csv-B.csv"
	realDBLstepSharedFileA       = "D3lstep-realdb-shared-A.pdf"
	realDBLstepSharedFileB       = "D3lstep-realdb-shared-B.pdf"
	realDBLstepLineUserIDA       = "D3lstep-line-user-A"
	realDBLstepLineUserIDB       = "D3lstep-line-user-B"
	realDBLstepFriendDisplayB    = "D3lstep-realdb-friend-B"
	realDBLstepSendSummaryB      = "D3lstep-realdb-send-summary-B"
	realDBLstepYearMonth         = "2026-09"
	realDBLstepPriorityMarker    = 77
)

type realDBStubStorage struct{}

func (realDBStubStorage) Upload(context.Context, string, io.Reader, string) error { return nil }
func (realDBStubStorage) GetSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example.test/signed/" + key, nil
}
func (realDBStubStorage) Delete(context.Context, string) error { return nil }

type realDBNoopAudit struct{}

func (realDBNoopAudit) LogLstepOperation(context.Context, uint64, *uint64, string, string, *uint64) error {
	return nil
}
func (realDBNoopAudit) LogLstepOperationWithMetadata(context.Context, uint64, *uint64, string, string, *uint64, any) error {
	return nil
}

type realDBAggregationAdapter struct {
	inner owner.LtvRepository
}

func (a realDBAggregationAdapter) FindOwnerLTV(ctx context.Context, params *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	rows, err := a.inner.FindOwnerLTV(ctx, &owner.FindOwnerLTVParams{
		ClinicID: params.ClinicID, Sort: params.Sort,
		MinTotalAmount: params.MinTotalAmount, MaxTotalAmount: params.MaxTotalAmount,
		MinVisitCount: params.MinVisitCount, LineLinked: params.LineLinked,
		Year: params.Year, From: params.From, To: params.To, AmountBasis: params.AmountBasis,
		IncludeZero: params.IncludeZero, Search: params.Search, PeriodPreset: params.PeriodPreset,
		MaxVisitCount: params.MaxVisitCount, LastVisitBucket: params.LastVisitBucket,
		IncludeNoVisit: params.IncludeNoVisit, Order: params.Order,
	})
	if err != nil {
		return nil, err
	}
	out := make([]OwnerLTVRow, len(rows))
	for i := range rows {
		r := &rows[i]
		out[i] = OwnerLTVRow{
			OwnerID: r.OwnerID, OwnerName: r.OwnerName, LineUserID: r.LineUserID,
			LstepOptOut: r.LstepOptOut, TotalAmount: r.TotalAmount,
			TotalVisitCount: r.TotalVisitCount, AnnualVisitCount: r.AnnualVisitCount,
			LastVisitDate: r.LastVisitDate, FirstVisitDate: r.FirstVisitDate,
			AnnualAmount: r.AnnualAmount, BillingCount: r.BillingCount,
			PeriodVisitCount: r.PeriodVisitCount, DaysSinceLastVisit: r.DaysSinceLastVisit,
			LastVisitBucket: r.LastVisitBucket, MaxSingleVisitAmount: r.MaxSingleVisitAmount,
		}
	}
	return out, nil
}

type realDBLstepFixture struct {
	fx            testdb.ClinicGrantFixture
	ownerA        *model.Owner
	ownerB        *model.Owner
	lineCustomerA *model.LineCustomer
	lineCustomerB *model.LineCustomer
	sharedFileA   *model.SharedFile
	sharedFileB   *model.SharedFile
	csvImportB    *model.LstepCsvImport
	handler       *Handler
}

func setupRealDBLstepIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.LineCustomer{},
		&model.LstepSettings{},
		&model.ClinicIntegration{},
		&model.LstepTagCodeMapping{},
		&model.LstepTagCache{},
		&model.LstepDeliveryTriggerLog{},
		&model.LstepFriendAttributeSnapshot{},
		&model.LstepCsvImport{},
		&model.LstepTriggerPriority{},
		&model.LineSendLog{},
		&model.SharedFile{},
		&model.MedicalRecord{},
	))
	testdb.Truncate(t, db,
		"lstep_delivery_trigger_log",
		"lstep_friend_attribute_snapshots",
		"lstep_csv_imports",
		"lstep_tag_cache",
		"lstep_tag_code_mappings",
		"lstep_trigger_priorities",
		"lstep_settings",
		"clinic_integrations",
		"line_send_logs",
		"shared_files",
		"line_customers",
		"medical_records",
		"pets",
		"owners",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

func guardClinicID(t *testing.T, label string, clinicID, forbidden uint64) {
	t.Helper()
	if clinicID == forbidden {
		t.Fatalf("%s must not query clinic %d without selected-clinic grant", label, clinicID)
	}
}

func newRealDBLstepHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	repos := newApplicationRepositories(db)
	ownerRepo := owner.NewRepository(db, nil)
	settingsSvc := LstepSettingsService(NewLstepSettingsService(repos.settings, repos.syncSettings, nil, realDBNoopAudit{}, nil))
	tagSvc := LstepTagService(NewLstepTagService(settingsSvc, ownerRepo, repos.tagCache, realDBNoopAudit{}, repos.tagConfig))
	tagCodeSvc := LstepTagCodeMappingService(NewLstepTagCodeMappingService(repos.tagCodeMapping, nil))
	tagSummarySvc := LstepTagSummaryService(NewLstepTagSummaryService(repos.tagCache))
	deliveryMonitorSvc := LstepDeliveryMonitorService(NewLstepDeliveryMonitorService(repos.deliveryTriggerLog))
	triggerPrioritySvc := LstepTriggerPriorityService(NewLstepTriggerPriorityService(repos.triggerPriority))
	analyticsSvc := LstepAnalyticsService(NewLstepAnalyticsService(ownerRepo, repos.deliveryTriggerLog, repos.friendAttributeSnapshot))
	csvImportSvc := LstepCsvImportService(NewLstepCsvImportService(db, repos.csvImport, nil, nil))
	lineCustomerSvc := LineCustomerService(NewLineCustomerService(repos.lineCustomers, ownerRepo))
	lineSendSvc := LineSendService(NewLineSendService(settingsSvc, ownerRepo, nil, repos.tagCache, realDBNoopAudit{}, repos.lineSendLog, repos.tagConfig))
	sharedFileSvc := SharedFileService(NewSharedFileService(repos.sharedFile, ownerRepo, realDBStubStorage{}))
	aggregationSvc := AggregationService(NewAggregationService(realDBAggregationAdapter{inner: owner.NewLtvRepository(db)}, settingsSvc))
	checkupSyncSvc := CheckupSyncService(NewCheckupSyncService(repos.checkupSync, ownerRepo, nil, repos.tagCache, settingsSvc, realDBNoopAudit{}))

	if forbiddenClinicID != 0 {
		lineCustomerSvc = &lineCustomerQueryGuard{LineCustomerService: lineCustomerSvc, t: t, forbidden: forbiddenClinicID}
		settingsSvc = &settingsQueryGuard{LstepSettingsService: settingsSvc, t: t, forbidden: forbiddenClinicID}
		tagCodeSvc = &tagCodeQueryGuard{LstepTagCodeMappingService: tagCodeSvc, t: t, forbidden: forbiddenClinicID}
		analyticsSvc = &analyticsQueryGuard{LstepAnalyticsService: analyticsSvc, t: t, forbidden: forbiddenClinicID}
		deliveryMonitorSvc = &deliveryMonitorQueryGuard{LstepDeliveryMonitorService: deliveryMonitorSvc, t: t, forbidden: forbiddenClinicID}
		tagSummarySvc = &tagSummaryQueryGuard{LstepTagSummaryService: tagSummarySvc, t: t, forbidden: forbiddenClinicID}
		triggerPrioritySvc = &triggerPriorityQueryGuard{LstepTriggerPriorityService: triggerPrioritySvc, t: t, forbidden: forbiddenClinicID}
		csvImportSvc = &csvImportQueryGuard{LstepCsvImportService: csvImportSvc, t: t, forbidden: forbiddenClinicID}
		lineSendSvc = &lineSendQueryGuard{LineSendService: lineSendSvc, t: t, forbidden: forbiddenClinicID}
		tagSvc = &tagQueryGuard{LstepTagService: tagSvc, t: t, forbidden: forbiddenClinicID}
		aggregationSvc = &aggregationQueryGuard{AggregationService: aggregationSvc, t: t, forbidden: forbiddenClinicID}
		checkupSyncSvc = &checkupSyncQueryGuard{CheckupSyncService: checkupSyncSvc, t: t, forbidden: forbiddenClinicID}
		sharedFileSvc = &sharedFileQueryGuard{SharedFileService: sharedFileSvc, t: t, forbidden: forbiddenClinicID}
	}

	return NewHandler(
		NewSettingsHandler(settingsSvc, testPermissionMiddleware),
		NewLineSendHandler(lineSendSvc, testPermissionMiddleware),
		NewLineLinkHandler(nil, testPermissionMiddleware),
		NewLineCustomerHandler(lineCustomerSvc, testPermissionMiddleware),
		tagSvc, tagCodeSvc, NewLstepTagConfigService(repos.tagConfig), tagSummarySvc,
		checkupSyncSvc, nil, deliveryMonitorSvc, triggerPrioritySvc,
		aggregationSvc, csvImportSvc, analyticsSvc, sharedFileSvc,
		nil, testPermissionMiddleware, noopPermissionAny,
	)
}

type lineCustomerQueryGuard struct {
	LineCustomerService
	t         *testing.T
	forbidden uint64
}

func (s *lineCustomerQueryGuard) List(ctx context.Context, clinicID uint64) (*LineCustomerListResult, error) {
	guardClinicID(s.t, "lineCustomer.List", clinicID, s.forbidden)
	return s.LineCustomerService.List(ctx, clinicID)
}

type settingsQueryGuard struct {
	LstepSettingsService
	t         *testing.T
	forbidden uint64
}

func (s *settingsQueryGuard) GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error) {
	guardClinicID(s.t, "settings.GetSettings", clinicID, s.forbidden)
	return s.LstepSettingsService.GetSettings(ctx, clinicID)
}

type tagCodeQueryGuard struct {
	LstepTagCodeMappingService
	t         *testing.T
	forbidden uint64
}

func (s *tagCodeQueryGuard) ListMappings(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	guardClinicID(s.t, "tagCode.ListMappings", clinicID, s.forbidden)
	return s.LstepTagCodeMappingService.ListMappings(ctx, clinicID)
}

type analyticsQueryGuard struct {
	LstepAnalyticsService
	t         *testing.T
	forbidden uint64
}

func (s *analyticsQueryGuard) GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error) {
	guardClinicID(s.t, "analytics.GetMonthlyDeliveryStats", clinicID, s.forbidden)
	return s.LstepAnalyticsService.GetMonthlyDeliveryStats(ctx, clinicID, yearMonth)
}
func (s *analyticsQueryGuard) GetVisitConversionSummary(ctx context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error) {
	guardClinicID(s.t, "analytics.GetVisitConversionSummary", clinicID, s.forbidden)
	return s.LstepAnalyticsService.GetVisitConversionSummary(ctx, clinicID, yearMonth, days)
}
func (s *analyticsQueryGuard) GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error) {
	guardClinicID(s.t, "analytics.GetLatestFriendAttributes", clinicID, s.forbidden)
	return s.LstepAnalyticsService.GetLatestFriendAttributes(ctx, clinicID, ownerID)
}

type deliveryMonitorQueryGuard struct {
	LstepDeliveryMonitorService
	t         *testing.T
	forbidden uint64
}

func (s *deliveryMonitorQueryGuard) GetSummary(ctx context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
	guardClinicID(s.t, "deliveryMonitor.GetSummary", input.ClinicID, s.forbidden)
	return s.LstepDeliveryMonitorService.GetSummary(ctx, input)
}
func (s *deliveryMonitorQueryGuard) GetLogs(ctx context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error) {
	guardClinicID(s.t, "deliveryMonitor.GetLogs", input.ClinicID, s.forbidden)
	return s.LstepDeliveryMonitorService.GetLogs(ctx, input)
}

type tagSummaryQueryGuard struct {
	LstepTagSummaryService
	t         *testing.T
	forbidden uint64
}

func (s *tagSummaryQueryGuard) GetTagSummary(ctx context.Context, clinicID uint64) (TagSummaryResponse, error) {
	guardClinicID(s.t, "tagSummary.GetTagSummary", clinicID, s.forbidden)
	return s.LstepTagSummaryService.GetTagSummary(ctx, clinicID)
}
func (s *tagSummaryQueryGuard) ListOwnersByTag(ctx context.Context, clinicID uint64, input ListOwnersByTagInput) (TagOwnerListResponse, error) {
	guardClinicID(s.t, "tagSummary.ListOwnersByTag", clinicID, s.forbidden)
	return s.LstepTagSummaryService.ListOwnersByTag(ctx, clinicID, input)
}

type triggerPriorityQueryGuard struct {
	LstepTriggerPriorityService
	t         *testing.T
	forbidden uint64
}

func (s *triggerPriorityQueryGuard) GetByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	guardClinicID(s.t, "triggerPriority.GetByClinicID", clinicID, s.forbidden)
	return s.LstepTriggerPriorityService.GetByClinicID(ctx, clinicID)
}

type csvImportQueryGuard struct {
	LstepCsvImportService
	t         *testing.T
	forbidden uint64
}

func (s *csvImportQueryGuard) ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	guardClinicID(s.t, "csvImport.ListByClinic", clinicID, s.forbidden)
	return s.LstepCsvImportService.ListByClinic(ctx, clinicID, limit)
}

type lineSendQueryGuard struct {
	LineSendService
	t         *testing.T
	forbidden uint64
}

func (s *lineSendQueryGuard) GetSendLogs(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error) {
	guardClinicID(s.t, "lineSend.GetSendLogs", clinicID, s.forbidden)
	return s.LineSendService.GetSendLogs(ctx, clinicID, ownerID)
}

type tagQueryGuard struct {
	LstepTagService
	t         *testing.T
	forbidden uint64
}

func (s *tagQueryGuard) GetOwnerTags(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error) {
	guardClinicID(s.t, "tag.GetOwnerTags", clinicID, s.forbidden)
	return s.LstepTagService.GetOwnerTags(ctx, clinicID, ownerID)
}

type aggregationQueryGuard struct {
	AggregationService
	t         *testing.T
	forbidden uint64
}

func (s *aggregationQueryGuard) ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
	guardClinicID(s.t, "aggregation.ListOwnerAggregation", clinicID, s.forbidden)
	return s.AggregationService.ListOwnerAggregation(ctx, clinicID, input)
}

type checkupSyncQueryGuard struct {
	CheckupSyncService
	t         *testing.T
	forbidden uint64
}

func (s *checkupSyncQueryGuard) PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
	guardClinicID(s.t, "checkupSync.PreviewCheckupSync", clinicID, s.forbidden)
	return s.CheckupSyncService.PreviewCheckupSync(ctx, clinicID, input, actorID)
}

type sharedFileQueryGuard struct {
	SharedFileService
	t         *testing.T
	forbidden uint64
}

func (s *sharedFileQueryGuard) FindAll(ctx context.Context, clinicID uint64) ([]*SharedFileResponse, error) {
	guardClinicID(s.t, "sharedFile.FindAll", clinicID, s.forbidden)
	return s.SharedFileService.FindAll(ctx, clinicID)
}
func (s *sharedFileQueryGuard) GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error) {
	guardClinicID(s.t, "sharedFile.GetSignedURL", clinicID, s.forbidden)
	return s.SharedFileService.GetSignedURL(ctx, clinicID, id)
}

func seedLivingPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) {
	t.Helper()
	makeSpeciesAndPet(t, db, clinicID, ownerID, petName)
}

func seedOwner(t *testing.T, db *gorm.DB, clinicID uint64, name, lineUserID string) *model.Owner {
	t.Helper()
	lineID := lineUserID
	o := &model.Owner{ClinicID: clinicID, Name: name, LineUserID: &lineID}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

func seedRealDBLstepFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBLstepFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3lstep realDB")
	ctx := context.Background()
	now := time.Now().In(config.JST)
	monthStart := time.Date(2026, 9, 1, 12, 0, 0, 0, config.JST)
	firedAt := monthStart

	ownerA := seedOwner(t, db, fx.ClinicA, realDBLstepOwnerNameA, realDBLstepLineUserIDA)
	ownerB := seedOwner(t, db, fx.ClinicB, realDBLstepOwnerNameB, realDBLstepLineUserIDB)

	lineCustomerA := &model.LineCustomer{
		ClinicID: fx.ClinicA, LineUserID: realDBLstepLineUserIDA,
		DisplayName: realDBLstepLineCustomerNameA, AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(ctx).Create(lineCustomerA).Error)
	lineCustomerB := &model.LineCustomer{
		ClinicID: fx.ClinicB, LineUserID: realDBLstepLineUserIDB,
		DisplayName: realDBLstepLineCustomerNameB, AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(ctx).Create(lineCustomerB).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepSettings{
		ClinicID: fx.ClinicA, IsSyncEnabled: false,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepSettings{
		ClinicID: fx.ClinicB, IsSyncEnabled: true,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepTagCodeMapping{
		ClinicID: fx.ClinicA, TagName: realDBLstepTagCodeNameA, CodeType: model.CodeTypeCheckupType, Codes: []string{"A1"},
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepTagCodeMapping{
		ClinicID: fx.ClinicB, TagName: realDBLstepTagCodeNameB, CodeType: model.CodeTypeCheckupType, Codes: []string{"B1"},
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepTagCache{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID, TagName: realDBLstepTagNameA, Category: "manual", SyncedAt: now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepTagCache{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID, TagName: realDBLstepTagNameB, Category: "manual", SyncedAt: now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepDeliveryTriggerLog{
		OwnerID: ownerA.ID, ClinicID: fx.ClinicA,
		TriggerType: model.TriggerTypeBirthdayMessage, ScheduledAt: now,
		Status: model.TriggerStatusFired, FiredAt: &firedAt,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepDeliveryTriggerLog{
		OwnerID: ownerB.ID, ClinicID: fx.ClinicB,
		TriggerType: model.TriggerTypeBirthdayMessage, ScheduledAt: now,
		Status: model.TriggerStatusFired, FiredAt: &firedAt,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepDeliveryTriggerLog{
		OwnerID: ownerB.ID, ClinicID: fx.ClinicB,
		TriggerType: model.TriggerTypeNextVisitReminder, ScheduledAt: monthStart,
		Status: model.TriggerStatusFired, FiredAt: &firedAt,
	}).Error)

	displayB := realDBLstepFriendDisplayB
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepFriendAttributeSnapshot{
		ClinicID: fx.ClinicA, LineUserID: realDBLstepLineUserIDA,
		DisplayName: strPtr("D3lstep-realdb-friend-A"), SnapshotTakenAt: now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LstepFriendAttributeSnapshot{
		ClinicID: fx.ClinicB, LineUserID: realDBLstepLineUserIDB,
		DisplayName: &displayB, SnapshotTakenAt: now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepCsvImport{
		ID: uuid.New(), ClinicID: fx.ClinicA, CsvType: "friend_attribute",
		FileName: realDBLstepCsvFileA, UploadedByUserID: fx.StaffID, Status: "completed",
	}).Error)
	csvImportB := &model.LstepCsvImport{
		ID: uuid.New(), ClinicID: fx.ClinicB, CsvType: "friend_attribute",
		FileName: realDBLstepCsvFileB, UploadedByUserID: fx.StaffID, Status: "completed",
	}
	require.NoError(t, db.WithContext(ctx).Create(csvImportB).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LstepTriggerPriority{
		ClinicID: fx.ClinicB, TriggerType: model.TriggerTypeBirthdayMessage, Priority: realDBLstepPriorityMarker,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LineSendLog{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID, SentByUserID: fx.StaffID,
		MessageType: "text", ContentSummary: "D3lstep-realdb-send-summary-A",
		Status: "sent", SentAt: now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LineSendLog{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID, SentByUserID: fx.StaffID,
		MessageType: "text", ContentSummary: realDBLstepSendSummaryB,
		Status: "sent", SentAt: now,
	}).Error)

	sharedFileA := &model.SharedFile{
		ClinicID: fx.ClinicA, UploadedBy: fx.StaffID,
		FileType: model.SharedFileTypePDF, FileName: realDBLstepSharedFileA,
		FileKey: "d3lstep/a.pdf", FileSize: 10, Purpose: model.SharedFilePurposeOther,
	}
	require.NoError(t, db.WithContext(ctx).Create(sharedFileA).Error)
	sharedFileB := &model.SharedFile{
		ClinicID: fx.ClinicB, UploadedBy: fx.StaffID,
		FileType: model.SharedFileTypePDF, FileName: realDBLstepSharedFileB,
		FileKey: "d3lstep/b.pdf", FileSize: 20, Purpose: model.SharedFilePurposeOther,
	}
	require.NoError(t, db.WithContext(ctx).Create(sharedFileB).Error)

	return realDBLstepFixture{
		fx: fx, ownerA: ownerA, ownerB: ownerB,
		lineCustomerA: lineCustomerA, lineCustomerB: lineCustomerB,
		sharedFileA: sharedFileA, sharedFileB: sharedFileB,
		csvImportB: csvImportB,
		handler:    newRealDBLstepHandler(t, db, forbiddenClinicID),
	}
}

func configureLstepGrant(fx realDBLstepFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}

func withOwnerID(configure func(*gin.Context), ownerID uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", ownerID)}}
	}
}

func withFileID(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}
}

type realDBCallResult struct {
	code int
	body []byte
}

func callRealDB(t *testing.T, method, path string, configure func(*gin.Context), invoke func(*gin.Context)) realDBCallResult {
	t.Helper()
	c, w := testdb.NewHTTPTestContext(t, method, path, configure)
	invoke(c)
	return realDBCallResult{code: w.Code, body: w.Body.Bytes()}
}

// TestRealDB_SelectedClinicBGrantAIsolation covers all 28 clinic-fixed lstep routes
// via the unique GET handlers (path aliases share handlers).
func TestRealDB_SelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("line_customers_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/line-customers",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA),
			fx.handler.lineCustomer.ListLineCustomers)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepLineCustomerNameA, realDBLstepLineCustomerNameB)
	})

	t.Run("line_customers_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/line-customers",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB),
			fx.handler.lineCustomer.ListLineCustomers)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepLineCustomerNameB)
		assert.NotContains(t, string(got.body), realDBLstepLineCustomerNameA)
	})

	t.Run("lstep_settings_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep-settings",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA),
			fx.handler.lstepSettings.GetLstepSettings)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("lstep_settings_grantB_sync_enabled", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep-settings",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB),
			fx.handler.lstepSettings.GetLstepSettings)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), `"is_sync_enabled":true`)
	})

	t.Run("tag_code_mappings_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep-tag-code-mappings",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA),
			fx.handler.ListTagCodeMappings)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepTagCodeNameA, realDBLstepTagCodeNameB)
	})

	t.Run("tag_code_mappings_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep-tag-code-mappings",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB),
			fx.handler.ListTagCodeMappings)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepTagCodeNameB)
		assert.NotContains(t, string(got.body), realDBLstepTagCodeNameA)
	})

	t.Run("delivery_stats_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats?year_month="+realDBLstepYearMonth,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.GetLstepMonthlyDeliveryStats)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("delivery_stats_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats?year_month="+realDBLstepYearMonth,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.GetLstepMonthlyDeliveryStats)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		var resp lstepMonthlyDeliveryStatsResponse
		require.NoError(t, json.Unmarshal(got.body, &resp))
		require.NotEmpty(t, resp.Rows, "B delivery-stats rows must be nonempty")
		found := false
		for _, row := range resp.Rows {
			if row.TriggerType == model.TriggerTypeNextVisitReminder && row.Count > 0 {
				found = true
			}
		}
		require.True(t, found, "B next_visit_reminder seed must appear: %s", got.body)
	})

	t.Run("visit_conversion_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion?year_month="+realDBLstepYearMonth,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.GetLstepVisitConversionSummary)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("visit_conversion_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion?year_month="+realDBLstepYearMonth,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.GetLstepVisitConversionSummary)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		var resp lstepVisitConversionResponse
		require.NoError(t, json.Unmarshal(got.body, &resp))
		require.Greater(t, resp.DeliveredCount, int64(0), "B visit-conversion delivered_count must be observable")
		require.NotEmpty(t, resp.Rows, "B visit-conversion rows must be nonempty")
	})

	t.Run("checkup_sync_preview_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/checkup-sync/preview?checkup_type=annual",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA),
			fx.handler.GetCheckupSyncPreview)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("checkup_sync_preview_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		seedLivingPet(t, db, fx.fx.ClinicA, fx.ownerA.ID, "D3lstep-pet-A")
		seedLivingPet(t, db, fx.fx.ClinicB, fx.ownerB.ID, "D3lstep-pet-B")
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/checkup-sync/preview?checkup_type=annual",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB),
			fx.handler.GetCheckupSyncPreview)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepOwnerNameB)
		assert.NotContains(t, string(got.body), realDBLstepOwnerNameA)
	})

	t.Run("csv_imports_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/csv-imports",
			configureLstepGrant(fx, string(model.ResourceLstepCsvImport), fx.fx.ClinicA),
			fx.handler.ListLstepCsvImports)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepCsvFileA, realDBLstepCsvFileB)
	})

	t.Run("csv_imports_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/lstep/csv-imports",
			configureLstepGrant(fx, string(model.ResourceLstepCsvImport), fx.fx.ClinicB),
			fx.handler.ListLstepCsvImports)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepCsvFileB)
		assert.NotContains(t, string(got.body), realDBLstepCsvFileA)
	})

	t.Run("delivery_monitor_summary_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/delivery-monitor/summary",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.GetLstepDeliveryTriggerSummary)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("delivery_monitor_summary_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/delivery-monitor/summary",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.GetLstepDeliveryTriggerSummary)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		var resp deliveryTriggerSummaryResponse
		require.NoError(t, json.Unmarshal(got.body, &resp))
		require.Greater(t, resp.Fired, int64(0), "B delivery-monitor summary fired must be observable")
	})

	t.Run("delivery_monitor_logs_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/delivery-monitor/logs",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.GetLstepDeliveryTriggerLogs)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepOwnerNameA, realDBLstepOwnerNameB)
	})

	t.Run("delivery_monitor_logs_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/delivery-monitor/logs",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.GetLstepDeliveryTriggerLogs)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		var resp deliveryTriggerLogsPageResponse
		require.NoError(t, json.Unmarshal(got.body, &resp))
		require.NotEmpty(t, resp.Items, "B delivery-monitor logs must be nonempty")
		assert.Contains(t, string(got.body), realDBLstepOwnerNameB)
		assert.NotContains(t, string(got.body), realDBLstepOwnerNameA)
	})

	t.Run("tag_summary_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/tag-summary",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.GetLstepTagSummary)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepTagNameA, realDBLstepTagNameB)
	})

	t.Run("tag_summary_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/tag-summary",
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.GetLstepTagSummary)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepTagNameB)
		assert.NotContains(t, string(got.body), realDBLstepTagNameA)
	})

	t.Run("owners_by_tag_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/owners?tag="+realDBLstepTagNameB,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA),
			fx.handler.SearchLstepOwnersByTag)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepOwnerNameA, realDBLstepOwnerNameB)
	})

	t.Run("owners_by_tag_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/owners?tag="+realDBLstepTagNameB,
			configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB),
			fx.handler.SearchLstepOwnersByTag)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepOwnerNameB)
		assert.NotContains(t, string(got.body), realDBLstepOwnerNameA)
	})

	t.Run("trigger_priorities_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/trigger-priorities",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA),
			fx.handler.GetLstepTriggerPriorities)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("trigger_priorities_grantB_marker", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/lstep/trigger-priorities",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB),
			fx.handler.GetLstepTriggerPriorities)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), fmt.Sprintf(`"priority":%d`, realDBLstepPriorityMarker))
		assert.Contains(t, string(got.body), fmt.Sprintf(`"clinic_id":"%d"`, fx.fx.ClinicB))
	})

	t.Run("line_send_logs_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/owners/%d/line/send-logs", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA), fx.ownerB.ID),
			fx.handler.lineSend.GetLineSendLogs)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepSendSummaryB)
	})

	t.Run("line_send_logs_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/owners/%d/line/send-logs", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.ownerB.ID),
			fx.handler.lineSend.GetLineSendLogs)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepSendSummaryB)
	})

	t.Run("friend_attributes_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/:clinic_id/owners/%d/lstep/friend-attributes", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicA), fx.ownerB.ID),
			fx.handler.GetLstepOwnerFriendAttributes)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepFriendDisplayB)
	})

	t.Run("friend_attributes_grantB_ok", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/:clinic_id/owners/%d/lstep/friend-attributes", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceLstepAnalytics), fx.fx.ClinicB), fx.ownerB.ID),
			fx.handler.GetLstepOwnerFriendAttributes)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepFriendDisplayB)
		assert.Contains(t, string(got.body), realDBLstepLineUserIDB)
	})

	t.Run("owner_tags_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/owners/%d/lstep/tags", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA), fx.ownerB.ID),
			fx.handler.GetOwnerLstepTags)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepTagNameA, realDBLstepTagNameB)
	})

	t.Run("owner_tags_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/owners/%d/lstep/tags", fx.ownerB.ID),
			withOwnerID(configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.ownerB.ID),
			fx.handler.GetOwnerLstepTags)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepTagNameB)
		assert.NotContains(t, string(got.body), realDBLstepTagNameA)
	})

	t.Run("owner_aggregation_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/owners/aggregations?include_zero=true&include_no_visit=true",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA),
			fx.handler.ListOwnerAggregation)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepOwnerNameA, realDBLstepOwnerNameB)
	})

	t.Run("owner_aggregation_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/clinics/:clinic_id/owners/aggregations?include_zero=true&include_no_visit=true",
			configureLstepGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB),
			fx.handler.ListOwnerAggregation)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepOwnerNameB)
		assert.NotContains(t, string(got.body), realDBLstepOwnerNameA)
	})

	t.Run("shared_files_list_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, "/api/v1/shared-files",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA),
			fx.handler.ListSharedFiles)
		assert.Equal(t, http.StatusForbidden, got.code)
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepSharedFileA, realDBLstepSharedFileB)
	})

	t.Run("shared_files_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, "/api/v1/shared-files",
			configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB),
			fx.handler.ListSharedFiles)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), realDBLstepSharedFileB)
		assert.NotContains(t, string(got.body), realDBLstepSharedFileA)
	})

	t.Run("shared_files_signed_url_grantA_403", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		fx.handler = newRealDBLstepHandler(t, db, fx.fx.ClinicB)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/shared-files/%d/signed-url", fx.sharedFileB.ID),
			withFileID(configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA), fx.sharedFileB.ID),
			fx.handler.GetSharedFileSignedURL)
		assert.Equal(t, http.StatusForbidden, got.code)
	})

	t.Run("shared_files_signed_url_grantB_ok", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/shared-files/%d/signed-url", fx.sharedFileB.ID),
			withFileID(configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB), fx.sharedFileB.ID),
			fx.handler.GetSharedFileSignedURL)
		require.Equal(t, http.StatusOK, got.code, string(got.body))
		assert.Contains(t, string(got.body), "https://example.test/signed/d3lstep/b.pdf")
	})

	t.Run("shared_files_signed_url_A_id_404", func(t *testing.T) {
		db := setupRealDBLstepIsolationTestDB(t)
		fx := seedRealDBLstepFixture(t, db, 0)
		got := callRealDB(t, http.MethodGet, fmt.Sprintf("/api/v1/shared-files/%d/signed-url", fx.sharedFileA.ID),
			withFileID(configureLstepGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB), fx.sharedFileA.ID),
			fx.handler.GetSharedFileSignedURL)
		require.Equal(t, http.StatusNotFound, got.code, string(got.body))
		testdb.AssertBodyOmitsClinicArtifacts(t, got.body, nil, realDBLstepSharedFileA, realDBLstepSharedFileB)
	})
}
