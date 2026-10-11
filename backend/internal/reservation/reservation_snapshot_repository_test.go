package reservation

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// This driver observes connection/transaction use, not PostgreSQL MVCC semantics.
type snapshotFixture struct {
	count      int64
	queries    []string
	args       [][]driver.NamedValue
	began      bool
	active     bool
	options    driver.TxOptions
	committed  bool
	rolledBack bool
	failAt     string
	deadlines  []time.Time
}
type snapshotConnector struct{ fixture *snapshotFixture }

func (c snapshotConnector) Connect(context.Context) (driver.Conn, error) {
	return &snapshotConn{fixture: c.fixture}, nil
}
func (c snapshotConnector) Driver() driver.Driver { return snapshotDriver{} }

type snapshotDriver struct{}

func (snapshotDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type snapshotConn struct{ fixture *snapshotFixture }

func (c *snapshotConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare forbidden")
}
func (c *snapshotConn) Close() error              { return nil }
func (c *snapshotConn) Begin() (driver.Tx, error) { return nil, errors.New("must use BeginTx") }
func (c *snapshotConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.fixture.active {
		return nil, errors.New("nested transaction")
	}
	c.fixture.active = true
	c.fixture.began = true
	c.fixture.options = options
	return snapshotTx{fixture: c.fixture}, nil
}
func (c *snapshotConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !c.fixture.active {
		return nil, errors.New("query outside snapshot transaction")
	}
	c.fixture.queries = append(c.fixture.queries, query)
	c.fixture.args = append(c.fixture.args, args)
	if deadline, ok := ctx.Deadline(); ok {
		c.fixture.deadlines = append(c.fixture.deadlines, deadline)
	}
	if c.fixture.failAt != "" && strings.Contains(query, c.fixture.failAt) {
		return nil, errors.New("fixture read failed")
	}
	switch {
	case strings.Contains(query, "count(*)"):
		return &snapshotRows{columns: []string{"count"}, values: [][]driver.Value{{c.fixture.count}}}, nil
	case strings.Contains(query, `SELECT * FROM "appointments"`):
		at := time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)
		return &snapshotRows{columns: []string{"id", "clinic_id", "owner_id", "start_time"}, values: [][]driver.Value{{int64(1), int64(1), int64(10), at}, {int64(2), int64(1), int64(10), at}}}, nil
	case strings.Contains(query, `SELECT * FROM "owners"`):
		return &snapshotRows{columns: []string{"id", "clinic_id", "name"}, values: [][]driver.Value{{int64(10), int64(1), "fixture owner"}}}, nil
	default:
		return nil, fmt.Errorf("unexpected fixture query: %s", query)
	}
}

type snapshotTx struct{ fixture *snapshotFixture }

func (tx snapshotTx) Commit() error {
	tx.fixture.active = false
	tx.fixture.committed = true
	return nil
}
func (tx snapshotTx) Rollback() error {
	tx.fixture.active = false
	tx.fixture.rolledBack = true
	return nil
}

type snapshotRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *snapshotRows) Columns() []string { return r.columns }
func (r *snapshotRows) Close() error      { return nil }
func (r *snapshotRows) Next(dest []driver.Value) error {
	if r.index == len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
func snapshotFixtureRepo(t *testing.T, fixture *snapshotFixture) *reservationRepository {
	t.Helper()
	db := sql.OpenDB(snapshotConnector{fixture: fixture})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return &reservationRepository{db: gdb}
}
func snapshotDayFilters() listReservationFilters {
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	return listReservationFilters{Date: &day}
}
func TestReservationSnapshotRepository(t *testing.T) {
	t.Run("count appointments and preload share read-only repeatable read", func(t *testing.T) {
		fixture := &snapshotFixture{count: 2}
		repo := snapshotFixtureRepo(t, fixture)
		rows, err := repo.findSnapshot(context.Background(), []uint64{1}, snapshotDayFilters())
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.NotNil(t, rows[0].Owner)
		require.Equal(t, driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelRepeatableRead), ReadOnly: true}, fixture.options)
		require.True(t, fixture.committed)
		require.False(t, fixture.rolledBack)
		require.Len(t, fixture.queries, 3)
		for _, query := range fixture.queries {
			require.Contains(t, query, "clinic_id")
		}
		require.Contains(t, fixture.queries[1], "LIMIT $4")
		require.NotContains(t, fixture.queries[1], "OFFSET")
		require.Len(t, fixture.deadlines, 3)
		for _, deadline := range fixture.deadlines {
			require.InDelta(t, 10, time.Until(deadline).Seconds(), 1)
		}
	})
	t.Run("overflow rolls back without fetching partial appointments", func(t *testing.T) {
		fixture := &snapshotFixture{count: 10001}
		repo := snapshotFixtureRepo(t, fixture)
		rows, err := repo.findSnapshot(context.Background(), []uint64{1}, snapshotDayFilters())
		require.Error(t, err)
		require.Nil(t, rows)
		require.True(t, fixture.rolledBack)
		require.False(t, fixture.committed)
		require.Len(t, fixture.queries, 1)
	})
	for _, failAt := range []string{"count(*)", `SELECT * FROM "appointments"`, `SELECT * FROM "owners"`} {
		t.Run("read error "+failAt, func(t *testing.T) {
			fixture := &snapshotFixture{count: 2, failAt: failAt}
			repo := snapshotFixtureRepo(t, fixture)
			rows, err := repo.findSnapshot(context.Background(), []uint64{1}, snapshotDayFilters())
			require.Error(t, err)
			require.Nil(t, rows)
			require.True(t, fixture.rolledBack)
			require.False(t, fixture.committed)
		})
	}
	t.Run("incomplete rows roll back instead of acknowledging a snapshot", func(t *testing.T) {
		fixture := &snapshotFixture{count: 3}
		repo := snapshotFixtureRepo(t, fixture)
		rows, err := repo.findSnapshot(context.Background(), []uint64{1}, snapshotDayFilters())
		require.Error(t, err)
		require.Nil(t, rows)
		require.True(t, fixture.rolledBack)
		require.False(t, fixture.committed)
	})
	t.Run("canceled caller never begins", func(t *testing.T) {
		fixture := &snapshotFixture{count: 2}
		repo := snapshotFixtureRepo(t, fixture)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		rows, err := repo.findSnapshot(ctx, []uint64{1}, snapshotDayFilters())
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, rows)
		require.False(t, fixture.began)
	})
	t.Run("empty clinic scope exposes no rows", func(t *testing.T) {
		fixture := &snapshotFixture{count: 2}
		repo := snapshotFixtureRepo(t, fixture)
		rows, err := repo.findSnapshot(context.Background(), nil, snapshotDayFilters())
		require.NoError(t, err)
		require.Empty(t, rows)
		require.False(t, fixture.began)
	})
}
