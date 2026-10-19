package csvimport

import (
	"strings"
	"testing"
)

// ddlSpecByName looks up a copied table in the selected profile's specs and fails the test if it is absent.
func ddlSpecByName(t *testing.T, table string, profiles ...CutoverProfile) CutoverTableSpec {
	t.Helper()
	for _, spec := range CutoverTableSpecs(profiles...) {
		if spec.Name == table {
			return spec
		}
	}
	t.Fatalf("spec is missing table %s", table)
	return CutoverTableSpec{}
}

// setFixtureCell replaces one cell in a fixture row for the given profile's bundle.
func setFixtureCell(t *testing.T, f *fixtureBundle, table, column, value string, profiles ...CutoverProfile) {
	t.Helper()
	spec := ddlSpecByName(t, table, profiles...)
	f.rows[table][0][columnIndex(spec.Columns, column)] = value
}

// addFixtureRow appends a fixture row with a new id and explicit cell overrides for the given profile.
func addFixtureRow(t *testing.T, f *fixtureBundle, table, newID string, edits map[string]string, profiles ...CutoverProfile) {
	t.Helper()
	spec := ddlSpecByName(t, table, profiles...)
	row := append([]string(nil), f.rows[table][0]...)
	row[columnIndex(spec.Columns, "id")] = newID
	for column, value := range edits {
		row[columnIndex(spec.Columns, column)] = value
	}
	f.rows[table] = append(f.rows[table], row)
	f.manifest.Tables[manifestTableIndex(t, f, table)].RowCount++
}

// TestPreflightCutoverQuantityNumericTypmod exercises line-item quantity acceptance and rejection at numeric(10,1) rounding boundaries, including NaN and malformed literals.
func TestPreflightCutoverQuantityNumericTypmod(t *testing.T) {
	for _, table := range []string{"billing_items", "estimate_items"} {
		for _, tc := range []struct {
			value string
			want  bool
		}{
			{"", false}, {"0", false}, {"-1", false}, {"-0.04", false}, {"0.04", false},
			{"999999999.95", false}, {"9999999999.4", false}, {"abc", false}, {"1.5.2", false},
			{"Infinity", false}, {"+Infinity", false}, {"-Infinity", false}, {"inf", false},
			{"1e99999999", false}, {"1e1073741824", false}, {"1e-1073741824", false},
			{"1e", false}, {"1E", false}, {"1.e", false}, {"1e+", false}, {"1e-", false},
			{"+NaN", false}, {"-NaN", false}, {"+nan", false},
			{"0.05", true}, {"1", true}, {"1.04", true}, {"1.05", true}, {"1e-1", true},
			{"999999999.94", true}, {"NaN", true}, {"nAn", true}, {"0.00000000000000000001e20", true},
			{"1e00000001", true},
			{strings.Repeat("0", 140001) + "1", true},
			{"1." + strings.Repeat("0", 140001) + "1", true},
		} {
			t.Run(table+"/"+tc.value, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setFixtureCell(t, f, table, "quantity", tc.value)
				})
				_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
				if tc.want && err != nil {
					t.Fatalf("quantity %q rejected: %v", tc.value, err)
				}
				if !tc.want && err == nil {
					t.Fatalf("quantity %q accepted", tc.value)
				}
				if err != nil && len(tc.value) >= 2 && strings.Contains(err.Error(), tc.value) {
					t.Fatalf("diagnostic leaks source value: %v", err)
				}
			})
		}
	}
}

