// model_ddl_column_drift_test.go — BUG-2026-07-27-01 の再発防止ゲート。
//
// 検出する障害クラス:
//
//	GORM model に列が増えたのに migrations/*.sql に対応する DDL が無い（= オフラインで
//	検出できる model→DDL ドリフト）。この状態は型検査でも lint でも捕まらず、
//	稼働時に初めて PostgreSQL 42703 (undefined_column) として現れる。
//
// なぜ「明示列挙」が効くのか:
//
//	GORM v2 は生 Joins() を含むクエリで SELECT * ではなく model の DBNames を
//	`<table>.<column>` として明示列挙する。したがって Joins を持つ一覧クエリ
//	（internal/pet/repository.go の FindAll 等）だけが 42703 で落ち、
//	Preload や単体取得（SELECT *）は成功する。症状が部分的に見えるため誤診しやすい。
//
// このゲートの限界（重要・過大評価しないこと）:
//
//	BUG-2026-07-27-01 の実際の原因は repo 内の model↔DDL 不整合ではなく、
//	「稼働中 DB のスキーマが repo の DDL より古い」ことだった
//	（002_add_pets_version.sql は commit 済みだが、稼働中コンテナは
//	その commit 以前に migration を適用していた）。本ゲートは repo 内の静的整合性
//	だけを見るため、その事象は検出できない。稼働 DB の遅れに対する実効的な防御は
//	httpapi 側の「未知 pg コード → 500 + SQLSTATE をサーバログへ」であり、
//	本ゲートは「migration を書き忘れた」別クラスを塞ぐ。
//
// 方向は model→DDL の一方向のみ:
//
//	逆方向（DDL にあるが model に無い）は正常な状態が多数ある
//	（model が一部列しか読まない、別 model が同一テーブルを扱う等）ため、
//	強制すると偽陽性が大量に出る。ここでは「model が読もうとする列が DDL に無い」
//	＝確実にランタイム 42703 になる方向だけを違反とする。
//
// 列名の導出は GORM 自身の schema.NamingStrategy を使う。独自 snake_case 実装は
// initialism（ID/LINE/URL 等）で GORM と乖離し偽陽性を生むため使わない。
package lintscan

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gorm.io/gorm/schema"
)

// modelDir は internal/lintscan から見た internal/model の相対パス。
// migrationsDir (= "../../migrations") は migration_cascade_lint_test.go が定義済み。
const modelDir = "../model"

// minModelStructsScanned は「AST 走査が壊れて 0 件になったのに PASS」という
// vacuous pass を防ぐ下限。2026-07-27 時点で TableName() を持つ model は 82 個。
const minModelStructsScanned = 60

// minDDLTablesParsed は同様に DDL パーサの vacuous pass を防ぐ下限。
const minDDLTablesParsed = 60

type modelTable struct {
	StructName string
	Table      string
	Columns    []string
}

type modelDDLDriftViolation struct {
	StructName string
	Table      string
	Column     string
	Reason     string
}

func (v modelDDLDriftViolation) String() string {
	return fmt.Sprintf("model.%s -> %s.%s: %s", v.StructName, v.Table, v.Column, v.Reason)
}

// TestModelDDLColumnDrift_EveryModelColumnExistsInMigrations は実ツリーに対する本番ゲート。
func TestModelDDLColumnDrift_EveryModelColumnExistsInMigrations(t *testing.T) {
	modelSources := mustReadGoSources(t, modelDir)
	models, err := extractModelTables(modelSources)
	if err != nil {
		t.Fatalf("extract model tables: %v", err)
	}
	if len(models) < minModelStructsScanned {
		t.Fatalf("only %d model structs with TableName() found in %s (expected >= %d). "+
			"AST scan is likely broken — refusing to pass vacuously.",
			len(models), modelDir, minModelStructsScanned)
	}

	ddl, err := parseDDLTableColumns(mustReadMigrationSQL(t))
	if err != nil {
		t.Fatalf("parse migration DDL: %v", err)
	}
	if len(ddl) < minDDLTablesParsed {
		t.Fatalf("only %d tables parsed from %s (expected >= %d). "+
			"DDL parser is likely broken — refusing to pass vacuously.",
			len(ddl), migrationsDir, minDDLTablesParsed)
	}

	violations := reconcileModelDDLColumns(models, ddl)
	if len(violations) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d model column(s) have no corresponding migration DDL.\n", len(violations))
		b.WriteString("These columns will be emitted by GORM as explicit `table.column` in any query\n")
		b.WriteString("carrying a raw Joins(), producing PostgreSQL 42703 at runtime (BUG-2026-07-27-01).\n")
		b.WriteString("Fix: add the column in a new backend/migrations/*.sql, or tag the field gorm:\"-\".\n")
		for _, v := range violations {
			b.WriteString("  - " + v.String() + "\n")
		}
		t.Fatal(b.String())
	}
}

