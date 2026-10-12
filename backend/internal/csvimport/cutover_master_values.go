package csvimport

import (
	"fmt"
	"strconv"
)

// validateCutoverNonNegativeInt enforces the master-price contract: empty
// stays NULL and the cell must be a non-negative int64. Diagnostics name the
// table, column, and row only — never the cell value.
func validateCutoverNonNegativeInt(table, column, value string, csvLine int64) error {
	if value == "" {
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
// dosage_form is the DB enum and default_quantity is a plain numeric literal
// (empty stays NULL). Diagnostics name table, column, and row only.
func validateCutoverMedicinesRow(row []string, indexes map[string]int, csvLine int64) error {
	if value := row[indexes["dosage_form"]]; value != "" {
		if _, ok := cutoverMedicinesDosageForms[value]; !ok {
			return fmt.Errorf("table medicines column dosage_form row %d: value is not a supported dosage form", csvLine)
		}
	}
	if value := row[indexes["default_quantity"]]; value != "" && !cutoverNumericLiteral.MatchString(value) {
		return fmt.Errorf("table medicines column default_quantity row %d: value must be a numeric quantity", csvLine)
	}
	return nil
}
