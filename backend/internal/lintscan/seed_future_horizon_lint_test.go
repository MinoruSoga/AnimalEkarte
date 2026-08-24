package lintscan

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	seedFutureHorizonBundle = "003_demo"
	seedFutureHorizonDays   = 30
)

// seedFutureHorizonTargets are the demo-seed tables whose future-dated rows
// UAT booking (S04 / V05-6) and clinic-holiday guard verification depend on.
// appointments.start_time is intentionally omitted: the 003_demo dump's
// appointment schedule values top out in 2023 even though created_at is 2026,
// so a start_time horizon would stay red under any same-offset rebase that
// keeps table-to-table relative dates.
var seedFutureHorizonTargets = []seedFutureHorizonTarget{
	{csvFile: "shift_entries.csv", column: "date"},
	{csvFile: "clinic_holidays.csv", column: "date"},
}

type seedFutureHorizonTarget struct {
	csvFile string
	column  string
}

type seedFutureHorizonViolation struct {
	csvFile string
	column  string
	minDate string
	maxDate string
	want    string
	reason  string
}

func TestDemoSeedFutureHorizon_ShiftEntriesAndHolidaysReachTodayPlus30(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)
	bundleDir := filepath.Join(moduleRoot, "migrations", "seeds", seedFutureHorizonBundle)
	today := seedFutureHorizonTodayJST()
	horizon := today.AddDate(0, 0, seedFutureHorizonDays)

	violations, err := findDemoSeedFutureHorizonViolations(bundleDir, today, horizon)
	if err != nil {
		t.Fatalf("scan %s future horizon: %v", seedFutureHorizonBundle, err)
	}
	for _, v := range violations {
		t.Errorf(
			"seed CSV %s/%s column %s min=%s max=%s want %s (%s)",
			seedFutureHorizonBundle, v.csvFile, v.column, v.minDate, v.maxDate, v.want, v.reason,
		)
	}
}

func TestDemoSeedFutureHorizon_DetectsZeroFutureFixture(t *testing.T) {
	bundleDir := t.TempDir()
	past := time.Date(2026, 5, 1, 0, 0, 0, 0, seedFutureHorizonJST())
	if err := writeSeedFutureHorizonFixture(bundleDir, past); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	today := time.Date(2026, 8, 24, 0, 0, 0, 0, seedFutureHorizonJST())
	horizon := today.AddDate(0, 0, seedFutureHorizonDays)
	violations, err := findDemoSeedFutureHorizonViolations(bundleDir, today, horizon)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	if len(violations) != len(seedFutureHorizonTargets) {
		t.Fatalf("violations = %d (%v), want %d (one per target on a past-only fixture)", len(violations), violations, len(seedFutureHorizonTargets))
	}
	for _, v := range violations {
		if v.maxDate != "2026-05-01" {
			t.Errorf("%s maxDate = %q, want 2026-05-01", v.csvFile, v.maxDate)
		}
		if v.reason != "future-dated seed rows exhausted" {
			t.Errorf("%s reason = %q, want exhausted", v.csvFile, v.reason)
		}
	}
}

func TestDemoSeedFutureHorizon_DetectsShiftWindowAfterToday(t *testing.T) {
	bundleDir := t.TempDir()
	// +180-day hole: max is past the 30-day horizon, but the window starts after today.
	if err := writeSeedFutureHorizonRangeFixture(
		bundleDir,
		time.Date(2026, 10, 28, 0, 0, 0, 0, seedFutureHorizonJST()),
		time.Date(2026, 12, 28, 0, 0, 0, 0, seedFutureHorizonJST()),
	); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	today := time.Date(2026, 8, 24, 0, 0, 0, 0, seedFutureHorizonJST())
	horizon := today.AddDate(0, 0, seedFutureHorizonDays)
	violations, err := findDemoSeedFutureHorizonViolations(bundleDir, today, horizon)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	foundShiftHole := false
	for _, v := range violations {
		if v.csvFile == "shift_entries.csv" && v.reason == "shift window starts after today" {
			foundShiftHole = true
		}
	}
	if !foundShiftHole {
		t.Fatalf("violations = %v, want shift_entries window-starts-after-today", violations)
	}
}

