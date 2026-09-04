package csvimport

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sys/unix"
)

func validateCutoverColumnTypes(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		bandColumns := make(map[string]struct{}, len(spec.BandColumns))
		for _, column := range spec.BandColumns {
			bandColumns[column] = struct{}{}
		}
		for _, column := range spec.Columns {
			var dataType string
			err := q.QueryRow(ctx, `SELECT data_type FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, spec.Name, column).Scan(&dataType)
			if err != nil {
				return fmt.Errorf("target schema is missing required column public.%s.%s", spec.Name, column)
			}
			if _, isBandColumn := bandColumns[column]; isBandColumn && dataType != "bigint" {
				return fmt.Errorf("target schema requires public.%s.%s to exist as bigint", spec.Name, column)
			}
		}
	}
	return nil
}

func validateCutoverSequences(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := q.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil || *sequence == "" {
			return fmt.Errorf("target table %s must have a serial id sequence", spec.Name)
		}
	}
	return nil
}

func validateCutoverBandEmpty(ctx context.Context, q cutoverQuerier, band CutoverIDBand) error {
	for _, spec := range CutoverTableSpecs() {
		occupied, err := cutoverBandOccupied(ctx, q, spec, band)
		if err != nil {
			return err
		}
		if occupied {
			// Non-PHI: table name + stable ref only. Rerun/idempotency guard (Issue #250).
			return fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
		}
	}
	return nil
}

func cutoverBandFloor(spec CutoverTableSpec, band CutoverIDBand) int64 {
	if spec.Name == "owners" {
		return band.OwnerFloor
	}
	return band.NonOwnerIDOffset
}

func cutoverBandOccupied(ctx context.Context, q cutoverQuerier, spec CutoverTableSpec, band CutoverIDBand) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id >= $1 AND id < $2)`, pgx.Identifier{spec.Name}.Sanitize())
	var occupied bool
	if err := q.QueryRow(ctx, query, cutoverBandFloor(spec, band), band.EndExclusive).Scan(&occupied); err != nil {
		return false, fmt.Errorf("inspect target band for table %s: %w", spec.Name, err)
	}
	return occupied, nil
}

func cutoverBandCount(ctx context.Context, q cutoverQuerier, spec CutoverTableSpec, band CutoverIDBand) (int64, error) {
	query := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2`, pgx.Identifier{spec.Name}.Sanitize())
	var count int64
	if err := q.QueryRow(ctx, query, cutoverBandFloor(spec, band), band.EndExclusive).Scan(&count); err != nil {
		return 0, fmt.Errorf("verify table %s row count: %w", spec.Name, err)
	}
	return count, nil
}

func lockCutoverTable(ctx context.Context, tx cutoverTransaction, table string) error {
	if _, err := tx.Exec(ctx, "LOCK TABLE "+pgx.Identifier{table}.Sanitize()+" IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("lock cutover table %s: %w", table, err)
	}
	return nil
}

func lockCutoverTables(ctx context.Context, tx cutoverTransaction) error {
	identifiers := make([]string, 0, len(CutoverTableSpecs()))
	for _, spec := range CutoverTableSpecs() {
		identifiers = append(identifiers, pgx.Identifier{spec.Name}.Sanitize())
	}
	if _, err := tx.Exec(ctx, "LOCK TABLE "+strings.Join(identifiers, ", ")+" IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("lock cutover tables: %w", err)
	}
	return nil
}

type cutoverTransformResult struct {
	count int64
	err   error
}

func runCutoverCSVTransform(
	ctx context.Context,
	writer *io.PipeWriter,
	result chan<- cutoverTransformResult,
	path string,
	seeds CutoverSeedIDs,
	expectedSHA256 string,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err := fmt.Errorf("cutover csv transform panicked: %v", recovered)
			_ = writer.CloseWithError(err)
			result <- cutoverTransformResult{err: err}
		}
	}()
	count, err := transformCutoverCSV(ctx, path, writer, seeds, expectedSHA256)
	_ = writer.CloseWithError(err)
	result <- cutoverTransformResult{count: count, err: err}
}

// PostgreSQL rejects COPY FROM on a table with row-level security unless the
// session role is superuser, has BYPASSRLS, or owns the table without FORCE RLS
// (SQLSTATE 0A000). app.bypass_rls only affects has_clinic_access() policies; it
// does not satisfy COPY's RLS check. Staging COPY into a TEMP table (no RLS)
// then INSERT SELECT into the target keeps the GUC WITH CHECK path.
const cutoverCopyFromBlockedByRLSQuery = `SELECT c.relrowsecurity
   AND NOT EXISTS (
     SELECT 1
     FROM pg_catalog.pg_roles AS role
     WHERE role.rolname = current_user
       AND (role.rolbypassrls OR role.rolsuper)
   )
   AND (
     pg_catalog.pg_get_userbyid(c.relowner) IS DISTINCT FROM current_user
     OR c.relforcerowsecurity
   )
FROM pg_catalog.pg_class AS c
JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relname = $1`

// cutoverStagingInsertBatchRows bounds each RLS-path INSERT. A single INSERT
// SELECT of a multi-million-row TEMP table terminated the PlanetScale backend
// (SQLSTATE 57P01). COPY itself streams; the following INSERT must also
// complete in bounded statements.
const cutoverStagingInsertBatchRows int64 = 10_000

func copyFromBlockedByRLS(ctx context.Context, q cutoverQuerier, table string) (bool, error) {
	var blocked bool
	if err := q.QueryRow(ctx, cutoverCopyFromBlockedByRLSQuery, table).Scan(&blocked); err != nil {
		return false, fmt.Errorf("inspect row-level security for table %s", table)
	}
	return blocked, nil
}

func cutoverCopyStagingRelation(table string) string {
	return "cutover_copy_" + table
}

func cutoverQuotedColumns(spec CutoverTableSpec) []string {
	columns := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		columns[i] = pgx.Identifier{column}.Sanitize()
	}
	return columns
}

func cutoverCopySQL(spec CutoverTableSpec, dest pgx.Identifier) string {
	// Empty CSV fields become SQL NULL, except ForceNotNullColumns which keep
	// empty strings (required text columns with no default in some paths).
	options := []string{"FORMAT csv", "HEADER true", "NULL ''"}
	if len(spec.ForceNotNullColumns) > 0 {
		forceNotNullColumns := make([]string, len(spec.ForceNotNullColumns))
		for i, column := range spec.ForceNotNullColumns {
			forceNotNullColumns[i] = pgx.Identifier{column}.Sanitize()
		}
		options = append(options, "FORCE_NOT_NULL ("+strings.Join(forceNotNullColumns, ", ")+")")
	}
	return "COPY " + dest.Sanitize() + " (" + strings.Join(cutoverQuotedColumns(spec), ", ") + ") FROM STDIN WITH (" + strings.Join(options, ", ") + ")"
}

func cutoverCopyStagingCreateSQL(table string) string {
	return "CREATE TEMP TABLE " + pgx.Identifier{cutoverCopyStagingRelation(table)}.Sanitize() +
		" (LIKE " + pgx.Identifier{"public", table}.Sanitize() + " INCLUDING DEFAULTS)"
}

func cutoverStagingInsertSQL(spec CutoverTableSpec) string {
	columns := strings.Join(cutoverQuotedColumns(spec), ", ")
	staging := pgx.Identifier{cutoverCopyStagingRelation(spec.Name)}.Sanitize()
	dest := pgx.Identifier{spec.Name}.Sanitize()
	returning := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		returning[i] = "staged." + pgx.Identifier{column}.Sanitize()
	}
	// ctid LIMIT batches keep each INSERT bounded. A single INSERT SELECT of
	// multi-million-row TEMP data (billing_items / exam_results) terminated the
	// PlanetScale backend with SQLSTATE 57P01. Deleting each batch from TEMP
	// also releases staging disk before commit.
	return "WITH pick AS MATERIALIZED (SELECT ctid FROM " + staging + " LIMIT $1), moved AS (" +
		"DELETE FROM " + staging + " AS staged USING pick WHERE staged.ctid = pick.ctid RETURNING " +
		strings.Join(returning, ", ") + ") INSERT INTO " + dest + " (" + columns +
		") OVERRIDING SYSTEM VALUE SELECT " + columns + " FROM moved"
}

func cutoverCopyStagingDropSQL(table string) string {
	return "DROP TABLE " + pgx.Identifier{cutoverCopyStagingRelation(table)}.Sanitize()
}

func insertCutoverStaging(ctx context.Context, tx cutoverTransaction, spec CutoverTableSpec, expected int64) (int64, error) {
	insertSQL := cutoverStagingInsertSQL(spec)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		tag, err := tx.Exec(ctx, insertSQL, cutoverStagingInsertBatchRows)
		if insertErr := cutoverStagingInsertError(spec.Name, err); insertErr != nil {
			return 0, insertErr
		}
		n := tag.RowsAffected()
		if n < 0 || (n > 0 && total > math.MaxInt64-n) {
			return 0, fmt.Errorf("table %s: imported row count does not match manifest", spec.Name)
		}
		total += n
		if n < cutoverStagingInsertBatchRows {
			break
		}
	}
	if total != expected {
		return 0, fmt.Errorf("table %s: imported row count does not match manifest", spec.Name)
	}
	return total, nil
}

func copyCutoverTable(ctx context.Context, tx cutoverTransaction, path string, spec CutoverTableSpec, table CutoverManifestTable, seeds CutoverSeedIDs) (int64, error) {
	blocked, err := copyFromBlockedByRLS(ctx, tx, spec.Name)
	if err != nil {
		return 0, err
	}
	dest := pgx.Identifier{spec.Name}
	if blocked {
		dest = pgx.Identifier{cutoverCopyStagingRelation(spec.Name)}
		if _, err := tx.Exec(ctx, cutoverCopyStagingCreateSQL(spec.Name)); err != nil {
			return 0, fmt.Errorf("table %s: create COPY staging table failed", spec.Name)
		}
	}

	copyCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, writer := io.Pipe()
	result := make(chan cutoverTransformResult, 1)
	go runCutoverCSVTransform(copyCtx, writer, result, path, seeds, table.SHA256)

	copySQL := cutoverCopySQL(spec, dest)
	tag, copyErr := tx.CopyFrom(copyCtx, reader, copySQL)
	if copyErr != nil {
		cancel()
		// Never forward the PostgreSQL COPY error through the pipe: it may echo
		// a source value and the transformer error is otherwise returned to the
		// caller. Closing without that error only unblocks the writer.
		_ = reader.Close()
	}
	transformed := <-result
	if err := cutoverCopyResultError(spec.Name, copyErr, transformed.err); err != nil {
		return 0, err
	}
	if tag.RowsAffected() != table.RowCount || transformed.count != table.RowCount {
		return 0, fmt.Errorf("table %s: imported row count does not match manifest", spec.Name)
	}
	if blocked {
		if _, err := insertCutoverStaging(ctx, tx, spec, table.RowCount); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, cutoverCopyStagingDropSQL(spec.Name)); err != nil {
			return 0, fmt.Errorf("table %s: drop COPY staging table failed", spec.Name)
		}
	}
	return tag.RowsAffected(), nil
}

func cutoverCopyResultError(table string, copyErr, transformErr error) error {
	if copyErr != nil {
		return cutoverSanitizedWriteError(table, "COPY", copyErr)
	}
	if transformErr != nil {
		return fmt.Errorf("table %s: source changed or transformation failed: %w", table, transformErr)
	}
	return nil
}

func cutoverStagingInsertError(table string, insertErr error) error {
	if insertErr == nil {
		return nil
	}
	return cutoverSanitizedWriteError(table, "INSERT", insertErr)
}

func cutoverSanitizedWriteError(table, action string, writeErr error) error {
	// COPY/INSERT errors may echo a source value. Never interpolate
	// driver/DETAIL/HINT text; operators get a constant plus SQLSTATE.
	type pgCoder interface{ SQLState() string }
	msg := "target database rejected the CSV " + action
	var coder pgCoder
	if errors.As(writeErr, &coder) {
		return fmt.Errorf("table %s: %s (SQLSTATE %s)", table, msg, coder.SQLState())
	}
	if cutoverConnectionClosed(writeErr) {
		return fmt.Errorf("table %s: target database connection closed during CSV %s", table, action)
	}
	return fmt.Errorf("table %s: %s", table, msg)
}

func cutoverConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var op *net.OpError
	if errors.As(err, &op) {
		return true
	}
	lower := strings.ToLower(err.Error())
	for _, needle := range []string{
		"conn closed", "connection reset", "unexpected eof", "broken pipe",
		"i/o timeout", "use of closed", "connection refused",
	} {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func transformCutoverCSV(ctx context.Context, path string, output io.Writer, seeds CutoverSeedIDs, expectedSHA256 string) (int64, error) {
	file, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || info.Size() > maxCutoverCSVBytes {
		return 0, fmt.Errorf("source CSV exceeds the size limit")
	}
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(file, maxCutoverCSVBytes+1), hash)))
	writer := csv.NewWriter(output)
	replacements := map[string]string{
		"{{CLINIC_ID}}":                     strconv.FormatInt(seeds.ClinicID, 10),
		"{{FALLBACK_ANIMAL_SPECIES_ID}}":    strconv.FormatInt(seeds.AnimalSpeciesID, 10),
		"{{FALLBACK_EXAM_TYPE_ID}}":         strconv.FormatInt(seeds.ExamTypeID, 10),
		"{{TRIMMING_RESERVATION_TYPE_ID}}":  strconv.FormatInt(seeds.TrimmingReservationTypeID, 10),
		"{{PAYMENT_METHOD_CASH_ID}}":        strconv.FormatInt(seeds.CashPaymentMethodID, 10),
		"{{PAYMENT_METHOD_CREDIT_CARD_ID}}": strconv.FormatInt(seeds.CreditCardPaymentMethodID, 10),
	}
	var records int64
	firstRecord := true
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return 0, fmt.Errorf("read CSV record %d", records+1)
		}
		mapped := append([]string(nil), record...)
		for i, value := range mapped {
			if replacement, ok := replacements[value]; ok {
				mapped[i] = replacement
			}
		}
		if err := writer.Write(mapped); err != nil {
			return 0, fmt.Errorf("write transformed CSV record %d: %w", records+1, err)
		}
		if firstRecord {
			firstRecord = false
		} else {
			records++
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return 0, fmt.Errorf("flush transformed CSV: %w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expectedSHA256) {
		return 0, fmt.Errorf("CSV sha256 changed after preflight")
	}
	return records, nil
}

func openStableOwnerOnlyFile(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect source CSV: %w", err)
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("source CSV must be a regular file and not a symbolic link")
	}
	if err := requireOwnerOnly(before, false); err != nil {
		return nil, err
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0) //nolint:gosec // operator path; O_NOFOLLOW enforces the source boundary
	if err != nil {
		return nil, fmt.Errorf("open source CSV: %w", err)
	}
	file := os.NewFile(uintptr(fd), filepath.Base(path))
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("open source CSV: invalid file descriptor")
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		_ = file.Close()
		return nil, fmt.Errorf("source CSV changed while it was opened")
	}
	if !after.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("source CSV must be a regular file")
	}
	if err := requireOwnerOnly(after, false); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func verifyCutoverRows(ctx context.Context, q cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs, provenance CutoverProvenanceContract) error {
	for i, spec := range CutoverTableSpecs() {
		if err := verifyCutoverTable(ctx, q, spec, manifest.Tables[i], manifest.IDBand, seeds); err != nil {
			return err
		}
	}
	return verifyCutoverPaymentGraph(ctx, q, &manifest, seeds, provenance)
}

func verifyCutoverTable(ctx context.Context, q cutoverQuerier, spec CutoverTableSpec, table CutoverManifestTable, band CutoverIDBand, seeds CutoverSeedIDs) error {
	count, err := cutoverBandCount(ctx, q, spec, band)
	if err != nil {
		return err
	}
	if count != table.RowCount {
		return fmt.Errorf("%s: table %s: committed row count does not match manifest", CutoverRefRowCount, spec.Name)
	}
	if !hasColumn(spec.Columns, "clinic_id") {
		return nil
	}
	clinicQuery := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2 AND clinic_id <> $3`, pgx.Identifier{spec.Name}.Sanitize())
	var mismatched int64
	if err := q.QueryRow(ctx, clinicQuery, cutoverBandFloor(spec, band), band.EndExclusive, seeds.ClinicID).Scan(&mismatched); err != nil {
		return fmt.Errorf("%s: verify table %s clinic isolation: %w", CutoverRefClinicIsolation, spec.Name, err)
	}
	if mismatched != 0 {
		return fmt.Errorf("%s: table %s contains rows assigned to another clinic", CutoverRefClinicIsolation, spec.Name)
	}
	return nil
}

