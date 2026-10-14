package support

// repository_plane_sync_test.go — SetPlaneSyncError の claim 競合ガード（PR650-02）。
//
// 起票失敗の記録は「まだ起票済みでない行」にだけ許可する。
// 自動起票・手動再送の失敗記録が、並行経路の成功 claim（plane_issue_id 非NULL）
// を遅れて上書きしないよう、UPDATE 条件に plane_issue_id IS NULL を原子適用する。
// 0 件時は tenant スコープの存在確認で「既に起票済み = 安全な no-op」と
// 「対象行なし/他医院 = NotFound」を区別する。
//
// 検証は database/sql driver のインメモリ double で行う（共有 DB 不使用）。
// UPDATE 文に IS NULL ガードが含まれることと、0 件時の分岐を検証する。
// 実 DB 上の原子的競合の証明は realdb_* テスト側の責務。

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type fakeDriverResult struct{ affected int64 }

func (r fakeDriverResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeDriverResult) RowsAffected() (int64, error) { return r.affected, nil }

type fakeDriverRows struct {
	cols []string
	vals [][]driver.Value
	pos  int
}

func (r *fakeDriverRows) Columns() []string { return r.cols }
func (r *fakeDriverRows) Close() error      { return nil }
func (r *fakeDriverRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.vals) {
		return io.EOF
	}
	copy(dest, r.vals[r.pos])
	r.pos++
	return nil
}

// fakeDriverConn は ExecContext/QueryContext を差し替えられる database/sql conn。
// GORM postgres dialector が発行する実 SQL と結果行をテストが制御する。
type fakeDriverConn struct {
	execFn  func(query string, args []driver.NamedValue) (driver.Result, error)
	queryFn func(query string, args []driver.NamedValue) (driver.Rows, error)
}

func (c *fakeDriverConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (c *fakeDriverConn) Close() error { return nil }
func (c *fakeDriverConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported")
}
func (c *fakeDriverConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.execFn == nil {
		return nil, fmt.Errorf("unexpected exec: %s", query)
	}
	return c.execFn(query, args)
}
func (c *fakeDriverConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.queryFn == nil {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	return c.queryFn(query, args)
}

type fakeDriverConnector struct{ conn *fakeDriverConn }

func (c *fakeDriverConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (c *fakeDriverConnector) Driver() driver.Driver                        { return fakeDriver{} }

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("fake driver has no DSN")
}

// newFakeDriverDB は実接続を持たない GORM postgres DB を組み立てる。
// SkipDefaultTransaction: fake conn はトランザクションを持たないため。
func newFakeDriverDB(t *testing.T, conn *fakeDriverConn) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sql.OpenDB(&fakeDriverConnector{conn: conn}),
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
	})
	require.NoError(t, err)
	return db
}

// 失敗記録の UPDATE は未 claim 行（plane_issue_id IS NULL）にだけ適用される。
func TestSetPlaneSyncError_WritesOnlyForUnclaimedRow(t *testing.T) {
	var statements []string
	conn := &fakeDriverConn{
		execFn: func(query string, _ []driver.NamedValue) (driver.Result, error) {
			statements = append(statements, query)
			return fakeDriverResult{affected: 1}, nil
		},
	}
	repo := NewRepository(newFakeDriverDB(t, conn))

	written, err := repo.SetPlaneSyncError(context.Background(), 7, 42, "plane api error (status 500)")

	require.NoError(t, err)
	assert.True(t, written)
	require.Len(t, statements, 1)
	// PR650-02: 並行起票の成功 claim を遅延失敗が上書きしないための原子ガード。
	assert.Contains(t, statements[0], "plane_issue_id IS NULL")
	assert.Contains(t, statements[0], "clinic_id")
}

// 並行経路が先に起票を claim 済み（UPDATE 0 件・行は存在する）場合、
// 遅延失敗記録は安全な no-op — NotFound にも成功状態の破壊にもならない。
func TestSetPlaneSyncError_AlreadyClaimedRowIsSafeNoOp(t *testing.T) {
	var statements []string
	conn := &fakeDriverConn{
		execFn: func(query string, _ []driver.NamedValue) (driver.Result, error) {
			statements = append(statements, query)
			return fakeDriverResult{affected: 0}, nil
		},
		queryFn: func(query string, _ []driver.NamedValue) (driver.Rows, error) {
			statements = append(statements, query)
			return &fakeDriverRows{cols: []string{"count"}, vals: [][]driver.Value{{int64(1)}}}, nil
		},
	}
	repo := NewRepository(newFakeDriverDB(t, conn))

	written, err := repo.SetPlaneSyncError(context.Background(), 7, 42, "late plane failure")

	// 既に起票済みの行は NotFound ではなく冪等 no-op として扱う。
	assert.NoError(t, err)
	assert.False(t, written)
	// UPDATE（IS NULL ガード）→ tenant スコープの存在確認、の順で発行される。
	require.Len(t, statements, 2)
	assert.Contains(t, statements[0], "plane_issue_id IS NULL")
	assert.Contains(t, statements[1], "clinic_id")
}