func TestFindDemoSeedFutureHorizonViolations_AllowsDatesOnHorizon(t *testing.T) {
	bundleDir := t.TempDir()
	today := time.Date(2026, 8, 24, 0, 0, 0, 0, seedFutureHorizonJST())
	horizon := time.Date(2026, 9, 23, 0, 0, 0, 0, seedFutureHorizonJST())
	if err := writeSeedFutureHorizonRangeFixture(bundleDir, today, horizon); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	violations, err := findDemoSeedFutureHorizonViolations(bundleDir, today, horizon)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %v, want none when the window covers today through the horizon", violations)
	}
}

func findDemoSeedFutureHorizonViolations(bundleDir string, today, horizon time.Time) ([]seedFutureHorizonViolation, error) {
	var violations []seedFutureHorizonViolation
	for _, target := range seedFutureHorizonTargets {
		earliest, latest, err := dateRangeInSeedCSVColumn(filepath.Join(bundleDir, target.csvFile), target.column)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", target.csvFile, err)
		}
		if latest.Before(horizon) {
			violations = append(violations, seedFutureHorizonViolation{
				csvFile: target.csvFile,
				column:  target.column,
				minDate: earliest.Format(time.DateOnly),
				maxDate: latest.Format(time.DateOnly),
				want:    horizon.Format(time.DateOnly),
				reason:  "future-dated seed rows exhausted",
			})
		}
		if target.csvFile == "shift_entries.csv" && earliest.After(today) {
			violations = append(violations, seedFutureHorizonViolation{
				csvFile: target.csvFile,
				column:  target.column,
				minDate: earliest.Format(time.DateOnly),
				maxDate: latest.Format(time.DateOnly),
				want:    today.Format(time.DateOnly),
				reason:  "shift window starts after today",
			})
		}
	}
	return violations, nil
}

func dateRangeInSeedCSVColumn(path, column string) (time.Time, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("read header: %w", err)
	}
	col := -1
	for i, name := range header {
		if name == column {
			col = i
			break
		}
	}
	if col < 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("column %q not in header %v", column, header)
	}

	var earliest, latest time.Time
	found := false
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if col >= len(record) {
			continue
		}
		raw := record[col]
		if raw == "" {
			continue
		}
		parsed, ok := parseSeedCSVDatePrefix(raw)
		if !ok {
			continue
		}
		if !found || parsed.Before(earliest) {
			earliest = parsed
		}
		if !found || parsed.After(latest) {
			latest = parsed
		}
		found = true
	}
	if !found {
		return time.Time{}, time.Time{}, fmt.Errorf("column %q has no parseable dates", column)
	}
	return earliest, latest, nil
}

func parseSeedCSVDatePrefix(raw string) (time.Time, bool) {
	if len(raw) < 10 {
		return time.Time{}, false
	}
	parsed, err := time.ParseInLocation(time.DateOnly, raw[:10], seedFutureHorizonJST())
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func seedFutureHorizonJST() *time.Location {
	return time.FixedZone("JST", 9*60*60)
}

func seedFutureHorizonTodayJST() time.Time {
	now := time.Now().In(seedFutureHorizonJST())
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func writeSeedFutureHorizonFixture(dir string, day time.Time) error {
	return writeSeedFutureHorizonRangeFixture(dir, day, day)
}

func writeSeedFutureHorizonRangeFixture(dir string, start, end time.Time) error {
	startValue := start.Format(time.DateOnly)
	endValue := end.Format(time.DateOnly)
	files := map[string]string{
		"shift_entries.csv": "id,clinic_id,staff_id,date\n" +
			"1,1,1," + startValue + "\n" +
			"2,1,1," + endValue + "\n",
		"clinic_holidays.csv": "id,clinic_id,date\n" +
			"1,1," + startValue + "\n" +
			"2,1," + endValue + "\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			return err
		}
	}
	return nil
}