func hasColumn(columns []string, target string) bool {
	for _, column := range columns {
		if column == target {
			return true
		}
	}
	return false
}

func advanceCutoverSequences(ctx context.Context, tx cutoverTransaction) error {
	// PostgreSQL sequence changes are non-transactional. This function only
	// advances values and runs after all row/count checks, so the only possible
	// rollback residue is a harmless jump to the reserved application range.
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := tx.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil {
			return fmt.Errorf("resolve sequence for table %s", spec.Name)
		}
		var current int64
		sequenceSQL := pgx.Identifier(strings.Split(*sequence, ".")).Sanitize()
		if err := tx.QueryRow(ctx, "SELECT last_value FROM "+sequenceSQL).Scan(&current); err != nil {
			return fmt.Errorf("read sequence for table %s: %w", spec.Name, err)
		}
		var maxID int64
		maxQuery := fmt.Sprintf(`SELECT COALESCE(max(id), 0) FROM %s`, pgx.Identifier{spec.Name}.Sanitize())
		if err := tx.QueryRow(ctx, maxQuery).Scan(&maxID); err != nil {
			return fmt.Errorf("read max id for table %s: %w", spec.Name, err)
		}
		lastValue := max(current, max(maxID, applicationIDFloor-1))
		if _, err := tx.Exec(ctx, `SELECT setval($1::regclass, $2, true)`, *sequence, lastValue); err != nil {
			return fmt.Errorf("advance sequence for table %s: %w", spec.Name, err)
		}
	}
	return nil
}

