package medicalrecord

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

const labDeviceMaxPayloadBytes = 8 * 1024

var labDeviceLocation = time.FixedZone("JST", 9*3600)

var knownFujiNX600Codes = map[string]struct{}{
	"Na-P": {}, "K-P": {}, "Cl-P": {}, "LIP-P": {}, "TP-P": {}, "ALB-P": {},
	"ALPi-P": {}, "GLU-P": {}, "TBIL-P": {}, "IP-P": {}, "TCHO-P": {},
	"GGT-P": {}, "GPT-P": {}, "Ca-P": {}, "CRE-P": {}, "BUN-P": {},
}

var knownFujiAU10VCodes = map[string]struct{}{
	"vf-SAA": {},
}

var knownUrineCodes = map[string]struct{}{
	"GLU": {}, "PRO": {}, "BIL": {}, "URO": {}, "PH": {}, "BLD": {}, "KET": {}, "NIT": {},
}

// DecodeLabDeviceFrames turns serial bytes into measurements.
// hint may be empty (auto), NX600 / fuji_nx600, AU10V / fuji_au10v, or PU-4010 / arkray_pu4010.
func DecodeLabDeviceFrames(payload []byte, hint string) ([]LabDeviceFrame, error) {
	if len(payload) == 0 || len(payload) > labDeviceMaxPayloadBytes {
		return nil, ErrInvalidLabDevicePayload
	}
	kind := normalizeLabDeviceHint(hint)
	stripped := stripLabDeviceBit7(payload)
	if kind == "urine" || (kind == "" && bytes.Contains(stripped, []byte("VPU-4010"))) {
		frame, err := decodeUrineMeasurement(stripped)
		if err != nil {
			return nil, err
		}
		return []LabDeviceFrame{frame}, nil
	}
	return decodeFujiFrames(payload, kind)
}

func normalizeLabDeviceHint(hint string) string {
	switch strings.ToLower(strings.TrimSpace(hint)) {
	case "", "auto":
		return ""
	case "nx600", "fuji_nx600":
		return "nx600"
	case "au10v", "fuji_au10v":
		return "au10v"
	case "pu-4010", "pu4010", "arkray_pu4010":
		return "urine"
	default:
		return strings.ToLower(strings.TrimSpace(hint))
	}
}

func stripLabDeviceBit7(in []byte) []byte {
	out := make([]byte, len(in))
	for i, b := range in {
		out[i] = b & 0x7F
	}
	return out
}

func labDeviceFingerprint(normalized []byte) string {
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:])
}

func isLabDeviceUnit(tok string) bool {
	return strings.Contains(tok, "/")
}

func parseLabDeviceClock(dateLayout, date, clockLayout, clock string) (*time.Time, error) {
	parsed, err := time.ParseInLocation(dateLayout+" "+clockLayout, date+" "+clock, labDeviceLocation)
	if err != nil {
		return nil, ErrInvalidLabDevicePayload
	}
	return &parsed, nil
}

func decodeSpecimenRaw(raw []byte) string {
	var b strings.Builder
	for _, c := range bytes.TrimSpace(raw) {
		switch {
		case c >= 0x20 && c <= 0x7E:
			b.WriteByte(c)
		case c >= 0xA1 && c <= 0xDF:
			b.WriteRune(rune(0xFF61 + int(c) - 0xA1))
		}
	}
	return strings.TrimSpace(b.String())
}

func markUnknownCodes(frame *LabDeviceFrame, known map[string]struct{}) {
	for _, it := range frame.Items {
		if _, ok := known[it.Code]; !ok {
			frame.Warnings = append(frame.Warnings, "unknown_code:"+it.Code)
			frame.NeedsReview = true
		}
	}
}
