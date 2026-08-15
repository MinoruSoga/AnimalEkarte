package lintscan

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const erdTableCountMarker = "<!-- ERD:TABLE_COUNT=N -->"

var (
	sqlIdentifierPattern = `(?:"(?:[^"]|"")*"|[a-z_][a-z0-9_$]*)`
	createTablePattern   = regexp.MustCompile(
		`(?i)^[\t ]*CREATE[\t ]+TABLE[\t ]+(?:IF[\t ]+NOT[\t ]+EXISTS[\t ]+)?(` +
			sqlIdentifierPattern +
			`)(?:[\t ]*\.[\t ]*(` +
			sqlIdentifierPattern +
			`))?[\t ]*(?:\(|$)`,
	)
	erdTableCountMarkerPattern = regexp.MustCompile(
		`(?m)^[\t ]*<!--[ \t]+ERD:TABLE_COUNT=([0-9]+)[ \t]+-->[\t ]*\r?$`,
	)
)

func normalizeSQLIdentifier(identifier string) string {
	if strings.HasPrefix(identifier, `"`) && strings.HasSuffix(identifier, `"`) {
		return strings.ReplaceAll(identifier[1:len(identifier)-1], `""`, `"`)
	}
	return strings.ToLower(identifier)
}

func distinctCreateTableNames(sql string) []string {
	distinctNames := make(map[string]struct{})
	for _, line := range strings.Split(sql, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}

		match := createTablePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := normalizeSQLIdentifier(match[1])
		if match[2] != "" {
			name += "." + normalizeSQLIdentifier(match[2])
		}
		distinctNames[name] = struct{}{}
	}

	if len(distinctNames) == 0 {
		return nil
	}

	names := make([]string, 0, len(distinctNames))
	for name := range distinctNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func parseERDTableCountMarker(markdown string) (int, error) {
	matches := erdTableCountMarkerPattern.FindAllStringSubmatch(markdown, -1)
	if len(matches) != 1 {
		return 0, fmt.Errorf(
			"parse ERD table count marker %q: expected exactly one anchored marker, found %d",
			erdTableCountMarker,
			len(matches),
		)
	}

	count, err := strconv.Atoi(matches[0][1])
	if err != nil {
		return 0, fmt.Errorf("parse ERD table count marker value %q: %w", matches[0][1], err)
	}
	return count, nil
}

func reconcileERDTableCount(schemaCount, declaredCount int) []string {
	if schemaCount == declaredCount {
		return nil
	}
	return []string{
		fmt.Sprintf(
			"ERD table count drift: schema defines %d distinct table(s), marker declares %d",
			schemaCount,
			declaredCount,
		),
	}
}

func loadERDTableCountSources() (string, string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("get working directory: %w", err)
	}

	moduleRoot, err := FindModuleRoot(workingDirectory)
	if err != nil {
		return "", "", fmt.Errorf("find backend module root: %w", err)
	}

	return erdCountLoadSourcesFromModuleRoot(moduleRoot)
}

func erdCountLoadSourcesFromModuleRoot(moduleRoot string) (string, string, error) {
	schemaPath := filepath.Join(moduleRoot, "migrations", "001_init.sql")
	schemaDDL, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", "", fmt.Errorf("read schema DDL %s: %w", schemaPath, err)
	}

	erdPath := filepath.Join(moduleRoot, "..", "docs", "architecture", "erd.md")
	erdMarkdown, err := os.ReadFile(erdPath)
	if err != nil {
		return "", "", fmt.Errorf("read ERD %s: %w", erdPath, err)
	}

	return string(schemaDDL), string(erdMarkdown), nil
}

func TestERDTableCount_SourceLoading(t *testing.T) {
	tests := []struct {
		name            string
		setupModuleRoot func(t *testing.T) string
		wantErrContains string
	}{
		{
			name: "unreachable ERD path returns hard error",
			setupModuleRoot: func(t *testing.T) string {
				t.Helper()

				moduleRoot := filepath.Join(t.TempDir(), "backend")
				migrationsDirectory := filepath.Join(moduleRoot, "migrations")
				if err := os.MkdirAll(migrationsDirectory, 0o755); err != nil {
					t.Fatalf("create migrations fixture directory: %v", err)
				}
				if err := os.WriteFile(
					filepath.Join(migrationsDirectory, "001_init.sql"),
					[]byte("CREATE TABLE owners (id bigint);\n"),
					0o600,
				); err != nil {
					t.Fatalf("write schema fixture: %v", err)
				}
				return moduleRoot
			},
			wantErrContains: "read ERD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := erdCountLoadSourcesFromModuleRoot(tt.setupModuleRoot(t))
			if err == nil {
				t.Fatal("erdCountLoadSourcesFromModuleRoot() error = nil, want hard error")
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Fatalf(
					"erdCountLoadSourcesFromModuleRoot() error = %q, want substring %q",
					err,
					tt.wantErrContains,
				)
			}
		})
	}
}

