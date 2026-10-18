package csvimport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const targetUniqueExpectedMerchandiseNameQuery = `SELECT EXISTS (SELECT 1 FROM public.merchandise_items WHERE clinic_id = $1 AND is_active = true AND deleted_at IS NULL AND name = ANY($2::text[]))`

const targetUniqueExpectedOwnerEmailQuery = `SELECT EXISTS (SELECT 1 FROM public.owners WHERE clinic_id = $1 AND deleted_at IS NULL AND email IS NOT NULL AND email != '' AND email = ANY($2::text[]))`

const targetUniqueExpectedOwnerPhoneQuery = `SELECT EXISTS (SELECT 1 FROM public.owners WHERE clinic_id = $1 AND deleted_at IS NULL AND phone <> '' AND phone = ANY($2::text[]))`

const targetUniqueExpectedMedicalRecordNumberQuery = `SELECT EXISTS (SELECT 1 FROM public.medical_records WHERE clinic_id = $1 AND record_no = ANY($2::text[]))`

const targetUniqueExpectedEstimateNumberQuery = `SELECT EXISTS (SELECT 1 FROM public.estimates WHERE clinic_id = $1 AND estimate_no = ANY($2::text[]))`

const targetUniqueExpectedMedicalRecordAppointmentQuery = `SELECT EXISTS (SELECT 1 FROM public.medical_records WHERE clinic_id = $1 AND deleted_at IS NULL AND appointment_id IS NOT NULL AND appointment_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedInquiryRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.inquiries AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedClinicalPlanRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.clinical_plans AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedTrimmingAppointmentQuery = `SELECT EXISTS (SELECT 1 FROM public.appointment_trimming_details AS child JOIN public.appointments AS parent ON parent.id = child.appointment_id WHERE parent.clinic_id = $1 AND child.appointment_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedBillingRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.billings AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id IS NOT NULL AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedPaymentBillingQuery = `SELECT EXISTS (SELECT 1 FROM public.payments AS child JOIN public.billings AS parent ON parent.id = child.billing_id WHERE parent.clinic_id = $1 AND child.billing_id = ANY($2::text[]::bigint[]))`

const targetUniqueExpectedTentativePetQuery = `SELECT EXISTS (SELECT 1 FROM public.appointments WHERE clinic_id = $1 AND status = 'tentative' AND deleted_at IS NULL AND pet_id = ANY($2::text[]::bigint[]))`

type targetUniqueRuleCase struct {
	name   string
	table  string
	column string
	query  string
	mutate func(t *testing.T, f *fixtureBundle, profile CutoverProfile)
}

// targetUniqueRuleCases lists the new unique-key rules with test-side expected SQL literals and the fixture mutation making each key probeable.
func targetUniqueRuleCases() []targetUniqueRuleCase {
	set := func(table, column, value string) func(*testing.T, *fixtureBundle, CutoverProfile) {
		return func(t *testing.T, f *fixtureBundle, profile CutoverProfile) {
			setFixtureCell(t, f, table, column, value, profile)
		}
	}
	return []targetUniqueRuleCase{
		{"merchandise-active-name", "merchandise_items", "name", targetUniqueExpectedMerchandiseNameQuery, func(t *testing.T, f *fixtureBundle, p CutoverProfile) {
			setFixtureCell(t, f, "merchandise_items", "name", "SOURCE_SECRET-merch-name", p)
			setFixtureCell(t, f, "merchandise_items", "is_active", "true", p)
		}},
		{"owner-email", "owners", "email", targetUniqueExpectedOwnerEmailQuery, set("owners", "email", "SOURCE_SECRET@example.test")},
		{"owner-phone", "owners", "phone", targetUniqueExpectedOwnerPhoneQuery, set("owners", "phone", "SOURCE_SECRET-phone")},
		{"medical-record-no", "medical_records", "record_no", targetUniqueExpectedMedicalRecordNumberQuery, set("medical_records", "record_no", "SOURCE_SECRET-record-no")},
		{"estimate-no", "estimates", "estimate_no", targetUniqueExpectedEstimateNumberQuery, set("estimates", "estimate_no", "SOURCE_SECRET-estimate-no")},
		{"medical-record-appointment", "medical_records", "appointment_id", targetUniqueExpectedMedicalRecordAppointmentQuery, set("medical_records", "appointment_id", "1000001")},
		{"inquiry-record", "inquiries", "medical_record_id", targetUniqueExpectedInquiryRecordQuery, set("inquiries", "medical_record_id", "1000001")},
		{"clinical-plan-record", "clinical_plans", "medical_record_id", targetUniqueExpectedClinicalPlanRecordQuery, set("clinical_plans", "medical_record_id", "1000001")},
		{"trimming-appointment", "appointment_trimming_details", "appointment_id", targetUniqueExpectedTrimmingAppointmentQuery, set("appointment_trimming_details", "appointment_id", "1000001")},
		{"billing-record", "billings", "medical_record_id", targetUniqueExpectedBillingRecordQuery, set("billings", "medical_record_id", "1000001")},
		{"payment-billing", "payments", "billing_id", targetUniqueExpectedPaymentBillingQuery, set("payments", "billing_id", "1000001")},
		{"tentative-appointment-pet", "appointments", "pet_id", targetUniqueExpectedTentativePetQuery, func(t *testing.T, f *fixtureBundle, p CutoverProfile) {
			setFixtureCell(t, f, "appointments", "status", "tentative", p)
			setFixtureCell(t, f, "appointments", "pet_id", "1000001", p)
		}},
	}
}

// targetUniqueAllExpectedQueries returns every expected query literal so fakes can route only the enumerated probes.
func targetUniqueAllExpectedQueries() []string {
	queries := make([]string, 0, len(targetUniqueRuleCases()))
	for _, c := range targetUniqueRuleCases() {
		queries = append(queries, c.query)
	}
	return queries
}

// targetUniqueApplyBundle builds a real preflight-verified apply bundle for a profile with the caller's fixture mutation.
func targetUniqueApplyBundle(t *testing.T, profile CutoverProfile, mutate func(*fixtureBundle)) CutoverBundle {
	t.Helper()
	if profile == CutoverProfileMedicinesCandidate {
		dir, digest := writeMedicinesCandidateFixture(t, mutate)
		bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
		if err != nil {
			t.Fatal(err)
		}
		return bundle
	}
	dir, digest := writeCutoverRehearsalFixture(t, mutate)
	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
		Provenance:     CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal},
	})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

