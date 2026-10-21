package medicalrecord

// vaccination_certificate_canonical.go — EMR-486 schema-v1 canonical JSON / hash.
//
// Contract (docs/work/emr/emr-486-certificate-api-contract-20261020.md):
// every JSON integer is first projected to its canonical decimal string so a
// wire int64 is never rounded through binary64; the projected tree is then
// serialized per RFC 8785 (JCS): object keys ordered by UTF-16 code units,
// only the required string escapes, no whitespace. SHA-256 over the UTF-8
// bytes, lowercase hex.
//
//   snapshot_sha256        = SHA256(projection(snapshot))
//   render_manifest_sha256 = SHA256(projection(snapshot.render_manifest))
//   preview_fingerprint    = SHA256(projection({layout, snapshot}))

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"unicode/utf16"
)

var errCertificateNonInteger = errors.New("certificate canonicalization requires integer-only numbers")

// projectCertificateIntegers returns a copy of the strict-decoded tree with
// every JSON integer replaced by its canonical decimal string. Fractions,
// exponents, NaN/Infinity and any non-integer number form are rejected;
// "-0" normalizes to "0".
func projectCertificateIntegers(v any) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			p, err := projectCertificateIntegers(e)
			if err != nil {
				return nil, err
			}
			out[k] = p
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			p, err := projectCertificateIntegers(e)
			if err != nil {
				return nil, err
			}
			out[i] = p
		}
		return out, nil
	case json.Number:
		s := t.String()
		if !jsonIntegerLiteral(s) {
			return nil, errCertificateNonInteger
		}
		if s == "-0" {
			return "0", nil
		}
		return s, nil
	case float64:
		return nil, errCertificateNonInteger
	default:
		return v, nil
	}
}

// certificateCanonicalBytes applies the integer projection then serializes
// per JCS. It is the single source for snapshot/manifest/fingerprint canonical
// bytes (the future writer persists these same bytes as immutable metadata).
func certificateCanonicalBytes(v any) ([]byte, error) {
	projected, err := projectCertificateIntegers(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := writeJCS(&buf, projected); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func certificateSHA256Hex(v any) (string, error) {
	b, err := certificateCanonicalBytes(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// certificateExpandInputs returns a copy of inputs where every schema-defined
// key is present: omitted fields expand to explicit null before hashing, so
// "omitted" and "null" produce identical canonical bytes.
func certificateExpandInputs(inputs map[string]any) map[string]any {
	out := make(map[string]any, len(certificateInputsKeys))
	for _, k := range certificateInputsKeys {
		out[k] = inputs[k] // absent -> nil
	}
	return out
}

// certificateNormalizedSnapshotDoc returns a copy of the decoded doc with the
// inputs object expanded to all schema keys. snapshot_sha256 and
// preview_fingerprint both hash this normalized tree so omitted and explicit
// null inputs produce identical digests.
func certificateNormalizedSnapshotDoc(doc map[string]any) map[string]any {
	normalized := make(map[string]any, len(doc))
	for k, v := range doc {
		normalized[k] = v
	}
	if inputs, ok := doc["inputs"].(map[string]any); ok {
		normalized["inputs"] = certificateExpandInputs(inputs)
	}
	return normalized
}

func certificateSnapshotSHA256(doc map[string]any) (string, error) {
	return certificateSHA256Hex(certificateNormalizedSnapshotDoc(doc))
}

func certificateRenderManifestSHA256(doc map[string]any) (string, error) {
	manifest, ok := doc["render_manifest"]
	if !ok {
		return "", errCertificateInvalidSnapshot
	}
	return certificateSHA256Hex(manifest)
}

// certificatePreviewFingerprint hashes projection({layout, snapshot}), where
// snapshot is the same inputs-expanded tree that certificateSnapshotSHA256
// commits to; exposed for the future issue writer, exercised here by
// determinism tests only.
func certificatePreviewFingerprint(layout string, doc map[string]any) (string, error) {
	return certificateSHA256Hex(map[string]any{"layout": layout, "snapshot": certificateNormalizedSnapshotDoc(doc)})
}

// ─── JCS writer ──────────────────────────────────────────────────────────────

func writeJCS(buf *bytes.Buffer, v any) error {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case string:
		writeJCSString(buf, t)
	case []any:
		buf.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeJCS(buf, e); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return utf16Less(keys[i], keys[j]) })
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeJCSString(buf, k)
			buf.WriteByte(':')
			if err := writeJCS(buf, t[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return errCertificateNonInteger
	}
	return nil
}

// writeJCSString emits a JCS string: only `"` `\` and control chars < 0x20 are
// escaped (\b \f \n \r \t, else \u00xx lowercase hex). `/`, `<`, `>`, `&`,
// U+2028 and U+2029 are written literally. Input is already UTF-8 validated by
// the strict decoder, so bytes are copied verbatim.
func writeJCSString(buf *bytes.Buffer, s string) {
	const lowerHex = "0123456789abcdef"
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if b < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte(lowerHex[b>>4])
				buf.WriteByte(lowerHex[b&0x0f])
			} else {
				buf.WriteByte(b)
			}
		}
	}
	buf.WriteByte('"')
}

// utf16Less orders strings by UTF-16 code units: characters above the BMP
// compare by their leading surrogate (0xD800-0xDBFF), which sorts before
// U+E000-U+FFFF, unlike code-point order.
func utf16Less(a, b string) bool {
	ua := utf16.Encode([]rune(a))
	ub := utf16.Encode([]rune(b))
	n := len(ua)
	if len(ub) < n {
		n = len(ub)
	}
	for i := 0; i < n; i++ {
		if ua[i] != ub[i] {
			return ua[i] < ub[i]
		}
	}
	return len(ua) < len(ub)
}
