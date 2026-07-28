package lstep

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// extractTagCodes は mappings から code_type が一致するコード一覧を返す（FEAT-379）。
func extractTagCodes(mappings []*model.LstepTagCodeMapping, codeType string) []string {
	var codes []string
	for _, m := range mappings {
		if m.CodeType == codeType {
			codes = append(codes, []string(m.Codes)...)
		}
	}
	return codes
}

// strSet はスライスを O(1) 検索用マップに変換する。
func strSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

// hasVaccineDeadlineSoon は vaccinations の中に NextDate が now から DefaultVaccineDeadlineDays
// 以内のものがあれば true を返す。pure function — モックなしでテスト可能。
func hasVaccineDeadlineSoon(vaccinations []model.Vaccination, now time.Time) bool {
	deadline := now.AddDate(0, 0, model.DefaultVaccineDeadlineDays)
	for i := range vaccinations {
		nd := vaccinations[i].NextDate
		if nd == nil {
			continue
		}
		// now <= NextDate <= deadline
		if !nd.Before(now) && !nd.After(deadline) {
			return true
		}
	}
	return false
}
