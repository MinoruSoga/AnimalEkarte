package csvimport

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type masterNameProbeSession struct {
	candidateFreshApplySession
	procExists    bool
	vaccExists    bool
	probeErrTable string
	procProbes    int
	vaccProbes    int
	procSQL       string
	vaccSQL       string
	procClinic    int64
	vaccClinic    int64
	procNames     []string
	vaccNames     []string
	order         []string
	probeErr      error
}

// QueryRow simulates procedure/vaccine name lookups, recording their SQL and arguments before delegating other queries.
func (s *masterNameProbeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "public.procedures") && strings.Contains(query, "name = ANY") {
		s.procProbes++
		s.procSQL = query
		if len(args) == 2 {
			s.procClinic, _ = args[0].(int64)
			if bound, ok := args[1].([]string); ok {
				s.procNames = append([]string(nil), bound...)
			}
		}
		s.order = append(s.order, "procedures")
		if s.probeErr != nil && (s.probeErrTable == "" || s.probeErrTable == "procedures") {
			return staticRow{err: s.probeErr}
		}
		return staticRow{values: []any{s.procExists}}
	}
	if strings.Contains(query, "public.vaccines") && strings.Contains(query, "name = ANY") {
		s.vaccProbes++
		s.vaccSQL = query
		if len(args) == 2 {
			s.vaccClinic, _ = args[0].(int64)
			if bound, ok := args[1].([]string); ok {
				s.vaccNames = append([]string(nil), bound...)
			}
		}
		s.order = append(s.order, "vaccines")
		if s.probeErr != nil && (s.probeErrTable == "" || s.probeErrTable == "vaccines") {
			return staticRow{err: s.probeErr}
		}
		return staticRow{values: []any{s.vaccExists}}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

// TestMasterNameProcedureCollisionCandidateRejected proves a same-clinic procedures name collision is denied before Begin under the candidate profile.
func TestMasterNameProcedureCollisionCandidateRejected(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &masterNameProbeSession{procExists: true}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("candidate procedures name collision accepted: %v", err)
	}
	if session.procProbes == 0 {
		t.Fatal("apply never probed candidate procedures names")
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// TestMasterNameVaccineCollisionCandidateRejected proves a same-clinic vaccines name collision is denied before Begin under the candidate profile.
func TestMasterNameVaccineCollisionCandidateRejected(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &masterNameProbeSession{vaccExists: true}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("candidate vaccines name collision accepted: %v", err)
	}
	if session.vaccProbes == 0 {
		t.Fatal("apply never probed candidate vaccines names")
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// TestMasterNameProbeBindsExactQueryAndNames asserts the exact static query text, clinic seed, verbatim names, and procedures-before-vaccines order.
func TestMasterNameProbeBindsExactQueryAndNames(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &masterNameProbeSession{}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	wantProc := `SELECT EXISTS (SELECT 1 FROM public.procedures WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`
	wantVacc := `SELECT EXISTS (SELECT 1 FROM public.vaccines WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`
	if session.procSQL != wantProc {
		t.Fatalf("procedures probe SQL = %q, want exact %q", session.procSQL, wantProc)
	}
	if session.vaccSQL != wantVacc {
		t.Fatalf("vaccines probe SQL = %q, want exact %q", session.vaccSQL, wantVacc)
	}
	if len(session.order) < 2 || session.order[0] != "procedures" || session.order[1] != "vaccines" {
		t.Fatalf("probe order = %v, want procedures first then vaccines", session.order)
	}
	seeds := validCutoverSeeds()
	if session.procClinic != seeds.ClinicID || session.vaccClinic != seeds.ClinicID {
		t.Fatalf("probe clinics = %d/%d, want %d", session.procClinic, session.vaccClinic, seeds.ClinicID)
	}
	for _, table := range []string{"procedures", "vaccines"} {
		names := sourceCSVNames(t, bundle, table)
		got := session.procNames
		if table == "vaccines" {
			got = session.vaccNames
		}
		if len(names) != len(got) {
			t.Fatalf("%s probe names = %v, want exact %v", table, got, names)
		}
		for i := range names {
			if got[i] != names[i] {
				t.Fatalf("%s probe name %d = %q, want exact %q", table, i, got[i], names[i])
			}
		}
	}
}

// TestMasterNameProbeQueryErrorFailsClosed proves a probe error surfaces a fixed name-preflight context without driver/path detail, before Begin.
func TestMasterNameProbeQueryErrorFailsClosed(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &masterNameProbeSession{probeErr: errors.New("DRIVER-SECRET-/tmp/patient-phi-path lookup failed")}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), "name preflight") {
		t.Fatalf("error = %v, want fixed table-labeled name preflight failure", err)
	}
	if strings.Contains(err.Error(), "DRIVER-SECRET") || strings.Contains(err.Error(), "patient-phi-path") {
		t.Fatalf("diagnostic leaks driver/source detail: %v", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

type legacyNameProbeSession struct {
	fakeCutoverSession
	procExists bool
	vaccExists bool
	procProbes int
	vaccProbes int
	probeErr   error
}

// QueryRow simulates legacy name lookups and records whether pending procedure/vaccine tables were probed.
func (s *legacyNameProbeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "public.procedures") && strings.Contains(query, "name = ANY") {
		s.procProbes++
		if s.probeErr != nil {
			return staticRow{err: s.probeErr}
		}
		return staticRow{values: []any{s.procExists}}
	}
	if strings.Contains(query, "public.vaccines") && strings.Contains(query, "name = ANY") {
		s.vaccProbes++
		if s.probeErr != nil {
			return staticRow{err: s.probeErr}
		}
		return staticRow{values: []any{s.vaccExists}}
	}
	return s.fakeCutoverSession.QueryRow(ctx, query, args...)
}

// TestMasterNameProcedureCollisionLegacyPendingRejected proves a pending procedures name collision denies the legacy path before any transaction.
func TestMasterNameProcedureCollisionLegacyPendingRejected(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &legacyNameProbeSession{procExists: true}
	session.inspectCounts = map[string]int64{}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("legacy pending procedures name collision accepted: %v", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// TestMasterNameVaccineCollisionLegacyPendingRejected proves a pending vaccines name collision denies the legacy path before any transaction.
func TestMasterNameVaccineCollisionLegacyPendingRejected(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &legacyNameProbeSession{vaccExists: true}
	session.inspectCounts = map[string]int64{}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("legacy pending vaccines name collision accepted: %v", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// TestMasterNameCompleteProcedurePrefixSkipsOwnNames proves a complete staffs+procedures prefix skips its own name probe while pending vaccines are still probed.
func TestMasterNameCompleteProcedurePrefixSkipsOwnNames(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &legacyNameProbeSession{procExists: true}
	session.inspectCounts = map[string]int64{"staffs": 1, "procedures": 1}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatalf("resume over complete staffs+procedures prefix failed: %v", err)
	}
	if session.procProbes != 0 {
		t.Fatalf("complete procedures prefix was probed %d times, want 0", session.procProbes)
	}
	if session.vaccProbes == 0 {
		t.Fatal("pending vaccines table was not probed")
	}
	for i, tx := range session.txs {
		for _, copySQL := range tx.copySQLs {
			if strings.Contains(copySQL, "procedures") || strings.Contains(copySQL, "staffs") {
				t.Fatalf("tx %d copied complete-prefix table: %s", i, copySQL)
			}
		}
	}
}

// TestMasterNameAtomicCollisionAfterBeginRollsBack proves a procedures collision inside the atomic transaction rolls back before any COPY or commit.
func TestMasterNameAtomicCollisionAfterBeginRollsBack(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := &atomicNameProbeTx{fakeCutoverTransaction: &fakeCutoverTransaction{}, procExists: true}
	begins := 0
	_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
		begins++
		return tx, nil
	}, bundle, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("atomic procedures collision accepted: %v", err)
	}
	if begins != 1 {
		t.Fatalf("begin calls = %d, want exactly 1 (atomic denial is after Begin)", begins)
	}
	if !tx.rolledBack {
		t.Fatal("collision must roll the transaction back")
	}
	if tx.committed {
		t.Fatal("collision must never commit")
	}
	if tx.copyCalls != 0 {
		t.Fatalf("copy calls = %d, want 0 before denial", tx.copyCalls)
	}
}

type atomicNameProbeTx struct {
	*fakeCutoverTransaction
	procExists bool
	vaccExists bool
	probes     int
}

// QueryRow injects procedure/vaccine name collisions inside the fake transaction and delegates other target checks.
func (tx *atomicNameProbeTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "public.procedures") && strings.Contains(query, "name = ANY") {
		tx.probes++
		return staticRow{values: []any{tx.procExists}}
	}
	if strings.Contains(query, "public.vaccines") && strings.Contains(query, "name = ANY") {
		tx.probes++
		return staticRow{values: []any{tx.vaccExists}}
	}
	return tx.fakeCutoverTransaction.QueryRow(ctx, query, args...)
}

// TestMasterNameAtomicVaccineCollisionRollsBack proves a vaccines collision inside the atomic transaction rolls back before any COPY or commit.
func TestMasterNameAtomicVaccineCollisionRollsBack(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := &atomicNameProbeTx{fakeCutoverTransaction: &fakeCutoverTransaction{}, vaccExists: true}
	begins := 0
	_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
		begins++
		return tx, nil
	}, bundle, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("atomic vaccines collision accepted: %v", err)
	}
	if begins != 1 {
		t.Fatalf("begin calls = %d, want exactly 1", begins)
	}
	if !tx.rolledBack || tx.committed {
		t.Fatalf("rollback=%v committed=%v, want rolled back and never committed", tx.rolledBack, tx.committed)
	}
	if tx.copyCalls != 0 {
		t.Fatalf("copy calls = %d, want 0 before denial", tx.copyCalls)
	}
}

// TestMasterNameLegacyQueryErrorBeforeBegin proves a legacy probe failure returns the fixed context with no Begin.
func TestMasterNameLegacyQueryErrorBeforeBegin(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &legacyNameProbeSession{probeErr: errors.New("driver failure")}
	session.inspectCounts = map[string]int64{}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), "name preflight") {
		t.Fatalf("error = %v, want fixed name preflight failure", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// sourceCSVNames reads the fixture CSV for a table and returns the verbatim name-column values in row order.
func sourceCSVNames(t *testing.T, bundle CutoverBundle, table string) []string {
	t.Helper()
	manifestTable, err := cutoverManifestTableByName(bundle.Manifest, table)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(cutoverCSVPath(bundle.SourceDir, manifestTable))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	rows, err := csv.NewReader(bufio.NewReader(f)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, row := range rows[1:] {
		var spec CutoverTableSpec
		for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
			if candidate.Name == table {
				spec = candidate
			}
		}
		names = append(names, row[columnIndex(spec.Columns, "name")])
	}
	return names
}

type masterReaderProbeQuerier struct {
	procProbes    int
	vaccProbes    int
	chunks        []int
	vaccChunks    []int
	procNamesFlat []string
	vaccNamesFlat []string
	exists        bool
}

// QueryRow records procedure/vaccine names and chunk sizes, answers enumerated extra unique-key probes, and rejects unknown SQL.
func (q *masterReaderProbeQuerier) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "public.procedures") && strings.Contains(query, "name = ANY") {
		q.procProbes++
		if len(args) == 2 {
			if bound, ok := args[1].([]string); ok {
				q.chunks = append(q.chunks, len(bound))
				q.procNamesFlat = append(q.procNamesFlat, bound...)
			}
		}
		return staticRow{values: []any{q.exists}}
	}
	if strings.Contains(query, "public.vaccines") && strings.Contains(query, "name = ANY") {
		q.vaccProbes++
		if len(args) == 2 {
			if bound, ok := args[1].([]string); ok {
				q.vaccChunks = append(q.vaccChunks, len(bound))
				q.vaccNamesFlat = append(q.vaccNamesFlat, bound...)
			}
		}
		return staticRow{values: []any{q.exists}}
	}
	for _, expected := range targetUniqueAllExpectedQueries() {
		if query == expected {
			return staticRow{values: []any{q.exists}}
		}
	}
	return staticRow{err: errors.New("unexpected query: " + query)}
}

