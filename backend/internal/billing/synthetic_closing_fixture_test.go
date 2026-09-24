package billing

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errRollbackSyntheticClosingCashTrigger = errors.New("rollback s09 cash trigger fixture")

func testdbSetupSyntheticClosing(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{},
		&model.Account{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{},
		&model.Billing{}, &model.BillingItem{}, &model.Payment{}, &model.PaymentSplit{},
		&model.PaymentMethodMaster{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	return db
}

func TestCreateSyntheticClosingFixture_RejectsUnsafeRequest(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)

	t.Run("staging env", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "staging", DBHost: "db", TargetDate: day, PasswordHash: "x",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "APP_ENV")
	})

	t.Run("existing billing ids", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "x", ExistingBillingIDs: []uint64{3},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "existing billing")
	})

	t.Run("empty password hash", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day,
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "password hash")
	})

	t.Run("weekend", func(t *testing.T) {
		saturday := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: saturday, PasswordHash: "x",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "weekday")
	})
}

func TestCreateSyntheticClosingFixture_CreatesFiveNewCompletedBillings(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)

	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "test-hash-not-for-login",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, RejectReservedClinicID(got.ClinicID))
	require.Len(t, got.BillingIDs, 5)
	require.Len(t, got.CompletedAt, 5)
	assert.Equal(t, SyntheticClosingLoginEmail(got.ClinicID), got.LoginEmail)
	assert.Equal(t, SyntheticClosingCleanupToken(got.ClinicID), got.CleanupToken)

	wantHours := [][2]int{{10, 0}, {13, 30}, {14, 0}, {20, 0}, {2, 0}}
	var persisted []model.Billing
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Order("completed_at ASC").Find(&persisted).Error)
	require.Len(t, persisted, 5)

	for i, b := range persisted {
		require.NotNil(t, b.CompletedAt)
		at := b.CompletedAt.In(jst)
		assert.Equal(t, wantHours[i][0], at.Hour(), "billing %d hour", i)
		assert.Equal(t, wantHours[i][1], at.Minute(), "billing %d minute", i)
		assert.Equal(t, model.BillingStatusCompleted, b.Status)
		assert.Equal(t, "s09-synthetic", b.Memo)
		assert.NotZero(t, b.ID)
	}
	assert.Equal(t, 8, persisted[4].CompletedAt.In(jst).Day(), "overnight EMG is next calendar day")

	var settings model.ClinicSettings
	require.NoError(t, db.WithContext(ctx).First(&settings, "clinic_id = ?", got.ClinicID).Error)
	assert.Equal(t, "09:00:00", settings.ClosingAmStart)
	assert.Equal(t, "13:30:00", settings.ClosingAmPmBoundary)
	assert.Equal(t, "19:00:00", settings.ClosingWeekdayEnd)

	var items []model.BillingItem
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Find(&items).Error)
	require.Len(t, items, 5)
	var splits []model.PaymentSplit
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Find(&splits).Error)
	require.Len(t, splits, 5)
	var account model.Account
	require.NoError(t, db.WithContext(ctx).Where("email = ?", got.LoginEmail).First(&account).Error)
	assert.True(t, account.IsSystemAdmin)

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
	var remaining int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Billing{}).Where("clinic_id = ?", got.ClinicID).Count(&remaining).Error)
	assert.Zero(t, remaining)
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
}

func TestInitSQL_ClinicInsertCreatesDefaultCashPaymentMethod(t *testing.T) {
	raw, err := os.ReadFile("../../migrations/001_init.sql") //nolint:gocritic // B5b requires this relative path.
	require.NoError(t, err)
	ddl := string(raw)
	assert.Contains(t, ddl, "CREATE TRIGGER trg_create_default_payment_methods")
	assert.Contains(t, ddl, "(NEW.id, '現金',            'cash',             1, true)")
	assert.Contains(t, ddl, "CREATE UNIQUE INDEX idx_payment_methods_clinic_system_key")
}

