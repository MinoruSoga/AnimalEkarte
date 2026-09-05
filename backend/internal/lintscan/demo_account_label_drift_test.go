package lintscan

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

const (
	demoAccountLabelDriftBundle        = "002_master"
	demoAccountExpectedMinCount        = 39
	demoAccountExpectedMaxCount        = 42
	gitLFSPointerPrefix                = "version https://git-lfs.github.com/spec/v1"
	demoAccountSystemAdminClinicLabel  = "全医院"
	demoAccountLoginFormRelativePath   = "frontend/src/features/auth/components/LoginForm.tsx"
	demoAccountStaffAttachEmailPattern = `^stg-staff-\d+@example\.test$`
)

var demoAccountComparedSeedTables = []string{
	"permission_groups",
	"clinics",
}

// demoAccountObjectLinePattern extracts one DEMO_ACCOUNTS object. Prettier may
// wrap fields across lines; field order matches the current source (email,
// labels, optional isSystemAdmin).
var demoAccountObjectLinePattern = regexp.MustCompile(
	`(?s)email:\s*"([^"]+)"` +
		`.*?occupationLabel:\s*"([^"]+)"` +
		`.*?permissionLabel:\s*"([^"]+)"` +
		`.*?clinicLabel:\s*"([^"]+)"` +
		`(?:.*?isSystemAdmin:\s*(true|false))?`,
)

var (
	demoAccountStaffAttachEmailRE = regexp.MustCompile(demoAccountStaffAttachEmailPattern)
	demoAccountRetiredDemoEmails  = []string{
		"hayashi@noah-vet.co.jp",
		"admin@noavet.jp",
		"admin@example.com",
		"vet@example.com",
		"nurse@example.com",
		"reception@example.com",
		"trimmer@example.com",
		"joto-vet@example.com",
		"shiki-vet@example.com",
	}
)

type demoAccountUILabels struct {
	email           string
	occupationLabel string
	permissionLabel string
	clinicLabel     string
	isSystemAdmin   bool
}

type demoAccountSeedLabels struct {
	occupationLabel string
	permissionLabel string
	clinicLabel     string
	isSystemAdmin   bool
}

type demoAccountSeedCSVTable struct {
	bundle  string
	table   string
	csvFile string
	header  []string
	rows    [][]string
	col     map[string]int
}

func TestDemoAccountLoginFormMatchesStaffAttachContract(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)

	source, err := os.ReadFile(mustFindDemoAccountLoginFormPath(t, moduleRoot))
	if err != nil {
		t.Fatalf("read LoginForm.tsx: %v", err)
	}

	uiAccounts := parseDemoAccountObjectLines(string(source))
	if len(uiAccounts) < demoAccountExpectedMinCount || len(uiAccounts) > demoAccountExpectedMaxCount {
		t.Fatalf(
			"DEMO_ACCOUNTS count = %d, want %d..%d staff-attach accounts",
			len(uiAccounts),
			demoAccountExpectedMinCount,
			demoAccountExpectedMaxCount,
		)
	}

	tables, err := loadDemoAccountComparedSeedTables(moduleRoot)
	if err != nil {
		t.Fatalf("load %s compared seed CSV: %v", demoAccountLabelDriftBundle, err)
	}

	for _, violation := range demoAccountStaffAttachContractViolations(uiAccounts, tables) {
		t.Errorf("%s", violation)
	}
}

