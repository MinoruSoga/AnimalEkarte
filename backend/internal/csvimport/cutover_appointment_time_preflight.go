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
	"math/big"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const cutoverAppointmentEpochQuery = `SELECT COALESCE(array_agg(EXTRACT(EPOCH FROM input.start_text::timestamptz)::text ORDER BY input.ordinal), ARRAY[]::text[]),
       COALESCE(array_agg(EXTRACT(EPOCH FROM input.end_text::timestamptz)::text ORDER BY input.ordinal), ARRAY[]::text[])
FROM unnest($1::text[], $2::text[]) WITH ORDINALITY AS input(start_text, end_text, ordinal)
;`

const cutoverAppointmentCollisionQuery = `SELECT EXISTS (
    SELECT 1
    FROM unnest($2::text[], $3::text[], $4::text[]) AS incoming(doctor_text, start_text, end_text)
    JOIN public.appointments AS target
      ON target.clinic_id = $1
     AND target.doctor_id = incoming.doctor_text::bigint
     AND target.deleted_at IS NULL
     AND target.status <> 'cancelled'
    WHERE target.start_time = incoming.start_text::timestamptz
       OR tstzrange(target.start_time, target.end_time, '[)') &&
          tstzrange(incoming.start_text::timestamptz, incoming.end_text::timestamptz, '[)')
)
;`

var cutoverCanonicalDecimal = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)

// cutoverAppointmentInstant is a canonical instant: a finite decimal or an infinity sentinel.
type cutoverAppointmentInstant struct {
	finite   *big.Rat
	infinity int8
}

// parseCutoverAppointmentInstant reads PostgreSQL epoch decimals exactly, without imposing a Go time range.
func parseCutoverAppointmentInstant(text string) (cutoverAppointmentInstant, error) {
	lower := strings.ToLower(text)
	switch lower {
	case "infinity", "+infinity":
		return cutoverAppointmentInstant{infinity: 1}, nil
	case "-infinity":
		return cutoverAppointmentInstant{infinity: -1}, nil
	}
	if !cutoverCanonicalDecimal.MatchString(text) {
		return cutoverAppointmentInstant{}, fmt.Errorf("non-canonical epoch literal")
	}
	r, ok := new(big.Rat).SetString(text)
	if !ok {
		return cutoverAppointmentInstant{}, fmt.Errorf("not a canonical epoch decimal")
	}
	return cutoverAppointmentInstant{finite: r}, nil
}

// cmp orders finite epochs and both infinity sentinels without floating-point conversion.
func (a cutoverAppointmentInstant) cmp(b cutoverAppointmentInstant) int {
	if a.infinity != 0 || b.infinity != 0 {
		return int(a.infinity - b.infinity)
	}
	return a.finite.Cmp(b.finite)
}

// identity provides a comparable key for equivalent decimal spellings and explicit infinities.
func (a cutoverAppointmentInstant) identity() string {
	if a.infinity > 0 {
		return "Infinity"
	}
	if a.infinity < 0 {
		return "-Infinity"
	}
	return a.finite.RatString()
}

type cutoverAppointmentRow struct {
	line      int64
	doctor    string
	startText string
	endText   string
	status    string
	start     cutoverAppointmentInstant
	end       cutoverAppointmentInstant
	eligible  bool
}

