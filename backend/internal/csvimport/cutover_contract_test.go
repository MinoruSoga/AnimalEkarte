package csvimport

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestPreflightCutoverBundleAcceptsExactContract(t *testing.T) {
	dir, manifestDigest := writeCutoverFixture(t, nil)

	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: manifestDigest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err != nil {
		t.Fatalf("PreflightCutoverBundle() error = %v", err)
	}
	if len(bundle.Manifest.Tables) != 21 {
		t.Fatalf("table count = %d, want 21", len(bundle.Manifest.Tables))
	}
	if bundle.Manifest.IDBand.ApplicationIDFloor != 1_000_000_000 {
		t.Fatalf("application floor = %d", bundle.Manifest.IDBand.ApplicationIDFloor)
	}
	if bundle.Manifest.HandoffEligibility != "TRUSTED_CANDIDATE" {
		t.Fatalf("handoff eligibility = %q, want TRUSTED_CANDIDATE", bundle.Manifest.HandoffEligibility)
	}
}

func TestPreflightCutoverBundleAcceptsNegativeRefundAmounts(t *testing.T) {
	dir, manifestDigest := writeCutoverFixture(t, func(f *fixtureBundle) {
		billingColumns := CutoverTableSpecs()[11].Columns
		paymentColumns := CutoverTableSpecs()[13].Columns
		splitColumns := CutoverTableSpecs()[14].Columns
		f.rows["billings"][0][columnIndex(billingColumns, "total_amount")] = "-3000"
		f.rows["payments"][0][columnIndex(paymentColumns, "subtotal")] = "-3000"
		f.rows["payments"][0][columnIndex(paymentColumns, "tax_total")] = "0"
		f.rows["payments"][0][columnIndex(paymentColumns, "total_amount")] = "-3000"
		f.rows["payments"][0][columnIndex(paymentColumns, "billing_amount")] = "-3000"
		f.rows["payments"][0][columnIndex(paymentColumns, "received_amount")] = "-3000"
		f.rows["payments"][0][columnIndex(paymentColumns, "change_amount")] = "0"
		f.rows["payment_splits"][0][columnIndex(splitColumns, "amount")] = "-3000"
		f.rows["payment_splits"][0][columnIndex(splitColumns, "received_amount")] = "-3000"
		f.rows["payment_splits"][0][columnIndex(splitColumns, "change_amount")] = "0"
	})

	if _, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: manifestDigest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	}); err != nil {
		t.Fatalf("PreflightCutoverBundle(negative refund) error = %v", err)
	}
}

func TestPreflightCutoverBundleAcceptsWindowZeroCompletedBillingWithoutPayment(t *testing.T) {
	dir, manifestDigest := writeCutoverFixture(t, func(f *fixtureBundle) {
		billingColumns := CutoverTableSpecs()[11].Columns
		zero := append([]string(nil), f.rows["billings"][0]...)
		zero[columnIndex(billingColumns, "id")] = "1000002"
		zero[columnIndex(billingColumns, "total_amount")] = "0"
		f.rows["billings"] = append(f.rows["billings"], zero)
		f.manifest.Tables[11].RowCount++
	})

	if _, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: manifestDigest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	}); err != nil {
		t.Fatalf("PreflightCutoverBundle(window-zero completed) error = %v", err)
	}
}

func TestPreflightCutoverBundleAcceptsLocalRehearsalBundle(t *testing.T) {
	dir, manifestDigest := writeCutoverFixture(t, func(f *fixtureBundle) {
		f.manifest.Status = "REHEARSAL_ONLY"
		f.manifest.HandoffEligibility = "REHEARSAL_ONLY"
		f.manifest.SourceCompletenessStatus = "UNVERIFIED"
		f.manifest.SourceComplete = false
		f.manifest.SourceProvenanceVerified = false
		f.manifest.StageMappingSHA256 = strings.Repeat("a", 64)
		empty := []string{}
		f.manifest.IncompleteSourceTables = &empty
		billingColumns := CutoverTableSpecs()[11].Columns
		f.rows["billings"][0][columnIndex(billingColumns, "status")] = "pending"
		f.rows["billings"][0][columnIndex(billingColumns, "completed_at")] = ""
		f.rows["payments"] = nil
		f.rows["payment_splits"] = nil
		f.manifest.Tables[13].RowCount = 0
		f.manifest.Tables[14].RowCount = 0
	})

	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: manifestDigest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
		Provenance:     CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal},
	})
	if err != nil {
		t.Fatalf("PreflightCutoverBundle(local rehearsal) error = %v", err)
	}
	if bundle.Manifest.Status != "REHEARSAL_ONLY" {
		t.Fatalf("status = %q, want REHEARSAL_ONLY", bundle.Manifest.Status)
	}
}