type targetUniqueProbeSession struct {
	candidateFreshApplySession
	collideQuery string
	probes       []string
}

// QueryRow simulates configured new-key collisions and records recognized queries before delegating candidate catalog and band lookups.
func (s *targetUniqueProbeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			s.probes = append(s.probes, expected)
			return staticRow{values: []any{query == s.collideQuery}}
		}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

type targetUniqueLegacyProbeSession struct {
	fakeCutoverSession
	collideQuery string
	probes       []string
}

// QueryRow simulates new-key conflicts while delegating band inspection to the legacy-session fake.
func (s *targetUniqueLegacyProbeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			s.probes = append(s.probes, expected)
			return staticRow{values: []any{query == s.collideQuery}}
		}
	}
	return s.fakeCutoverSession.QueryRow(ctx, query, args...)
}

// TestTargetUniqueNewKeyCollisionBeforeBegin proves each of the 12 new unique keys denies apply before Begin under both profiles, with fixed table/column context and no value echo.
func TestTargetUniqueNewKeyCollisionBeforeBegin(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, c := range targetUniqueRuleCases() {
			t.Run(string(profile)+"/"+c.name, func(t *testing.T) {
				bundle := targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
					c.mutate(t, f, profile)
				})
				seeds := validCutoverSeeds()
				var session cutoverSession
				txCount := func() int { return 0 }
				if profile == CutoverProfileMedicinesCandidate {
					s := &targetUniqueProbeSession{collideQuery: c.query}
					session = s
					txCount = func() int { return len(s.txs) }
				} else {
					s := &targetUniqueLegacyProbeSession{collideQuery: c.query}
					s.inspectCounts = map[string]int64{}
					session = s
					txCount = func() int { return len(s.txs) }
				}
				_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, seeds, pgx.RepeatableRead)
				if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
					t.Fatalf("%s.%s collision accepted: %v", c.table, c.column, err)
				}
				if !strings.Contains(err.Error(), "table "+c.table+" column "+c.column) {
					t.Fatalf("missing fixed table/column context: %v", err)
				}
				if strings.Contains(err.Error(), "SOURCE_SECRET") || strings.Contains(err.Error(), "1000001") {
					t.Fatalf("diagnostic leaks bound value: %v", err)
				}
				if txCount() != 0 {
					t.Fatalf("transactions = %d, want 0 before denial", txCount())
				}
			})
		}
	}
}

