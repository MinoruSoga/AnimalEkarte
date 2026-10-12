package reservation

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type transferFixture struct {
	petOwnerID      int64
	ownerLineUserID *string
	ownerConfirmed  bool
	targets         [][]driver.Value
	medicalHit      bool
	lineCustomerIDs []int64
	rowsAffected    int64
	failAt          string
	failCode        string
	queries         []string
	qargs           [][]driver.NamedValue
	execs           []string
	eargs           [][]driver.NamedValue
	began           bool
	active          bool
	committed       bool
	rolledBack      bool
}

type transferConnector struct{ fixture *transferFixture }

func (c transferConnector) Connect(context.Context) (driver.Conn, error) {
	return &transferConn{fixture: c.fixture}, nil
}
func (c transferConnector) Driver() driver.Driver { return transferDriver{} }

type transferDriver struct{}

func (transferDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type transferConn struct{ fixture *transferFixture }

func (c *transferConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare forbidden")
}
func (c *transferConn) Close() error              { return nil }
func (c *transferConn) Begin() (driver.Tx, error) { return nil, errors.New("must use BeginTx") }
func (c *transferConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.fixture.active {
		return nil, errors.New("nested transaction")
	}
	c.fixture.active = true
	c.fixture.began = true
	return transferTx{fixture: c.fixture}, nil
}
func (c *transferConn) fail(query string) error {
	if c.fixture.failAt != "" && strings.Contains(query, c.fixture.failAt) {
		if c.fixture.failCode != "" {
			return &pgconn.PgError{Code: c.fixture.failCode}
		}
		return errors.New("fixture failure")
	}
	return nil
}
func (c *transferConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !c.fixture.active {
		return nil, errors.New("query outside ambient transaction")
	}
	c.fixture.queries = append(c.fixture.queries, query)
	c.fixture.qargs = append(c.fixture.qargs, args)
	if err := c.fail(query); err != nil {
		return nil, err
	}
	f := c.fixture
	switch {
	case strings.Contains(query, `FROM "pets"`):
		return &snapshotRows{
			columns: []string{"id", "clinic_id", "owner_id"},
			values:  [][]driver.Value{{int64(2), int64(1), f.petOwnerID}},
		}, nil
	case strings.Contains(query, `FROM "owners"`):
		var lineUserID driver.Value
		if f.ownerLineUserID != nil {
			lineUserID = *f.ownerLineUserID
		}
		var confirmedAt driver.Value
		if f.ownerConfirmed {
			confirmedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		return &snapshotRows{
			columns: []string{"id", "clinic_id", "line_user_id", "line_id_confirmed_at"},
			values:  [][]driver.Value{{int64(30), int64(1), lineUserID, confirmedAt}},
		}, nil
	case strings.Contains(query, `FROM "appointments"`):
		return &snapshotRows{
			columns: []string{"id", "clinic_id", "pet_id", "owner_id", "line_customer_id", "status", "start_time", "customer_fields"},
			values:  f.targets,
		}, nil
	case strings.Contains(query, `FROM "medical_records"`):
		values := [][]driver.Value{}
		if f.medicalHit {
			values = [][]driver.Value{{int64(900)}}
		}
		return &snapshotRows{columns: []string{"id"}, values: values}, nil
	case strings.Contains(query, `FROM "line_customers"`):
		values := make([][]driver.Value, 0, len(f.lineCustomerIDs))
		for _, id := range f.lineCustomerIDs {
			values = append(values, []driver.Value{id})
		}
		return &snapshotRows{columns: []string{"id"}, values: values}, nil
	default:
		return nil, fmt.Errorf("unexpected fixture query: %s", query)
	}
}
func (c *transferConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !c.fixture.active {
		return nil, errors.New("exec outside ambient transaction")
	}
	c.fixture.execs = append(c.fixture.execs, query)
	c.fixture.eargs = append(c.fixture.eargs, args)
	if err := c.fail(query); err != nil {
		return nil, err
	}
	return transferResult{rows: c.fixture.rowsAffected}, nil
}

type transferResult struct{ rows int64 }

func (r transferResult) LastInsertId() (int64, error) { return 0, nil }
func (r transferResult) RowsAffected() (int64, error) { return r.rows, nil }

type transferTx struct{ fixture *transferFixture }

func (tx transferTx) Commit() error {
	tx.fixture.active = false
	tx.fixture.committed = true
	return nil
}
func (tx transferTx) Rollback() error {
	tx.fixture.active = false
	tx.fixture.rolledBack = true
	return nil
}

func transferFixtureGorm(t *testing.T, fixture *transferFixture) *gorm.DB {
	t.Helper()
	db := sql.OpenDB(transferConnector{fixture: fixture})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return gdb
}

var transferFrom = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

func transferTargetRow(id, ownerID int64, lineCustomerID driver.Value) []driver.Value {
	return []driver.Value{
		id, int64(1), int64(2), ownerID, lineCustomerID, "pending",
		transferFrom.Add(2 * time.Hour), []byte("{}"),
	}
}

func runTransfer(
	ctx context.Context,
	gdb *gorm.DB,
	transfer *PetOwnerTransfer,
) ([]uint64, error) {
	var got []uint64
	err := gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		got, err = transfer.TransferUpcomingPetReservations(
			persistence.WithTxValue(ctx, tx), 1, 2, 10, 30, transferFrom)
		return err
	})
	return got, err
}