// corruptManifestHash replaces a manifest table's digest with an all-zero hash.
func corruptManifestHash(t *testing.T, bundle CutoverBundle, table string) CutoverBundle {
	t.Helper()
	for i := range bundle.Manifest.Tables {
		if bundle.Manifest.Tables[i].Table == table {
			bundle.Manifest.Tables[i].SHA256 = strings.Repeat("0", 64)
		}
	}
	return bundle
}

// corruptManifestRowCount bumps a manifest table's declared row count by one.
func corruptManifestRowCount(t *testing.T, bundle CutoverBundle, table string) CutoverBundle {
	t.Helper()
	for i := range bundle.Manifest.Tables {
		if bundle.Manifest.Tables[i].Table == table {
			bundle.Manifest.Tables[i].RowCount++
		}
	}
	return bundle
}

// TestMasterNameReaderCorruptHashFailsBeforeQuery checks that a digest mismatch prevents probing the affected table's names.
func TestMasterNameReaderCorruptHashFailsBeforeQuery(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table, func(t *testing.T) {
			bundle := corruptManifestHash(t, candidateFreshBundleForApply(t), table)
			q := &masterReaderProbeQuerier{}
			err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
			if err == nil || !strings.Contains(err.Error(), "sha256") {
				t.Fatalf("error = %v, want source sha256 mismatch", err)
			}
			if table == "procedures" && q.procProbes+q.vaccProbes != 0 {
				t.Fatal("name query ran before hash verification")
			}
			if table == "vaccines" && q.vaccProbes != 0 {
				t.Fatal("vaccines name query ran before its hash verification")
			}
		})
	}
}

