package model

// ReservationBusinessKind は reservation_types.business_kind の保存値（001_init.sql L8896 の CHECK 制約と一致）。
// 旧 category 列は 001_init.sql L10878 で DROP 済み。NULL は「未分類」で、general などで補完してはならない
// （001_init.sql L8893-8894: 旧 category は dispatch の fallback ではない）。
type ReservationBusinessKind string

const (
	ReservationBusinessKindGeneral   ReservationBusinessKind = "general"
	ReservationBusinessKindTrimming  ReservationBusinessKind = "trimming"
	ReservationBusinessKindInpatient ReservationBusinessKind = "inpatient"
	ReservationBusinessKindHotel     ReservationBusinessKind = "hotel"
)

// HasBusinessKind は business_kind が kind と一致するときだけ true を返す。
// nil レシーバと NULL（未分類）は常に false — 呼び出し側の判定は拒否・スキップ側へ倒れる（fail-closed）。
func (rt *ReservationType) HasBusinessKind(kind ReservationBusinessKind) bool {
	return rt != nil && rt.BusinessKind != nil && ReservationBusinessKind(*rt.BusinessKind) == kind
}

// DerivedCategory は旧 category の wire 互換値を business_kind から導出する。
// trimming は "trimming"、general / inpatient / hotel は "general"、NULL（未分類）は ""（unknown）。
// NULL を "general" で補完せず、"trimming" を返すのは business_kind が trimming の場合だけ。
func (rt *ReservationType) DerivedCategory() ReservationTypeCategory {
	switch {
	case rt.HasBusinessKind(ReservationBusinessKindTrimming):
		return ReservationTypeCategoryTrimming
	case rt.HasBusinessKind(ReservationBusinessKindGeneral),
		rt.HasBusinessKind(ReservationBusinessKindInpatient),
		rt.HasBusinessKind(ReservationBusinessKindHotel):
		return ReservationTypeCategoryGeneral
	default:
		return ""
	}
}

// GroupColor は表示色を予約区分グループの色で返す。グループ未設定・未読込は nil（色なし）。
// 予約区分自身の color 列は 001_init.sql L10878 で DROP 済みで、既定色の定数は発明しない。
func (rt *ReservationType) GroupColor() *string {
	if rt == nil || rt.Group == nil {
		return nil
	}
	color := rt.Group.Color
	return &color
}