// TestTargetUniqueLegacyParentPrefixCollisionBeforeWrite proves ID-key collisions deny the legacy resume before writing, even with a complete parent prefix.
func TestTargetUniqueLegacyParentPrefixCollisionBeforeWrite(t *testing.T) {
	legacyRules := []targetUniqueRuleCase{}
	for _, c := range targetUniqueRuleCases() {
		if c.column == "medical_record_id" || c.column == "appointment_id" || c.column == "billing_id" || c.column == "pet_id" {
			legacyRules = append(legacyRules, c)
		}
	}
	for _, c := range legacyRules {
		t.Run(c.name, func(t *testing.T) {
			bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, func(f *fixtureBundle) {
				c.mutate(t, f, CutoverProfileLegacy21)
			})
			counts := map[string]int64{}
			childSeen := false
			for _, spec := range CutoverTableSpecs(bundle.Manifest.profile) {
				if spec.Name == c.table {
					childSeen = true
				}
				if !childSeen {
					counts[spec.Name] = 1
				}
			}
			s := &targetUniqueLegacyProbeSession{collideQuery: c.query}
			s.inspectCounts = counts
			_, err := applyCutoverCommittingEachTable(context.Background(), s, bundle, validCutoverSeeds(), pgx.RepeatableRead)
			if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
				t.Fatalf("%s.%s collision accepted under complete parent prefix: %v", c.table, c.column, err)
			}
			if !strings.Contains(err.Error(), "table "+c.table+" column "+c.column) {
				t.Fatalf("missing fixed table/column context: %v", err)
			}
			if len(s.txs) != 0 {
				t.Fatalf("transactions = %d, want 0 before denial", len(s.txs))
			}
		})
	}
}

type targetUniqueRecorder struct {
	queries []string
	binds   [][]string
	clinics []int64
	exists  bool
	errAt   string
	err     error
}

// QueryRow captures seed IDs and raw batches for recognized queries, can inject a chosen cause, and rejects unknown SQL.
func (r *targetUniqueRecorder) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			r.queries = append(r.queries, query)
			if len(args) == 2 {
				r.clinics = append(r.clinics, args[0].(int64))
				if bound, ok := args[1].([]string); ok {
					r.binds = append(r.binds, append([]string(nil), bound...))
				}
			}
			if r.err != nil && query == r.errAt {
				return staticRow{err: r.err}
			}
			return staticRow{values: []any{r.exists}}
		}
	}
	return staticRow{err: errors.New("unexpected query: " + query)}
}

// targetUniqueRuleByName looks up the actual production rule for a named test case so tests bind the real configured rule.
func targetUniqueRuleByName(t *testing.T, name string) cutoverTargetUniqueRule {
	t.Helper()
	for _, c := range targetUniqueRuleCases() {
		if c.name == name {
			for _, rule := range cutoverTargetUniqueRules {
				if rule.table == c.table && rule.column == c.column {
					return rule
				}
			}
			t.Fatalf("production rule for %s.%s missing", c.table, c.column)
		}
	}
	t.Fatal("unknown case " + name)
	return cutoverTargetUniqueRule{}
}

// TestTargetUniqueRuleMapPinned pins the production rule list to the 14 distinct table/column pairs including the two original name rules.
func TestTargetUniqueRuleMapPinned(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range cutoverTargetUniqueRules {
		key := rule.table + "." + rule.column
		if seen[key] {
			t.Fatalf("duplicate rule %s", key)
		}
		seen[key] = true
	}
	if len(seen) != 14 {
		t.Fatalf("rule count = %d, want 14", len(seen))
	}
	for _, c := range targetUniqueRuleCases() {
		if !seen[c.table+"."+c.column] {
			t.Fatalf("root rule %s.%s missing from production list", c.table, c.column)
		}
	}
	if !seen["procedures.name"] || !seen["vaccines.name"] {
		t.Fatal("existing name rules dropped")
	}
}

