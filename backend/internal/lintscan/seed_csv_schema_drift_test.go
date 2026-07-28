package lintscan

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const seedCSVSQLIdentifierPattern = `(?:"(?:[^"]|"")*"|[A-Za-z_][A-Za-z0-9_$]*)`

var (
	createTableStatementPattern = regexp.MustCompile(
		`(?is)^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(` +
			seedCSVSQLIdentifierPattern + `(?:\s*\.\s*` + seedCSVSQLIdentifierPattern + `)?)\s*\(`,
	)
	alterTableStatementPattern = regexp.MustCompile(
		`(?is)^\s*ALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:ONLY\s+)?(` +
			seedCSVSQLIdentifierPattern + `(?:\s*\.\s*` + seedCSVSQLIdentifierPattern + `)?)\s+`,
	)
	seedCSVMigrationFilenamePattern = regexp.MustCompile(`^([0-9]{3})_[A-Za-z0-9_]+\.sql$`)
	seedCSVColumnPositionPattern    = regexp.MustCompile(
		`(?is)(?:\s+(?:BEFORE|AFTER)\s+` + seedCSVSQLIdentifierPattern + `|\s+FIRST)\s*$`,
	)
)

var seedCSVBundleNames = []string{"002_master", "003_demo", "004_staging"}

type seedCSVSchemaTarget struct {
	bundle       string
	table        string
	csvFile      string
	tableColumns []string
	csvHeader    []string
}

type seedCSVSchemaViolation struct {
	bundle          string
	table           string
	csvFile         string
	expectedColumns []string
	actualColumns   []string
}

type seedCSVMigrationFile struct {
	number int
	name   string
}

func TestSeedCSVSchemaDrift_AllManifestHeadersMatchTableColumnOrder(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)

	scanned, violations, err := findSeedCSVSchemaDrift(moduleRoot)
	if err != nil {
		t.Fatalf("scan seed CSV schema drift from module root %s: %v", moduleRoot, err)
	}

	t.Logf("scanned %d seed CSV manifest entries", scanned)
	for _, violation := range violations {
		t.Errorf(
			"seed CSV schema drift in %s/%s for table %s: expected columns %v, got %v",
			violation.bundle,
			violation.csvFile,
			violation.table,
			violation.expectedColumns,
			violation.actualColumns,
		)
	}
}