func TestPetOwnerTransfer_TransfersUpcomingRowsInsideAmbientTx(t *testing.T) {
	fixture := &transferFixture{
		petOwnerID:   30,
		rowsAffected: 1,
		targets: [][]driver.Value{
			transferTargetRow(1, 10, int64(55)),
			transferTargetRow(2, 10, nil),
		},
	}
	gdb := transferFixtureGorm(t, fixture)
	got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, got)
	require.True(t, fixture.committed)
	require.False(t, fixture.rolledBack)

	require.Len(t, fixture.queries, 4)
	petQuery, ownerQuery, targetQuery, recordQuery := fixture.queries[0], fixture.queries[1], fixture.queries[2], fixture.queries[3]
	assert.Contains(t, petQuery, `FROM "pets"`)
	assert.Contains(t, petQuery, "clinic_id")
	assert.Contains(t, ownerQuery, `FROM "owners"`)
	assert.Contains(t, ownerQuery, "FOR SHARE NOWAIT")
	assert.Contains(t, targetQuery, `FROM "appointments"`)
	assert.Contains(t, targetQuery, "FOR UPDATE NOWAIT")
	assert.Contains(t, targetQuery, "start_time >=")
	assert.Contains(t, targetQuery, "status IN")
	assert.Contains(t, recordQuery, `FROM "medical_records"`)
	assert.Contains(t, recordQuery, "appointment_id IN")

	targetArgs := fixture.qargs[2]
	require.Len(t, targetArgs, 6)
	assert.Equal(t, int64(1), targetArgs[0].Value, "clinic scope")
	assert.Equal(t, int64(2), targetArgs[1].Value, "pet scope")
	assert.Equal(t, transferFrom, targetArgs[2].Value, "JST day boundary")
	assert.Equal(t, "pending", targetArgs[3].Value)
	assert.Equal(t, "confirmed", targetArgs[4].Value)
	assert.Equal(t, int64(petOwnerTransferMaxRows+1), targetArgs[5].Value)

	require.Len(t, fixture.execs, 2)
	for _, exec := range fixture.execs {
		assert.Contains(t, exec, `UPDATE "appointments"`)
		assert.Contains(t, exec, "owner_id")
		assert.Contains(t, exec, "updated_at")
	}
	assert.Contains(t, fixture.execs[0], "line_customer_id")
	assert.Contains(t, fixture.execs[0], "customer_fields")
	assert.NotContains(t, fixture.execs[1], "line_customer_id")
	assert.NotContains(t, fixture.execs[1], "customer_fields")

	firstArgs := fixture.eargs[0]
	var sawNewOwner bool
	for _, arg := range firstArgs {
		if v, ok := arg.Value.(int64); ok && v == 30 {
			sawNewOwner = true
		}
	}
	assert.True(t, sawNewOwner, "UPDATE must set owner_id to the new owner")
	guard := firstArgs[len(firstArgs)-7:]
	assert.Equal(t, int64(1), guard[0].Value, "guard binds clinic")
	assert.Equal(t, int64(1), guard[1].Value, "guard binds appointment id")
	assert.Equal(t, int64(2), guard[2].Value, "guard binds pet")
	assert.Equal(t, int64(10), guard[3].Value, "guard binds previous owner")
	assert.Equal(t, "pending", guard[4].Value)
	assert.Equal(t, "confirmed", guard[5].Value)
	assert.Equal(t, transferFrom, guard[6].Value, "guard rebinds the same boundary")
}

