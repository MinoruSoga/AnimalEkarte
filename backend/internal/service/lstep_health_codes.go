package service

// HLTH / PREV / LTV 健診・予防・物販タグ名定数（FEAT-379）。
const (
	HlthHealthcheckDoneTag         = "HLTH_健診あり"
	HlthHealthcheckNeverTag        = "HLTH_健診未受診"
	HlthAnnual4CheckupTag          = "HLTH_年4回候補"
	PrevVaccineDeadlineTag         = "PREV_ワクチン期限"
	PrevFilariaTag                 = "PREV_フィラリア未完了"
	PrevFleaTickTag                = "PREV_ノミダニ対象"
	LtvFoodPurchaseTag             = "LTV_フード購入あり"
	LtvSuppPurchaseTag             = "LTV_サプリ購入あり"
)

// SPEC-002 Q5/Q6 確定待ち空プレースホルダー。
// var で宣言しテストで差し替え可能にする。本番コードは空のまま（noop）。
var (
	HealthCheckupCodes        []string
	FilariaTestCodes          []string
	FilariaPrescriptionCodes  []string
	FleaTickPrescriptionCodes []string
	FoodPurchaseCodes         []string
	SuppPurchaseCodes         []string
)

// HealthPreventionLookbackDays は健診・予防履歴の参照期間（日数）。
const HealthPreventionLookbackDays = 365

// VaccineDeadlineDays はワクチン期限間近とみなす残日数の閾値。
const VaccineDeadlineDays = 60