func TestCreateSyntheticClosingFixture_ReusesTriggerCreatedCash(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	var got *SyntheticClosingResult
	txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		require.NoError(t, tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_s09_test_payment_methods_clinic_system_key
			  ON payment_methods (clinic_id, system_key)
			  WHERE system_key IS NOT NULL AND deleted_at IS NULL
		`).Error)
		require.NoError(t, tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_s09_test_payment_methods_clinic_name
			  ON payment_methods (clinic_id, name)
			  WHERE deleted_at IS NULL
		`).Error)
		require.NoError(t, tx.Exec(`
			CREATE OR REPLACE FUNCTION s09_test_create_default_payment_methods()
			RETURNS TRIGGER AS $$
			BEGIN
			    INSERT INTO payment_methods (clinic_id, name, system_key, display_order, is_active, created_at, updated_at)
			    VALUES
			        (NEW.id, '現金',            'cash',             1, true, now(), now()),
			        (NEW.id, 'クレジットカード', 'credit_card',      2, true, now(), now()),
			        (NEW.id, '電子マネー',       'electronic_money', 3, true, now(), now()),
			        (NEW.id, '銀行振込',         'bank_transfer',    4, true, now(), now());
			    RETURN NEW;
			END;
			$$ LANGUAGE plpgsql
		`).Error)
		require.NoError(t, tx.Exec(`
			DROP TRIGGER IF EXISTS trg_s09_test_create_default_payment_methods ON clinics
		`).Error)
		require.NoError(t, tx.Exec(`
			CREATE TRIGGER trg_s09_test_create_default_payment_methods
			    AFTER INSERT ON clinics
			    FOR EACH ROW
			    EXECUTE FUNCTION s09_test_create_default_payment_methods()
		`).Error)

		created, createErr := CreateSyntheticClosingFixture(ctx, tx, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: time.Date(2026, 9, 7, 0, 0, 0, 0, jst), PasswordHash: "x",
		})
		if createErr != nil {
			return createErr
		}
		got = created

		var cashCount int64
		if err := tx.Model(&model.PaymentMethodMaster{}).
			Where("clinic_id = ? AND system_key = ?", got.ClinicID, "cash").
			Count(&cashCount).Error; err != nil {
			return err
		}
		if cashCount != 1 {
			return errors.New("expected exactly one cash payment method after trigger-backed setup")
		}
		var methodCount int64
		if err := tx.Model(&model.PaymentMethodMaster{}).Where("clinic_id = ?", got.ClinicID).Count(&methodCount).Error; err != nil {
			return err
		}
		if methodCount != 4 {
			return errors.New("expected trigger defaults only, without a duplicate cash insert")
		}
		return errRollbackSyntheticClosingCashTrigger
	})
	require.ErrorIs(t, txErr, errRollbackSyntheticClosingCashTrigger)
	require.NotNil(t, got)
	require.Len(t, got.BillingIDs, 5)
}

func TestDeleteSyntheticClosingFixture_RejectsWrongToken(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: time.Date(2026, 9, 7, 0, 0, 0, 0, jst), PasswordHash: "x",
	})
	require.NoError(t, err)
	err = DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, "deadbeef")
	require.Error(t, err)
	assert.ErrorContains(t, err, "cleanup token")
}

// EMR-210: teardown が cash_register_closes / cash_register_close_adjustments を含む
// 全ブロッキング FK 子孫を消し切ること。testdb は GORM double で実 RESTRICT/composite
// FK・append-only trigger を持たないため、ここでは系列の実行経路と削除結果を検証する。
// 実スキーマでの完走保証は MIGRATE_SQL_INTEGRATION の disposable DB テストと
// internal/lintscan のクロージャゲートが担う。
func TestDeleteSyntheticClosingFixture_RemovesCashRegisterCloseGraph(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.CashRegisterCloseAdjustment{}, &model.AuditLog{},
	))
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)

	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "x",
	})
	require.NoError(t, err)

	var staffRow model.Staff
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).First(&staffRow).Error)

	close := &model.CashRegisterClose{
		ClinicID:  got.ClinicID,
		CloseDate: day,
		Period:    "am",
		ClosedBy:  &staffRow.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(close).Error)
	adjustment := &model.CashRegisterCloseAdjustment{
		ClinicID:  got.ClinicID,
		CloseID:   close.ID,
		BillingID: got.BillingIDs[0],
		Reason:    "s09 teardown regression",
		ActorID:   &staffRow.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(adjustment).Error)

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))

	for name, modelPtr := range map[string]any{
		"cash_register_close_adjustments": &model.CashRegisterCloseAdjustment{},
		"cash_register_closes":            &model.CashRegisterClose{},
		"payment_splits":                  &model.PaymentSplit{},
		"payments":                        &model.Payment{},
		"billing_items":                   &model.BillingItem{},
		"billings":                        &model.Billing{},
		"owners":                          &model.Owner{},
		"pets":                            &model.Pet{},
		"staffs":                          &model.Staff{},
	} {
		var remaining int64
		require.NoError(t, db.WithContext(ctx).Model(modelPtr).Unscoped().Where("clinic_id = ?", got.ClinicID).Count(&remaining).Error)
		assert.Zero(t, remaining, "%s rows must be removed", name)
	}
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
}

