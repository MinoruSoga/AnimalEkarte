package lintscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedCSVColumnMeta is the NOT NULL / type slice of migration DDL used by the
// unquoted-empty gate. FORCE_NOT_NULL in the seed loader can only absorb text
// family empties; non-text NOT NULL empties are real data defects.
type seedCSVColumnMeta struct {
	Name    string
	Type    string
	NotNull bool
}

type seedCSVNotNullEmptyViolation struct {
	bundle   string
	table    string
	csvFile  string
	column   string
	dataType string
	rows     int
}

func TestSeedCSVNotNullNonTextUnquotedEmpty_CurrentSeedsHaveNone(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)

	violations, err := findSeedCSVNotNullNonTextUnquotedEmpties(moduleRoot)
	if err != nil {
		t.Fatalf("scan not-null non-text unquoted empties: %v", err)
	}
	for _, v := range violations {
		t.Errorf(
			"seed CSV %s/%s table %s column %s (%s NOT NULL) has %d unquoted empty field(s); "+
				"loader FORCE_NOT_NULL cannot absorb non-text empties",
			v.bundle, v.csvFile, v.table, v.column, v.dataType, v.rows,
		)
	}
}

func TestFindSeedCSVNotNullNonTextUnquotedEmpties_DetectsFixture(t *testing.T) {
	moduleRoot := t.TempDir()
	migrationsDir := filepath.Join(moduleRoot, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o750); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	ddl := `
CREATE TABLE demo_force (
  id bigint PRIMARY KEY,
  note text NOT NULL DEFAULT '',
  amount numeric NOT NULL DEFAULT 0,
  optional_note text
);
`
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_init.sql"), []byte(ddl), 0o600); err != nil {
		t.Fatalf("write ddl: %v", err)
	}

	// Only 003_demo is required for this unit; other bundles get empty manifests
	// with a matching table so loaders that walk all bundles stay happy.
	for _, bundle := range seedCSVBundleNames {
		bundleDir := filepath.Join(migrationsDir, "seeds", bundle)
		if err := os.MkdirAll(bundleDir, 0o750); err != nil {
			t.Fatalf("mkdir bundle: %v", err)
		}
		manifest := `{"tables":[{"table":"demo_force","csvFile":"demo_force.csv"}]}`
		if err := os.WriteFile(filepath.Join(bundleDir, "manifest.json"), []byte(manifest), 0o600); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		// amount (numeric NOT NULL) has unquoted empty; note (text NOT NULL) empty is allowed;
		// optional_note empty is allowed (nullable).
		csv := "id,note,amount,optional_note\n1,,,x\n"
		if err := os.WriteFile(filepath.Join(bundleDir, "demo_force.csv"), []byte(csv), 0o600); err != nil {
			t.Fatalf("write csv: %v", err)
		}
	}

	violations, err := findSeedCSVNotNullNonTextUnquotedEmpties(moduleRoot)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	if len(violations) == 0 {
		t.Fatal("expected at least one non-text NOT NULL unquoted-empty violation, got none")
	}
	foundAmount := false
	for _, v := range violations {
		if v.column == "note" {
			t.Fatalf("text NOT NULL unquoted empty must be allowed, got %#v", v)
		}
		if v.column == "optional_note" {
			t.Fatalf("nullable column must not be reported, got %#v", v)
		}
		if v.column == "amount" {
			foundAmount = true
			if v.rows < 1 {
				t.Fatalf("amount rows = %d, want >= 1", v.rows)
			}
		}
	}
	if !foundAmount {
		t.Fatalf("expected amount violation, got %#v", violations)
	}
}