// TestModelDDLColumnDrift_EveryGormStructIsCovered は、上のゲートの適用対象から
// 静かに漏れる model が無いことを保証する。
//
// 上のゲートは TableName() を宣言する構造体だけを検査する。GORM は TableName() が
// 無い場合、構造体名の複数形 snake_case を既定のテーブル名にするため、TableName() を
// 書かない model も動作してしまう。その model は上のゲートに現れず、
// 「ドリフトがあるのにゲートが緑」という偽陰性になる。
//
// そこで「gorm タグを1つでも持つ = GORM 永続化を意図した構造体」でありながら
// TableName() を持たないものを違反とする。2026-07-27 時点で内部 DTO
// （*Response / *Thresholds / *Breakdown / Lab* 等）は gorm タグを持たないため緑。
func TestModelDDLColumnDrift_EveryGormStructIsCovered(t *testing.T) {
	sources := mustReadGoSources(t, modelDir)
	uncovered, err := findGormStructsWithoutTableName(sources)
	if err != nil {
		t.Fatalf("scan model sources: %v", err)
	}
	if len(uncovered) > 0 {
		t.Fatalf("%d struct(s) in internal/model carry gorm tags but declare no TableName(): %v\n"+
			"TestModelDDLColumnDrift_EveryModelColumnExistsInMigrations silently skips these, "+
			"so column drift in them would go undetected. "+
			"Add a TableName() method, or remove the gorm tags if the struct is not persisted.",
			len(uncovered), uncovered)
	}
}

func TestFindGormStructsWithoutTableName_DetectsUncovered(t *testing.T) {
	src := []byte(`package model

type Covered struct {
	ID uint64 ` + "`gorm:\"primaryKey\"`" + `
}

func (Covered) TableName() string { return "covered" }

// gorm タグを持つのに TableName() が無い = 上位ゲートから漏れる
type Uncovered struct {
	ID uint64 ` + "`gorm:\"primaryKey\"`" + `
}

// タグは一切無いが ID を持つ。GORM は無タグでも ID を主キーとみなし
// テーブル名も自動導出するため、これも永続化モデルであり得る。
type UntaggedWithID struct {
	ID   uint64
	Name string
}

// gorm.Model の匿名 embed。embed 行自体に gorm タグは付かない。
type EmbedsGormModel struct {
	gorm.Model
	Name string
}

// gorm タグも ID も embed も無い純粋な DTO は対象外
type PlainDTO struct {
	Total int ` + "`json:\"total\"`" + `
}
`)

	got, err := findGormStructsWithoutTableName(map[string][]byte{"m.go": src})
	if err != nil {
		t.Fatalf("findGormStructsWithoutTableName: %v", err)
	}
	want := []string{"EmbedsGormModel", "Uncovered", "UntaggedWithID"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// isGormModelSignal は「この field の存在をもって、その構造体が GORM 永続化モデルを
// 意図していると判断してよいか」を返す。
//
// gorm タグの有無だけでは不十分である。GORM は無タグでも `ID` field を主キーとみなし、
// テーブル名も構造体名の複数形 snake_case へ自動フォールバックするため、
//
//	type Foo struct { ID uint64; Name string }
//
// はタグゼロでも有効な永続化モデルになる。gorm.Model の匿名 embed も同様に無タグ。
// これらを候補から外すと、TableName() を書かない新規 model が
// 上位ゲートにも本カバレッジゲートにも現れない二重の偽陰性になる。
func isGormModelSignal(field *ast.Field) bool {
	if gormTag(field) != "" {
		return true
	}
	// 匿名 embed（gorm.Model 等）
	if len(field.Names) == 0 {
		return true
	}
	for _, name := range field.Names {
		if name.Name == "ID" {
			return true
		}
	}
	return false
}

// findGormStructsWithoutTableName は GORM 永続化モデルと判断できるが
// TableName() を宣言しない構造体名を返す。
func findGormStructsWithoutTableName(sources map[string][]byte) ([]string, error) {
	fset := token.NewFileSet()
	gormTagged := map[string]struct{}{}
	hasTableName := map[string]struct{}{}

	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		file, err := parser.ParseFile(fset, name, sources[name], 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					for _, field := range st.Fields.List {
						if isGormModelSignal(field) {
							gormTagged[ts.Name.Name] = struct{}{}
							break
						}
					}
				}
			case *ast.FuncDecl:
				if recv, _, ok := parseTableNameMethod(d); ok {
					hasTableName[recv] = struct{}{}
				}
			}
		}
	}

	var uncovered []string
	for name := range gormTagged {
		if _, ok := hasTableName[name]; !ok {
			uncovered = append(uncovered, name)
		}
	}
	sort.Strings(uncovered)
	return uncovered, nil
}

