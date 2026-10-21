package medicalrecord

// vaccination_certificate_render.go — EMR-487 server renderer (single drawing
// owner for the three layouts). Renders one page purely from the stored
// snapshot; null fields render blank, never sample values, and no fields
// beyond the lifecycle.md layout table are emitted.
//
// Supported versions (the saved preview path compares the stored
// render_manifest against these constants; a mismatch is dependency-changed):

import (
	"fmt"
	"strconv"
	"time"
)

const (
	certificateRendererVersion   = "ekarte-certificate-pdf/1"
	certificateFontAssetVersion  = "adobe-japan1-heisei-kakugo-w5-unembedded/1"
	certificateTemplatePortrait  = "normal-portrait/1"
	certificateTemplateLandscape = "normal-landscape/1"
	certificateTemplateRabies    = "rabies-a4-landscape/1"
)

// Paper: rabies is user-decided A4 landscape (841.89x595.28pt). The normal
// paper size is not fixed by the spec — assumption: A4 portrait for the
// portrait layout and A4 landscape for the landscape layout.
const (
	certificateA4W         = 595.28
	certificateA4H         = 841.89
	certificateMinFontSize = 6.0
	certificateLabelSize   = 8.0
	certificateValueSize   = 10.0
)

func certificateTemplateForLayout(layout string) string {
	switch layout {
	case "portrait":
		return certificateTemplatePortrait
	case "landscape":
		return certificateTemplateLandscape
	default:
		return certificateTemplateRabies
	}
}

func certificateLayoutAllowed(kind, layout string) bool {
	if kind == "rabies" {
		return layout == "rabies"
	}
	return layout == "portrait" || layout == "landscape"
}

// renderCertificatePDF draws the requested layout into a deterministic PDF.
func renderCertificatePDF(snap *certificateSnapshot, layout string) ([]byte, error) {
	c := &pdfCanvas{}
	var w, h float64
	var err error
	switch layout {
	case "portrait":
		w, h = certificateA4W, certificateA4H
		err = drawCertificateNormalPortrait(c, snap)
	case "landscape":
		w, h = certificateA4H, certificateA4W
		err = drawCertificateNormalLandscape(c, snap)
	case "rabies":
		w, h = certificateA4H, certificateA4W
		err = drawCertificateRabies(c, snap)
	default:
		return nil, errCertificateRenderFailed
	}
	if err != nil {
		return nil, err
	}
	manifestSHA, err := certificateRenderManifestSHA256(snap.Doc)
	if err != nil {
		return nil, errCertificateRenderFailed
	}
	snapshotSHA, err := certificateSnapshotSHA256(snap.Doc)
	if err != nil {
		return nil, errCertificateRenderFailed
	}
	pdfDate, err := certificatePDFDate(snap.DisplayIssueDate)
	if err != nil {
		return nil, errCertificateRenderFailed
	}
	return assembleCertificatePDF(w, h, c.bytes(), pdfDate, certificatePDFFileID(snapshotSHA, layout, manifestSHA)), nil
}

// ─── Text fitting ────────────────────────────────────────────────────────────

// certificateTextWidth estimates rendered width in points: CJK runs at ~1em,
// ASCII at ~0.55em for the non-embedded UniJIS font.
func certificateTextWidth(s string, size float64) float64 {
	var w float64
	for _, r := range s {
		if r < 0x80 {
			w += size * 0.55
		} else {
			w += size
		}
	}
	return w
}

func certificateWrapText(s string, size, maxW float64) []string {
	var lines []string
	var cur []rune
	var curW float64
	for _, r := range s {
		rw := size
		if r < 0x80 {
			rw = size * 0.55
		}
		if len(cur) > 0 && curW+rw > maxW {
			lines = append(lines, string(cur))
			cur = cur[:0]
			curW = 0
		}
		cur = append(cur, r)
		curW += rw
	}
	if len(cur) > 0 {
		lines = append(lines, string(cur))
	}
	return lines
}

// certificateFitText wraps s inside (w x h) starting at baseline y-h+size,
// shrinking the font down to certificateMinFontSize; render_failed if it still
// does not fit. Never silently truncates.
func certificateFitText(c *pdfCanvas, s string, x, yTop, w, h, startSize float64) error {
	if s == "" {
		return nil
	}
	for size := startSize; size >= certificateMinFontSize; size -= 0.5 {
		lines := certificateWrapText(s, size, w)
		fits := true
		for _, ln := range lines {
			if certificateTextWidth(ln, size) > w {
				fits = false
				break
			}
		}
		if !fits || float64(len(lines))*size*1.3 > h {
			continue
		}
		for i, ln := range lines {
			c.text(x, yTop-size-float64(i)*size*1.3, size, ln)
		}
		return nil
	}
	return errCertificateRenderFailed
}