// TestPreflightCutoverVitalBounds pins the vital measurement floors, ranges, and numeric bounds for each vitals column.
func TestPreflightCutoverVitalBounds(t *testing.T) {
	for _, tc := range []struct {
		column string
		values map[string]bool
	}{
		{"temperature", map[string]bool{
			"": true, "30": true, "50": true, "30.0": true, "50.0": true, "40.5": true,
			"29.999999999999999999": false, "50.1": false, "29.9": false, "0": false,
			"50." + strings.Repeat("0", 140001) + "1": false,
			"NaN": false, "+NaN": false, "-NaN": false, "Infinity": false, "-Infinity": false, "abc": false, "30e": false,
		}},
		{"heart_rate", map[string]bool{
			"": true, "1": true, "499": true, "250": true,
			"0": false, "500": false, "-1": false, "1.5": false, "abc": false, "99999999999999999999": false,
		}},
		{"respiration_rate", map[string]bool{
			"": true, "1": true, "199": true,
			"0": false, "200": false, "2.5": false,
		}},
		{"weight", map[string]bool{
			"": true, "0.5": true, "1e-1000": true, "0.00001": true, "NaN": true, "nAN": true, "Infinity": true, "1e100000": true,
			strings.Repeat("0", 140001) + "1": true, "1e00000001": true,
			"0": false, "-1": false, "+NaN": false, "-NaN": false, "-Infinity": false, "abc": false, "1e": false,
		}},
		{"weight_unit", map[string]bool{
			"": true, "Kg": true, "g": true,
			"kg": false, "lb": false, "KG": false,
		}},
	} {
		for value, want := range tc.values {
			t.Run(tc.column+"/"+value, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setFixtureCell(t, f, "vital_records", tc.column, value)
				})
				_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
				if want && err != nil {
					t.Fatalf("%s=%q rejected: %v", tc.column, value, err)
				}
				if !want && err == nil {
					t.Fatalf("%s=%q accepted", tc.column, value)
				}
			})
		}
	}
}

// TestPreflightCutoverKanaCheck rejects its katakana cases and accepts the hiragana and empty name_kana controls.
func TestPreflightCutoverKanaCheck(t *testing.T) {
	for _, tc := range []struct {
		table string
		value string
		want  bool
	}{
		{"owners", "ア", false}, {"owners", "ヶ", false}, {"owners", "タロウ", false},
		{"owners", "あいう", true}, {"owners", "", true}, {"owners", "かな", true},
		{"pets", "ポチ", false}, {"pets", "", true}, {"pets", "ぽち", true},
	} {
		t.Run(tc.table+"/"+tc.value, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, tc.table, "name_kana", tc.value)
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if tc.want && err != nil {
				t.Fatalf("name_kana %q rejected: %v", tc.value, err)
			}
			if !tc.want && err == nil {
				t.Fatalf("name_kana %q accepted", tc.value)
			}
			if err != nil && strings.Contains(err.Error(), tc.value) && tc.value != "" {
				t.Fatalf("diagnostic leaks source value: %v", err)
			}
		})
	}
}

// TestPreflightCutoverEnumDomains exercises accepted and rejected enum spellings for the copied source columns.
func TestPreflightCutoverEnumDomains(t *testing.T) {
	for _, tc := range []struct {
		table, column string
		values        map[string]bool
	}{
		{"owners", "membership_type", map[string]bool{
			"non_member": true, "member": true, "deceased": true, "transferred": true,
			"": false, "vip": false, "Member": false}},
		{"pets", "gender", map[string]bool{
			"male": true, "female": true, "unknown": true, "": false, "other": false}},
		{"pets", "status", map[string]bool{
			"alive": true, "deceased": true, "": false, "missing": false}},
		{"procedures", "anesthesia", map[string]bool{
			"": true, "none": true, "local": true, "sedation": true, "general": true, "epidural": false}},
		{"procedures", "tax_type", map[string]bool{
			"included": true, "excluded": true, "exempt": true, "": false, "included20": false}},
		{"appointments", "visit_type", map[string]bool{
			"first": true, "revisit": true, "": false, "third": false}},
		{"appointments", "status", map[string]bool{
			"": true, "confirmed": true, "tentative": true, "no_show": true, "bogus": false}},
		{"appointments", "source", map[string]bool{
			"manual": true, "line": true, "": false, "phone": false}},
		{"medical_records", "status", map[string]bool{
			"": true, "draft": true, "finalized": true, "open": false}},
		{"medical_records", "visit_type", map[string]bool{
			"": true, "first": true, "revisit": true, "followup": false}},
		{"billing_items", "category", map[string]bool{
			"examination": true, "house_call": true, "trimming": true, "": false, "apparel": false}},
		{"billing_items", "tax_type", map[string]bool{"included": true, "exempt": true, "": false}},
		{"estimates", "status", map[string]bool{
			"": true, "draft": true, "sent": true, "approved": true, "rejected": true, "expired": false}},
		{"estimate_items", "category", map[string]bool{
			"procedure": true, "house_call": true, "": false, "service": false}},
		{"estimate_items", "tax_type", map[string]bool{"excluded": true, "": false}},
		{"merchandise_items", "category", map[string]bool{
			"food": true, "goods": true, "other": true, "house_call": true, "": false, "toy": false}},
		{"vaccinations", "next_schedule_type", map[string]bool{
			"": true, "3weeks": true, "4weeks": true, "1year": true, "other": true, "2weeks": false}},
		{"vaccines", "species", map[string]bool{
			"": true, "dog": true, "cat": true, "both": true, "dogs": false}},
	} {
		for value, want := range tc.values {
			t.Run(tc.table+"."+tc.column+"/"+value, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setFixtureCell(t, f, tc.table, tc.column, value)
				})
				_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
				if want && err != nil {
					t.Fatalf("%s.%s=%q rejected: %v", tc.table, tc.column, value, err)
				}
				if !want && err == nil {
					t.Fatalf("%s.%s=%q accepted", tc.table, tc.column, value)
				}
				if err != nil && !strings.Contains(err.Error(), tc.column) {
					t.Fatalf("diagnostic does not name column: %v", err)
				}
			})
		}
	}
}