func TestStagingRehearsalProvenanceRequiresTargetBinding(t *testing.T) {
	manifest := CutoverManifest{}
	for _, contract := range []CutoverProvenanceContract{
		{Mode: CutoverProvenanceStagingRehearsal},
		{Mode: CutoverProvenanceStagingRehearsal, Target: CutoverTargetBinding{Environment: "local", Host: "db", Database: "ekarte", ClinicID: 1}},
	} {
		if err := validateCutoverProducerProvenance(manifest, contract); err == nil || !strings.Contains(err.Error(), "target binding") {
			t.Fatalf("contract %+v error = %v, want target binding rejection", contract, err)
		}
	}
}

func TestPreflightCutoverBundleRejectsInvalidProducerProvenance(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fixtureBundle)
		wantErr string
	}{
		{
			name: "non-trusted handoff",
			mutate: func(f *fixtureBundle) {
				f.manifest.HandoffEligibility = "REHEARSAL_ONLY"
			},
			wantErr: "handoff eligibility",
		},
		{
			name: "unsupported manifest schema",
			mutate: func(f *fixtureBundle) {
				f.manifest.ManifestSchemaVersion = "animalekarte-cutover-v2"
			},
			wantErr: "mapping contract",
		},
		{
			name: "untrusted stage mapping",
			mutate: func(f *fixtureBundle) {
				f.manifest.StageMappingSHA256 = strings.Repeat("0", 64)
			},
			wantErr: "mapping contract",
		},
		{
			name: "untrusted CSV contract",
			mutate: func(f *fixtureBundle) {
				f.manifest.CSVContractSHA256 = strings.Repeat("0", 64)
			},
			wantErr: "mapping contract",
		},
		{
			name: "partial source",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceCompletenessStatus = "PARTIAL"
				f.manifest.SourceComplete = false
			},
			wantErr: "source completeness",
		},
		{
			name: "unverified source identity",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceIdentity.Verified = false
			},
			wantErr: "source identity",
		},
		{
			name: "missing source backup digest",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceIdentity.SourceBackupSHA256 = nil
			},
			wantErr: "source identity",
		},
		{
			name: "invalid stage build ID",
			mutate: func(f *fixtureBundle) {
				f.manifest.StageBuildID = "not-a-uuid"
			},
			wantErr: "stage build",
		},
		{
			name: "incomplete source table",
			mutate: func(f *fixtureBundle) {
				f.manifest.IncompleteSourceTables = stringSlicePointer([]string{"TBL_KNJO_DATA"})
			},
			wantErr: "incomplete source",
		},
		{
			name: "missing incomplete source table inventory",
			mutate: func(f *fixtureBundle) {
				f.manifest.IncompleteSourceTables = nil
			},
			wantErr: "incomplete source",
		},
		{
			name: "invalid summary digest",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceSummarySHA256.Raw = "invalid"
			},
			wantErr: "summary digest",
		},
		{
			name: "out-of-order summary timestamps",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceSummaryGeneratedAt.Raw = "2026-07-22T00:00:03Z"
			},
			wantErr: "summary generation",
		},
		{
			name: "stage summary is newer than manifest",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceSummaryGeneratedAt.Stage = "2026-07-22T00:00:04Z"
			},
			wantErr: "summary generation",
		},
		{
			name: "invalid evidence digest",
			mutate: func(f *fixtureBundle) {
				f.manifest.SourceEvidenceSHA256.BaseLoad = "invalid"
			},
			wantErr: "evidence digest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, tt.mutate)
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Fatalf("error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

func TestCutoverTableSpecsMatchPaymentContract(t *testing.T) {
	specs := CutoverTableSpecs()
	if len(specs) != 21 {
		t.Fatalf("table count = %d, want 21", len(specs))
	}
	gotNames := make([]string, 0, len(specs))
	for _, spec := range specs {
		gotNames = append(gotNames, spec.Name)
	}
	wantNames := []string{
		"staffs", "procedures", "merchandise_items", "owners", "pets",
		"medical_records", "inquiries", "clinical_plans", "vital_records",
		"appointments", "appointment_trimming_details", "billings",
		"billing_items", "payments", "payment_splits", "estimates",
		"estimate_items", "exams", "exam_results", "vaccines", "vaccinations",
	}
	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("table order = %v, want %v", gotNames, wantNames)
	}

	billings := specs[11]
	if !slices.Equal(billings.Columns, []string{
		"id", "clinic_id", "medical_record_id", "owner_id", "pet_id",
		"total_amount", "status", "scheduled_date", "completed_at",
	}) {
		t.Fatalf("billings columns = %v", billings.Columns)
	}
	payments := specs[13]
	if !slices.Equal(payments.Columns, []string{
		"id", "clinic_id", "billing_id", "subtotal", "tax_total", "total_amount",
		"insurance_name", "insurance_ratio", "insurance_amount", "discount_amount",
		"billing_amount", "received_amount", "change_amount", "method",
		"payment_method_id", "paid_by", "created_at",
	}) {
		t.Fatalf("payments columns = %v", payments.Columns)
	}
	if !slices.Equal(payments.BandColumns, []string{"id", "billing_id", "paid_by"}) {
		t.Fatalf("payments band columns = %v", payments.BandColumns)
	}
	paymentSplits := specs[14]
	if !slices.Equal(paymentSplits.Columns, []string{
		"id", "clinic_id", "billing_id", "method", "payment_method_id", "amount",
		"received_amount", "change_amount", "paid_by", "created_at",
	}) {
		t.Fatalf("payment_splits columns = %v", paymentSplits.Columns)
	}
	if !slices.Equal(paymentSplits.BandColumns, []string{"id", "billing_id", "paid_by"}) {
		t.Fatalf("payment_splits band columns = %v", paymentSplits.BandColumns)
	}

	wantPlaceholders := map[string]string{
		"staffs.clinic_id":            "{{CLINIC_ID}}",
		"procedures.clinic_id":        "{{CLINIC_ID}}",
		"merchandise_items.clinic_id": "{{CLINIC_ID}}",
		"owners.clinic_id":            "{{CLINIC_ID}}",
		"pets.clinic_id":              "{{CLINIC_ID}}",
		"pets.animal_species_id (fallback only, when unresolved)": "{{FALLBACK_ANIMAL_SPECIES_ID}}",
		"medical_records.clinic_id":                               "{{CLINIC_ID}}",
		"vital_records.clinic_id":                                 "{{CLINIC_ID}}",
		"appointments.clinic_id":                                  "{{CLINIC_ID}}",
		"appointments.reservation_type_id":                        "{{TRIMMING_RESERVATION_TYPE_ID}}",
		"appointment_trimming_details.clinic_id":                  "{{CLINIC_ID}}",
		"billings.clinic_id":                                      "{{CLINIC_ID}}",
		"payments.clinic_id":                                      "{{CLINIC_ID}}",
		"payments.payment_method_id (cash)":                       "{{PAYMENT_METHOD_CASH_ID}}",
		"payments.payment_method_id (credit_card)":                "{{PAYMENT_METHOD_CREDIT_CARD_ID}}",
		"payment_splits.clinic_id":                                "{{CLINIC_ID}}",
		"payment_splits.payment_method_id (cash)":                 "{{PAYMENT_METHOD_CASH_ID}}",
		"payment_splits.payment_method_id (credit_card)":          "{{PAYMENT_METHOD_CREDIT_CARD_ID}}",
		"estimates.clinic_id":                                     "{{CLINIC_ID}}",
		"exams.clinic_id":                                         "{{CLINIC_ID}}",
		"exams.exam_type_id":                                      "{{FALLBACK_EXAM_TYPE_ID}}",
		"vaccines.clinic_id":                                      "{{CLINIC_ID}}",
		"vaccinations.clinic_id":                                  "{{CLINIC_ID}}",
	}
	if got := CutoverPlaceholderColumns(); !reflect.DeepEqual(got, wantPlaceholders) {
		t.Fatalf("placeholder inventory = %v, want %v", got, wantPlaceholders)
	}
}

func TestCutoverTableSpecsDeclareNonNullableTextColumns(t *testing.T) {
	expected := map[string][]string{
		"staffs":                       {"name", "license_number"},
		"procedures":                   {"name", "description"},
		"merchandise_items":            {"name"},
		"owners":                       {"name", "name_kana", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks"},
		"pets":                         {"pet_number", "name", "name_kana", "breed", "color", "food", "remarks"},
		"medical_records":              {"record_no"},
		"inquiries":                    {"chief_complaint", "owner_observations", "history", "notes", "allergy_info", "current_medications"},
		"clinical_plans":               {"physical_exam", "diagnosis_details", "treatment_policy"},
		"vital_records":                {"notes"},
		"appointments":                 nil,
		"appointment_trimming_details": {"remarks"},
		"billings":                     nil,
		"billing_items":                {"name"},
		"payments":                     {"insurance_name"},
		"payment_splits":               nil,
		"estimates":                    {"estimate_no", "title", "comment", "notes"},
		"estimate_items":               {"name"},
		"exams":                        {"result_summary"},
		"exam_results":                 {"name", "inspection_value", "normal_value"},
		"vaccines":                     {"name", "description", "interval"},
		"vaccinations":                 {"supplemental", "lot1", "lot2", "lot3", "lot4", "remarks"},
	}

	for _, spec := range CutoverTableSpecs() {
		if !slices.Equal(spec.ForceNotNullColumns, expected[spec.Name]) {
			t.Errorf("%s force-not-null columns = %v, want %v", spec.Name, spec.ForceNotNullColumns, expected[spec.Name])
		}
	}
}

// TestCutoverMappingCoverageCoversAllFormalTables locks the Issue #250
// source→target inventory: 21 formal tables, isolation mode, non-PHI fields only.
func TestCutoverMappingCoverageCoversAllFormalTables(t *testing.T) {
	specs := CutoverTableSpecs()
	coverage := CutoverMappingCoverage()
	if len(coverage) != len(specs) {
		t.Fatalf("coverage count = %d, want %d", len(coverage), len(specs))
	}
	if len(coverage) != 21 {
		t.Fatalf("formal table count = %d, want 21", len(coverage))
	}

	// Child tables without clinic_id must still declare parent-FK isolation.
	wantParentFK := map[string]struct{}{
		"inquiries": {}, "clinical_plans": {}, "billing_items": {},
		"estimate_items": {}, "exam_results": {},
	}

	seen := make(map[string]struct{}, len(coverage))
	for i, entry := range coverage {
		spec := specs[i]
		if entry.TargetTable != spec.Name {
			t.Fatalf("coverage[%d].TargetTable = %q, want %q", i, entry.TargetTable, spec.Name)
		}
		if entry.CSVFile != spec.Name+".csv" {
			t.Fatalf("%s CSVFile = %q", entry.TargetTable, entry.CSVFile)
		}
		if entry.ColumnCount != len(spec.Columns) {
			t.Fatalf("%s ColumnCount = %d, want %d", entry.TargetTable, entry.ColumnCount, len(spec.Columns))
		}
		if !slices.Equal(entry.BandColumns, spec.BandColumns) {
			t.Fatalf("%s BandColumns = %v, want %v", entry.TargetTable, entry.BandColumns, spec.BandColumns)
		}
		if entry.ConsumerStatus != "formal_cutover_v1" {
			t.Fatalf("%s ConsumerStatus = %q", entry.TargetTable, entry.ConsumerStatus)
		}
		expectClinicID := hasColumn(spec.Columns, "clinic_id")
		if entry.HasClinicID != expectClinicID {
			t.Fatalf("%s HasClinicID = %v, want %v", entry.TargetTable, entry.HasClinicID, expectClinicID)
		}
		if expectClinicID {
			if entry.IsolationCheck != CutoverIsolationClinicID {
				t.Fatalf("%s IsolationCheck = %q, want %q", entry.TargetTable, entry.IsolationCheck, CutoverIsolationClinicID)
			}
		} else {
			if _, ok := wantParentFK[entry.TargetTable]; !ok {
				t.Fatalf("unexpected table without clinic_id: %s", entry.TargetTable)
			}
			if entry.IsolationCheck != CutoverIsolationParentFK {
				t.Fatalf("%s IsolationCheck = %q, want %q", entry.TargetTable, entry.IsolationCheck, CutoverIsolationParentFK)
			}
		}
		// Inventory must stay non-PHI: no free-text payload fields.
		blob := entry.TargetTable + entry.CSVFile + entry.IsolationCheck + entry.ConsumerStatus
		for _, forbidden := range []string{"private-", "password", "phone", "address", "email@", "owner name"} {
			if strings.Contains(strings.ToLower(blob), forbidden) {
				t.Fatalf("mapping inventory looks PHI-bearing (%q): %+v", forbidden, entry)
			}
		}
		if _, dup := seen[entry.TargetTable]; dup {
			t.Fatalf("duplicate mapping entry for %s", entry.TargetTable)
		}
		seen[entry.TargetTable] = struct{}{}
	}

	// Mutation safety: returned BandColumns must be a copy.
	coverage[0].BandColumns[0] = "mutated"
	fresh := CutoverMappingCoverage()
	if fresh[0].BandColumns[0] == "mutated" {
		t.Fatal("CutoverMappingCoverage returned a shared BandColumns slice")
	}
}

func TestPreflightCutoverBundleFailsClosed(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fixtureBundle)
		wantErr string
	}{
		{
			name: "manifest table order drift",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0], f.manifest.Tables[1] = f.manifest.Tables[1], f.manifest.Tables[0]
			},
			wantErr: "table order",
		},
		{
			name: "CSV digest mismatch",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					file := filepath.Join(dir, "owners.csv")
					contents, _ := os.ReadFile(file)
					_ = os.WriteFile(file, append(contents, '\n'), 0o600)
				}
			},
			wantErr: "sha256",
		},
		{
			name: "row count mismatch",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].RowCount = 2
			},
			wantErr: "row count",
		},
		{
			name: "unsafe manifest filename",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].File = "../staffs.csv"
			},
			wantErr: "filename",
		},
		{
			name: "unknown placeholder",
			mutate: func(f *fixtureBundle) {
				f.rows["owners"][0][2] = "{{UNKNOWN_ID}}"
			},
			wantErr: "placeholder",
		},
		{
			name: "ID outside clinic band",
			mutate: func(f *fixtureBundle) {
				f.rows["staffs"][0][0] = "10000000"
			},
			wantErr: "clinic band",
		},
		{
			name: "duplicate primary ID",
			mutate: func(f *fixtureBundle) {
				f.rows["owners"] = append(f.rows["owners"], append([]string(nil), f.rows["owners"][0]...))
				f.manifest.Tables[3].RowCount = 2
			},
			wantErr: "duplicate",
		},
		{
			name: "unexpected file",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.WriteFile(filepath.Join(dir, "unexpected.csv"), []byte("id\n"), 0o600)
				}
			},
			wantErr: "unexpected file",
		},
		{
			name: "group readable source",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.Chmod(filepath.Join(dir, "owners.csv"), 0o640)
				}
			},
			wantErr: "owner-only",
		},
		{
			name: "oversized CSV",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.Truncate(filepath.Join(dir, "owners.csv"), maxCutoverCSVBytes+1)
				}
			},
			wantErr: "size limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, tt.mutate)
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

