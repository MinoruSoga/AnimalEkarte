package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestAllowlistVsCutoverTableSpecs(t *testing.T) {
	t.Parallel()

	specs := csvimport.CutoverTableSpecs()
	require.Len(t, specs, 21)
	require.Equal(t, "staffs", specs[0].Name)

	cutoverNames := make([]string, 0, len(specs))
	cutover := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		cutoverNames = append(cutoverNames, spec.Name)
		cutover[spec.Name] = struct{}{}
	}
	require.Equal(t, cutoverNames, cutoverTableNames())
	require.Contains(t, cutover, "staffs")

	allow := skeletonAllowlistTables()
	require.NotEmpty(t, allow)

	required := []string{
		"clinics",
		"payment_methods",
		"exam_types",
		"reservation_types",
		"permission_groups",
		"permission_group_rules",
	}
	for _, name := range required {
		assert.Contains(t, allow, name)
	}

	for _, name := range allow {
		_, forbidden := cutover[name]
		assert.Falsef(t, forbidden, "skeleton allowlist must not include cutover table %s", name)
	}
	assert.NotContains(t, allow, "staffs")
}

func TestRejectForbiddenWrite_CutoverTables(t *testing.T) {
	t.Parallel()

	specs := csvimport.CutoverTableSpecs()
	require.NotEmpty(t, specs)

	tests := make([]struct {
		name  string
		table string
		sql   string
	}, 0, len(specs)*2+1)
	for _, spec := range specs {
		tests = append(tests,
			struct {
				name  string
				table string
				sql   string
			}{spec.Name + " insert", spec.Name, "INSERT INTO " + spec.Name + " (id) VALUES (1)"},
			struct {
				name  string
				table string
				sql   string
			}{spec.Name + " copy", spec.Name, "COPY " + spec.Name + " FROM STDIN"},
		)
	}
	tests = append(tests, struct {
		name  string
		table string
		sql   string
	}{"staffs insert into hachioji band", "staffs", "INSERT INTO staffs (id, clinic_id, name) VALUES (1, 1, 'x')"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := rejectForbiddenWrite(tt.table, tt.sql)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.table)
			assert.Contains(t, err.Error(), "cutover")
		})
	}
}

func TestApplySkeleton(t *testing.T) {
	db := newMemDB()
	db.seedCompany(1)

	require.NoError(t, applySkeleton(context.Background(), db, bootstrapOpts{}))

	t.Run("does not write cutover tables including staffs", func(t *testing.T) {
		for _, spec := range csvimport.CutoverTableSpecs() {
			assert.Zerof(t, db.count(spec.Name), "cutover table %s", spec.Name)
			for _, w := range db.recordedWrites() {
				if strings.EqualFold(w.table, spec.Name) {
					t.Errorf("skeleton wrote cutover table %s sql=%s", spec.Name, w.sql)
				}
				lower := strings.ToLower(w.sql)
				if strings.Contains(lower, "into "+spec.Name) || strings.Contains(lower, "copy "+spec.Name) {
					t.Errorf("skeleton SQL targets cutover table %s: %s", spec.Name, w.sql)
				}
			}
		}
	})

	tests := []struct {
		name       string
		clinicID   int64
		clinicName string
	}{
		{name: "hachioji", clinicID: 1, clinicName: "八王子病院"},
		{name: "jouto", clinicID: 2, clinicName: "城東センター病院"},
	}
	for _, tt := range tests {
		t.Run(tt.name+" f6 bindings", func(t *testing.T) {
			c := db.clinic(tt.clinicID)
			require.NotNil(t, c)
			assert.Equal(t, int64(1), c.companyID)
			assert.Equal(t, tt.clinicName, c.name)
			assert.Truef(t, db.hasPaymentMethod(tt.clinicID, "cash"), "clinic %d missing cash", tt.clinicID)
			assert.Truef(t, db.hasPaymentMethod(tt.clinicID, "credit_card"), "clinic %d missing credit_card", tt.clinicID)
			assert.Truef(t, db.hasExamType(tt.clinicID, "検査"), "clinic %d missing exam 検査", tt.clinicID)
			assert.Truef(t, db.hasTrimmingType(tt.clinicID), "clinic %d missing trimming reservation type", tt.clinicID)
			assert.Truef(t, db.hasClinicSettings(tt.clinicID), "clinic %d missing clinic_settings", tt.clinicID)
			execGroup, generalGroup := db.permissionGroupsForClinic(tt.clinicID)
			require.NotNilf(t, execGroup, "clinic %d missing 執行 group", tt.clinicID)
			require.NotNilf(t, generalGroup, "clinic %d missing 一般 group", tt.clinicID)
			assert.Equal(t, "執行", execGroup.name)
			assert.Equal(t, "一般", generalGroup.name)
			assert.Equal(t, tt.clinicID, execGroup.clinicID)
			assert.Equal(t, tt.clinicID, generalGroup.clinicID)
			assert.Truef(t, db.hasRulesForGroup(execGroup.id), "clinic %d 執行 rules missing", tt.clinicID)
			assert.Truef(t, db.hasRulesForGroup(generalGroup.id), "clinic %d 一般 rules missing", tt.clinicID)
		})
	}

	t.Run("staffs count 0 in hachioji band", func(t *testing.T) {
		assert.Equal(t, int64(0), db.countStaffsInBand(0, 10_000_000))
		assert.Zero(t, db.count("staffs"))
	})

	t.Run("does not insert payment_methods when clinic trigger already filled them", func(t *testing.T) {
		for _, w := range db.recordedWrites() {
			if strings.EqualFold(w.table, "payment_methods") && insertRe.MatchString(w.sql) {
				t.Errorf("skeleton inserted payment_methods after clinic trigger: %s", w.sql)
			}
		}
		assert.Equal(t, 4, db.countPaymentMethods(1), "clinic trigger seeds 4 system methods")
		assert.Equal(t, 4, db.countPaymentMethods(2), "clinic trigger seeds 4 system methods")
	})
}

