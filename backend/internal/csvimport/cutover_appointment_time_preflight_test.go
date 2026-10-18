package csvimport

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const appointmentTimeExpectedEpochQuery = `SELECT COALESCE(array_agg(EXTRACT(EPOCH FROM input.start_text::timestamptz)::text ORDER BY input.ordinal), ARRAY[]::text[]),
       COALESCE(array_agg(EXTRACT(EPOCH FROM input.end_text::timestamptz)::text ORDER BY input.ordinal), ARRAY[]::text[])
FROM unnest($1::text[], $2::text[]) WITH ORDINALITY AS input(start_text, end_text, ordinal)
;`

const appointmentTimeExpectedCollisionQuery = `SELECT EXISTS (
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

// appointmentTimeFakeTx is a cutoverTransaction wrapper that answers the two
// root temporal queries inside the transaction from the configured maps.
type appointmentTimeFakeTx struct {
	*fakeCutoverTransaction
	catalog  cutoverQuerier
	epochErr bool
	collide  bool
	epochs   map[string]string
	probes   []string
	calls    []appointmentTimeCall
	events   []string
	queryFn  func(string, ...any) pgx.Row
	occupied bool
}

type appointmentTimeCall struct {
	query string
	args  []any
}

// QueryRow records exact temporal SQL and raw bindings; epoch answers are explicit fake data, not a PG parser.
func (tx *appointmentTimeFakeTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if query == appointmentTimeExpectedEpochQuery || query == appointmentTimeExpectedCollisionQuery {
		tx.calls = append(tx.calls, appointmentTimeCall{query, args})
		event := "epoch"
		if query == appointmentTimeExpectedCollisionQuery {
			event = "collision"
		}
		tx.probes = append(tx.probes, event)
		tx.events = append(tx.events, event)
		if tx.queryFn != nil {
			return tx.queryFn(query, args...)
		}
	}
	if query == appointmentTimeExpectedEpochQuery {
		if tx.epochErr {
			return staticRow{err: context.Canceled}
		}
		starts := args[0].([]string)
		ends := args[1].([]string)
		sa := make([]string, len(starts))
		ea := make([]string, len(ends))
		for i, v := range starts {
			ep, ok := tx.epochs[v]
			if !ok {
				return staticRow{err: fmt.Errorf("no canonical answer for %q", v)}
			}
			sa[i] = ep
		}
		for i, v := range ends {
			ep, ok := tx.epochs[v]
			if !ok {
				return staticRow{err: fmt.Errorf("no canonical answer for %q", v)}
			}
			ea[i] = ep
		}
		return staticRow{values: []any{sa, ea}}
	}
	if query == appointmentTimeExpectedCollisionQuery {
		return staticRow{values: []any{tx.collide}}
	}
	if tx.occupied && strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "id >= $1 AND id < $2") {
		return staticRow{values: []any{true}}
	}
	if tx.catalog != nil {
		return tx.catalog.QueryRow(ctx, query, args...)
	}
	return tx.fakeCutoverTransaction.QueryRow(ctx, query, args...)
}

// Exec records actual table locks while preserving the candidate catalog fake's atomic-band behavior.
func (tx *appointmentTimeFakeTx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(query, "LOCK TABLE ") {
		tx.events = append(tx.events, "lock")
	}
	if candidate, ok := tx.catalog.(*candidateAtomicTransaction); ok {
		return candidate.Exec(ctx, query, args...)
	}
	return tx.fakeCutoverTransaction.Exec(ctx, query, args...)
}

// CopyFrom records COPY on the same transaction that handled temporal queries.
func (tx *appointmentTimeFakeTx) CopyFrom(ctx context.Context, reader io.Reader, query string) (pgconn.CommandTag, error) {
	tx.events = append(tx.events, "copy")
	if candidate, ok := tx.catalog.(*candidateAtomicTransaction); ok {
		return candidate.CopyFrom(ctx, reader, query)
	}
	return tx.fakeCutoverTransaction.CopyFrom(ctx, reader, query)
}

// Commit records the transaction boundary without treating the deferred Rollback call as a rollback of committed data.
func (tx *appointmentTimeFakeTx) Commit(ctx context.Context) error {
	tx.events = append(tx.events, "commit")
	return tx.fakeCutoverTransaction.Commit(ctx)
}

// Rollback records denied import cleanup for ordering assertions.
func (tx *appointmentTimeFakeTx) Rollback(ctx context.Context) error {
	tx.events = append(tx.events, "rollback")
	return tx.fakeCutoverTransaction.Rollback(ctx)
}

// appointmentTimeBundle creates a source-preflight-verified bundle with explicit appointment cells.
func appointmentTimeBundle(t *testing.T, profile CutoverProfile, status, start, end string) CutoverBundle {
	t.Helper()
	return targetUniqueApplyBundle(t, profile, func(f *fixtureBundle) {
		setFixtureCell(t, f, "appointments", "status", status, profile)
		setFixtureCell(t, f, "appointments", "start_time", start, profile)
		setFixtureCell(t, f, "appointments", "end_time", end, profile)
		setFixtureCell(t, f, "appointments", "doctor_id", "1000001", profile)
	})
}

// TestCutoverAppointmentTimesReversedRejectedBeforeWrite requires transaction-local ordering rejection before any atomic COPY.
func TestCutoverAppointmentTimesReversedRejectedBeforeWrite(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile), func(t *testing.T) {
			bundle := appointmentTimeBundle(t, profile, "confirmed", "2026-08-27T10:00:00Z", "2026-08-27T09:00:00Z")
			inner := &fakeCutoverTransaction{}
			tx := &appointmentTimeFakeTx{fakeCutoverTransaction: inner, epochs: map[string]string{"2026-08-27T10:00:00Z": "1756305600", "2026-08-27T09:00:00Z": "1756302000"}}
			if profile == CutoverProfileMedicinesCandidate {
				tx.catalog = &candidateAtomicTransaction{fakeCutoverTransaction: inner, counts: map[string]int64{}}
			}
			begins := 0
			_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
				begins++
				return tx, nil
			}, bundle, validCutoverSeeds())
			if err == nil || !strings.Contains(err.Error(), "end_time") {
				t.Fatalf("reversed appointment time accepted: %v", err)
			}
			if begins != 1 || !tx.rolledBack || tx.committed || tx.copyCalls != 0 {
				t.Fatalf("begins=%d rollback=%v commit=%v copies=%d", begins, tx.rolledBack, tx.committed, tx.copyCalls)
			}
		})
	}
}

// TestCutoverAppointmentTimesTargetCollisionDenied requires rollback when the target temporal probe reports a collision.
func TestCutoverAppointmentTimesTargetCollisionDenied(t *testing.T) {
	bundle := appointmentTimeBundle(t, CutoverProfileLegacy21, "confirmed", "2026-08-27T09:00:00Z", "2026-08-27T10:00:00Z")
	tx := &appointmentTimeFakeTx{fakeCutoverTransaction: &fakeCutoverTransaction{}, collide: true, epochs: map[string]string{"2026-08-27T09:00:00Z": "1756302000", "2026-08-27T10:00:00Z": "1756305600"}}
	begins := 0
	_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
		begins++
		return tx, nil
	}, bundle, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("atomic target time collision accepted: %v", err)
	}
	if begins != 1 || !tx.rolledBack || tx.committed || tx.copyCalls != 0 {
		t.Fatalf("begins=%d rollback=%v commit=%v copies=%d", begins, tx.rolledBack, tx.committed, tx.copyCalls)
	}
}

type appointmentTimeInput struct {
	start, end, doctor, status string
}

// appointmentTimeRows rewrites only a copied test bundle's appointments and binds its new row count and digest.
func appointmentTimeRows(t *testing.T, rows []appointmentTimeInput) CutoverBundle {
	t.Helper()
	bundle := validCutoverBundleForApply(t)
	digest := rewriteTableCSV(t, bundle, "appointments", func(header []string, source [][]string) [][]string {
		out := make([][]string, len(rows))
		for i, input := range rows {
			row := append([]string(nil), source[0]...)
			for column, value := range map[string]string{
				"id": strconv.Itoa(1000001 + i), "start_time": input.start, "end_time": input.end,
				"doctor_id": input.doctor, "status": input.status,
			} {
				row[columnIndex(header, column)] = value
			}
			out[i] = row
		}
		return out
	})
	return updateManifestEntry(t, bundle, "appointments", digest, int64(len(rows)))
}

// appointmentTimeTestTx starts a fake with exact, test-supplied PostgreSQL canonical responses.
func appointmentTimeTestTx(epochs map[string]string) *appointmentTimeFakeTx {
	return &appointmentTimeFakeTx{fakeCutoverTransaction: &fakeCutoverTransaction{}, epochs: epochs}
}

// TestCutoverAppointmentTimesSourceMatrix pins storage semantics independently of timestamp string syntax.
func TestCutoverAppointmentTimesSourceMatrix(t *testing.T) {
	row := func(start, end string) appointmentTimeInput {
		return appointmentTimeInput{start, end, "1000001", "confirmed"}
	}
	epochs := map[string]string{
		"a": "0", "b": "1", "c": "2", "d": "3", "z": "10",
		"same-offset-SOURCE_SECRET": "0.000000", "same-precision": "+000.0",
		"2026-08-27T00:00:00+09:00":        "1787756400",
		"2026-08-26T15:00:00.000000Z":      "1787756400.000000",
		"2026-08-27T00:00:00.000001+09:00": "1787756400.000001",
		"2026-08-27T23:00:00+09:00":        "1787839200",
		"2026-08-29T01:00:00+09:00":        "1787932800",
		"negative":                         "-62167219200.000001",
		"negative-next":                    "-62167219200.000000",
		"extreme":                          "1000000000000000000.000001",
		"extreme-next":                     "1000000000000000000.000002",
		"bc-SOURCE_SECRET":                 "-200000000000000.000001",
		"-infinity":                        "-Infinity", "infinity": "iNfInItY",
	}
	for _, tc := range []struct {
		name string
		rows []appointmentTimeInput
		want string
	}{
		{"equal-historical-shell", []appointmentTimeInput{row("a", "a")}, ""},
		{"cross-JST-day", []appointmentTimeInput{row("2026-08-27T23:00:00+09:00", "2026-08-29T01:00:00+09:00")}, ""},
		{"reverse", []appointmentTimeInput{row("b", "a")}, "column end_time row 2"},
		{"reverse-cancelled", []appointmentTimeInput{{"b", "a", "1000001", "cancelled"}}, "column end_time row 2"},
		{"reverse-null-status", []appointmentTimeInput{{"b", "a", "1000001", ""}}, "column end_time row 2"},
		{"reverse-null-doctor", []appointmentTimeInput{{"b", "a", "", "confirmed"}}, "column end_time row 2"},
		{"offset-same-start", []appointmentTimeInput{
			row("2026-08-27T00:00:00+09:00", "2026-08-27T00:00:00+09:00"),
			row("2026-08-26T15:00:00.000000Z", "2026-08-26T15:00:00.000000Z")}, "duplicates row 2"},
		{"precision-same-start", []appointmentTimeInput{row("a", "a"), row("same-precision", "same-precision")}, "duplicates row 2"},
		{"doctor-numeric-identity", []appointmentTimeInput{
			row("a", "a"), {"same-offset-SOURCE_SECRET", "same-offset-SOURCE_SECRET", "+0001000001", "confirmed"}}, "duplicates row 2"},
		{"different-doctors", []appointmentTimeInput{row("a", "z"), {"a", "z", "1000002", "confirmed"}}, ""},
		{"null-doctors", []appointmentTimeInput{{"a", "z", "", "confirmed"}, {"a", "z", "", "confirmed"}}, ""},
		{"cancelled-excluded", []appointmentTimeInput{row("a", "z"), {"a", "z", "1000001", "cancelled"}}, ""},
		{"null-status-excluded", []appointmentTimeInput{row("a", "z"), {"a", "z", "1000001", ""}}, ""},
		{"positive-overlap", []appointmentTimeInput{row("a", "c"), row("b", "d")}, "overlaps"},
		{"contained-unsorted", []appointmentTimeInput{row("b", "c"), row("a", "z")}, "overlaps"},
		{"touching", []appointmentTimeInput{row("b", "c"), row("a", "b")}, ""},
		{"empty-inside-positive", []appointmentTimeInput{row("a", "z"), row("b", "b")}, ""},
		{"empty-same-start-still-unique", []appointmentTimeInput{row("a", "z"), row("a", "a")}, "duplicates"},
		{"negative-microsecond", []appointmentTimeInput{row("negative", "negative-next")}, ""},
		{"beyond-Go-range-exact", []appointmentTimeInput{row("extreme", "extreme-next")}, ""},
		{"beyond-Go-range-reverse", []appointmentTimeInput{row("extreme-next", "extreme")}, "end_time"},
		{"BC-epoch", []appointmentTimeInput{row("bc-SOURCE_SECRET", "negative")}, ""},
		{"infinity-empty", []appointmentTimeInput{row("infinity", "infinity")}, ""},
		{"negative-infinity-empty", []appointmentTimeInput{row("-infinity", "-infinity")}, ""},
		{"both-infinities", []appointmentTimeInput{row("-infinity", "infinity")}, ""},
		{"infinity-reverse", []appointmentTimeInput{row("infinity", "a")}, "end_time"},
		{"infinity-containing", []appointmentTimeInput{row("-infinity", "infinity"), row("a", "b")}, "overlaps"},
		{"infinity-adjacent", []appointmentTimeInput{row("-infinity", "a"), row("a", "infinity")}, ""},
		{"microsecond-adjacent", []appointmentTimeInput{
			row("2026-08-27T00:00:00+09:00", "2026-08-27T00:00:00.000001+09:00"),
			row("2026-08-27T00:00:00.000001+09:00", "2026-08-27T23:00:00+09:00")}, ""},
		{"empty-source", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := appointmentTimeTestTx(epochs)
			err := validateCutoverAppointmentTimes(context.Background(), tx, appointmentTimeRows(t, tc.rows), validCutoverSeeds())
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
			if err != nil {
				for _, probe := range tx.probes {
					if probe == "collision" {
						t.Fatal("target collision queried before source validation completed")
					}
				}
				if strings.Contains(err.Error(), "SOURCE_SECRET") {
					t.Fatal("source value leaked")
				}
			}
			if len(tc.rows) == 0 && len(tx.calls) != 0 {
				t.Fatal("empty source queried")
			}
		})
	}
}

// TestCutoverAppointmentTimesBatching checks both 256-item limits and collisions straddling canonical batches.
func TestCutoverAppointmentTimesBatching(t *testing.T) {
	for _, mode := range []string{"valid", "duplicate", "overlap"} {
		t.Run(mode, func(t *testing.T) {
			input := make([]appointmentTimeInput, 257)
			epochs := map[string]string{}
			for i := range input {
				start, end := fmt.Sprintf("SOURCE_SECRET_start_%d", i), fmt.Sprintf("SOURCE_SECRET_end_%d", i)
				epochs[start], epochs[end] = strconv.Itoa(i*2), strconv.Itoa(i*2+1)
				input[i] = appointmentTimeInput{start, end, "+0001000001", "confirmed"}
			}
			if mode == "duplicate" {
				epochs[input[256].start] = "0.000000"
				epochs[input[256].end] = "1"
			}
			if mode == "overlap" {
				epochs[input[0].end] = "0.75"
				epochs[input[256].start], epochs[input[256].end] = "0.5", "0.9"
			}
			tx := appointmentTimeTestTx(epochs)
			seeds := validCutoverSeeds()
			err := validateCutoverAppointmentTimes(context.Background(), tx, appointmentTimeRows(t, input), seeds)
			if mode != "valid" {
				if err == nil || strings.Contains(err.Error(), "SOURCE_SECRET") {
					t.Fatalf("source collision not safely denied: %v", err)
				}
				if !reflect.DeepEqual(tx.probes, []string{"epoch", "epoch"}) {
					t.Fatalf("probes = %v, want canonical batches only", tx.probes)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tx.probes, []string{"epoch", "epoch", "collision", "collision"}) {
				t.Fatalf("batch order = %v", tx.probes)
			}
			for i, call := range tx.calls {
				offset, size := (i%2)*256, 256
				if i%2 != 0 {
					size = 1
				}
				var gotStarts, gotEnds []string
				if i < 2 {
					if call.query != appointmentTimeExpectedEpochQuery || len(call.args) != 2 {
						t.Fatal("wrong canonical query or arity")
					}
					gotStarts, gotEnds = call.args[0].([]string), call.args[1].([]string)
				} else {
					if call.query != appointmentTimeExpectedCollisionQuery || len(call.args) != 4 || call.args[0] != seeds.ClinicID {
						t.Fatal("wrong target query, arity, or clinic")
					}
					doctors := call.args[1].([]string)
					if len(doctors) != size {
						t.Fatal("doctor batch wrong size")
					}
					for _, doctor := range doctors {
						if doctor != "1000001" {
							t.Fatalf("doctor was not normalized: %q", doctor)
						}
					}
					gotStarts, gotEnds = call.args[2].([]string), call.args[3].([]string)
				}
				if len(gotStarts) != size || len(gotEnds) != size {
					t.Fatal("timestamp batch wrong size")
				}
				for j := range gotStarts {
					if gotStarts[j] != input[offset+j].start || gotEnds[j] != input[offset+j].end {
						t.Fatal("timestamp source bytes or order changed")
					}
				}
			}
		})
	}
}

// TestCutoverAppointmentTimesCanonicalFailures rejects malformed PostgreSQL answer shapes and preserves query causes privately.
func TestCutoverAppointmentTimesCanonicalFailures(t *testing.T) {
	bundle := appointmentTimeRows(t, []appointmentTimeInput{{"SOURCE_SECRET", "END_SECRET", "", "cancelled"}})
	for _, tc := range []struct {
		name string
		a, b []string
	}{
		{"null-start", nil, []string{"1"}},
		{"null-end", []string{"0"}, nil},
		{"extra-start", []string{"0", "1"}, []string{"1"}},
		{"extra-end", []string{"0"}, []string{"1", "2"}},
		{"empty-start", []string{""}, []string{"1"}},
		{"NaN", []string{"NaN"}, []string{"1"}},
		{"fraction", []string{"1/2"}, []string{"1"}},
		{"exponent", []string{"1e3"}, []string{"1001"}},
		{"hex", []string{"0x10"}, []string{"17"}},
		{"whitespace", []string{" 0"}, []string{"1"}},
		{"malformed-end", []string{"0"}, []string{"PG_DRIVER_SECRET"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := appointmentTimeTestTx(nil)
			tx.queryFn = func(query string, _ ...any) pgx.Row {
				if query != appointmentTimeExpectedEpochQuery {
					t.Fatal("target query after malformed canonical answer")
				}
				return staticRow{values: []any{tc.a, tc.b}}
			}
			err := validateCutoverAppointmentTimes(context.Background(), tx, bundle, validCutoverSeeds())
			if err == nil {
				t.Fatal("invalid canonical result accepted")
			}
			for _, private := range []string{"SOURCE_SECRET", "END_SECRET", "PG_DRIVER_SECRET"} {
				if strings.Contains(err.Error(), private) {
					t.Fatal("canonical failure leaked private detail")
				}
			}
		})
	}
	for _, query := range []string{appointmentTimeExpectedEpochQuery, appointmentTimeExpectedCollisionQuery} {
		t.Run("cause-"+strconv.Itoa(len(query)), func(t *testing.T) {
			cause := errors.New("PG_DRIVER_SECRET-/private/path")
			tx := appointmentTimeTestTx(map[string]string{"SOURCE_SECRET": "0", "END_SECRET": "1"})
			tx.queryFn = func(actual string, _ ...any) pgx.Row {
				if actual == query {
					return staticRow{err: cause}
				}
				return staticRow{values: []any{[]string{"0"}, []string{"1"}}}
			}
			eligible := appointmentTimeRows(t, []appointmentTimeInput{{"SOURCE_SECRET", "END_SECRET", "1000001", "confirmed"}})
			err := validateCutoverAppointmentTimes(context.Background(), tx, eligible, validCutoverSeeds())
			if err == nil || !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET") || strings.Contains(err.Error(), "/private/path") {
				t.Fatalf("cause not safely preserved: %v", err)
			}
		})
	}
}

// TestCutoverAppointmentTimesReadGuards proves complete source binding precedes every PostgreSQL query.
func TestCutoverAppointmentTimesReadGuards(t *testing.T) {
	for _, mode := range []string{"hash", "rowcount", "manifest", "header", "parse", "symlink", "unreadable", "cancel", "doctor"} {
		t.Run(mode, func(t *testing.T) {
			bundle := appointmentTimeRows(t, []appointmentTimeInput{{"SOURCE_SECRET", "END_SECRET", "1000001", "cancelled"}})
			table, err := cutoverManifestTableByName(bundle.Manifest, "appointments")
			if err != nil {
				t.Fatal(err)
			}
			path := cutoverCSVPath(bundle.SourceDir, table)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch mode {
			case "hash":
				bundle = corruptManifestHash(t, bundle, "appointments")
			case "rowcount":
				bundle = corruptManifestRowCount(t, bundle, "appointments")
			case "manifest":
				for i, entry := range bundle.Manifest.Tables {
					if entry.Table == "appointments" {
						bundle.Manifest.Tables = append(bundle.Manifest.Tables[:i], bundle.Manifest.Tables[i+1:]...)
						break
					}
				}
			case "header":
				digest := rewriteTableCSV(t, bundle, "appointments", func(header []string, rows [][]string) [][]string {
					header[0] = "PRIVATE_HEADER_SECRET"
					return rows
				})
				bundle = updateManifestEntry(t, bundle, "appointments", digest, 1)
			case "parse":
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := f.WriteString("\"PRIVATE_PARSE_SECRET")
				closeErr := f.Close()
				if writeErr != nil || closeErr != nil {
					t.Fatalf("append = %v; close = %v", writeErr, closeErr)
				}
			case "symlink":
				moved := filepath.Join(bundle.SourceDir, "PRIVATE_PATH_SECRET.csv")
				if err := os.Rename(path, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(moved, path); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if err := os.Chmod(path, 0o644); err != nil {
					t.Fatal(err)
				}
			case "cancel":
				cancel()
			case "doctor":
				digest := rewriteTableCSV(t, bundle, "appointments", func(header []string, rows [][]string) [][]string {
					rows[0][columnIndex(header, "doctor_id")] = "PRIVATE_DOCTOR_SECRET"
					return rows
				})
				bundle = updateManifestEntry(t, bundle, "appointments", digest, 1)
			}
			tx := appointmentTimeTestTx(nil)
			err = validateCutoverAppointmentTimes(ctx, tx, bundle, validCutoverSeeds())
			if err == nil || len(tx.calls) != 0 || strings.Contains(err.Error(), "SECRET") || strings.Contains(err.Error(), bundle.SourceDir) {
				t.Fatalf("unsafe read failure/query count %d: %v", len(tx.calls), err)
			}
			if mode == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation cause lost")
			}
			if mode == "parse" {
				var parseErr *csv.ParseError
				if !errors.As(err, &parseErr) {
					t.Fatal("CSV parse cause lost")
				}
			}
		})
	}
}

// TestCutoverAppointmentTimesHeaderPermutation confines tolerated header drift to rehearsal contracts.
func TestCutoverAppointmentTimesHeaderPermutation(t *testing.T) {
	for _, allowed := range []bool{false, true} {
		t.Run(strconv.FormatBool(allowed), func(t *testing.T) {
			bundle := rehearsalBundle(t)
			if !allowed {
				bundle.Provenance.Mode = CutoverProvenanceFormal
			}
			digest := rewriteTableCSV(t, bundle, "appointments", func(header []string, rows [][]string) [][]string {
				a, b := columnIndex(header, "start_time"), columnIndex(header, "end_time")
				header[a], header[b] = header[b], header[a]
				for _, row := range rows {
					row[a], row[b] = row[b], row[a]
				}
				return rows
			})
			bundle = updateManifestEntry(t, bundle, "appointments", strings.ToUpper(digest), 1)
			tx := appointmentTimeTestTx(map[string]string{
				"2026-07-22T00:00:00.000000Z": "0", "2026-07-22T01:00:00.000000Z": "3600",
			})
			err := validateCutoverAppointmentTimes(context.Background(), tx, bundle, validCutoverSeeds())
			if allowed && err != nil || !allowed && (err == nil || len(tx.calls) != 0) {
				t.Fatalf("allowed=%v error=%v calls=%d", allowed, err, len(tx.calls))
			}
		})
	}
}

// TestCutoverAppointmentTimesCancelBetweenBatches stops even when a fake querier ignores cancellation.
func TestCutoverAppointmentTimesCancelBetweenBatches(t *testing.T) {
	for _, phase := range []string{"epoch", "collision"} {
		t.Run(phase, func(t *testing.T) {
			input := make([]appointmentTimeInput, 257)
			for i := range input {
				input[i] = appointmentTimeInput{fmt.Sprint(i * 2), fmt.Sprint(i*2 + 1), "1000001", "confirmed"}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tx := appointmentTimeTestTx(nil)
			tx.queryFn = func(query string, args ...any) pgx.Row {
				if query == appointmentTimeExpectedEpochQuery {
					if phase == "epoch" {
						cancel()
					}
					return staticRow{values: []any{args[0], args[1]}}
				}
				cancel()
				return staticRow{values: []any{false}}
			}
			err := validateCutoverAppointmentTimes(ctx, tx, appointmentTimeRows(t, input), validCutoverSeeds())
			want := []string{"epoch"}
			if phase == "collision" {
				want = []string{"epoch", "epoch", "collision"}
			}
			if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(tx.probes, want) {
				t.Fatalf("cancel error=%v probes=%v", err, tx.probes)
			}
		})
	}
}

type appointmentTimeSession struct {
	fakeCutoverSession
	epochs  map[string]string
	collide bool
	txs     []*appointmentTimeFakeTx
}

// QueryRow refuses session-level temporal probes; only the transaction may canonicalize timestamps.
func (s *appointmentTimeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if query == appointmentTimeExpectedEpochQuery || query == appointmentTimeExpectedCollisionQuery {
		return staticRow{err: errors.New("temporal query outside transaction")}
	}
	return s.fakeCutoverSession.QueryRow(ctx, query, args...)
}

// Begin retains each transaction so tests can distinguish committed prefixes from denied appointment writes.
func (s *appointmentTimeSession) Begin(context.Context, pgx.TxIsoLevel) (cutoverTransaction, error) {
	tx := appointmentTimeTestTx(s.epochs)
	tx.collide = s.collide
	tx.events = append(tx.events, "begin")
	s.txs = append(s.txs, tx)
	return tx, nil
}

// appointmentTimeAssertTrace checks lock, canonicalization, collision probe, and COPY on one transaction.
func appointmentTimeAssertTrace(t *testing.T, tx *appointmentTimeFakeTx, success bool) {
	t.Helper()
	want := []string{"begin", "lock", "epoch", "collision"}
	if success {
		want = append(want, "copy", "commit", "rollback")
	} else {
		want = append(want, "rollback")
	}
	if !success {
		if !reflect.DeepEqual(tx.events, want) {
			t.Fatalf("denied event trace = %v", tx.events)
		}
		if tx.copyCalls != 0 || tx.committed || !tx.rolledBack {
			t.Fatal("denied transaction copied or committed")
		}
		return
	}
	if len(tx.events) < len(want) || !reflect.DeepEqual(tx.events[:4], want[:4]) || tx.events[4] != "copy" ||
		tx.events[len(tx.events)-2] != "commit" || tx.events[len(tx.events)-1] != "rollback" {
		t.Fatalf("successful event trace = %v", tx.events)
	}
	if tx.copyCalls == 0 || !tx.committed {
		t.Fatal("successful transaction never copied/committed")
	}
}

// TestCutoverAppointmentTimesTransactionSeams keeps all temporal parsing on the locked COPY transaction.
func TestCutoverAppointmentTimesTransactionSeams(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, collision := range []bool{false, true} {
			t.Run(string(profile)+"/atomic/"+strconv.FormatBool(collision), func(t *testing.T) {
				bundle := appointmentTimeBundle(t, profile, "confirmed", "SOURCE_SECRET_START", "SOURCE_SECRET_END")
				inner := &fakeCutoverTransaction{}
				tx := &appointmentTimeFakeTx{fakeCutoverTransaction: inner, collide: collision,
					epochs: map[string]string{"SOURCE_SECRET_START": "0", "SOURCE_SECRET_END": "1"}}
				if profile == CutoverProfileMedicinesCandidate {
					tx.catalog = &candidateAtomicTransaction{fakeCutoverTransaction: inner, counts: map[string]int64{}}
				}
				_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
					tx.events = append(tx.events, "begin")
					return tx, nil
				}, bundle, validCutoverSeeds())
				if collision && (err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied)) || !collision && err != nil {
					t.Fatalf("collision=%v error=%v", collision, err)
				}
				appointmentTimeAssertTrace(t, tx, !collision)
			})
		}
	}
	for _, collision := range []bool{false, true} {
		t.Run("legacy-table/"+strconv.FormatBool(collision), func(t *testing.T) {
			bundle := appointmentTimeBundle(t, CutoverProfileLegacy21, "confirmed", "SOURCE_SECRET_START", "SOURCE_SECRET_END")
			session := &appointmentTimeSession{collide: collision,
				epochs: map[string]string{"SOURCE_SECRET_START": "0", "SOURCE_SECRET_END": "1"}}
			spec := ddlSpecByName(t, "appointments")
			table, err := cutoverManifestTableByName(bundle.Manifest, "appointments")
			if err != nil {
				t.Fatal(err)
			}
			_, err = importOneCutoverTable(context.Background(), session, pgx.RepeatableRead, bundle, spec, table, validCutoverSeeds())
			if collision && (err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied)) || !collision && err != nil {
				t.Fatalf("collision=%v error=%v", collision, err)
			}
			if len(session.txs) != 1 {
				t.Fatal("wrong transaction count")
			}
			appointmentTimeAssertTrace(t, session.txs[0], !collision)
		})
	}
}

// TestCutoverAppointmentTimesLegacyFailureRetainsPrefix preserves earlier table commits on appointment denial.
func TestCutoverAppointmentTimesLegacyFailureRetainsPrefix(t *testing.T) {
	bundle := appointmentTimeBundle(t, CutoverProfileLegacy21, "confirmed", "SOURCE_SECRET_START", "SOURCE_SECRET_END")
	session := &appointmentTimeSession{collide: true,
		epochs: map[string]string{"SOURCE_SECRET_START": "0", "SOURCE_SECRET_END": "1"}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("legacy collision accepted: %v", err)
	}
	specs := CutoverTableSpecs()
	appointmentIndex := -1
	for i, spec := range specs {
		if spec.Name == "appointments" {
			appointmentIndex = i
		}
	}
	if appointmentIndex < 1 || len(session.txs) != appointmentIndex+2 {
		t.Fatal("did not reach intended appointment transaction")
	}
	early := session.txs[0]
	if !early.rolledBack || early.committed || early.copyCalls != 0 {
		t.Fatal("early source tx must roll back with no writes")
	}
	for _, tx := range session.txs[1 : appointmentIndex+1] {
		if !tx.committed || tx.copyCalls != 1 || len(tx.calls) != 0 {
			t.Fatal("earlier table commit was erased or received temporal queries")
		}
	}
	appointmentTimeAssertTrace(t, session.txs[appointmentIndex+1], false)
}

// TestCutoverAppointmentTimesCompletedPrefixSkipsSelf avoids target-probing already-adopted appointment rows on legacy resume.
func TestCutoverAppointmentTimesCompletedPrefixSkipsSelf(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	complete := map[string]int64{}
	for _, spec := range CutoverTableSpecs() {
		complete[spec.Name] = 1
		if spec.Name == "appointments" {
			break
		}
	}
	session := &appointmentTimeSession{fakeCutoverSession: fakeCutoverSession{inspectCounts: complete},
		epochs: map[string]string{"2026-07-22T00:00:00.000000Z": "0", "2026-07-22T01:00:00.000000Z": "3600"}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.txs) < 2 || !reflect.DeepEqual(session.txs[0].events, []string{"begin", "epoch", "rollback"}) ||
		session.txs[0].copyCalls != 0 || session.txs[0].committed {
		t.Fatal("completed prefix did not receive only the early source transaction")
	}
	for _, tx := range session.txs[1:] {
		if len(tx.calls) != 0 {
			t.Fatal("completed appointments were temporally re-probed")
		}
		for _, query := range tx.copySQLs {
			if strings.Contains(query, `"appointments"`) {
				t.Fatal("completed appointments copied again")
			}
		}
	}
}

// TestCutoverAppointmentTimesOccupiedBandStopsBeforeParsing keeps the original locked band guard ahead of temporal queries.
func TestCutoverAppointmentTimesOccupiedBandStopsBeforeParsing(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := appointmentTimeTestTx(nil)
	tx.occupied = true
	session := &appointmentTimeOneTxSession{tx: tx}
	spec := ddlSpecByName(t, "appointments")
	table, err := cutoverManifestTableByName(bundle.Manifest, spec.Name)
	if err != nil {
		t.Fatal(err)
	}
	_, err = importOneCutoverTable(context.Background(), session, pgx.RepeatableRead, bundle, spec, table, validCutoverSeeds())
	if err == nil || len(tx.calls) != 0 || tx.copyCalls != 0 || tx.committed {
		t.Fatalf("occupied guard error=%v calls=%d", err, len(tx.calls))
	}
}

type appointmentTimeOneTxSession struct {
	fakeCutoverSession
	tx *appointmentTimeFakeTx
}

// Begin supplies the preconfigured appointment transaction for the occupied-band regression.
func (s *appointmentTimeOneTxSession) Begin(context.Context, pgx.TxIsoLevel) (cutoverTransaction, error) {
	return s.tx, nil
}

// appointmentTimeRound3Session records every transaction Begin and answers
// only the canonical epoch query inside each tx via an explicit per-text map.
type appointmentTimeRound3Session struct {
	fakeCutoverSession
	beginErr      error
	epochs        map[string]string
	txs           []*appointmentTimeFakeTx
	queryErr      error
	reverseActual bool
	events        []string
}

// Begin records the early transaction separately and refuses to hide an unclosed source transaction.
func (s *appointmentTimeRound3Session) Begin(ctx context.Context, iso pgx.TxIsoLevel) (cutoverTransaction, error) {
	s.events = append(s.events, "begin")
	if s.beginErr != nil {
		return nil, s.beginErr
	}
	if len(s.txs) == 1 && (!s.txs[0].rolledBack || s.txs[0].committed || s.txs[0].copyCalls != 0) {
		return nil, errors.New("source transaction was not rolled back before a write Begin")
	}
	inner := &fakeCutoverTransaction{}
	tx := &appointmentTimeFakeTx{fakeCutoverTransaction: inner, epochs: s.epochs}
	tx.events = append(tx.events, "begin")
	isEarly := len(s.txs) == 0
	if s.queryErr != nil && isEarly {
		tx.queryFn = func(query string, _ ...any) pgx.Row {
			if query != appointmentTimeExpectedEpochQuery {
				return staticRow{err: errUnexpectedQuery}
			}
			return staticRow{err: s.queryErr}
		}
	}
	if s.reverseActual && !isEarly {
		tx.queryFn = func(query string, args ...any) pgx.Row {
			if query != appointmentTimeExpectedEpochQuery || len(args[0].([]string)) != 1 {
				return staticRow{err: errUnexpectedQuery}
			}
			return staticRow{values: []any{[]string{"1"}, []string{"0"}}}
		}
	}
	s.txs = append(s.txs, tx)
	return tx, nil
}

// QueryRow rejects session-level canonicalization and target temporal probes.
func (s *appointmentTimeRound3Session) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if query == appointmentTimeExpectedEpochQuery || query == appointmentTimeExpectedCollisionQuery {
		return staticRow{err: errors.New("temporal query outside explicit transaction")}
	}
	return s.fakeCutoverSession.QueryRow(ctx, query, args...)
}

// appointmentTimeRound3Bundle rewrites the appointments CSV to the given raw
// (start,end) pairs, keeps the manifest digest/row-count consistent, and maps
// each raw timestamp string to a synthetic canonical epoch.
func appointmentTimeRound3Bundle(t *testing.T, rows [][2]string) CutoverBundle {
	t.Helper()
	bundle := targetUniqueApplyBundle(t, CutoverProfileLegacy21, nil)
	newHash := rewriteTableCSV(t, bundle, "appointments", func(header []string, csvRows [][]string) [][]string {
		si, ei := columnIndex(header, "start_time"), columnIndex(header, "end_time")
		di, st := columnIndex(header, "doctor_id"), columnIndex(header, "status")
		out := make([][]string, len(rows))
		for i, tr := range rows {
			r := append([]string(nil), csvRows[i%len(csvRows)]...)
			r[si], r[ei], r[di], r[st] = tr[0], tr[1], "1000001", "confirmed"
			r[columnIndex(header, "id")] = strconv.Itoa(1000001 + i)
			out[i] = r
		}
		return out
	})
	return updateManifestEntry(t, bundle, "appointments", newHash, int64(len(rows)))
}

// TestCutoverAppointmentTimesRound3SourceFailuresNoPrefix stops known source errors before any legacy table write.
func TestCutoverAppointmentTimesRound3SourceFailuresNoPrefix(t *testing.T) {
	cases := []struct {
		name   string
		rows   [][2]string
		epochs map[string]string
		want   string
	}{
		{"malformed-pg", [][2]string{{"T0", "T1"}}, nil, "temporal preflight"},
		{"reversed", [][2]string{{"T1", "T0"}}, map[string]string{"T0": "0", "T1": "1"}, "end_time"},
		{"duplicate-start", [][2]string{{"T0", "T1"}, {"T0", "T1"}}, map[string]string{"T0": "0", "T1": "1"}, "start_time"},
		{"overlap", [][2]string{{"T0", "T3"}, {"T1", "T2"}}, map[string]string{"T0": "0", "T1": "1", "T2": "2", "T3": "3"}, "start_time"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bundle := appointmentTimeRound3Bundle(t, c.rows)
			s := &appointmentTimeRound3Session{epochs: c.epochs}
			_, err := applyCutoverCommittingEachTable(context.Background(), s, bundle, validCutoverSeeds(), pgx.RepeatableRead)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("%s: error = %v, want %q", c.name, err, c.want)
			}
			if len(s.txs) != 1 {
				t.Fatalf("txs = %d, want exactly 1 early source tx", len(s.txs))
			}
			tx := s.txs[0]
			if tx.copyCalls != 0 || tx.committed || !tx.rolledBack ||
				!reflect.DeepEqual(tx.events, []string{"begin", "epoch", "rollback"}) {
				t.Fatalf("early tx must not copy/commit: copies=%d commit=%v", tx.copyCalls, tx.committed)
			}
		})
	}
}

// TestCutoverAppointmentTimesRound3EarlyFailures preserves safe causes without beginning a write transaction.
func TestCutoverAppointmentTimesRound3EarlyFailures(t *testing.T) {
	for _, phase := range []string{"begin", "query"} {
		t.Run(phase, func(t *testing.T) {
			cause := errors.New("DRIVER_SOURCE_SECRET /private/path")
			session := &appointmentTimeRound3Session{}
			if phase == "begin" {
				session.beginErr = cause
			} else {
				session.queryErr = cause
			}
			bundle := appointmentTimeBundle(t, CutoverProfileLegacy21, "confirmed",
				"2026-07-22T00:00:00.000000Z", "2026-07-22T01:00:00.000000Z")
			_, err := applyCutoverCommittingEachTable(context.Background(), session,
				bundle, validCutoverSeeds(), pgx.RepeatableRead)
			if !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET") ||
				strings.Contains(err.Error(), "/private/path") {
				t.Fatalf("cause not safely preserved: %v", err)
			}
			if len(session.events) != 1 {
				t.Fatalf("write Begin after early failure: %v", session.events)
			}
			for _, tx := range session.txs {
				if !tx.rolledBack || tx.copyCalls != 0 || tx.committed ||
					!reflect.DeepEqual(tx.events, []string{"begin", "epoch", "rollback"}) {
					t.Fatal("early query failure wrote or failed to roll back")
				}
			}
		})
	}
}

// TestCutoverAppointmentTimesRound3EarlySuccessAndRevalidation separates early parsing from locked COPY-time parsing.
func TestCutoverAppointmentTimesRound3EarlySuccessAndRevalidation(t *testing.T) {
	for _, reverseActual := range []bool{false, true} {
		t.Run(strconv.FormatBool(reverseActual), func(t *testing.T) {
			session := &appointmentTimeRound3Session{
				epochs:        map[string]string{"2026-07-22T00:00:00.000000Z": "0", "2026-07-22T01:00:00.000000Z": "3600"},
				reverseActual: reverseActual,
			}
			bundle := appointmentTimeBundle(t, CutoverProfileLegacy21, "confirmed",
				"2026-07-22T00:00:00.000000Z", "2026-07-22T01:00:00.000000Z")
			_, err := applyCutoverCommittingEachTable(context.Background(), session,
				bundle, validCutoverSeeds(), pgx.RepeatableRead)
			if reverseActual && (err == nil || !strings.Contains(err.Error(), "end_time")) || !reverseActual && err != nil {
				t.Fatalf("actual revalidation reverse=%v error=%v", reverseActual, err)
			}
			early := session.txs[0]
			if !reflect.DeepEqual(early.events, []string{"begin", "epoch", "rollback"}) ||
				early.copyCalls != 0 || early.committed || !early.rolledBack {
				t.Fatal("early transaction targeted or wrote appointments")
			}
			var actual *appointmentTimeFakeTx
			for _, tx := range session.txs[1:] {
				if len(tx.calls) > 0 {
					if actual != nil {
						t.Fatal("multiple actual appointment transactions")
					}
					actual = tx
				}
			}
			if actual == nil {
				t.Fatal("no fresh canonicalization in the appointments transaction")
			}
			if reverseActual {
				if !reflect.DeepEqual(actual.events, []string{"begin", "lock", "epoch", "rollback"}) ||
					actual.committed || actual.copyCalls != 0 {
					t.Fatal("early canonical values were reused at COPY time")
				}
			} else {
				appointmentTimeAssertTrace(t, actual, true)
			}
		})
	}
}

// TestCutoverAppointmentTimesRound3CandidateKeepsSingleTransaction excludes the legacy source-only transaction from candidates.
func TestCutoverAppointmentTimesRound3CandidateKeepsSingleTransaction(t *testing.T) {
	session := &candidateFreshApplySession{}
	bundle := targetUniqueApplyBundle(t, CutoverProfileMedicinesCandidate, nil)
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.txs) != 1 || !session.txs[0].committed || session.txs[0].copyCalls == 0 {
		t.Fatal("candidate gained an extra source-only transaction")
	}
}
