package medicalrecord

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

// synthIDEXXShortFrame builds a short IDEXX status ping frame (body <= idexxShortFrameBodyMaxBytes).
// Spec (LAB_DEVICE_CONNECTIVITY.md §IDEXX): "02 31 30 … 03" — approximately 7 bytes total.
func synthIDEXXShortFrame() []byte {
	body := []byte("10   ") // 5 bytes — matches "02 31 30 … 03" ping pattern from spec
	var buf bytes.Buffer
	buf.WriteByte(0x02)
	buf.Write(body)
	buf.WriteByte(0x03)
	return buf.Bytes()
}

// synthIDEXXLongFrame builds a synthetic IDEXX VetLab long frame containing the
// blood-count labels listed in LAB_DEVICE_CONNECTIVITY.md §IDEXX.
func synthIDEXXLongFrame() []byte {
	lines := []string{
		"WBC  5.50  K/uL",
		"RBC  6.20  M/uL",
		"HCT  38.5  %",
		"HGB  12.8  g/dL",
		"PLT  312   K/uL",
		"NEU  3.80  K/uL",
		"LYM  1.20  K/uL",
		"MONO 0.30  K/uL",
		"EOS  0.10  K/uL",
		"BASO 0.05  K/uL",
		"RETIC 0.20 K/uL",
	}
	body := strings.Join(lines, "\n")
	var buf bytes.Buffer
	buf.WriteByte(0x02)
	buf.WriteString(body)
	buf.WriteByte(0x03)
	return buf.Bytes()
}

// TestDecodeIDEXXFrames_ShortFrameOnly verifies that a payload containing only short
// status frames (spec: "約7バイト ping") is rejected as invalid.
func TestDecodeIDEXXFrames_ShortFrameOnly(t *testing.T) {
	t.Parallel()
	payload := synthIDEXXShortFrame()
	_, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) {
		t.Fatalf("short-only payload must be invalid_payload, got %v", err)
	}
}

// TestDecodeIDEXXFrames_ShortThenLong verifies that a short ping followed by a long frame
// discards the ping and returns exactly one measurement.
func TestDecodeIDEXXFrames_ShortThenLong(t *testing.T) {
	t.Parallel()
	payload := append(append([]byte{}, synthIDEXXShortFrame()...), synthIDEXXLongFrame()...)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatalf("short+long: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: got %d want 1", len(frames))
	}
}

// TestDecodeIDEXXFrames_AllKnownCodes verifies all 11 expected blood-count codes are decoded.
func TestDecodeIDEXXFrames_AllKnownCodes(t *testing.T) {
	t.Parallel()
	frames, err := DecodeLabDeviceFrames(synthIDEXXLongFrame(), "idexx_vetlab")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: got %d want 1", len(frames))
	}
	f := frames[0]
	if f.SourceType != model.LabImportSourceTypeIDEXXVetLab {
		t.Errorf("source_type: got %s", f.SourceType)
	}
	if f.DeviceHint != "VetLab" {
		t.Errorf("device_hint: got %s", f.DeviceHint)
	}
	wantCodes := []string{"WBC", "RBC", "HCT", "HGB", "PLT", "NEU", "LYM", "MONO", "EOS", "BASO", "RETIC"}
	if len(f.Items) != len(wantCodes) {
		t.Fatalf("item count: got %d want %d", len(f.Items), len(wantCodes))
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
	if byCode["WBC"].ValueRaw != "5.50" || byCode["WBC"].Unit != "K/uL" {
		t.Errorf("WBC: %+v", byCode["WBC"])
	}
	if byCode["HCT"].ValueRaw != "38.5" || byCode["HCT"].Unit != "%" {
		t.Errorf("HCT: %+v", byCode["HCT"])
	}
	if byCode["HGB"].ValueRaw != "12.8" || byCode["HGB"].Unit != "g/dL" {
		t.Errorf("HGB: %+v", byCode["HGB"])
	}
	if f.NeedsReview {
		t.Errorf("all-known frame must not need review, warnings=%v", f.Warnings)
	}
}

// TestDecodeIDEXXFrames_FingerprintStable verifies fingerprint stability across identical payloads.
func TestDecodeIDEXXFrames_FingerprintStable(t *testing.T) {
	t.Parallel()
	payload := synthIDEXXLongFrame()
	a, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatal(err)
	}
	if a[0].SourceFingerprint != b[0].SourceFingerprint {
		t.Fatal("same bytes must yield the same fingerprint")
	}
	if a[0].SourceFingerprint == "" {
		t.Fatal("fingerprint must not be empty")
	}
}

// TestDecodeIDEXXFrames_RepeatedLong verifies that two identical long frames produce
// two frame results with the same fingerprint (service-level dedup handles eliminating dupes).
func TestDecodeIDEXXFrames_RepeatedLong(t *testing.T) {
	t.Parallel()
	long := synthIDEXXLongFrame()
	payload := append(append([]byte{}, long...), long...)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatalf("repeated long: %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("frames: got %d want 2", len(frames))
	}
	if frames[0].SourceFingerprint != frames[1].SourceFingerprint {
		t.Errorf("repeated frames must share fingerprint for service-level dedup")
	}
}

// TestDecodeIDEXXFrames_EmptyBody verifies that a frame with no usable content is invalid.
func TestDecodeIDEXXFrames_EmptyBody(t *testing.T) {
	t.Parallel()
	// STX immediately followed by ETX — zero-byte body
	_, err := DecodeLabDeviceFrames([]byte{0x02, 0x03}, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) {
		t.Fatalf("zero-body frame must be invalid_payload, got %v", err)
	}
}

// TestDecodeIDEXXFrames_HintAlias verifies the "VetLab" hint alias routes to the IDEXX decoder.
func TestDecodeIDEXXFrames_HintAlias(t *testing.T) {
	t.Parallel()
	frames, err := DecodeLabDeviceFrames(synthIDEXXLongFrame(), "VetLab")
	if err != nil {
		t.Fatalf("VetLab hint: %v", err)
	}
	if frames[0].SourceType != model.LabImportSourceTypeIDEXXVetLab {
		t.Errorf("source_type: got %s", frames[0].SourceType)
	}
}

// TestDecodeIDEXXFrames_WBCWithoutUnit verifies that a value without a recognized unit is
// still extracted (unit left empty).
func TestDecodeIDEXXFrames_WBCWithoutUnit(t *testing.T) {
	t.Parallel()
	// Body exceeds short-frame threshold; WBC value present but no unit token
	body := strings.Repeat(" ", 20) + "WBC 5.50"
	var buf bytes.Buffer
	buf.WriteByte(0x02)
	buf.WriteString(body)
	buf.WriteByte(0x03)
	frames, err := DecodeLabDeviceFrames(buf.Bytes(), "idexx_vetlab")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 || len(frames[0].Items) == 0 {
		t.Fatalf("expected one frame with one item")
	}
	it := frames[0].Items[0]
	if it.Code != "WBC" || it.ValueRaw != "5.50" {
		t.Errorf("WBC without unit: %+v", it)
	}
	if it.Unit != "" {
		t.Errorf("unit should be empty, got %q", it.Unit)
	}
}