func TestParseSeedCSVTableColumns(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want map[string][]string
	}{
		{
			name: "CREATE and ALTER columns preserve final order while constraints and comments are ignored",
			sql: `
				-- CREATE TABLE ignored_line_comment (bad text);
				/* CREATE TABLE ignored_block_comment (bad text); */
				CREATE TABLE app_private.example (
					id bigint PRIMARY KEY,
					amount numeric(10, 2),
					"quoted_column" text,
					CONSTRAINT example_pk PRIMARY KEY (id),
					CHECK (amount > 0)
				);
				ALTER TABLE app_private.example
					ADD CONSTRAINT example_unique UNIQUE (amount),
					ADD COLUMN IF NOT EXISTS added_after_create text,
					ALTER COLUMN amount TYPE numeric(12, 2);
				CREATE TABLE IF NOT EXISTS app_private.example (
					must_not_replace_existing_columns text
				);
			`,
			want: map[string][]string{
				"example": {"id", "amount", "quoted_column", "added_after_create"},
			},
		},
		{
			name: "multiple ADD COLUMN clauses append in statement order",
			sql: `
				CREATE TABLE treatments (id bigint);
				ALTER TABLE treatments
					ALTER COLUMN id TYPE bigint,
					ADD COLUMN dose_weight_kg numeric(6, 2),
					ADD COLUMN dose_weight_source varchar(255),
					ADD COLUMN dose_amount_mg numeric(12, 6);
			`,
			want: map[string][]string{
				"treatments": {"id", "dose_weight_kg", "dose_weight_source", "dose_amount_mg"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSeedCSVTableColumns(tt.sql)
			if err != nil {
				t.Fatalf("parse fixture DDL: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parsed table columns = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestFindSeedCSVSchemaDrift_UsesOrderedMigrations(t *testing.T) {
	tests := []struct {
		name           string
		migrationFiles []struct {
			name string
			sql  string
		}
		wantScanned int
	}{
		{
			name: "001 and later migrations append pets version in migration number order",
			migrationFiles: []struct {
				name string
				sql  string
			}{
				{
					name: "003_add_after_version.sql",
					sql:  "ALTER TABLE pets ADD COLUMN after_version text;",
				},
				{
					name: "001_init.sql",
					sql:  "CREATE TABLE pets (id bigint PRIMARY KEY);",
				},
				{
					name: "002_add_pets_version.sql",
					sql:  "ALTER TABLE pets ADD COLUMN version integer NOT NULL DEFAULT 1;",
				},
			},
			wantScanned: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleRoot := t.TempDir()
			migrationsDir := filepath.Join(moduleRoot, "migrations")
			if err := os.MkdirAll(migrationsDir, 0o750); err != nil {
				t.Fatalf("create migrations directory: %v", err)
			}

			for _, migration := range tt.migrationFiles {
				path := filepath.Join(migrationsDir, migration.name)
				if err := os.WriteFile(path, []byte(migration.sql), 0o600); err != nil {
					t.Fatalf("write migration fixture %s: %v", migration.name, err)
				}
			}

			for _, bundle := range seedCSVBundleNames {
				bundleDir := filepath.Join(migrationsDir, "seeds", bundle)
				if err := os.MkdirAll(bundleDir, 0o750); err != nil {
					t.Fatalf("create seed bundle directory %s: %v", bundle, err)
				}
				manifest := `{"tables":[{"table":"pets","csvFile":"pets.csv"}]}`
				if err := os.WriteFile(
					filepath.Join(bundleDir, "manifest.json"),
					[]byte(manifest),
					0o600,
				); err != nil {
					t.Fatalf("write seed bundle manifest %s: %v", bundle, err)
				}
				if err := os.WriteFile(
					filepath.Join(bundleDir, "pets.csv"),
					[]byte("id,version,after_version\n"),
					0o600,
				); err != nil {
					t.Fatalf("write seed bundle CSV %s: %v", bundle, err)
				}
			}

			scanned, violations, err := findSeedCSVSchemaDrift(moduleRoot)
			if err != nil {
				t.Fatalf("scan ordered migration fixture: %v", err)
			}
			if scanned != tt.wantScanned {
				t.Fatalf("scanned manifest entries = %d, want %d", scanned, tt.wantScanned)
			}
			if len(violations) != 0 {
				t.Fatalf("ordered migration fixture violations = %v, want none", violations)
			}
			t.Log("pets final columns: [id version after_version]")
		})
	}
}

func TestSeedCSVSchemaDrift_OrderedMigrationsAppendPetsVersion(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		wantTail []string
	}{
		{
			name:     "pets version follows danger reason",
			table:    "pets",
			wantTail: []string{"danger_reason", "version"},
		},
	}

	moduleRoot := mustFindSeedCSVModuleRoot(t)
	ddl, _, err := seedCSVReadOrderedMigrationDDL(moduleRoot)
	if err != nil {
		t.Fatalf("read ordered migrations: %v", err)
	}
	tableColumns, err := parseSeedCSVTableColumns(ddl)
	if err != nil {
		t.Fatalf("parse ordered migrations: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns, ok := tableColumns[tt.table]
			if !ok {
				t.Fatalf("ordered migrations have no table %s", tt.table)
			}
			if len(columns) < len(tt.wantTail) {
				t.Fatalf("table %s columns = %v, want tail %v", tt.table, columns, tt.wantTail)
			}
			gotTail := columns[len(columns)-len(tt.wantTail):]
			if !reflect.DeepEqual(gotTail, tt.wantTail) {
				t.Fatalf("table %s column tail = %v, want %v", tt.table, gotTail, tt.wantTail)
			}
			t.Logf("%s final column tail: %v", tt.table, gotTail)
		})
	}
}

func TestSeedCSVSchemaDrift_RejectsUnsupportedColumnOrderDDL(t *testing.T) {
	tests := []struct {
		name    string
		alter   string
		wantErr string
	}{
		{
			name:    "DROP COLUMN fails explicitly",
			alter:   "ALTER TABLE pets DROP COLUMN legacy_code;",
			wantErr: "unsupported ALTER TABLE pets",
		},
		{
			name:    "RENAME COLUMN fails explicitly",
			alter:   "ALTER TABLE pets RENAME COLUMN legacy_code TO current_code;",
			wantErr: "unsupported ALTER TABLE pets",
		},
		{
			name:    "column position modifier fails explicitly",
			alter:   "ALTER TABLE pets ADD COLUMN legacy_code text BEFORE id;",
			wantErr: "unsupported ALTER TABLE pets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSeedCSVTableColumns(
				"CREATE TABLE pets (id bigint);" + tt.alter,
			)
			if err == nil {
				t.Fatal("parse unsupported ALTER DDL: got nil error, want explicit failure")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("parse unsupported ALTER DDL error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestSeedCSVSchemaDrift_GateDetectsViolations(t *testing.T) {
	tests := []struct {
		name           string
		target         seedCSVSchemaTarget
		wantViolations int
	}{
		{
			name: "swapped CSV columns are rejected",
			target: seedCSVSchemaTarget{
				bundle:       "002_master",
				table:        "animal_species",
				csvFile:      "animal_species.csv",
				tableColumns: []string{"id", "name", "is_active", "sort_order", "created_at", "updated_at"},
				csvHeader:    []string{"name", "id", "is_active", "sort_order", "created_at", "updated_at"},
			},
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := analyzeSeedCSVSchemaTargets([]seedCSVSchemaTarget{tt.target})
			if len(violations) != tt.wantViolations {
				t.Fatalf(
					"analyze swapped columns for %s/%s: got %d violation(s), want %d",
					tt.target.bundle,
					tt.target.csvFile,
					len(violations),
					tt.wantViolations,
				)
			}
		})
	}
}

func mustFindSeedCSVModuleRoot(t *testing.T) string {
	t.Helper()

	moduleRoot, err := findSeedCSVModuleRoot()
	if err != nil {
		t.Fatalf("resolve backend module root for seed CSV schema drift gate: %v", err)
	}
	return moduleRoot
}

func findSeedCSVModuleRoot() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	moduleRoot, err := FindModuleRoot(workingDirectory)
	if err != nil {
		return "", fmt.Errorf("find module root from %s: %w", workingDirectory, err)
	}
	return moduleRoot, nil
}

func findSeedCSVSchemaDrift(moduleRoot string) (int, []seedCSVSchemaViolation, error) {
	ddl, migrationNames, err := seedCSVReadOrderedMigrationDDL(moduleRoot)
	if err != nil {
		return 0, nil, err
	}

	tableColumns, err := parseSeedCSVTableColumns(ddl)
	if err != nil {
		return 0, nil, fmt.Errorf(
			"parse ordered migration schema %v: %w",
			migrationNames,
			err,
		)
	}

	targets, err := loadSeedCSVSchemaTargets(moduleRoot, tableColumns)
	if err != nil {
		return 0, nil, fmt.Errorf("load seed CSV schema targets: %w", err)
	}

	return len(targets), analyzeSeedCSVSchemaTargets(targets), nil
}

func seedCSVReadOrderedMigrationDDL(moduleRoot string) (string, []string, error) {
	migrationsDir := filepath.Join(moduleRoot, "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", nil, fmt.Errorf("read migrations directory %s: %w", migrationsDir, err)
	}

	var migrations []seedCSVMigrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := seedCSVMigrationFilenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		number, err := strconv.Atoi(match[1])
		if err != nil {
			return "", nil, fmt.Errorf("parse migration number from %s: %w", entry.Name(), err)
		}
		migrations = append(migrations, seedCSVMigrationFile{number: number, name: entry.Name()})
	}
	if len(migrations) == 0 {
		return "", nil, fmt.Errorf("no NNN_*.sql migrations found in %s", migrationsDir)
	}

	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].number == migrations[j].number {
			return migrations[i].name < migrations[j].name
		}
		return migrations[i].number < migrations[j].number
	})
	if migrations[0].number != 1 {
		return "", nil, fmt.Errorf(
			"ordered migration set in %s starts at %03d, want 001",
			migrationsDir,
			migrations[0].number,
		)
	}

	var ddl strings.Builder
	migrationNames := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		path := filepath.Join(migrationsDir, migration.name)
		raw, err := os.ReadFile(path) //nolint:gosec // path is a repository-owned NNN_*.sql migration.
		if err != nil {
			return "", nil, fmt.Errorf("read migration %s: %w", path, err)
		}
		ddl.Write(raw)
		ddl.WriteByte('\n')
		migrationNames = append(migrationNames, migration.name)
	}
	return ddl.String(), migrationNames, nil
}

