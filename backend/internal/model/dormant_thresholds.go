package model

// Q21 SPEC-004 dormant prevention 閾値のデフォルト値。
// clinic_settings に値が設定されていない or 0 の場合の fallback として使用。
const (
	DefaultDormantPrevention180Days = 180
	DefaultDormantPrevention210Days = 210
	DefaultDormantPrevention240Days = 240
	DefaultDormantPrevention365Days = 365
)

// DormantThresholds は dormant prevention 4 段階閾値を集約した DTO。
type DormantThresholds struct {
	Stage180 int
	Stage210 int
	Stage240 int
	Stage365 int
}

// WithDefaults は 0 以下のフィールドをデフォルトで埋めた新インスタンスを返す。
func (t DormantThresholds) WithDefaults() DormantThresholds {
	out := t
	if out.Stage180 <= 0 {
		out.Stage180 = DefaultDormantPrevention180Days
	}
	if out.Stage210 <= 0 {
		out.Stage210 = DefaultDormantPrevention210Days
	}
	if out.Stage240 <= 0 {
		out.Stage240 = DefaultDormantPrevention240Days
	}
	if out.Stage365 <= 0 {
		out.Stage365 = DefaultDormantPrevention365Days
	}
	return out
}