// TestPreflightCutoverRequiredTimestamps proves required timestamp columns reject blank values only; no date grammar or ordering is asserted.
func TestPreflightCutoverRequiredTimestamps(t *testing.T) {
	for _, column := range []string{"start_time", "end_time"} {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			setFixtureCell(t, f, "appointments", column, "")
		})
		if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
			t.Fatalf("empty %s accepted", column)
		}
	}
}

// TestPreflightCutoverMerchandiseActiveBoolean pins is_active boolean literal acceptance and rejection for merchandise_items.
func TestPreflightCutoverMerchandiseActiveBoolean(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"true", true}, {"t", true}, {"TRUE", true}, {"yes", true}, {"On", true}, {"1", true}, {" of ", true},
		{"false", true}, {"0", true}, {"off", true},
		{"", false}, {"maybe", false}, {"o", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, "merchandise_items", "is_active", tc.value)
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if tc.want && err != nil {
				t.Fatalf("is_active %q rejected: %v", tc.value, err)
			}
			if !tc.want && err == nil {
				t.Fatalf("is_active %q accepted", tc.value)
			}
		})
	}
}

// TestPreflightCutoverMasterNameUnique rejects duplicate procedure and vaccine names inside the source bundle.
func TestPreflightCutoverMasterNameUnique(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "name", "shared-name")
				addFixtureRow(t, f, table, "1000002", map[string]string{"name": "shared-name"})
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), "duplicate") {
				t.Fatalf("duplicate %s name accepted: %v", table, err)
			}
			if err != nil && strings.Contains(err.Error(), "shared-name") {
				t.Fatalf("diagnostic leaks source value: %v", err)
			}
		})
		t.Run(table+"-inactive-still-unique", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "name", "shared-name")
				setFixtureCell(t, f, table, "is_active", "false")
				addFixtureRow(t, f, table, "1000002", map[string]string{"name": "shared-name", "is_active": "false"})
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("inactive duplicate %s name accepted", table)
			}
		})
	}
}

// TestPreflightCutoverMerchandiseActiveOnlyUnique proves duplicate names among inactive merchandise do not collide while active duplicates do.
func TestPreflightCutoverMerchandiseActiveOnlyUnique(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "merchandise_items", "name", "snack")
		addFixtureRow(t, f, "merchandise_items", "1000002", map[string]string{"name": "snack", "is_active": "false"})
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("inactive merchandise duplicate rejected: %v", err)
	}
	for _, truthy := range []string{"true", "t", "yes", "ON", "1", " y "} {
		t.Run("active-"+truthy, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, "merchandise_items", "name", "snack")
				addFixtureRow(t, f, "merchandise_items", "1000002", map[string]string{"name": "snack", "is_active": truthy})
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("active duplicate merchandise name with is_active=%q accepted", truthy)
			}
		})
	}
}

// TestPreflightCutoverOwnerContactUnique pins owner email/phone uniqueness semantics, including empty-value handling.
func TestPreflightCutoverOwnerContactUnique(t *testing.T) {
	for _, column := range []string{"email", "phone"} {
		t.Run(column+"-duplicate", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, "owners", column, "contact-a")
				addFixtureRow(t, f, "owners", "300002", map[string]string{column: "contact-a"})
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), column) {
				t.Fatalf("duplicate owner %s accepted: %v", column, err)
			}
			if err != nil && strings.Contains(err.Error(), "contact-a") {
				t.Fatalf("diagnostic leaks source value: %v", err)
			}
		})
		t.Run(column+"-empty-excluded", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				addFixtureRow(t, f, "owners", "300002", nil)
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("empty %s duplicates rejected: %v", column, err)
			}
		})
	}
}

