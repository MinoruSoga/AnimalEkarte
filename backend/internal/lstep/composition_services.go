package lstep

type applicationCore struct {
	settings        LstepSettingsService
	tagSync         LstepTagSyncService
	sharedFile      SharedFileService
	triggerPriority LstepTriggerPriorityService
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
	tagSync := NewLstepTagSyncService(
		settings, deps.Owners, deps.Vaccinations, deps.MedicalRecords, deps.Accounting,
		repos.tagCache, deps.Pets, deps.Prescriptions, deps.Checkups,
		repos.syncErrorCounter, repos.tagCodeMapping, deps.BillingItems, repos.tagConfig,
	)
	return applicationCore{
		settings:        settings,
		tagSync:         tagSync,
		sharedFile:      NewSharedFileService(repos.sharedFile, deps.Owners, deps.SharedFileStorage),
		triggerPriority: NewLstepTriggerPriorityService(repos.triggerPriority),
	}
}

func newApplicationGraph(
	deps *Dependencies,
	repos *applicationRepositories,
	core applicationCore,
) applicationGraph {
	return applicationGraph{
		settings:        core.settings,
		lineSend:        NewLineSendService(core.settings, deps.Owners, core.sharedFile, repos.tagCache, deps.Audit, repos.lineSendLog, repos.tagConfig),
		lineLink:        NewLineLinkService(deps.Owners, repos.lineLinkToken, deps.ReservationSettings, repos.settings, deps.Transactor, deps.LineLinkAuditTx, deps.Cipher),
		lineCustomer:    NewLineCustomerService(repos.lineCustomers, deps.Owners),
		tag:             NewLstepTagService(core.settings, deps.Owners, repos.tagCache, deps.Audit, repos.tagConfig),
		tagCodeMapping:  NewLstepTagCodeMappingService(repos.tagCodeMapping, deps.Transactor),
		tagConfig:       NewLstepTagConfigService(repos.tagConfig),
		tagSummary:      NewLstepTagSummaryService(repos.tagCache),
		checkupSync:     NewCheckupSyncService(repos.checkupSync, deps.Owners, deps.Pets, repos.tagCache, core.settings, deps.Audit),
		lifecycle:       NewLstepLifecycleService(core.settings, newOwnerLifecyclePorts(deps.Owners, deps.OwnerLifecycle), newPetLifecyclePorts(deps.Pets, deps.PetLifecycle), repos.tagCache, core.tagSync, deps.Audit, repos.tagConfig, deps.Transactor, deps.LifecycleAuditTx),
		deliveryMonitor: NewLstepDeliveryMonitorService(repos.deliveryTriggerLog),
		triggerPriority: core.triggerPriority,
		aggregation:     NewAggregationService(deps.AggregationRepository, core.settings),
		csvImport:       NewLstepCsvImportService(deps.DB, repos.csvImport, NewLstepCSVImportOwnerLookup(), deps.Staff),
		analytics:       NewLstepAnalyticsService(deps.Owners, repos.deliveryTriggerLog, repos.friendAttributeSnapshot),
		sharedFile:      core.sharedFile,
	}
}