func TestERDTableCount_MatchesSchema(t *testing.T) {
	schemaDDL, erdMarkdown, err := loadERDTableCountSources()
	if err != nil {
		t.Fatal(err)
	}

	schemaCount := len(distinctCreateTableNames(schemaDDL))
	declaredCount, err := parseERDTableCountMarker(erdMarkdown)
	if err != nil {
		t.Fatalf("parse %q marker: %v", erdTableCountMarker, err)
	}

	t.Logf("ERD table count: schema=%d declared=%d", schemaCount, declaredCount)
	if schemaCount != 123 {
		t.Errorf("distinct CREATE TABLE count = %d, want 123", schemaCount)
	}
	for _, violation := range reconcileERDTableCount(schemaCount, declaredCount) {
		t.Error(violation)
	}
}

func TestERDTableCount_ParserFixtures(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{
			name: "comment-only CREATE TABLE is ignored",
			sql:  "-- CREATE TABLE ghost_table (id bigint);\n",
			want: nil,
		},
		{
			name: "duplicate declarations count one distinct table",
			sql:  "CREATE TABLE pets (id bigint);\nCREATE TABLE pets (id bigint);\n",
			want: []string{"pets"},
		},
		{
			name: "IF NOT EXISTS declaration is recognized",
			sql:  "CREATE TABLE IF NOT EXISTS owners (id bigint);\n",
			want: []string{"owners"},
		},
		{
			name: "schema-qualified names remain distinct",
			sql:  "CREATE TABLE app_private.events (id bigint);\nCREATE TABLE public.events (id bigint);\n",
			want: []string{"app_private.events", "public.events"},
		},
		{
			name: "unquoted names are normalized to lowercase",
			sql:  "create table APP_PRIVATE.Audit_Events (id bigint);\n",
			want: []string{"app_private.audit_events"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := distinctCreateTableNames(tt.sql); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("distinctCreateTableNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestERDTableCount_MarkerUniqueness(t *testing.T) {
	tests := []struct {
		name      string
		markdown  string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "exactly one marker is accepted",
			markdown:  "<!-- ERD:TABLE_COUNT=110 -->\n| ERD domain physical table count | 109 |\n",
			wantCount: 110,
		},
		{
			name:     "missing marker is rejected",
			markdown: "# ERD\n",
			wantErr:  true,
		},
		{
			name:     "duplicate markers are rejected",
			markdown: "<!-- ERD:TABLE_COUNT=110 -->\n<!-- ERD:TABLE_COUNT=110 -->\n",
			wantErr:  true,
		},
		{
			name:     "inline marker text is not authoritative",
			markdown: "prose <!-- ERD:TABLE_COUNT=110 -->\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseERDTableCountMarker(tt.markdown)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseERDTableCountMarker() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseERDTableCountMarker() unexpected error: %v", err)
			}
			if got != tt.wantCount {
				t.Errorf("parseERDTableCountMarker() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestERDTableCount_GateDetectsViolations(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
	}{
		{
			name:     "shifted marker below schema count",
			markdown: "<!-- ERD:TABLE_COUNT=109 -->\n",
		},
		{
			name:     "shifted marker above schema count",
			markdown: "<!-- ERD:TABLE_COUNT=111 -->\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			declaredCount, err := parseERDTableCountMarker(tt.markdown)
			if err != nil {
				t.Fatalf("parse shifted ERD table count marker: %v", err)
			}

			if violations := reconcileERDTableCount(110, declaredCount); len(violations) == 0 {
				t.Fatalf(
					"reconcileERDTableCount(%d, %d) returned no violation for shifted ERD marker",
					110,
					declaredCount,
				)
			}
		})
	}
}