func TestPreflightCutoverBundleRejectsPaymentContractViolations(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fixtureBundle)
		wantErr string
	}{
		{
			name: "duplicate payment billing",
			mutate: func(f *fixtureBundle) {
				row := append([]string(nil), f.rows["payments"][0]...)
				row[columnIndex(CutoverTableSpecs()[13].Columns, "id")] = "1000002"
				f.rows["payments"] = append(f.rows["payments"], row)
				f.manifest.Tables[13].RowCount++
			},
			wantErr: "billing_id",
		},
		{
			name: "payment split without payment parent",
			mutate: func(f *fixtureBundle) {
				f.rows["payment_splits"][0][columnIndex(CutoverTableSpecs()[14].Columns, "billing_id")] = "1000002"
			},
			wantErr: "payment parent",
		},
		{
			name: "payment method placeholder disagrees with method",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "payment_method_id")] =
					"{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
			},
			wantErr: "does not match method",
		},
		{
			name: "payment clinic is not an explicit placeholder",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "clinic_id")] = "1"
			},
			wantErr: "clinic placeholder",
		},
		{
			name: "payment split clinic is not an explicit placeholder",
			mutate: func(f *fixtureBundle) {
				f.rows["payment_splits"][0][columnIndex(CutoverTableSpecs()[14].Columns, "clinic_id")] = "1"
			},
			wantErr: "clinic placeholder",
		},
		{
			name: "payment billing amount is zero",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "billing_amount")] = "0"
				f.rows["payment_splits"][0][columnIndex(CutoverTableSpecs()[14].Columns, "amount")] = "0"
			},
			wantErr: "violate the cutover contract",
		},
		{
			name: "payment total amount disagrees with billing",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "total_amount")] = "999"
			},
			wantErr: "payment snapshot",
		},
		{
			name: "insurance ratio is outside contract range",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "insurance_ratio")] = "1.01"
			},
			wantErr: "insurance_ratio",
		},
		{
			name: "insurance amount is negative",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "insurance_amount")] = "-1"
			},
			wantErr: "insurance_amount",
		},
		{
			name: "payment split amount is zero",
			mutate: func(f *fixtureBundle) {
				f.rows["payment_splits"][0][columnIndex(CutoverTableSpecs()[14].Columns, "amount")] = "0"
			},
			wantErr: "must not be zero",
		},
		{
			name: "completed billing timestamp is missing",
			mutate: func(f *fixtureBundle) {
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "completed_at")] = ""
			},
			wantErr: "completed_at",
		},
		{
			name: "completed billing payment graph is missing",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"] = nil
				f.rows["payment_splits"] = nil
				f.manifest.Tables[13].RowCount = 0
				f.manifest.Tables[14].RowCount = 0
			},
			wantErr: "must contain formal rows",
		},
		{
			name: "formal handoff cannot omit the entire payment graph",
			mutate: func(f *fixtureBundle) {
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "status")] = "waiting"
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "completed_at")] = ""
				f.rows["payments"] = nil
				f.rows["payment_splits"] = nil
				f.manifest.Tables[13].RowCount = 0
				f.manifest.Tables[14].RowCount = 0
			},
			wantErr: "must contain formal rows",
		},
		{
			name: "payment parent billing is not completed",
			mutate: func(f *fixtureBundle) {
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "status")] = "waiting"
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "completed_at")] = ""
			},
			wantErr: "billing status",
		},
		{
			name: "empty billing status cannot bypass payment graph",
			mutate: func(f *fixtureBundle) {
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "status")] = ""
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "completed_at")] = ""
			},
			wantErr: "billing status",
		},
		{
			name: "unknown billing status cannot bypass payment graph",
			mutate: func(f *fixtureBundle) {
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "status")] = "unknown"
				f.rows["billings"][0][columnIndex(CutoverTableSpecs()[11].Columns, "completed_at")] = ""
			},
			wantErr: "billing status",
		},
		{
			name: "payment timestamp disagrees with billing completion",
			mutate: func(f *fixtureBundle) {
				f.rows["payments"][0][columnIndex(CutoverTableSpecs()[13].Columns, "created_at")] = "2026-07-22T00:00:01Z"
				f.rows["payment_splits"][0][columnIndex(CutoverTableSpecs()[14].Columns, "created_at")] = "2026-07-22T00:00:01Z"
			},
			wantErr: "completion timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, tt.mutate)
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCutoverPaymentGraphRejectsUnboundedPaymentInventory(t *testing.T) {
	manifest := CutoverManifest{Tables: []CutoverManifestTable{
		{Table: "payments", File: "payments.csv", RowCount: maxCutoverPaymentRows + 1},
		{Table: "payment_splits", File: "payment_splits.csv", RowCount: 0},
	}}
	err := validateCutoverPaymentGraph(t.TempDir(), &manifest)
	if err == nil || !strings.Contains(err.Error(), "payment limit") {
		t.Fatalf("validateCutoverPaymentGraph() error = %v, want payment limit rejection", err)
	}
}

func TestPreflightCutoverBundleAcceptsCreditCardAndMixedPaymentGraphs(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*fixtureBundle)
	}{
		{
			name: "positive insurance amount",
			mutate: func(f *fixtureBundle) {
				paymentColumns := CutoverTableSpecs()[13].Columns
				f.rows["payments"][0][columnIndex(paymentColumns, "insurance_ratio")] = "0.50"
				f.rows["payments"][0][columnIndex(paymentColumns, "insurance_amount")] = "500"
			},
		},
		{
			name: "credit card only",
			mutate: func(f *fixtureBundle) {
				paymentColumns := CutoverTableSpecs()[13].Columns
				splitColumns := CutoverTableSpecs()[14].Columns
				f.rows["payments"][0][columnIndex(paymentColumns, "method")] = "credit_card"
				f.rows["payments"][0][columnIndex(paymentColumns, "payment_method_id")] =
					"{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
				f.rows["payments"][0][columnIndex(paymentColumns, "received_amount")] = "0"
				f.rows["payment_splits"][0][columnIndex(splitColumns, "method")] = "credit_card"
				f.rows["payment_splits"][0][columnIndex(splitColumns, "payment_method_id")] =
					"{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
				f.rows["payment_splits"][0][columnIndex(splitColumns, "received_amount")] = "0"
			},
		},
		{
			name: "mixed",
			mutate: func(f *fixtureBundle) {
				paymentColumns := CutoverTableSpecs()[13].Columns
				splitColumns := CutoverTableSpecs()[14].Columns
				f.rows["payments"][0][columnIndex(paymentColumns, "billing_amount")] = "1500"
				card := append([]string(nil), f.rows["payment_splits"][0]...)
				card[columnIndex(splitColumns, "id")] = "1000002"
				card[columnIndex(splitColumns, "method")] = "credit_card"
				card[columnIndex(splitColumns, "payment_method_id")] =
					"{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
				card[columnIndex(splitColumns, "amount")] = "500"
				card[columnIndex(splitColumns, "received_amount")] = "0"
				f.rows["payment_splits"] = append(f.rows["payment_splits"], card)
				f.manifest.Tables[14].RowCount++
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, tt.mutate)
			if _, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			}); err != nil {
				t.Fatalf("PreflightCutoverBundle() error = %v", err)
			}
		})
	}
}