// certRow is one label/value row inside a bordered section.
type certRow struct {
	label string
	value string
	h     float64 // row height; 0 -> default
}

// drawCertificateTable renders a bordered label|value table and returns the
// bottom Y. labelW is the label column width.
func drawCertificateTable(c *pdfCanvas, x, yTop, w, labelW float64, rows []certRow) (float64, error) {
	y := yTop
	c.hline(x, x+w, y)
	for _, r := range rows {
		rowH := r.h
		if rowH == 0 {
			rowH = 16
		}
		if err := certificateFitText(c, r.label, x+3, y-3, labelW-6, rowH-4, certificateLabelSize); err != nil {
			return 0, err
		}
		if r.value != "" {
			if err := certificateFitText(c, r.value, x+labelW+3, y-3, w-labelW-6, rowH-4, certificateValueSize); err != nil {
				return 0, err
			}
		}
		y -= rowH
		c.hline(x, x+w, y)
	}
	c.vline(x, y, yTop)
	c.vline(x+w, y, yTop)
	c.vline(x+labelW, y, yTop)
	return y, nil
}

// drawCertificateSection draws a bordered block with an optional title and
// free-positioned rows of "label value" lines; returns the bottom Y.
func drawCertificateSection(c *pdfCanvas, x, yTop, w float64, rows []certRow) (float64, error) {
	y := yTop
	c.hline(x, x+w, y)
	for _, r := range rows {
		rowH := r.h
		if rowH == 0 {
			rowH = 16
		}
		if err := certificateFitText(c, r.label, x+4, y-4, w-8, rowH-6, certificateLabelSize); err != nil {
			return 0, err
		}
		if r.value != "" {
			if err := certificateFitText(c, r.value, x+70, y-4, w-74, rowH-6, certificateValueSize); err != nil {
				return 0, err
			}
		}
		y -= rowH
	}
	c.hline(x, x+w, y)
	c.vline(x, y, yTop)
	c.vline(x+w, y, yTop)
	return y, nil
}

// ─── Wareki / formatting ─────────────────────────────────────────────────────

// certificateWareki renders YYYY-MM-DD as a Japanese-era date. Reiwa starts
// 2019-05-01, Heisei 1989-01-08, Showa 1926-12-25; era year 1 is 元年. Earlier
// dates are render_failed (never guessed).
func certificateWareki(date string) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil || !isCertificateDate(date) {
		return "", errCertificateRenderFailed
	}
	var era string
	var year int
	switch {
	case !t.Before(time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC)):
		era, year = "令和", t.Year()-2018
	case !t.Before(time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC)):
		era, year = "平成", t.Year()-1988
	case !t.Before(time.Date(1926, 12, 25, 0, 0, 0, 0, time.UTC)):
		era, year = "昭和", t.Year()-1925
	default:
		return "", errCertificateRenderFailed
	}
	y := strconv.Itoa(year)
	if year == 1 {
		y = "元"
	}
	return fmt.Sprintf("%s%s年%d月%d日", era, y, int(t.Month()), t.Day()), nil
}

// certificateGregorian renders YYYY-MM-DD as 西暦 Y年M月D日 (rabies 接種日).
func certificateGregorian(date string) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil || !isCertificateDate(date) {
		return "", errCertificateRenderFailed
	}
	return fmt.Sprintf("%d年%d月%d日", t.Year(), int(t.Month()), t.Day()), nil
}

// certificateWarekiDate formats a nullable date in wareki; null -> "".
func certificateWarekiDate(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	return certificateWareki(s)
}

// formatCertificateYen renders a JPY amount as "#,###円". Null is blank and 0
// is "0円"; totals are never recomputed.
func formatCertificateYen(v *int64) string {
	if v == nil {
		return ""
	}
	n := *v
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	out := ""
	for len(s) > 3 {
		out = "," + s[len(s)-3:] + out
		s = s[:len(s)-3]
	}
	out = s + out
	if neg {
		out = "-" + out
	}
	return out + "円"
}

func certificateBodySize(s string) string {
	switch s {
	case "small":
		return "小"
	case "medium":
		return "中"
	case "large":
		return "大"
	default:
		return ""
	}
}

func certificateLot(lots []any, i int) string {
	if i < 0 || i >= len(lots) {
		return ""
	}
	s, _ := lots[i].(string)
	return s
}

// ─── Layouts ─────────────────────────────────────────────────────────────────

