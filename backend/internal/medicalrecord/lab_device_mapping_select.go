package medicalrecord

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/model"
)

// errMultipleLabResults means the payload held more than one distinct result.
var errMultipleLabResults = errors.New("multiple_results")

const (
	labMappingDeviceUnitMax = 64
	labMappingChartUnitMax  = 32
	labMappingCodeMax       = 64
)

// selectOneLabResult keeps a single measurement frame.
// Identical fingerprints collapse. Distinct frames are an error.
// Empty fingerprints and empty item lists are skipped.
func selectOneLabResult(payload []byte, hint string) (LabDeviceFrame, error) {
	frames, err := DecodeLabDeviceFrames(payload, hint)
	if err != nil {
		return LabDeviceFrame{}, err
	}
	chosen := make([]LabDeviceFrame, 0, 1)
	seen := make(map[string]struct{}, len(frames))
	for _, frame := range frames {
		if frame.SourceFingerprint == "" || len(frame.Items) == 0 {
			continue
		}
		if _, ok := seen[frame.SourceFingerprint]; ok {
			continue
		}
		seen[frame.SourceFingerprint] = struct{}{}
		frame.Species = normalizeLabDeviceSpecies(frame.Species)
		chosen = append(chosen, frame)
	}
	if len(chosen) == 0 {
		return LabDeviceFrame{}, ErrInvalidLabDevicePayload
	}
	if len(chosen) > 1 {
		return LabDeviceFrame{}, errMultipleLabResults
	}
	return chosen[0], nil
}

// normalizeLabDeviceSpecies returns canine, feline, or empty.
// Free text and pet names become empty. dog/cat are accepted only as a whole token.
func normalizeLabDeviceSpecies(species string) string {
	switch strings.ToLower(strings.TrimSpace(species)) {
	case "canine", "dog", "犬":
		return "canine"
	case "feline", "cat", "猫":
		return "feline"
	default:
		return ""
	}
}

// labDeviceSpeciesFromText reads a species token from one result frame.
// canine/feline/犬/猫 must be a whole word. A character inside a longer word is ignored,
// so another species' confirmed map is not selected by a partial match.
// Bare dog/cat are not species.
func labDeviceSpeciesFromText(text string) string {
	canineAt, felineAt := -1, -1
	for _, token := range labSpeciesTokens(text) {
		switch strings.ToLower(token.text) {
		case "canine", "犬":
			if canineAt < 0 {
				canineAt = token.at
			}
		case "feline", "猫":
			if felineAt < 0 {
				felineAt = token.at
			}
		}
	}
	switch {
	case canineAt >= 0 && (felineAt < 0 || canineAt <= felineAt):
		return "canine"
	case felineAt >= 0:
		return "feline"
	default:
		return ""
	}
}

type labSpeciesToken struct {
	text string
	at   int
}

func labSpeciesTokens(text string) []labSpeciesToken {
	tokens := make([]labSpeciesToken, 0, 4)
	start := -1
	for i, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			tokens = append(tokens, labSpeciesToken{text: text[start:i], at: start})
			start = -1
		}
	}
	if start >= 0 {
		tokens = append(tokens, labSpeciesToken{text: text[start:], at: start})
	}
	return tokens
}

// isLabMappingComment reports items that are not analyte codes.
func isLabMappingComment(item LabDeviceItem) bool {
	if item.Flag == "comment" || strings.TrimSpace(item.Code) == "" {
		return true
	}
	if strings.ContainsAny(item.Code, " \t") || utf8.RuneCountInString(item.Code) > labMappingCodeMax {
		return true
	}
	return false
}

// unitsMismatch reports a display-only difference. Empty units are not a mismatch.
// Values are never converted.
func unitsMismatch(deviceUnit, chartUnit string) bool {
	device := foldLabUnit(deviceUnit)
	chart := foldLabUnit(chartUnit)
	if device == "" || chart == "" {
		return false
	}
	return device != chart
}

func foldLabUnit(unit string) string {
	folded := strings.ToLower(strings.TrimSpace(unit))
	folded = strings.ReplaceAll(folded, " ", "")
	folded = strings.ReplaceAll(folded, "µ", "u")
	folded = strings.ReplaceAll(folded, "μ", "u")
	return folded
}

func clipRunes(value string, maxLen int) string {
	if utf8.RuneCountInString(value) <= maxLen {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxLen])
}

func mappingSelectionCategory(err error) string {
	if errors.Is(err, errMultipleLabResults) {
		return "multiple_results"
	}
	return "invalid_payload"
}

func mappingSourceOrEmpty(frame LabDeviceFrame) string {
	source := string(frame.SourceType)
	if isLabDeviceSourceType(source) {
		return source
	}
	return ""
}

func draftItemFromFrame(clinicID uint64, sort int, item LabDeviceItem, source string) model.LabDeviceMappingDraftItem {
	return model.LabDeviceMappingDraftItem{
		ClinicID:       clinicID,
		DeviceItemCode: clipRunes(strings.TrimSpace(item.Code), labMappingCodeMax),
		ValueRaw:       item.ValueRaw,
		DeviceUnit:     clipRunes(strings.TrimSpace(item.Unit), labMappingDeviceUnitMax),
		ItemSource:     source,
		SortOrder:      sort,
	}
}