// TestPreflightCutoverNumberUniqueIncludesEmpty proves record_no/estimate_no uniqueness also binds empty text values because the DDL unique constraint covers them.
func TestPreflightCutoverNumberUniqueIncludesEmpty(t *testing.T) {
	for _, tc := range []struct {
		table, column string
	}{
		{"medical_records", "record_no"},
		{"estimates", "estimate_no"},
	} {
		t.Run(tc.column+"-duplicate", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, tc.table, tc.column, "N-1")
				addFixtureRow(t, f, tc.table, "1000002", map[string]string{tc.column: "N-1"})
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("duplicate %s accepted", tc.column)
			}
		})
		t.Run(tc.column+"-empty-duplicate", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				addFixtureRow(t, f, tc.table, "1000002", nil)
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("duplicate empty %s accepted", tc.column)
			}
		})
	}
}

// TestPreflightCutoverOneToOneUnique rejects shared non-empty reference keys across the one-to-one child tables.
func TestPreflightCutoverOneToOneUnique(t *testing.T) {
	for _, tc := range []struct {
		table, column, parent string
	}{
		{"inquiries", "medical_record_id", "1000001"},
		{"clinical_plans", "medical_record_id", "1000001"},
		{"appointment_trimming_details", "appointment_id", "1000001"},
	} {
		t.Run(tc.table, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				addFixtureRow(t, f, tc.table, "1000002", map[string]string{tc.column: "+001000001"})
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("numeric-equivalent duplicate %s.%s accepted", tc.table, tc.column)
			}
		})
		t.Run(tc.table+"-empty", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, tc.table, tc.column, "")
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
				t.Fatalf("empty %s.%s accepted", tc.table, tc.column)
			}
		})
	}
}

// TestPreflightCutoverTentativePetUnique rejects two tentative appointments for the same pet while other statuses do not collide.
func TestPreflightCutoverTentativePetUnique(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "status", "tentative")
		addFixtureRow(t, f, "appointments", "1000002", map[string]string{"status": "tentative"})
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
		t.Fatal("duplicate tentative appointment for one pet accepted")
	}
	dir, digest = writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "status", "tentative")
		addFixtureRow(t, f, "appointments", "1000002", map[string]string{"status": "tentative", "pet_id": ""})
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("tentative with NULL pet rejected: %v", err)
	}
	dir, digest = writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "status", "confirmed")
		addFixtureRow(t, f, "appointments", "1000002", map[string]string{"status": "confirmed"})
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("non-tentative duplicate rejected: %v", err)
	}
}