func drawCertificateNormalPortrait(c *pdfCanvas, snap *certificateSnapshot) error {
	issueDate, err := certificateWareki(snap.DisplayIssueDate)
	if err != nil {
		return err
	}
	administered, err := certificateWareki(certString(snap.Source, "administered_on"))
	if err != nil {
		return err
	}
	nextScheduled, err := certificateWarekiDate(certString(snap.Source, "next_scheduled_on"))
	if err != nil {
		return err
	}
	birthDate, err := certificateWarekiDate(certString(snap.Pet, "birth_date"))
	if err != nil {
		return err
	}
	c.text(400, 812, 9, "発行日 "+issueDate)
	c.text(235, 780, 16, "予防接種証明書")
	y, err := drawCertificateSection(c, 50, 756, 495, []certRow{
		{label: "郵便番号", value: certString(snap.Owner, "postal_code")},
		{label: "住所", value: certString(snap.Owner, "address"), h: 24},
		{label: "電話番号", value: certString(snap.Owner, "phone")},
		{label: "氏名", value: certString(snap.Owner, "name")},
	})
	if err != nil {
		return err
	}
	y, err = drawCertificateSection(c, 50, y-10, 495, []certRow{
		{label: "名前", value: certString(snap.Pet, "name")},
		{label: "動物種", value: certString(snap.Pet, "species")},
		{label: "種別", value: certString(snap.Pet, "breed")},
		{label: "性別", value: certString(snap.Pet, "sex")},
		{label: "生年月日", value: birthDate},
	})
	if err != nil {
		return err
	}
	lots, ok := snap.Source["lots"].([]any)
	if !ok {
		return errCertificateRenderFailed
	}
	y, err = drawCertificateSection(c, 50, y-10, 495, []certRow{
		{label: "接種日", value: administered},
		{label: "種類", value: certString(snap.Source, "vaccine_name")},
		{label: "LOT.1", value: certificateLot(lots, 0)},
		{label: "LOT.2", value: certificateLot(lots, 1)},
		{label: "LOT.3", value: certificateLot(lots, 2)},
		{label: "LOT.4", value: certificateLot(lots, 3)},
		{label: "貼付欄", value: "", h: 48},
		{label: "診察担当", value: certString(snap.Source, "doctor_name")},
		{label: "次回予定日", value: nextScheduled},
	})
	if err != nil {
		return err
	}
	_, err = drawCertificateSection(c, 50, y-10, 495, []certRow{
		{label: "郵便番号", value: certString(snap.Clinic, "postal_code")},
		{label: "住所", value: certString(snap.Clinic, "address"), h: 24},
		{label: "名称", value: certString(snap.Clinic, "name")},
		{label: "院長", value: certString(snap.Clinic, "director_name")},
		{label: "TEL", value: certString(snap.Clinic, "phone")},
		{label: "FAX", value: certString(snap.Clinic, "fax")},
		{label: "URL", value: certString(snap.Clinic, "url")},
	})
	return err
}

func drawCertificateNormalLandscape(c *pdfCanvas, snap *certificateSnapshot) error {
	issueDate, err := certificateWareki(snap.DisplayIssueDate)
	if err != nil {
		return err
	}
	administered, err := certificateWareki(certString(snap.Source, "administered_on"))
	if err != nil {
		return err
	}
	nextScheduled, err := certificateWarekiDate(certString(snap.Source, "next_scheduled_on"))
	if err != nil {
		return err
	}
	birthDate, err := certificateWarekiDate(certString(snap.Pet, "birth_date"))
	if err != nil {
		return err
	}
	c.text(650, 565, 9, "発行日 "+issueDate)
	c.text(330, 540, 16, "予防接種証明書")
	y, err := drawCertificateSection(c, 50, 520, 741, []certRow{
		{label: "郵便番号", value: certString(snap.Owner, "postal_code")},
		{label: "住所", value: certString(snap.Owner, "address"), h: 24},
		{label: "電話番号", value: certString(snap.Owner, "phone")},
		{label: "氏名", value: certString(snap.Owner, "name")},
		{label: "名前", value: certString(snap.Pet, "name")},
		{label: "動物種", value: certString(snap.Pet, "species")},
		{label: "種別", value: certString(snap.Pet, "breed")},
		{label: "性別", value: certString(snap.Pet, "sex")},
		{label: "生年月日", value: birthDate},
	})
	if err != nil {
		return err
	}
	_, err = drawCertificateSection(c, 50, y-14, 360, []certRow{
		{label: "接種日", value: administered},
		{label: "種類", value: certString(snap.Source, "vaccine_name")},
		{label: "貼付欄", value: "", h: 60},
		{label: "診察担当", value: certString(snap.Source, "doctor_name")},
		{label: "次回予定日", value: nextScheduled},
	})
	if err != nil {
		return err
	}
	_, err = drawCertificateSection(c, 430, y-14, 361, []certRow{
		{label: "郵便番号", value: certString(snap.Clinic, "postal_code")},
		{label: "住所", value: certString(snap.Clinic, "address"), h: 24},
		{label: "名称", value: certString(snap.Clinic, "name")},
		{label: "院長", value: certString(snap.Clinic, "director_name")},
		{label: "TEL", value: certString(snap.Clinic, "phone")},
		{label: "FAX", value: certString(snap.Clinic, "fax")},
		{label: "URL", value: certString(snap.Clinic, "url")},
		{label: "E-Mail", value: certString(snap.Clinic, "email")},
	})
	return err
}