// EMR-211 の受け口: auditRowsPolicy が teardown tx 内で呼ばれ、その失敗は
// teardown 全体をロールバックさせる。既定（nil）は audit 行を温存する。
func TestDeleteSyntheticClosingFixture_AuditRowsPolicySeam(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)

	t.Run("policy error aborts teardown", func(t *testing.T) {
		got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "x",
		})
		require.NoError(t, err)

		policyErr := errors.New("audit policy refused")
		err = DeleteSyntheticClosingFixtureWithAuditPolicy(ctx, db, "development", "db", got.ClinicID, got.CleanupToken,
			func(context.Context, *gorm.DB, uint64) error { return policyErr })
		require.ErrorIs(t, err, policyErr)

		// ロールバックされるので clinic は残る。
		require.NoError(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
	})

	t.Run("policy resolves audit rows inside tx", func(t *testing.T) {
		got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "x",
		})
		require.NoError(t, err)
		var staffRow model.Staff
		require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).First(&staffRow).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.AuditLog{
			ClinicID: &got.ClinicID, ActorID: &staffRow.ID,
			ActorType: "staff", Action: "login", Resource: "session",
		}).Error)

		var sawClinic uint64
		err = DeleteSyntheticClosingFixtureWithAuditPolicy(ctx, db, "development", "db", got.ClinicID, got.CleanupToken,
			func(pctx context.Context, tx *gorm.DB, clinicID uint64) error {
				sawClinic = clinicID
				return tx.Exec("DELETE FROM audit_logs WHERE clinic_id = ?", clinicID).Error
			})
		require.NoError(t, err)
		assert.Equal(t, got.ClinicID, sawClinic)
		require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
	})
}

// ---------------------------------------------------------------------------
// EMR-72 オフライン ambient-tx 証跡
//
// testdb 系テストは -short のオフライン検証で skip されるため、transaction の
// 接続ルーティングは記録型の database/sql driver で直接証明する。各 statement が
// どの driver.Conn 上で実行されたかを記録し、teardown トランザクションが ambient
// tx を逸脱して別接続で走る故障（委譲 cleanup が ambient tx に逃げる残存 bug）を
// DB なしで検出する。
// ---------------------------------------------------------------------------

var errRollbackSyntheticClosingAmbient = errors.New("rollback s09 ambient teardown fixture")

// s09FakeEntry は driver レベルで観測した 1 操作の記録。
type s09FakeEntry struct {
	connID int
	op     string // begin|exec|query|commit|rollback
	sql    string
	args   []driver.NamedValue
}

// s09FakeDriver は接続ごとの statement を記録する最小 driver.Driver。
type s09FakeDriver struct {
	mu        sync.Mutex
	nextID    int
	entries   []s09FakeEntry
	responder func(query string, args []driver.NamedValue) driver.Rows
}

func (d *s09FakeDriver) record(connID int, op, query string, args []driver.NamedValue) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = append(d.entries, s09FakeEntry{
		connID: connID, op: op, sql: query,
		args: append([]driver.NamedValue(nil), args...),
	})
}

// Open は driver.Driver インターフェースのためだけに存在する。sql.OpenDB は
// connector 経由でしか使わない。
func (d *s09FakeDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("s09 fake driver: direct Open is unsupported")
}

func (d *s09FakeDriver) ops(op string) []s09FakeEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []s09FakeEntry
	for _, e := range d.entries {
		if e.op == op {
			out = append(out, e)
		}
	}
	return out
}

func (d *s09FakeDriver) statements() []s09FakeEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []s09FakeEntry
	for _, e := range d.entries {
		if e.op == "exec" || e.op == "query" {
			out = append(out, e)
		}
	}
	return out
}

