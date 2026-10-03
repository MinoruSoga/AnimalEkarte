// Package labdevicecap issues and verifies clinic-bound consumer capabilities
// for the local lab device agent.
//
// codex-security ad9e0152 finding-1: previously the deployment-global
// LAB_DEVICE_AGENT_CONSUMER_TOKEN was returned verbatim to any staff member
// holding the lab-import grant on their selected clinic — that staff member
// could reuse it on another clinic's workstation and read or decide that
// clinic's queued frames. The backend now mints a short-lived HMAC-signed
// capability bound to the authorized selected clinic. The shared secret stays
// on the backend and the workstation flag; it never crosses the wire, so a
// capability issued for clinic A is rejected by clinic B's agent.
package labdevicecap

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TTL is how long an issued capability stays valid. Short enough that a
// leaked capability is useless quickly; the frontend renews via the API and
// re-verification happens on every agent call, so TTL bounds session lifetime.
const TTL = 5 * time.Minute

// verifyLeeway absorbs modest workstation clock skew on the expiry check so a
// slightly-fast clinic Mac does not reject a freshly minted capability.
const verifyLeeway = 30 * time.Second

var (
	ErrMalformed      = errors.New("malformed consumer capability")
	ErrBadSignature   = errors.New("invalid consumer capability signature")
	ErrExpired        = errors.New("consumer capability expired")
	ErrClinicMismatch = errors.New("consumer capability is bound to a different clinic")
)

// Issue returns "v1.<b64(clinic)>.<b64(expUnix)>.<b64(nonce)>.<b64(mac)>" where
// mac = HMAC-SHA256(secret, "v1.<b64(clinic)>.<b64(expUnix)>.<b64(nonce)>").
func Issue(secret, clinicID string, now time.Time) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("capability nonce: %w", err)
	}
	expiry := now.Add(TTL).Unix()
	payload := "v1." +
		base64.RawURLEncoding.EncodeToString([]byte(clinicID)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(expiry, 10))) + "." +
		base64.RawURLEncoding.EncodeToString(nonce)
	return payload + "." + sign(secret, payload), nil
}

// Verify accepts capability only when it is well-formed, authentic under
// secret, unexpired, and bound to expectedClinic. ErrClinicMismatch means the
// MAC is valid but the bound clinic differs — the agent maps that to 403
// (same as the historical clinic mismatch) while the rest map to 401.
func Verify(secret, expectedClinic, capability string, now time.Time) error {
	parts := strings.Split(capability, ".")
	if len(parts) != 5 || parts[0] != "v1" {
		return ErrMalformed
	}
	payload := strings.Join(parts[:4], ".")
	expectedMAC, err := base64.RawURLEncoding.DecodeString(sign(secret, payload))
	if err != nil {
		return ErrMalformed
	}
	providedMAC, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil || !hmac.Equal(expectedMAC, providedMAC) {
		return ErrBadSignature
	}
	clinic, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrMalformed
	}
	if string(clinic) != expectedClinic {
		return ErrClinicMismatch
	}
	expiryRaw, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrMalformed
	}
	expiry, err := strconv.ParseInt(string(expiryRaw), 10, 64)
	if err != nil {
		return ErrMalformed
	}
	if now.After(time.Unix(expiry, 0).Add(verifyLeeway)) {
		return ErrExpired
	}
	return nil
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