func drawCertificateRabies(c *pdfCanvas, snap *certificateSnapshot) error {
	administered, err := certificateGregorian(certString(snap.Source, "administered_on"))
	if err != nil {
		return err
	}
	birthDate, err := certificateWarekiDate(certString(snap.Pet, "birth_date"))
	if err != nil {
		return err
	}
	if age := certString(snap.Pet, "age_text"); age != "" {
		if birthDate != "" {
			birthDate += "（" + age + "）"
		} else {
			birthDate = age
		}
	}
	c.text(320, 562, 16, "狂犬病予防注射済証")
	c.text(640, 562, 10, "証明番号 "+certString(snap.Inputs, "certificate_number"))
	c.text(54, 548, 7.5, "所有者（管理者）")
	y, err := drawCertificateTable(c, 50, 542, 500, 100, []certRow{
		{label: "住所", value: certString(snap.Owner, "address"), h: 24},
		{label: "電話番号", value: certString(snap.Owner, "phone")},
		{label: "ふりがな", value: certString(snap.Owner, "name_kana")},
		{label: "氏名", value: certString(snap.Owner, "name")},
	})
	if err != nil {
		return err
	}
	c.text(54, y+4, 7.5, "犬情報")
	y, err = drawCertificateTable(c, 50, y-2, 500, 100, []certRow{
		{label: "種類", value: certString(snap.Pet, "breed")},
		{label: "毛色", value: certString(snap.Pet, "coat_color")},
		{label: "性別", value: certString(snap.Pet, "sex")},
		{label: "名前", value: certString(snap.Pet, "name")},
		{label: "生年月日", value: birthDate},
		{label: "体格", value: certificateBodySize(certString(snap.Inputs, "body_size"))},
		{label: "特徴", value: certString(snap.Inputs, "distinguishing_features"), h: 24},
		{label: "注射量", value: certString(snap.Inputs, "injection_amount_text")},
		{label: "注射済票番号", value: certString(snap.Inputs, "injection_tag_number")},
		{label: "犬鑑札番号", value: certString(snap.Inputs, "dog_registration_tag_number")},
	})
	if err != nil {
		return err
	}
	statement := fmt.Sprintf("上記の犬に、%sに狂犬病予防注射を実施したことを証明する。", administered)
	y, err = drawCertificateSection(c, 50, y-8, 500, []certRow{
		{label: statement, h: 22},
	})
	if err != nil {
		return err
	}
	y, err = drawCertificateTable(c, 50, y-8, 500, 120, []certRow{
		{label: "登録手数料", value: formatCertificateYen(certInt64Ptr(snap.Inputs, "registration_fee"))},
		{label: "注射料", value: formatCertificateYen(certInt64Ptr(snap.Inputs, "injection_fee"))},
		{label: "注射済票費用", value: formatCertificateYen(certInt64Ptr(snap.Inputs, "tag_fee"))},
		{label: "合計", value: formatCertificateYen(certInt64Ptr(snap.Inputs, "total"))},
	})
	if err != nil {
		return err
	}
	_, err = drawCertificateSection(c, 570, 540, 221, []certRow{
		{label: "獣医師会", value: certString(snap.Clinic, "veterinarian_association"), h: 24},
		{label: "指定獣医師", value: certString(snap.Clinic, "designated_veterinarian"), h: 24},
		{label: "住所", value: certString(snap.Clinic, "address"), h: 32},
		{label: "名称", value: certString(snap.Clinic, "name")},
		{label: "TEL", value: certString(snap.Clinic, "phone")},
		{label: "FAX", value: certString(snap.Clinic, "fax")},
		{label: "獣医師氏名", value: certString(snap.Clinic, "veterinarian_name")},
		{label: "印", value: "", h: 40},
	})
	return err
}