func (d *s09FakeDriver) hasSQL(connID int, needle string) bool {
	for _, e := range d.statements() {
		if e.connID == connID && strings.Contains(strings.ToUpper(e.sql), strings.ToUpper(needle)) {
			return true
		}
	}
	return false
}

type s09FakeConnector struct{ drv *s09FakeDriver }

func (c *s09FakeConnector) Connect(context.Context) (driver.Conn, error) {
	c.drv.mu.Lock()
	defer c.drv.mu.Unlock()
	c.drv.nextID++
	return &s09FakeConn{id: c.drv.nextID, drv: c.drv}, nil
}

func (c *s09FakeConnector) Driver() driver.Driver { return c.drv }

type s09FakeConn struct {
	id   int
	drv  *s09FakeDriver
	inTx bool
}

func (c *s09FakeConn) Prepare(query string) (driver.Stmt, error) {
	return &s09FakeStmt{conn: c, sql: query}, nil
}

func (c *s09FakeConn) Close() error { return nil }

func (c *s09FakeConn) Begin() (driver.Tx, error) { return c.begin() }

func (c *s09FakeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}

func (c *s09FakeConn) begin() (driver.Tx, error) {
	c.drv.record(c.id, "begin", "", nil)
	c.inTx = true
	return &s09FakeTx{conn: c}, nil
}

func (c *s09FakeConn) Ping(context.Context) error { return nil }

// CheckNamedValue は raw 値（uint64 等）をそのまま通す。
func (c *s09FakeConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *s09FakeConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.drv.record(c.id, "exec", query, args)
	return s09FakeResult(0), nil
}

func (c *s09FakeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.drv.record(c.id, "query", query, args)
	c.drv.mu.Lock()
	responder := c.drv.responder
	c.drv.mu.Unlock()
	if responder == nil {
		return &s09FakeRows{cols: []string{"id"}}, nil
	}
	return responder(query, args), nil
}

type s09FakeStmt struct {
	conn *s09FakeConn
	sql  string
}

func (s *s09FakeStmt) Close() error  { return nil }
func (s *s09FakeStmt) NumInput() int { return -1 }

func (s *s09FakeStmt) named(args []driver.Value) []driver.NamedValue {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return named
}

func (s *s09FakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.ExecContext(context.Background(), s.sql, s.named(args))
}

func (s *s09FakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.sql, s.named(args))
}

type s09FakeTx struct{ conn *s09FakeConn }

func (t *s09FakeTx) Commit() error {
	t.conn.drv.record(t.conn.id, "commit", "", nil)
	t.conn.inTx = false
	return nil
}

func (t *s09FakeTx) Rollback() error {
	t.conn.drv.record(t.conn.id, "rollback", "", nil)
	t.conn.inTx = false
	return nil
}

type s09FakeResult int64

func (r s09FakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r s09FakeResult) RowsAffected() (int64, error) { return int64(r), nil }

type s09FakeRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

func (r *s09FakeRows) Columns() []string { return r.cols }
func (r *s09FakeRows) Close() error      { return nil }

