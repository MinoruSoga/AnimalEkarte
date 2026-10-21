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

// certJSONKind classifies a decoded JSON value.
type certJSONKind int

const (
	certJSONNull certJSONKind = iota
	certJSONBool
	certJSONNumber
	certJSONString
	certJSONArray
	certJSONObject
)

// certJSONValue is the typed JSON tree produced by the strict decoder in
// vaccination_certificate_snapshot.go. It replaces the former
// map[string]any/[]any plumbing so validation, integer projection and the JCS
// writer stay fully typed (no `any` escape hatches). Num keeps the exact
// source literal via json.Number; a zero value is JSON null.
type certJSONValue struct {
	Kind certJSONKind
	Bool bool
	Num  json.Number
	Str  string
	Arr  []certJSONValue
	Obj  map[string]certJSONValue
}

// certJSONNullValue is the explicit-null value used when expanding omitted
// inputs keys before hashing.
var certJSONNullValue = certJSONValue{Kind: certJSONNull}

// projectCertificateIntegers returns a copy of the strict-decoded tree with
// every JSON integer replaced by its canonical decimal string. Fractions,
// exponents, NaN/Infinity and any non-integer number form are rejected;
// "-0" normalizes to "0".
func projectCertificateIntegers(v certJSONValue) (certJSONValue, error) {
	switch v.Kind {
	case certJSONObject:
		out := certJSONValue{Kind: certJSONObject, Obj: make(map[string]certJSONValue, len(v.Obj))}
		for k, e := range v.Obj {
			p, err := projectCertificateIntegers(e)
			if err != nil {
				return certJSONValue{}, err
			}
			out.Obj[k] = p
		}
		return out, nil
	case certJSONArray:
		out := certJSONValue{Kind: certJSONArray, Arr: make([]certJSONValue, len(v.Arr))}
		for i, e := range v.Arr {
			p, err := projectCertificateIntegers(e)
			if err != nil {
				return certJSONValue{}, err
			}
			out.Arr[i] = p
		}
		return out, nil
	case certJSONNumber:
		s := v.Num.String()
		if !jsonIntegerLiteral(s) {
			return certJSONValue{}, errCertificateNonInteger
		}
		if s == "-0" {
			return certJSONValue{Kind: certJSONString, Str: "0"}, nil
		}
		return certJSONValue{Kind: certJSONString, Str: s}, nil
	default:
		return v, nil
	}
}

// certificateCanonicalBytes applies the integer projection then serializes
// per JCS. It is the single source for snapshot/manifest/fingerprint canonical
// bytes (the future writer persists these same bytes as immutable metadata).
func certificateCanonicalBytes(v certJSONValue) ([]byte, error) {
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

func certificateSHA256Hex(v certJSONValue) (string, error) {
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
func certificateExpandInputs(inputs map[string]certJSONValue) map[string]certJSONValue {
	out := make(map[string]certJSONValue, len(certificateInputsKeys))
	for _, k := range certificateInputsKeys {
		v, ok := inputs[k]
		if !ok {
			v = certJSONNullValue
		}
		out[k] = v
	}
	return out
}

// certificateNormalizedSnapshotDoc returns a copy of the decoded doc with the
// inputs object expanded to all schema keys. snapshot_sha256 and
// preview_fingerprint both hash this normalized tree so omitted and explicit
// null inputs produce identical digests.
func certificateNormalizedSnapshotDoc(doc certJSONValue) certJSONValue {
	normalized := certJSONValue{Kind: certJSONObject, Obj: make(map[string]certJSONValue, len(doc.Obj))}
	for k, v := range doc.Obj {
		normalized.Obj[k] = v
	}
	if inputs, ok := doc.Obj["inputs"]; ok && inputs.Kind == certJSONObject {
		normalized.Obj["inputs"] = certJSONValue{Kind: certJSONObject, Obj: certificateExpandInputs(inputs.Obj)}
	}
	return normalized
}

func certificateSnapshotSHA256(doc certJSONValue) (string, error) {
	return certificateSHA256Hex(certificateNormalizedSnapshotDoc(doc))
}

func certificateRenderManifestSHA256(doc certJSONValue) (string, error) {
	manifest, ok := doc.Obj["render_manifest"]
	if !ok {
		return "", errCertificateInvalidSnapshot
	}
	return certificateSHA256Hex(manifest)
}

// certificatePreviewFingerprint hashes projection({layout, snapshot}), where
// snapshot is the same inputs-expanded tree that certificateSnapshotSHA256
// commits to; exposed for the future issue writer, exercised here by
// determinism tests only.
func certificatePreviewFingerprint(layout string, doc certJSONValue) (string, error) {
	return certificateSHA256Hex(certJSONValue{
		Kind: certJSONObject,
		Obj: map[string]certJSONValue{
			"layout":   {Kind: certJSONString, Str: layout},
			"snapshot": certificateNormalizedSnapshotDoc(doc),
		},
	})
}

// ─── JCS writer ──────────────────────────────────────────────────────────────

func writeJCS(buf *bytes.Buffer, v certJSONValue) error {
	switch v.Kind {
	case certJSONNull:
		buf.WriteString("null")
	case certJSONBool:
		if v.Bool {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case certJSONString:
		writeJCSString(buf, v.Str)
	case certJSONArray:
		buf.WriteByte('[')
		for i, e := range v.Arr {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeJCS(buf, e); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case certJSONObject:
		keys := make([]string, 0, len(v.Obj))
		for k := range v.Obj {
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
			if err := writeJCS(buf, v.Obj[k]); err != nil {
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
