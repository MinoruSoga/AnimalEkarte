package model

// CPM V1 判定閾値のデフォルト値。
// clinic_settings に値が設定されていない or 0 以下の場合の fallback として使用。
const (
	DefaultCPMV1DormantDays            = 240
	DefaultCPMV1NoahDays               = 365
	DefaultCPMV1NoahAnnualVisits       = 3
	DefaultCPMV1NoahLTV          int64 = 80_000
	DefaultCPMV1CoreDays               = 180
	DefaultCPMV1CoreAnnualVisits       = 2
	DefaultCPMV1CoreLTV          int64 = 50_000
	DefaultCPMV1SpotMinAmount    int64 = 30_000
	DefaultCPMV1SpotInactiveDays       = 90
	DefaultCPMV1GrowingMaxDays         = 90
	DefaultCPMV1GrowingMinVisits       = 2
	DefaultCPMV1GrowingMaxVisits       = 3
	DefaultCPMV1LTVBreakLow      int64 = 20_000
)

// CPMV1Thresholds は CPM V1 ステージ判定の全閾値を集約した DTO。
// ゼロ値は WithDefaults() で互換デフォルトに補完される。
type CPMV1Thresholds struct {
	DormantDays      int   // cpm_dormant: 最終来院からの経過日数 >= DormantDays (default 240)
	NoahDays         int   // cpm_noah: 在籍日数 >= NoahDays (default 365)
	NoahAnnualVisits int   // cpm_noah: 年間来院回数 >= NoahAnnualVisits (default 3)
	NoahLTV          int64 // cpm_noah: 累計金額 >= NoahLTV (default 80_000)
	CoreDays         int   // cpm_core: 在籍日数 >= CoreDays (default 180)
	CoreAnnualVisits int   // cpm_core: 年間来院回数 >= CoreAnnualVisits (default 2)
	CoreLTV          int64 // cpm_core: 累計金額 >= CoreLTV (default 50_000); growing 上限にも兼用
	SpotMinAmount    int64 // cpm_spot: 単回最大金額 >= SpotMinAmount (default 30_000)
	SpotInactiveDays int   // cpm_spot: 最終来院からの経過日数 > SpotInactiveDays (default 90)
	GrowingMaxDays   int   // cpm_growing: 初来院からの経過日数 <= GrowingMaxDays (default 90)
	GrowingMinVisits int   // cpm_growing: 総来院回数 >= GrowingMinVisits (default 2)
	GrowingMaxVisits int   // cpm_growing: 総来院回数 <= GrowingMaxVisits (default 3)
	LTVBreakLow      int64 // growing 下限 / encounter 上限境界 (default 20_000)
}

// WithDefaults は 0 以下のフィールドをデフォルトで埋めた新インスタンスを返す。
func (t CPMV1Thresholds) WithDefaults() CPMV1Thresholds {
	out := t
	if out.DormantDays <= 0 {
		out.DormantDays = DefaultCPMV1DormantDays
	}
	if out.NoahDays <= 0 {
		out.NoahDays = DefaultCPMV1NoahDays
	}
	if out.NoahAnnualVisits <= 0 {
		out.NoahAnnualVisits = DefaultCPMV1NoahAnnualVisits
	}
	if out.NoahLTV <= 0 {
		out.NoahLTV = DefaultCPMV1NoahLTV
	}
	if out.CoreDays <= 0 {
		out.CoreDays = DefaultCPMV1CoreDays
	}
	if out.CoreAnnualVisits <= 0 {
		out.CoreAnnualVisits = DefaultCPMV1CoreAnnualVisits
	}
	if out.CoreLTV <= 0 {
		out.CoreLTV = DefaultCPMV1CoreLTV
	}
	if out.SpotMinAmount <= 0 {
		out.SpotMinAmount = DefaultCPMV1SpotMinAmount
	}
	if out.SpotInactiveDays <= 0 {
		out.SpotInactiveDays = DefaultCPMV1SpotInactiveDays
	}
	if out.GrowingMaxDays <= 0 {
		out.GrowingMaxDays = DefaultCPMV1GrowingMaxDays
	}
	if out.GrowingMinVisits <= 0 {
		out.GrowingMinVisits = DefaultCPMV1GrowingMinVisits
	}
	if out.GrowingMaxVisits <= 0 {
		out.GrowingMaxVisits = DefaultCPMV1GrowingMaxVisits
	}
	if out.LTVBreakLow <= 0 {
		out.LTVBreakLow = DefaultCPMV1LTVBreakLow
	}
	return out
}