func TestEnsureDefaultPaymentMethods_InsertsWhenTriggerAbsent(t *testing.T) {
	db := newMemDB()
	db.skipPaymentMethodTrigger = true
	db.seedCompany(1)
	require.NoError(t, applySkeleton(context.Background(), db, bootstrapOpts{}))
	assert.True(t, db.hasPaymentMethod(1, "cash"))
	assert.True(t, db.hasPaymentMethod(1, "credit_card"))
	assert.True(t, db.hasPaymentMethod(2, "cash"))
	assert.True(t, db.hasPaymentMethod(2, "credit_card"))
	var inserts int
	for _, w := range db.recordedWrites() {
		if strings.EqualFold(w.table, "payment_methods") && insertRe.MatchString(w.sql) {
			inserts++
		}
	}
	assert.Equal(t, 4, inserts)
}

func TestApplySkeleton_BootstrapAccountDoesNotCreateStaffs(t *testing.T) {
	db := newMemDB()
	db.seedCompany(1)
	require.NoError(t, applySkeleton(context.Background(), db, bootstrapOpts{
		email:        "ops@example.test",
		passwordHash: "hash",
	}))
	assert.Equal(t, 1, db.count("accounts"))
	assert.Zero(t, db.count("staffs"))
	assert.Equal(t, int64(0), db.countStaffsInBand(0, 10_000_000))
}

func TestPermissionBits_MasterAnimalSpeciesIsViewOnly(t *testing.T) {
	t.Parallel()
	view, create, edit, del := permissionBits(model.ResourceMasterAnimalSpecies, true)
	assert.True(t, view)
	assert.False(t, create)
	assert.False(t, edit)
	assert.False(t, del)
	view, create, edit, del = permissionBits(model.ResourceMasterAnimalSpecies, false)
	assert.True(t, view)
	assert.False(t, create)
	assert.False(t, edit)
	assert.False(t, del)
}

func TestPermissionBits_GeneralDeniesMasterPermissionAndDiscount(t *testing.T) {
	t.Parallel()
	for _, resource := range []model.Resource{model.ResourceMasterPermission, model.ResourceDiscount} {
		view, create, edit, del := permissionBits(resource, false)
		assert.Falsef(t, view, "%s general view", resource)
		assert.Falsef(t, create, "%s general create", resource)
		assert.Falsef(t, edit, "%s general edit", resource)
		assert.Falsef(t, del, "%s general delete", resource)
	}
}

