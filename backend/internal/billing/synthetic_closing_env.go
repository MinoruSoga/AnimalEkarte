package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

const (
	syntheticClosingClinicPrefix  = "s09-clinic-"
	syntheticClosingCompanyPrefix = "s09-synthetic-"
	syntheticClosingCleanupMACKey = "s09-synthetic-closing-cleanup-v1"
)

var uatSyntheticClosingHTTPHosts = map[string]struct{}{
	"backend":   {},
	"localhost": {},
	"127.0.0.1": {},
}

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

// AllowUATSyntheticClosingHTTPHost はブラウザ/HTTP 呼び出し元ホストを fail-closed で判定する。
func AllowUATSyntheticClosingHTTPHost(host string) error {
	normalized := strings.ToLower(strings.TrimSpace(host))
	if hostname, _, err := net.SplitHostPort(normalized); err == nil {
		normalized = hostname
	}
	if _, ok := uatSyntheticClosingHTTPHosts[normalized]; !ok {
		return fmt.Errorf("http host %q is not allowed for synthetic closing fixtures", host)
	}
	return nil
}

// SyntheticClosingLoginEmail は合成 staff の公開メール規約。パスワードは含めない。
func SyntheticClosingLoginEmail(clinicID uint64) string {
	return fmt.Sprintf("s09-%d@example.test", clinicID)
}

// SyntheticClosingCleanupToken は clinic 単位の回収トークン。秘密は env ではなく MAC で束ねる。
func SyntheticClosingCleanupToken(clinicID uint64) string {
	mac := hmac.New(sha256.New, []byte(syntheticClosingCleanupMACKey))
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], clinicID)
	_, _ = mac.Write(buf[:])
	return hex.EncodeToString(mac.Sum(nil))
}

// MatchSyntheticClosingCleanupToken は回収トークンを定数時間比較する。
func MatchSyntheticClosingCleanupToken(clinicID uint64, token string) bool {
	expected := SyntheticClosingCleanupToken(clinicID)
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(token)))
}
