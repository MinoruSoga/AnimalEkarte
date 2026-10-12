package medicalrecord

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type hospSnapshotFixture struct {
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

type hospSnapshotConnector struct{ fixture *hospSnapshotFixture }

func (c hospSnapshotConnector) Connect(context.Context) (driver.Conn, error) {
	return &hospSnapshotConn{fixture: c.fixture}, nil
}
func (c hospSnapshotConnector) Driver() driver.Driver { return hospSnapshotDriver{} }

type hospSnapshotDriver struct{}

func (hospSnapshotDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type hospSnapshotConn struct{ fixture *hospSnapshotFixture }

func (c *hospSnapshotConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare forbidden")
}
func (c *hospSnapshotConn) Close() error              { return nil }
func (c *hospSnapshotConn) Begin() (driver.Tx, error) { return nil, errors.New("must use BeginTx") }
func (c *hospSnapshotConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.fixture.active {
		return nil, errors.New("nested transaction")
	}
	c.fixture.active = true
	c.fixture.began = true
	c.fixture.options = options
	return hospSnapshotTx{fixture: c.fixture}, nil
}
func (c *hospSnapshotConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
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
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	switch {
	case strings.Contains(query, "count(*)"):
		return &hospSnapshotRows{columns: []string{"count"}, values: [][]driver.Value{{c.fixture.count}}}, nil
	case strings.Contains(query, `SELECT * FROM "hospitalizations"`):
		return &hospSnapshotRows{
			columns: []string{
				"id", "clinic_id", "owner_id", "pet_id", "cage_id", "doctor_id",
				"hospitalization_type", "status", "start_date", "end_date",
			},
			values: [][]driver.Value{
				{int64(1), int64(1), int64(10), int64(5), int64(7), int64(3), "hospitalization", "admitted", day, day.Add(72 * time.Hour)},
				{int64(2), int64(1), int64(10), int64(5), int64(7), int64(3), "hospitalization", "admitted", day, day.Add(72 * time.Hour)},
			},
		}, nil
	case strings.Contains(query, `FROM "pets"`):
		return &hospSnapshotRows{
			columns: []string{"id", "clinic_id", "owner_id", "animal_species_id", "name"},
			values:  [][]driver.Value{{int64(5), int64(1), int64(10), int64(9), "fixture pet"}},
		}, nil
	case strings.Contains(query, `FROM "animal_species"`):
		return &hospSnapshotRows{
			columns: []string{"id", "name"},
			values:  [][]driver.Value{{int64(9), "fixture species"}},
		}, nil
	case strings.Contains(query, `FROM "owners"`):
		return &hospSnapshotRows{
			columns: []string{"id", "clinic_id", "name"},
			values:  [][]driver.Value{{int64(10), int64(1), "fixture owner"}},
		}, nil
	case strings.Contains(query, `FROM "cages"`):
		return &hospSnapshotRows{
			columns: []string{"id", "clinic_id", "name"},
			values:  [][]driver.Value{{int64(7), int64(1), "fixture cage"}},
		}, nil
	case strings.Contains(query, `FROM "staffs"`):
		return &hospSnapshotRows{
			columns: []string{"id", "clinic_id", "name"},
			values:  [][]driver.Value{{int64(3), int64(1), "fixture doctor"}},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected fixture query: %s", query)
	}
}

type hospSnapshotTx struct{ fixture *hospSnapshotFixture }

func (tx hospSnapshotTx) Commit() error {
	tx.fixture.active = false
	tx.fixture.committed = true
	return nil
}
func (tx hospSnapshotTx) Rollback() error {
	tx.fixture.active = false
	tx.fixture.rolledBack = true
	return nil
}

type hospSnapshotRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *hospSnapshotRows) Columns() []string { return r.columns }
func (r *hospSnapshotRows) Close() error      { return nil }
func (r *hospSnapshotRows) Next(dest []driver.Value) error {
	if r.index == len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func hospSnapshotFixtureRepo(t *testing.T, fixture *hospSnapshotFixture) hospitalizationSnapshotFinder {
	t.Helper()
	db := sql.OpenDB(hospSnapshotConnector{fixture: fixture})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return &hospitalizationRepository{db: gdb}
}

func hospSnapshotFilters() listHospitalizationFilters {
	petID := uint64(5)
	ownerID := uint64(10)
	status := "admitted"
	return listHospitalizationFilters{PetID: &petID, OwnerID: &ownerID, Status: &status}
}

func TestHospitalizationSnapshotRepository(t *testing.T) {
	t.Run("count rows and preloads share read-only repeatable read", func(t *testing.T) {
		fixture := &hospSnapshotFixture{count: 2}
		repo := hospSnapshotFixtureRepo(t, fixture)
		rows, total, err := repo.findSnapshot(context.Background(), 1, hospSnapshotFilters())
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, int64(2), total)
		assert.NotNil(t, rows[0].Pet)
		assert.NotNil(t, rows[0].Pet.AnimalSpecies)
		assert.NotNil(t, rows[0].Owner)
		assert.NotNil(t, rows[0].Cage)
		assert.NotNil(t, rows[0].Doctor)
		assert.Equal(t,
			driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelRepeatableRead), ReadOnly: true},
			fixture.options)
		assert.True(t, fixture.committed)
		assert.False(t, fixture.rolledBack)
		require.Len(t, fixture.queries, 7, "count + rows + 5 preloads must run in one transaction")
		for _, index := range []int{0, 1, 2, 4, 5} {
			assert.Contains(t, fixture.queries[index], "clinic_id")
		}
		require.Contains(t, fixture.queries[1], "LIMIT $")
		assert.NotContains(t, fixture.queries[1], "OFFSET")
		assert.Contains(t, fixture.queries[0], "pet_id = $")
		assert.Contains(t, fixture.queries[0], "status = $")
		require.Len(t, fixture.deadlines, len(fixture.queries))
		for _, deadline := range fixture.deadlines {
			assert.InDelta(t, 10, time.Until(deadline).Seconds(), 1)
		}
	})
	t.Run("filter values are bound on the same predicates as legacy list", func(t *testing.T) {
		fixture := &hospSnapshotFixture{count: 2}
		repo := hospSnapshotFixtureRepo(t, fixture)
		_, _, err := repo.findSnapshot(context.Background(), 1, hospSnapshotFilters())
		require.NoError(t, err)
		require.NotEmpty(t, fixture.args)
		var bound []any
		for _, callArgs := range fixture.args {
			for _, arg := range callArgs {
				bound = append(bound, arg.Value)
			}
		}
		assert.Contains(t, bound, int64(5))
		assert.Contains(t, bound, int64(10))
		assert.Contains(t, bound, "admitted")
	})
	t.Run("overflow rolls back without fetching rows", func(t *testing.T) {
		fixture := &hospSnapshotFixture{count: hospitalizationSnapshotMaxRows + 1}
		repo := hospSnapshotFixtureRepo(t, fixture)
		rows, _, err := repo.findSnapshot(context.Background(), 1, hospSnapshotFilters())
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.True(t, fixture.rolledBack)
		assert.False(t, fixture.committed)
		assert.Len(t, fixture.queries, 1)
	})
	for _, failAt := range []string{"count(*)", `SELECT * FROM "hospitalizations"`, `FROM "pets"`, `FROM "owners"`} {
		t.Run("read error "+failAt, func(t *testing.T) {
			fixture := &hospSnapshotFixture{count: 2, failAt: failAt}
			repo := hospSnapshotFixtureRepo(t, fixture)
			rows, _, err := repo.findSnapshot(context.Background(), 1, hospSnapshotFilters())
			require.Error(t, err)
			assert.Nil(t, rows)
			assert.True(t, fixture.rolledBack)
			assert.False(t, fixture.committed)
		})
	}
	t.Run("incomplete rows roll back instead of acknowledging a snapshot", func(t *testing.T) {
		fixture := &hospSnapshotFixture{count: 3}
		repo := hospSnapshotFixtureRepo(t, fixture)
		rows, _, err := repo.findSnapshot(context.Background(), 1, hospSnapshotFilters())
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.True(t, fixture.rolledBack)
		assert.False(t, fixture.committed)
	})
	t.Run("canceled caller never begins", func(t *testing.T) {
		fixture := &hospSnapshotFixture{count: 2}
		repo := hospSnapshotFixtureRepo(t, fixture)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		rows, _, err := repo.findSnapshot(ctx, 1, hospSnapshotFilters())
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, rows)
		assert.False(t, fixture.began)
	})
}