// TestMasterNameReaderRowCountMismatchFailsBeforeQuery proves a declared row-count mismatch is rejected before the affected table's target query.
func TestMasterNameReaderRowCountMismatchFailsBeforeQuery(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table, func(t *testing.T) {
			bundle := corruptManifestRowCount(t, candidateFreshBundleForApply(t), table)
			q := &masterReaderProbeQuerier{}
			err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
			if err == nil || !strings.Contains(err.Error(), "row count mismatch") {
				t.Fatalf("error = %v, want source row count mismatch", err)
			}
		})
	}
}

// TestMasterNameReaderMalformedCSVFailsPrivacy proves unparseable CSV returns fixed context with no lexeme or path echo.
func TestMasterNameReaderMalformedCSVFailsPrivacy(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	manifestTable, err := cutoverManifestTableByName(bundle.Manifest, "procedures")
	if err != nil {
		t.Fatal(err)
	}
	path := cutoverCSVPath(bundle.SourceDir, manifestTable)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(raw), "\n", 2)
	broken := lines[0] + "\n\"unterminated-secret-name\n"
	if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}
	q := &masterReaderProbeQuerier{}
	err = validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
	if err == nil || !strings.Contains(err.Error(), "name preflight") {
		t.Fatalf("error = %v, want fixed name preflight failure", err)
	}
	if strings.Contains(err.Error(), "unterminated-secret-name") || strings.Contains(err.Error(), path) {
		t.Fatalf("diagnostic leaks source detail: %v", err)
	}
}