// validateCutoverAppointmentSourceTimes returns all canonical rows in CSV order after source-only checks; it does not query target appointments.
func validateCutoverAppointmentSourceTimes(ctx context.Context, q cutoverQuerier, bundle CutoverBundle) ([]cutoverAppointmentRow, error) {
	const context0 = "table appointments temporal preflight"
	table, err := cutoverManifestTableByName(bundle.Manifest, "appointments")
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": manifest entry is missing", cause: err}
	}
	var spec CutoverTableSpec
	for _, c := range CutoverTableSpecs(bundle.Manifest.profile) {
		if c.Name == "appointments" {
			spec = c
		}
	}
	resolved, err := resolveCutoverAccountCSV(cutoverCSVPath(bundle.SourceDir, table))
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": source csv path is not resolvable", cause: err}
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": source csv is not inspectable", cause: err}
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: csv must be a regular file", context0)
	}
	f, err := openStableOwnerOnlyFile(resolved)
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": source csv is not readable", cause: err}
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	limited := &io.LimitedReader{R: f, N: maxCutoverCSVBytes + 1}
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(limited, hash)))
	if err := ctx.Err(); err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
	}
	header, err := reader.Read()
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": source csv header is not readable", cause: err}
	}
	indexes, _, err := cutoverColumnIndexes(spec, header, rehearsalContractDrift(bundle.Manifest, bundle.Provenance.Mode))
	if err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": source csv header does not match the expected columns", cause: err}
	}
	var rows []cutoverAppointmentRow
	rowCount := int64(0)
	for {
		if err := ctx.Err(); err != nil {
			return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
		}
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, &cutoverNamePreflightError{message: fmt.Sprintf("%s: source csv row %d is not parseable", context0, rowCount+2), cause: readErr}
		}
		rowCount++
		dr := row[indexes["doctor_id"]]
		doctor := ""
		if dr != "" {
			n, perr := strconv.ParseInt(dr, 10, 64)
			if perr != nil {
				return nil, &cutoverNamePreflightError{message: fmt.Sprintf("table appointments column doctor_id row %d: invalid staff id", rowCount+1), cause: perr}
			}
			doctor = strconv.FormatInt(n, 10)
		}
		rows = append(rows, cutoverAppointmentRow{line: rowCount + 1, doctor: doctor, status: row[indexes["status"]], startText: row[indexes["start_time"]], endText: row[indexes["end_time"]]})
	}
	if limited.N <= 0 {
		return nil, fmt.Errorf("%s: source csv exceeds size limit", context0)
	}
	if rowCount != table.RowCount {
		return nil, fmt.Errorf("%s: source row count mismatch: got %d, want %d", context0, rowCount, table.RowCount)
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), table.SHA256) {
		return nil, fmt.Errorf("%s: source csv sha256 does not match manifest", context0)
	}
	starts := make([]string, len(rows))
	ends := make([]string, len(rows))
	for i, r := range rows {
		starts[i], ends[i] = r.startText, r.endText
	}
	var scanStart, scanEnd []string
	for s := 0; s < len(rows); s += cutoverMedicineNameBatchSize {
		if err := ctx.Err(); err != nil {
			return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
		}
		e := s + cutoverMedicineNameBatchSize
		if e > len(rows) {
			e = len(rows)
		}
		var a, b []string
		if err := q.QueryRow(ctx, cutoverAppointmentEpochQuery, starts[s:e], ends[s:e]).Scan(&a, &b); err != nil {
			return nil, &cutoverNamePreflightError{message: context0 + ": canonicalization failed", cause: err}
		}
		if len(a) != e-s || len(b) != e-s {
			return nil, fmt.Errorf("%s: canonicalization returned wrong shape", context0)
		}
		scanStart, scanEnd = append(scanStart, a...), append(scanEnd, b...)
	}
	for i := range rows {
		if err := ctx.Err(); err != nil {
			return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
		}
		var err error
		if rows[i].start, err = parseCutoverAppointmentInstant(scanStart[i]); err != nil {
			return nil, &cutoverNamePreflightError{message: fmt.Sprintf("table appointments column start_time row %d: non-canonical instant", rows[i].line), cause: err}
		}
		if rows[i].end, err = parseCutoverAppointmentInstant(scanEnd[i]); err != nil {
			return nil, &cutoverNamePreflightError{message: fmt.Sprintf("table appointments column end_time row %d: non-canonical instant", rows[i].line), cause: err}
		}
		if rows[i].end.cmp(rows[i].start) < 0 {
			return nil, fmt.Errorf("table appointments column end_time row %d: end precedes start", rows[i].line)
		}
		rows[i].eligible = rows[i].status != "" && rows[i].status != "cancelled" && rows[i].doctor != ""
	}
	seen := map[string]map[string]int64{}
	for _, r := range rows {
		if !r.eligible {
			continue
		}
		m := seen[r.doctor]
		if m == nil {
			m = map[string]int64{}
			seen[r.doctor] = m
		}
		if prev, dup := m[r.start.identity()]; dup {
			return nil, fmt.Errorf("table appointments column start_time row %d: duplicates row %d for the same doctor", r.line, prev)
		}
		m[r.start.identity()] = r.line
	}
	eligible := make([]cutoverAppointmentRow, 0, len(rows))
	for _, r := range rows {
		if r.eligible {
			eligible = append(eligible, r)
		}
	}
	groups := map[string][]cutoverAppointmentRow{}
	for _, r := range eligible {
		groups[r.doctor] = append(groups[r.doctor], r)
	}
	for _, rs := range groups {
		if err := ctx.Err(); err != nil {
			return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
		}
		sort.Slice(rs, func(i, j int) bool {
			if c := rs[i].start.cmp(rs[j].start); c != 0 {
				return c < 0
			}
			if c := rs[i].end.cmp(rs[j].end); c != 0 {
				return c < 0
			}
			return rs[i].line < rs[j].line
		})
		var maxEnd *cutoverAppointmentInstant
		for _, r := range rs {
			if maxEnd != nil && r.end.cmp(r.start) > 0 && r.start.cmp(*maxEnd) < 0 {
				return nil, fmt.Errorf("table appointments column start_time row %d: overlaps a prior appointment for the same doctor", r.line)
			}
			if r.end.cmp(r.start) > 0 && (maxEnd == nil || r.end.cmp(*maxEnd) > 0) {
				e := r.end
				maxEnd = &e
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
	}
	return rows, nil
}

// validateCutoverAppointmentTimes checks appointment storage constraints using PostgreSQL parsing in the
// locked COPY transaction. It permits historical empty ranges and preserves original timestamp cells.
func validateCutoverAppointmentTimes(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs) error {
	const context0 = "table appointments temporal preflight"
	rows, err := validateCutoverAppointmentSourceTimes(ctx, q, bundle)
	if err != nil {
		return err
	}
	eligible := make([]cutoverAppointmentRow, 0, len(rows))
	for _, r := range rows {
		if r.eligible {
			eligible = append(eligible, r)
		}
	}
	if err := ctx.Err(); err != nil {
		return &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
	}
	for s := 0; s < len(eligible); s += cutoverMedicineNameBatchSize {
		if err := ctx.Err(); err != nil {
			return &cutoverNamePreflightError{message: context0 + ": cancelled", cause: err}
		}
		e := s + cutoverMedicineNameBatchSize
		if e > len(eligible) {
			e = len(eligible)
		}
		ds := make([]string, e-s)
		ss := make([]string, e-s)
		es := make([]string, e-s)
		for i := s; i < e; i++ {
			ds[i-s], ss[i-s], es[i-s] = eligible[i].doctor, eligible[i].startText, eligible[i].endText
		}
		var exists bool
		if err := q.QueryRow(ctx, cutoverAppointmentCollisionQuery, seeds.ClinicID, ds, ss, es).Scan(&exists); err != nil {
			return &cutoverNamePreflightError{message: context0 + ": target check failed", cause: err}
		}
		if exists {
			return fmt.Errorf("%s: %s already exists in the target clinic", CutoverRefBandOccupied, context0)
		}
	}
	return nil
}