func TestReadOwnerOnlyRegularFileRejectsOversizedManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxCutoverManifestBytes + 1); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readOwnerOnlyRegularFile(path); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("readOwnerOnlyRegularFile() error = %v, want size-limit rejection", err)
	}
}

func TestPreflightCutoverBundleRejectsSymlinkedCSV(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	owners := filepath.Join(dir, "owners.csv")
	target := filepath.Join(dir, "owners-real.csv")
	if err := os.Rename(owners, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, owners); err != nil {
		t.Fatal(err)
	}

	_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "symbolic link") {
		t.Fatalf("error = %v, want symbolic link rejection", err)
	}
}

type fixtureBundle struct {
	manifest      CutoverManifest
	rows          map[string][][]string
	afterManifest func(string)
}

func writeCutoverFixture(t *testing.T, mutate func(*fixtureBundle)) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	f := fixtureBundle{
		manifest: CutoverManifest{
			GeneratedAt:               "2026-07-22T00:00:03Z",
			Status:                    "PASS",
			SourceLayer:               "animalekarte_stage",
			SourceRunID:               "run-1",
			ClinicCode:                "hachioji",
			ClinicOrdinal:             1,
			ClinicBandBase:            0,
			ClinicBandEndExclusive:    10_000_000,
			StageIDOffset:             1_000_000,
			IDBand:                    CutoverIDBand{Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000, OwnerFloor: 300_000, ApplicationIDFloor: 1_000_000_000},
			OutputDir:                 "sensitive-local/animalekarte-csv-export/hachioji/run-1",
			Format:                    "csv-with-header",
			ImportablePredicate:       "mapping_status IN ('confirmed','inferred'), plus per-table CSV FK-parent eligibility guards (see scripts/lib/stage-csv-columns.mjs)",
			PlaceholderColumns:        CutoverPlaceholderColumns(),
			PlaceholderResolutionNote: "source placeholders",
			ManifestSchemaVersion:     cutoverManifestSchema,
			StageMappingSHA256:        cutoverStageMappingSHA256,
			CSVContractSHA256:         cutoverCSVContractSHA256,
			SourceCompletenessStatus:  "PASS",
			SourceComplete:            true,
			SourceProvenanceVerified:  true,
			SourceIdentity: CutoverSourceIdentity{
				SourceBackupSHA256:    stringPointer(strings.Repeat("a", 64)),
				SourceBackupSizeBytes: int64Pointer(7_475_357_184),
				BaseArchiveSHA256:     stringPointer(strings.Repeat("b", 64)),
				KNJOArchiveSHA256:     stringPointer(strings.Repeat("c", 64)),
				Verified:              true,
			},
			StageBuildID:           "3ed169df-441b-4515-b128-3b182e53f84a",
			IncompleteSourceTables: stringSlicePointer([]string{}),
			HandoffEligibility:     "TRUSTED_CANDIDATE",
			SourceSummarySHA256: CutoverLayerDigests{
				Raw: strings.Repeat("d", 64), Intermediate: strings.Repeat("e", 64), Stage: strings.Repeat("f", 64),
			},
			SourceSummaryGeneratedAt: CutoverLayerTimestamps{
				Raw: "2026-07-22T00:00:00Z", Intermediate: "2026-07-22T00:00:01Z", Stage: "2026-07-22T00:00:02Z",
			},
			SourceEvidenceSHA256: CutoverEvidenceDigests{
				BaseLoad: strings.Repeat("1", 64), KNJORecovery: strings.Repeat("2", 64),
			},
		},
		rows: map[string][][]string{},
	}

	for _, spec := range CutoverTableSpecs() {
		row := make([]string, len(spec.Columns))
		for i, column := range spec.Columns {
			switch column {
			case "id":
				if spec.Name == "owners" {
					row[i] = "300001"
				} else {
					row[i] = "1000001"
				}
			case "clinic_id":
				row[i] = "{{CLINIC_ID}}"
			case "animal_species_id":
				row[i] = "{{FALLBACK_ANIMAL_SPECIES_ID}}"
			case "exam_type_id":
				row[i] = "{{FALLBACK_EXAM_TYPE_ID}}"
			case "reservation_type_id":
				row[i] = "{{TRIMMING_RESERVATION_TYPE_ID}}"
			case "method":
				if spec.Name == "payments" || spec.Name == "payment_splits" {
					row[i] = "cash"
				}
			case "status":
				if spec.Name == "billings" {
					row[i] = "completed"
				}
			case "payment_method_id":
				if spec.Name == "payments" || spec.Name == "payment_splits" {
					row[i] = "{{PAYMENT_METHOD_CASH_ID}}"
				}
			case "subtotal", "total_amount", "billing_amount", "received_amount", "amount":
				if spec.Name == "billings" && column == "total_amount" {
					row[i] = "1000"
				} else if spec.Name == "payments" || spec.Name == "payment_splits" {
					row[i] = "1000"
				}
			case "tax_total", "insurance_ratio", "insurance_amount", "discount_amount", "change_amount":
				if spec.Name == "payments" || spec.Name == "payment_splits" {
					row[i] = "0"
				}
			case "created_at":
				if spec.Name == "payments" || spec.Name == "payment_splits" {
					row[i] = "2026-07-22T00:00:00.000000Z"
				}
			case "completed_at":
				if spec.Name == "billings" {
					row[i] = "2026-07-22T00:00:00.000000Z"
				}
			}
		}
		for _, column := range spec.BandColumns {
			idx := columnIndex(spec.Columns, column)
			if idx < 0 || row[idx] != "" || column == "clinic_id" {
				continue
			}
			if column == "owner_id" {
				row[idx] = "300001"
			} else {
				row[idx] = "1000001"
			}
		}
		f.rows[spec.Name] = [][]string{row}
		f.manifest.Tables = append(f.manifest.Tables, CutoverManifestTable{
			Table: spec.Name, File: spec.Name + ".csv", RowCount: 1,
		})
	}
	if mutate != nil {
		mutate(&f)
	}

	for i, spec := range CutoverTableSpecs() {
		fileName := spec.Name + ".csv"
		file := filepath.Join(dir, fileName)
		out, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		w := csv.NewWriter(out)
		if err := w.Write(spec.Columns); err != nil {
			t.Fatal(err)
		}
		for _, row := range f.rows[spec.Name] {
			if err := w.Write(row); err != nil {
				t.Fatal(err)
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			t.Fatal(err)
		}
		if err := out.Close(); err != nil {
			t.Fatal(err)
		}
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(contents)
		f.manifest.Tables[i].SHA256 = hex.EncodeToString(sum[:])
	}

	manifestBytes, err := json.MarshalIndent(f.manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if f.afterManifest != nil {
		f.afterManifest(dir)
	}
	sum := sha256.Sum256(manifestBytes)
	return dir, hex.EncodeToString(sum[:])
}

func stringPointer(value string) *string {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}

func stringSlicePointer(value []string) *[]string {
	return &value
}

func columnIndex(columns []string, name string) int {
	for i, column := range columns {
		if column == name {
			return i
		}
	}
	return -1
}
