package csvimport

import (
	"strings"
	"testing"
)

func lateReviewFixture(t *testing.T, profile CutoverProfile, mutate func(*fixtureBundle)) (string, ExpectedCutoverSource) {
	t.Helper()
	if profile == CutoverProfileMedicinesCandidate {
		dir, digest := writeMedicinesCandidateFixture(t, mutate)
		return dir, candidateExpectedSource(digest)
	}
	dir, digest := writeCutoverFixture(t, mutate)
	return dir, referenceFixtureSource(digest)
}

func TestPreflightCutoverLateReviewMerchandiseTaxType(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, tc := range []struct {
			name  string
			value string
			want  bool
		}{
			{"included", "included", true},
			{"excluded", "excluded", true},
			{"exempt", "exempt", true},
			{"empty", "", false},
			{"invalid", "sentinel-invalid-tax", false},
			{"wrong-case", "Excluded", false},
		} {
			t.Run(string(profile)+"/"+tc.name, func(t *testing.T) {
				dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, "merchandise_items", "tax_type", tc.value, profile)
				})
				_, err := PreflightCutoverBundle(dir, source)
				if tc.want && err != nil {
					t.Fatalf("tax_type %s rejected under %s: %v", tc.name, profile, err)
				}
				if !tc.want {
					if err == nil || !strings.Contains(err.Error(), "table merchandise_items column tax_type row 2") {
						t.Fatalf("tax_type %s accepted under %s: %v", tc.name, profile, err)
					}
					if len(tc.value) >= 4 && strings.Contains(err.Error(), tc.value) {
						t.Fatalf("diagnostic leaks source value: %v", err)
					}
				}
			})
		}
	}
}

func TestPreflightCutoverLateReviewBillingStatus(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile), func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "billings", "status", "sentinel-invalid-status", profile)
			})
			if _, err := PreflightCutoverBundle(dir, source); err == nil {
				t.Fatalf("invalid billings.status accepted under %s", profile)
			}
		})
	}
}

func TestPreflightCutoverLateReviewRecordAppointmentUnique(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile)+"/duplicate", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "medical_records", "appointment_id", "1000001", profile)
				addFixtureRow(t, f, "medical_records", "1000002", map[string]string{
					"record_no":      "MR-2",
					"appointment_id": "1000001",
				}, profile)
			})
			_, err := PreflightCutoverBundle(dir, source)
			if err == nil || !strings.Contains(err.Error(), "table medical_records column appointment_id row 3") {
				t.Fatalf("duplicate appointment_id accepted under %s: %v", profile, err)
			}
		})
		t.Run(string(profile)+"/alias", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "medical_records", "appointment_id", "1000001", profile)
				addFixtureRow(t, f, "medical_records", "1000002", map[string]string{
					"record_no":      "MR-2",
					"appointment_id": "+01000001",
				}, profile)
			})
			_, err := PreflightCutoverBundle(dir, source)
			if err == nil || !strings.Contains(err.Error(), "table medical_records column appointment_id row 3") {
				t.Fatalf("aliased duplicate appointment_id accepted under %s: %v", profile, err)
			}
		})
		t.Run(string(profile)+"/both-null", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "medical_records", "appointment_id", "", profile)
				addFixtureRow(t, f, "medical_records", "1000002", map[string]string{
					"record_no":      "MR-2",
					"appointment_id": "",
				}, profile)
			})
			if _, err := PreflightCutoverBundle(dir, source); err != nil {
				t.Fatalf("both-empty appointment_id rejected under %s: %v", profile, err)
			}
		})
	}
}

func TestPreflightCutoverLateReviewBillingRecordUnique(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile)+"/duplicate", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				addFixtureRow(t, f, "billings", "1000002", nil, profile)
			})
			_, err := PreflightCutoverBundle(dir, source)
			if err == nil || !strings.Contains(err.Error(), "table billings column medical_record_id row 3") {
				t.Fatalf("duplicate billing medical_record_id accepted under %s: %v", profile, err)
			}
		})
		t.Run(string(profile)+"/alias", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				addFixtureRow(t, f, "billings", "1000002", map[string]string{
					"medical_record_id": "+01000001",
				}, profile)
			})
			_, err := PreflightCutoverBundle(dir, source)
			if err == nil || !strings.Contains(err.Error(), "table billings column medical_record_id row 3") {
				t.Fatalf("aliased duplicate billing medical_record_id accepted under %s: %v", profile, err)
			}
		})
	}
}

func TestPreflightCutoverLateReviewVitalRecordRequired(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile)+"/empty", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "vital_records", "medical_record_id", "", profile)
			})
			_, err := PreflightCutoverBundle(dir, source)
			if err == nil || !strings.Contains(err.Error(), "table vital_records column medical_record_id row 2") {
				t.Fatalf("empty vital medical_record_id accepted under %s: %v", profile, err)
			}
		})
		t.Run(string(profile)+"/baseline", func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {})
			if _, err := PreflightCutoverBundle(dir, source); err != nil {
				t.Fatalf("baseline rejected under %s: %v", profile, err)
			}
		})
	}
}

func TestPreflightCutoverLateReviewWeightBounds(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, tc := range []struct {
			name  string
			value string
			want  bool
		}{
			{"max-exponent", "1e131071", true},
			{"max-exponent-frac", "9.99e131071", true},
			{"min-exponent", "1e-16383", true},
			{"max-fraction", "1." + strings.Repeat("0", 16383), true},
			{"nan", "NaN", true},
			{"pos-inf", "Infinity", true},
			{"over-exponent", "1e131072", false},
			{"huge-exponent", "1e200000", false},
			{"under-exponent", "1e-16384", false},
			{"over-fraction", "1." + strings.Repeat("0", 16384), false},
		} {
			t.Run(string(profile)+"/"+tc.name, func(t *testing.T) {
				dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, "vital_records", "weight", tc.value, profile)
				})
				_, err := PreflightCutoverBundle(dir, source)
				if tc.want && err != nil {
					t.Fatalf("weight %s rejected under %s: %v", tc.name, profile, err)
				}
				if !tc.want {
					if err == nil || !strings.Contains(err.Error(), "table vital_records column weight row 2") {
						t.Fatalf("weight %s accepted under %s: %v", tc.name, profile, err)
					}
				}
			})
		}
	}
}