func TestDemoAccountLoginFormMatchesSeedloginCatalog(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)
	source, err := os.ReadFile(mustFindDemoAccountLoginFormPath(t, moduleRoot))
	if err != nil {
		t.Fatalf("read LoginForm.tsx: %v", err)
	}

	uiAccounts := parseDemoAccountObjectLines(string(source))
	catalog := seedlogin.Catalog()
	if len(uiAccounts) != len(catalog) {
		t.Fatalf("DEMO_ACCOUNTS count = %d, seedlogin catalog = %d", len(uiAccounts), len(catalog))
	}
	for i, spec := range catalog {
		ui := uiAccounts[i]
		if ui.email != spec.Email {
			t.Errorf("index %d email: ui=%s catalog=%s", i, ui.email, spec.Email)
		}
		if ui.occupationLabel != spec.OccupationLabel {
			t.Errorf("%s occupationLabel: ui=%q catalog=%q", spec.Email, ui.occupationLabel, spec.OccupationLabel)
		}
		if ui.permissionLabel != seedlogin.PermissionGroupName {
			t.Errorf("%s permissionLabel: ui=%q want %q", spec.Email, ui.permissionLabel, seedlogin.PermissionGroupName)
		}
		if ui.clinicLabel != spec.ClinicLabel {
			t.Errorf("%s clinicLabel: ui=%q catalog=%q", spec.Email, ui.clinicLabel, spec.ClinicLabel)
		}
		if ui.isSystemAdmin {
			t.Errorf("%s must not be system admin", spec.Email)
		}
	}
}

func TestParseDemoAccountObjectLines(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []demoAccountUILabels
	}{
		{
			name: "empty source yields no accounts",
			src:  "",
			want: nil,
		},
		{
			name: "source with no DEMO_ACCOUNTS objects yields no accounts",
			src:  "const DEMO_ACCOUNTS = SHOW_DEMO ? [] : [];\n",
			want: nil,
		},
		{
			name: "system admin line captures isSystemAdmin true",
			src:  `{ email: "hayashi@noah-vet.co.jp", displayName: "林 文明", occupationLabel: "獣医師", permissionLabel: "執行", clinicLabel: "全医院", isSystemAdmin: true },`,
			want: []demoAccountUILabels{{
				email:           "hayashi@noah-vet.co.jp",
				occupationLabel: "獣医師",
				permissionLabel: "執行",
				clinicLabel:     "全医院",
				isSystemAdmin:   true,
			}},
		},
		{
			name: "omitted isSystemAdmin is false",
			src:  `{ email: "admin@example.com", displayName: "安田 希恵", occupationLabel: "看護師", permissionLabel: "執行", clinicLabel: "八王子病院" },`,
			want: []demoAccountUILabels{{
				email:           "admin@example.com",
				occupationLabel: "看護師",
				permissionLabel: "執行",
				clinicLabel:     "八王子病院",
				isSystemAdmin:   false,
			}},
		},
		{
			name: "isSystemAdmin false is captured",
			src:  `{ email: "vet@example.com", occupationLabel: "看護師", permissionLabel: "一般", clinicLabel: "八王子病院", isSystemAdmin: false },`,
			want: []demoAccountUILabels{{
				email:           "vet@example.com",
				occupationLabel: "看護師",
				permissionLabel: "一般",
				clinicLabel:     "八王子病院",
				isSystemAdmin:   false,
			}},
		},
		{
			name: "comment lines without objects are ignored",
			src:  "// 八王子病院\nconst x = 1;\n",
			want: nil,
		},
		{
			name: "prettier multiline objects are parsed",
			src: `{
        email: "stg-staff-11000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-21000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },`,
			want: []demoAccountUILabels{
				{
					email:           "stg-staff-11000021@example.test",
					occupationLabel: "獣医師",
					permissionLabel: "一般",
					clinicLabel:     "城東センター病院",
					isSystemAdmin:   false,
				},
				{
					email:           "stg-staff-21000021@example.test",
					occupationLabel: "獣医師",
					permissionLabel: "一般",
					clinicLabel:     "ノア動物病院　敷島病院",
					isSystemAdmin:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDemoAccountObjectLines(tt.src)
			if len(got) != len(tt.want) {
				t.Fatalf("parsed %d account(s), want %d (%#v)", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("account[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDemoAccountLabelDriftViolations_FailClosedOnCount(t *testing.T) {
	tables := demoAccountJoinFixtureTables(t, 10)

	tests := []struct {
		name string
		ui   []demoAccountUILabels
	}{
		{name: "nil slice", ui: nil},
		{name: "empty slice", ui: []demoAccountUILabels{}},
		{name: "one account", ui: demoAccountJoinFixtureUI(t, 1, nil)},
		{name: "eight accounts", ui: demoAccountJoinFixtureUI(t, 8, nil)},
		{name: "thirteen accounts", ui: demoAccountJoinFixtureUI(t, 13, nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := demoAccountStaffAttachContractViolations(tt.ui, tables)
			if len(violations) == 0 {
				t.Fatalf("violations = %v, want count failure", violations)
			}
			wantRange := fmt.Sprintf("want %d..%d", demoAccountExpectedMinCount, demoAccountExpectedMaxCount)
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, wantRange) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("violations = %v, want count range failure", violations)
			}
		})
	}
}

func TestDemoAccountLabelDriftViolations_ComparesLabels(t *testing.T) {
	tables := demoAccountJoinFixtureTables(t, demoAccountExpectedMinCount)

	validUI := demoAccountJoinFixtureUI(t, demoAccountExpectedMinCount, nil)
	if violations := demoAccountStaffAttachContractViolations(validUI, tables); len(violations) != 0 {
		t.Fatalf("valid fixture violations = %v, want none", violations)
	}

	tests := []struct {
		name          string
		modify        func(int, *demoAccountUILabels)
		wantSubstring string
	}{
		{
			name: "permissionLabel drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 2 {
					account.permissionLabel = "存在しない権限"
				}
			},
			wantSubstring: `permissionLabel "存在しない権限" is not in seed CSV 002_master/permission_groups.csv`,
		},
		{
			name: "occupationLabel empty fails closed",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 0 {
					account.occupationLabel = ""
				}
			},
			wantSubstring: "occupationLabel is empty (fail closed)",
		},
		{
			name: "non-admin clinicLabel drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 3 {
					account.clinicLabel = "敷島医院"
				}
			},
			wantSubstring: `clinicLabel "敷島医院" is not in seed CSV 002_master/clinics.csv`,
		},
		{
			name: "retired demo email fails closed",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 0 {
					account.email = "admin@example.com"
				}
			},
			wantSubstring: "retired 003_demo email admin@example.com is not allowed",
		},
		{
			name: "non staff-attach email fails closed",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 1 {
					account.email = "someone@example.com"
				}
			},
			wantSubstring: "email someone@example.com must match stg-staff-{id}@example.test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := demoAccountJoinFixtureUI(t, demoAccountExpectedMinCount, tt.modify)
			violations := demoAccountStaffAttachContractViolations(ui, tables)
			if len(violations) == 0 {
				t.Fatal("got no violations, want a contract mismatch")
			}
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, tt.wantSubstring) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("violations = %v, want substring %q", violations, tt.wantSubstring)
			}
		})
	}
}

