package lstep

import "gorm.io/gorm"

type applicationRepositories struct {
	settings                LstepSettingsRepository
	syncSettings            LstepSyncSettingsRepository
	sharedFile              SharedFileRepository
	tagCache                LstepTagCacheRepository
	lineSendLog             LineSendLogRepository
	lineLinkToken           LineLinkTokenRepository
	checkupSync             CheckupSyncRepository
	syncErrorCounter        LstepSyncErrorCounterRepository
	tagCodeMapping          LstepTagCodeMappingRepository
	deliveryTriggerLog      LstepDeliveryTriggerLogRepository
	triggerPriority         LstepTriggerPriorityRepository
	csvImport               LstepCsvImportRepository
	friendAttributeSnapshot LstepFriendAttributeSnapshotRepository
	tagConfig               LstepTagConfigRepository
	lineCustomers           LineCustomerRepository
}

func newApplicationRepositories(db *gorm.DB) *applicationRepositories {
	return &applicationRepositories{
		settings:                NewLstepSettingsRepository(db),
		syncSettings:            NewLstepSyncSettingsRepository(db),
		sharedFile:              NewSharedFileRepository(db),
		tagCache:                NewLstepTagCacheRepository(db),
		lineSendLog:             NewLineSendLogRepository(db),
		lineLinkToken:           NewLineLinkTokenRepository(db),
		checkupSync:             NewCheckupSyncRepository(db),
		syncErrorCounter:        NewLstepSyncErrorCounterRepository(db),
		tagCodeMapping:          NewLstepTagCodeMappingRepository(db),
		deliveryTriggerLog:      NewLstepDeliveryTriggerLogRepository(db),
		triggerPriority:         NewLstepTriggerPriorityRepository(db),
		csvImport:               NewLstepCsvImportRepository(db),
		friendAttributeSnapshot: NewLstepFriendAttributeSnapshotRepository(db),
		tagConfig:               NewLstepTagConfigRepository(db),
		lineCustomers:           NewLineCustomerRepository(db),
	}
}