// TestReconcileModelDDLColumns_GateDetectsDrift は「意図的にドリフトさせたら FAIL する」ことを
// 合成入力で恒久的に固定する。実 model ファイルを一時編集して確認する運用は、
// 並行セッションで作業ツリーが汚れている状況では revert 漏れリスクがあるため採らない
// （既存 TestSeedCSVSchemaDrift_GateDetectsViolations と同じ手法）。
func TestReconcileModelDDLColumns_GateDetectsDrift(t *testing.T) {
	ddl := map[string][]string{
		"pets": {"id", "clinic_id", "name"},
	}

	t.Run("model column missing from DDL is a violation", func(t *testing.T) {
		models := []modelTable{{
			StructName: "Pet",
			Table:      "pets",
			// version は DDL に無い = BUG-2026-07-27-01 と同じ形状
			Columns: []string{"id", "clinic_id", "name", "version"},
		}}
		got := reconcileModelDDLColumns(models, ddl)
		if len(got) != 1 {
			t.Fatalf("expected exactly 1 violation, got %d (%v)", len(got), got)
		}
		if got[0].Column != "version" || got[0].Table != "pets" {
			t.Errorf("unexpected violation: %v", got[0])
		}
	})

	t.Run("fully aligned model produces no violation", func(t *testing.T) {
		models := []modelTable{{
			StructName: "Pet",
			Table:      "pets",
			Columns:    []string{"id", "clinic_id", "name"},
		}}
		if got := reconcileModelDDLColumns(models, ddl); len(got) != 0 {
			t.Errorf("expected no violations, got %v", got)
		}
	})

	t.Run("DDL column absent from model is NOT a violation (one-way gate)", func(t *testing.T) {
		models := []modelTable{{
			StructName: "Pet",
			Table:      "pets",
			Columns:    []string{"id"},
		}}
		if got := reconcileModelDDLColumns(models, ddl); len(got) != 0 {
			t.Errorf("reverse direction must not be enforced, got %v", got)
		}
	})

	t.Run("model whose table has no DDL at all is a violation", func(t *testing.T) {
		models := []modelTable{{
			StructName: "Ghost",
			Table:      "ghosts",
			Columns:    []string{"id"},
		}}
		got := reconcileModelDDLColumns(models, ddl)
		if len(got) != 1 || got[0].Table != "ghosts" {
			t.Fatalf("expected a missing-table violation, got %v", got)
		}
	})
}

func TestExtractModelTables_DerivesColumnsLikeGORM(t *testing.T) {
	src := []byte(`package model

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
	ID              uint64         ` + "`gorm:\"primaryKey\"`" + `
	AnimalSpeciesID uint64         ` + "`gorm:\"not null\"`" + `
	NameKana        string         ` + "`gorm:\"column:name_kana;default:''\"`" + `
	LINEUserID      string
	Ignored         string         ` + "`gorm:\"-\"`" + `
	unexported      string
	DeletedAt       gorm.DeletedAt
	CreatedAt       time.Time
	Owner           *Other         ` + "`gorm:\"foreignKey:OwnerID\"`" + `
	Children        []Other        ` + "`gorm:\"foreignKey:SampleID\"`" + `
}

func (Sample) TableName() string { return "samples" }

type Other struct {
	ID uint64
}

func (Other) TableName() string { return "others" }
`)

	got, err := extractModelTables(map[string][]byte{"sample.go": src})
	if err != nil {
		t.Fatalf("extractModelTables: %v", err)
	}

	byName := map[string]modelTable{}
	for _, m := range got {
		byName[m.StructName] = m
	}

	sample, ok := byName["Sample"]
	if !ok {
		t.Fatalf("Sample not extracted, got %v", got)
	}
	if sample.Table != "samples" {
		t.Errorf("table = %q, want %q", sample.Table, "samples")
	}

	want := []string{
		"animal_species_id",
		"created_at",
		"deleted_at",
		"id",
		"line_user_id", // initialism handling must match GORM exactly
		"name_kana",    // explicit column tag wins
	}
	if !reflect.DeepEqual(sample.Columns, want) {
		t.Errorf("columns = %v, want %v", sample.Columns, want)
	}
}