func TestDemoAccountClinicLabelAllowed(t *testing.T) {
	tests := []struct {
		name          string
		isSystemAdmin bool
		uiLabel       string
		seedClinic    string
		want          bool
	}{
		{
			name:          "system admin 全医院 is allowed even when seed clinic differs",
			isSystemAdmin: true,
			uiLabel:       demoAccountSystemAdminClinicLabel,
			seedClinic:    "八王子病院",
			want:          true,
		},
		{
			name:          "system admin exact clinic match is allowed",
			isSystemAdmin: true,
			uiLabel:       "八王子病院",
			seedClinic:    "八王子病院",
			want:          true,
		},
		{
			name:          "system admin wrong clinic is rejected",
			isSystemAdmin: true,
			uiLabel:       "城東センター病院",
			seedClinic:    "八王子病院",
			want:          false,
		},
		{
			name:          "non-admin exact clinic match is required",
			isSystemAdmin: false,
			uiLabel:       "八王子病院",
			seedClinic:    "八王子病院",
			want:          true,
		},
		{
			name:          "non-admin 全医院 is rejected",
			isSystemAdmin: false,
			uiLabel:       demoAccountSystemAdminClinicLabel,
			seedClinic:    "八王子病院",
			want:          false,
		},
		{
			name:          "non-admin empty clinicLabel is rejected",
			isSystemAdmin: false,
			uiLabel:       "",
			seedClinic:    "八王子病院",
			want:          false,
		},
		{
			name:          "non-admin mismatch is rejected",
			isSystemAdmin: false,
			uiLabel:       "敷島医院",
			seedClinic:    "八王子病院",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demoAccountClinicLabelAllowed(tt.isSystemAdmin, tt.uiLabel, tt.seedClinic)
			if got != tt.want {
				t.Fatalf(
					"demoAccountClinicLabelAllowed(%v, %q, %q) = %v, want %v",
					tt.isSystemAdmin,
					tt.uiLabel,
					tt.seedClinic,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestErrIfSeedCSVLFSPointer(t *testing.T) {
	tests := []struct {
		name      string
		firstLine string
		wantErr   bool
	}{
		{
			name:      "exact Git LFS pointer first line",
			firstLine: gitLFSPointerPrefix,
			wantErr:   true,
		},
		{
			name:      "pointer with trailing CR",
			firstLine: gitLFSPointerPrefix + "\r",
			wantErr:   true,
		},
		{
			name:      "real CSV header is not a pointer",
			firstLine: "id,email,password_hash,is_active,is_system_admin",
			wantErr:   false,
		},
		{
			name:      "empty first line is not a pointer",
			firstLine: "",
			wantErr:   false,
		},
		{
			name:      "pointer text not at start is not a pointer",
			firstLine: "x " + gitLFSPointerPrefix,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errIfSeedCSVLFSPointer(demoAccountLabelDriftBundle, "accounts.csv", tt.firstLine)
			if tt.wantErr {
				if err == nil {
					t.Fatal("got nil error, want LFS pointer fail-closed")
				}
				if !strings.Contains(err.Error(), "not real data") {
					t.Fatalf("error = %q, want substring %q", err, "not real data")
				}
				if !strings.Contains(err.Error(), "002_master/accounts.csv") {
					t.Fatalf("error = %q, want seed CSV path", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}
		})
	}
}

func TestReadComparedDemoSeedCSV_LFSPointerIsNotRealData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.csv")
	pointer := gitLFSPointerPrefix + "\noid sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\nsize 42\n"
	if err := os.WriteFile(path, []byte(pointer), 0o600); err != nil {
		t.Fatalf("write LFS pointer fixture: %v", err)
	}

	_, err := readComparedDemoSeedCSV(demoAccountLabelDriftBundle, "accounts.csv", path)
	if err == nil {
		t.Fatal("expected LFS pointer fixture to fail closed as not real data")
	}
	if !strings.Contains(err.Error(), "not real data") {
		t.Fatalf("error = %q, want substring %q", err, "not real data")
	}
	if !strings.Contains(err.Error(), "Git LFS pointer") {
		t.Fatalf("error = %q, want Git LFS pointer", err)
	}
}

func TestReadComparedDemoSeedCSV_RealHeaderIsAccepted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.csv")
	contents := "id,email\n1,admin@example.com\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write CSV fixture: %v", err)
	}

	table, err := readComparedDemoSeedCSV(demoAccountLabelDriftBundle, "accounts.csv", path)
	if err != nil {
		t.Fatalf("read real CSV fixture: %v", err)
	}
	if table.csvFile != "accounts.csv" || len(table.rows) != 1 {
		t.Fatalf("table = %#v, want accounts.csv with 1 data row", table)
	}
}

func parseDemoAccountObjectLines(source string) []demoAccountUILabels {
	matches := demoAccountObjectLinePattern.FindAllStringSubmatch(source, -1)
	if len(matches) == 0 {
		return nil
	}
	accounts := make([]demoAccountUILabels, 0, len(matches))
	for _, match := range matches {
		accounts = append(accounts, demoAccountUILabels{
			email:           match[1],
			occupationLabel: match[2],
			permissionLabel: match[3],
			clinicLabel:     match[4],
			isSystemAdmin:   match[5] == "true",
		})
	}
	return accounts
}

func demoAccountStaffAttachContractViolations(
	uiAccounts []demoAccountUILabels,
	tables map[string]demoAccountSeedCSVTable,
) []string {
	if len(uiAccounts) < demoAccountExpectedMinCount || len(uiAccounts) > demoAccountExpectedMaxCount {
		return []string{fmt.Sprintf(
			"extracted DEMO_ACCOUNTS count = %d, want %d..%d (fail closed)",
			len(uiAccounts),
			demoAccountExpectedMinCount,
			demoAccountExpectedMaxCount,
		)}
	}

	permissionNames, err := demoAccountSeedNameSet(tables, "permission_groups", "name")
	if err != nil {
		return []string{err.Error()}
	}
	clinicNames, err := demoAccountSeedNameSet(tables, "clinics", "name")
	if err != nil {
		return []string{err.Error()}
	}

	var violations []string
	seenEmail := make(map[string]struct{}, len(uiAccounts))
	for _, ui := range uiAccounts {
		if _, dup := seenEmail[ui.email]; dup {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS email %s is duplicated (fail closed)",
				ui.email,
			))
		}
		seenEmail[ui.email] = struct{}{}

		for _, retired := range demoAccountRetiredDemoEmails {
			if ui.email == retired {
				violations = append(violations, fmt.Sprintf(
					"DEMO_ACCOUNTS retired 003_demo email %s is not allowed",
					ui.email,
				))
			}
		}
		if !demoAccountStaffAttachEmailRE.MatchString(ui.email) {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS email %s must match stg-staff-{id}@example.test",
				ui.email,
			))
		}
		if strings.TrimSpace(ui.occupationLabel) == "" {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s occupationLabel is empty (fail closed)",
				ui.email,
			))
		}
		if _, ok := permissionNames[ui.permissionLabel]; !ok {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s permissionLabel %q is not in seed CSV %s/%s",
				ui.email,
				ui.permissionLabel,
				demoAccountLabelDriftBundle,
				tables["permission_groups"].csvFile,
			))
		}
		if !demoAccountClinicLabelInMaster(ui.isSystemAdmin, ui.clinicLabel, clinicNames) {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s clinicLabel %q is not in seed CSV %s/%s",
				ui.email,
				ui.clinicLabel,
				demoAccountLabelDriftBundle,
				tables["clinics"].csvFile,
			))
		}
	}
	return violations
}

