package medicalrecord

import (
	"bytes"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	urineSTX     = 0x02
	urineETX     = 0x03
	urineETB     = 0x17
	urineItemLen = 25
)

func decodeUrineMeasurement(stripped []byte) (LabDeviceFrame, error) {
	if !bytes.Contains(stripped, []byte("VPU-4010")) {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}
	start := bytes.IndexByte(stripped, urineSTX)
	end := bytes.LastIndexByte(stripped, urineETX)
	if start < 0 || end <= start {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}
	meas := stripped[start : end+1]
	records := splitUrineRecords(meas)
	if len(records) == 0 {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}

	measuredAt, specimen, err := parseUrineHeader(records[0])
	if err != nil {
		return LabDeviceFrame{}, err
	}

	var items []LabDeviceItem
	var warnings []string
	for _, rec := range records[1:] {
		recItems, recWarn := parseUrineRecord(rec)
		items = append(items, recItems...)
		warnings = append(warnings, recWarn...)
	}
	if len(items) == 0 {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}

	out := LabDeviceFrame{
		SourceType:        model.LabImportSourceTypeArkrayPU4010,
		SourceFingerprint: labDeviceFingerprint(meas),
		MeasuredAt:        measuredAt,
		SpecimenIDRaw:     specimen,
		DeviceHint:        "PU-4010",
		Items:             items,
		Warnings:          warnings,
	}
	markUnknownCodes(&out, knownUrineCodes)
	return out, nil
}

func splitUrineRecords(meas []byte) [][]byte {
	var records [][]byte
	for i := 0; i < len(meas); i++ {
		if meas[i] != urineSTX {
			continue
		}
		stop := indexByteFrom(meas, urineETB, i+1)
		if stop < 0 {
			stop = indexByteFrom(meas, urineETX, i+1)
		}
		if stop < 0 {
			return nil
		}
		records = append(records, meas[i+1:stop])
		i = stop
	}
	return records
}

func parseUrineHeader(header []byte) (*time.Time, string, error) {
	text := string(header)
	if !strings.Contains(text, "VPU-4010") {
		return nil, "", ErrInvalidLabDevicePayload
	}
	date := findTokenWithSep(text, '/')
	clock := ""
	if idx := strings.Index(text, "T"); idx >= 0 && idx+6 <= len(text) {
		clock = text[idx+1 : idx+6]
	}
	if len(date) != 10 || len(clock) != 5 {
		return nil, "", ErrInvalidLabDevicePayload
	}
	measuredAt, err := parseLabDeviceClock("2006/01/02", date, "15:04", clock)
	if err != nil {
		return nil, "", err
	}
	return measuredAt, parseUrineSpecimen(text), nil
}

func findTokenWithSep(text string, sep byte) string {
	for i := 0; i+9 < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			continue
		}
		cand := text[i : i+10]
		if cand[4] == sep && cand[7] == sep {
			return cand
		}
	}
	return ""
}

func parseUrineSpecimen(header string) string {
	search := header
	if t := strings.Index(header, "T"); t >= 0 {
		search = header[t:]
	}
	p := strings.Index(search, "P")
	if p < 0 {
		return ""
	}
	rest := search[p+1:]
	if s := strings.Index(rest, "S"); s >= 0 {
		rest = rest[:s]
	}
	id := strings.TrimSpace(rest)
	if id == "" || id == "-" {
		return ""
	}
	return id
}

func parseUrineRecord(rec []byte) ([]LabDeviceItem, []string) {
	if strings.Contains(string(rec), "T.26") || strings.Contains(string(rec), "COM.") {
		return nil, []string{"trailer:T.26 COM."}
	}
	var items []LabDeviceItem
	for off := 0; off+urineItemLen <= len(rec); off += urineItemLen {
		item, ok := parseUrineItemSlot(rec[off : off+urineItemLen])
		if ok {
			items = append(items, item)
		}
	}
	return items, nil
}

func parseUrineItemSlot(slot []byte) (LabDeviceItem, bool) {
	fields := strings.Fields(string(slot))
	if len(fields) == 0 {
		return LabDeviceItem{}, false
	}
	code := fields[0]
	if code == "T.26" || code == "COM." {
		return LabDeviceItem{}, false
	}
	item := LabDeviceItem{Code: code}
	var values []string
	for _, tok := range fields[1:] {
		if isLabDeviceUnit(tok) {
			item.Unit = tok
			continue
		}
		values = append(values, tok)
	}
	item.ValueRaw = strings.Join(values, " ")
	if item.ValueRaw == "" {
		return LabDeviceItem{}, false
	}
	return item, true
}