func TestParseDDLTableColumns_HandlesCreateAndAlter(t *testing.T) {
	sql := `
-- a comment mentioning CREATE TABLE decoys (x INT);
CREATE TABLE pets (
    id BIGSERIAL PRIMARY KEY,
    clinic_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL DEFAULT '',
    weight NUMERIC(6,2),
    CONSTRAINT chk_pets_name CHECK (name <> ''),
    UNIQUE (clinic_id, name)
);

/* block comment
CREATE TABLE more_decoys (y INT);
*/

CREATE TABLE IF NOT EXISTS "owners" (
    id BIGSERIAL PRIMARY KEY
);

ALTER TABLE pets ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE pets ADD COLUMN IF NOT EXISTS danger_reason TEXT;
ALTER TABLE pets ADD CONSTRAINT fk_pets_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id);

-- 実データ由来の回帰ケース1: 複数アクション ALTER。先頭が ALTER COLUMN で
-- 後続の ADD COLUMN が続く（001_init.sql の treatments dose_* と同形）。
ALTER TABLE pets
  ALTER COLUMN weight TYPE numeric(10,2),
  ADD COLUMN dose_weight_kg numeric(6,2),
  ADD COLUMN dose_amount_mg numeric(12,6);
`

	got, err := parseDDLTableColumns(sql)
	if err != nil {
		t.Fatalf("parseDDLTableColumns: %v", err)
	}

	wantPets := []string{
		"clinic_id", "danger_reason", "dose_amount_mg", "dose_weight_kg",
		"id", "name", "version", "weight",
	}
	gotPets := append([]string(nil), got["pets"]...)
	sort.Strings(gotPets)
	if !reflect.DeepEqual(gotPets, wantPets) {
		t.Errorf("pets columns = %v, want %v", gotPets, wantPets)
	}

	if _, ok := got["owners"]; !ok {
		t.Errorf("quoted table name not parsed, got tables %v", tableNames(got))
	}

	// 実データ由来の回帰ケース2: 文字列リテラル内の閉じ括弧・カンマ・セミコロンを
	// 構文要素と誤認しないこと（001_init.sql の line_reservation_settings
	// additional_fields の JSON DEFAULT に "例) 090-1234-5678" が含まれる）。
	quoted := `
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    additional_fields jsonb NOT NULL DEFAULT '[
        {"key":"phone","label":"電話番号","placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","placeholder":""}
    ]',
    line_channel_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
`
	gotQuoted, err := parseDDLTableColumns(quoted)
	if err != nil {
		t.Fatalf("parseDDLTableColumns(quoted): %v", err)
	}
	wantQuoted := []string{"additional_fields", "created_at", "id", "line_channel_id"}
	gotSettings := append([]string(nil), gotQuoted["settings"]...)
	sort.Strings(gotSettings)
	if !reflect.DeepEqual(gotSettings, wantQuoted) {
		t.Errorf("settings columns = %v, want %v "+
			"(columns after a quoted literal containing ')' and ',' were dropped)",
			gotSettings, wantQuoted)
	}
	for _, decoy := range []string{"decoys", "more_decoys"} {
		if _, ok := got[decoy]; ok {
			t.Errorf("comment content %q was parsed as DDL", decoy)
		}
	}
}

// ---------------------------------------------------------------------------
// Pure core
// ---------------------------------------------------------------------------

