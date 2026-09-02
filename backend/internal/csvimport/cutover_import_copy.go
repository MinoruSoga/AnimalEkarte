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
		floor := band.NonOwnerIDOffset
		if spec.Name == "owners" {
			floor = band.OwnerFloor
		}
		query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id >= $1 AND id < $2)`, pgx.Identifier{spec.Name}.Sanitize())
		var occupied bool
		if err := q.QueryRow(ctx, query, floor, band.EndExclusive).Scan(&occupied); err != nil {
			return fmt.Errorf("inspect target band for table %s: %w", spec.Name, err)
		}
		if occupied {
			// Non-PHI: table name + stable ref only. Rerun/idempotency guard (Issue #250).
			return fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
		}
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

func copyCutoverTable(ctx context.Context, tx cutoverTransaction, path string, spec CutoverTableSpec, table CutoverManifestTable, seeds CutoverSeedIDs) (int64, error) {
	copyCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, writer := io.Pipe()
	result := make(chan cutoverTransformResult, 1)
	go runCutoverCSVTransform(copyCtx, writer, result, path, seeds, table.SHA256)

	columns := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		columns[i] = pgx.Identifier{column}.Sanitize()
	}
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
	copySQL := "COPY " + pgx.Identifier{spec.Name}.Sanitize() + " (" + strings.Join(columns, ", ") + ") FROM STDIN WITH (" + strings.Join(options, ", ") + ")"
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
	return tag.RowsAffected(), nil
}

func cutoverCopyResultError(table string, copyErr, transformErr error) error {
	if copyErr != nil {
		// COPY errors may echo a source value. Never interpolate driver/DETAIL/HINT
		// text; operators get a constant plus SQLSTATE from coder.SQLState().
		type pgCoder interface{ SQLState() string }
		msg := "target database rejected the CSV COPY"
		var coder pgCoder
		if errors.As(copyErr, &coder) {
			return fmt.Errorf("table %s: %s (SQLSTATE %s)", table, msg, coder.SQLState())
		}
		return fmt.Errorf("table %s: %s", table, msg)
	}
	if transformErr != nil {
		return fmt.Errorf("table %s: source changed or transformation failed: %w", table, transformErr)
	}
	return nil
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

func verifyCutoverRows(ctx context.Context, q cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	for i, spec := range CutoverTableSpecs() {
		floor := manifest.IDBand.NonOwnerIDOffset
		if spec.Name == "owners" {
			floor = manifest.IDBand.OwnerFloor
		}
		query := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2`, pgx.Identifier{spec.Name}.Sanitize())
		var count int64
		if err := q.QueryRow(ctx, query, floor, manifest.IDBand.EndExclusive).Scan(&count); err != nil {
			return fmt.Errorf("verify table %s row count: %w", spec.Name, err)
		}
		if count != manifest.Tables[i].RowCount {
			// Non-PHI: table name + stable ref only (no row payloads).
			return fmt.Errorf("%s: table %s: committed row count does not match manifest", CutoverRefRowCount, spec.Name)
		}
		if hasColumn(spec.Columns, "clinic_id") {
			clinicQuery := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2 AND clinic_id <> $3`, pgx.Identifier{spec.Name}.Sanitize())
			var mismatched int64
			if err := q.QueryRow(ctx, clinicQuery, floor, manifest.IDBand.EndExclusive, seeds.ClinicID).Scan(&mismatched); err != nil {
				return fmt.Errorf("%s: verify table %s clinic isolation: %w", CutoverRefClinicIsolation, spec.Name, err)
			}
			if mismatched != 0 {
				// Non-PHI: table name + stable ref only. Never echo foreign clinic IDs or row values.
				return fmt.Errorf("%s: table %s contains rows assigned to another clinic", CutoverRefClinicIsolation, spec.Name)
			}
		}
	}
	return verifyCutoverPaymentGraph(ctx, q, &manifest, seeds)
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
