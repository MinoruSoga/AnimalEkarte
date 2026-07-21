package service

import "github.com/animal-ekarte/backend/internal/lstep"

// BE9-2C L① transitional aliases — Lステップ設定/認証系は internal/lstep へ移動済み。
// 残留 consumer（L②〜L⑤ で移動予定の lstep 系 service/handler）互換のための alias。
// REMOVE: 各 consumer の移動時。
type (
	LstepSettingsService      = lstep.LstepSettingsService
	LstepSettingsResponse     = lstep.LstepSettingsResponse
	UpdateLstepSettingsInput  = lstep.UpdateLstepSettingsInput
	LstepConnectionTestResult = lstep.LstepConnectionTestResult
)
