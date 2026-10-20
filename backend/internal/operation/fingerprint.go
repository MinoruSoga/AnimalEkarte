package operation

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"sort"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const (
	// fingerprintSize is the only accepted digest length.
	// operation_bindings.fingerprint enforces octet_length(fingerprint) = 32.
	fingerprintSize = 32
	// fingerprintDomain separates these digests from any other HMAC computed
	// with the same key ring.
	fingerprintDomain = "c06-operation-fingerprint-v1"
	// minFingerprintKeySize is the shortest accepted key ring entry.
	minFingerprintKeySize = 32
	// maxContentFields and maxContentFieldSize bound one Content so that the
	// canonical form stays small and every part is length prefixed.
	maxContentFields    = 64
	maxContentFieldSize = 256
	// ownerKindField and intentField label the two scope inputs of the
	// canonical form. Content fields follow them in ascending key order.
	ownerKindField = "owner_kind"
	intentField    = "intent"
)

// NewContent validates and copies the immutable business content of one
// intent. The returned Content is safe to fingerprint: later mutation of the
// caller's map cannot change it. A field name must be non-blank, at most 256
// bytes, and at most 64 fields are accepted; values must be at most 256 bytes
// of canonical text (decimal integers instead of floats).
func NewContent(fields map[string]string) (Content, error) {
	if len(fields) > maxContentFields {
		return Content{}, apperrors.WrapInvalidInput("operation content has too many fields")
	}
	copied := make(map[string]string, len(fields))
	for name, value := range fields {
		if strings.TrimSpace(name) == "" {
			return Content{}, apperrors.WrapInvalidInput("operation content field name is blank")
		}
		if len(name) > maxContentFieldSize || len(value) > maxContentFieldSize {
			return Content{}, apperrors.WrapInvalidInput("operation content field is too long")
		}
		copied[name] = value
	}
	return Content{fields: copied}, nil
}

// Equal reports whether both contents carry the same field names and values.
// A zero Content and NewContent(nil) are equal: both carry no fields.
func (c Content) Equal(o Content) bool {
	if len(c.fields) != len(o.fields) {
		return false
	}
	for name, value := range c.fields {
		other, ok := o.fields[name]
		if !ok || other != value {
			return false
		}
	}
	return true
}

// Fingerprinter computes and re-checks operation fingerprints with an injected
// key ring. The ring is never read from the environment, a file or a literal.
type Fingerprinter struct {
	current string
	keys    map[string][]byte
}

// NewFingerprinter returns a fingerprinter using current as the key version for
// new bindings. Every key must be at least 32 bytes, every version must be
// non-blank, current must exist in keys, and both the version set and the key
// bytes are copied.
func NewFingerprinter(current string, keys map[string][]byte) (*Fingerprinter, error) {
	if strings.TrimSpace(current) == "" {
		return nil, apperrors.WrapInternalServerError("operation fingerprint key version is required")
	}
	if len(keys) == 0 {
		return nil, apperrors.WrapInternalServerError("operation fingerprint key ring is empty")
	}
	copied := make(map[string][]byte, len(keys))
	for version, key := range keys {
		if strings.TrimSpace(version) == "" {
			return nil, apperrors.WrapInternalServerError("operation fingerprint key version is blank")
		}
		if len(key) < minFingerprintKeySize {
			return nil, apperrors.WrapInternalServerError("operation fingerprint key is too short")
		}
		keyCopy := make([]byte, len(key))
		copy(keyCopy, key)
		copied[version] = keyCopy
	}
	if _, ok := copied[current]; !ok {
		return nil, apperrors.WrapInternalServerError("operation fingerprint current key version is unknown")
	}
	return &Fingerprinter{current: current, keys: copied}, nil
}

// Compute returns the HMAC-SHA-256 fingerprint of the scope/ content identity
// and the key version that produced it. Scope.ClinicID and the actor are not
// inputs: bindings are unique per clinic and actor in storage, and the actor is
// re-authorised on every attempt.
func (f *Fingerprinter) Compute(scope Scope, c Content) ([]byte, string) {
	key, ok := f.keys[f.current]
	if !ok {
		// Unreachable for a value built by NewFingerprinter; fail closed with an
		// empty sum so a caller can never store an unkeyed fingerprint.
		return nil, f.current
	}
	return fingerprint(key, scope, c), f.current
}

// Matches re-computes the fingerprint with the key version recorded on the
// stored binding and compares it in constant time. An unknown key version or a
// stored digest of the wrong size is an error (fail closed): the attempt must
// not proceed as if the content matched.
func (f *Fingerprinter) Matches(scope Scope, c Content, stored []byte, keyVersion string) (bool, error) {
	key, ok := f.keys[keyVersion]
	if !ok {
		return false, apperrors.WrapInternalServerError("operation fingerprint key version is not configured")
	}
	if len(stored) != fingerprintSize {
		return false, apperrors.WrapInternalServerError("operation fingerprint has an unexpected size")
	}
	return hmac.Equal(fingerprint(key, scope, c), stored), nil
}

// fingerprint is the canonical form: the domain separation prefix, then the
// OwnerKind and Intent fields, then every Content field in ascending key order.
// Each part is netstring framed (decimal byte length, colon, bytes, comma), so
// no two distinct identities share an encoding.
func fingerprint(key []byte, scope Scope, c Content) []byte {
	mac := hmac.New(sha256.New, key)
	writeFingerprintField(mac, fingerprintDomain)
	writeFingerprintField(mac, ownerKindField)
	writeFingerprintField(mac, scope.OwnerKind)
	writeFingerprintField(mac, intentField)
	writeFingerprintField(mac, scope.Intent)

	names := make([]string, 0, len(c.fields))
	for name := range c.fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		writeFingerprintField(mac, name)
		writeFingerprintField(mac, c.fields[name])
	}
	return mac.Sum(nil)
}

func writeFingerprintField(mac hash.Hash, value string) {
	// Netstring framing: "<decimal byte length>:<bytes>,". hash.Hash.Write never
	// returns an error.
	_, _ = mac.Write([]byte(strconv.Itoa(len(value)) + ":" + value + ","))
}