// TestTargetUniqueRuleBindingExact asserts each production rule's query equals the independent expected literal and binds seeds.ClinicID plus the raw fixture value.
func TestTargetUniqueRuleBindingExact(t *testing.T) {
	for _, c := range targetUniqueRuleCases() {
		t.Run(c.name, func(t *testing.T) {
			bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
				c.mutate(t, f, CutoverProfileMedicinesCandidate)
			})
			r := &targetUniqueRecorder{}
			if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, c.name)); err != nil {
				t.Fatal(err)
			}
			if len(r.queries) != 1 || r.queries[0] != c.query {
				t.Fatalf("queries = %v, want exact %q", r.queries, c.query)
			}
			seeds := validCutoverSeeds()
			if r.clinics[0] != seeds.ClinicID {
				t.Fatalf("clinic = %d, want %d", r.clinics[0], seeds.ClinicID)
			}
			want := expectedTargetValue(c.name)
			if len(r.binds) != 1 || len(r.binds[0]) != 1 || r.binds[0][0] != want {
				t.Fatalf("bound keys = %q, want exactly [%q]", r.binds, want)
			}
		})
	}
}

// TestTargetUniqueSelectionFilters proves empty owner contacts are skipped while whitespace, quotes, and Unicode values bind verbatim under both profiles.
func TestTargetUniqueSelectionFilters(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, c := range []targetUniqueRuleCase{targetUniqueRuleCases()[1], targetUniqueRuleCases()[2]} {
			t.Run(string(profile)+"/"+c.name+"/empty", func(t *testing.T) {
				bundle := targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, "owners", c.column, "", profile)
				})
				r := &targetUniqueRecorder{}
				if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, c.name)); err != nil {
					t.Fatal(err)
				}
				for _, q := range r.queries {
					if q == c.query {
						t.Fatalf("empty %s must not be probed", c.column)
					}
				}
			})
			for _, tc := range []struct {
				name  string
				value string
			}{{"whitespace-only", " \t "}, {"quotes-unicode", " \t ' 名 ' "}} {
				t.Run(string(profile)+"/"+c.name+"/"+tc.name, func(t *testing.T) {
					bundle := targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
						setFixtureCell(t, f, "owners", c.column, tc.value, profile)
					})
					r := &targetUniqueRecorder{}
					if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, c.name)); err != nil {
						t.Fatal(err)
					}
					matched := 0
					for _, q := range r.queries {
						if q == c.query {
							matched++
						}
					}
					if matched != 1 {
						t.Fatalf("probe count for %s = %d, want exactly 1", c.name, matched)
					}
					if len(r.binds) != 1 || len(r.binds[0]) != 1 || r.binds[0][0] != tc.value {
						t.Fatalf("raw bind = %v, want exact %q", r.binds, tc.value)
					}
				})
			}
		}
	}
}

// TestTargetUniqueEmptyTextStillProbed proves empty record_no/estimate_no still bind as raw empty strings because the DDL uniqueness covers them.
func TestTargetUniqueEmptyTextStillProbed(t *testing.T) {
	for _, c := range []targetUniqueRuleCase{targetUniqueRuleCases()[3], targetUniqueRuleCases()[4]} {
		bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
			setFixtureCell(t, f, c.table, c.column, "", CutoverProfileMedicinesCandidate)
		})
		r := &targetUniqueRecorder{}
		if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, c.name)); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, q := range r.queries {
			if q == c.query {
				found = true
			}
		}
		if !found {
			t.Fatalf("empty %s.%s must still be probed (DDL permits empty text)", c.table, c.column)
		}
		if len(r.binds) != 1 || r.binds[0][0] != "" {
			t.Fatalf("expected raw empty string bound: %v", r.binds)
		}
	}
}

// TestTargetUniqueMerchActiveOnlyFilter pins the merchandise active-only selector across PostgreSQL boolean spellings and invalid-literal rejection.
func TestTargetUniqueMerchActiveOnlyFilter(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{{"false", false}, {"0", false}, {"off", false}, {"n", false}, {"true", true}, {"1", true}, {"t", true}, {"on", true}, {"yes", true}} {
		t.Run(tc.value, func(t *testing.T) {
			bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
				setFixtureCell(t, f, "merchandise_items", "is_active", tc.value, CutoverProfileMedicinesCandidate)
				setFixtureCell(t, f, "merchandise_items", "name", "SOURCE_SECRET-merch-"+tc.value, CutoverProfileMedicinesCandidate)
			})
			r := &targetUniqueRecorder{}
			if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "merchandise-active-name")); err != nil {
				t.Fatal(err)
			}
			probed := false
			for _, q := range r.queries {
				if q == targetUniqueExpectedMerchandiseNameQuery {
					probed = true
				}
			}
			if probed != tc.want {
				t.Fatalf("is_active=%s probed=%v, want %v", tc.value, probed, tc.want)
			}
		})
	}
	bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, nil)
	digest := rewriteTableCSV(t, bundle, "merchandise_items", func(header []string, rows [][]string) [][]string {
		idx := columnIndex(merchSpec(t, bundle).Columns, "is_active")
		for _, row := range rows {
			row[idx] = "notabool"
		}
		return rows
	})
	bundle = updateManifestEntry(t, bundle, "merchandise_items", digest, 1)
	r := &targetUniqueRecorder{}
	err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "merchandise-active-name"))
	if err == nil || !strings.Contains(err.Error(), "table merchandise_items column is_active row 2:") {
		t.Fatalf("invalid boolean not rejected with fixed coordinate: %v", err)
	}
	if strings.Contains(err.Error(), "notabool") {
		t.Fatalf("diagnostic leaks cell value: %v", err)
	}
	if len(r.queries) != 0 {
		t.Fatalf("queries ran before invalid boolean rejection: %v", r.queries)
	}
}

