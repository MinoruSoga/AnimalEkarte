package lstep

import "github.com/animal-ekarte/backend/internal/persistence"

type applicationCore struct {
	settings        LstepSettingsService
	tagSync         LstepTagSyncService
	sharedFile      SharedFileService
	triggerPriority LstepTriggerPriorityService
	recipients      lineRecipientReader
}

type applicationGraph struct {
	settings        LstepSettingsService
	lineSend        LineSendService
	lineLink        LineLinkService
	lineCustomer    LineCustomerService
	tag             LstepTagService
	tagCodeMapping  LstepTagCodeMappingService
	tagConfig       LstepTagConfigService
	tagSummary      LstepTagSummaryService
	checkupSync     CheckupSyncService
	lifecycle       LstepLifecycleService
	deliveryMonitor LstepDeliveryMonitorService
	triggerPriority LstepTriggerPriorityService
	aggregation     AggregationService
	csvImport       LstepCsvImportService
	analytics       LstepAnalyticsService
	sharedFile      SharedFileService
}

func newApplicationCore(deps *Dependencies, repos *applicationRepositories) applicationCore {
	settings := NewLstepSettingsService(
		repos.settings, repos.syncSettings, deps.Cipher, deps.Audit, deps.ClinicSettings, deps.Transactor,
	)
	recipients := persistence.NewLineBindingReader(deps.DB)
	tagSync := NewLstepTagSyncService(
		settings, deps.Owners, deps.Vaccinations, deps.MedicalRecords, deps.Accounting,
		repos.tagCache, deps.Pets, deps.Checkups,
		repos.syncErrorCounter, repos.tagCodeMapping, deps.BillingItems, repos.tagConfig,
		recipients, deps.Appointments,
	)
	return applicationCore{
		settings:        settings,
		tagSync:         tagSync,
		sharedFile:      NewSharedFileService(repos.sharedFile, deps.Owners, deps.SharedFileStorage),
		triggerPriority: NewLstepTriggerPriorityService(repos.triggerPriority),
		recipients:      recipients,
	}
}

func newApplicationGraph(
	deps *Dependencies,
	repos *applicationRepositories,
	core applicationCore,
) applicationGraph {
	return applicationGraph{
		settings:        core.settings,
		lineSend:        NewLineSendService(),
		lineLink:        NewLineLinkService(deps.Owners, repos.lineLinkToken, deps.WebhookRoutes, repos.settings, deps.Transactor, deps.LineLinkAuditTx, deps.Cipher),
		lineCustomer:    NewLineCustomerService(repos.lineCustomers, deps.Owners),
		tag:             NewLstepTagService(core.settings, deps.Owners, repos.tagCache, deps.Audit, repos.tagConfig, core.recipients),
		tagCodeMapping:  NewLstepTagCodeMappingService(repos.tagCodeMapping, deps.Transactor),
		tagConfig:       NewLstepTagConfigService(repos.tagConfig),
		tagSummary:      NewLstepTagSummaryService(repos.tagCache),
		checkupSync:     NewCheckupSyncService(repos.checkupSync, deps.Owners, deps.Pets, repos.tagCache, core.settings, deps.Audit, core.recipients),
		lifecycle:       NewLstepLifecycleService(core.settings, newOwnerLifecyclePorts(deps.Owners, deps.OwnerLifecycle), newPetLifecyclePorts(deps.Pets, deps.PetLifecycle), repos.tagCache, core.tagSync, deps.Audit, repos.tagConfig, deps.Transactor, deps.LifecycleAuditTx, core.recipients, deps.PetLife),
		deliveryMonitor: NewLstepDeliveryMonitorService(repos.deliveryTriggerLog),
		triggerPriority: core.triggerPriority,
		aggregation:     NewAggregationService(deps.AggregationRepository, core.settings),
		csvImport:       NewLstepCsvImportService(deps.DB, repos.csvImport, NewLstepCSVImportOwnerLookup(), deps.Staff),
		analytics:       NewLstepAnalyticsService(deps.Owners, repos.deliveryTriggerLog, repos.friendAttributeSnapshot, core.recipients),
		sharedFile:      core.sharedFile,
	}
}