func TestParseSeedCSVColumnMeta_NotNullOutsideCheck(t *testing.T) {
	ddl := `
CREATE TABLE reservation_type_unavailable_times (
  id bigint PRIMARY KEY,
  specific_date date CHECK (
    (unavailable_type = 'specific' AND specific_date IS NOT NULL)
    OR (unavailable_type = 'weekly' AND specific_date IS NULL)
  ),
  memo text NOT NULL DEFAULT '',
  amount numeric(10,2) NOT NULL
);
ALTER TABLE reservation_type_unavailable_times
  ADD COLUMN label varchar(50) NOT NULL DEFAULT '';
`
	meta, err := parseSeedCSVColumnMeta(ddl)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cols := meta["reservation_type_unavailable_times"]
	byName := map[string]seedCSVColumnMeta{}
	for _, c := range cols {
		byName[c.Name] = c
	}
	if byName["specific_date"].NotNull {
		t.Fatalf("specific_date must not be treated as column NOT NULL (CHECK only): %#v", byName["specific_date"])
	}
	if !byName["memo"].NotNull || !isSeedCSVTextDataType(byName["memo"].Type) {
		t.Fatalf("memo meta = %#v", byName["memo"])
	}
	if !byName["amount"].NotNull || isSeedCSVTextDataType(byName["amount"].Type) {
		t.Fatalf("amount meta = %#v", byName["amount"])
	}
	if !byName["label"].NotNull || !isSeedCSVTextDataType(byName["label"].Type) {
		t.Fatalf("label meta = %#v", byName["label"])
	}
	if !byName["id"].NotNull {
		t.Fatalf("PRIMARY KEY id should imply NOT NULL: %#v", byName["id"])
	}
}

func findSeedCSVNotNullNonTextUnquotedEmpties(moduleRoot string) ([]seedCSVNotNullEmptyViolation, error) {
	ddl, migrationNames, err := seedCSVReadOrderedMigrationDDL(moduleRoot)
	if err != nil {
		return nil, err
	}
	metaByTable, err := parseSeedCSVColumnMeta(ddl)
	if err != nil {
		return nil, fmt.Errorf("parse column meta from %v: %w", migrationNames, err)
	}

	var violations []seedCSVNotNullEmptyViolation
	for _, bundle := range seedCSVBundleNames {
		bundleDir := filepath.Join(moduleRoot, "migrations", "seeds", bundle)
		manifestPath := filepath.Join(bundleDir, "manifest.json")
		manifest, err := readSeedCSVManifest(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("load bundle %s: %w", bundle, err)
		}
		for _, entry := range manifest.Tables {
			tableName := normalizeSeedCSVSQLIdentifier(entry.Table)
			cols, ok := metaByTable[tableName]
			if !ok {
				return nil, fmt.Errorf("bundle %s table %s missing from migration DDL", bundle, entry.Table)
			}
			byName := make(map[string]seedCSVColumnMeta, len(cols))
			for _, c := range cols {
				byName[c.Name] = c
			}
			csvPath := filepath.Join(bundleDir, entry.CSVFile)
			counts, err := countUnquotedEmptyFieldsByColumn(csvPath)
			if err != nil {
				return nil, fmt.Errorf("scan %s/%s: %w", bundle, entry.CSVFile, err)
			}
			for col, n := range counts {
				if n == 0 {
					continue
				}
				meta, ok := byName[col]
				if !ok || !meta.NotNull {
					continue
				}
				if isSeedCSVTextDataType(meta.Type) {
					continue // loader FORCE_NOT_NULL absorbs these
				}
				violations = append(violations, seedCSVNotNullEmptyViolation{
					bundle:   bundle,
					table:    entry.Table,
					csvFile:  entry.CSVFile,
					column:   col,
					dataType: meta.Type,
					rows:     n,
				})
			}
		}
	}
	return violations, nil
}