// TestTargetUniqueLeadingZeroRawBind proves leading-zero ID spellings reach the bind slice verbatim for the SQL cast to compare numerically.
func TestTargetUniqueLeadingZeroRawBind(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
		setFixtureCell(t, f, "medical_records", "appointment_id", "0001000001", CutoverProfileMedicinesCandidate)
	})
	r := &targetUniqueRecorder{}
	if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "medical-record-appointment")); err != nil {
		t.Fatal(err)
	}
	if len(r.binds) != 1 || r.binds[0][0] != "0001000001" {
		t.Fatalf("leading-zero ID not bound verbatim: %v", r.binds)
	}
}

// TestTargetUniqueAtomicCollisionRollsBack proves a new-key collision inside the atomic transaction rolls back after Begin with no COPY or commit.
func TestTargetUniqueAtomicCollisionRollsBack(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, func(f *fixtureBundle) {
		setFixtureCell(t, f, "owners", "email", "SOURCE_SECRET@example.test", CutoverProfileLegacy21)
	})
	tx := &targetUniqueAtomicTx{fakeCutoverTransaction: &fakeCutoverTransaction{}, collideQuery: targetUniqueExpectedOwnerEmailQuery}
	begins := 0
	_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
		begins++
		return tx, nil
	}, bundle, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("atomic collision accepted: %v", err)
	}
	if begins != 1 || !tx.rolledBack || tx.committed || tx.copyCalls != 0 {
		t.Fatalf("begins=%d rollback=%v commit=%v copies=%d", begins, tx.rolledBack, tx.committed, tx.copyCalls)
	}
}

// TestTargetUniqueNonNameCausePreserved proves a non-name probe error preserves its cause while Error() stays sanitized.
func TestTargetUniqueNonNameCausePreserved(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
		setFixtureCell(t, f, "owners", "email", "SOURCE_SECRET@example.test", CutoverProfileMedicinesCandidate)
	})
	cause := errors.New("DRIVER-SECRET-/tmp/patient-phi-path")
	r := &targetUniqueRecorder{err: cause, errAt: targetUniqueExpectedOwnerEmailQuery}
	err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "owner-email"))
	if err == nil || !errors.Is(err, cause) {
		t.Fatalf("cause not preserved: %v", err)
	}
	if strings.Contains(err.Error(), "DRIVER-SECRET") || strings.Contains(err.Error(), "patient-phi-path") {
		t.Fatalf("diagnostic leaks driver detail: %v", err)
	}
}

type targetUniqueAtomicTx struct {
	*fakeCutoverTransaction
	collideQuery string
}

// QueryRow supplies in-transaction unique-key collision responses and delegates the remaining target checks.
func (tx *targetUniqueAtomicTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			return staticRow{values: []any{query == tx.collideQuery}}
		}
	}
	return tx.fakeCutoverTransaction.QueryRow(ctx, query, args...)
}

// merchSpec resolves the merchandise_items spec for the bundle's profile.
func merchSpec(t *testing.T, bundle CutoverBundle) CutoverTableSpec {
	t.Helper()
	for _, spec := range CutoverTableSpecs(bundle.Manifest.profile) {
		if spec.Name == "merchandise_items" {
			return spec
		}
	}
	t.Fatal("merchandise_items spec missing")
	return CutoverTableSpec{}
}

// bindsContain reports whether the bound slice contains the wanted value.
func bindsContain(bound []string, want string) bool {
	for _, v := range bound {
		if v == want {
			return true
		}
	}
	return false
}

