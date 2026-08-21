package medicalrecord

import (
	"strings"

	"github.com/animal-ekarte/backend/internal/model"
)

// idexxShortFrameBodyMaxBytes is the body-length threshold below which a frame is
// considered a short status ping (spec: "約7バイト" total, so body ≤ ~5 bytes).
// Using 16 as a conservative upper bound to safely discard all status frames.
const idexxShortFrameBodyMaxBytes = 16

// isIDEXXUnit reports whether s looks like a unit token in VetLab output.
// Extends the general "/" check with flat units observed in the spec ("fL", "pg", "%").
func isIDEXXUnit(s string) bool {
	if strings.Contains(s, "/") {
		return true
	}
	switch s {
	case "fL", "pg", "%":
		return true
	}
	return false
}

// decodeIDEXXFrames processes a raw serial payload from the IDEXX VetLab Station PIMS port.
//
// Protocol facts (LAB_DEVICE_CONNECTIVITY.md §IDEXX):
//   - 9600 8N1, cu.usbserial (COM5)
//   - STX (0x02) … ETX (0x03) framing; no CR/LF between delimiters
//   - Short frames (body ≤ idexxShortFrameBodyMaxBytes) are status pings → discard
//   - Long frames (~2 KB) contain blood-count labels with values and units
//   - The same long frame repeats until the port is released → deduplicate by fingerprint
//   - Session replies (ACK/A/IM/SM) are built in lab_device_idexx_pims.go but not
//     written to the serial port until Source/Port are confirmed
//   - No ASTM/HL7 protocol is assumed
func decodeIDEXXFrames(payload []byte) ([]LabDeviceFrame, error) {
	// Reuse STX/ETX splitter from lab_device_fuji.go
	chunks := splitFujiFrames(payload)
	var frames []LabDeviceFrame
	for _, chunk := range chunks {
		// chunk includes the STX and ETX bytes; body is between them
		if len(chunk) < 2 {
			continue
		}
		body := chunk[1 : len(chunk)-1]
		if len(body) <= idexxShortFrameBodyMaxBytes || looksLikeIDEXXPIMSControlFrame(chunk) {
			// I/s session inquiries (and A/SM replies) are not measurements.
			// Port inquiry I is 21 bytes and exceeds the short-body cutoff.
			continue
		}
		frame, err := decodeIDEXXLongFrame(body, chunk)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	if len(frames) == 0 {
		return nil, ErrInvalidLabDevicePayload
	}
	return frames, nil
}

// decodeIDEXXLongFrame extracts blood-count items from one long IDEXX frame body.
// rawFrame is the full STX…ETX bytes used for fingerprinting.
func decodeIDEXXLongFrame(body, rawFrame []byte) (LabDeviceFrame, error) {
	// Normalize: replace control bytes with spaces, keep printable ASCII
	text := make([]byte, 0, len(body))
	for _, b := range body {
		if b >= 0x20 && b <= 0x7E {
			text = append(text, b)
		} else {
			text = append(text, ' ')
		}
	}

	tokens := strings.Fields(string(text))
	items := parseIDEXXTokens(tokens)

	if len(items) == 0 {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}

	out := LabDeviceFrame{
		SourceType:        model.LabImportSourceTypeIDEXXVetLab,
		SourceFingerprint: labDeviceFingerprint(rawFrame),
		DeviceHint:        "VetLab",
		Items:             items,
	}
	markUnknownCodes(&out, knownIDEXXCodes)
	return out, nil
}

// parseIDEXXTokens scans a token slice for known IDEXX codes and extracts CODE VALUE [UNIT].
// Unknown tokens are skipped; unknown codes in the result are flagged by markUnknownCodes.
func parseIDEXXTokens(tokens []string) []LabDeviceItem {
	var items []LabDeviceItem
	for i := 0; i < len(tokens); i++ {
		code := tokens[i]
		if _, isKnown := knownIDEXXCodes[code]; !isKnown {
			continue
		}
		if i+1 >= len(tokens) {
			break
		}
		value := tokens[i+1]
		unit := ""
		if i+2 < len(tokens) && isIDEXXUnit(tokens[i+2]) {
			unit = tokens[i+2]
			i += 2
		} else {
			i += 1
		}
		items = append(items, LabDeviceItem{
			Code:     code,
			ValueRaw: value,
			Unit:     unit,
		})
	}
	return items
}
