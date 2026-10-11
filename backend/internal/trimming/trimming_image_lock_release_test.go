package trimming

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLockDriverState struct {
	connects         int
	closed           int
	unlockVal        bool
	unlockErr        error
	execErr          error
	queryCtxErr      error
	queryDeadline    time.Time
	queryHasDeadline bool
}

type fakeLockConnector struct {
	state *fakeLockDriverState
}

func (c *fakeLockConnector) Connect(context.Context) (driver.Conn, error) {
	c.state.connects++
	return &fakeLockConn{state: c.state}, nil
}

func (c *fakeLockConnector) Driver() driver.Driver { return fakeLockDriver{} }

type fakeLockDriver struct{}

func (fakeLockDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("fakeLockDriver.Open must not be called; use Connector")
}

type fakeLockConn struct {
	state *fakeLockDriverState
}

func (c *fakeLockConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *fakeLockConn) Close() error {
	c.state.closed++
	return nil
}

func (c *fakeLockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *fakeLockConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	if c.state.execErr != nil {
		return nil, c.state.execErr
	}
	return driver.RowsAffected(0), nil
}

func (c *fakeLockConn) QueryContext(ctx context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.queryCtxErr = ctx.Err()
	c.state.queryDeadline, c.state.queryHasDeadline = ctx.Deadline()
	if c.state.unlockErr != nil {
		return nil, c.state.unlockErr
	}
	return &fakeLockRows{values: []driver.Value{c.state.unlockVal}}, nil
}

type fakeLockRows struct {
	values []driver.Value
	pos    int
}

func (r *fakeLockRows) Columns() []string { return []string{"pg_advisory_unlock"} }
func (r *fakeLockRows) Close() error      { return nil }

func (r *fakeLockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.values) {
		return io.EOF
	}
	dest[0] = r.values[r.pos]
	r.pos++
	return nil
}

func TestReleaseTrimmingImageKeySessionLock_SuccessReturnsConnToPool(t *testing.T) {
	ctx := context.Background()
	state := &fakeLockDriverState{unlockVal: true}
	db := sql.OpenDB(&fakeLockConnector{state: state})
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, state.connects)

	releaseTrimmingImageKeySessionLock(ctx, conn, "trimming-images/1/ok.jpg")
	require.NoError(t, conn.Close())

	conn2, err := db.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, conn2.Close())

	assert.Equal(t, 1, state.connects,
		"unlock 成功時は同じ物理接続が pool から再利用されなければならない")
	assert.Equal(t, 0, state.closed,
		"unlock 成功時に物理接続を close してはいけない")
}

func TestReleaseTrimmingImageKeySessionLock_CanceledParentUsesBoundedDetachedCtx(t *testing.T) {
	ctx := context.Background()
	state := &fakeLockDriverState{unlockVal: true}
	db := sql.OpenDB(&fakeLockConnector{state: state})
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	parent, cancel := context.WithCancel(ctx)
	cancel()
	releaseTrimmingImageKeySessionLock(parent, conn, "trimming-images/1/canceled.jpg")
	require.NoError(t, conn.Close())

	assert.NoError(t, state.queryCtxErr,
		"detached release ctx は親のキャンセルを引き継がない")
	require.True(t, state.queryHasDeadline, "release ctx は deadline を持つ")
	remaining := time.Until(state.queryDeadline)
	assert.Greater(t, remaining, time.Duration(0))
	assert.LessOrEqual(t, remaining, supersededImageCleanupTimeout)
}

func TestReleaseTrimmingImageKeySessionLock_UnlockFalseDiscardsConn(t *testing.T) {
	ctx := context.Background()
	state := &fakeLockDriverState{unlockVal: false}
	db := sql.OpenDB(&fakeLockConnector{state: state})
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	releaseTrimmingImageKeySessionLock(ctx, conn, "trimming-images/1/false.jpg")
	assert.ErrorIs(t, conn.Close(), sql.ErrConnDone)

	conn2, err := db.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, conn2.Close())

	assert.Equal(t, 2, state.connects,
		"unlock=false は不確実なセッションなので新しい物理接続が必要")
	assert.Equal(t, 1, state.closed,
		"不確実なセッションの物理接続は pool へ戻さず close しなければならない")
}

func TestReleaseTrimmingImageKeySessionLock_UnlockErrorDiscardsConn(t *testing.T) {
	ctx := context.Background()
	state := &fakeLockDriverState{unlockErr: errors.New("unlock query failed")}
	db := sql.OpenDB(&fakeLockConnector{state: state})
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	releaseTrimmingImageKeySessionLock(ctx, conn, "trimming-images/1/err.jpg")
	assert.ErrorIs(t, conn.Close(), sql.ErrConnDone)

	conn2, err := db.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, conn2.Close())

	assert.Equal(t, 2, state.connects,
		"unlock エラーは不確実なセッションなので新しい物理接続が必要")
	assert.Equal(t, 1, state.closed,
		"不確実なセッションの物理接続は pool へ戻さず close しなければならない")
}

func TestAcquireTrimmingImageKeySessionLock_FailureDiscardsConn(t *testing.T) {
	ctx := context.Background()
	state := &fakeLockDriverState{execErr: errors.New("acquire exec failed")}
	db := sql.OpenDB(&fakeLockConnector{state: state})
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	err = acquireTrimmingImageKeySessionLock(ctx, conn, "trimming-images/1/acq.jpg")
	require.Error(t, err, "前提: acquire は失敗する")
	assert.ErrorIs(t, conn.Close(), sql.ErrConnDone)

	conn2, err := db.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, conn2.Close())

	assert.Equal(t, 2, state.connects,
		"acquire 失敗の不確実なセッションは破棄され新しい物理接続が必要")
	assert.Equal(t, 1, state.closed,
		"acquire 失敗後の物理接続は pool へ戻さず close しなければならない")
}