// demoAccountClinicLabelInMaster allows system-admin rows to use clinicLabel "全医院"
// or any 002_master clinic name. Non-admin rows must match a master clinic name.
func demoAccountClinicLabelInMaster(isSystemAdmin bool, uiLabel string, clinicNames map[string]struct{}) bool {
	if _, ok := clinicNames[uiLabel]; ok {
		return true
	}
	return isSystemAdmin && uiLabel == demoAccountSystemAdminClinicLabel
}

// demoAccountClinicLabelAllowed keeps the historical unit-test helper for
// system-admin "全医院" vs exact clinic matching.
func demoAccountClinicLabelAllowed(isSystemAdmin bool, uiLabel, seedClinic string) bool {
	if uiLabel == seedClinic {
		return true
	}
	return isSystemAdmin && uiLabel == demoAccountSystemAdminClinicLabel
}

func demoAccountSeedNameSet(
	tables map[string]demoAccountSeedCSVTable,
	tableName, column string,
) (map[string]struct{}, error) {
	table, err := demoAccountRequiredTable(tables, tableName)
	if err != nil {
		return nil, err
	}
	index, ok := table.col[column]
	if !ok {
		return nil, fmt.Errorf(
			"seed CSV %s/%s missing column %s (fail closed)",
			table.bundle,
			table.csvFile,
			column,
		)
	}
	names := make(map[string]struct{}, len(table.rows))
	for _, row := range table.rows {
		if index >= len(row) {
			continue
		}
		name := strings.TrimSpace(row[index])
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf(
			"seed CSV %s/%s has no %s values (fail closed)",
			table.bundle,
			table.csvFile,
			column,
		)
	}
	return names, nil
}