// TestMasterNameReaderWrongHeaderRejected proves a wrong header is rejected without echoing header bytes.
func TestMasterNameReaderWrongHeaderRejected(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	manifestTable, err := cutoverManifestTableByName(bundle.Manifest, "procedures")
	if err != nil {
		t.Fatal(err)
	}
	path := cutoverCSVPath(bundle.SourceDir, manifestTable)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(raw), "\n", 2)
	if err := os.WriteFile(path, []byte("unexpected,header\n"+parts[1]), 0o600); err != nil {
		t.Fatal(err)
	}
	q := &masterReaderProbeQuerier{}
	err = validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
	if err == nil || !strings.Contains(err.Error(), "header") {
		t.Fatalf("error = %v, want header mismatch failure", err)
	}
	if strings.Contains(err.Error(), "unexpected,header") {
		t.Fatalf("diagnostic leaks header bytes: %v", err)
	}
}

// TestMasterNameReaderUppercaseDigestAccepted proves uppercase manifest digests still match the computed lowercase hash.
func TestMasterNameReaderUppercaseDigestAccepted(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	for i := range bundle.Manifest.Tables {
		bundle.Manifest.Tables[i].SHA256 = strings.ToUpper(bundle.Manifest.Tables[i].SHA256)
	}
	q := &masterReaderProbeQuerier{}
	if err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false); err != nil {
		t.Fatalf("uppercase manifest digest rejected: %v", err)
	}
}

