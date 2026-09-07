package billing

import (
	"fmt"
	"strings"
)

var uatSyntheticClosingEnvs = map[string]struct{}{
	"test":        {},
	"development": {},
	"local":       {},
	"dev":         {},
}

var uatSyntheticClosingHosts = map[string]struct{}{
	"db":        {},
	"localhost": {},
	"127.0.0.1": {},
}

// AllowUATSyntheticClosing は S09 合成 helper を動かしてよい環境かを fail-closed で判定する。
func AllowUATSyntheticClosing(appEnv, dbHost string) error {
	env := strings.ToLower(strings.TrimSpace(appEnv))
	if _, ok := uatSyntheticClosingEnvs[env]; !ok {
		return fmt.Errorf("APP_ENV %q is not allowed for synthetic closing fixtures", appEnv)
	}
	host := strings.ToLower(strings.TrimSpace(dbHost))
	if _, ok := uatSyntheticClosingHosts[host]; !ok {
		return fmt.Errorf("db host %q is not allowed for synthetic closing fixtures", dbHost)
	}
	return nil
}

// RejectExistingBillingIDs は既存会計 ID の改変を拒否する。
func RejectExistingBillingIDs(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return fmt.Errorf("existing billing IDs are forbidden")
}

// RejectReservedClinicID は八王子/城東の予約 ID を拒否する。
func RejectReservedClinicID(clinicID uint64) error {
	if clinicID == 1 || clinicID == 2 {
		return fmt.Errorf("clinic_id %d is reserved", clinicID)
	}
	return nil
}