func loadSeedCSVSchemaTargets(
	moduleRoot string,
	tableColumns map[string][]string,
) ([]seedCSVSchemaTarget, error) {
	var targets []seedCSVSchemaTarget
	for _, bundle := range seedCSVBundleNames {
		bundleDir := filepath.Join(moduleRoot, "migrations", "seeds", bundle)
		manifestPath := filepath.Join(bundleDir, "manifest.json")

		manifest, err := readSeedCSVManifest(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("load bundle %s manifest: %w", bundle, err)
		}
		if len(manifest.Tables) == 0 {
			return nil, fmt.Errorf("bundle %s manifest %s contains no table entries", bundle, manifestPath)
		}

		for index, entry := range manifest.Tables {
			tableName := normalizeSeedCSVSQLIdentifier(entry.Table)
			expectedColumns, ok := tableColumns[tableName]
			if !ok {
				return nil, fmt.Errorf(
					"bundle %s manifest entry %d references table %q absent from ordered migration schema",
					bundle,
					index,
					entry.Table,
				)
			}

			csvPath := filepath.Join(bundleDir, entry.CSVFile)
			header, err := readSeedCSVHeader(csvPath)
			if err != nil {
				return nil, fmt.Errorf(
					"read bundle %s manifest entry %d table %s CSV %s: %w",
					bundle,
					index,
					entry.Table,
					entry.CSVFile,
					err,
				)
			}

			targets = append(targets, seedCSVSchemaTarget{
				bundle:       bundle,
				table:        entry.Table,
				csvFile:      entry.CSVFile,
				tableColumns: append([]string(nil), expectedColumns...),
				csvHeader:    header,
			})
		}
	}
	return targets, nil
}