func parseSeedCSVColumnMeta(sql string) (map[string][]seedCSVColumnMeta, error) {
	withoutComments, err := stripSQLComments(sql)
	if err != nil {
		return nil, err
	}
	statements, err := splitSQLTopLevel(withoutComments, ';')
	if err != nil {
		return nil, err
	}

	out := make(map[string][]seedCSVColumnMeta)
	for statementIndex, statement := range statements {
		if match := createTableStatementPattern.FindStringSubmatchIndex(statement); match != nil {
			tableName := normalizeSeedCSVSQLIdentifier(statement[match[2]:match[3]])
			if _, exists := out[tableName]; exists {
				continue
			}
			openParen := match[1] - 1
			closeParen, err := findMatchingSQLParen(statement, openParen)
			if err != nil {
				return nil, fmt.Errorf("CREATE TABLE %s statement %d: %w", tableName, statementIndex+1, err)
			}
			cols, err := parseCreateTableColumnMeta(statement[openParen+1 : closeParen])
			if err != nil {
				return nil, fmt.Errorf("CREATE TABLE %s columns: %w", tableName, err)
			}
			out[tableName] = cols
			continue
		}

		if match := alterTableStatementPattern.FindStringSubmatchIndex(statement); match != nil {
			tableName := normalizeSeedCSVSQLIdentifier(statement[match[2]:match[3]])
			cols, exists := out[tableName]
			if !exists {
				// Ignore ALTERs for tables we do not track (same as partial applies).
				continue
			}
			clauses, err := splitSQLTopLevel(statement[match[1]:], ',')
			if err != nil {
				return nil, fmt.Errorf("ALTER TABLE %s: %w", tableName, err)
			}
			for _, clause := range clauses {
				if col, ifNotExists, isAdd, err := parseAddColumnMeta(clause); err != nil {
					return nil, err
				} else if isAdd {
					if containsColumnMeta(cols, col.Name) {
						if ifNotExists {
							continue
						}
						return nil, fmt.Errorf("ALTER TABLE %s duplicate column %s", tableName, col.Name)
					}
					cols = append(cols, col)
					continue
				}
				if name, notNull, ok := parseAlterColumnNotNull(clause); ok {
					for i := range cols {
						if cols[i].Name == name {
							cols[i].NotNull = notNull
						}
					}
				}
			}
			out[tableName] = cols
		}
	}
	return out, nil
}

func parseCreateTableColumnMeta(body string) ([]seedCSVColumnMeta, error) {
	definitions, err := splitSQLTopLevel(body, ',')
	if err != nil {
		return nil, err
	}
	var cols []seedCSVColumnMeta
	for index, definition := range definitions {
		identifier, quoted, ok, err := parseLeadingSQLIdentifier(definition)
		if err != nil {
			return nil, fmt.Errorf("definition %d: %w", index+1, err)
		}
		if !ok {
			continue
		}
		if !quoted && isTableConstraintKeyword(identifier) {
			continue
		}
		rest := strings.TrimSpace(definition)
		// strip leading identifier from rest
		_, _, restAfter, okID, err := seedCSVConsumeLeadingSQLIdentifier(rest)
		if err != nil || !okID {
			return nil, fmt.Errorf("definition %d: cannot re-consume identifier", index+1)
		}
		dataType := parseLeadingSQLType(restAfter)
		notNull := sqlHasNotNullAtDepthZero(restAfter) || sqlHasPrimaryKeyAtDepthZero(restAfter)
		cols = append(cols, seedCSVColumnMeta{
			Name:    identifier,
			Type:    dataType,
			NotNull: notNull,
		})
	}
	return cols, nil
}