func TestPetOwnerTransfer_RequiresAmbientTransaction(t *testing.T) {
	fixture := &transferFixture{petOwnerID: 30, rowsAffected: 1}
	gdb := transferFixtureGorm(t, fixture)
	got, err := NewPetOwnerTransfer(gdb).TransferUpcomingPetReservations(
		context.Background(), 1, 2, 10, 30, transferFrom)
	require.Error(t, err)
	assert.Nil(t, got)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.Empty(t, fixture.queries)
	assert.False(t, fixture.began)
}

func TestPetOwnerTransfer_Guards(t *testing.T) {
	newFixture := func(mutate func(*transferFixture)) *transferFixture {
		f := &transferFixture{
			petOwnerID:   30,
			rowsAffected: 1,
			targets:      [][]driver.Value{transferTargetRow(1, 10, nil)},
		}
		if mutate != nil {
			mutate(f)
		}
		return f
	}

	t.Run("pet not yet owned by new owner conflicts", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) { f.petOwnerID = 10 })
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, fixture.execs)
		assert.True(t, fixture.rolledBack)
	})

	t.Run("appointment owned by another owner conflicts without writes", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) {
			f.targets = [][]driver.Value{transferTargetRow(1, 99, nil)}
		})
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, fixture.execs)
		assert.True(t, fixture.rolledBack)
	})

	t.Run("nil appointment owner conflicts", func(t *testing.T) {
		row := transferTargetRow(1, 10, nil)
		row[3] = nil
		fixture := newFixture(func(f *transferFixture) { f.targets = [][]driver.Value{row} })
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, fixture.execs)
	})

	t.Run("active medical record blocks the whole transfer", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) { f.medicalHit = true })
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, fixture.execs, "no appointment writes once a record exists")
		assert.True(t, fixture.rolledBack)
	})

	t.Run("overflow over max rows rejects before writes", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) {
			f.targets = make([][]driver.Value, petOwnerTransferMaxRows+1)
			for i := range f.targets {
				f.targets[i] = transferTargetRow(int64(i+1), 10, nil)
			}
		})
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INVALID_INPUT", appErr.Code)
		assert.Empty(t, fixture.execs)
		assert.True(t, fixture.rolledBack)
	})

	t.Run("cas zero rows conflicts", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) { f.rowsAffected = 0 })
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.True(t, fixture.rolledBack)
	})

	t.Run("nowait lock contention maps to conflict", func(t *testing.T) {
		fixture := newFixture(func(f *transferFixture) {
			f.failAt = `FROM "appointments"`
			f.failCode = "55P03"
		})
		gdb := transferFixtureGorm(t, fixture)
		_, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, fixture.execs)
	})

	t.Run("canceled context stops before the first query", func(t *testing.T) {
		fixture := newFixture(nil)
		gdb := transferFixtureGorm(t, fixture)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := runTransfer(ctx, gdb, NewPetOwnerTransfer(gdb))
		require.Error(t, err)
		assert.Empty(t, fixture.execs)
	})
}

