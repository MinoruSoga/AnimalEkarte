package inventory

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// EMR-449 A.5 fixture F「旧 writer/import 保持・過去 snapshot 不変」（review B-5）。
// 契約: docs/work/emr/emr-master-update-selected-contract-20261022.md :65, :75-77。
// DryRun のみ・実 DB/実行は NOT_RUN（Lead 指示）。入口別受領表は unit report 側に集約する。

// mbdwpWrite は gorm callback が組み立てた実発行予定 statement の記録。
type mbdwpWrite struct {
	op    string // "insert" | "update" | "delete"
	table string
	sql   string
	vars  []any
}

type mbdwpCapture struct {
	writes []mbdwpWrite
}

func mbdwpTableName(tx *gorm.DB) string {
	if tx.Statement.Table != "" {
		return tx.Statement.Table
	}
	if tx.Statement.Schema != nil {
		return tx.Statement.Schema.Table
	}
	return ""
}

// mbdwpEntrance は新列を所有しない1つの更新入口。table はその入口が発行してよい唯一の表。
type mbdwpEntrance struct {
	name  string
	table string
	run   func(ctx context.Context) error
}

// mbdwpNewDB は DryRun の gorm session と書込 capture を返す。
// RowsAffected=1 を立てて、repository の not-found / conflict 分岐や再読取を通す。
func mbdwpNewDB(t *testing.T) (*gorm.DB, *mbdwpCapture) {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	cap := &mbdwpCapture{}
	record := func(op string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			cap.writes = append(cap.writes, mbdwpWrite{op: op, table: mbdwpTableName(tx), sql: tx.Statement.SQL.String(), vars: append([]any(nil), tx.Statement.Vars...)})
			tx.RowsAffected = 1
		}
	}
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("mbdwp_capture_create", record("insert")))
	require.NoError(t, db.Callback().Update().After("gorm:update").Register("mbdwp_capture_update", record("update")))
	require.NoError(t, db.Callback().Delete().After("gorm:delete").Register("mbdwp_capture_delete", record("delete")))
	return db, cap
}

// mbdwpAssertColumnNotWritten は UPDATE/DELETE が is_bill_discount_excluded を SET しないことを確認する。
// SET 句に列が無い = 保存済みの NULL/false/true のいずれも上書きされない（A.5）。
// INSERT は列の所有写経路（旧 source fixture が bind 値を検査する）なので対象外。
func mbdwpAssertColumnNotWritten(t *testing.T, writes []mbdwpWrite) {
	t.Helper()
	for _, w := range writes {
		if w.op == "insert" {
			continue
		}
		assert.NotContains(t, w.sql, colIsBillDiscountExcluded, "%s on %s must not set %s: %s", w.op, w.table, colIsBillDiscountExcluded, w.sql)
	}
}

// mbdwpInsertColumnValue は INSERT の列リスト位置に対応する bind 値を取り出す。
func mbdwpInsertColumnValue(t *testing.T, w mbdwpWrite) any {
	t.Helper()
	open := strings.Index(w.sql, "(")
	closeIdx := strings.Index(w.sql, ") VALUES")
	require.GreaterOrEqual(t, open, 0, w.sql)
	require.Greater(t, closeIdx, open, w.sql)
	idx := -1
	for i, col := range strings.Split(w.sql[open+1:closeIdx], ",") {
		if strings.TrimSpace(col) == `"`+colIsBillDiscountExcluded+`"` {
			idx = i
			break
		}
	}
	require.GreaterOrEqual(t, idx, 0, "INSERT must bind %s: %s", colIsBillDiscountExcluded, w.sql)
	require.Less(t, idx, len(w.vars), w.sql)
	return w.vars[idx]
}

func mbdwpUnwrapVar(v any) any {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}