// reconcileModelDDLColumns は model が要求する列が DDL に存在するかを一方向で検証する。
func reconcileModelDDLColumns(models []modelTable, ddl map[string][]string) []modelDDLDriftViolation {
	var violations []modelDDLDriftViolation
	for _, m := range models {
		ddlColumns, ok := ddl[m.Table]
		if !ok {
			violations = append(violations, modelDDLDriftViolation{
				StructName: m.StructName,
				Table:      m.Table,
				Column:     "*",
				Reason:     "no CREATE TABLE for this table in migrations",
			})
			continue
		}
		present := make(map[string]struct{}, len(ddlColumns))
		for _, c := range ddlColumns {
			present[c] = struct{}{}
		}
		for _, col := range m.Columns {
			if _, ok := present[col]; !ok {
				violations = append(violations, modelDDLDriftViolation{
					StructName: m.StructName,
					Table:      m.Table,
					Column:     col,
					Reason:     "column declared by GORM model but absent from migration DDL",
				})
			}
		}
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Table != violations[j].Table {
			return violations[i].Table < violations[j].Table
		}
		return violations[i].Column < violations[j].Column
	})
	return violations
}

// extractModelTables は (filename -> source) から、TableName() を宣言する構造体ごとに
// GORM が生成する列名集合を導出する純関数。
func extractModelTables(sources map[string][]byte) ([]modelTable, error) {
	fset := token.NewFileSet()

	structs := map[string]*ast.StructType{}
	tableNameByStruct := map[string]string{}

	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		file, err := parser.ParseFile(fset, name, sources[name], 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if st, ok := ts.Type.(*ast.StructType); ok {
						structs[ts.Name.Name] = st
					}
				}
			case *ast.FuncDecl:
				if recv, table, ok := parseTableNameMethod(d); ok {
					tableNameByStruct[recv] = table
				}
			}
		}
	}

	ns := schema.NamingStrategy{}
	result := make([]modelTable, 0, len(tableNameByStruct))
	for structName, table := range tableNameByStruct {
		st, ok := structs[structName]
		if !ok {
			continue
		}
		columns, err := modelStructColumns(structName, st, structs, ns)
		if err != nil {
			return nil, err
		}
		sort.Strings(columns)
		result = append(result, modelTable{StructName: structName, Table: table, Columns: columns})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StructName < result[j].StructName })
	return result, nil
}

// parseTableNameMethod は `func (X) TableName() string { return "t" }` から (X, t) を取り出す。
func parseTableNameMethod(d *ast.FuncDecl) (recv, table string, ok bool) {
	if d.Name == nil || d.Name.Name != "TableName" || d.Recv == nil || len(d.Recv.List) != 1 {
		return "", "", false
	}
	recvType := d.Recv.List[0].Type
	if star, isStar := recvType.(*ast.StarExpr); isStar {
		recvType = star.X
	}
	ident, isIdent := recvType.(*ast.Ident)
	if !isIdent {
		return "", "", false
	}
	if d.Body == nil || len(d.Body.List) != 1 {
		return "", "", false
	}
	ret, isRet := d.Body.List[0].(*ast.ReturnStmt)
	if !isRet || len(ret.Results) != 1 {
		return "", "", false
	}
	lit, isLit := ret.Results[0].(*ast.BasicLit)
	if !isLit || lit.Kind != token.STRING {
		return "", "", false
	}
	value := strings.Trim(lit.Value, "`\"")
	if value == "" {
		return "", "", false
	}
	return ident.Name, value, true
}