// 行が存在しない・他医院の行に対する記録は従来どおり NotFound。
func TestSetPlaneSyncError_MissingOrWrongClinicIsNotFound(t *testing.T) {
	conn := &fakeDriverConn{
		execFn: func(string, []driver.NamedValue) (driver.Result, error) {
			return fakeDriverResult{affected: 0}, nil
		},
		queryFn: func(string, []driver.NamedValue) (driver.Rows, error) {
			return &fakeDriverRows{cols: []string{"count"}, vals: [][]driver.Value{{int64(0)}}}, nil
		},
	}
	repo := NewRepository(newFakeDriverDB(t, conn))

	written, err := repo.SetPlaneSyncError(context.Background(), 999, 42, "late plane failure")

	require.Error(t, err)
	assert.False(t, written)
	assert.True(t, apperrors.IsNotFound(err))
}

// UPDATE 自体の DB 障害はラップしたエラーとして返す（no-op と混同しない）。
func TestSetPlaneSyncError_UpdateErrorIsWrapped(t *testing.T) {
	conn := &fakeDriverConn{
		execFn: func(string, []driver.NamedValue) (driver.Result, error) {
			return nil, errors.New("connection lost")
		},
	}
	repo := NewRepository(newFakeDriverDB(t, conn))

	written, err := repo.SetPlaneSyncError(context.Background(), 7, 42, "x")

	require.Error(t, err)
	assert.False(t, written)
	assert.False(t, apperrors.IsNotFound(err))
}

func TestSetPlaneSyncError_StateFetchUsesClaimedRow(t *testing.T) {
	var statement string
	conn := &fakeDriverConn{execFn: func(query string, _ []driver.NamedValue) (driver.Result, error) {
		statement = query
		return fakeDriverResult{affected: 1}, nil
	}}
	err := NewRepository(newFakeDriverDB(t, conn)).SetPlaneStateSyncError(context.Background(), 7, 42, "state fetch failed")
	require.NoError(t, err)
	assert.Contains(t, statement, "plane_issue_id IS NOT NULL")
	assert.Contains(t, statement, "clinic_id")
}

func TestSetPlaneSyncError_ExistenceQueryErrorIsWrapped(t *testing.T) {
	conn := &fakeDriverConn{
		execFn:  func(string, []driver.NamedValue) (driver.Result, error) { return fakeDriverResult{affected: 0}, nil },
		queryFn: func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("connection lost") },
	}
	written, err := NewRepository(newFakeDriverDB(t, conn)).SetPlaneSyncError(context.Background(), 7, 42, "late failure")
	require.Error(t, err)
	assert.False(t, written)
	assert.False(t, apperrors.IsNotFound(err))
}

func TestApplyPlaneStateSync_NonCompletedClearsErrorOnly(t *testing.T) {
	var statements []string
	var args []driver.NamedValue
	conn := &fakeDriverConn{execFn: func(query string, a []driver.NamedValue) (driver.Result, error) {
		statements = append(statements, query)
		args = append(args, a...)
		return fakeDriverResult{affected: 1}, nil
	}}
	repo := NewRepository(newFakeDriverDB(t, conn))

	err := repo.ApplyPlaneStateSync(context.Background(), 42, 7, false)

	require.NoError(t, err)
	require.Len(t, statements, 1, "one atomic UPDATE, no second write")
	assert.Contains(t, statements[0], "plane_sync_error")
	assert.Contains(t, statements[0], "plane_issue_id IS NOT NULL")
	assert.Contains(t, statements[0], "clinic_id")
	assert.Contains(t, statements[0], "deleted_at")
	assert.NotContains(t, statements[0], `"status"`, "non-completed must not touch status")
	require.Len(t, args, 4)
	assert.Nil(t, args[0].Value)
	assert.Equal(t, int64(7), args[2].Value)
	assert.Equal(t, int64(42), args[3].Value)
}

func TestApplyPlaneStateSync_CompletedClearsErrorAndResolves(t *testing.T) {
	var statements []string
	var args []driver.NamedValue
	conn := &fakeDriverConn{execFn: func(query string, a []driver.NamedValue) (driver.Result, error) {
		statements = append(statements, query)
		args = append(args, a...)
		return fakeDriverResult{affected: 1}, nil
	}}
	repo := NewRepository(newFakeDriverDB(t, conn))

	err := repo.ApplyPlaneStateSync(context.Background(), 42, 7, true)

	require.NoError(t, err)
	require.Len(t, statements, 1, "clear + resolve in the same UPDATE")
	assert.Contains(t, statements[0], "plane_sync_error")
	assert.Contains(t, statements[0], `"status"`)
	assert.Contains(t, statements[0], "plane_issue_id IS NOT NULL")
	assert.Contains(t, statements[0], "clinic_id")
	assert.Contains(t, statements[0], "deleted_at")
	require.Len(t, args, 5)
	assert.Nil(t, args[0].Value)
	assert.Equal(t, "resolved", args[1].Value)
	assert.Equal(t, int64(7), args[3].Value)
	assert.Equal(t, int64(42), args[4].Value)
}

func TestApplyPlaneStateSync_ZeroRowsIsNotFound(t *testing.T) {
	conn := &fakeDriverConn{
		execFn: func(string, []driver.NamedValue) (driver.Result, error) {
			return fakeDriverResult{affected: 0}, nil
		},
	}
	err := NewRepository(newFakeDriverDB(t, conn)).ApplyPlaneStateSync(context.Background(), 999, 42, false)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestApplyPlaneStateSync_UpdateErrorIsWrapped(t *testing.T) {
	conn := &fakeDriverConn{
		execFn: func(string, []driver.NamedValue) (driver.Result, error) {
			return nil, errors.New("connection lost")
		},
	}
	err := NewRepository(newFakeDriverDB(t, conn)).ApplyPlaneStateSync(context.Background(), 42, 7, false)
	require.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
}