// expectedTargetValue returns the fixture-mutated key value each case expects bound verbatim.
func expectedTargetValue(name string) string {
	switch name {
	case "merchandise-active-name":
		return "SOURCE_SECRET-merch-name"
	case "owner-email":
		return "SOURCE_SECRET@example.test"
	case "owner-phone":
		return "SOURCE_SECRET-phone"
	case "medical-record-no":
		return "SOURCE_SECRET-record-no"
	case "estimate-no":
		return "SOURCE_SECRET-estimate-no"
	case "medical-record-appointment", "inquiry-record", "clinical-plan-record", "trimming-appointment", "billing-record", "payment-billing", "tentative-appointment-pet":
		return "1000001"
	}
	return ""
}

// TestTargetUniqueExcludedRowsStillHashed proves filtered-out rows still consume the full CSV so tampered data and wrong row counts are rejected before any query.
func TestTargetUniqueExcludedRowsStillHashed(t *testing.T) {
	t.Run("stale-hash-on-excluded-row", func(t *testing.T) {
		bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
			setFixtureCell(t, f, "merchandise_items", "is_active", "false", CutoverProfileMedicinesCandidate)
		})
		rewriteTableCSV(t, bundle, "merchandise_items", func(header []string, rows [][]string) [][]string {
			idx := columnIndex(merchSpec(t, bundle).Columns, "name")
			for _, row := range rows {
				row[idx] = "TAMPERED-excluded-row"
			}
			return rows
		})
		r := &targetUniqueRecorder{}
		err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "merchandise-active-name"))
		if err == nil || !strings.Contains(err.Error(), "sha256") {
			t.Fatalf("tampered excluded row not detected: %v", err)
		}
		if len(r.queries) != 0 {
			t.Fatalf("query ran before hash verification: %v", r.queries)
		}
	})
	t.Run("wrong-rowcount-on-excluded", func(t *testing.T) {
		bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
			setFixtureCell(t, f, "merchandise_items", "is_active", "false", CutoverProfileMedicinesCandidate)
		})
		bundle = corruptManifestRowCount(t, bundle, "merchandise_items")
		r := &targetUniqueRecorder{}
		err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "merchandise-active-name"))
		if err == nil || !strings.Contains(err.Error(), "row count") {
			t.Fatalf("row count mismatch on excluded rows accepted: %v", err)
		}
	})
}

// TestTargetUniqueNullableNumericEmptySkipped proves empty nullable numeric keys are skipped rather than queried.
func TestTargetUniqueNullableNumericEmptySkipped(t *testing.T) {
	for _, c := range []targetUniqueRuleCase{targetUniqueRuleCases()[5], targetUniqueRuleCases()[9]} {
		t.Run(c.name, func(t *testing.T) {
			bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
				setFixtureCell(t, f, c.table, c.column, "", CutoverProfileMedicinesCandidate)
			})
			r := &targetUniqueRecorder{}
			if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, c.name)); err != nil {
				t.Fatal(err)
			}
			for _, q := range r.queries {
				if q == c.query {
					t.Fatalf("empty nullable %s.%s must not be probed", c.table, c.column)
				}
			}
		})
	}
}

// TestTargetUniqueTentativeSelector pins pet binding to tentative status only; confirmed, cancelled, and empty statuses are excluded.
func TestTargetUniqueTentativeSelector(t *testing.T) {
	for _, tc := range []struct {
		status string
		want   bool
	}{{"tentative", true}, {"confirmed", false}, {"cancelled", false}, {"", false}} {
		t.Run("status="+tc.status, func(t *testing.T) {
			bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, func(f *fixtureBundle) {
				setFixtureCell(t, f, "appointments", "status", tc.status, CutoverProfileMedicinesCandidate)
				setFixtureCell(t, f, "appointments", "pet_id", "1000001", CutoverProfileMedicinesCandidate)
			})
			r := &targetUniqueRecorder{}
			if err := validateCutoverTargetUniqueValues(context.Background(), r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "tentative-appointment-pet")); err != nil {
				t.Fatal(err)
			}
			probed := false
			for _, q := range r.queries {
				if q == targetUniqueExpectedTentativePetQuery {
					probed = true
				}
			}
			if probed != tc.want {
				t.Fatalf("status=%q probed=%v, want %v", tc.status, probed, tc.want)
			}
		})
	}
}

type targetUniqueOrderTx struct {
	*fakeCutoverTransaction
	catalog cutoverQuerier
	events  []string
}