func (r *s09FakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

var s09ReturningColRe = regexp.MustCompile(`"(\w+)"`)

// s09FakeReturnValue は RETURNING 列名から database/sql が scan 可能な
// driver.Value を選ぶ。配列列（pq.StringArray 等）は "{}" リテラル、
// 時刻列は time.Time、bool 列は true、それ以外は int64 を返す。
func s09FakeReturnValue(col string, seq int64) driver.Value {
	switch {
	case strings.HasSuffix(col, "_order"), strings.HasSuffix(col, "_ids"),
		strings.HasSuffix(col, "_list"), strings.HasSuffix(col, "_tags"),
		strings.HasSuffix(col, "_weekdays"):
		return "{}"
	case strings.HasSuffix(col, "_at"), strings.HasSuffix(col, "_date"),
		strings.HasSuffix(col, "_time"), col == "birthday":
		return time.Now()
	case strings.HasPrefix(col, "is_"), strings.HasPrefix(col, "show_"),
		strings.HasPrefix(col, "has_"), strings.HasPrefix(col, "include_"),
		strings.HasSuffix(col, "_flag"), col == "active":
		return true
	default:
		return int64(600000 + seq)
	}
}

// s09FakeReturningRows は INSERT/DELETE ... RETURNING "a","b" を受理し、
// 要求された列名をエコーするダミー行を返す。
func s09FakeReturningRows(query string) driver.Rows {
	idx := strings.Index(strings.ToUpper(query), "RETURNING")
	cols := []string{"id"}
	if idx >= 0 {
		var parsed []string
		for _, m := range s09ReturningColRe.FindAllStringSubmatch(query[idx:], -1) {
			parsed = append(parsed, m[1])
		}
		if len(parsed) > 0 {
			cols = parsed
		}
	}
	vals := make([]driver.Value, len(cols))
	for i, col := range cols {
		vals[i] = s09FakeReturnValue(col, int64(i))
	}
	return &s09FakeRows{cols: cols, rows: [][]driver.Value{vals}}
}

// s09FakeColumnRows は teardown の既存列判定が全削除系列を実行するよう、
// syntheticClosingDeleteStatements / append-only テーブル / exams の列を
// information_schema 相当の行として返す。
func s09FakeColumnRows() [][]driver.Value {
	seen := map[[2]string]bool{}
	var rows [][]driver.Value
	add := func(table, column string) {
		key := [2]string{table, column}
		if seen[key] {
			return
		}
		seen[key] = true
		rows = append(rows, []driver.Value{table, column})
	}
	for _, stmt := range syntheticClosingDeleteStatements {
		m := syntheticClosingDeleteTargetRe.FindStringSubmatch(stmt)
		if len(m) >= 3 {
			add(m[1], m[2])
		}
	}
	for _, table := range syntheticClosingAppendOnlyTables {
		add(table, "clinic_id")
	}
	add("exams", "current_revision_version")
	return rows
}

// s09TeardownResponder は teardown 経路の SELECT に対し、合成 clinic・staff・
// スキーマ列を返す。RETURNING は列名エコー、その他の SELECT は空行
// （ErrRecordNotFound 相当）を返す。
func s09TeardownResponder(clinicID uint64) func(string, []driver.NamedValue) driver.Rows {
	return func(query string, _ []driver.NamedValue) driver.Rows {
		upper := strings.ToUpper(query)
		switch {
		case strings.Contains(upper, "INFORMATION_SCHEMA.COLUMNS"):
			return &s09FakeRows{cols: []string{"table_name", "column_name"}, rows: s09FakeColumnRows()}
		case strings.Contains(upper, "TO_REGCLASS"):
			return &s09FakeRows{cols: []string{"bool"}, rows: [][]driver.Value{{true}}}
		case strings.Contains(upper, "RETURNING"):
			return s09FakeReturningRows(query)
		case strings.Contains(upper, `FROM "CLINICS"`):
			return &s09FakeRows{
				cols: []string{"id", "name", "company_id", "is_active", "created_at", "updated_at"},
				rows: [][]driver.Value{{
					int64(clinicID),
					fmt.Sprintf("%s%d", syntheticClosingClinicPrefix, clinicID),
					int64(910001), true, time.Now(), time.Now(),
				}},
			}
		case strings.Contains(upper, `FROM "STAFFS"`):
			return &s09FakeRows{
				cols: []string{"id", "clinic_id", "account_id", "name", "is_active"},
				rows: [][]driver.Value{{
					int64(555001), int64(clinicID), int64(777001), "s09-staff", true,
				}},
			}
		default:
			return &s09FakeRows{cols: []string{"id"}}
		}
	}
}

// s09CreateResponder は create 経路向けに RETURNING をエコーし、
// SELECT（payment_methods の Take 等）は空行を返す。
func s09CreateResponder() func(string, []driver.NamedValue) driver.Rows {
	return func(query string, _ []driver.NamedValue) driver.Rows {
		if strings.Contains(strings.ToUpper(query), "RETURNING") {
			return s09FakeReturningRows(query)
		}
		return &s09FakeRows{cols: []string{"id"}}
	}
}

func s09OpenFakeGorm(t *testing.T) (*s09FakeDriver, *gorm.DB) {
	t.Helper()
	drv := &s09FakeDriver{}
	sqlDB := sql.OpenDB(&s09FakeConnector{drv: drv})
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	return drv, db
}

// requireSingleTxConn は begin が 1 回だけ記録され、全 statement がその接続で
// 実行されたことを検証する（teardown tx からの逸脱 = 2 回目の begin または
// 別 connID の statement として観測される）。
func s09RequireSingleTxConn(t *testing.T, drv *s09FakeDriver) int {
	t.Helper()
	begins := drv.ops("begin")
	require.Len(t, begins, 1, "exactly one transaction must be opened; a second Begin means the fixture tx escaped the ambient tx")
	connID := begins[0].connID
	for _, e := range drv.statements() {
		assert.Equal(t, connID, e.connID, "statement ran off the transaction connection: %q", e.sql)
	}
	return connID
}

// s09RequirePhaseSingleTx は 1 phase 分の entries が begin 1 回・単一接続・
// commit で終わることを検証する。
func s09RequirePhaseSingleTx(t *testing.T, entries []s09FakeEntry, phase string) {
	t.Helper()
	var beginConn, begins, stmts, commits int
	for _, e := range entries {
		switch e.op {
		case "begin":
			begins++
			beginConn = e.connID
		case "exec", "query":
			stmts++
			assert.Equal(t, beginConn, e.connID, "%s statement escaped the transaction connection: %q", phase, e.sql)
		case "commit":
			commits++
		}
	}
	require.Equal(t, 1, begins, "%s: expected exactly one begin", phase)
	require.Equal(t, 1, commits, "%s: expected exactly one commit", phase)
	assert.Positive(t, stmts, "%s: expected statements", phase)
}

// EMR-72: ambient tx が ctx にあるとき、teardown は別接続で tx を開かず
// ambient tx に SAVEPOINT で join し、委譲 cleanup もその tx で実行する。
func TestDeleteSyntheticClosingFixture_AmbientTxJoinsSavepoint(t *testing.T) {
	ctx := context.Background()
	drv, db := s09OpenFakeGorm(t)
	clinicID := uint64(920424)
	drv.responder = s09TeardownResponder(clinicID)

	ambient := db.WithContext(ctx).Begin()
	require.NoError(t, ambient.Error)

	var policyTx, policyCtxTx *gorm.DB
	err := DeleteSyntheticClosingFixtureWithAuditPolicy(
		persistence.WithTxValue(ctx, ambient), db, "development", "db", clinicID,
		SyntheticClosingCleanupToken(clinicID),
		func(pctx context.Context, tx *gorm.DB, _ uint64) error {
			policyTx = tx
			policyCtxTx = persistence.TxFromContext(pctx)
			return nil
		})
	require.NoError(t, err)
	require.NoError(t, ambient.Rollback().Error)

	connID := s09RequireSingleTxConn(t, drv)
	assert.True(t, drv.hasSQL(connID, "SAVEPOINT"),
		"teardown must join the ambient tx via SAVEPOINT instead of opening a separate connection")
	require.Same(t, policyTx, policyCtxTx,
		"delegated seams must resolve the teardown tx session via WithTxValue, not the caller ambient tx")
	require.NotSame(t, ambient, policyCtxTx)
	for _, table := range []string{"appointments", "reservation_types", "shift_entries", "staffs"} {
		assert.True(t,
			drv.hasSQL(connID, fmt.Sprintf("DELETE FROM %s", table)) ||
				drv.hasSQL(connID, fmt.Sprintf("DELETE FROM \"%s\"", table)),
			"delegated cleanup for %s must run on the teardown transaction connection", table)
	}
}

// EMR-72: create 経路も同じ契約 — ambient tx があるとき SAVEPOINT join し、
// write-owner の staff insert を含む全 INSERT がその接続に留まる。
func TestCreateSyntheticClosingFixture_AmbientTxJoinsSavepoint(t *testing.T) {
	ctx := context.Background()
	drv, db := s09OpenFakeGorm(t)
	drv.responder = s09CreateResponder()

	ambient := db.WithContext(ctx).Begin()
	require.NoError(t, ambient.Error)

	got, err := CreateSyntheticClosingFixture(
		persistence.WithTxValue(ctx, ambient), db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db",
			TargetDate:   time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC),
			PasswordHash: "x",
		})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, ambient.Rollback().Error)

	connID := s09RequireSingleTxConn(t, drv)
	assert.True(t, drv.hasSQL(connID, "SAVEPOINT"),
		"create must join the ambient tx via SAVEPOINT")
	assert.True(t, drv.hasSQL(connID, `INSERT INTO "staffs"`),
		"write-owner staff insert must run inside the fixture transaction")
	assert.True(t, drv.hasSQL(connID, `INSERT INTO "billings"`))
}

