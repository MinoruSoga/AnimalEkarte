package lintscan

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedCSVQuotedEmptyOccurrenceFloors is the 7d286ca71^ count of `""` bytes
// in each 003_demo CSV that stores quoted empty strings. QUOTE_MINIMAL rewrites
// drop these to unquoted empty and zero the count (the 7d286ca71 defect).
// shift_entries.csv may exceed the floor after a window clone.
var seedCSVQuotedEmptyOccurrenceFloors = map[string]int{
	"appointments.csv":                     72074,
	"audit_logs.csv":                       74,
	"cash_register_closes.csv":             12,
	"checkup_type_fields.csv":              976,
	"chief_complaint_types.csv":            18,
	"clinic_integrations.csv":              2,
	"clinical_plans.csv":                   348340,
	"clinics.csv":                          38,
	"consultations.csv":                    17,
	"exam_results.csv":                     5716588,
	"exam_type_fields.csv":                 144,
	"inquiries.csv":                        2782176,
	"insurances.csv":                       3,
	"line_customers.csv":                   48,
	"line_reservation_settings.csv":        243,
	"lstep_friend_attribute_snapshots.csv": 12,
	"merchandise_items.csv":                2,
	"occupations.csv":                      12,
	"owners.csv":                           72448,
	"pets.csv":                             90594,
	"reservation_types.csv":                174,
	"shift_entries.csv":                    1461,
	"shift_templates.csv":                  18,
	"staffs.csv":                           844,
	"vaccinations.csv":                     365684,
	"vaccines.csv":                         56,
	"vital_records.csv":                    415119,
}

type seedCSVQuotedEmptyViolation struct {
	csvFile string
	got     int
	want    int
}

func TestSeedCSVQuotedEmptyLines_CurrentDemoMeetsParentFloors(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)
	bundleDir := filepath.Join(moduleRoot, "migrations", "seeds", "003_demo")
	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		t.Skip("003_demo retired")
	}

	violations, err := findSeedCSVQuotedEmptyFloorViolations(bundleDir, seedCSVQuotedEmptyOccurrenceFloors)
	if err != nil {
		t.Fatalf("scan quoted-empty floors: %v", err)
	}
	for _, v := range violations {
		t.Errorf(
			"seed CSV 003_demo/%s has %d occurrences of %q, want at least %d (quoted empty dropped)",
			v.csvFile, v.got, `""`, v.want,
		)
	}
}

func TestFindSeedCSVQuotedEmptyFloorViolations_DetectsDroppedQuotes(t *testing.T) {
	dir := t.TempDir()
	quoted := "id,name,note\n1,a,\"\"\n"
	dropped := "id,name,note\n1,a,\n"
	if err := os.WriteFile(filepath.Join(dir, "pets.csv"), []byte(dropped), 0o600); err != nil {
		t.Fatalf("write dropped quotes: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "owners.csv"), []byte(quoted), 0o600); err != nil {
		t.Fatalf("write quoted: %v", err)
	}

	floors := map[string]int{"pets.csv": 1, "owners.csv": 1}
	violations, err := findSeedCSVQuotedEmptyFloorViolations(dir, floors)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	if len(violations) != 1 || violations[0].csvFile != "pets.csv" || violations[0].got != 0 {
		t.Fatalf("violations = %+v, want pets.csv got=0", violations)
	}

	if err := os.WriteFile(filepath.Join(dir, "pets.csv"), []byte(quoted), 0o600); err != nil {
		t.Fatalf("write restored: %v", err)
	}
	violations, err = findSeedCSVQuotedEmptyFloorViolations(dir, floors)
	if err != nil {
		t.Fatalf("scan restored: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none after restoring quotes", violations)
	}
}

func TestFindSeedCSVQuotedEmptyFloorViolations_DetectsCSVWriterMinimalQuotes(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "name", "note"}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if err := writer.Write([]string{"1", "a", ""}); err != nil {
		t.Fatalf("write row: %v", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pets.csv"), buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write writer output: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "owners.csv"), []byte("id,name,note\n1,a,\"\"\n"), 0o600); err != nil {
		t.Fatalf("write quoted owners: %v", err)
	}

	violations, err := findSeedCSVQuotedEmptyFloorViolations(dir, map[string]int{"pets.csv": 1, "owners.csv": 1})
	if err != nil {
		t.Fatalf("scan csv.Writer output: %v", err)
	}
	if len(violations) != 1 || violations[0].csvFile != "pets.csv" {
		t.Fatalf("violations = %+v, want pets.csv from csv.Writer QUOTE_MINIMAL", violations)
	}
}

func TestParseCSVLineFields_DistinguishesQuotedEmpty(t *testing.T) {
	fields, err := parseCSVLineFields(`1000001,1,"",,alive`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(fields) != 5 {
		t.Fatalf("len(fields)=%d, want 5", len(fields))
	}
	if !fields[2].quoted || fields[2].value != "" {
		t.Fatalf("field 2 = %+v, want quoted empty", fields[2])
	}
	if fields[3].quoted || fields[3].value != "" {
		t.Fatalf("field 3 = %+v, want unquoted empty", fields[3])
	}
}

func findSeedCSVQuotedEmptyFloorViolations(bundleDir string, floors map[string]int) ([]seedCSVQuotedEmptyViolation, error) {
	var violations []seedCSVQuotedEmptyViolation
	for name, want := range floors {
		got, err := countQuotedEmptyOccurrences(filepath.Join(bundleDir, name))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if got < want {
			violations = append(violations, seedCSVQuotedEmptyViolation{csvFile: name, got: got, want: want})
		}
	}
	return violations, nil
}

func countQuotedEmptyOccurrences(path string) (int, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	firstLine, _, _ := strings.Cut(string(body), "\n")
	if err := errIfSeedCSVLFSPointer("003_demo", filepath.Base(path), firstLine); err != nil {
		return 0, err
	}
	return bytes.Count(body, []byte(`""`)), nil
}

type csvLineField struct {
	value  string
	quoted bool
}

func parseCSVLineFields(line string) ([]csvLineField, error) {
	line = strings.TrimRight(line, "\r\n")
	var fields []csvLineField
	i := 0
	n := len(line)
	if n == 0 {
		return []csvLineField{{}}, nil
	}
	for i <= n {
		if i == n {
			break
		}
		if line[i] == '"' {
			var b strings.Builder
			i++
			closed := false
			for i < n {
				if line[i] == '"' {
					if i+1 < n && line[i+1] == '"' {
						b.WriteByte('"')
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				b.WriteByte(line[i])
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated quoted field")
			}
			fields = append(fields, csvLineField{value: b.String(), quoted: true})
			if i == n {
				break
			}
			if line[i] != ',' {
				return nil, fmt.Errorf("expected comma after quoted field")
			}
			i++
			if i == n {
				fields = append(fields, csvLineField{})
				break
			}
			continue
		}
		j := i
		for j < n && line[j] != ',' {
			j++
		}
		fields = append(fields, csvLineField{value: line[i:j], quoted: false})
		if j == n {
			break
		}
		i = j + 1
		if i == n {
			fields = append(fields, csvLineField{})
			break
		}
	}
	return fields, nil
}
