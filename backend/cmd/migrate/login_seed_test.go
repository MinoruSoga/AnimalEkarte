package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

func TestExpectedSeedBundleDirs_IncludesLoginForLocalAndStaging(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	got := expectedSeedBundleDirs()
	want := []string{"002_master", seedlogin.BundleDir}
	if !slices.Equal(got, want) {
		t.Fatalf("development plan = %v, want %v", got, want)
	}

	t.Setenv("APP_ENV", "staging")
	got = expectedSeedBundleDirs()
	if !slices.Equal(got, want) {
		t.Fatalf("staging plan = %v, want %v", got, want)
	}
}

func TestExpectedSeedBundleDirs_ProductionOmitsLogin(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	got := expectedSeedBundleDirs()
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("production plan = %v, want master only", got)
	}
}

func TestSeedBundlesForEnv_StillExcludesCSVAccounts(t *testing.T) {
	t.Parallel()
	got := seedBundlesForEnv("staging")
	if slices.Contains(got, seedlogin.BundleDir) {
		t.Fatalf("CSV plan must not include %s: %v", seedlogin.BundleDir, got)
	}
}

func TestLoginSeedNeedsApply_NoRowNeedsApply(t *testing.T) {
	db := openLoginSeedChecksumDB(t, "", false)
	needs, err := loginSeedNeedsApply(db, seedlogin.MigrationKey(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if !needs {
		t.Fatal("expected needsApply=true when no schema_migrations row")
	}
}

func TestLoginSeedNeedsApply_SameChecksumSkips(t *testing.T) {
	db := openLoginSeedChecksumDB(t, "same-checksum", true)
	needs, err := loginSeedNeedsApply(db, seedlogin.MigrationKey(), "same-checksum")
	if err != nil {
		t.Fatal(err)
	}
	if needs {
		t.Fatal("expected needsApply=false when checksum matches")
	}
}

func TestLoginSeedNeedsApply_ChangedChecksumReapplies(t *testing.T) {
	db := openLoginSeedChecksumDB(t, "old-checksum", true)
	needs, err := loginSeedNeedsApply(db, seedlogin.MigrationKey(), "new-checksum")
	if err != nil {
		t.Fatal(err)
	}
	if !needs {
		t.Fatal("expected needsApply=true when catalog checksum changed")
	}
}

type loginSeedChecksumDriver struct {
	checksum string
	found    bool
}

type loginSeedChecksumConn struct {
	checksum string
	found    bool
}

type loginSeedChecksumRows struct {
	checksum string
	found    bool
	done     bool
}

func openLoginSeedChecksumDB(t *testing.T, checksum string, found bool) *sql.DB {
	t.Helper()
	name := "login_seed_checksum_" + t.Name()
	sql.Register(name, &loginSeedChecksumDriver{checksum: checksum, found: found})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func (d *loginSeedChecksumDriver) Open(string) (driver.Conn, error) {
	return &loginSeedChecksumConn{checksum: d.checksum, found: d.found}, nil
}
func (c *loginSeedChecksumConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*loginSeedChecksumConn) Close() error { return nil }
func (*loginSeedChecksumConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected begin")
}
func (c *loginSeedChecksumConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return &loginSeedChecksumRows{checksum: c.checksum, found: c.found}, nil
}
func (*loginSeedChecksumConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("unexpected exec")
}
func (r *loginSeedChecksumRows) Columns() []string { return []string{"checksum"} }
func (r *loginSeedChecksumRows) Close() error      { return nil }
func (r *loginSeedChecksumRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	if !r.found {
		return io.EOF
	}
	dest[0] = r.checksum
	return nil
}
