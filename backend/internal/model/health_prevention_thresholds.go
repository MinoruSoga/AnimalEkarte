package model

// 健診・予防タグ判定閾値のデフォルト値。
// clinic_settings に値が設定されていない or 0 以下の場合の fallback として使用。
const (
	DefaultHealthPreventionLookbackDays = 365 // 健診・予防履歴の参照期間（日数）
	DefaultVaccineDeadlineDays          = 60  // ワクチン期限間近とみなす残日数
)

// HealthPreventionThresholds は健診・予防タグ判定の全閾値を集約した DTO。
// ゼロ値は WithDefaults() で互換デフォルトに補完される。
type HealthPreventionThresholds struct {
	LookbackDays    int // 健診・予防履歴の参照期間（日数, default 365）
	VaccineDeadline int // ワクチン期限間近とみなす残日数（default 60）
}

// WithDefaults は 0 以下のフィールドをデフォルトで埋めた新インスタンスを返す。
func (t HealthPreventionThresholds) WithDefaults() HealthPreventionThresholds {
	out := t
	if out.LookbackDays <= 0 {
		out.LookbackDays = DefaultHealthPreventionLookbackDays
	}
	if out.VaccineDeadline <= 0 {
		out.VaccineDeadline = DefaultVaccineDeadlineDays
	}
	return out
}