// mbdwpAssertInsertPreserves は旧 source の新 INSERT が model の値をそのまま bind することを
// 確認する（明示設定を missing へ落とさず、missing を自動 false 化しない: A.5）。
func mbdwpAssertInsertPreserves(t *testing.T, writes []mbdwpWrite, want any) {
	t.Helper()
	var inserts []mbdwpWrite
	for _, w := range writes {
		if w.op == "insert" {
			inserts = append(inserts, w)
		}
	}
	require.Len(t, inserts, 1, "expected exactly one INSERT: %+v", writes)
	got := mbdwpUnwrapVar(mbdwpInsertColumnValue(t, inserts[0]))
	assert.Equal(t, want, got, "the carried value must be bound as-is without coercion")
}

// mbdwpRunPreservationFixtures は1 master の A.5 保持 fixture を実行する。
// entrances: 新列を所有しない更新入口。allowedTables: 入口群が書いてよい表の集合。
// insert: 旧 source 相当の新 INSERT 入口。
func mbdwpRunPreservationFixtures(t *testing.T, allowedTables map[string]bool, entrances []mbdwpEntrance, insert func(ctx context.Context, value *bool) error) {
	t.Helper()

	t.Run("non-owning writers keep stored NULL/false/true", func(t *testing.T) {
		for _, e := range entrances {
			t.Run(e.name, func(t *testing.T) {
				db, cap := mbdwpNewDB(t)
				ctx := persistence.WithTxValue(t.Context(), db)
				require.NoError(t, e.run(ctx))
				require.NotEmpty(t, cap.writes, "entrance produced no write statements")
				for _, w := range cap.writes {
					assert.Equal(t, e.table, w.table, "write escaped the owning table: %s %s", w.op, w.sql)
				}
				mbdwpAssertColumnNotWritten(t, cap.writes)
			})
		}
	})

	t.Run("old-source INSERT preserves the carried value", func(t *testing.T) {
		for _, tt := range []struct {
			name  string
			value *bool
			want  any
		}{
			{name: "explicit true is stored, not dropped to missing", value: boolPtrForBillDiscountTest(true), want: true},
			{name: "explicit false is stored as false", value: boolPtrForBillDiscountTest(false), want: false},
			{name: "legacy NULL stays missing, not auto-false", value: nil, want: nil},
		} {
			t.Run(tt.name, func(t *testing.T) {
				db, cap := mbdwpNewDB(t)
				ctx := persistence.WithTxValue(t.Context(), db)
				require.NoError(t, insert(ctx, tt.value))
				mbdwpAssertInsertPreserves(t, cap.writes, tt.want)
			})
		}
	})

	t.Run("master writes never reach snapshot/history tables", func(t *testing.T) {
		for _, e := range entrances {
			t.Run(e.name, func(t *testing.T) {
				db, cap := mbdwpNewDB(t)
				ctx := persistence.WithTxValue(t.Context(), db)
				require.NoError(t, e.run(ctx))
				require.NotEmpty(t, cap.writes)
				for _, w := range cap.writes {
					assert.True(t, allowedTables[w.table],
						"master write must not touch snapshot/history table %q: %s %s", w.table, w.op, w.sql)
				}
			})
		}
	})
}

func TestBillDiscountWriterPreservation_MerchandiseItem(t *testing.T) {
	repo := &merchandiseItemRepository{}
	name := "renamed"
	entrances := []mbdwpEntrance{
		{name: "PATCH omitted (name only)", table: "merchandise_items", run: func(ctx context.Context) error {
			cmd := UpdateMerchandiseItemInput{Name: &name}
			return repo.update(ctx, 7, 5, buildMerchandiseItemUpdate(&cmd), cmd.Version)
		}},
		{name: "create is_active compensation UPDATE", table: "merchandise_items", run: func(ctx context.Context) error {
			return repo.Create(ctx, &model.MerchandiseItem{ClinicID: 7, Name: "x", IsActive: false, IsBillDiscountExcluded: boolPtrForBillDiscountTest(true)})
		}},
	}
	mbdwpRunPreservationFixtures(t, map[string]bool{"merchandise_items": true}, entrances,
		func(ctx context.Context, value *bool) error {
			return repo.Create(ctx, &model.MerchandiseItem{ClinicID: 7, Name: "x", IsActive: true, IsBillDiscountExcluded: value})
		})
}
