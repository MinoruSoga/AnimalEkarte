package lstep

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// This connector tests actual repository SQL/Scan and connection/transaction
// routing. Its staged writes simulate atomicity; it does not prove PostgreSQL
// MVCC, advisory lock contention, sequence allocation, constraints, or RLS.
type retrySQLCall struct {
	query       string
	args        []driver.NamedValue
	connection  int
	transaction bool
	ambient     bool
	deadline    time.Time
}
type retrySQLFixture struct {
	calls        []retrySQLCall
	connections  int
	marker       uint64
	cacheDeleted bool
	noOwner      bool
	queryErr     error
	ackErr       error
	commitErr    error
	commits      int
	rollbacks    int
}
type retrySQLConnector struct{ f *retrySQLFixture }

func (c retrySQLConnector) Connect(context.Context) (driver.Conn, error) {
	c.f.connections++
	return &retrySQLConn{f: c.f, id: c.f.connections}, nil
}
func (c retrySQLConnector) Driver() driver.Driver { return retrySQLDriver{} }

type retrySQLDriver struct{}

func (retrySQLDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type retrySQLConn struct {
	f            *retrySQLFixture
	id           int
	active       bool
	ack          uint64
	cacheDeleted bool
}

func (c *retrySQLConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *retrySQLConn) Close() error              { return nil }
func (c *retrySQLConn) Begin() (driver.Tx, error) { return nil, errors.New("use BeginTx") }
func (c *retrySQLConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.active {
		return nil, errors.New("nested transaction")
	}
	c.active = true
	c.ack, c.cacheDeleted = 0, false
	return retrySQLTx{c: c}, nil
}
func (c *retrySQLConn) record(ctx context.Context, query string, args []driver.NamedValue) {
	deadline, _ := ctx.Deadline()
	c.f.calls = append(c.f.calls, retrySQLCall{
		query: strings.Join(strings.Fields(query), " "), args: append([]driver.NamedValue(nil), args...),
		connection: c.id, transaction: c.active, ambient: persistence.TxFromContext(ctx) != nil, deadline: deadline,
	})
}
func (c *retrySQLConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.record(ctx, query, args)
	if c.f.queryErr != nil {
		return nil, c.f.queryErr
	}
	if !strings.Contains(query, "INSERT INTO lstep_next_visit_sync_retries") || !strings.Contains(query, "RETURNING revision") {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	if c.active {
		return nil, errors.New("enqueue must autocommit independently")
	}
	rows := &retrySQLRows{}
	if !c.f.noOwner {
		c.f.marker++
		rows.revision = int64(c.f.marker)
	}
	return rows, nil
}
func (c *retrySQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.record(ctx, query, args)
	if !c.active {
		return nil, errors.New("write outside owned transaction")
	}
	switch {
	case strings.Contains(query, "pg_advisory_xact_lock"):
		return driver.RowsAffected(1), nil
	case strings.Contains(query, "DELETE FROM lstep_next_visit_sync_retries"):
		if c.f.ackErr != nil {
			return nil, c.f.ackErr
		}
		if len(args) != 3 {
			return nil, errors.New("ack requires scoped revision")
		}
		revision, ok := args[2].Value.(int64)
		if !ok {
			return nil, errors.New("revision is not an integer")
		}
		if uint64(revision) != c.f.marker {
			return driver.RowsAffected(0), nil
		}
		c.ack = uint64(revision)
		return driver.RowsAffected(1), nil
	case strings.Contains(query, `DELETE FROM "lstep_tag_cache"`):
		c.cacheDeleted = true
		return driver.RowsAffected(1), nil
	default:
		return nil, fmt.Errorf("unexpected exec: %s", query)
	}
}

type retrySQLTx struct{ c *retrySQLConn }

func (tx retrySQLTx) Commit() error {
	tx.c.active = false
	tx.c.f.commits++
	if tx.c.f.commitErr != nil {
		return tx.c.f.commitErr
	}
	if tx.c.ack != 0 && tx.c.ack == tx.c.f.marker {
		tx.c.f.marker = 0
	}
	if tx.c.cacheDeleted {
		tx.c.f.cacheDeleted = true
	}
	return nil
}
func (tx retrySQLTx) Rollback() error {
	tx.c.active = false
	tx.c.f.rollbacks++
	return nil
}

type retrySQLRows struct {
	revision int64
	read     bool
}

func (*retrySQLRows) Columns() []string { return []string{"revision"} }
func (*retrySQLRows) Close() error      { return nil }
func (r *retrySQLRows) Next(dest []driver.Value) error {
	if r.read || r.revision == 0 {
		return io.EOF
	}
	r.read = true
	dest[0] = r.revision
	return nil
}
func retrySQLRepo(t *testing.T, f *retrySQLFixture, maxConnections int) *lstepTagCacheRepository {
	t.Helper()
	db := sql.OpenDB(retrySQLConnector{f: f})
	db.SetMaxOpenConns(maxConnections)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return &lstepTagCacheRepository{db: gdb}
}
func retrySQLValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i := range args {
		values[i] = args[i].Value
	}
	return values
}

func TestNextVisitSyncRetryRepository_Enqueue(t *testing.T) {
	t.Run("returning revision preserves scope deadline and removes ambient transaction", func(t *testing.T) {
		f := &retrySQLFixture{marker: 40}
		r := retrySQLRepo(t, f, 2)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		deadline, _ := ctx.Deadline()
		abort := errors.New("business rollback")
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			revision, err := r.EnqueueNextVisitSync(persistence.WithTxValue(ctx, tx), 3, 9)
			require.NoError(t, err)
			require.Equal(t, uint64(41), revision)
			return abort
		})
		require.ErrorIs(t, err, abort)
		require.Equal(t, uint64(41), f.marker, "enqueue survives business rollback")
		require.Len(t, f.calls, 1)
		call := f.calls[0]
		require.False(t, call.transaction)
		require.False(t, call.ambient)
		require.Equal(t, deadline, call.deadline)
		require.Equal(t, []driver.Value{int64(3), int64(9)}, retrySQLValues(call.args))
		require.Contains(t, call.query, "SELECT clinic_id, id FROM owners WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL")
		require.Contains(t, call.query, "ON CONFLICT (clinic_id, owner_id) DO UPDATE SET revision = EXCLUDED.revision")
		require.Contains(t, call.query, "RETURNING revision")
		require.Equal(t, 2, f.connections, "enqueue uses independent autocommit connection")
	})
	for _, name := range []string{"owner missing", "query fails", "caller canceled"} {
		t.Run(name, func(t *testing.T) {
			f := &retrySQLFixture{noOwner: name == "owner missing"}
			r := retrySQLRepo(t, f, 1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sentinel := errors.New("insert failed")
			if name == "query fails" {
				f.queryErr = sentinel
			}
			if name == "caller canceled" {
				cancel()
			}
			revision, err := r.EnqueueNextVisitSync(ctx, 3, 9)
			require.Error(t, err)
			require.Zero(t, revision)
			if name == "owner missing" {
				require.True(t, apperrors.IsNotFound(err))
			}
			if name == "query fails" {
				require.ErrorIs(t, err, sentinel)
			}
			if name == "caller canceled" {
				require.ErrorIs(t, err, context.Canceled)
				require.Empty(t, f.calls)
			}
		})
	}
}