// ambient tx がない経路では新規 tx を 1 本だけ開き、全 statement がその接続で
// commit される（既存挙動の維持）。
func TestDeleteSyntheticClosingFixture_PlainCtxSingleTxConnection(t *testing.T) {
	ctx := context.Background()
	drv, db := s09OpenFakeGorm(t)
	clinicID := uint64(920555)
	drv.responder = s09TeardownResponder(clinicID)

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", clinicID,
		SyntheticClosingCleanupToken(clinicID)))

	connID := s09RequireSingleTxConn(t, drv)
	assert.Equal(t, 1, len(drv.ops("commit")), "the teardown transaction must commit")
	assert.False(t, drv.hasSQL(connID, "SAVEPOINT"),
		"without an ambient tx the fixture opens a fresh transaction, not a savepoint")
	assert.True(t, drv.hasSQL(connID, `DELETE FROM "clinics"`))
}

// EMR-72 A2 のオフライン証跡: RegisterUATRoutes の実 HTTP 経路で POST→DELETE を
// 流し、POST が作成した全テーブルに対して DELETE フェーズの teardown が DELETE を
// 発行し、全 write が各フェーズ 1 本の tx 接続に留まることを検証する。
// （行レベルの残存ゼロ証明は testdb 版 TestRegisterUATRoutes_DeleteLeavesNoResidue が担う。）
func TestRegisterUATRoutes_TeardownCoversCreatedTables(t *testing.T) {
	gin.SetMode(gin.TestMode)
	drv, db := s09OpenFakeGorm(t)
	drv.responder = s09CreateResponder()
	h := &SyntheticClosingHandler{
		DB: db, AppEnv: "development", DBHost: "db", Password: "s09-local-password",
	}
	r := gin.New()
	RegisterUATRoutes(r.Group("/api/v1"), h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uat/synthetic-closings",
		bytes.NewBufferString(`{"targetDate":"2026-09-07"}`))
	req.Host = "localhost"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var body syntheticClosingHTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotZero(t, body.ClinicID)

	drv.responder = s09TeardownResponder(body.ClinicID)
	postEnd := len(drv.entries)

	del := httptest.NewRequest(http.MethodDelete,
		"/api/v1/uat/synthetic-closings/"+strconv.FormatUint(body.ClinicID, 10), nil)
	del.Host = "127.0.0.1"
	del.Header.Set(syntheticClosingCleanupHeader, body.CleanupToken)
	dw := httptest.NewRecorder()
	r.ServeHTTP(dw, del)
	require.Equal(t, http.StatusNoContent, dw.Code, dw.Body.String())

	insertRe := regexp.MustCompile(`^INSERT INTO "?(\w+)"?`)
	deleteRe := regexp.MustCompile(`^DELETE FROM "?(\w+)"?`)
	inserted := map[string]bool{}
	deleted := map[string]bool{}
	for i, e := range drv.entries {
		sql := strings.TrimSpace(e.sql)
		if i < postEnd {
			if m := insertRe.FindStringSubmatch(sql); m != nil {
				inserted[m[1]] = true
			}
			continue
		}
		if m := deleteRe.FindStringSubmatch(sql); m != nil {
			deleted[m[1]] = true
		}
	}
	for table := range inserted {
		assert.True(t, deleted[table],
			"teardown issued no DELETE for fixture-created table %s", table)
	}
	s09RequirePhaseSingleTx(t, drv.entries[postEnd:], "DELETE /synthetic-closings")
}