// rewriteTableCSV applies a row-level edit to one fixture CSV, rewrites it, and returns the new SHA-256 for manifest updates.
func rewriteTableCSV(t *testing.T, bundle CutoverBundle, table string, edit func(header []string, rows [][]string) [][]string) string {
	t.Helper()
	manifestTable, err := cutoverManifestTableByName(bundle.Manifest, table)
	if err != nil {
		t.Fatal(err)
	}
	path := cutoverCSVPath(bundle.SourceDir, manifestTable)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bufio.NewReader(f)).ReadAll()
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	body := edit(rows[0], rows[1:])
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	if err := w.WriteAll(append([][]string{rows[0]}, body...)); err != nil {
		t.Fatal(err)
	}
	content := b.Bytes()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// updateManifestEntry overwrites one manifest table's digest and row count to match a rewritten CSV.
func updateManifestEntry(t *testing.T, bundle CutoverBundle, table, digest string, rowCount int64) CutoverBundle {
	t.Helper()
	for i := range bundle.Manifest.Tables {
		if bundle.Manifest.Tables[i].Table == table {
			bundle.Manifest.Tables[i].SHA256 = digest
			bundle.Manifest.Tables[i].RowCount = rowCount
		}
	}
	return bundle
}

// rehearsalBundle builds a candidate apply bundle with local-rehearsal provenance so header permutations are tolerated.
func rehearsalBundle(t *testing.T) CutoverBundle {
	t.Helper()
	bundle := candidateFreshBundleForApply(t)
	bundle.Provenance = CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal}
	return bundle
}

// TestMasterNameReaderHeaderPermutation proves a header-column swap is accepted under rehearsal provenance and rejected under formal provenance for both tables.
func TestMasterNameReaderHeaderPermutation(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table+"/rehearsal-allowed", func(t *testing.T) {
			bundle := rehearsalBundle(t)
			digest := rewriteTableCSV(t, bundle, table, func(header []string, rows [][]string) [][]string {
				header[0], header[1] = header[1], header[0]
				for _, row := range rows {
					row[0], row[1] = row[1], row[0]
				}
				return rows
			})
			for i := range bundle.Manifest.Tables {
				if bundle.Manifest.Tables[i].Table == table {
					bundle.Manifest.Tables[i].SHA256 = digest
				}
			}
			q := &masterReaderProbeQuerier{}
			if err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false); err != nil {
				t.Fatalf("rehearsal header permutation rejected: %v", err)
			}
		})
		t.Run(table+"/formal-rejected", func(t *testing.T) {
			bundle := candidateFreshBundleForApply(t)
			bundle.Provenance = CutoverProvenanceContract{Mode: CutoverProvenanceFormal}
			digest := rewriteTableCSV(t, bundle, table, func(header []string, rows [][]string) [][]string {
				header[0], header[1] = header[1], header[0]
				for _, row := range rows {
					row[0], row[1] = row[1], row[0]
				}
				return rows
			})
			for i := range bundle.Manifest.Tables {
				if bundle.Manifest.Tables[i].Table == table {
					bundle.Manifest.Tables[i].SHA256 = digest
				}
			}
			q := &masterReaderProbeQuerier{}
			err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
			if err == nil || !strings.Contains(err.Error(), "header") {
				t.Fatalf("formal header permutation accepted: %v", err)
			}
			if strings.Contains(err.Error(), ",") && strings.Contains(err.Error(), "id,") {
				t.Fatalf("diagnostic leaks header bytes: %v", err)
			}
		})
	}
}

