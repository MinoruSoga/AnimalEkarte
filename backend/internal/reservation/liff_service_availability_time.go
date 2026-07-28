package reservation

import "strings"

// timeToHHMM は "HH:MM:SS" / "HH:MM" / "HHMM" を timeslot_engine が要求する "HHMM" 形式に変換する。
// PostgreSQL の time 型は GORM 経由で "HH:MM:SS" (8文字) として返るため正規化が必要。
func timeToHHMM(s string) string {
	clean := strings.ReplaceAll(s, ":", "")
	if len(clean) >= 4 {
		return clean[:4]
	}
	return clean
}

func ptrTimeToHHMM(s *string) *string {
	if s == nil {
		return nil
	}
	result := timeToHHMM(*s)
	return &result
}
