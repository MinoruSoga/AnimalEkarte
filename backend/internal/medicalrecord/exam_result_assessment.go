package medicalrecord

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	qualitativeNegative       = "(-)"
	qualitativePlusMinus      = "(±)"
	qualitativePositive       = "(+)"
	qualitativeDoublePositive = "(++)"
	qualitativeTriplePositive = "(+++)"
)

const (
	qualitativeNegativeIndex = iota
	qualitativePlusMinusIndex
	qualitativePositiveIndex
	qualitativeDoublePositiveIndex
	qualitativeTriplePositiveIndex
)

type examResultAssessment struct {
	status     model.ExaminationResultStatus
	isAbnormal bool
	isAssessed bool
}

func assessExamResult(
	inspectionValue string,
	refMin, refMax *float64,
	qualitativeMin, qualitativeMax *string,
) examResultAssessment {
	hasNumericBounds := refMin != nil || refMax != nil
	hasQualitativeBounds := qualitativeMin != nil || qualitativeMax != nil
	if hasNumericBounds && hasQualitativeBounds {
		return unassessedExamResult()
	}
	if hasNumericBounds {
		return assessNumericExamResult(inspectionValue, refMin, refMax)
	}
	if hasQualitativeBounds {
		return assessQualitativeExamResult(inspectionValue, qualitativeMin, qualitativeMax)
	}
	return unassessedExamResult()
}

func assessNumericExamResult(
	inspectionValue string,
	refMin, refMax *float64,
) examResultAssessment {
	if !isFiniteOptionalFloat64(refMin) || !isFiniteOptionalFloat64(refMax) {
		return unassessedExamResult()
	}
	if refMin != nil && refMax != nil && *refMin > *refMax {
		return unassessedExamResult()
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(inspectionValue), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return unassessedExamResult()
	}
	if refMin != nil && value < *refMin {
		return abnormalExamResult(model.ExaminationResultStatusLow)
	}
	if refMax != nil && value > *refMax {
		return abnormalExamResult(model.ExaminationResultStatusHigh)
	}
	return normalAssessedExamResult()
}

func isFiniteOptionalFloat64(value *float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0))
}

func assessQualitativeExamResult(
	inspectionValue string,
	qualitativeMin, qualitativeMax *string,
) examResultAssessment {
	valueIndex, ok := qualitativeValueIndex(inspectionValue)
	if !ok {
		return unassessedExamResult()
	}

	var minIndex, maxIndex int
	if qualitativeMin != nil {
		minIndex, ok = qualitativeValueIndex(*qualitativeMin)
		if !ok {
			return unassessedExamResult()
		}
	}
	if qualitativeMax != nil {
		maxIndex, ok = qualitativeValueIndex(*qualitativeMax)
		if !ok {
			return unassessedExamResult()
		}
	}
	if qualitativeMin != nil && qualitativeMax != nil && minIndex > maxIndex {
		return unassessedExamResult()
	}
	if qualitativeMin != nil && valueIndex < minIndex {
		return abnormalExamResult(model.ExaminationResultStatusLow)
	}
	if qualitativeMax != nil && valueIndex > maxIndex {
		return abnormalExamResult(model.ExaminationResultStatusHigh)
	}
	return normalAssessedExamResult()
}

func qualitativeValueIndex(value string) (int, bool) {
	switch normalizeQualitativeValue(value) {
	case qualitativeNegative:
		return qualitativeNegativeIndex, true
	case qualitativePlusMinus:
		return qualitativePlusMinusIndex, true
	case qualitativePositive:
		return qualitativePositiveIndex, true
	case qualitativeDoublePositive:
		return qualitativeDoublePositiveIndex, true
	case qualitativeTriplePositive:
		return qualitativeTriplePositiveIndex, true
	default:
		return 0, false
	}
}

func normalizeQualitativeValue(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsSpace(r):
			return -1
		case r == '（':
			return '('
		case r == '）':
			return ')'
		default:
			return r
		}
	}, value)
}

func unassessedExamResult() examResultAssessment {
	return examResultAssessment{status: model.ExaminationResultStatusNormal}
}

func normalAssessedExamResult() examResultAssessment {
	return examResultAssessment{
		status:     model.ExaminationResultStatusNormal,
		isAssessed: true,
	}
}

func abnormalExamResult(status model.ExaminationResultStatus) examResultAssessment {
	return examResultAssessment{
		status:     status,
		isAbnormal: true,
		isAssessed: true,
	}
}