// modelStructColumns は構造体フィールドから GORM の DB 列名集合を導出する。
//
// 埋め込み（匿名）フィールドが現れた場合はエラーを返す。2026-07-27 時点で
// internal/model に埋め込みフィールドと gorm.Model は1件も存在しないため実害は無いが、
// 黙って読み飛ばすと将来 gorm.Model 等が導入された際に埋め込み側の列
// （id / created_at / updated_at / deleted_at）を取りこぼし、
// 「ドリフトがあるのにゲートが緑」という偽陰性になる。失敗させて対応を促す。
func modelStructColumns(
	structName string,
	st *ast.StructType,
	structs map[string]*ast.StructType,
	ns schema.NamingStrategy,
) ([]string, error) {
	var columns []string
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf(
				"model %s has an embedded field; modelStructColumns cannot derive its columns. "+
					"Extend the extractor to flatten embedded structs before relying on this gate",
				structName)
		}
		tag := gormTag(field)
		if tag == "-" || hasGormTagKey(tag, "-") {
			continue
		}
		// gorm:"embedded" 付きの named field は、その構造体のフィールドが
		// フラット化されて実列になる（embeddedPrefix 併用で prefix_xxx 列）。
		// 下の isRelationType は「同一 package の構造体型 = 関連（列なし）」と
		// 判定するため、embedded タグ付き field を素通りさせるとその列が
		// 突合対象から永久に消える＝本ゲートが検出すべき障害クラスそのものが
		// ゲートの死角になる。匿名 embedding と同様に失敗させる。
		if hasGormTagKey(tag, "embedded") {
			return nil, fmt.Errorf(
				"model %s has a gorm:\"embedded\" field; modelStructColumns cannot derive its "+
					"flattened columns. Extend the extractor to expand embedded structs "+
					"(honouring embeddedPrefix) before relying on this gate",
				structName)
		}
		// 関連（foreignKey/many2many）は列を持たない
		if hasGormTagKey(tag, "foreignkey") || hasGormTagKey(tag, "many2many") {
			continue
		}
		if isRelationType(field.Type, structs) {
			continue
		}
		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}
			if col, ok := gormTagValue(tag, "column"); ok {
				columns = append(columns, col)
				continue
			}
			columns = append(columns, ns.ColumnName("", name.Name))
		}
	}
	return columns, nil
}

// isRelationType は「その型が他 model 構造体への関連であり列を持たない」かを判定する。
// gorm.DeletedAt / time.Time / *string 等のスカラは false（＝列）を返す。
func isRelationType(expr ast.Expr, structs map[string]*ast.StructType) bool {
	switch t := expr.(type) {
	case *ast.ArrayType:
		// []byte はスカラ列。それ以外のスライスは has-many 関連。
		if ident, ok := t.Elt.(*ast.Ident); ok && ident.Name == "byte" {
			return false
		}
		return true
	case *ast.StarExpr:
		return isRelationType(t.X, structs)
	case *ast.Ident:
		// 同一 package 内で struct として宣言されている型のみ関連とみなす。
		_, ok := structs[t.Name]
		return ok
	}
	// *ast.SelectorExpr (gorm.DeletedAt, time.Time 等) は列として扱う。
	return false
}

func gormTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
	return tag.Get("gorm")
}

func hasGormTagKey(tag, key string) bool {
	for _, part := range strings.Split(tag, ";") {
		name := strings.TrimSpace(part)
		if idx := strings.Index(name, ":"); idx >= 0 {
			name = name[:idx]
		}
		if strings.EqualFold(strings.TrimSpace(name), key) {
			return true
		}
	}
	return false
}

func gormTagValue(tag, key string) (string, bool) {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, ":")
		if idx < 0 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(part[:idx]), key) {
			value := strings.TrimSpace(part[idx+1:])
			if value != "" {
				return value, true
			}
		}
	}
	return "", false
}

// ---------------------------------------------------------------------------
// DDL parsing
//
// 注: seed_csv_schema_drift_test.go にも同種の CREATE/ALTER 列抽出があるが、
// 2026-07-27 時点でそのファイルは未コミット（別セッションの作業中成果物）である。
// 未追跡ファイルの helper に依存すると、本ファイルだけを commit した際に
// main の lintscan package がコンパイル不能になるため、意図的に独立実装とする。
// 両者が commit 済みで安定したら統合してよい。
// ---------------------------------------------------------------------------

var errUnterminatedDDL = errors.New("unterminated quote or comment in DDL")

// parseDDLTableColumns は CREATE TABLE と ALTER TABLE ... ADD COLUMN から
// テーブル名 -> 列名集合を抽出する。
func parseDDLTableColumns(sql string) (map[string][]string, error) {
	stripped, err := stripDDLComments(sql)
	if err != nil {
		return nil, err
	}

	tables := map[string][]string{}
	add := func(table, column string) {
		table, column = normalizeDDLIdentifier(table), normalizeDDLIdentifier(column)
		if table == "" || column == "" {
			return
		}
		for _, existing := range tables[table] {
			if existing == column {
				return
			}
		}
		tables[table] = append(tables[table], column)
	}

	for _, stmt := range splitDDLStatements(stripped) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		upper := strings.ToUpper(stmt)
		switch {
		case strings.HasPrefix(upper, "CREATE TABLE"):
			table, body, ok := splitCreateTable(stmt)
			if !ok {
				continue
			}
			// CREATE TABLE は必ず登録する（列ゼロでもテーブル存在は真）
			if t := normalizeDDLIdentifier(table); t != "" {
				if _, exists := tables[t]; !exists {
					tables[t] = nil
				}
			}
			for _, clause := range splitTopLevelCommas(body) {
				if col, ok := createTableColumnName(clause); ok {
					add(table, col)
				}
			}
		case strings.HasPrefix(upper, "ALTER TABLE"):
			if table, columns, ok := parseAlterTableAddColumns(stmt); ok {
				for _, col := range columns {
					add(table, col)
				}
			}
		}
	}
	return tables, nil
}

