package medicalrecord

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificateWareki(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"2026-10-20", "令和8年10月20日"},
		{"2019-05-01", "令和元年5月1日"},
		{"2019-04-30", "平成31年4月30日"},
		{"1989-01-08", "平成元年1月8日"},
		{"1989-01-07", "昭和64年1月7日"},
		{"1926-12-25", "昭和元年12月25日"},
	}
	for _, tc := range cases {
		got, err := certificateWareki(tc.in)
		require.NoError(t, err, tc.in)
		assert.Equal(t, tc.want, got)
	}
	_, err := certificateWareki("1926-12-24")
	assert.ErrorIs(t, err, errCertificateRenderFailed, "pre-Showa dates must not be guessed")
	_, err = certificateWareki("not-a-date")
	assert.ErrorIs(t, err, errCertificateRenderFailed)
}

func TestCertificateGregorian(t *testing.T) {
	got, err := certificateGregorian("2026-10-20")
	require.NoError(t, err)
	assert.Equal(t, "2026年10月20日", got)
}

func TestFormatCertificateYen(t *testing.T) {
	assert.Equal(t, "", formatCertificateYen(nil))
	i64 := func(v int64) *int64 { return &v }
	assert.Equal(t, "0円", formatCertificateYen(i64(0)))
	assert.Equal(t, "3000円", formatCertificateYen(i64(3000)))
	assert.Equal(t, "1,000円", formatCertificateYen(i64(1000)))
	assert.Equal(t, "6,550円", formatCertificateYen(i64(6550)))
	assert.Equal(t, "1,234,567円", formatCertificateYen(i64(1234567)))
}

func TestCertificateBodySize(t *testing.T) {
	assert.Equal(t, "小", certificateBodySize("small"))
	assert.Equal(t, "中", certificateBodySize("medium"))
	assert.Equal(t, "大", certificateBodySize("large"))
	assert.Equal(t, "", certificateBodySize(""))
}

func mustRenderPDF(t *testing.T, raw, layout string) []byte {
	t.Helper()
	snap := mustDecodeSnapshot(t, raw)
	pdf, err := renderCertificatePDF(snap, layout)
	require.NoError(t, err)
	return pdf
}

func TestRenderCertificatePDF_DeterministicAndWellFormed(t *testing.T) {
	for _, layout := range []string{"portrait", "landscape"} {
		pdf1 := mustRenderPDF(t, testCertificateSnapshotJSON, layout)
		pdf2 := mustRenderPDF(t, testCertificateSnapshotJSON, layout)
		assert.Equal(t, pdf1, pdf2, "%s render must be byte-deterministic", layout)
		assert.True(t, bytes.HasPrefix(pdf1, []byte("%PDF-1.7\n")))
		assert.True(t, bytes.HasSuffix(pdf1, []byte("%%EOF\n")))
		assert.Contains(t, string(pdf1), "xref\n0 9\n")
		assert.Contains(t, string(pdf1), "/UniJIS-UTF16-H")
		assert.Contains(t, string(pdf1), "/ID [<")
		assert.Contains(t, string(pdf1), "(D:20261020000000Z)")
	}
	portrait := mustRenderPDF(t, testCertificateSnapshotJSON, "portrait")
	landscape := mustRenderPDF(t, testCertificateSnapshotJSON, "landscape")
	assert.NotEqual(t, portrait, landscape)
	rabies := mustRenderPDF(t, testCertificateRabiesSnapshot(100, 2), "rabies")
	assert.True(t, bytes.HasPrefix(rabies, []byte("%PDF-1.7\n")))
}

// TestRenderCertificatePDF_NoRuntimeClock pins that output contains no
// wall-clock dependence: same snapshot bytes -> same document ID and dates.
func TestRenderCertificatePDF_NoRuntimeClock(t *testing.T) {
	pdf := mustRenderPDF(t, testCertificateSnapshotJSON, "portrait")
	assert.Contains(t, string(pdf), "/CreationDate (D:20261020000000Z)")
	assert.Contains(t, string(pdf), "/ModDate (D:20261020000000Z)")
	id := certificatePDFFileID("abc", "portrait", "def")
	raw, err := hex.DecodeString(id)
	require.NoError(t, err)
	assert.Len(t, raw, 16)
	assert.Len(t, id, 32)
}

func TestRenderCertificatePDF_RenderFailures(t *testing.T) {
	snap := mustDecodeSnapshot(t, testCertificateSnapshotJSON)
	_, err := renderCertificatePDF(snap, "bogus")
	assert.ErrorIs(t, err, errCertificateRenderFailed)
	preShowa := mustDecodeSnapshot(t, string(bytes.Replace(
		[]byte(testCertificateSnapshotJSON),
		[]byte(`"display_issue_date": "2026-10-20"`),
		[]byte(`"display_issue_date": "1900-01-01"`), 1)))
	_, err = renderCertificatePDF(preShowa, "portrait")
	assert.ErrorIs(t, err, errCertificateRenderFailed)
}

func TestCertificateFitText_OverflowFails(t *testing.T) {
	c := &pdfCanvas{}
	require.NoError(t, certificateFitText(c, "short", 0, 100, 200, 20, 10))
	long := ""
	for i := 0; i < 500; i++ {
		long += "あ"
	}
	// Even shrinking to 6pt, 500 full-width glyphs cannot fit in a 40x20 box.
	err := certificateFitText(c, long, 0, 100, 40, 20, 10)
	assert.ErrorIs(t, err, errCertificateRenderFailed, "must never silently truncate")
}

func TestCertificatePDFDate(t *testing.T) {
	got, err := certificatePDFDate("2026-10-20")
	require.NoError(t, err)
	assert.Equal(t, "D:20261020000000Z", got)
	_, err = certificatePDFDate("2026-1-1")
	assert.ErrorIs(t, err, errCertificateRenderFailed)
}
