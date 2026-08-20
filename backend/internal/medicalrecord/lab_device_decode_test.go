package medicalrecord

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestDecodeLabDeviceFrames_NX600Synthetic(t *testing.T) {
	t.Parallel()
	payload := synthFujiNX600()
	frames, err := DecodeLabDeviceFrames(payload, "")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: got %d want 1", len(frames))
	}
	f := frames[0]
	if f.SourceType != model.LabImportSourceTypeFujiNX600 {
		t.Errorf("source_type: got %s", f.SourceType)
	}
	if f.DeviceHint != "NX600" {
		t.Errorf("device_hint: got %s", f.DeviceHint)
	}
	if f.SpecimenIDRaw != "TEST1" {
		t.Errorf("specimen: got %q", f.SpecimenIDRaw)
	}
	if f.MeasuredAt == nil || !f.MeasuredAt.Equal(jstTime(t, "2026-01-15 10:30")) {
		t.Errorf("measured_at: got %v", f.MeasuredAt)
	}
	wantCodes := []string{
		"Na-P", "K-P", "Cl-P", "LIP-P", "TP-P", "ALB-P", "ALPi-P", "GLU-P",
		"TBIL-P", "IP-P", "TCHO-P", "GGT-P", "GPT-P", "Ca-P", "CRE-P", "BUN-P",
	}
	if len(f.Items) != len(wantCodes) {
		t.Fatalf("items: got %d want %d", len(f.Items), len(wantCodes))
	}
	byCode := map[string]LabDeviceItem{}
	for _, it := range f.Items {
		byCode[it.Code] = it
	}
	for _, code := range wantCodes {
		if _, ok := byCode[code]; !ok {
			t.Errorf("missing code %s", code)
		}
	}
	if byCode["Na-P"].ValueRaw != "151" || byCode["Na-P"].Unit != "mEq/l" {
		t.Errorf("Na-P: %+v", byCode["Na-P"])
	}
	if byCode["CRE-P"].ValueRaw != "2.30" {
		t.Errorf("CRE-P value: %q", byCode["CRE-P"].ValueRaw)
	}
	if byCode["ALPi-P"].Flag != "@" {
		t.Errorf("ALPi-P flag: %q", byCode["ALPi-P"].Flag)
	}
	if f.SourceFingerprint != sha256Hex(payload) {
		t.Errorf("fingerprint mismatch")
	}
}

func TestDecodeLabDeviceFrames_AU10VInequalityAndResend(t *testing.T) {
	t.Parallel()
	one := synthFujiAU10V()
	frames, err := DecodeLabDeviceFrames(one, "AU10V")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: got %d", len(frames))
	}
	f := frames[0]
	if f.SourceType != model.LabImportSourceTypeFujiAU10V {
		t.Errorf("source_type: got %s", f.SourceType)
	}
	if len(f.Items) != 1 || f.Items[0].Code != "vf-SAA" {
		t.Fatalf("items: %+v", f.Items)
	}
	if f.Items[0].ValueRaw != "<3.75" || f.Items[0].Unit != "ug/mL" {
		t.Errorf("vf-SAA: %+v", f.Items[0])
	}
	if !strings.Contains(f.Items[0].Flag, "@") || !strings.Contains(f.Items[0].Flag, "D") {
		t.Errorf("flag should keep @ and D: %q", f.Items[0].Flag)
	}

	twice := append(append([]byte{}, one...), one...)
	got, err := DecodeLabDeviceFrames(twice, "")
	if err != nil {
		t.Fatalf("double frame: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("double frame count: %d", len(got))
	}
	if got[0].SourceFingerprint != got[1].SourceFingerprint {
		t.Errorf("resend fingerprints differ")
	}
}