// TestMasterNameReaderChunkingBothTables uses 257-row fixtures to check 256-name batching and exact bound-name order for both master tables.
func TestMasterNameReaderChunkingBothTables(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table, func(t *testing.T) {
			bundle := candidateFreshBundleForApply(t)
			var spec CutoverTableSpec
			for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
				if candidate.Name == table {
					spec = candidate
				}
			}
			nameIdx := columnIndex(spec.Columns, "name")
			wantNames := make([]string, 0, 257)

			digest := rewriteTableCSV(t, bundle, table, func(header []string, rows [][]string) [][]string {
				if len(rows) == 0 {
					t.Fatal("fixture source has no data rows to clone")
				}
				base := append([]string(nil), rows[0]...)
				out := make([][]string, 0, 257)
				for i := 0; i < 257; i++ {
					row := append([]string(nil), base...)
					row[nameIdx] = fmt.Sprintf(" name-%03d ' 名 ", i)
					out = append(out, row)
					wantNames = append(wantNames, row[nameIdx])
				}
				return out
			})
			bundle = updateManifestEntry(t, bundle, table, digest, 257)
			q := &masterReaderProbeQuerier{}
			if err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false); err != nil {
				t.Fatal(err)
			}
			gotChunks := q.chunks
			gotCount := q.procProbes
			gotNames := q.procNamesFlat
			otherCount := q.vaccProbes
			if table == "vaccines" {
				gotChunks = q.vaccChunks
				gotCount = q.vaccProbes
				gotNames = q.vaccNamesFlat
				otherCount = q.procProbes
			}
			if gotCount != 2 {
				t.Fatalf("%s probes = %d, want 2 chunks", table, gotCount)
			}
			if len(gotChunks) != 2 || gotChunks[0] != 256 || gotChunks[1] != 1 {
				t.Fatalf("%s chunk sizes = %v, want [256 1]", table, gotChunks)
			}
			if len(gotNames) != len(wantNames) {
				t.Fatalf("%s flattened names = %d, want %d", table, len(gotNames), len(wantNames))
			}
			for i := range wantNames {
				if gotNames[i] != wantNames[i] {
					t.Fatalf("%s name %d = %q, want %q", table, i, gotNames[i], wantNames[i])
				}
			}
			if otherCount != 1 {
				t.Fatalf("other table probes = %d, want 1", otherCount)
			}
		})
	}
}

// TestMasterNameReaderEmptySelectedTableNoQuery proves a zero-row source skips its own query while the other table is still probed.
func TestMasterNameReaderEmptySelectedTableNoQuery(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table, func(t *testing.T) {
			bundle := candidateFreshBundleForApply(t)
			digest := rewriteTableCSV(t, bundle, table, func(header []string, rows [][]string) [][]string {
				return nil
			})
			rows0 := int64(0)
			bundle = updateManifestEntry(t, bundle, table, digest, rows0)
			q := &masterReaderProbeQuerier{}
			if err := validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false); err != nil {
				t.Fatal(err)
			}
			sel := q.procProbes
			other := q.vaccProbes
			if table == "vaccines" {
				sel, other = q.vaccProbes, q.procProbes
			}
			if sel != 0 {
				t.Fatalf("%s empty file probed %d times, want 0", table, sel)
			}
			if other != 1 {
				t.Fatalf("other table probed %d times, want 1", other)
			}
		})
	}
}