func splitCreateTable(stmt string) (table, body string, ok bool) {
	open := strings.Index(stmt, "(")
	if open < 0 {
		return "", "", false
	}
	head := stmt[:open]
	closeIdx := matchingParen(stmt, open)
	if closeIdx < 0 {
		return "", "", false
	}

	fields := strings.Fields(head)
	// CREATE TABLE [IF NOT EXISTS] <name>
	if len(fields) < 3 {
		return "", "", false
	}
	table = fields[len(fields)-1]
	return table, stmt[open+1 : closeIdx], true
}

// createTableColumnName は CREATE TABLE の1句から列名を返す。
// テーブル制約句 (PRIMARY KEY / CONSTRAINT / UNIQUE ...) は列ではないので false。
func createTableColumnName(clause string) (string, bool) {
	clause = strings.TrimSpace(clause)
	if clause == "" {
		return "", false
	}
	fields := strings.Fields(clause)
	if len(fields) == 0 {
		return "", false
	}
	first := strings.ToUpper(strings.Trim(fields[0], `"`))
	switch first {
	case "CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "CHECK", "EXCLUDE", "LIKE":
		return "", false
	}
	return fields[0], true
}

// parseAlterTableAddColumns は ALTER TABLE 文から追加される列をすべて返す。
//
// PostgreSQL の ALTER TABLE は複数アクションをカンマ区切りで取れる。001_init.sql の
// treatments には
//
//	ALTER TABLE treatments
//	  ALTER COLUMN quantity TYPE numeric(10,2),
//	  ADD COLUMN dose_weight_kg numeric(6,2),
//	  ADD COLUMN dose_amount_mg numeric(12,6), ...;
//
// のように先頭が ALTER COLUMN で後続に ADD COLUMN が続く形がある。先頭アクションだけを
// 見る実装では dose_* 列を丸ごと取りこぼす。
func parseAlterTableAddColumns(stmt string) (table string, columns []string, ok bool) {
	fields := strings.Fields(stmt)
	if len(fields) < 4 || !strings.EqualFold(fields[0], "ALTER") || !strings.EqualFold(fields[1], "TABLE") {
		return "", nil, false
	}

	// ALTER TABLE [IF EXISTS] [ONLY] <name>
	idx := 2
	for idx < len(fields) {
		switch strings.ToUpper(fields[idx]) {
		case "IF", "EXISTS", "ONLY":
			idx++
			continue
		}
		break
	}
	if idx >= len(fields) {
		return "", nil, false
	}
	table = fields[idx]

	// 表名の直後からアクションリストが始まる。トップレベルのカンマで分割する
	// （型の numeric(12,6) 等を割らないため）。
	//
	// 位置の特定に strings.Index(stmt, table) を使ってはならない。表名が先行トークンの
	// 部分文字列になり得る（例: 表名 "able" は "ALTER TABLE" の "TABLE" 内に一致する）。
	// トークン境界で数えた offset を使う。
	tableEnd := offsetAfterToken(stmt, idx)
	if tableEnd < 0 {
		return "", nil, false
	}
	actions := splitTopLevelCommas(stmt[tableEnd:])

	for _, action := range actions {
		af := strings.Fields(action)
		if len(af) == 0 || !strings.EqualFold(af[0], "ADD") {
			continue
		}
		j := 1
		if j < len(af) && strings.EqualFold(af[j], "COLUMN") {
			j++
		}
		if j+2 < len(af) &&
			strings.EqualFold(af[j], "IF") &&
			strings.EqualFold(af[j+1], "NOT") &&
			strings.EqualFold(af[j+2], "EXISTS") {
			j += 3
		}
		if j >= len(af) {
			continue
		}
		// ADD CONSTRAINT / PRIMARY KEY / UNIQUE ... は列ではない
		switch strings.ToUpper(strings.Trim(af[j], `"`)) {
		case "CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "CHECK", "EXCLUDE":
			continue
		}
		columns = append(columns, af[j])
	}
	return table, columns, len(columns) > 0
}

