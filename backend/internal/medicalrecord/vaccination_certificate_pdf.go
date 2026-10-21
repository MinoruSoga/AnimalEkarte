package medicalrecord

// vaccination_certificate_pdf.go — EMR-487 minimal deterministic PDF 1.7 writer.
//
// Stdlib only (no new modules). The output is byte-deterministic for a fixed
// input: fixed object order, correct xref byte offsets, trailer /Size /Root
// /Info /ID. /ID (both halves) is the first 16 bytes of
// sha256(snapshot_sha256 + "|" + layout + "|" + render_manifest_sha256), hex.
// Info /CreationDate and /ModDate come from snapshot.display_issue_date —
// never time.Now, no randomness, no map-iteration-order output. Content
// streams are uncompressed.
//
// Japanese text uses a Type0 font (/Encoding /UniJIS-UTF16-H, descendant
// CIDFontType0 /HeiseiKakuGo-W5, NOT embedded — no font file ships in this
// repo). This choice is flagged for the PR; it stays isolated in this file so
// an embedded font can replace it later.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

// pdfWriter collects numbered indirect objects then emits header, objects,
// xref table and trailer in one deterministic pass.
type pdfWriter struct {
	objects [][]byte
}

func (w *pdfWriter) add(body []byte) int {
	w.objects = append(w.objects, body)
	return len(w.objects)
}

func (w *pdfWriter) finish(infoObj int, fileIDHex string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.7\n")
	offsets := make([]int, len(w.objects)+1)
	for i, obj := range w.objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n", i+1)
		buf.Write(obj)
		buf.WriteString("\nendobj\n")
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(w.objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(w.objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R /Info %d 0 R /ID [<%s><%s>] >>\nstartxref\n%d\n%%%%EOF\n",
		len(w.objects)+1, infoObj, fileIDHex, fileIDHex, xref)
	return buf.Bytes()
}

// assembleCertificatePDF builds the fixed 8-object document:
//
//	1 Catalog / 2 Pages / 3 Page / 4 Contents stream /
//	5 Type0 font / 6 CIDFontType0 / 7 FontDescriptor / 8 Info
func assembleCertificatePDF(mediaBoxWidth, mediaBoxHeight float64, content []byte, pdfDate, fileIDHex string) []byte {
	w := &pdfWriter{}
	w.add([]byte("<< /Type /Catalog /Pages 2 0 R >>"))
	w.add([]byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"))
	w.add([]byte(fmt.Sprintf(
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		pdfNum(mediaBoxWidth), pdfNum(mediaBoxHeight))))
	var stream bytes.Buffer
	fmt.Fprintf(&stream, "<< /Length %d >>\nstream\n", len(content))
	stream.Write(content)
	stream.WriteString("endstream")
	w.add(stream.Bytes())
	w.add([]byte("<< /Type /Font /Subtype /Type0 /BaseFont /HeiseiKakuGo-W5 /Encoding /UniJIS-UTF16-H /DescendantFonts [6 0 R] >>"))
	w.add([]byte("<< /Type /Font /Subtype /CIDFontType0 /BaseFont /HeiseiKakuGo-W5 /CIDSystemInfo << /Registry (Adobe) /Ordering (Japan1) /Supplement 2 >> /FontDescriptor 7 0 R /DW 1000 /W [1 95 500] >>"))
	// FontBBox/Ascent/Descent/CapHeight/StemV are nominal metrics for the
	// non-embedded base font; consumers substitute a real HeiseiKakuGo-W5.
	w.add([]byte("<< /Type /FontDescriptor /FontName /HeiseiKakuGo-W5 /Flags 4 /FontBBox [-250 -300 1250 1050] /ItalicAngle 0 /Ascent 880 /Descent -120 /CapHeight 700 /StemV 80 >>"))
	w.add([]byte(fmt.Sprintf("<< /Producer (ekarte-certificate-pdf) /CreationDate (%s) /ModDate (%s) >>", pdfDate, pdfDate)))
	return w.finish(len(w.objects), fileIDHex)
}

// pdfCanvas accumulates page content-stream operators. Origin is bottom-left.
type pdfCanvas struct {
	buf bytes.Buffer
}

func (c *pdfCanvas) text(x, y, size float64, s string) {
	if s == "" {
		return
	}
	fmt.Fprintf(&c.buf, "BT /F1 %s Tf 1 0 0 1 %s %s Tm <%s> Tj ET\n",
		pdfNum(size), pdfNum(x), pdfNum(y), pdfHexUTF16(s))
}

func (c *pdfCanvas) hline(x1, x2, y float64) {
	fmt.Fprintf(&c.buf, "%s %s m %s %s l S\n", pdfNum(x1), pdfNum(y), pdfNum(x2), pdfNum(y))
}

func (c *pdfCanvas) vline(x, y1, y2 float64) {
	fmt.Fprintf(&c.buf, "%s %s m %s %s l S\n", pdfNum(x), pdfNum(y1), pdfNum(x), pdfNum(y2))
}

func (c *pdfCanvas) rect(x, y, w, h float64) {
	fmt.Fprintf(&c.buf, "%s %s %s %s re S\n", pdfNum(x), pdfNum(y), pdfNum(w), pdfNum(h))
}

func (c *pdfCanvas) bytes() []byte {
	return c.buf.Bytes()
}

func pdfNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// pdfHexUTF16 encodes a Go string as UTF-16BE hex for UniJIS-UTF16-H text
// objects (surrogate pairs for non-BMP code points).
func pdfHexUTF16(s string) string {
	units := utf16.Encode([]rune(s))
	var b strings.Builder
	b.Grow(len(units) * 4)
	for _, u := range units {
		fmt.Fprintf(&b, "%04X", u)
	}
	return b.String()
}

// certificatePDFFileID derives the deterministic document /ID: first 16 bytes
// of sha256(snapshot_sha256 + "|" + layout + "|" + render_manifest_sha256),
// hex (32 lowercase chars, used for both halves).
func certificatePDFFileID(snapshotSHA256, layout, renderManifestSHA256 string) string {
	sum := sha256.Sum256([]byte(snapshotSHA256 + "|" + layout + "|" + renderManifestSHA256))
	return hex.EncodeToString(sum[:16])
}

// certificatePDFDate converts a YYYY-MM-DD snapshot date into a PDF date
// literal D:YYYYMMDD000000Z. No wall-clock time is ever consulted.
func certificatePDFDate(date string) (string, error) {
	if !isCertificateDate(date) {
		return "", errCertificateRenderFailed
	}
	return "D:" + date[0:4] + date[5:7] + date[8:10] + "000000Z", nil
}
