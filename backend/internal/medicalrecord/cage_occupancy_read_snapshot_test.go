package medicalrecord

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// This in-memory driver records transaction options and SQL shape. It neither
// connects to PostgreSQL nor proves MVCC, RLS, locks or concurrent behavior.
type cageSnapshotFixture struct {
	begins    []driver.TxOptions
	queries   []string
	isolation string
	readOnly  bool
	commits   int
	rollbacks int
}

type cageSnapshotConnector struct{ fixture *cageSnapshotFixture }
type cageSnapshotDriver struct{}
type cageSnapshotConnection struct{ fixture *cageSnapshotFixture }
type cageSnapshotTransaction struct{ fixture *cageSnapshotFixture }
type cageSnapshotRows struct {
	columns []string
	values  [][]driver.Value
	next    int
}

func (c cageSnapshotConnector) Connect(context.Context) (driver.Conn, error) {
	return &cageSnapshotConnection{fixture: c.fixture}, nil
}

func (cageSnapshotConnector) Driver() driver.Driver { return cageSnapshotDriver{} }
func (cageSnapshotDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("snapshot fixture uses its connector")
}

func (*cageSnapshotConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepared statement in snapshot fixture")
}

func (*cageSnapshotConnection) Close() error { return nil }
func (c *cageSnapshotConnection) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *cageSnapshotConnection) BeginTx(_ context.Context, options driver.TxOptions) (driver.Tx, error) {
	c.fixture.begins = append(c.fixture.begins, options)
	c.fixture.readOnly = options.ReadOnly
	switch sql.IsolationLevel(options.Isolation) {
	case sql.LevelRepeatableRead:
		c.fixture.isolation = "repeatable read"
	case sql.LevelSerializable:
		c.fixture.isolation = "serializable"
	default:
		c.fixture.isolation = "read committed"
	}
	return cageSnapshotTransaction{fixture: c.fixture}, nil
}

func (c *cageSnapshotConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.fixture.queries = append(c.fixture.queries, query)
	switch {
	case strings.Contains(query, "current_setting('transaction_isolation')"):
		return &cageSnapshotRows{
			columns: []string{"isolation", "read_only"},
			values:  [][]driver.Value{{c.fixture.isolation, c.fixture.readOnly}},
		}, nil
	case strings.Contains(query, `"stay_cases"`):
		return &cageSnapshotRows{
			columns: []string{"id", "clinic_id", "stay_kind", "version"},
			values:  [][]driver.Value{{int64(11), int64(7), "hotel", int64(4)}},
		}, nil
	case strings.Contains(query, `"stay_movement_events"`):
		return &cageSnapshotRows{columns: []string{"id"}}, nil
	default:
		return nil, errors.New("unexpected query in snapshot fixture")
	}
}

func (t cageSnapshotTransaction) Commit() error {
	t.fixture.commits++
	return nil
}

func (t cageSnapshotTransaction) Rollback() error {
	t.fixture.rollbacks++
	return nil
}

func (r *cageSnapshotRows) Columns() []string { return r.columns }
func (*cageSnapshotRows) Close() error { return nil }
func (r *cageSnapshotRows) Next(dest []driver.Value) error {
	if r.next == len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.next])
	r.next++
	return nil
}

func cageSnapshotFixtureDB(t *testing.T, fixture *cageSnapshotFixture) *gorm.DB {
	t.Helper()
	pool := sql.OpenDB(cageSnapshotConnector{fixture: fixture})
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	return db
}

func TestCageOccupancySnapshot_RootReadOnlyAndReaderDoesNotLock(t *testing.T) {
	fixture := &cageSnapshotFixture{}
	db := cageSnapshotFixtureDB(t, fixture)
	reader := NewStayCageEffectiveStateReader(db)
	result, err := reader.ReadEffectiveStatesForCageAssignment(t.Context(), 7, []uint64{11})
	require.NoError(t, err)
	require.Len(t, fixture.begins, 1)
	require.Equal(t, driver.IsolationLevel(sql.LevelRepeatableRead), fixture.begins[0].Isolation)
	require.True(t, fixture.begins[0].ReadOnly)
	require.Equal(t, 1, fixture.commits)
	require.Len(t, result, 1)
	require.Equal(t, StayCageUnknown, result[0].State)
	require.Equal(t, StayCageEvidence{StayVersion: 4, MovementHeadRevision: 0}, result[0].Evidence)
	require.Len(t, fixture.queries, 2)
	for _, query := range fixture.queries {
		require.NotContains(t, query, "FOR SHARE")
	}
}