func verifyCutoverSequences(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := q.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil {
			return fmt.Errorf("resolve sequence for table %s", spec.Name)
		}
		sequenceSQL := pgx.Identifier(strings.Split(*sequence, ".")).Sanitize()
		var lastValue int64
		var isCalled bool
		if err := q.QueryRow(ctx, "SELECT last_value, is_called FROM "+sequenceSQL).Scan(&lastValue, &isCalled); err != nil {
			return fmt.Errorf("verify sequence for table %s: %w", spec.Name, err)
		}
		nextValue := lastValue
		if isCalled {
			if lastValue == math.MaxInt64 {
				return fmt.Errorf("table %s sequence exhausted", spec.Name)
			}
			nextValue++
		}
		if nextValue < applicationIDFloor {
			return fmt.Errorf("table %s sequence next value is below application floor", spec.Name)
		}
		var maxID int64
		maxQuery := fmt.Sprintf(`SELECT COALESCE(max(id), 0) FROM %s`, pgx.Identifier{spec.Name}.Sanitize())
		if err := q.QueryRow(ctx, maxQuery).Scan(&maxID); err != nil {
			return fmt.Errorf("read max id for table %s during sequence verification: %w", spec.Name, err)
		}
		if nextValue <= maxID {
			return fmt.Errorf("table %s sequence next value does not exceed the current max id", spec.Name)
		}
	}
	return nil
}