type seedCSVManifest struct {
	Tables []seedCSVManifestEntry `json:"tables"`
}

type seedCSVManifestEntry struct {
	Table   string `json:"table"`
	CSVFile string `json:"csvFile"`
}

func readSeedCSVManifest(path string) (seedCSVManifest, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // path is derived from the fixed bundle allowlist.
	if err != nil {
		return seedCSVManifest{}, fmt.Errorf("read manifest %s: %w", path, err)
	}

	var manifest seedCSVManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return seedCSVManifest{}, fmt.Errorf("decode manifest %s: %w", path, err)
	}
	return manifest, nil
}

func readSeedCSVHeader(path string) ([]string, error) {
	file, err := os.Open(path) //nolint:gosec // path is declared by a repository-owned manifest.
	if err != nil {
		return nil, fmt.Errorf("open CSV %s: %w", path, err)
	}

	header, readErr := csv.NewReader(file).Read()
	closeErr := file.Close()
	if readErr != nil {
		if readErr == io.EOF {
			return nil, fmt.Errorf("read CSV header %s: file is empty", path)
		}
		return nil, fmt.Errorf("read CSV header %s: %w", path, readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close CSV %s: %w", path, closeErr)
	}
	return header, nil
}

func analyzeSeedCSVSchemaTargets(targets []seedCSVSchemaTarget) []seedCSVSchemaViolation {
	var violations []seedCSVSchemaViolation
	for _, target := range targets {
		if reflect.DeepEqual(target.tableColumns, target.csvHeader) {
			continue
		}
		violations = append(violations, seedCSVSchemaViolation{
			bundle:          target.bundle,
			table:           target.table,
			csvFile:         target.csvFile,
			expectedColumns: append([]string(nil), target.tableColumns...),
			actualColumns:   append([]string(nil), target.csvHeader...),
		})
	}
	return violations
}

func parseSeedCSVTableColumns(sql string) (map[string][]string, error) {
	withoutComments, err := stripSQLComments(sql)
	if err != nil {
		return nil, fmt.Errorf("strip SQL comments: %w", err)
	}

	statements, err := splitSQLTopLevel(withoutComments, ';')
	if err != nil {
		return nil, fmt.Errorf("split SQL statements: %w", err)
	}

	tableColumns := make(map[string][]string)
	for statementIndex, statement := range statements {
		if match := createTableStatementPattern.FindStringSubmatchIndex(statement); match != nil {
			tableName := normalizeSeedCSVSQLIdentifier(statement[match[2]:match[3]])
			if _, exists := tableColumns[tableName]; exists {
				continue
			}

			openParen := match[1] - 1
			closeParen, err := findMatchingSQLParen(statement, openParen)
			if err != nil {
				return nil, fmt.Errorf(
					"parse CREATE TABLE %s at statement %d: %w",
					tableName,
					statementIndex+1,
					err,
				)
			}
			columns, err := parseCreateTableColumns(statement[openParen+1 : closeParen])
			if err != nil {
				return nil, fmt.Errorf(
					"parse columns for CREATE TABLE %s at statement %d: %w",
					tableName,
					statementIndex+1,
					err,
				)
			}
			tableColumns[tableName] = columns
			continue
		}

		if match := alterTableStatementPattern.FindStringSubmatchIndex(statement); match != nil {
			tableName := normalizeSeedCSVSQLIdentifier(statement[match[2]:match[3]])
			columns, exists := tableColumns[tableName]
			if !exists {
				return nil, fmt.Errorf(
					"ALTER TABLE %s at statement %d appears before its CREATE TABLE",
					tableName,
					statementIndex+1,
				)
			}

			clauses, err := splitSQLTopLevel(statement[match[1]:], ',')
			if err != nil {
				return nil, fmt.Errorf(
					"split ALTER TABLE %s clauses at statement %d: %w",
					tableName,
					statementIndex+1,
					err,
				)
			}
			for clauseIndex, clause := range clauses {
				if err := seedCSVValidateAlterTableClause(clause); err != nil {
					return nil, fmt.Errorf(
						"unsupported ALTER TABLE %s clause %d at statement %d: %w",
						tableName,
						clauseIndex+1,
						statementIndex+1,
						err,
					)
				}
				column, ifNotExists, isAddColumn, err := parseAddColumnClause(clause)
				if err != nil {
					return nil, fmt.Errorf(
						"parse ALTER TABLE %s clause %d at statement %d: %w",
						tableName,
						clauseIndex+1,
						statementIndex+1,
						err,
					)
				}
				if !isAddColumn {
					continue
				}
				if containsString(columns, column) {
					if ifNotExists {
						continue
					}
					return nil, fmt.Errorf(
						"ALTER TABLE %s adds duplicate column %s at statement %d",
						tableName,
						column,
						statementIndex+1,
					)
				}
				columns = append(columns, column)
			}
			tableColumns[tableName] = columns
		}
	}
	return tableColumns, nil
}

func seedCSVValidateAlterTableClause(clause string) error {
	if _, ok := consumeSQLKeyword(clause, "ADD"); ok {
		if seedCSVColumnPositionPattern.MatchString(clause) {
			return fmt.Errorf("ADD COLUMN position modifier can change column order")
		}
		return nil
	}

	if rest, ok := consumeSQLKeyword(clause, "DROP"); ok {
		if _, isConstraint := consumeSQLKeyword(rest, "CONSTRAINT"); isConstraint {
			return nil
		}
		return fmt.Errorf("DROP operation can change the expected column set or order")
	}

	if rest, ok := consumeSQLKeyword(clause, "RENAME"); ok {
		if _, isConstraint := consumeSQLKeyword(rest, "CONSTRAINT"); isConstraint {
			return nil
		}
		return fmt.Errorf("RENAME operation can change the expected column set or order")
	}

	rest, ok := consumeSQLKeyword(clause, "ALTER")
	if !ok {
		return fmt.Errorf("unrecognized ALTER TABLE operation")
	}
	rest, ok = consumeSQLKeyword(rest, "COLUMN")
	if !ok {
		return fmt.Errorf("ALTER operation without COLUMN is not supported")
	}

	_, _, rest, ok, err := seedCSVConsumeLeadingSQLIdentifier(rest)
	if err != nil {
		return fmt.Errorf("parse ALTER COLUMN target: %w", err)
	}
	if !ok {
		return fmt.Errorf("ALTER COLUMN has no target identifier")
	}

	if _, isType := consumeSQLKeyword(rest, "TYPE"); isType {
		return nil
	}
	if rest, isSet := consumeSQLKeyword(rest, "SET"); isSet {
		if _, isNot := consumeSQLKeyword(rest, "NOT"); isNot {
			return nil
		}
		if _, isDefault := consumeSQLKeyword(rest, "DEFAULT"); isDefault {
			return nil
		}
	}
	if rest, isDrop := consumeSQLKeyword(rest, "DROP"); isDrop {
		if _, isNot := consumeSQLKeyword(rest, "NOT"); isNot {
			return nil
		}
		if _, isDefault := consumeSQLKeyword(rest, "DEFAULT"); isDefault {
			return nil
		}
	}
	return fmt.Errorf("ALTER COLUMN operation is not known to preserve column order")
}

func seedCSVConsumeLeadingSQLIdentifier(
	input string,
) (string, bool, string, bool, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false, "", false, nil
	}

	if trimmed[0] == '"' {
		var identifier strings.Builder
		for index := 1; index < len(trimmed); index++ {
			if trimmed[index] != '"' {
				identifier.WriteByte(trimmed[index])
				continue
			}
			if index+1 < len(trimmed) && trimmed[index+1] == '"' {
				identifier.WriteByte('"')
				index++
				continue
			}
			return identifier.String(), true, strings.TrimSpace(trimmed[index+1:]), true, nil
		}
		return "", false, "", false, fmt.Errorf("unterminated quoted identifier")
	}

	end := 0
	for end < len(trimmed) && isSQLIdentifierByte(trimmed[end]) {
		end++
	}
	if end == 0 {
		return "", false, "", false, nil
	}
	return strings.ToLower(trimmed[:end]), false, strings.TrimSpace(trimmed[end:]), true, nil
}

