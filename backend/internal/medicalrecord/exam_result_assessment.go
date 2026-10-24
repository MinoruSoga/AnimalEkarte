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
	refMin, refMax *model.ExactNumeric,
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
	refMin, refMax *model.ExactNumeric,
) examResultAssessment {
	if !isFiniteOptionalExactNumeric(refMin) || !isFiniteOptionalExactNumeric(refMax) {
		return unassessedExamResult()
	}
	if refMin != nil && refMax != nil {
		if cmp, ok := refMin.Cmp(*refMax); !ok || cmp > 0 {
			return unassessedExamResult()
		}
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(inspectionValue), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return unassessedExamResult()
	}
	// 判定は exact 同士の厳密比較。入力文字列の受理構文は従来どおり ParseFloat とし、
	// 最短10進表現で exact へ橋渡しする。橋渡しできない極端な値は判定不能（unassessed）とする。
	value, err := model.ExactNumericFromFloat64(parsed)
	if err != nil {
		return unassessedExamResult()
	}
	if refMin != nil {
		cmp, ok := value.Cmp(*refMin)
		if !ok {
			return unassessedExamResult()
		}
		if cmp < 0 {
			return abnormalExamResult(model.ExaminationResultStatusLow)
		}
	}
	if refMax != nil {
		cmp, ok := value.Cmp(*refMax)
		if !ok {
			return unassessedExamResult()
		}
		if cmp > 0 {
			return abnormalExamResult(model.ExaminationResultStatusHigh)
		}
	}
	return normalAssessedExamResult()
}

func isFiniteOptionalExactNumeric(value *model.ExactNumeric) bool {
	return value == nil || value.IsFinite()
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