func TestDecodeLabDeviceFrames_UrineBit7AndDashValues(t *testing.T) {
	t.Parallel()
	ascii := synthUrinePU4010()
	raw := setHighBit(ascii)
	frames, err := DecodeLabDeviceFrames(raw, "")
	if err != nil {
		t.Fatalf("decode high-bit urine: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("urine should be one measurement, got %d", len(frames))
	}
	f := frames[0]
	if f.SourceType != model.LabImportSourceTypeArkrayPU4010 {
		t.Errorf("source_type: got %s", f.SourceType)
	}
	if f.MeasuredAt == nil || !f.MeasuredAt.Equal(jstTime(t, "2026-01-15 11:51")) {
		t.Errorf("measured_at: got %v", f.MeasuredAt)
	}
	if f.SpecimenIDRaw != "" {
		t.Errorf("dash specimen should be empty, got %q", f.SpecimenIDRaw)
	}
	byCode := map[string]LabDeviceItem{}
	for _, it := range f.Items {
		byCode[it.Code] = it
	}
	want := []string{"GLU", "PRO", "BIL", "URO", "PH", "BLD", "KET", "NIT"}
	for _, code := range want {
		if _, ok := byCode[code]; !ok {
			t.Errorf("missing %s", code)
		}
	}
	if byCode["GLU"].ValueRaw != "-" || byCode["GLU"].Unit != "mg/dL" {
		t.Errorf("GLU: %+v", byCode["GLU"])
	}
	if byCode["PRO"].ValueRaw != "*3+ 300" {
		t.Errorf("PRO should keep qual and num, got %q", byCode["PRO"].ValueRaw)
	}
	if byCode["URO"].ValueRaw != "NORMAL" {
		t.Errorf("URO: %q", byCode["URO"].ValueRaw)
	}
	if byCode["PH"].ValueRaw != "9.0" {
		t.Errorf("PH: %+v", byCode["PH"])
	}
	joined := strings.Join(f.Warnings, " ")
	if !strings.Contains(joined, "T.26") && !strings.Contains(joined, "COM.") {
		t.Errorf("trailer should be a warning, got %v", f.Warnings)
	}
	if f.SourceFingerprint != sha256Hex(ascii) {
		t.Errorf("urine fingerprint must be SHA-256 of bit7-stripped bytes")
	}
}

func TestDecodeLabDeviceFrames_UrineWithoutBit7AsFujiFails(t *testing.T) {
	t.Parallel()
	raw := setHighBit(synthUrinePU4010())
	if bytes.Contains(raw, []byte("VPU-4010")) {
		t.Fatal("high-bit urine must not contain ASCII VPU-4010")
	}
	_, err := DecodeLabDeviceFrames(raw, "fuji_nx600")
	if !errors.Is(err, ErrInvalidLabDevicePayload) {
		t.Fatalf("urine read without bit7 strip must be invalid_payload, got %v", err)
	}
}

func TestDecodeLabDeviceFrames_FingerprintStable(t *testing.T) {
	t.Parallel()
	payload := synthFujiAU10V()
	a, err := DecodeLabDeviceFrames(payload, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeLabDeviceFrames(payload, "")
	if err != nil {
		t.Fatal(err)
	}
	if a[0].SourceFingerprint != b[0].SourceFingerprint {
		t.Fatal("same bytes must yield the same fingerprint")
	}
}

func TestDecodeLabDeviceFrames_InvalidCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		payload []byte
		hint    string
	}{
		{"empty", nil, ""},
		{"too_large", bytes.Repeat([]byte{0x02}, 9000), ""},
		{"no_stx", []byte("NORMAL 2026-01-15"), ""},
		{"fuji_bad_date", synthFujiHeaderOnly("NORMAL ", "XXXXXXXXXX", "10:00", "TEST1", nil), ""},
		{"urine_no_vpu", []byte{0x02, 'X', 0x03}, "arkray_pu4010"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeLabDeviceFrames(tc.payload, tc.hint)
			if !errors.Is(err, ErrInvalidLabDevicePayload) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestDecodeLabDeviceFrames_UnknownCodeWarning(t *testing.T) {
	t.Parallel()
	item := fujiEqualsSlot("ZZZ-P", "1", "U/l", "")
	payload := synthFujiHeaderOnly("NORMAL ", "2026-01-15", "10:00", "TEST1", [][]byte{item})
	frames, err := DecodeLabDeviceFrames(payload, "NX600")
	if err != nil {
		t.Fatal(err)
	}
	if !frames[0].NeedsReview {
		t.Fatal("unknown code should need review")
	}
	found := false
	for _, w := range frames[0].Warnings {
		if strings.Contains(w, "ZZZ-P") {
			found = true
		}
	}
	if !found {
		t.Errorf("warnings: %v", frames[0].Warnings)
	}
}

// TestDecodeLabDeviceFrames_IDEXXHintRoutes verifies that "idexx_vetlab" hint reaches the IDEXX decoder.
func TestDecodeLabDeviceFrames_IDEXXHintRoutes(t *testing.T) {
	t.Parallel()
	frames, err := DecodeLabDeviceFrames(synthIDEXXLongFrame(), "idexx_vetlab")
	if err != nil {
		t.Fatalf("idexx_vetlab hint: %v", err)
	}
	if len(frames) == 0 {
		t.Fatal("no frames returned")
	}
}

func TestDecodeLabDeviceFrames_NonNormalStatus(t *testing.T) {
	t.Parallel()
	item := fujiEqualsSlot("Na-P", "1", "mEq/l", "")
	payload := synthFujiHeaderOnly("ERROR  ", "2026-01-15", "10:00", "TEST1", [][]byte{item})
	frames, err := DecodeLabDeviceFrames(payload, "NX600")
	if err != nil {
		t.Fatal(err)
	}
	if !frames[0].NeedsReview {
		t.Fatal("non-NORMAL status should need review")
	}
}

func jstTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, labDeviceLocation)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func setHighBit(in []byte) []byte {
	out := make([]byte, len(in))
	for i, b := range in {
		out[i] = b | 0x80
	}
	return out
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func fujiEqualsSlot(code, value, unit, flag string) []byte {
	slot := bytes.Repeat([]byte(" "), 36)
	copy(slot[0:7], padRight(code, 7))
	slot[7] = '='
	copy(slot[8:17], padRight(value, 9))
	copy(slot[17:23], padRight(unit, 6))
	copy(slot[23:25], "01")
	if flag != "" {
		copy(slot[26:36], padRight(flag, 10))
	}
	return slot
}

func synthFujiHeaderOnly(status7, date10, clock5, specimen string, items [][]byte) []byte {
	header := bytes.Repeat([]byte(" "), 50)
	copy(header[0:7], padRight(status7, 7))
	copy(header[7:17], padRight(date10, 10))
	copy(header[17:22], padRight(clock5, 5))
	copy(header[22:25], "1  ")
	copy(header[35:48], padRight(specimen, 13))
	copy(header[48:50], "01")
	var buf bytes.Buffer
	buf.WriteByte(0x02)
	buf.Write(header)
	for _, item := range items {
		buf.Write(item)
	}
	buf.WriteByte(0x03)
	return buf.Bytes()
}

func synthFujiNX600() []byte {
	type row struct{ code, value, unit, flag string }
	rows := []row{
		{"Na-P", "151", "mEq/l", ""},
		{"K-P", "5.4", "mEq/l", ""},
		{"Cl-P", "121", "mEq/l", ""},
		{"LIP-P", "24", "U/l", ""},
		{"TP-P", "6.4", "g/dl", ""},
		{"ALB-P", "2.9", "g/dl", ""},
		{"ALPi-P", "17", "U/l", "@"},
		{"GLU-P", "237", "mg/dl", ""},
		{"TBIL-P", "0.2", "mg/dl", ""},
		{"IP-P", "4.5", "mg/dl", ""},
		{"TCHO-P", "163", "mg/dl", ""},
		{"GGT-P", "3", "U/l", "@"},
		{"GPT-P", "58", "U/l", ""},
		{"Ca-P", "9.7", "mg/dl", ""},
		{"CRE-P", "2.30", "mg/dl", ""},
		{"BUN-P", "82.6", "mg/dl", ""},
	}
	items := make([][]byte, 0, len(rows))
	for _, r := range rows {
		items = append(items, fujiEqualsSlot(r.code, r.value, r.unit, r.flag))
	}
	return synthFujiHeaderOnly("NORMAL ", "2026-01-15", "10:30", "TEST1", items)
}

func synthFujiAU10V() []byte {
	item := []byte("vf-SAA <3.75     ug/mL 10 @     D   ")
	if len(item) != 36 {
		panic("au item slot must be 36 bytes")
	}
	return synthFujiHeaderOnly("NORMAL ", "2026-01-15", "16:33", "TEST1", [][]byte{item})
}

func synthUrinePU4010() []byte {
	header := padRight("VPU-4010  V01.06 B-------------D2026/01/15T11:51N01.8062P   -  S8EA", 74)
	items := []string{
		"GLU    -           mg/dL ",
		"PRO   *3+    300   mg/dL ",
		"BIL    -           mg/dL ",
		"URO    NORMAL      mg/dL ",
		"PH           9.0         ",
		"BLD   *2+    0.2   mg/dL ",
		"KET    -           mg/dL ",
		"NIT    -                 ",
	}
	for _, it := range items {
		if len(it) != 25 {
			panic("urine item must be 25 bytes: " + it)
		}
	}
	var buf bytes.Buffer
	buf.WriteByte(0x02)
	buf.WriteString(header)
	buf.WriteByte(0x17)
	buf.WriteByte(0x02)
	for _, it := range items {
		buf.WriteString(it)
	}
	buf.WriteByte(0x17)
	buf.WriteByte(0x02)
	buf.WriteString(padRight("T.26  COM.", 67))
	buf.WriteByte(0x03)
	return buf.Bytes()
}
