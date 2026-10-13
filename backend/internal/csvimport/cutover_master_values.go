package csvimport

import (
	"fmt"
	"strconv"
	"strings"
)

// validateCutoverNonNegativeInt enforces the master-price contract: empty
// stays NULL for nullable prices; merchandise unit_price is required. A value
// must be a non-negative int64. Diagnostics name the
// table, column, and row only — never the cell value.
func validateCutoverNonNegativeInt(table, column, value string, csvLine int64) error {
	if value == "" {
		if table == "merchandise_items" && column == "unit_price" {
			return fmt.Errorf("table %s column %s row %d: value is required", table, column, csvLine)
		}
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("table %s column %s row %d: value must be an integer", table, column, csvLine)
	}
	if parsed < 0 {
		return fmt.Errorf("table %s column %s row %d: value must not be negative", table, column, csvLine)
	}
	return nil
}

// cutoverMedicinesDosageForms are the DB dosage_form enum labels exported by
// the producer. Empty stays NULL.
var cutoverMedicinesDosageForms = map[string]struct{}{
	"tablet": {}, "liquid": {}, "injection": {}, "topical": {}, "powder": {},
}

// validateCutoverMedicinesRow enforces the medicines-only lexical checks:
// dosage_form is a DB enum, medicine_unit has no old-database source and must
// stay empty on every row (EMR-342 DESIGN 3.2/3.3), and default_quantity is a
// plain numeric literal exactly representable in numeric(10,2), with zero
// valid and an explicit value required. Diagnostics name table, column, and
// row only.
func validateCutoverMedicinesRow(row []string, indexes map[string]int, csvLine int64) error {
	// EMR-342 DESIGN 3.2 pins is_active=true and is_non_insurance=false.
	// The producer canonical t/f and the semantic true/false lexical forms
	// are accepted; wrong semantics, empty, and non-boolean lexical values
	// fail closed without coercion and without echoing the cell value.
	if value := row[indexes["is_active"]]; value != "t" && value != "true" {
		return fmt.Errorf("table medicines column is_active row %d: is_active must be t or true", csvLine)
	}
	if value := row[indexes["is_non_insurance"]]; value != "f" && value != "false" {
		return fmt.Errorf("table medicines column is_non_insurance row %d: is_non_insurance must be f or false", csvLine)
	}
	if value := row[indexes["dosage_form"]]; value != "" {
		if _, ok := cutoverMedicinesDosageForms[value]; !ok {
			return fmt.Errorf("table medicines column dosage_form row %d: value is not a supported dosage form", csvLine)
		}
	}
	if value := row[indexes["medicine_unit"]]; value != "" {
		return fmt.Errorf("table medicines column medicine_unit row %d: medicine_unit has no old-database source and must stay empty", csvLine)
	}
	if err := validateCutoverMedicineQuantity(row[indexes["default_quantity"]], csvLine); err != nil {
		return err
	}
	return nil
}

// validateCutoverMedicineQuantity requires lossless numeric(10,2) storage.
// Decimal digits are inspected without floating point or rounding. Zero is valid.
func validateCutoverMedicineQuantity(value string, csvLine int64) error {
	if value == "" {
		return fmt.Errorf("table medicines column default_quantity row %d: value is required", csvLine)
	}
	if !cutoverNumericLiteral.MatchString(value) || strings.HasPrefix(value, "-") {
		return fmt.Errorf("table medicines column default_quantity row %d: value must be a non-negative decimal quantity", csvLine)
	}
	unsigned := strings.TrimPrefix(value, "+")
	integer, fraction, _ := strings.Cut(unsigned, ".")
	if len(strings.TrimLeft(integer, "0")) > 8 || len(strings.TrimRight(fraction, "0")) > 2 {
		return fmt.Errorf("table medicines column default_quantity row %d: value must fit numeric(10,2) exactly without rounding", csvLine)
	}
	return nil
}

// validateCutoverMedicinesName rejects blank and whitespace-only source
// names at preflight. The producer classifies such rows as needs_review and
// removes them, so they can never be imported. Names are not normalized:
// an accepted name is imported exactly as written. The error names the
// column and CSV row only, never the rejected cell value.
func validateCutoverMedicinesName(value string, csvLine int64) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("table medicines column name row %d: name is required and must not be blank", csvLine)
	}
	return nil
}

// validateCutoverEstimateMedicineRewrite enforces the candidate22 medicine
// rewrite joint invariant (EMR-342 DESIGN 3.4): a populated medicine_id
// means the estimate row was rewritten from procedures to medicines, so the
// category must be medicine and consultation_id, procedure_id, and
// merchandise_item_id must all be empty. A medicine category also requires
// medicine_id. Non-medicine rows and the legacy21 profile keep their
// existing behavior; legacy rejects a populated medicine_id through the
// unavailable-parent reference check instead.
func validateCutoverEstimateMedicineRewrite(profile CutoverProfile, row []string, indexes map[string]int, csvLine int64) error {
	if profile != CutoverProfileMedicinesCandidate {
		return nil
	}
	medicineID := row[indexes["medicine_id"]]
	category := row[indexes["category"]]
	if medicineID != "" {
		if category != "medicine" {
			return fmt.Errorf("table estimate_items column category row %d: populated medicine_id requires category medicine", csvLine)
		}
		for _, column := range []string{"consultation_id", "procedure_id", "merchandise_item_id"} {
			if row[indexes[column]] != "" {
				return fmt.Errorf("table estimate_items column %s row %d: medicine rewrite requires an empty %s", column, csvLine, column)
			}
		}
	}
	if category == "medicine" && medicineID == "" {
		return fmt.Errorf("table estimate_items column medicine_id row %d: category medicine requires medicine_id", csvLine)
	}
	return nil
}
