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
)

const (
	demoAccountLabelDriftBundle       = "003_demo"
	demoAccountExpectedCount          = 9
	gitLFSPointerPrefix               = "version https://git-lfs.github.com/spec/v1"
	demoAccountSystemAdminClinicLabel = "全医院"
	demoAccountLoginFormRelativePath  = "frontend/src/features/auth/components/LoginForm.tsx"
)

var demoAccountComparedSeedTables = []string{
	"accounts",
	"staffs",
	"staff_permission_groups",
	"permission_groups",
	"occupations",
	"clinics",
}

// demoAccountObjectLinePattern extracts one DEMO_ACCOUNTS object from a single
// LoginForm.tsx line. Field order matches the current source (email, labels,
// optional isSystemAdmin). Reordered lines fail closed via count != 9.
var demoAccountObjectLinePattern = regexp.MustCompile(
	`email:\s*"([^"]+)"` +
		`.*?occupationLabel:\s*"([^"]+)"` +
		`.*?permissionLabel:\s*"([^"]+)"` +
		`.*?clinicLabel:\s*"([^"]+)"` +
		`(?:.*?isSystemAdmin:\s*(true|false))?`,
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

func TestDemoAccountLoginFormLabelsMatch003DemoSeed(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)

	source, err := os.ReadFile(mustFindDemoAccountLoginFormPath(t, moduleRoot))
	if err != nil {
		t.Fatalf("read LoginForm.tsx: %v", err)
	}

	uiAccounts := parseDemoAccountObjectLines(string(source))
	tables, err := loadDemoAccountComparedSeedTables(moduleRoot)
	if err != nil {
		t.Fatalf("load %s compared seed CSV: %v", demoAccountLabelDriftBundle, err)
	}

	for _, violation := range demoAccountLabelDriftViolations(uiAccounts, tables) {
		t.Errorf("%s", violation)
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
	tables := demoAccountJoinFixtureTables(t, 9)

	tests := []struct {
		name string
		ui   []demoAccountUILabels
	}{
		{name: "nil slice", ui: nil},
		{name: "empty slice", ui: []demoAccountUILabels{}},
		{name: "one account", ui: demoAccountJoinFixtureUI(t, 1, nil)},
		{name: "eight accounts", ui: demoAccountJoinFixtureUI(t, 8, nil)},
		{name: "ten accounts", ui: demoAccountJoinFixtureUI(t, 10, nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := demoAccountLabelDriftViolations(tt.ui, tables)
			if len(violations) != 1 {
				t.Fatalf("violations = %v, want exactly one count failure", violations)
			}
			want := fmt.Sprintf(
				"extracted DEMO_ACCOUNTS count = %d, want %d (fail closed)",
				len(tt.ui),
				demoAccountExpectedCount,
			)
			if violations[0] != want {
				t.Fatalf("violation = %q, want %q", violations[0], want)
			}
		})
	}
}

func TestDemoAccountLabelDriftViolations_ComparesLabels(t *testing.T) {
	tables := demoAccountJoinFixtureTables(t, 9)

	tests := []struct {
		name          string
		modify        func(int, *demoAccountUILabels)
		wantSubstring string
	}{
		{
			name: "matching labels have no violations",
			modify: func(int, *demoAccountUILabels) {
			},
		},
		{
			name: "permissionLabel drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 2 {
					account.permissionLabel = "一般"
				}
			},
			wantSubstring: `permissionLabel "一般" does not match seed CSV 003_demo/permission_groups.csv permission group "執行"`,
		},
		{
			name: "occupationLabel drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 0 {
					account.occupationLabel = "受付"
				}
			},
			wantSubstring: `occupationLabel "受付" does not match seed CSV 003_demo/occupations.csv occupation "獣医師"`,
		},
		{
			name: "non-admin clinicLabel drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 3 {
					account.clinicLabel = "敷島医院"
				}
			},
			wantSubstring: `clinicLabel "敷島医院" does not match seed CSV 003_demo/clinics.csv clinic "八王子病院"`,
		},
		{
			name: "isSystemAdmin drift",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 1 {
					account.isSystemAdmin = true
					account.clinicLabel = demoAccountSystemAdminClinicLabel
				}
			},
			wantSubstring: "isSystemAdmin = true, seed CSV 003_demo/accounts.csv is_system_admin = false",
		},
		{
			name: "missing seed email fails closed",
			modify: func(i int, account *demoAccountUILabels) {
				if i == 0 {
					account.email = "missing@example.com"
				}
			},
			wantSubstring: "accounts.csv email missing@example.com matched 0 row(s), want 1 (fail closed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := demoAccountJoinFixtureUI(t, 9, tt.modify)
			violations := demoAccountLabelDriftViolations(ui, tables)
			if tt.wantSubstring == "" {
				if len(violations) != 0 {
					t.Fatalf("violations = %v, want none", violations)
				}
				return
			}
			if len(violations) == 0 {
				t.Fatal("got no violations, want a label mismatch")
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
				if !strings.Contains(err.Error(), "003_demo/accounts.csv") {
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
	var accounts []demoAccountUILabels
	for _, line := range strings.Split(source, "\n") {
		match := demoAccountObjectLinePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
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

func demoAccountLabelDriftViolations(
	uiAccounts []demoAccountUILabels,
	tables map[string]demoAccountSeedCSVTable,
) []string {
	if len(uiAccounts) != demoAccountExpectedCount {
		return []string{fmt.Sprintf(
			"extracted DEMO_ACCOUNTS count = %d, want %d (fail closed)",
			len(uiAccounts),
			demoAccountExpectedCount,
		)}
	}

	var violations []string
	for _, ui := range uiAccounts {
		seed, err := seedLabelsForDemoAccountEmail(ui.email, tables)
		if err != nil {
			violations = append(violations, fmt.Sprintf("DEMO_ACCOUNTS %s: %v", ui.email, err))
			continue
		}
		if ui.occupationLabel != seed.occupationLabel {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s occupationLabel %q does not match seed CSV %s/%s occupation %q",
				ui.email,
				ui.occupationLabel,
				demoAccountLabelDriftBundle,
				tables["occupations"].csvFile,
				seed.occupationLabel,
			))
		}
		if ui.permissionLabel != seed.permissionLabel {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s permissionLabel %q does not match seed CSV %s/%s permission group %q",
				ui.email,
				ui.permissionLabel,
				demoAccountLabelDriftBundle,
				tables["permission_groups"].csvFile,
				seed.permissionLabel,
			))
		}
		if ui.isSystemAdmin != seed.isSystemAdmin {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s isSystemAdmin = %v, seed CSV %s/%s is_system_admin = %v",
				ui.email,
				ui.isSystemAdmin,
				demoAccountLabelDriftBundle,
				tables["accounts"].csvFile,
				seed.isSystemAdmin,
			))
		}
		if !demoAccountClinicLabelAllowed(ui.isSystemAdmin, ui.clinicLabel, seed.clinicLabel) {
			violations = append(violations, fmt.Sprintf(
				"DEMO_ACCOUNTS %s clinicLabel %q does not match seed CSV %s/%s clinic %q",
				ui.email,
				ui.clinicLabel,
				demoAccountLabelDriftBundle,
				tables["clinics"].csvFile,
				seed.clinicLabel,
			))
		}
	}
	return violations
}

