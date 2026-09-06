package clinicale2e

import (
	"fmt"
	"strings"
)

var allowedAppEnvs = map[string]struct{}{
	"test": {},
}

var allowedDBHosts = map[string]struct{}{
	"db":        {},
	"localhost": {},
	"127.0.0.1": {},
}

const (
	clinicNamePrefix  = "e2e-clinical-"
	companyNamePrefix = "e2e-clinical-"
	clinicIDBase      = uint64(991000)
)

// Allow は clinical E2E 合成 fixture を動かしてよい環境かを fail-closed で判定する。
func Allow(appEnv, dbHost string) error {
	env := strings.ToLower(strings.TrimSpace(appEnv))
	if _, ok := allowedAppEnvs[env]; !ok {
		return fmt.Errorf("APP_ENV %q is not allowed for clinical e2e fixtures", appEnv)
	}
	host := strings.ToLower(strings.TrimSpace(dbHost))
	if _, ok := allowedDBHosts[host]; !ok {
		return fmt.Errorf("db host %q is not allowed for clinical e2e fixtures", dbHost)
	}
	return nil
}

// RejectReservedClinicID は八王子/城東の予約 ID を拒否する。
func RejectReservedClinicID(clinicID uint64) error {
	if clinicID == 1 || clinicID == 2 {
		return fmt.Errorf("clinic_id %d is reserved", clinicID)
	}
	return nil
}

// LoginEmail は合成 staff の公開メール規約。パスワードは含めない。
func LoginEmail(clinicID uint64) string {
	return fmt.Sprintf("e2e-clinical-%d@example.test", clinicID)
}
