package operation

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeHMACKeyV2 is the rotated test key. Matches must still verify sums computed
// with the older version while new sums use the current one.
var fakeHMACKeyV2 = []byte("fedcba9876543210fedcba9876543210")

func TestNewContent_Validation(t *testing.T) {
	t.Run("rejects empty field key", func(t *testing.T) {
		_, err := NewContent(map[string]string{"": "value"})
		require.Error(t, err)
	})

	t.Run("accepts the 64 field bound", func(t *testing.T) {
		_, err := NewContent(fields(64, "value"))
		require.NoError(t, err)
	})

	t.Run("rejects more than 64 fields", func(t *testing.T) {
		_, err := NewContent(fields(65, "value"))
		require.Error(t, err)
	})

	t.Run("accepts 256 byte key and value", func(t *testing.T) {
		key := strings.Repeat("k", 256)
		_, err := NewContent(map[string]string{key: strings.Repeat("v", 256)})
		require.NoError(t, err)
	})

	t.Run("rejects key longer than 256 bytes", func(t *testing.T) {
		_, err := NewContent(map[string]string{strings.Repeat("k", 257): "value"})
		require.Error(t, err)
	})

	t.Run("rejects value longer than 256 bytes", func(t *testing.T) {
		_, err := NewContent(map[string]string{"amount": strings.Repeat("v", 257)})
		require.Error(t, err)
	})

	t.Run("copies the input map", func(t *testing.T) {
		source := map[string]string{"amount": "12000"}
		content, err := NewContent(source)
		require.NoError(t, err)
		source["amount"] = "13000"

		fp := newTestFingerprinter(t)
		sum, _ := fp.Compute(testScope(), content)

		expected, err := NewContent(map[string]string{"amount": "12000"})
		require.NoError(t, err)
		expectedSum, _ := fp.Compute(testScope(), expected)
		require.Equal(t, expectedSum, sum)
	})
}

func TestContent_Equal(t *testing.T) {
	first, err := NewContent(map[string]string{"amount": "12000", "billing_id": "42"})
	require.NoError(t, err)
	same, err := NewContent(map[string]string{"billing_id": "42", "amount": "12000"})
	require.NoError(t, err)
	other, err := NewContent(map[string]string{"billing_id": "42", "amount": "12001"})
	require.NoError(t, err)

	require.True(t, first.Equal(same))
	require.False(t, first.Equal(other))
}

func TestNewFingerprinter_Validation(t *testing.T) {
	t.Run("rejects a key shorter than 32 bytes", func(t *testing.T) {
		_, err := NewFingerprinter("k1", map[string][]byte{"k1": []byte("too-short-key")})
		require.Error(t, err)
	})

	t.Run("rejects a missing current version", func(t *testing.T) {
		_, err := NewFingerprinter("k2", map[string][]byte{"k1": fakeHMACKeyV1})
		require.Error(t, err)
	})

	t.Run("accepts current version with rotated keys", func(t *testing.T) {
		fp, err := NewFingerprinter("k2", map[string][]byte{"k1": fakeHMACKeyV1, "k2": fakeHMACKeyV2})
		require.NoError(t, err)
		require.NotNil(t, fp)
	})
}