func parseCreateTableColumns(body string) ([]string, error) {
	definitions, err := splitSQLTopLevel(body, ',')
	if err != nil {
		return nil, fmt.Errorf("split CREATE TABLE definitions: %w", err)
	}

	var columns []string
	for index, definition := range definitions {
		identifier, quoted, ok, err := parseLeadingSQLIdentifier(definition)
		if err != nil {
			return nil, fmt.Errorf("parse definition %d: %w", index+1, err)
		}
		if !ok {
			continue
		}
		if !quoted && isTableConstraintKeyword(identifier) {
			continue
		}
		columns = append(columns, identifier)
	}
	return columns, nil
}

func parseAddColumnClause(clause string) (string, bool, bool, error) {
	rest, ok := consumeSQLKeyword(clause, "ADD")
	if !ok {
		return "", false, false, nil
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
			return "", false, false, fmt.Errorf("ADD IF is not followed by NOT")
		}
		next, hasExists := consumeSQLKeyword(next, "EXISTS")
		if !hasExists {
			return "", false, false, fmt.Errorf("ADD IF NOT is not followed by EXISTS")
		}
		ifNotExists = true
		rest = next
	}

	identifier, quoted, ok, err := parseLeadingSQLIdentifier(rest)
	if err != nil {
		return "", false, false, fmt.Errorf("parse ADD target: %w", err)
	}
	if !ok {
		return "", false, false, fmt.Errorf("ADD clause has no target identifier")
	}
	if !explicitColumn && !quoted && isTableConstraintKeyword(identifier) {
		return "", false, false, nil
	}
	return identifier, ifNotExists, true, nil
}