// TestPreflightCutoverPetOwnerHistoryPreserved proves historical owner snapshots survive pet transfers without requiring current-owner pairing.
func TestPreflightCutoverPetOwnerHistoryPreserved(t *testing.T) {
	for _, table := range []string{"appointments", "medical_records", "billings", "estimates"} {
		t.Run(table+"-transferred-owner-preserved", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				addFixtureRow(t, f, "owners", "300002", nil)
				setFixtureCell(t, f, table, "owner_id", "300002")
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("historical snapshot owner in %s rejected: %v", table, err)
			}
		})
		t.Run(table+"-null-owner", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "owner_id", "")
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("NULL owner in %s rejected: %v", table, err)
			}
		})
		t.Run(table+"-null-pet", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "pet_id", "")
				if table == "medical_records" {
					f.rows["vital_records"] = nil
					f.manifest.Tables[manifestTableIndex(t, f, "vital_records")].RowCount = 0
					for _, child := range []string{"exams", "vaccinations"} {
						setFixtureCell(t, f, child, "medical_record_id", "")
					}
				}
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("NULL pet in %s rejected: %v", table, err)
			}
		})
	}
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run("independent-snapshots-"+string(profile), func(t *testing.T) {
			mutate := func(f *fixtureBundle) {
				addFixtureRow(t, f, "owners", "300002", nil, profile)
				addFixtureRow(t, f, "owners", "300003", nil, profile)
				setFixtureCell(t, f, "pets", "owner_id", "300003", profile)
				setFixtureCell(t, f, "medical_records", "owner_id", "300001", profile)
				setFixtureCell(t, f, "billings", "owner_id", "300002", profile)
				setFixtureCell(t, f, "estimates", "owner_id", "300001", profile)
				setFixtureCell(t, f, "appointments", "owner_id", "300002", profile)
				setFixtureCell(t, f, "appointments", "status", "tentative", profile)
				for _, table := range []string{"appointments", "medical_records", "billings", "estimates"} {
					setFixtureCell(t, f, table, "pet_id", "1000001", profile)
				}
			}
			var dir, digest string
			var source ExpectedCutoverSource
			if profile == CutoverProfileMedicinesCandidate {
				dir, digest = writeMedicinesCandidateFixture(t, mutate)
				source = candidateExpectedSource(digest)
			} else {
				dir, digest = writeCutoverFixture(t, mutate, profile)
				source = referenceFixtureSource(digest)
			}
			if _, err := PreflightCutoverBundle(dir, source); err != nil {
				t.Fatalf("independent historical snapshots rejected under %s: %v", profile, err)
			}
		})
	}
	t.Run("tentative-stale-owner-preserved", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			addFixtureRow(t, f, "owners", "300002", nil)
			setFixtureCell(t, f, "appointments", "status", "tentative")
			setFixtureCell(t, f, "appointments", "owner_id", "300002")
		})
		if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
			t.Fatalf("tentative appointment with stale snapshot owner rejected: %v", err)
		}
	})
	t.Run("tentative-duplicate-pet-regardless-of-snapshot", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			addFixtureRow(t, f, "owners", "300002", nil)
			setFixtureCell(t, f, "appointments", "status", "tentative")
			setFixtureCell(t, f, "appointments", "owner_id", "300002")
			addFixtureRow(t, f, "appointments", "1000002", map[string]string{
				"status": "tentative", "pet_id": "1000001", "owner_id": "300001",
			})
		})
		_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
		if err == nil || !strings.Contains(err.Error(), "duplicate tentative appointment") {
			t.Fatalf("duplicate tentative pet accepted or wrong diagnostic: %v", err)
		}
	})
}

// TestPreflightCutoverPairingDistinctFromAbsentParent separates missing-parent errors from same-pet mismatch errors on record-pet pairing.
func TestPreflightCutoverPairingDistinctFromAbsentParent(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "pet_id", "1000084")
	})
	_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("absent parent error = %v, want missing reference", err)
	}
	dir, digest = writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "pet_id", "20000001")
	})
	if _, err = PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil || !strings.Contains(err.Error(), "band") {
		t.Fatalf("out-of-band parent error = %v, want band rejection", err)
	}
}

// TestPreflightCutoverRecordPetPairing pins medical_record/pet pairing acceptance and cross-pet rejection cases.
func TestPreflightCutoverRecordPetPairing(t *testing.T) {
	tables := []string{"vital_records", "exams", "vaccinations"}
	for _, table := range tables {
		t.Run(table+"-mismatch", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				addFixtureRow(t, f, "pets", "1000002", nil)
				setFixtureCell(t, f, table, "pet_id", "1000002")
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), "pet does not match the referenced medical record") {
				t.Fatalf("record/pet mismatch in %s accepted: %v", table, err)
			}
			if err != nil && !strings.Contains(err.Error(), "table "+table+" column pet_id row 2") {
				t.Fatalf("mismatch diagnostic names wrong position: %v", err)
			}
		})
		t.Run(table+"-null-record", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "medical_record_id", "")
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if table == "vital_records" {
				if err == nil || !strings.Contains(err.Error(), "table vital_records column medical_record_id row 2") || !strings.Contains(err.Error(), "must not be empty") {
					t.Fatalf("empty required vital medical_record_id accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NULL record in %s rejected: %v", table, err)
			}
		})
		t.Run(table+"-record-null-pet-mismatch", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, "medical_records", "pet_id", "")
				for _, earlier := range tables {
					if earlier == table {
						break
					}
					if earlier == "vital_records" {
						f.rows["vital_records"] = nil
						f.manifest.Tables[manifestTableIndex(t, f, "vital_records")].RowCount = 0
						continue
					}
					setFixtureCell(t, f, earlier, "medical_record_id", "")
				}
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), "pet does not match the referenced medical record") {
				t.Fatalf("present record with NULL pet accepted in %s: %v", table, err)
			}
			if err != nil && !strings.Contains(err.Error(), "table "+table+" column pet_id row 2") {
				t.Fatalf("mismatch diagnostic names wrong position: %v", err)
			}
		})
		t.Run(table+"-absent-record", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "medical_record_id", "1000084")
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), "missing") {
				t.Fatalf("absent record in %s accepted: %v", table, err)
			}
			if err != nil && !strings.Contains(err.Error(), "table "+table+" ") {
				t.Fatalf("absent-record diagnostic names wrong table: %v", err)
			}
		})
		t.Run(table+"-wrong-band-record", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "medical_record_id", "20000001")
			})
			_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
			if err == nil || !strings.Contains(err.Error(), "band") {
				t.Fatalf("out-of-band record in %s accepted: %v", table, err)
			}
			if err != nil && !strings.Contains(err.Error(), "table "+table+" ") {
				t.Fatalf("band diagnostic names wrong table: %v", err)
			}
		})
	}
	for _, table := range []string{"exams", "vaccinations"} {
		t.Run(table+"-null-pet", func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setFixtureCell(t, f, table, "pet_id", "")
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("NULL pet in %s rejected: %v", table, err)
			}
		})
	}
	t.Run("record-null-pet-child-null-pet", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			setFixtureCell(t, f, "medical_records", "pet_id", "")
			f.rows["vital_records"] = nil
			f.manifest.Tables[manifestTableIndex(t, f, "vital_records")].RowCount = 0
			setFixtureCell(t, f, "exams", "pet_id", "")
			setFixtureCell(t, f, "vaccinations", "pet_id", "")
		})
		if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
			t.Fatalf("NULL pet on both sides rejected: %v", err)
		}
	})
}

