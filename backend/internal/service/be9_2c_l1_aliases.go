package service

import (
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	PrevFilariaTag         = lstep.PrevFilariaTag
	PrevFleaTickTag        = lstep.PrevFleaTickTag
	LtvFoodPurchaseTag     = lstep.LtvFoodPurchaseTag
	HlthHealthcheckDoneTag = lstep.HlthHealthcheckDoneTag

	CPMStageEncounter    = lstep.CPMStageEncounter
	CPMStageGrowing      = lstep.CPMStageGrowing
	CPMStageCore         = lstep.CPMStageCore
	CPMStageSpot         = lstep.CPMStageSpot
	CPMStageNoah         = lstep.CPMStageNoah
	CPMStageDormant      = lstep.CPMStageDormant
	CPMStageUnclassified = lstep.CPMStageUnclassified

	CPMStageV2Encounter = lstep.CPMStageV2Encounter
	CPMStageV2Coming    = lstep.CPMStageV2Coming
	CPMStageV2Good      = lstep.CPMStageV2Good
	CPMStageV2Family    = lstep.CPMStageV2Family
	CPMStageV2Noah      = lstep.CPMStageV2Noah
)

// BE9-2C L① transitional aliases — Lステップ設定/認証系は internal/lstep へ移動済み。
// 残留 consumer（L③〜L⑤ で移動予定の lstep 系 service）互換のための alias。
// REMOVE: 各 consumer の移動時。L② で LINE 送信系 6型（LineCustomer/LineSend/LineLink/
// LineMessaging Service・SendLineMessageInput/Result）は consumer ゼロになったため削除済み。
type (
	LstepSettingsService      = lstep.LstepSettingsService
	LstepSettingsResponse     = lstep.LstepSettingsResponse
	UpdateLstepSettingsInput  = lstep.UpdateLstepSettingsInput
	LstepConnectionTestResult = lstep.LstepConnectionTestResult

	// L③a transitional aliases — tag-sync core now lives in internal/lstep.
	// Residual service consumers are removed in L③b-L⑤. REMOVE: BE9-2F.
	LstepTagSyncService        = lstep.LstepTagSyncService
	LstepTagService            = lstep.LstepTagService
	LstepTagCodeMappingService = lstep.LstepTagCodeMappingService
	LstepTagConfigService      = lstep.LstepTagConfigService
	LstepTagSummaryService     = lstep.LstepTagSummaryService
	CPMStage                   = lstep.CPMStage
	CPMStageV2                 = lstep.CPMStageV2
	CPMData                    = lstep.CPMData
	CPMStageV2Input            = lstep.CPMStageV2Input
)

//nolint:gocritic // by-value signature is the transitional API contract; changing it would break residual consumers before L③b/L⑤.
func CalculateCPMStage(input CPMData) CPMStage {
	return lstep.CalculateCPMStage(input)
}

func CalculateCPMStageV2(input CPMStageV2Input) CPMStageV2 {
	return lstep.CalculateCPMStageV2(input)
}

// NewLstepTagSyncFromRepos keeps the two composition roots on one dependency
// order while the repository aggregate still exists. The reservation repository
// was never consumed by tag sync and is intentionally no longer forwarded.
func NewLstepTagSyncFromRepos(repos *repository.Repositories, settings LstepSettingsService) LstepTagSyncService {
	return lstep.NewLstepTagSyncService(
		settings,
		repos.Owner,
		repos.Vaccination,
		repos.MedicalRecord,
		repos.Accounting,
		repos.LstepTagCache,
		repos.Pet,
		repos.Prescription,
		repos.Checkup,
		repos.LstepSyncErrorCounter,
		repos.LstepTagCodeMapping,
		repos.BillingItem,
		repos.LstepTagConfig,
	)
}

const tagPrefixCheckupDone = lstep.TagPrefixCheckupDone

func isSystemManagedTag(tagName string) bool {
	return lstep.IsSystemManagedTag(tagName)
}