// TestFingerprinter_Compute covers what is and is not a fingerprint input.
//
// Preconditions (CAS versions, fences) are never fingerprinted: Compute accepts
// a Content value only, and Preconditions is a distinct type that cannot be
// passed here, so there is no code path that folds a precondition into the sum.
func TestFingerprinter_Compute(t *testing.T) {
	fp := newTestFingerprinter(t)
	scope := testScope()

	base, err := NewContent(map[string]string{
		"billing_id":              "42",
		"ack_before_calc_version": "7",
		"after_amount":            "12000",
	})
	require.NoError(t, err)

	t.Run("same content in a different insert order has the same sum", func(t *testing.T) {
		reordered, err := NewContent(map[string]string{
			"after_amount":            "12000",
			"billing_id":              "42",
			"ack_before_calc_version": "7",
		})
		require.NoError(t, err)

		first, version := fp.Compute(scope, base)
		second, otherVersion := fp.Compute(scope, reordered)

		require.Equal(t, first, second)
		require.Equal(t, version, otherVersion)
		require.Len(t, first, fingerprintSize)
	})

	t.Run("one character of a field value changes the sum", func(t *testing.T) {
		changed, err := NewContent(map[string]string{
			"billing_id":              "42",
			"ack_before_calc_version": "7",
			"after_amount":            "12001",
		})
		require.NoError(t, err)

		first, _ := fp.Compute(scope, base)
		second, _ := fp.Compute(scope, changed)
		require.False(t, bytes.Equal(first, second))
	})

	t.Run("a different owner kind changes the sum", func(t *testing.T) {
		other := scope
		other.OwnerKind = "medicalrecord"

		first, _ := fp.Compute(scope, base)
		second, _ := fp.Compute(other, base)
		require.False(t, bytes.Equal(first, second))
	})

	t.Run("a different intent changes the sum", func(t *testing.T) {
		other := scope
		other.Intent = "billing.other_acknowledgement"

		first, _ := fp.Compute(scope, base)
		second, _ := fp.Compute(other, base)
		require.False(t, bytes.Equal(first, second))
	})
}

func TestFingerprinter_Matches(t *testing.T) {
	content, err := NewContent(map[string]string{"billing_id": "42", "after_amount": "12000"})
	require.NoError(t, err)
	changed, err := NewContent(map[string]string{"billing_id": "42", "after_amount": "13000"})
	require.NoError(t, err)
	scope := testScope()

	// A sum stored under an older key version must keep verifying.
	oldFP, err := NewFingerprinter("k1", map[string][]byte{"k1": fakeHMACKeyV1})
	require.NoError(t, err)
	oldSum, oldVersion := oldFP.Compute(scope, content)
	require.Equal(t, "k1", oldVersion)

	fp, err := NewFingerprinter("k2", map[string][]byte{"k1": fakeHMACKeyV1, "k2": fakeHMACKeyV2})
	require.NoError(t, err)

	t.Run("verifies a sum from an earlier key version", func(t *testing.T) {
		matched, err := fp.Matches(scope, content, oldSum, oldVersion)
		require.NoError(t, err)
		require.True(t, matched)
	})

	t.Run("verifies a sum from the current key version", func(t *testing.T) {
		sum, version := fp.Compute(scope, content)
		require.Equal(t, "k2", version)
		require.False(t, bytes.Equal(oldSum, sum))

		matched, err := fp.Matches(scope, content, sum, version)
		require.NoError(t, err)
		require.True(t, matched)
	})

	t.Run("rejects different content under the same key", func(t *testing.T) {
		matched, err := fp.Matches(scope, changed, oldSum, oldVersion)
		require.NoError(t, err)
		require.False(t, matched)
	})

	t.Run("rejects a tampered sum", func(t *testing.T) {
		tampered := append([]byte(nil), oldSum...)
		tampered[0] ^= 0xff

		matched, err := fp.Matches(scope, content, tampered, oldVersion)
		require.NoError(t, err)
		require.False(t, matched)
	})

	t.Run("rejects a truncated sum", func(t *testing.T) {
		matched, _ := fp.Matches(scope, content, oldSum[:16], oldVersion)
		require.False(t, matched)
	})

	t.Run("rejects an unknown key version", func(t *testing.T) {
		_, err := fp.Matches(scope, content, oldSum, "k9")
		require.Error(t, err)
	})
}

func fields(count int, value string) map[string]string {
	out := make(map[string]string, count)
	for i := 0; i < count; i++ {
		out[fmt.Sprintf("field_%02d", i)] = value
	}
	return out
}