func TestNextVisitSyncRetryRepository_Acknowledge(t *testing.T) {
	t.Run("without ambient transaction no deletion occurs", func(t *testing.T) {
		f := &retrySQLFixture{marker: 41}
		r := retrySQLRepo(t, f, 1)
		require.Error(t, r.AcknowledgeNextVisitSync(context.Background(), 3, 9, 41))
		require.Empty(t, f.calls)
		require.Equal(t, uint64(41), f.marker)
	})
	for _, name := range []string{"commit", "newer revision", "callback rollback", "commit failure", "ack failure"} {
		t.Run(name, func(t *testing.T) {
			f := &retrySQLFixture{marker: 40}
			r := retrySQLRepo(t, f, 1)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			revision, err := r.EnqueueNextVisitSync(ctx, 3, 9)
			require.NoError(t, err)
			sentinel := errors.New(name)
			if name == "newer revision" {
				_, err = r.EnqueueNextVisitSync(ctx, 3, 9)
				require.NoError(t, err)
			}
			if name == "commit failure" {
				f.commitErr = sentinel
			}
			if name == "ack failure" {
				f.ackErr = sentinel
			}
			err = r.WithNextVisitOwnerLock(persistence.WithTxValue(ctx, r.db), 3, 9, func(txCtx context.Context) error {
				require.NotNil(t, persistence.TxFromContext(txCtx))
				if err := r.DeleteTag(txCtx, 3, 9, "fixture tag"); err != nil {
					return err
				}
				if err := r.AcknowledgeNextVisitSync(txCtx, 3, 9, revision); err != nil {
					return err
				}
				if name == "callback rollback" {
					return sentinel
				}
				return nil
			})
			failed := name == "callback rollback" || name == "commit failure" || name == "ack failure"
			if failed {
				require.ErrorIs(t, err, sentinel)
			} else {
				require.NoError(t, err)
			}
			wantMarker := uint64(0)
			if failed {
				wantMarker = revision
			}
			if name == "newer revision" {
				wantMarker = revision + 1
			}
			require.Equal(t, wantMarker, f.marker)
			require.Equal(t, !failed, f.cacheDeleted, "cache and acknowledgement share commit outcome")
			require.Equal(t, 1, f.connections, "enqueue before owner lock and callback ACK must not borrow a second connection")
			calls := f.calls[len(f.calls)-3:]
			require.Contains(t, calls[0].query, "pg_advisory_xact_lock")
			require.Equal(t, []driver.Value{"lstep-next-visit:3:9"}, retrySQLValues(calls[0].args))
			require.False(t, calls[0].ambient, "owner transaction detaches the caller transaction")
			require.Contains(t, calls[1].query, `DELETE FROM "lstep_tag_cache"`)
			require.Contains(t, calls[1].query, "owner_id = $1 AND tag_name = $2")
			require.Contains(t, calls[1].query, "clinic_id = $3")
			require.Equal(t, []driver.Value{int64(9), "fixture tag", int64(3)}, retrySQLValues(calls[1].args))
			require.Equal(t, "DELETE FROM lstep_next_visit_sync_retries WHERE clinic_id = $1 AND owner_id = $2 AND revision = $3", calls[2].query)
			require.Equal(t, []driver.Value{int64(3), int64(9), int64(revision)}, retrySQLValues(calls[2].args))
			deadline, _ := ctx.Deadline()
			for _, call := range calls {
				require.True(t, call.transaction)
				require.Equal(t, calls[0].connection, call.connection)
				require.Equal(t, deadline, call.deadline)
			}
			require.True(t, calls[1].ambient)
			require.True(t, calls[2].ambient)
		})
	}
}
