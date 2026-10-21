package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCanonicalBytes(t *testing.T, raw string) []byte {
	t.Helper()
	obj, err := strictJSONObject([]byte(raw))
	require.NoError(t, err)
	b, err := certificateCanonicalBytes(obj)
	require.NoError(t, err)
	return b
}

// TestCertificateCanonical_ContractVectors pins the contract's canonicalization
// table (integer projection then JCS): every integer becomes a decimal string,
// key order is UTF-16 code-unit order, no whitespace.
func TestCertificateCanonical_ContractVectors(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"key sort + int to string", `{"b":2,"a":null}`, `{"a":null,"b":"2"}`},
		{"zero, string kept, key order", `{"amount":0,"number":"0001","name":"狂犬病"}`, `{"amount":"0","name":"狂犬病","number":"0001"}`},
		{"int64-precision integer", `{"amount":9007199254740993}`, `{"amount":"9007199254740993"}`},
		{"negative integer", `{"a":-42}`, `{"a":"-42"}`},
		{"minus zero normalizes", `{"a":-0}`, `{"a":"0"}`},
		{"array order preserved", `{"lots":[null,"B","A",null]}`, `{"lots":[null,"B","A",null]}`},
		{"booleans kept", `{"f":true,"t":false,"n":null}`, `{"f":true,"n":null,"t":false}`},
		{"unicode literal not escaped", `{"s":"é<>&/\u2028"}`, "{\"s\":\"é<>&/\u2028\"}"},
		{"control char escapes", `{"s":"a\nb\tcd"}`, `{"s":"a\nb\tcd"}`},
		// UTF-16 ordering: U+1F600 encodes as D83D DE00, which sorts before U+FFFF.
		{"utf16 key order above BMP", `{"￿":2,"😀":1}`, `{"😀":"1","￿":"2"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, string(mustCanonicalBytes(t, tc.in)))
		})
	}
}

func TestCertificateCanonical_OmittedNullEquivalence(t *testing.T) {
	parse := func(raw string) map[string]any {
		obj, err := strictJSONObject([]byte(raw))
		require.NoError(t, err)
		return obj
	}
	withNull, err := certificateSnapshotSHA256(parse(`{"inputs":{"registration_fee":null},"x":1}`))
	require.NoError(t, err)
	omitted, err := certificateSnapshotSHA256(parse(`{"inputs":{},"x":1}`))
	require.NoError(t, err)
	explicitZero, err := certificateSnapshotSHA256(parse(`{"inputs":{"registration_fee":0},"x":1}`))
	require.NoError(t, err)
	assert.Equal(t, withNull, omitted, "omitted and explicit-null inputs hash identically")
	assert.NotEqual(t, withNull, explicitZero, "explicit zero differs from null")
	expanded := certificateExpandInputs(parse(`{"inputs":{"registration_fee":1}}`)["inputs"].(map[string]any))
	assert.Len(t, expanded, 10)
	assert.Nil(t, expanded["certificate_number"])
	assert.Equal(t, "1", func() string {
		p, _ := projectCertificateIntegers(expanded)
		return p.(map[string]any)["registration_fee"].(string)
	}())
}

func TestCertificateCanonical_NFCvsNFD(t *testing.T) {
	nfc := mustCanonicalBytes(t, `{"s":"\u00e9"}`)  // U+00E9 precomposed
	nfd := mustCanonicalBytes(t, `{"s":"e\u0301"}`) // U+0065 U+0301 decomposed
	assert.NotEqual(t, nfc, nfd, "canonicalization must not normalize Unicode")
}

func TestCertificateCanonical_RejectsNonIntegers(t *testing.T) {
	for _, in := range []string{`{"a":1.5}`, `{"a":1e2}`, `{"a":0.0}`, `{"a":[1.5]}`} {
		obj, err := strictJSONObject([]byte(in))
		require.NoError(t, err)
		_, err = certificateCanonicalBytes(obj)
		assert.ErrorIs(t, err, errCertificateNonInteger, in)
	}
}

func TestCertificateStrictJSON_RejectsDuplicatesAndGarbage(t *testing.T) {
	_, err := strictJSONObject([]byte(`{"a":1,"a":2}`))
	assert.Error(t, err)
	_, err = strictJSONObject([]byte(`{"a":{"b":1,"b":2}}`))
	assert.Error(t, err)
	_, err = strictJSONObject([]byte(`{"a":1} trailing`))
	assert.Error(t, err)
	_, err = strictJSONObject([]byte(`{"s":"\ud800"}`))
	assert.Error(t, err, "lone surrogate must be rejected")
	_, err = strictJSONObject([]byte(`{"s":"\ud83d\ude00"}`))
	assert.NoError(t, err, "paired surrogates pass")
}

func TestCertificateHashes_Deterministic(t *testing.T) {
	obj, err := strictJSONObject([]byte(testCertificateSnapshotJSON))
	require.NoError(t, err)
	doc := obj
	s1, err := certificateSnapshotSHA256(doc)
	require.NoError(t, err)
	s2, err := certificateSnapshotSHA256(doc)
	require.NoError(t, err)
	assert.Equal(t, s1, s2)
	assert.Len(t, s1, 64)
	m1, err := certificateRenderManifestSHA256(doc)
	require.NoError(t, err)
	assert.Len(t, m1, 64)
	p1, err := certificatePreviewFingerprint("portrait", doc)
	require.NoError(t, err)
	p2, err := certificatePreviewFingerprint("portrait", doc)
	require.NoError(t, err)
	assert.Equal(t, p1, p2)
	p3, err := certificatePreviewFingerprint("landscape", doc)
	require.NoError(t, err)
	assert.NotEqual(t, p1, p3)
}
