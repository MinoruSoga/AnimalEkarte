package csvimport

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreflightCutoverDDLBooleanColumns(t *testing.T) {
	columns := []struct {
		table    string
		column   string
		nullable bool
	}{
		{"staffs", "is_active", false},
		{"staffs", "reservation_visible", false},
		{"procedures", "is_active", false},
		{"procedures", "is_surgery", false},
		{"merchandise_items", "is_active", false},
		{"owners", "is_dangerous", false},
		{"owners", "dm_preference", true},
		{"billing_items", "is_insurance_applicable", true},
		{"estimate_items", "is_insurance_applicable", true},
		{"vaccines", "is_active", false},
	}
	values := []struct {
		name  string
		value string
		valid bool
	}{
		{"true-short", "t", true},
		{"false-short", "f", true},
		{"true-case", "TRUE", true},
		{"false-case", "False", true},
		{"true-number", "1", true},
		{"false-number", "0", true},
		{"true-prefix", "y", true},
		{"false-prefix", "n", true},
		{"true-on", "on", true},
		{"false-prefix-off", "of", true},
		{"true-space", " true ", true},
		{"false-space", " off ", true},
		{"empty", "", false},
		{"invalid-sentinel", "sentinel-invalid-bool", false},
		{"invalid-number", "2", false},
		{"ambiguous-prefix", "o", false},
		{"invalid-suffix", "truefalse", false},
		{"whitespace-not-null", " \t ", false},
	}
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, column := range columns {
			for _, tc := range values {
				t.Run(string(profile)+"/"+column.table+"/"+column.column+"/"+tc.name, func(t *testing.T) {
					dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
						setFixtureCell(t, f, column.table, column.column, tc.value, profile)
					})
					_, err := PreflightCutoverBundle(dir, source)
					want := tc.valid || (tc.value == "" && column.nullable)
					if want {
						if err != nil {
							t.Fatalf("valid %s.%s/%s rejected: %v", column.table, column.column, tc.name, err)
						}
						return
					}
					coordinate := fmt.Sprintf("table %s column %s row 2:", column.table, column.column)
					if err == nil || !strings.Contains(err.Error(), coordinate) || !strings.Contains(err.Error(), "value must be a boolean literal") {
						t.Fatalf("invalid %s.%s/%s did not reach intended guard: %v", column.table, column.column, tc.name, err)
					}
					if tc.value == "sentinel-invalid-bool" && strings.Contains(err.Error(), tc.value) {
						t.Fatalf("diagnostic echoes source value: %v", err)
					}
				})
			}
		}
	}
}