func demoAccountRequiredTable(
	tables map[string]demoAccountSeedCSVTable,
	table string,
) (demoAccountSeedCSVTable, error) {
	loaded, ok := tables[table]
	if !ok {
		return demoAccountSeedCSVTable{}, fmt.Errorf(
			"compared seed table %s is missing (fail closed)",
			table,
		)
	}
	return loaded, nil
}

func loadDemoAccountComparedSeedTables(moduleRoot string) (map[string]demoAccountSeedCSVTable, error) {
	bundleDir := filepath.Join(moduleRoot, "migrations", "seeds", demoAccountLabelDriftBundle)
	manifest, err := readSeedCSVManifest(filepath.Join(bundleDir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("load bundle %s manifest: %w", demoAccountLabelDriftBundle, err)
	}

	byTable := make(map[string]seedCSVManifestEntry, len(manifest.Tables))
	for _, entry := range manifest.Tables {
		byTable[entry.Table] = entry
	}

	loaded := make(map[string]demoAccountSeedCSVTable, len(demoAccountComparedSeedTables))
	for _, table := range demoAccountComparedSeedTables {
		entry, ok := byTable[table]
		if !ok {
			return nil, fmt.Errorf(
				"bundle %s manifest missing table %s (fail closed)",
				demoAccountLabelDriftBundle,
				table,
			)
		}
		csvPath := filepath.Join(bundleDir, entry.CSVFile)
		parsed, err := readComparedDemoSeedCSV(demoAccountLabelDriftBundle, entry.CSVFile, csvPath)
		if err != nil {
			return nil, err
		}
		parsed.table = table
		loaded[table] = parsed
	}
	return loaded, nil
}

func readComparedDemoSeedCSV(bundle, csvFile, path string) (demoAccountSeedCSVTable, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // path is a repository-owned seed CSV under the compared bundle.
	if err != nil {
		return demoAccountSeedCSVTable{}, fmt.Errorf("read seed CSV %s/%s: %w", bundle, csvFile, err)
	}

	if err := errIfSeedCSVLFSPointer(bundle, csvFile, seedCSVFirstLine(raw)); err != nil {
		return demoAccountSeedCSVTable{}, err
	}

	records, err := csv.NewReader(bytes.NewReader(raw)).ReadAll()
	if err != nil {
		return demoAccountSeedCSVTable{}, fmt.Errorf("parse seed CSV %s/%s: %w", bundle, csvFile, err)
	}
	if len(records) < 1 {
		return demoAccountSeedCSVTable{}, fmt.Errorf("seed CSV %s/%s has no header (fail closed)", bundle, csvFile)
	}

	header := records[0]
	col := make(map[string]int, len(header))
	for index, name := range header {
		col[strings.TrimSpace(name)] = index
	}
	return demoAccountSeedCSVTable{
		bundle:  bundle,
		csvFile: csvFile,
		header:  append([]string(nil), header...),
		rows:    records[1:],
		col:     col,
	}, nil
}

func seedCSVFirstLine(raw []byte) string {
	text := strings.TrimPrefix(string(raw), "\uFEFF")
	line, _, _ := strings.Cut(text, "\n")
	return strings.TrimRight(line, "\r")
}

func errIfSeedCSVLFSPointer(bundle, csvFile, firstLine string) error {
	if strings.HasPrefix(strings.TrimSpace(firstLine), gitLFSPointerPrefix) {
		return fmt.Errorf(
			"seed CSV %s/%s first line is a Git LFS pointer; this is not real data",
			bundle,
			csvFile,
		)
	}
	return nil
}

func (table demoAccountSeedCSVTable) rowsWhere(column, value string) ([][]string, error) {
	index, ok := table.col[column]
	if !ok {
		return nil, fmt.Errorf(
			"seed CSV %s/%s table %s missing column %s (fail closed)",
			table.bundle,
			table.csvFile,
			table.table,
			column,
		)
	}
	var matches [][]string
	for _, row := range table.rows {
		if index < len(row) && row[index] == value {
			matches = append(matches, row)
		}
	}
	return matches, nil
}

func (table demoAccountSeedCSVTable) field(row []string, column string) (string, error) {
	index, ok := table.col[column]
	if !ok {
		return "", fmt.Errorf(
			"seed CSV %s/%s table %s missing column %s (fail closed)",
			table.bundle,
			table.csvFile,
			table.table,
			column,
		)
	}
	if index >= len(row) {
		return "", fmt.Errorf(
			"seed CSV %s/%s table %s row is too short for column %s (fail closed)",
			table.bundle,
			table.csvFile,
			table.table,
			column,
		)
	}
	return row[index], nil
}

func parseDemoAccountSeedBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "t", "true":
		return true, nil
	case "f", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q (fail closed)", raw)
	}
}