func parseAddColumnMeta(clause string) (seedCSVColumnMeta, bool, bool, error) {
	rest, ok := consumeSQLKeyword(clause, "ADD")
	if !ok {
		return seedCSVColumnMeta{}, false, false, nil
	}
	explicitColumn := false
	if next, hasColumn := consumeSQLKeyword(rest, "COLUMN"); hasColumn {
		explicitColumn = true
		rest = next
	}
	ifNotExists := false
	if next, hasIf := consumeSQLKeyword(rest, "IF"); hasIf {
		next, hasNot := consumeSQLKeyword(next, "NOT")
		if !hasNot {
			return seedCSVColumnMeta{}, false, false, fmt.Errorf("ADD IF without NOT")
		}
		next, hasExists := consumeSQLKeyword(next, "EXISTS")
		if !hasExists {
			return seedCSVColumnMeta{}, false, false, fmt.Errorf("ADD IF NOT without EXISTS")
		}
		ifNotExists = true
		rest = next
	}
	identifier, quoted, restAfter, ok, err := seedCSVConsumeLeadingSQLIdentifier(rest)
	if err != nil {
		return seedCSVColumnMeta{}, false, false, err
	}
	if !ok {
		return seedCSVColumnMeta{}, false, false, fmt.Errorf("ADD clause has no target identifier")
	}
	if !explicitColumn && !quoted && isTableConstraintKeyword(identifier) {
		return seedCSVColumnMeta{}, false, false, nil
	}
	return seedCSVColumnMeta{
		Name:    identifier,
		Type:    parseLeadingSQLType(restAfter),
		NotNull: sqlHasNotNullAtDepthZero(restAfter) || sqlHasPrimaryKeyAtDepthZero(restAfter),
	}, ifNotExists, true, nil
}

func parseAlterColumnNotNull(clause string) (name string, notNull bool, ok bool) {
	rest, okKW := consumeSQLKeyword(clause, "ALTER")
	if !okKW {
		return "", false, false
	}
	rest, okKW = consumeSQLKeyword(rest, "COLUMN")
	if !okKW {
		return "", false, false
	}
	identifier, _, rest, okID, err := seedCSVConsumeLeadingSQLIdentifier(rest)
	if err != nil || !okID {
		return "", false, false
	}
	if hasKeywordSequence(rest, "SET", "NOT", "NULL") {
		return identifier, true, true
	}
	if hasKeywordSequence(rest, "DROP", "NOT", "NULL") {
		return identifier, false, true
	}
	return "", false, false
}

func hasKeywordSequence(input string, keywords ...string) bool {
	rest := input
	for _, kw := range keywords {
		next, ok := consumeSQLKeyword(rest, kw)
		if !ok {
			return false
		}
		rest = next
	}
	return true
}

func parseLeadingSQLType(rest string) string {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return ""
	}
	first, _, afterFirst, ok, err := seedCSVConsumeLeadingSQLIdentifier(rest)
	if err != nil || !ok {
		return ""
	}
	base := strings.ToLower(first)
	scan := afterFirst
	if base == "character" {
		if next, ok := consumeSQLKeyword(afterFirst, "VARYING"); ok {
			base = "character varying"
			scan = next
		}
	}
	if base == "double" {
		if next, ok := consumeSQLKeyword(afterFirst, "PRECISION"); ok {
			base = "double precision"
			scan = next
		}
	}
	scan = strings.TrimSpace(scan)
	if strings.HasPrefix(scan, "(") {
		end, err := findMatchingSQLParen(scan, 0)
		if err == nil {
			return base + strings.ToLower(scan[:end+1])
		}
	}
	return base
}

func sqlHasNotNullAtDepthZero(input string) bool {
	return sqlHasKeywordPairAtDepthZero(input, "NOT", "NULL")
}

func sqlHasPrimaryKeyAtDepthZero(input string) bool {
	return sqlHasKeywordPairAtDepthZero(input, "PRIMARY", "KEY")
}