func TestRun_ApplyRefusesNonLocalWithoutOverride(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		override      string
		wantSubstring string
	}{
		{
			name:          "example.invalid without override",
			host:          "example.invalid",
			override:      "",
			wantSubstring: "non-local",
		},
		{
			name:          "example.invalid names own allow-remote env",
			host:          "example.invalid",
			override:      "",
			wantSubstring: "STG_UAT_SKELETON_ALLOW_REMOTE=YES_I_UNDERSTAND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_NAME", "animalekarte")
			t.Setenv("STG_UAT_SKELETON_ALLOW_REMOTE", tt.override)

			opened := false
			deps := runDependencies{
				fromEnv: func() (dbconn.ConnParams, error) {
					return dbconn.ConnParams{
						Host: tt.host, Port: "5432", User: "u", Password: "p", SSLMode: "disable",
					}, nil
				},
				openDB: func(context.Context, string) (dbSession, error) {
					opened = true
					t.Fatal("openDB must not be called when remote apply is refused")
					return nil, nil
				},
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := run(context.Background(), []string{"apply"}, logger, deps)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantSubstring)
			assert.False(t, opened)
		})
	}
}

type clinicRow struct {
	id        int64
	companyID int64
	name      string
}

type recordedWrite struct {
	table string
	sql   string
	args  []any
}

type memRow struct {
	err  error
	vals []any
}

func (r memRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.vals) {
		return fmt.Errorf("scan dest=%d vals=%d", len(dest), len(r.vals))
	}
	for i := range dest {
		if err := assignScan(dest[i], r.vals[i]); err != nil {
			return err
		}
	}
	return nil
}

type memDB struct {
	mu                       sync.Mutex
	writes                   []recordedWrite
	tables                   map[string][]map[string]any
	seq                      map[string]int64
	skipPaymentMethodTrigger bool
}

func newMemDB() *memDB {
	return &memDB{
		tables: map[string][]map[string]any{},
		seq:    map[string]int64{},
	}
}

func (m *memDB) seedCompany(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tables["companies"] = append(m.tables["companies"], map[string]any{"id": id})
}

func (m *memDB) Exec(_ context.Context, sql string, args ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	table, row, isInsert := parseInsert(sql, args)
	if table == "" {
		table = tableFromSQL(sql)
	}
	m.writes = append(m.writes, recordedWrite{table: table, sql: sql, args: append([]any(nil), args...)})
	if !isInsert {
		return nil
	}
	if table == "payment_methods" && m.paymentMethodExistsLocked(asInt64(row["clinic_id"]), asString(row["system_key"])) {
		return fmt.Errorf("duplicate key idx_payment_methods_clinic_system_key clinic_id=%v system_key=%v", row["clinic_id"], row["system_key"])
	}
	if _, ok := row["id"]; !ok {
		m.seq[table]++
		row["id"] = m.seq[table]
	} else if n := asInt64(row["id"]); n > m.seq[table] {
		m.seq[table] = n
	}
	m.tables[table] = append(m.tables[table], row)
	if table == "clinics" && !m.skipPaymentMethodTrigger {
		m.seedDefaultPaymentMethodsLocked(asInt64(row["id"]))
	}
	return nil
}

