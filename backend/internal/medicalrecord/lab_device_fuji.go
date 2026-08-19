package medicalrecord

import (
	"strings"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	fujiSTX        = 0x02
	fujiETX        = 0x03
	fujiHeaderLen  = 50
	fujiItemSlot   = 36
	fujiStatusLen  = 7
	fujiDateStart  = 7
	fujiDateLen    = 10
	fujiClockStart = 17
	fujiClockLen   = 5
	fujiSpecStart  = 35
	fujiSpecLen    = 13
)

func decodeFujiFrames(payload []byte, hint string) ([]LabDeviceFrame, error) {
	chunks := splitFujiFrames(payload)
	if len(chunks) == 0 {
		return nil, ErrInvalidLabDevicePayload
	}
	frames := make([]LabDeviceFrame, 0, len(chunks))
	for _, chunk := range chunks {
		frame, err := decodeFujiFrame(chunk, hint)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

func splitFujiFrames(payload []byte) [][]byte {
	var out [][]byte
	i := 0
	for i < len(payload) {
		if payload[i] != fujiSTX {
			i++
			continue
		}
		end := indexByteFrom(payload, fujiETX, i+1)
		if end < 0 {
			return nil
		}
		out = append(out, payload[i:end+1])
		i = end + 1
	}
	return out
}

func indexByteFrom(b []byte, c byte, start int) int {
	for i := start; i < len(b); i++ {
		if b[i] == c {
			return i
		}
	}
	return -1
}

func decodeFujiFrame(frame []byte, hint string) (LabDeviceFrame, error) {
	if len(frame) < 2+fujiHeaderLen || frame[0] != fujiSTX || frame[len(frame)-1] != fujiETX {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}
	body := frame[1 : len(frame)-1]
	header := body[:fujiHeaderLen]
	status := strings.TrimSpace(string(header[:fujiStatusLen]))
	date := string(header[fujiDateStart : fujiDateStart+fujiDateLen])
	clock := string(header[fujiClockStart : fujiClockStart+fujiClockLen])
	measuredAt, err := parseLabDeviceClock("2006-01-02", date, "15:04", clock)
	if err != nil {
		return LabDeviceFrame{}, err
	}

	items := parseFujiItemSlots(body[fujiHeaderLen:])
	if len(items) == 0 {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}

	sourceType, deviceHint := classifyFuji(hint, items)
	out := LabDeviceFrame{
		SourceType:        sourceType,
		SourceFingerprint: labDeviceFingerprint(frame),
		MeasuredAt:        measuredAt,
		SpecimenIDRaw:     decodeSpecimenRaw(header[fujiSpecStart : fujiSpecStart+fujiSpecLen]),
		DeviceHint:        deviceHint,
		Items:             items,
	}
	if status != "NORMAL" {
		out.Warnings = append(out.Warnings, "status:"+status)
		out.NeedsReview = true
	}
	known := knownFujiNX600Codes
	if sourceType == model.LabImportSourceTypeFujiAU10V {
		known = knownFujiAU10VCodes
	}
	markUnknownCodes(&out, known)
	for _, it := range items {
		if it.Flag != "" {
			out.Warnings = append(out.Warnings, "flag:"+it.Flag)
		}
	}
	return out, nil
}

func classifyFuji(hint string, items []LabDeviceItem) (model.LabImportSourceType, string) {
	if hint == "au10v" {
		return model.LabImportSourceTypeFujiAU10V, "AU10V"
	}
	if hint == "nx600" {
		return model.LabImportSourceTypeFujiNX600, "NX600"
	}
	for _, it := range items {
		if it.Code == "vf-SAA" {
			return model.LabImportSourceTypeFujiAU10V, "AU10V"
		}
	}
	return model.LabImportSourceTypeFujiNX600, "NX600"
}

func parseFujiItemSlots(body []byte) []LabDeviceItem {
	var items []LabDeviceItem
	for off := 0; off+fujiItemSlot <= len(body); off += fujiItemSlot {
		item, ok := parseFujiItemSlot(body[off : off+fujiItemSlot])
		if ok {
			items = append(items, item)
		}
	}
	return items
}

func parseFujiItemSlot(slot []byte) (LabDeviceItem, bool) {
	raw := strings.TrimRight(string(slot), " ")
	if strings.TrimSpace(raw) == "" {
		return LabDeviceItem{}, false
	}
	flag := ""
	body := raw
	if idx := strings.LastIndex(raw, "01"); idx >= 0 {
		flag = strings.TrimSpace(raw[idx+2:])
		body = strings.TrimSpace(raw[:idx])
	}
	if body == "" {
		return LabDeviceItem{}, false
	}
	item := LabDeviceItem{Flag: flag}
	if eq := strings.IndexByte(body, '='); eq >= 0 {
		item.Code = strings.TrimSpace(body[:eq])
		applyFujiValueUnit(&item, strings.Fields(strings.TrimSpace(body[eq+1:])))
	} else {
		fields := strings.Fields(body)
		if len(fields) == 0 {
			return LabDeviceItem{}, false
		}
		item.Code = fields[0]
		applyFujiValueUnit(&item, fields[1:])
	}
	if item.Code == "" || item.ValueRaw == "" {
		return LabDeviceItem{}, false
	}
	return item, true
}

func applyFujiValueUnit(item *LabDeviceItem, fields []string) {
	unitAt := -1
	for i, tok := range fields {
		if isLabDeviceUnit(tok) {
			unitAt = i
			break
		}
	}
	if unitAt >= 0 {
		item.ValueRaw = strings.Join(fields[:unitAt], " ")
		item.Unit = fields[unitAt]
		extra := strings.TrimSpace(strings.Join(fields[unitAt+1:], " "))
		item.Flag = strings.TrimSpace(strings.Join([]string{extra, item.Flag}, " "))
		return
	}
	item.ValueRaw = strings.Join(fields, " ")
}