func mustFindDemoAccountLoginFormPath(t *testing.T, moduleRoot string) string {
	t.Helper()

	path, err := findDemoAccountLoginFormPath(moduleRoot)
	if err != nil {
		t.Fatalf("resolve LoginForm.tsx: %v", err)
	}
	return path
}

func findDemoAccountLoginFormPath(moduleRoot string) (string, error) {
	// moduleRoot is backend/ (host) or /app (compose). The sibling frontend
	// tree is therefore repo/frontend on the host and /frontend in compose
	// (docker-compose.yml mounts ./frontend:/frontend:ro on backend).
	candidates := []string{
		filepath.Clean(filepath.Join(moduleRoot, "..", demoAccountLoginFormRelativePath)),
		filepath.Clean(filepath.Join(string(os.PathSeparator)+"frontend", "src", "features", "auth", "components", "LoginForm.tsx")),
	}
	if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
		candidates = append(candidates, filepath.Join(workspace, demoAccountLoginFormRelativePath))
	}

	var searched []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		searched = append(searched, candidate)
		info, err := os.Stat(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("stat LoginForm.tsx %s: %w", candidate, err)
		}
		if info.IsDir() {
			continue
		}
		return candidate, nil
	}
	return "", fmt.Errorf(
		"LoginForm.tsx not found (fail closed); searched %v",
		searched,
	)
}

