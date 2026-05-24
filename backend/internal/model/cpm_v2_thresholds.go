package model

// P1 CPM V2 来院回数閾値のデフォルト値。
// clinic_settings に値が設定されていない or 0 以下の場合の fallback として使用。
const (
	DefaultCPMV2ComingThreshold = 2
	DefaultCPMV2GoodThreshold   = 4
	DefaultCPMV2FamilyThreshold = 8
	DefaultCPMV2NoahThreshold   = 13
)

// CPMV2Thresholds は CPM V2 来院回数 4 段階閾値を集約した DTO。
type CPMV2Thresholds struct {
	Coming int // これから ステージ開始来院回数 (default 2)
	Good   int // いいかんじ ステージ開始来院回数 (default 4)
	Family int // ファミリー ステージ開始来院回数 (default 8)
	Noah   int // ノア ステージ開始来院回数 (default 13)
}

// WithDefaults は 0 以下のフィールドをデフォルトで埋めた新インスタンスを返す。
func (t CPMV2Thresholds) WithDefaults() CPMV2Thresholds {
	out := t
	if out.Coming <= 0 {
		out.Coming = DefaultCPMV2ComingThreshold
	}
	if out.Good <= 0 {
		out.Good = DefaultCPMV2GoodThreshold
	}
	if out.Family <= 0 {
		out.Family = DefaultCPMV2FamilyThreshold
	}
	if out.Noah <= 0 {
		out.Noah = DefaultCPMV2NoahThreshold
	}
	return out
}