func TestPetOwnerTransfer_LineCustomerRebind(t *testing.T) {
	lineUserID := "U-transfer-1"
	lineTarget := [][]driver.Value{transferTargetRow(1, 10, int64(55))}

	t.Run("confirmed unique mapping rebinds to the new customer", func(t *testing.T) {
		fixture := &transferFixture{
			petOwnerID:      30,
			rowsAffected:    1,
			ownerLineUserID: &lineUserID,
			ownerConfirmed:  true,
			lineCustomerIDs: []int64{88},
			targets:         lineTarget,
		}
		gdb := transferFixtureGorm(t, fixture)
		got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, got)
		require.Len(t, fixture.execs, 1)
		var sawCustomer, sawFields bool
		for _, arg := range fixture.eargs[0] {
			if v, ok := arg.Value.(int64); ok && v == 88 {
				sawCustomer = true
			}
			if v, ok := arg.Value.([]byte); ok && string(v) == "{}" {
				sawFields = true
			}
		}
		assert.True(t, sawCustomer, "UPDATE binds the uniquely matched line_customer id")
		assert.True(t, sawFields, "UPDATE clears customer_fields on rebind")
		lineQuery := fixture.queries[4]
		assert.Contains(t, lineQuery, `FROM "line_customers"`)
		assert.Contains(t, lineQuery, "FOR SHARE NOWAIT")
		lineArgs := fixture.qargs[4]
		require.Len(t, lineArgs, 4)
		assert.Equal(t, int64(1), lineArgs[0].Value)
		assert.Equal(t, int64(30), lineArgs[1].Value, "lookup is scoped to the new owner")
		assert.Equal(t, "U-transfer-1", lineArgs[2].Value)
		assert.Equal(t, int64(2), lineArgs[3].Value, "LIMIT 2 detects non-unique mappings")
	})

	for _, tc := range []struct {
		name            string
		confirmed       bool
		lineCustomerIDs []int64
	}{
		{name: "unconfirmed identity detaches", confirmed: false, lineCustomerIDs: []int64{88}},
		{name: "zero mappings detach", confirmed: true, lineCustomerIDs: nil},
		{name: "multiple mappings detach", confirmed: true, lineCustomerIDs: []int64{88, 89}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := &transferFixture{
				petOwnerID:      30,
				rowsAffected:    1,
				ownerLineUserID: &lineUserID,
				ownerConfirmed:  tc.confirmed,
				lineCustomerIDs: tc.lineCustomerIDs,
				targets:         lineTarget,
			}
			gdb := transferFixtureGorm(t, fixture)
			got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
			require.NoError(t, err)
			assert.Equal(t, []uint64{1}, got)
			require.Len(t, fixture.execs, 1)
			var sawNilCustomer, sawFields bool
			for _, arg := range fixture.eargs[0] {
				if arg.Value == nil {
					sawNilCustomer = true
				}
				if v, ok := arg.Value.([]byte); ok && string(v) == "{}" {
					sawFields = true
				}
			}
			assert.True(t, sawNilCustomer, "ambiguous or unconfirmed identity detaches line_customer_id")
			assert.True(t, sawFields, "detach clears customer_fields")
		})
	}
}

func TestPetOwnerTransfer_SkipsLineLookupWithoutLineTargets(t *testing.T) {
	lineUserID := "U-transfer-1"

	t.Run("manual-only targets skip the customer lookup entirely", func(t *testing.T) {
		fixture := &transferFixture{
			petOwnerID:      30,
			rowsAffected:    1,
			ownerLineUserID: &lineUserID,
			ownerConfirmed:  true,
			failAt:          `FROM "line_customers"`,
			targets:         [][]driver.Value{transferTargetRow(1, 10, nil)},
		}
		gdb := transferFixtureGorm(t, fixture)
		got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, got)
		for _, q := range fixture.queries {
			assert.NotContains(t, q, `FROM "line_customers"`)
		}
		require.Len(t, fixture.execs, 1)
		assert.NotContains(t, fixture.execs[0], "line_customer_id")
		assert.NotContains(t, fixture.execs[0], "customer_fields")
	})

	t.Run("zero targets do not touch line_customers or medical_records", func(t *testing.T) {
		fixture := &transferFixture{
			petOwnerID:      30,
			rowsAffected:    1,
			ownerLineUserID: &lineUserID,
			ownerConfirmed:  true,
			failAt:          `FROM "line_customers"`,
			targets:         [][]driver.Value{},
		}
		gdb := transferFixtureGorm(t, fixture)
		got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.NoError(t, err)
		assert.Empty(t, got)
		require.Len(t, fixture.queries, 3)
		assert.Empty(t, fixture.execs)
	})

	t.Run("whitespace line identity detaches without lookup", func(t *testing.T) {
		blank := "   "
		fixture := &transferFixture{
			petOwnerID:      30,
			rowsAffected:    1,
			ownerLineUserID: &blank,
			ownerConfirmed:  true,
			failAt:          `FROM "line_customers"`,
			targets:         [][]driver.Value{transferTargetRow(1, 10, int64(55))},
		}
		gdb := transferFixtureGorm(t, fixture)
		got, err := runTransfer(context.Background(), gdb, NewPetOwnerTransfer(gdb))
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, got)
		for _, q := range fixture.queries {
			assert.NotContains(t, q, `FROM "line_customers"`)
		}
		require.Len(t, fixture.execs, 1)
		var sawNilCustomer, sawFields bool
		for _, arg := range fixture.eargs[0] {
			if arg.Value == nil {
				sawNilCustomer = true
			}
			if v, ok := arg.Value.([]byte); ok && string(v) == "{}" {
				sawFields = true
			}
		}
		assert.True(t, sawNilCustomer)
		assert.True(t, sawFields)
	})
}