func consumeSQLKeyword(input, keyword string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if len(trimmed) < len(keyword) || !strings.EqualFold(trimmed[:len(keyword)], keyword) {
		return input, false
	}
	if len(trimmed) > len(keyword) && isSQLIdentifierByte(trimmed[len(keyword)]) {
		return input, false
	}
	return strings.TrimSpace(trimmed[len(keyword):]), true
}

func parseLeadingSQLIdentifier(input string) (string, bool, bool, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false, false, nil
	}

	if trimmed[0] == '"' {
		var identifier strings.Builder
		for index := 1; index < len(trimmed); index++ {
			if trimmed[index] != '"' {
				identifier.WriteByte(trimmed[index])
				continue
			}
			if index+1 < len(trimmed) && trimmed[index+1] == '"' {
				identifier.WriteByte('"')
				index++
				continue
			}
			return identifier.String(), true, true, nil
		}
		return "", false, false, fmt.Errorf("unterminated quoted identifier")
	}

	end := 0
	for end < len(trimmed) && isSQLIdentifierByte(trimmed[end]) {
		end++
	}
	if end == 0 {
		return "", false, false, nil
	}
	return strings.ToLower(trimmed[:end]), false, true, nil
}

func normalizeSeedCSVSQLIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	lastDot := -1
	inQuotes := false
	for index := 0; index < len(identifier); index++ {
		switch identifier[index] {
		case '"':
			if inQuotes && index+1 < len(identifier) && identifier[index+1] == '"' {
				index++
				continue
			}
			inQuotes = !inQuotes
		case '.':
			if !inQuotes {
				lastDot = index
			}
		}
	}
	if lastDot >= 0 {
		identifier = strings.TrimSpace(identifier[lastDot+1:])
	}
	if len(identifier) >= 2 && identifier[0] == '"' && identifier[len(identifier)-1] == '"' {
		return strings.ReplaceAll(identifier[1:len(identifier)-1], `""`, `"`)
	}
	return strings.ToLower(identifier)
}