func TestCageOccupancySnapshot_SaveReaderKeepsAmbientTransactionAndShareLock(t *testing.T) {
	fixture := &cageSnapshotFixture{}
	db := cageSnapshotFixtureDB(t, fixture)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	ctx := persistence.WithTxValue(t.Context(), tx)
	_, err := NewStayCageEffectiveStateReader(db).ReadEffectiveStatesForCageAssignment(ctx, 7, []uint64{11})
	require.NoError(t, err)
	require.Len(t, fixture.begins, 1)
	require.False(t, fixture.begins[0].ReadOnly)
	require.Same(t, tx, persistence.TxFromContext(ctx))
	require.Equal(t, 0, fixture.commits)
	require.Len(t, fixture.queries, 2)
	require.Contains(t, fixture.queries[0], "FOR SHARE")
	require.NoError(t, tx.Rollback().Error)
}

func TestCageOccupancySnapshot_AmbientCandidatesNeverStartAnotherTransaction(t *testing.T) {
	for _, tc := range []struct {
		name      string
		isolation sql.IsolationLevel
		readOnly  bool
		wantError bool
	}{
		{"read committed refused", sql.LevelReadCommitted, false, true},
		{"write repeatable read retained", sql.LevelRepeatableRead, false, false},
		{"read-only repeatable read retained", sql.LevelRepeatableRead, true, false},
		{"serializable retained", sql.LevelSerializable, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := &cageSnapshotFixture{}
			db := cageSnapshotFixtureDB(t, fixture)
			tx := db.Begin(&sql.TxOptions{Isolation: tc.isolation, ReadOnly: tc.readOnly})
			require.NoError(t, tx.Error)
			ctx := persistence.WithTxValue(t.Context(), tx)
			called := false
			err := withCageOccupancyReadSnapshot(ctx, db, func(readCtx context.Context) error {
				called = true
				require.Same(t, tx, persistence.TxFromContext(readCtx))
				require.Equal(t, tc.readOnly, cageReadOnlySnapshot(readCtx))
				return nil
			})
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, !tc.wantError, called)
			require.Len(t, fixture.begins, 1)
			require.Equal(t, 0, fixture.commits)
			require.NoError(t, tx.Rollback().Error)
		})
	}
}

func TestCageOccupancySnapshot_TransactionBoundDatabaseNeedsAmbientContext(t *testing.T) {
	fixture := &cageSnapshotFixture{}
	db := cageSnapshotFixtureDB(t, fixture)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	called := false
	err := withCageOccupancyReadSnapshot(t.Context(), tx, func(context.Context) error {
		called = true
		return nil
	})
	require.Error(t, err)
	require.False(t, called)
	require.Len(t, fixture.begins, 1)
	require.Empty(t, fixture.queries)
	require.NoError(t, tx.Rollback().Error)
}

func TestCageOccupancySnapshot_ReadParticipantDoesNotNestAndFailureRollsBack(t *testing.T) {
	fixture := &cageSnapshotFixture{}
	db := cageSnapshotFixtureDB(t, fixture)
	failure := errors.New("read failed")
	err := withCageOccupancyReadSnapshot(t.Context(), db, func(ctx context.Context) error {
		tx := persistence.TxFromContext(ctx)
		return withCageOccupancyReadSnapshot(ctx, db, func(inner context.Context) error {
			require.Same(t, tx, persistence.TxFromContext(inner))
			require.True(t, cageReadOnlySnapshot(inner))
			return failure
		})
	})
	require.ErrorIs(t, err, failure)
	require.Len(t, fixture.begins, 1)
	require.Empty(t, fixture.queries)
	require.Equal(t, 0, fixture.commits)
	require.Equal(t, 1, fixture.rollbacks)
}

func TestCageOccupancySnapshot_ReadOnlyMarkerCannotMoveToAnotherTransaction(t *testing.T) {
	readTx, writeTx := &gorm.DB{}, &gorm.DB{}
	ctx := persistence.WithTxValue(t.Context(), readTx)
	ctx = context.WithValue(ctx, cageReadOnlySnapshotKey{}, readTx)
	require.True(t, cageReadOnlySnapshot(ctx))
	writeCtx := persistence.WithTxValue(ctx, writeTx)
	require.False(t, cageReadOnlySnapshot(writeCtx))
	require.Same(t, writeTx, persistence.TxFromContext(writeCtx))
}