func (m *memDB) QueryRow(_ context.Context, sql string, args ...any) rowScanner {
	m.mu.Lock()
	defer m.mu.Unlock()
	nsql := compactSQL(sql)

	switch {
	case strings.Contains(nsql, "from companies"):
		id := asInt64(args[0])
		return memRow{vals: []any{m.existsLocked("companies", "id", id)}}
	case strings.Contains(nsql, "from clinics where id"):
		id := asInt64(args[0])
		row := m.findLocked("clinics", "id", id)
		if row == nil {
			return memRow{err: fmt.Errorf("clinic %d not found", id)}
		}
		return memRow{vals: []any{asInt64(row["id"]), asInt64(row["company_id"]), asString(row["name"])}}
	case strings.Contains(nsql, "from payment_methods"):
		clinicID := asInt64(args[0])
		key := asString(args[1])
		n := m.countWhereLocked("payment_methods", func(r map[string]any) bool {
			return asInt64(r["clinic_id"]) == clinicID && asString(r["system_key"]) == key
		})
		return memRow{vals: []any{n}}
	case strings.Contains(nsql, "from exam_types"):
		clinicID := asInt64(args[0])
		name := asString(args[1])
		n := m.countWhereLocked("exam_types", func(r map[string]any) bool {
			return asInt64(r["clinic_id"]) == clinicID && asString(r["name"]) == name
		})
		return memRow{vals: []any{n}}
	case strings.Contains(nsql, "from reservation_types"):
		clinicID := asInt64(args[0])
		category := asString(args[1])
		n := m.countWhereLocked("reservation_types", func(r map[string]any) bool {
			return asInt64(r["clinic_id"]) == clinicID && asString(r["category"]) == category
		})
		return memRow{vals: []any{n}}
	case strings.Contains(nsql, "from clinic_settings"):
		clinicID := asInt64(args[0])
		n := m.countWhereLocked("clinic_settings", func(r map[string]any) bool {
			return asInt64(r["clinic_id"]) == clinicID
		})
		return memRow{vals: []any{n}}
	case strings.Contains(nsql, "from permission_groups where id"):
		id := asInt64(args[0])
		row := m.findLocked("permission_groups", "id", id)
		if row == nil {
			return memRow{err: fmt.Errorf("permission group %d not found", id)}
		}
		return memRow{vals: []any{asInt64(row["clinic_id"]), asString(row["name"])}}
	case strings.Contains(nsql, "from permission_group_rules"):
		groupID := asInt64(args[0])
		n := m.countWhereLocked("permission_group_rules", func(r map[string]any) bool {
			return asInt64(r["group_id"]) == groupID
		})
		return memRow{vals: []any{n}}
	case strings.Contains(nsql, "from staffs") && strings.Contains(nsql, "id >="):
		start, end := asInt64(args[0]), asInt64(args[1])
		n := m.countWhereLocked("staffs", func(r map[string]any) bool {
			id := asInt64(r["id"])
			return id >= start && id < end
		})
		return memRow{vals: []any{n}}
	case strings.HasPrefix(nsql, "select count(*) from "):
		table := strings.TrimPrefix(nsql, "select count(*) from ")
		table = strings.Trim(table, `"`)
		table = strings.TrimPrefix(table, "public.")
		if fields := strings.Fields(table); len(fields) > 0 {
			table = strings.Trim(fields[0], `"`)
		}
		return memRow{vals: []any{int64(len(m.tables[table]))}}
	default:
		return memRow{err: fmt.Errorf("unsupported query %q", sql)}
	}
}

func (m *memDB) recordedWrites() []recordedWrite {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]recordedWrite, len(m.writes))
	copy(out, m.writes)
	return out
}

func (m *memDB) count(table string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tables[table])
}

func (m *memDB) clinic(id int64) *clinicRow {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.findLocked("clinics", "id", id)
	if row == nil {
		return nil
	}
	return &clinicRow{
		id:        asInt64(row["id"]),
		companyID: asInt64(row["company_id"]),
		name:      asString(row["name"]),
	}
}

func (m *memDB) hasPaymentMethod(clinicID int64, systemKey string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.paymentMethodExistsLocked(clinicID, systemKey)
}

func (m *memDB) countPaymentMethods(clinicID int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int(m.countWhereLocked("payment_methods", func(r map[string]any) bool {
		return asInt64(r["clinic_id"]) == clinicID
	}))
}

func (m *memDB) paymentMethodExistsLocked(clinicID int64, systemKey string) bool {
	if systemKey == "" {
		return false
	}
	return m.countWhereLocked("payment_methods", func(r map[string]any) bool {
		return asInt64(r["clinic_id"]) == clinicID && asString(r["system_key"]) == systemKey
	}) > 0
}

