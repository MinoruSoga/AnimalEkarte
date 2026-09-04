package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const checkedInInitialMigrationSHA256 = "d792bca4f4d985f1b2acdc7cd9ac0c9af80fe28aa3135712ec70896086c8ba2d"

type failClosedDriver struct{ state *failClosedDriverState }
type failClosedDriverState struct {
	storedChecksum               string
	began, rolledBack, committed bool
	statements                   []string
}
type failClosedConn struct{ state *failClosedDriverState }
type failClosedTx struct{ state *failClosedDriverState }
type checksumRows struct {
	checksum string
	done     bool
}

func (d *failClosedDriver) Open(string) (driver.Conn, error) {
	return &failClosedConn{state: d.state}, nil
}
func (c *failClosedConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*failClosedConn) Close() error { return nil }
func (c *failClosedConn) Begin() (driver.Tx, error) {
	c.state.began = true
	return &failClosedTx{state: c.state}, nil
}
func (c *failClosedConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.statements = append(c.state.statements, query)
	return &checksumRows{checksum: c.state.storedChecksum}, nil
}
func (c *failClosedConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.statements = append(c.state.statements, query)
	return nil, errors.New("unexpected mutation")
}
func (t *failClosedTx) Commit() error   { t.state.committed = true; return nil }
func (t *failClosedTx) Rollback() error { t.state.rolledBack = true; return nil }
func (*checksumRows) Columns() []string { return []string{"checksum"} }
func (*checksumRows) Close() error      { return nil }
func (r *checksumRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	dest[0] = r.checksum
	r.done = true
	return nil
}

func checkedInInitialMigrationChecksum(t *testing.T) string {
	t.Helper()
	migrationPath := filepath.Join("..", "..", "migrations", "001_init.sql")
	contents, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read %s: %v", migrationPath, err)
	}
	actual := sha256.Sum256(contents)
	return hex.EncodeToString(actual[:])
}

func TestCheckedInInitialMigrationChecksum(t *testing.T) {
	if got := checkedInInitialMigrationChecksum(t); got != checkedInInitialMigrationSHA256 {
		t.Fatalf("001_init.sql SHA-256 = %s, want %s; review schema drift and the reset/rebuild contract", got, checkedInInitialMigrationSHA256)
	}
}

func TestKnownOldChecksumsToCheckedInMigrationFailClosedWithoutCompanionDDLOrCAS(t *testing.T) {
	oldChecksums := []string{
		"921c88ba86159a501f86abcddd8639fd65dee37f7d9d3f1c811844f7294c6cc5",
		"6b618736ca0dba65408b7bf2a6d5d8c607f12c3adf1c544321d17f3004b62fcb",
		"aa30c742a20d4a9fa7a9d323cfa0314d4c83bec57712f7984228c4451c089691",
		"28e954b32fd606a122e0cb29815ea277f8a96cb0966208f39e6fe69dd8cb9c4e",
		"287bfce66c810503c43c8a5c1d4cf414f561af2555314eb4119be74253ce77ce",
		"d92b3c7af70c00ac305ba33d20e1aa3b2de9de55a97919cc98021f2e88926e1b",
		"60477e0ba76116a38ce2ac0f9563e9ba39aa88388b6dddca029f4899f6808ea4",
	}
	current := checkedInInitialMigrationChecksum(t)
	for i, oldChecksum := range oldChecksums {
		t.Run(oldChecksum[:12], func(t *testing.T) {
			state := &failClosedDriverState{storedChecksum: oldChecksum}
			driverName := fmt.Sprintf("checksum-fail-closed-test-%d", i)
			sql.Register(driverName, &failClosedDriver{state: state})
			db, err := sql.Open(driverName, "")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			applied, err := isAlreadyApplied(db, "001_init.sql", current)
			if !applied || err == nil {
				t.Fatalf("result applied=%v err=%v, want applied migration and fail-closed error", applied, err)
			}
			if !strings.Contains(err.Error(), "docs/ops/deploy/LOCAL_DB_RESET.md") || !strings.Contains(err.Error(), "DB_RESET=true") {
				t.Fatalf("error %q does not provide the explicit reset/rebuild path", err)
			}
			if state.began || state.rolledBack || state.committed {
				t.Fatalf("unexpected repair transaction state = %+v", state)
			}
			if len(state.statements) != 1 || !strings.Contains(state.statements[0], "SELECT checksum FROM schema_migrations") {
				t.Fatalf("statements = %q, want checksum SELECT only (no companion DDL or checksum CAS)", state.statements)
			}
		})
	}
}