// demoAccountClinicLabelAllowed documents the system-admin clinicLabel allowlist.
// System-admin DEMO_ACCOUNTS may show clinicLabel "全医院" instead of the staff
// row's clinic name: 003_demo still pins those accounts to a home clinic, and
// the login UI is an explicit all-clinics label, not a clinic-name match.
// Non-admin rows must match clinics.csv name exactly. Exact clinic match remains
// valid for system admins; any other clinicLabel is rejected.
func demoAccountClinicLabelAllowed(isSystemAdmin bool, uiLabel, seedClinic string) bool {
	if uiLabel == seedClinic {
		return true
	}
	return isSystemAdmin && uiLabel == demoAccountSystemAdminClinicLabel
}

func seedLabelsForDemoAccountEmail(
	email string,
	tables map[string]demoAccountSeedCSVTable,
) (demoAccountSeedLabels, error) {
	accounts, err := demoAccountRequiredTable(tables, "accounts")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	accountRows, err := accounts.rowsWhere("email", email)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(accountRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"accounts.csv email %s matched %d row(s), want 1 (fail closed)",
			email,
			len(accountRows),
		)
	}

	accountID, err := accounts.field(accountRows[0], "id")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	adminRaw, err := accounts.field(accountRows[0], "is_system_admin")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	isSystemAdmin, err := parseDemoAccountSeedBool(adminRaw)
	if err != nil {
		return demoAccountSeedLabels{}, fmt.Errorf("accounts.csv is_system_admin: %w", err)
	}

	staffs, err := demoAccountRequiredTable(tables, "staffs")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	staffRows, err := staffs.rowsWhere("account_id", accountID)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(staffRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"staffs.csv account_id %s matched %d row(s), want 1 (fail closed)",
			accountID,
			len(staffRows),
		)
	}
	staffID, err := staffs.field(staffRows[0], "id")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	occupationID, err := staffs.field(staffRows[0], "occupation_id")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	clinicID, err := staffs.field(staffRows[0], "clinic_id")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}

	groups, err := demoAccountRequiredTable(tables, "staff_permission_groups")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	groupRows, err := groups.rowsWhere("staff_id", staffID)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(groupRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"staff_permission_groups.csv staff_id %s matched %d row(s), want 1 (fail closed)",
			staffID,
			len(groupRows),
		)
	}
	groupID, err := groups.field(groupRows[0], "group_id")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}

	permissionGroups, err := demoAccountRequiredTable(tables, "permission_groups")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	permissionRows, err := permissionGroups.rowsWhere("id", groupID)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(permissionRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"permission_groups.csv id %s matched %d row(s), want 1 (fail closed)",
			groupID,
			len(permissionRows),
		)
	}
	permissionLabel, err := permissionGroups.field(permissionRows[0], "name")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}

	occupations, err := demoAccountRequiredTable(tables, "occupations")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	occupationRows, err := occupations.rowsWhere("id", occupationID)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(occupationRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"occupations.csv id %s matched %d row(s), want 1 (fail closed)",
			occupationID,
			len(occupationRows),
		)
	}
	occupationLabel, err := occupations.field(occupationRows[0], "name")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}

	clinics, err := demoAccountRequiredTable(tables, "clinics")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	clinicRows, err := clinics.rowsWhere("id", clinicID)
	if err != nil {
		return demoAccountSeedLabels{}, err
	}
	if len(clinicRows) != 1 {
		return demoAccountSeedLabels{}, fmt.Errorf(
			"clinics.csv id %s matched %d row(s), want 1 (fail closed)",
			clinicID,
			len(clinicRows),
		)
	}
	clinicLabel, err := clinics.field(clinicRows[0], "name")
	if err != nil {
		return demoAccountSeedLabels{}, err
	}

	return demoAccountSeedLabels{
		occupationLabel: occupationLabel,
		permissionLabel: permissionLabel,
		clinicLabel:     clinicLabel,
		isSystemAdmin:   isSystemAdmin,
	}, nil
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
	raw, err := os.ReadFile(path) //nolint:gosec // path is a repository-owned 003_demo seed CSV.
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
			email:           fmt.Sprintf("user%d@example.com", index+1),
			occupationLabel: "獣医師",
			permissionLabel: "執行",
			clinicLabel:     "八王子病院",
		}
		if modify != nil {
			modify(index, &accounts[index])
		}
	}
	return accounts
}