// EMR-72 A1 の実 DB 版: ambient tx 内の teardown が ambient rollback で完全に
// 復元されること（委譲 cleanup が ambient/teardown tx に join し、外部へ commit
// しないこと）と、commit 済み teardown が fixture 行を残さないことを検証する。
// testdb は -short で skip される。
func TestDeleteSyntheticClosingFixture_AmbientTxJoinLeavesNoResidue(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db",
		TargetDate:   time.Date(2026, 9, 7, 0, 0, 0, 0, jst),
		PasswordHash: "x",
	})
	require.NoError(t, err)

	txErr := db.WithContext(ctx).Transaction(func(ambient *gorm.DB) error {
		if err := DeleteSyntheticClosingFixtureWithAuditPolicy(
			persistence.WithTxValue(ctx, ambient), db,
			"development", "db", got.ClinicID, got.CleanupToken, nil); err != nil {
			return err
		}
		return errRollbackSyntheticClosingAmbient
	})
	require.ErrorIs(t, txErr, errRollbackSyntheticClosingAmbient)

	var clinic model.Clinic
	require.NoError(t, db.WithContext(ctx).First(&clinic, got.ClinicID).Error,
		"ambient rollback must restore the synthetic clinic")
	var billingCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Billing{}).
		Where("clinic_id = ?", got.ClinicID).Count(&billingCount).Error)
	assert.Equal(t, int64(5), billingCount,
		"ambient rollback must restore all fixture billings")

	require.NoError(t, db.WithContext(ctx).Transaction(func(ambient *gorm.DB) error {
		return DeleteSyntheticClosingFixtureWithAuditPolicy(
			persistence.WithTxValue(ctx, ambient), db,
			"development", "db", got.ClinicID, got.CleanupToken, nil)
	}))
	for name, modelPtr := range map[string]any{
		"payment_splits": &model.PaymentSplit{},
		"payments":       &model.Payment{},
		"billing_items":  &model.BillingItem{},
		"billings":       &model.Billing{},
		"owners":         &model.Owner{},
		"pets":           &model.Pet{},
		"staffs":         &model.Staff{},
	} {
		var remaining int64
		require.NoError(t, db.WithContext(ctx).Model(modelPtr).Unscoped().
			Where("clinic_id = ?", got.ClinicID).Count(&remaining).Error)
		assert.Zero(t, remaining, "%s rows must be removed", name)
	}
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
}

