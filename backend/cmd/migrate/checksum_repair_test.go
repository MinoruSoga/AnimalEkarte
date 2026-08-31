package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
)

type repairDriver struct{ state *repairDriverState }
type repairDriverState struct {
	began, rolledBack, committed bool
	statements                   []string
}
type repairConn struct {
	state *repairDriverState
	inTx  bool
}
type repairTx struct{ state *repairDriverState }

func (d *repairDriver) Open(string) (driver.Conn, error) { return &repairConn{state: d.state}, nil }
func (c *repairConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*repairConn) Close() error { return nil }
func (c *repairConn) Begin() (driver.Tx, error) {
	c.state.began = true
	c.inTx = true
	return &repairTx{state: c.state}, nil
}
func (c *repairConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.statements = append(c.state.statements, query)
	if c.inTx && strings.Contains(query, "ADD CONSTRAINT chk_lab_device_item_masters_source_type") {
		return nil, errors.New("injected DDL failure")
	}
	return driver.RowsAffected(1), nil
}
func (t *repairTx) Commit() error   { t.state.committed = true; return nil }
func (t *repairTx) Rollback() error { t.state.rolledBack = true; return nil }

func TestChecksumRepairRollsBackCompanionDDLAndChecksumTogether(t *testing.T) {
	state := &repairDriverState{}
	name := "checksum-repair-transaction-test"
	sql.Register(name, &repairDriver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repaired, err := tryRepairKnownChecksumDrift(db, "001_init.sql", "28e954b32fd606a122e0cb29815ea277f8a96cb0966208f39e6fe69dd8cb9c4e", "287bfce66c810503c43c8a5c1d4cf414f561af2555314eb4119be74253ce77ce")
	if err == nil || repaired {
		t.Fatalf("result repaired=%v err=%v, want rollback failure", repaired, err)
	}
	if !state.began || !state.rolledBack || state.committed {
		t.Fatalf("transaction state = %+v", state)
	}
	for _, stmt := range state.statements {
		if strings.Contains(stmt, "UPDATE schema_migrations") {
			t.Fatal("checksum updated after companion DDL failure")
		}
	}
}