func demoAccountJoinFixtureTables(t *testing.T, count int) map[string]demoAccountSeedCSVTable {
	t.Helper()

	accountRows := make([][]string, 0, count)
	staffRows := make([][]string, 0, count)
	groupRows := make([][]string, 0, count)
	for index := 0; index < count; index++ {
		id := fmt.Sprintf("%d", index+1)
		accountRows = append(accountRows, []string{id, fmt.Sprintf("user%d@example.com", index+1), "f"})
		staffRows = append(staffRows, []string{id, id, "1", "1"})
		groupRows = append(groupRows, []string{id, "1"})
	}

	return map[string]demoAccountSeedCSVTable{
		"accounts": fixtureDemoAccountSeedCSVTable(
			"accounts",
			[]string{"id", "email", "is_system_admin"},
			accountRows,
		),
		"staffs": fixtureDemoAccountSeedCSVTable(
			"staffs",
			[]string{"id", "account_id", "occupation_id", "clinic_id"},
			staffRows,
		),
		"staff_permission_groups": fixtureDemoAccountSeedCSVTable(
			"staff_permission_groups",
			[]string{"staff_id", "group_id"},
			groupRows,
		),
		"permission_groups": fixtureDemoAccountSeedCSVTable(
			"permission_groups",
			[]string{"id", "name"},
			[][]string{{"1", "執行"}},
		),
		"occupations": fixtureDemoAccountSeedCSVTable(
			"occupations",
			[]string{"id", "name"},
			[][]string{{"1", "獣医師"}},
		),
		"clinics": fixtureDemoAccountSeedCSVTable(
			"clinics",
			[]string{"id", "name"},
			[][]string{{"1", "八王子病院"}},
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