// TestMasterNameReaderMalformedPrivacyBothTables proves malformed rows and wrong headers fail with fixed context for both tables, never leaking cell or path detail.
func TestMasterNameReaderMalformedPrivacyBothTables(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		t.Run(table+"/malformed", func(t *testing.T) {
			bundle := candidateFreshBundleForApply(t)
			manifestTable, err := cutoverManifestTableByName(bundle.Manifest, table)
			if err != nil {
				t.Fatal(err)
			}
			path := cutoverCSVPath(bundle.SourceDir, manifestTable)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.SplitN(string(raw), "\n", 2)
			if err := os.WriteFile(path, []byte(lines[0]+"\n\"unterminated-SECRET-name\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			q := &masterReaderProbeQuerier{}
			err = validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
			if err == nil || !strings.Contains(err.Error(), "name preflight") {
				t.Fatalf("error = %v, want fixed name preflight failure", err)
			}
			if strings.Contains(err.Error(), "unterminated-SECRET-name") || strings.Contains(err.Error(), path) {
				t.Fatalf("diagnostic leaks source detail: %v", err)
			}
			sel := q.procProbes
			if table == "vaccines" {
				sel = q.vaccProbes
			}
			if sel != 0 {
				t.Fatalf("%s was queried before its source integrity passed", table)
			}
		})
		t.Run(table+"/wrong-header", func(t *testing.T) {
			bundle := candidateFreshBundleForApply(t)
			manifestTable, err := cutoverManifestTableByName(bundle.Manifest, table)
			if err != nil {
				t.Fatal(err)
			}
			path := cutoverCSVPath(bundle.SourceDir, manifestTable)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			parts := strings.SplitN(string(raw), "\n", 2)
			if err := os.WriteFile(path, []byte("bogus,header\n"+parts[1]), 0o600); err != nil {
				t.Fatal(err)
			}
			q := &masterReaderProbeQuerier{}
			err = validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
			if err == nil || !strings.Contains(err.Error(), "header") {
				t.Fatalf("error = %v, want header mismatch failure", err)
			}
			if strings.Contains(err.Error(), "bogus,header") {
				t.Fatalf("diagnostic leaks header bytes: %v", err)
			}
		})
	}
}

// TestMasterNameReaderExplicitNamesBoundVerbatim proves mixed-case, whitespace-padded, and quoted Unicode source names reach the bind slice verbatim for both tables.
func TestMasterNameReaderExplicitNamesBoundVerbatim(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	probe := &masterNameProbeSession{}
	wantProc := "  MiXeD procedure ' 名  "
	wantVacc := "  MiXeD vaccine ' 名  "
	for _, tc := range []struct{ table, name string }{{"procedures", wantProc}, {"vaccines", wantVacc}} {
		var spec CutoverTableSpec
		for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
			if candidate.Name == tc.table {
				spec = candidate
			}
		}
		idx := columnIndex(spec.Columns, "name")
		digest := rewriteTableCSV(t, bundle, tc.table, func(header []string, rows [][]string) [][]string {
			for _, row := range rows {
				row[idx] = tc.name
			}
			return rows
		})
		bundle = updateManifestEntry(t, bundle, tc.table, digest, 1)
	}
	if _, err := applyCutoverCommittingEachTable(context.Background(), probe, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	if len(probe.procNames) != 1 || probe.procNames[0] != wantProc {
		t.Fatalf("procedures bound names = %q, want exact %q", probe.procNames, wantProc)
	}
	if len(probe.vaccNames) != 1 || probe.vaccNames[0] != wantVacc {
		t.Fatalf("vaccines bound names = %q, want exact %q", probe.vaccNames, wantVacc)
	}
	seeds := validCutoverSeeds()
	if probe.procClinic != seeds.ClinicID || probe.vaccClinic != seeds.ClinicID {
		t.Fatal("clinic binding mismatch")
	}
	if len(probe.order) < 2 || probe.order[0] != "procedures" || probe.order[1] != "vaccines" {
		t.Fatalf("probe order = %v", probe.order)
	}
}

// TestMasterNameProbePreservesSafeCauses proves probe causes (cancel or driver error) survive errors.Is while Error() keeps the fixed safe text, before Begin.
func TestMasterNameProbePreservesSafeCauses(t *testing.T) {
	for _, table := range []string{"procedures", "vaccines"} {
		for _, cause := range []error{context.Canceled, errors.New("DRIVER-SECRET-/tmp/patient-phi-path")} {
			t.Run(table+"/"+cause.Error()[:10], func(t *testing.T) {
				bundle := candidateFreshBundleForApply(t)
				session := &masterNameProbeSession{probeErr: cause, probeErrTable: table}
				_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
				if err == nil {
					t.Fatal("probe error swallowed")
				}
				if !errors.Is(err, cause) {
					t.Fatalf("error cause not preserved: %v", err)
				}
				if strings.Contains(err.Error(), "DRIVER-SECRET") || strings.Contains(err.Error(), "patient-phi-path") {
					t.Fatalf("diagnostic leaks driver/source detail: %v", err)
				}
				if len(session.txs) != 0 {
					t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
				}
			})
		}
	}
}

// TestMasterNameReaderMalformedCSVClassifiesParseCause proves malformed CSV unwraps to *csv.ParseError while Error() stays sanitized.
func TestMasterNameReaderMalformedCSVClassifiesParseCause(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	manifestTable, err := cutoverManifestTableByName(bundle.Manifest, "procedures")
	if err != nil {
		t.Fatal(err)
	}
	path := cutoverCSVPath(bundle.SourceDir, manifestTable)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(raw), "\n", 2)
	if err := os.WriteFile(path, []byte(lines[0]+"\n\"unterminated-SECRET-name\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	q := &masterReaderProbeQuerier{}
	err = validateCutoverUniqueMasterTargetNames(context.Background(), q, bundle, validCutoverSeeds(), false)
	if err == nil || !strings.Contains(err.Error(), "name preflight") {
		t.Fatalf("error = %v, want fixed name preflight failure", err)
	}
	var parseErr *csv.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("malformed csv cause not classified as *csv.ParseError: %v", err)
	}
	if strings.Contains(err.Error(), "unterminated-SECRET-name") || strings.Contains(err.Error(), path) {
		t.Fatalf("diagnostic leaks source detail: %v", err)
	}
}