func isTableConstraintKeyword(identifier string) bool {
	switch strings.ToUpper(identifier) {
	case "CHECK", "CONSTRAINT", "EXCLUDE", "FOREIGN", "LIKE", "PRIMARY", "UNIQUE":
		return true
	default:
		return false
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isSQLIdentifierByte(value byte) bool {
	return value == '_' || value == '$' ||
		value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9'
}

func stripSQLComments(sql string) (string, error) {
	var result strings.Builder
	result.Grow(len(sql))

	for index := 0; index < len(sql); {
		switch {
		case index+1 < len(sql) && sql[index:index+2] == "--":
			result.WriteString("  ")
			index += 2
			for index < len(sql) && sql[index] != '\n' {
				result.WriteByte(' ')
				index++
			}
		case index+1 < len(sql) && sql[index:index+2] == "/*":
			depth := 1
			result.WriteString("  ")
			index += 2
			for index < len(sql) && depth > 0 {
				switch {
				case index+1 < len(sql) && sql[index:index+2] == "/*":
					depth++
					result.WriteString("  ")
					index += 2
				case index+1 < len(sql) && sql[index:index+2] == "*/":
					depth--
					result.WriteString("  ")
					index += 2
				default:
					if sql[index] == '\n' {
						result.WriteByte('\n')
					} else {
						result.WriteByte(' ')
					}
					index++
				}
			}
			if depth != 0 {
				return "", fmt.Errorf("unterminated block comment")
			}
		case sql[index] == '\'' || sql[index] == '"':
			next, err := copySQLQuoted(sql, index, &result, sql[index])
			if err != nil {
				return "", err
			}
			index = next
		default:
			if tag, ok := sqlDollarQuoteTag(sql, index); ok {
				next, err := copySQLDollarQuoted(sql, index, tag, &result)
				if err != nil {
					return "", err
				}
				index = next
				continue
			}
			result.WriteByte(sql[index])
			index++
		}
	}
	return result.String(), nil
}

func copySQLQuoted(sql string, start int, result *strings.Builder, quote byte) (int, error) {
	result.WriteByte(quote)
	for index := start + 1; index < len(sql); index++ {
		result.WriteByte(sql[index])
		if sql[index] == '\\' && index+1 < len(sql) {
			index++
			result.WriteByte(sql[index])
			continue
		}
		if sql[index] != quote {
			continue
		}
		if index+1 < len(sql) && sql[index+1] == quote {
			index++
			result.WriteByte(sql[index])
			continue
		}
		return index + 1, nil
	}
	return 0, fmt.Errorf("unterminated %q quoted SQL text", quote)
}

func sqlDollarQuoteTag(sql string, start int) (string, bool) {
	if start >= len(sql) || sql[start] != '$' {
		return "", false
	}
	for index := start + 1; index < len(sql); index++ {
		switch {
		case sql[index] == '$':
			return sql[start : index+1], true
		case isSQLIdentifierByte(sql[index]):
			continue
		default:
			return "", false
		}
	}
	return "", false
}

func copySQLDollarQuoted(
	sql string,
	start int,
	tag string,
	result *strings.Builder,
) (int, error) {
	endOffset := strings.Index(sql[start+len(tag):], tag)
	if endOffset < 0 {
		return 0, fmt.Errorf("unterminated dollar-quoted SQL text with tag %s", tag)
	}
	end := start + len(tag) + endOffset + len(tag)
	result.WriteString(sql[start:end])
	return end, nil
}

func splitSQLTopLevel(input string, delimiter byte) ([]string, error) {
	var parts []string
	start := 0
	depth := 0

	for index := 0; index < len(input); {
		switch input[index] {
		case '\'', '"':
			next, err := skipSQLQuoted(input, index, input[index])
			if err != nil {
				return nil, err
			}
			index = next
		case '$':
			tag, ok := sqlDollarQuoteTag(input, index)
			if !ok {
				index++
				continue
			}
			next, err := skipSQLDollarQuoted(input, index, tag)
			if err != nil {
				return nil, err
			}
			index = next
		case '(':
			depth++
			index++
		case ')':
			if depth == 0 {
				return nil, fmt.Errorf("unexpected closing parenthesis at byte %d", index)
			}
			depth--
			index++
		default:
			if input[index] == delimiter && depth == 0 {
				if part := strings.TrimSpace(input[start:index]); part != "" {
					parts = append(parts, part)
				}
				start = index + 1
			}
			index++
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unclosed parenthesis depth %d", depth)
	}
	if part := strings.TrimSpace(input[start:]); part != "" {
		parts = append(parts, part)
	}
	return parts, nil
}

func findMatchingSQLParen(input string, open int) (int, error) {
	if open < 0 || open >= len(input) || input[open] != '(' {
		return 0, fmt.Errorf("byte %d is not an opening parenthesis", open)
	}

	depth := 0
	for index := open; index < len(input); {
		switch input[index] {
		case '\'', '"':
			next, err := skipSQLQuoted(input, index, input[index])
			if err != nil {
				return 0, err
			}
			index = next
		case '$':
			tag, ok := sqlDollarQuoteTag(input, index)
			if !ok {
				index++
				continue
			}
			next, err := skipSQLDollarQuoted(input, index, tag)
			if err != nil {
				return 0, err
			}
			index = next
		case '(':
			depth++
			index++
		case ')':
			depth--
			if depth == 0 {
				return index, nil
			}
			index++
		default:
			index++
		}
	}
	return 0, fmt.Errorf("opening parenthesis at byte %d has no match", open)
}

func skipSQLQuoted(input string, start int, quote byte) (int, error) {
	var discard strings.Builder
	return copySQLQuoted(input, start, &discard, quote)
}

func skipSQLDollarQuoted(input string, start int, tag string) (int, error) {
	var discard strings.Builder
	return copySQLDollarQuoted(input, start, tag, &discard)
}
