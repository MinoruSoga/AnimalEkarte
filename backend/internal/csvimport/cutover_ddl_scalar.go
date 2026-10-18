package csvimport

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	cutoverPGNumericMaxIntegerDigits    = 131072
	cutoverPGNumericMaxFractionalDigits = 16383
)

type cutoverDDLScalarKind uint8

const (
	cutoverDDLInt32 cutoverDDLScalarKind = iota
	cutoverDDLInt64
	cutoverDDLNumeric
)

type cutoverDDLScalarRule struct {
	column    string
	kind      cutoverDDLScalarKind
	nullable  bool
	precision int
	scale     int
}

var cutoverDDLScalarColumns = map[string][]cutoverDDLScalarRule{
	"staffs": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
	},
	"procedures": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"price", cutoverDDLInt64, true, 0, 0},
		{"duration", cutoverDDLInt32, true, 0, 0},
		{"parent_id", cutoverDDLInt64, true, 0, 0},
		{"tax_rate", cutoverDDLNumeric, false, 0, 0},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"medicines": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"price", cutoverDDLInt64, true, 0, 0},
		{"parent_id", cutoverDDLInt64, true, 0, 0},
		{"inventory_id", cutoverDDLInt64, true, 0, 0},
		{"default_quantity", cutoverDDLNumeric, true, 10, 2},
		{"tax_rate", cutoverDDLNumeric, false, 0, 0},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"merchandise_items": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"unit_price", cutoverDDLInt64, false, 0, 0},
		{"tax_rate", cutoverDDLNumeric, false, 0, 0},
		{"sort_order", cutoverDDLInt32, false, 0, 0},
	},
	"owners": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"discount_rate", cutoverDDLNumeric, false, 5, 2},
	},
	"pets": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"owner_id", cutoverDDLInt64, false, 0, 0},
		{"animal_species_id", cutoverDDLInt64, false, 0, 0},
		{"weight", cutoverDDLNumeric, true, 6, 2},
	},
	"appointments": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"owner_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"reservation_type_id", cutoverDDLInt64, false, 0, 0},
		{"doctor_id", cutoverDDLInt64, true, 0, 0},
	},
	"appointment_trimming_details": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"appointment_id", cutoverDDLInt64, false, 0, 0},
	},
	"medical_records": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"owner_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"appointment_id", cutoverDDLInt64, true, 0, 0},
		{"doctor_id", cutoverDDLInt64, true, 0, 0},
		{"entered_by", cutoverDDLInt64, true, 0, 0},
	},
	"inquiries": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, false, 0, 0},
		{"staff_id", cutoverDDLInt64, true, 0, 0},
	},
	"clinical_plans": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, false, 0, 0},
	},
	"vital_records": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, false, 0, 0},
		{"temperature", cutoverDDLNumeric, true, 0, 0},
		{"weight", cutoverDDLNumeric, true, 0, 0},
		{"heart_rate", cutoverDDLInt32, true, 0, 0},
		{"respiration_rate", cutoverDDLInt32, true, 0, 0},
		{"staff_id", cutoverDDLInt64, true, 0, 0},
	},
	"billings": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, true, 0, 0},
		{"owner_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"total_amount", cutoverDDLInt64, false, 0, 0},
	},
	"billing_items": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"billing_id", cutoverDDLInt64, false, 0, 0},
		{"unit_price", cutoverDDLInt64, false, 0, 0},
		{"quantity", cutoverDDLNumeric, false, 10, 1},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"payments": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"billing_id", cutoverDDLInt64, false, 0, 0},
		{"subtotal", cutoverDDLInt64, false, 0, 0},
		{"tax_total", cutoverDDLInt64, false, 0, 0},
		{"total_amount", cutoverDDLInt64, false, 0, 0},
		{"insurance_ratio", cutoverDDLNumeric, true, 3, 2},
		{"insurance_amount", cutoverDDLInt64, true, 0, 0},
		{"discount_amount", cutoverDDLInt64, true, 0, 0},
		{"billing_amount", cutoverDDLInt64, false, 0, 0},
		{"received_amount", cutoverDDLInt64, true, 0, 0},
		{"change_amount", cutoverDDLInt64, true, 0, 0},
		{"payment_method_id", cutoverDDLInt64, true, 0, 0},
		{"paid_by", cutoverDDLInt64, true, 0, 0},
	},
	"payment_splits": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"billing_id", cutoverDDLInt64, false, 0, 0},
		{"payment_method_id", cutoverDDLInt64, true, 0, 0},
		{"amount", cutoverDDLInt64, false, 0, 0},
		{"received_amount", cutoverDDLInt64, false, 0, 0},
		{"change_amount", cutoverDDLInt64, false, 0, 0},
		{"paid_by", cutoverDDLInt64, true, 0, 0},
	},
	"estimates": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, true, 0, 0},
		{"owner_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"subtotal", cutoverDDLInt64, false, 0, 0},
		{"tax_total", cutoverDDLInt64, false, 0, 0},
		{"total_amount", cutoverDDLInt64, false, 0, 0},
		{"insurance_amount", cutoverDDLInt64, true, 0, 0},
		{"discount_amount", cutoverDDLInt64, true, 0, 0},
		{"created_by", cutoverDDLInt64, true, 0, 0},
	},
	"estimate_items": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"estimate_id", cutoverDDLInt64, false, 0, 0},
		{"unit_price", cutoverDDLInt64, false, 0, 0},
		{"quantity", cutoverDDLNumeric, false, 10, 1},
		{"tax_rate", cutoverDDLNumeric, true, 3, 2},
		{"discount_rate", cutoverDDLNumeric, true, 5, 2},
		{"discount_amount", cutoverDDLInt64, true, 0, 0},
		{"consultation_id", cutoverDDLInt64, true, 0, 0},
		{"procedure_id", cutoverDDLInt64, true, 0, 0},
		{"medicine_id", cutoverDDLInt64, true, 0, 0},
		{"merchandise_item_id", cutoverDDLInt64, true, 0, 0},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"exams": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"exam_type_id", cutoverDDLInt64, false, 0, 0},
	},
	"exam_results": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"exam_id", cutoverDDLInt64, false, 0, 0},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"vaccines": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"price", cutoverDDLInt64, true, 0, 0},
		{"inventory_id", cutoverDDLInt64, true, 0, 0},
		{"parent_id", cutoverDDLInt64, true, 0, 0},
		{"sort_order", cutoverDDLInt32, true, 0, 0},
	},
	"vaccinations": {
		{"id", cutoverDDLInt64, false, 0, 0},
		{"clinic_id", cutoverDDLInt64, false, 0, 0},
		{"medical_record_id", cutoverDDLInt64, true, 0, 0},
		{"pet_id", cutoverDDLInt64, true, 0, 0},
		{"vaccine_id", cutoverDDLInt64, false, 0, 0},
		{"doctor_id", cutoverDDLInt64, true, 0, 0},
	},
}