// Exec records LOCK TABLE statements in the event log and delegates to the wrapped transaction.
func (tx *targetUniqueOrderTx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(query, "LOCK TABLE ") {
		tx.events = append(tx.events, "table-lock")
	}
	return tx.fakeCutoverTransaction.Exec(ctx, query, args...)
}

// QueryRow records recognized unique-key probes before returning the simulated owner-email conflict or delegating catalog checks.
func (tx *targetUniqueOrderTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			tx.events = append(tx.events, "probe:"+query[:50])
			return staticRow{values: []any{query == targetUniqueExpectedOwnerEmailQuery}}
		}
	}
	if tx.catalog != nil {
		return tx.catalog.QueryRow(ctx, query, args...)
	}
	return tx.fakeCutoverTransaction.QueryRow(ctx, query, args...)
}

// CopyFrom records COPY calls in the event log and delegates to the wrapped transaction.
func (tx *targetUniqueOrderTx) CopyFrom(ctx context.Context, reader io.Reader, sql string) (pgconn.CommandTag, error) {
	tx.events = append(tx.events, "copy")
	return tx.fakeCutoverTransaction.CopyFrom(ctx, reader, sql)
}

// TestTargetUniqueAtomicLocksBeforeProbeBeforeCopy proves the single LOCK TABLE runs before the target-unique probe and denied probes never reach COPY, under both profiles.
func TestTargetUniqueAtomicLocksBeforeProbeBeforeCopy(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile), func(t *testing.T) {
			targetUniqueAtomicCollisionOnce(t, profile)
		})
	}
}

// targetUniqueAtomicCollisionOnce drives one atomic-seam collision for a profile and asserts the observed lock/probe/copy ordering.
func targetUniqueAtomicCollisionOnce(t *testing.T, profile CutoverProfile) {
	bundle := targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
		setFixtureCell(t, f, "owners", "email", "SOURCE_SECRET@example.test", profile)
	})
	inner := &fakeCutoverTransaction{}
	tx := &targetUniqueOrderTx{fakeCutoverTransaction: inner}
	if profile == CutoverProfileMedicinesCandidate {
		tx.catalog = &candidateAtomicTransaction{fakeCutoverTransaction: inner, counts: map[string]int64{}}
	}
	begins := 0
	_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
		begins++
		return tx, nil
	}, bundle, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("atomic collision accepted: %v", err)
	}
	if begins != 1 || !tx.rolledBack || tx.committed || tx.copyCalls != 0 {
		t.Fatalf("begins=%d rollback=%v commit=%v copies=%d", begins, tx.rolledBack, tx.committed, tx.copyCalls)
	}
	lastLock, firstProbe, copyIdx := -1, -1, -1
	lockCount := 0
	for i, e := range tx.events {
		if e == "table-lock" {
			lastLock = i
			lockCount++
		}
		if strings.HasPrefix(e, "probe:") && firstProbe < 0 {
			firstProbe = i
		}
		if e == "copy" {
			copyIdx = i
		}
	}
	if firstProbe < 0 {
		t.Fatal("no target-unique probe recorded")
	}
	if lastLock < 0 || firstProbe < lastLock {
		t.Fatalf("locks not before probe: events=%v", tx.events)
	}
	if copyIdx >= 0 {
		t.Fatalf("COPY ran after denied probe: events=%v", tx.events)
	}
	if lockCount != 1 {
		t.Fatalf("locks = %d, want the single LOCK TABLE statement", lockCount)
	}
}

// TestTargetUniqueNonNameCauseApplyMatrix proves non-name probe causes (cancel or driver error) survive errors.Is with sanitized Error() and no Begin, under both profiles.
func TestTargetUniqueNonNameCauseApplyMatrix(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, cause := range []error{context.Canceled, errors.New("DRIVER-SECRET-/tmp/patient-phi-path")} {
			t.Run(string(profile), func(t *testing.T) {
				bundle := targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, "owners", "email", "SOURCE_SECRET@example.test", profile)
				})
				s := &targetUniqueApplyErrSession{probeErr: cause, probeErrQuery: targetUniqueExpectedOwnerEmailQuery}
				_, err := applyCutoverCommittingEachTable(context.Background(), s, bundle, validCutoverSeeds(), pgx.RepeatableRead)
				if err == nil || !errors.Is(err, cause) {
					t.Fatalf("cause not preserved: %v", err)
				}
				if strings.Contains(err.Error(), "DRIVER-SECRET") || strings.Contains(err.Error(), "patient-phi-path") {
					t.Fatalf("diagnostic leaks driver detail: %v", err)
				}
				if len(s.txs) != 0 {
					t.Fatalf("transactions = %d, want 0", len(s.txs))
				}
			})
		}
	}
}