func demoAccountJoinFixtureUI(
	t *testing.T,
	count int,
	modify func(int, *demoAccountUILabels),
) []demoAccountUILabels {
	t.Helper()

	accounts := make([]demoAccountUILabels, count)
	for index := range accounts {
		accounts[index] = demoAccountUILabels{
			email:           fmt.Sprintf("stg-staff-%d@example.test", 11000001+index),
			occupationLabel: "獣医師",
			permissionLabel: "一般",
			clinicLabel:     "城東センター病院",
		}
		if modify != nil {
			modify(index, &accounts[index])
		}
	}
	return accounts
}

func demoAccountJoinFixtureTables(t *testing.T, count int) map[string]demoAccountSeedCSVTable {
	t.Helper()
	_ = count

	return map[string]demoAccountSeedCSVTable{
		"permission_groups": fixtureDemoAccountSeedCSVTable(
			"permission_groups",
			[]string{"id", "name"},
			[][]string{{"1", "執行"}, {"4", "一般"}},
		),
		"clinics": fixtureDemoAccountSeedCSVTable(
			"clinics",
			[]string{"id", "name"},
			[][]string{{"1", "八王子病院"}, {"2", "城東センター病院"}},
		),
	}
}

func fixtureDemoAccountSeedCSVTable(table string, header []string, rows [][]string) demoAccountSeedCSVTable {
	col := make(map[string]int, len(header))
	for index, name := range header {
		col[name] = index
	}
	return demoAccountSeedCSVTable{
		bundle:  demoAccountLabelDriftBundle,
		table:   table,
		csvFile: table + ".csv",
		header:  append([]string(nil), header...),
		rows:    rows,
		col:     col,
	}
}