// offsetAfterToken は空白区切りで数えて n 番目（0 始まり）のトークンの直後の
// byte offset を返す。トークンが足りない場合は -1。
func offsetAfterToken(s string, n int) int {
	i, seen := 0, 0
	for i < len(s) {
		for i < len(s) && isASCIISpace(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}
		for i < len(s) && !isASCIISpace(s[i]) {
			i++
		}
		if seen == n {
			return i
		}
		seen++
	}
	return -1
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

func normalizeDDLIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	identifier = strings.TrimSuffix(identifier, ",")
	if strings.HasPrefix(identifier, `"`) && strings.HasSuffix(identifier, `"`) && len(identifier) >= 2 {
		return strings.ReplaceAll(identifier[1:len(identifier)-1], `""`, `"`)
	}
	// public.pets のような修飾名は最後の要素を使う
	if idx := strings.LastIndex(identifier, "."); idx >= 0 {
		identifier = identifier[idx+1:]
	}
	return strings.ToLower(identifier)
}

func stripDDLComments(sql string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(sql); {
		switch {
		case sql[i] == '\'' || sql[i] == '"':
			quote := sql[i]
			b.WriteByte(sql[i])
			i++
			for i < len(sql) {
				b.WriteByte(sql[i])
				if sql[i] == quote {
					// '' / "" によるエスケープ
					if i+1 < len(sql) && sql[i+1] == quote {
						b.WriteByte(sql[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
		case strings.HasPrefix(sql[i:], "--"):
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
		case strings.HasPrefix(sql[i:], "/*"):
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 {
				return "", errUnterminatedDDL
			}
			i += 2 + end + 2
		default:
			b.WriteByte(sql[i])
			i++
		}
	}
	return b.String(), nil
}

// skipQuotedLiteral は s[i] が引用符の開始ならその文字列リテラルを読み飛ばし、
// 閉じ引用符の次の位置を返す。引用符でなければ i をそのまま返す。
//
// これが無いと DDL 内の文字列リテラルに含まれる括弧・カンマ・セミコロンを
// 構文要素と誤認する。実例: 001_init.sql の line_reservation_settings.additional_fields
// の DEFAULT は JSON 文字列で、内部に "例) 090-1234-5678" という閉じ括弧単独と
// 多数のカンマを含む。quote 非対応スキャナはここで CREATE TABLE を早期終了し、
// 以降の列（line_channel_id・created_at 等）を丸ごと取りこぼす。
func skipQuotedLiteral(s string, i int) int {
	if i >= len(s) {
		return i
	}
	quote := s[i]
	if quote != '\'' && quote != '"' {
		return i
	}
	for j := i + 1; j < len(s); j++ {
		if s[j] != quote {
			continue
		}
		// '' / "" は引用符自身のエスケープ
		if j+1 < len(s) && s[j+1] == quote {
			j++
			continue
		}
		return j + 1
	}
	return len(s) // 終端されていない場合は末尾まで消費する
}

func matchingParen(s string, open int) int {
	depth := 0
	for i := open; i < len(s); {
		if next := skipQuotedLiteral(s, i); next != i {
			i = next
			continue
		}
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

func splitTopLevelCommas(body string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(body); {
		if next := skipQuotedLiteral(body, i); next != i {
			i = next
			continue
		}
		switch body[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
		i++
	}
	parts = append(parts, body[start:])
	return parts
}

// splitDDLStatements は引用符を尊重して文をセミコロンで分割する。
func splitDDLStatements(sql string) []string {
	var stmts []string
	start := 0
	for i := 0; i < len(sql); {
		if next := skipQuotedLiteral(sql, i); next != i {
			i = next
			continue
		}
		if sql[i] == ';' {
			stmts = append(stmts, sql[start:i])
			start = i + 1
		}
		i++
	}
	return append(stmts, sql[start:])
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func mustReadGoSources(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	sources := map[string][]byte{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		sources[e.Name()] = raw
	}
	return sources
}

func mustReadMigrationSQL(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	var b strings.Builder
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		b.Write(raw)
		b.WriteString("\n")
	}
	return b.String()
}

func tableNames(m map[string][]string) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