func cutoverASCIITrimSpace(text string) string {
	return strings.Trim(text, " \t\n\r\f\v")
}

func validateCutoverDDLScalarValue(table string, rule cutoverDDLScalarRule, text string, csvLine int64) error {
	if text == "" {
		if rule.nullable {
			return nil
		}
		return fmt.Errorf("table %s column %s row %d: value must not be empty", table, rule.column, csvLine)
	}
	if strings.HasPrefix(text, "{{") && placeholderAllowed(table, rule.column, text, text) {
		return nil
	}
	trimmed := cutoverASCIITrimSpace(text)
	if strings.TrimSpace(trimmed) != trimmed {
		return fmt.Errorf("table %s column %s row %d: value must be a numeric literal", table, rule.column, csvLine)
	}
	switch rule.kind {
	case cutoverDDLInt32, cutoverDDLInt64:
		if !cutoverIntegerLiteral.MatchString(trimmed) {
			return fmt.Errorf("table %s column %s row %d: value must be an integer", table, rule.column, csvLine)
		}
		bits := 32
		if rule.kind == cutoverDDLInt64 {
			bits = 64
		}
		if _, err := strconv.ParseInt(trimmed, 10, bits); err != nil {
			return fmt.Errorf("table %s column %s row %d: value is outside the integer%d range", table, rule.column, csvLine, bits)
		}
		return nil
	case cutoverDDLNumeric:
		value, class := parseCutoverPGNumeric(text)
		switch class {
		case cutoverNumericInvalid:
			return fmt.Errorf("table %s column %s row %d: value must be a numeric literal", table, rule.column, csvLine)
		case cutoverNumericNaN:
			return nil
		case cutoverNumericPosInf, cutoverNumericNegInf:
			if rule.precision == 0 {
				return nil
			}
			return fmt.Errorf("table %s column %s row %d: value exceeds PostgreSQL numeric bounds", table, rule.column, csvLine)
		}
		if rule.precision == 0 {
			if value.decPos > cutoverPGNumericMaxIntegerDigits || value.dscale > cutoverPGNumericMaxFractionalDigits {
				return fmt.Errorf("table %s column %s row %d: value exceeds PostgreSQL numeric bounds", table, rule.column, csvLine)
			}
			return nil
		}
		if value.sign == 0 {
			return nil
		}
		bound := value
		bound.sign = 1
		if cutoverDecimalCompare(bound, strings.Repeat("9", rule.precision)+"5", int64(rule.precision-rule.scale)) >= 0 {
			return fmt.Errorf("table %s column %s row %d: value exceeds the numeric(%d,%d) precision bound", table, rule.column, csvLine, rule.precision, rule.scale)
		}
		return nil
	}
	return fmt.Errorf("table %s column %s row %d: unsupported scalar kind", table, rule.column, csvLine)
}

func validateCutoverDDLScalarRow(spec CutoverTableSpec, row []string, indexes map[string]int, csvLine int64) error {
	for _, rule := range cutoverDDLScalarColumns[spec.Name] {
		if err := validateCutoverDDLScalarValue(spec.Name, rule, row[indexes[rule.column]], csvLine); err != nil {
			return err
		}
	}
	return nil
}
