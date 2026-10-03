package labdevicecap

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-consumer-secret"

func TestIssueVerifyRoundTrip(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	capability, err := Issue(testSecret, "clinic-2", now)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(capability, "v1."))

	require.NoError(t, Verify(testSecret, "clinic-2", capability, now))
	require.NoError(t, Verify(testSecret, "clinic-2", capability, now.Add(TTL-time.Second)))
}

func TestVerifyRejectsCapabilityForOtherClinic(t *testing.T) {
	now := time.Now()
	capability, err := Issue(testSecret, "clinic-1", now)
	require.NoError(t, err)

	err = Verify(testSecret, "clinic-2", capability, now)
	require.ErrorIs(t, err, ErrClinicMismatch)
}

func TestVerifyRejectsExpiredCapability(t *testing.T) {
	now := time.Now()
	capability, err := Issue(testSecret, "clinic-2", now)
	require.NoError(t, err)

	require.NoError(t, Verify(testSecret, "clinic-2", capability, now.Add(TTL)))
	require.ErrorIs(t, Verify(testSecret, "clinic-2", capability, now.Add(TTL+verifyLeeway+time.Second)), ErrExpired)
}

func TestVerifyRejectsTamperedClaims(t *testing.T) {
	now := time.Now()
	capability, err := Issue(testSecret, "clinic-2", now)
	require.NoError(t, err)
	parts := strings.Split(capability, ".")

	// clinic claim を差し替えても MAC が一致しない
	parts[1] = "Y2xpbmljLTE" // base64url("clinic-1")
	require.ErrorIs(t, Verify(testSecret, "clinic-1", strings.Join(parts, "."), now), ErrBadSignature)
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	now := time.Now()
	capability, err := Issue("other-secret", "clinic-2", now)
	require.NoError(t, err)

	require.ErrorIs(t, Verify(testSecret, "clinic-2", capability, now), ErrBadSignature)
}

func TestVerifyRejectsMalformedCapabilities(t *testing.T) {
	now := time.Now()
	for _, capability := range []string{
		"",
		"garbage",
		"v1.only.three.parts",
		"v2." + strings.Repeat("a.", 3) + "b",
		"v1.!!!.!!.$.%%%",
	} {
		assert.Error(t, Verify(testSecret, "clinic-2", capability, now), capability)
	}
}