// TestPreflightCutoverInactiveHistoricalActorsPreserved preserves historical rows referencing inactive staff and deceased owners/pets.
func TestPreflightCutoverInactiveHistoricalActorsPreserved(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		setFixtureCell(t, f, "staffs", "is_active", "false")
		setFixtureCell(t, f, "owners", "membership_type", "deceased")
		setFixtureCell(t, f, "pets", "status", "deceased")
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("inactive/deceased historical rows rejected: %v", err)
	}
}

// TestPreflightCutoverDDLBundleWiringBothProfiles exercises the DDL floor end-to-end through PreflightCutoverBundle under both profiles.
func TestPreflightCutoverDDLBundleWiringBothProfiles(t *testing.T) {
	t.Run("legacy-quantity", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			setFixtureCell(t, f, "billing_items", "quantity", "0")
		})
		if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
			t.Fatal("invalid quantity accepted in legacy profile")
		}
	})
	t.Run("candidate-quantity", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			setFixtureCell(t, f, "billing_items", "quantity", "0")
		}, CutoverProfileMedicinesCandidate)
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err == nil {
			t.Fatal("invalid quantity accepted in candidate profile")
		}
	})
	t.Run("candidate-history-preserved", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			addFixtureRow(t, f, "owners", "300002", nil, CutoverProfileMedicinesCandidate)
			setFixtureCell(t, f, "appointments", "owner_id", "300002", CutoverProfileMedicinesCandidate)
		})
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
			t.Fatalf("historical snapshot owner rejected in candidate profile: %v", err)
		}
	})
	t.Run("candidate-procedures-name-unique", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			setFixtureCell(t, f, "procedures", "name", "shared-name", CutoverProfileMedicinesCandidate)
			addFixtureRow(t, f, "procedures", "1000004", map[string]string{"name": "shared-name"}, CutoverProfileMedicinesCandidate)
		}, CutoverProfileMedicinesCandidate)
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err == nil {
			t.Fatal("duplicate procedures name accepted in candidate profile")
		}
	})
}

// TestPreflightCutoverDDLRulesUnderPermutedHeader keeps DDL checks bound to the named column after an allowed rehearsal header permutation.
func TestPreflightCutoverDDLRulesUnderPermutedHeader(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		mutateRehearsalOnlyHandoff(f)
		spec := ddlSpecByName(t, "owners")
		permuted := make([]string, len(spec.Columns))
		for i, name := range spec.Columns {
			permuted[len(spec.Columns)-1-i] = name
		}
		if f.headerOrder == nil {
			f.headerOrder = map[string][]string{}
		}
		f.headerOrder["owners"] = permuted
		setFixtureCell(t, f, "owners", "email", "contact-a")
		addFixtureRow(t, f, "owners", "300002", map[string]string{"email": "contact-a"})
	})
	_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err == nil || !strings.Contains(err.Error(), "email") {
		t.Fatalf("duplicate owner email under permuted header accepted: %v", err)
	}
}