// EMR-72 A2 の実 DB 版: RegisterUATRoutes の POST→DELETE が 204 を返し、
// fixture 保有テーブル全件が残らないことを行レベルで証明する。
// testdb は -short で skip される。
func TestRegisterUATRoutes_DeleteLeavesNoResidue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdbSetupSyntheticClosing(t)
	h := &SyntheticClosingHandler{
		DB: db, AppEnv: "development", DBHost: "db", Password: "s09-local-password",
	}
	r := gin.New()
	RegisterUATRoutes(r.Group("/api/v1"), h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uat/synthetic-closings",
		bytes.NewBufferString(`{"targetDate":"2026-09-07"}`))
	req.Host = "localhost"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var body syntheticClosingHTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotZero(t, body.ClinicID)

	del := httptest.NewRequest(http.MethodDelete,
		"/api/v1/uat/synthetic-closings/"+strconv.FormatUint(body.ClinicID, 10), nil)
	del.Host = "127.0.0.1"
	del.Header.Set(syntheticClosingCleanupHeader, body.CleanupToken)
	dw := httptest.NewRecorder()
	r.ServeHTTP(dw, del)
	require.Equal(t, http.StatusNoContent, dw.Code, dw.Body.String())

	ctx := context.Background()
	for name, modelPtr := range map[string]any{
		"billings":                 &model.Billing{},
		"billing_items":            &model.BillingItem{},
		"payments":                 &model.Payment{},
		"payment_splits":           &model.PaymentSplit{},
		"owners":                   &model.Owner{},
		"pets":                     &model.Pet{},
		"staffs":                   &model.Staff{},
		"staff_clinic_assignments": &model.StaffClinicAssignment{},
		"payment_methods":          &model.PaymentMethodMaster{},
	} {
		var remaining int64
		require.NoError(t, db.WithContext(ctx).Model(modelPtr).Unscoped().
			Where("clinic_id = ?", body.ClinicID).Count(&remaining).Error)
		assert.Zero(t, remaining, "%s rows must be removed", name)
	}
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, body.ClinicID).Error,
		"synthetic clinic row must be gone")
	var companyCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Company{}).Unscoped().
		Where("name LIKE ?", syntheticClosingCompanyPrefix+"%").Count(&companyCount).Error)
	assert.Zero(t, companyCount, "synthetic company rows must be removed")
}
