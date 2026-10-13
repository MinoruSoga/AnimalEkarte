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

// cutoverMedicineUnits are the DB medicine_unit enum labels exported by the
// producer. Empty stays NULL.
var cutoverMedicineUnits = map[string]struct{}{
	"per_tablet": {}, "per_ml": {}, "per_dose": {}, "per_gram": {},
}

// validateCutoverMedicinesRow enforces the medicines-only lexical checks:
// dosage_form and medicine_unit are DB enums and default_quantity is a plain
// numeric literal exactly representable in numeric(10,2), with zero valid and
// empty=NULL. Diagnostics name table, column, and row only.
func validateCutoverMedicinesRow(row []string, indexes map[string]int, csvLine int64) error {
	if value := row[indexes["dosage_form"]]; value != "" {
		if _, ok := cutoverMedicinesDosageForms[value]; !ok {
			return fmt.Errorf("table medicines column dosage_form row %d: value is not a supported dosage form", csvLine)
		}
	}
	if value := row[indexes["medicine_unit"]]; value != "" {
		if _, ok := cutoverMedicineUnits[value]; !ok {
			return fmt.Errorf("table medicines column medicine_unit row %d: value is not a supported medicine unit", csvLine)
		}
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
		return nil
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