func sqlHasKeywordPairAtDepthZero(input, first, second string) bool {
	upper := strings.ToUpper(input)
	depth := 0
	inSingle := false
	inDouble := false
	for i := 0; i < len(upper); i++ {
		c := upper[i]
		if inSingle {
			if c == '\'' {
				if i+1 < len(upper) && upper[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if c == '"' {
				if i+1 < len(upper) && upper[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}
		switch c {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth != 0 {
				continue
			}
			if !isSQLIdentStart(c) {
				continue
			}
			start := i
			for i < len(upper) && isSQLIdentCont(upper[i]) {
				i++
			}
			word := upper[start:i]
			i-- // loop will ++
			if word != first {
				continue
			}
			// skip whitespace
			j := i + 1
			for j < len(upper) && (upper[j] == ' ' || upper[j] == '\t' || upper[j] == '\n' || upper[j] == '\r') {
				j++
			}
			end := j
			for end < len(upper) && isSQLIdentCont(upper[end]) {
				end++
			}
			if strings.EqualFold(upper[j:end], second) {
				return true
			}
		}
	}
	return false
}

func isSQLIdentStart(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func isSQLIdentCont(c byte) bool {
	return isSQLIdentStart(c) || (c >= '0' && c <= '9')
}

// isSeedCSVTextDataType mirrors cmd/migrate textDataTypes (information_schema
// data_type values that FORCE_NOT_NULL can safely absorb). Keep these lists
// identical: a type allowed here but not forced by the loader would still
// fail COPY with 23502, while a type forced but not allowed here would hide
// real non-text defects. name/citext/domains are intentionally excluded until
// the loader also forces them.
func isSeedCSVTextDataType(dataType string) bool {
	d := strings.ToLower(strings.TrimSpace(dataType))
	switch {
	case d == "text":
		return true
	case d == "character varying" || strings.HasPrefix(d, "character varying("):
		return true
	case d == "character" || strings.HasPrefix(d, "character("):
		return true
	case d == "varchar" || strings.HasPrefix(d, "varchar("):
		// DDL parsers often emit varchar; information_schema uses character varying.
		return true
	case d == "char" || strings.HasPrefix(d, "char("):
		return true
	default:
		return false
	}
}

func containsColumnMeta(cols []seedCSVColumnMeta, name string) bool {
	for _, c := range cols {
		if c.Name == name {
			return true
		}
	}
	return false
}

// countUnquotedEmptyFieldsByColumn quote-aware scans a CSV and counts data rows
// where a field is empty and not quoted (Postgres COPY treats those as NULL).
func countUnquotedEmptyFieldsByColumn(path string) (map[string]int, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // repository-owned seed path
	if err != nil {
		return nil, err
	}
	text := string(raw)
	if strings.HasPrefix(text, "\ufeff") {
		text = text[1:]
	}
	type field struct {
		val    string
		quoted bool
	}
	var records [][]field
	i, n := 0, len(text)
	for i < n {
		var row []field
		for {
			if i >= n {
				row = append(row, field{"", false})
				break
			}
			if text[i] == '"' {
				i++
				var b strings.Builder
				for i < n {
					if text[i] == '"' {
						if i+1 < n && text[i+1] == '"' {
							b.WriteByte('"')
							i += 2
							continue
						}
						i++
						break
					}
					b.WriteByte(text[i])
					i++
				}
				row = append(row, field{b.String(), true})
			} else {
				start := i
				for i < n && text[i] != ',' && text[i] != '\n' && text[i] != '\r' {
					i++
				}
				row = append(row, field{text[start:i], false})
			}
			if i >= n {
				break
			}
			if text[i] == ',' {
				i++
				continue
			}
			if text[i] == '\r' || text[i] == '\n' {
				if text[i] == '\r' && i+1 < n && text[i+1] == '\n' {
					i += 2
				} else {
					i++
				}
				break
			}
		}
		records = append(records, row)
	}
	if len(records) > 0 {
		last := records[len(records)-1]
		if len(last) == 1 && last[0].val == "" && !last[0].quoted {
			records = records[:len(records)-1]
		}
	}
	out := map[string]int{}
	if len(records) == 0 {
		return out, nil
	}
	header := make([]string, len(records[0]))
	for j, f := range records[0] {
		header[j] = f.val
	}
	for _, row := range records[1:] {
		for j, col := range header {
			if j >= len(row) {
				continue
			}
			if row[j].val == "" && !row[j].quoted {
				out[col]++
			}
		}
	}
	return out, nil
}