type targetUniqueApplyErrSession struct {
	candidateFreshApplySession
	probeErr      error
	probeErrQuery string
}

// QueryRow injects the configured cause only for its selected unique-key probe and delegates other session lookups.
func (s *targetUniqueApplyErrSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			if s.probeErr != nil && query == s.probeErrQuery {
				return staticRow{err: s.probeErr}
			}
			return staticRow{values: []any{false}}
		}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

// cancelledContext deterministically reports its cause after cancelAt Err() checks.
type cancelledContext struct {
	context.Context
	checks   int
	cancelAt int
	cause    error
}

func (c *cancelledContext) Err() error {
	c.checks++
	if c.checks >= c.cancelAt {
		return c.cause
	}
	return nil
}

func TestTargetUniqueCancelDuringScan(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, func(f *fixtureBundle) {
		setFixtureCell(t, f, "merchandise_items", "is_active", "false", CutoverProfileLegacy21)
		for i := 0; i < 32; i++ {
			addFixtureRow(t, f, "merchandise_items", fmt.Sprintf("%d", 1000030+i), map[string]string{
				"name":      fmt.Sprintf("SOURCE_SECRET-%d", i),
				"is_active": "false",
			}, CutoverProfileLegacy21)
		}
	})
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(cause.Error(), func(t *testing.T) {
			ctx := &cancelledContext{Context: context.Background(), cancelAt: 5, cause: cause}
			r := &targetUniqueRecorder{}
			err := validateCutoverTargetUniqueValues(ctx, r, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "merchandise-active-name"))
			if !errors.Is(err, cause) {
				t.Fatalf("err = %v, want errors.Is %v", err, cause)
			}
			if !strings.Contains(err.Error(), ": cancelled") {
				t.Fatalf("missing fixed cancel context: %v", err)
			}
			if len(r.queries) != 0 {
				t.Fatalf("queries ran during cancelled scan: %v", r.queries)
			}
			if ctx.checks != ctx.cancelAt {
				t.Fatalf("checks = %d, want exactly %d (stop during read)", ctx.checks, ctx.cancelAt)
			}
		})
	}
}

// TestTargetUniqueCancelBeforeOpen prefers the safe cancellation cause over opening a missing source.
func TestTargetUniqueCancelBeforeOpen(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, nil)
	bundle.SourceDir = "/SOURCE_SECRET-missing"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	recorder := &targetUniqueRecorder{}
	err := validateCutoverTargetUniqueValues(ctx, recorder, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "owner-email"))
	if !errors.Is(err, context.Canceled) || err.Error() != "table owners column email unique preflight: cancelled" ||
		len(recorder.queries) != 0 {
		t.Fatalf("pre-open cancellation = %v", err)
	}
}

type targetUniqueCancelBatchQuerier struct {
	t      *testing.T
	cancel context.CancelFunc
	calls  int
}

// QueryRow cancels only after the first explicitly matched batch, independently of driver cancellation behavior.
func (q *targetUniqueCancelBatchQuerier) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	q.calls++
	if query != targetUniqueExpectedOwnerEmailQuery || len(args) != 2 || len(args[1].([]string)) != 256 {
		q.t.Fatal("unexpected query or batch after cancellation")
	}
	q.cancel()
	return staticRow{values: []any{false}}
}

// TestTargetUniqueCancelBetweenBatches prevents a second query when the first successful batch cancels the context.
func TestTargetUniqueCancelBetweenBatches(t *testing.T) {
	bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, nil)
	digest := rewriteTableCSV(t, bundle, "owners", func(header []string, rows [][]string) [][]string {
		out := make([][]string, 257)
		for i := range out {
			out[i] = append([]string(nil), rows[0]...)
			out[i][columnIndex(header, "email")] = fmt.Sprintf("SOURCE_SECRET-%d@example.test", i)
		}
		return out
	})
	bundle = updateManifestEntry(t, bundle, "owners", digest, 257)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := &targetUniqueCancelBatchQuerier{t: t, cancel: cancel}
	err := validateCutoverTargetUniqueValues(ctx, q, bundle, validCutoverSeeds(), targetUniqueRuleByName(t, "owner-email"))
	if !errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "SOURCE_SECRET") || q.calls != 1 {
		t.Fatalf("between-batch cancellation error=%v calls=%d", err, q.calls)
	}
}