func (m *memDB) seedDefaultPaymentMethodsLocked(clinicID int64) {
	defaults := []struct {
		name      string
		systemKey string
		order     int
	}{
		{"現金", "cash", 1},
		{"クレジットカード", "credit_card", 2},
		{"電子マネー", "electronic_money", 3},
		{"銀行振込", "bank_transfer", 4},
	}
	for _, pm := range defaults {
		if m.paymentMethodExistsLocked(clinicID, pm.systemKey) {
			continue
		}
		m.seq["payment_methods"]++
		m.tables["payment_methods"] = append(m.tables["payment_methods"], map[string]any{
			"id":            m.seq["payment_methods"],
			"clinic_id":     clinicID,
			"name":          pm.name,
			"system_key":    pm.systemKey,
			"display_order": pm.order,
		})
	}
}

func (m *memDB) hasExamType(clinicID int64, name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countWhereLocked("exam_types", func(r map[string]any) bool {
		return asInt64(r["clinic_id"]) == clinicID && asString(r["name"]) == name
	}) > 0
}

func (m *memDB) hasTrimmingType(clinicID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countWhereLocked("reservation_types", func(r map[string]any) bool {
		return asInt64(r["clinic_id"]) == clinicID && asString(r["category"]) == "trimming"
	}) > 0
}

type groupRow struct {
	id       int64
	clinicID int64
	name     string
}

func (m *memDB) hasClinicSettings(clinicID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findLocked("clinic_settings", "clinic_id", clinicID) != nil
}

func (m *memDB) permissionGroupsForClinic(clinicID int64) (execGroup, generalGroup *groupRow) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.tables["permission_groups"] {
		if asInt64(row["clinic_id"]) != clinicID {
			continue
		}
		g := &groupRow{id: asInt64(row["id"]), clinicID: clinicID, name: asString(row["name"])}
		switch g.name {
		case "執行":
			execGroup = g
		case "一般":
			generalGroup = g
		}
	}
	return execGroup, generalGroup
}

func (m *memDB) hasRulesForGroup(groupID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countWhereLocked("permission_group_rules", func(r map[string]any) bool {
		return asInt64(r["group_id"]) == groupID
	}) > 0
}

func (m *memDB) countStaffsInBand(start, end int64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countWhereLocked("staffs", func(r map[string]any) bool {
		id := asInt64(r["id"])
		return id >= start && id < end
	})
}

func (m *memDB) existsLocked(table, col string, id int64) bool {
	return m.findLocked(table, col, id) != nil
}

func (m *memDB) findLocked(table, col string, id int64) map[string]any {
	for _, row := range m.tables[table] {
		if asInt64(row[col]) == id {
			return row
		}
	}
	return nil
}

func (m *memDB) countWhereLocked(table string, match func(map[string]any) bool) int64 {
	var n int64
	for _, row := range m.tables[table] {
		if match(row) {
			n++
		}
	}
	return n
}

var insertRe = regexp.MustCompile(`(?is)INSERT\s+INTO\s+(?:public\.)?([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]+)\)\s*VALUES\s*\(`)

func parseInsert(sql string, args []any) (string, map[string]any, bool) {
	m := insertRe.FindStringSubmatch(sql)
	if len(m) != 3 {
		return "", nil, false
	}
	table := strings.ToLower(m[1])
	cols := strings.Split(m[2], ",")
	row := make(map[string]any, len(cols))
	for i, col := range cols {
		name := strings.ToLower(strings.TrimSpace(col))
		if i < len(args) {
			row[name] = args[i]
		}
	}
	return table, row, true
}

func tableFromSQL(sql string) string {
	re := regexp.MustCompile(`(?is)\b(?:INSERT\s+INTO|COPY|UPDATE|DELETE\s+FROM|TRUNCATE(?:\s+TABLE)?)\s+(?:ONLY\s+)?(?:public\.)?"?([A-Za-z_][A-Za-z0-9_]*)"?`)
	m := re.FindStringSubmatch(sql)
	if len(m) != 2 {
		return ""
	}
	return strings.ToLower(m[1])
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(strings.ToLower(sql)), " ")
}

func asInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint64:
		return int64(n)
	default:
		return 0
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func assignScan(dest, src any) error {
	switch d := dest.(type) {
	case *int64:
		*d = asInt64(src)
	case *string:
		*d = asString(src)
	case *bool:
		b, ok := src.(bool)
		if !ok {
			return fmt.Errorf("scan bool from %T", src)
		}
		*d = b
	default:
		return fmt.Errorf("unsupported scan dest %T", dest)
	}
	return nil
}
