package apicontract

// openapi_date_format_drift_test.go — BE-refactor.md R3-3 (D9・P1): OpenAPI ↔ response DTO
// serialization format drift の検出ゲート。
//
// ─── Background ──────────────────────────────────────────────────────────────────────
//
// docs/api.yaml が `format: date`（日付のみ）と宣言したプロパティを、target package の response 構造体が
// Go の time.Time / *time.Time フィールドで持つと、JSON encoding は RFC3339 datetime
// （`2020-01-15T00:00:00+09:00`）になり、宣言（`2020-01-15`）と wire 表現が乖離する。これは
// R2-1（inventory expiry_date）で顕在化したバグクラスそのもの。
//
// ⚠️ 実測（2026-07-02）: BE-refactor.md R3-3 が前提とした「6/30 調査で format↔実装は整合（0/76）」は
// 現 HEAD では成立しない。docs/api.yaml が `format: date` を宣言する date 系フィールド（birth_date /
// last_visit / neutered_date / date / scheduled_date / valid_until / expiry_date …）の多くが、
// target response DTO では *time.Time（datetime wire）で配信されており、response 側だけで 22 箇所の drift が既存する
// （下記 knownDateFormatDrifts）。これは本タスクが導入する前から存在する systemic な状態で、
// 「openapi を date-time に直す」か「response DTO を date-only 文字列にする」かは FE との contract 判断
// （PO follow-up）。本ゲートはその判断を下さず、現状を allowlist に固定して **CI 可視化** し、
// **新規 drift の混入と、allowlist と実態の乖離**を fail させる（migration_cascade_lint と同枠組み）。
//
// ─── Scope ──────────────────────────────────────────────────────────────────────────
//
// - 対象: response serialization（target package の *_response.go）。request DTO（*_request.go）は入力
//   binding の別関心事のため対象外。
// - 検出クラス: openapi `format: date`（date-only）↔ response time.Time/*time.Time（datetime wire）。
//   逆方向（date-time 宣言を date-only 文字列で配信）は severity が低く別途（対象外）。
// - keying は json タグ名（openapi property 名）。schema 単位の厳密対応はしない（drift 検出には十分・
//   name 衝突は allowlist で吸収）。

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const openapiPath = "../../docs/api.yaml"

// responseScanDirs enumerates every target domain package that owns HTTP response DTOs.
// Keys include the package-relative path so same-named files in different domains cannot
// collapse into one allowlist entry.
var responseScanDirs = []string{
	"../auth",
	"../billing",
	"../clinic",
	"../inventory",
	"../lstep",
	"../manualarticle",
	"../medicalrecord",
	"../owner",
	"../pet",
	"../reservation",
	"../staff",
	"../trimming",
}

// knownDateFormatDrifts は現 HEAD に存在する「openapi format:date ↔ response time.Time」drift を
// (file, json名) → 出現数 で固定する。値は response 側の time.Time/*time.Time フィールド数。
// 新規 drift（未登録キー）・出現数変化・stale エントリ（drift が解消された）はすべて fail。
//
// これらは本ゲート導入前から存在する既存 drift。解消（openapi を date-time にするか response DTO を
// date-only 文字列にするか）は FE contract 判断の follow-up。解消した際は該当エントリを削除すること。
var knownDateFormatDrifts = map[string]int{
	"billing/accounting_response.go|scheduled_date":        1,
	"billing/estimate_response.go|valid_until":             1,
	"inventory/inventory_response.go|expiry_date":          1,
	"inventory/inventory_response.go|last_restocked":       1,
	"medicalrecord/daily_record_response.go|date":          1,
	"medicalrecord/examination_response.go|date":           1,
	"medicalrecord/hospitalization_response.go|start_date": 1,
	"medicalrecord/hospitalization_response.go|end_date":   1,
	"medicalrecord/medical_record_response.go|date":        1,
	"medicalrecord/treatment_response.go|date":             1,
	"medicalrecord/vaccination_response.go|date":           1,
	"medicalrecord/vaccination_response.go|next_date":      1,
	"owner/http_response.go|birth_date":                    2,
	"owner/http_response.go|neutered_date":                 1,
	"owner/http_response.go|last_visit":                    1,
	"pet/pet_response.go|birth_date":                       2,
	"pet/pet_response.go|neutered_date":                    2,
	"pet/pet_response.go|last_visit":                       2,

	// pet/pet_response.go|first_visit_date: benign json-name collision, NOT a real drift.
	// docs/api.yaml declares two unrelated `first_visit_date` properties: `PetFirstVisit`
	// (line ~1126, `format: date-time`, correctly backed by pet/pet_response.go's
	// petFirstVisitResponse.FirstVisitDate *time.Time) and the newer owner-aggregation
	// schema (line ~7977, `format: date`, correctly backed by the lstep aggregation
	// response's FirstVisitDate *string, already date-only).
	// dateOnly-prop parsing (parseOpenAPIDateOnlyProps) and the response AST scan
	// (analyzeResponseFileDateDrift) both key by bare json tag name only, not by schema/
	// struct identity, so the new `format: date` property makes the matcher pick up
	// pet/pet_response.go's *unrelated*, pre-existing, correctly date-time-typed field as if
	// it were drift against the new date-only schema. It isn't: pet/pet_response.go never
	// serves the aggregation endpoint, and the lstep aggregation response already serves date-only
	// correctly. Pinned here rather than reworking the matcher to be schema-scoped (would
	// require ownership tracking not worth it for one collision).
	"pet/pet_response.go|first_visit_date": 1,

	// owner/http_response.go|deceased_at, pet/pet_response.go|deceased_at: benign json-name collision,
	// NOT a real drift (same class as first_visit_date above). docs/api.yaml declares two
	// unrelated `deceased_at` properties: `PatchPetDeathRequest.deceased_at` (request DTO,
	// `format: date`, date-only — matches lstep_lifecycle_request.go's jsonDate input type)
	// and the response-side `deceased_at` on OwnerPetSummary/Pet (`format: date-time`,
	// correctly backed by owner/http_response.go and pet/pet_response.go's *time.Time,
	// since pets.deceased_at is a `timestamptz` column, not a date column). The matcher keys
	// by bare json tag name only (not by schema/request-vs-response identity), so the
	// pre-existing date-only request property makes it flag the new datetime response field.
	// It isn't real drift: the response schemas declare `format: date-time` for deceased_at,
	// matching the target response DTO's time.Time wire format exactly.
	"owner/http_response.go|deceased_at": 1,
	"pet/pet_response.go|deceased_at":    1,
}

func driftKey(file, jsonName string) string { return file + "|" + jsonName }

// ─── OpenAPI parsing ─────────────────────────────────────────────────────────────────

// parseOpenAPIDateOnlyProps は docs/api.yaml から「date-only」body スキーマプロパティ名の集合を返す。
//
// 実 YAML（gopkg.in/yaml.v3）でツリーを構築し、`properties:` マッピング配下の各プロパティ名のうち
// 値が `type: string` かつ `format: date` の mapping であるものを収集する。`properties:` 配下だけを
// 見ることで、path parameter の `schema: {type: string, format: date}`（structural key `schema` が
// leaf に見える形）や `parameters`/`schema` そのものを構造的に除外する（go-reviewer HIGH-1 の是正）。
// 行順・コメント・nullable の位置に非依存（YAML パーサがトークン化するため。HIGH-2 の是正）。
// `format: date-time` は type/format 不一致で自然に除外される。
func parseOpenAPIDateOnlyProps(yamlSrc []byte) (map[string]struct{}, error) {
	var root any
	if err := yaml.Unmarshal(yamlSrc, &root); err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	var walk func(node any)
	walk = func(node any) {
		switch n := node.(type) {
		case map[string]any:
			if props, ok := n["properties"].(map[string]any); ok {
				for name, v := range props {
					if pm, ok := v.(map[string]any); ok {
						if pm["type"] == "string" && pm["format"] == "date" {
							out[name] = struct{}{}
						}
					}
				}
			}
			for _, v := range n {
				walk(v)
			}
		case []any:
			for _, e := range n {
				walk(e)
			}
		}
	}
	walk(root)
	return out, nil
}

// ─── Response DTO AST scan ───────────────────────────────────────────────────────────

type driftFinding struct {
	file     string
	jsonName string
}

// analyzeResponseFileDateDrift は1つの Go ソースから、time.Time/*time.Time 型で json タグ名が
// dateOnly 集合に属するフィールドを検出する（純粋関数・fixture と実ソースで同一ロジック）。
// go-reviewer M-1: トップレベルの named type 宣言（response DTO 構造体）のみを対象とし、関数内
// ローカル/インラインの匿名 struct（ログ用途等・偶然 time.Time+date タグを持ちうる）は無視する。
func analyzeResponseFileDateDrift(filename string, src []byte, dateOnly map[string]struct{}) ([]driftFinding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, err
	}
	fileKey := filepath.ToSlash(filename)
	var findings []driftFinding
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				continue
			}
			for _, field := range st.Fields.List {
				if field.Tag == nil || !isTimeTimeType(field.Type) {
					continue
				}
				tag, err := strconv.Unquote(field.Tag.Value)
				if err != nil {
					continue
				}
				name, _, _ := strings.Cut(reflect.StructTag(tag).Get("json"), ",")
				if name == "" || name == "-" {
					continue
				}
				if _, isDateOnly := dateOnly[name]; isDateOnly {
					findings = append(
						findings,
						driftFinding{file: fileKey, jsonName: name},
					)
				}
			}
		}
	}
	return findings, nil
}

// isTimeTimeType は expr が time.Time または *time.Time かを判定する。
// go-reviewer M-2: AST-only 判定のため `import t "time"` のエイリアスや `type Date = time.Time` の
// 型エイリアスは見逃す（go/types なしの構造的限界）。現行 target response package にエイリアスは無く、
// sql.NullTime / civil.Date 等の別 time 型は正しく除外される。エイリアス導入時は本判定の拡張が要る。
func isTimeTimeType(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "time" && sel.Sel.Name == "Time"
}

// walkResponseDrifts は responseScanDirs 配下の *_response.go 全ソースを走査し drift を集計する。
func walkResponseDrifts(t *testing.T, dateOnly map[string]struct{}) map[string]int {
	t.Helper()
	agg := map[string]int{}
	for _, dir := range responseScanDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*_response.go"))
		if err != nil {
			t.Fatalf("glob response files in %s: %v", dir, err)
		}
		// glob "*_response.go" は *_response_test.go にマッチしない（末尾が _test.go のため）ので
		// test ファイル除外の追加チェックは不要。
		for _, fp := range files {
			src, err := os.ReadFile(fp) //nolint:gosec // fixed source dirs enumerated in this test, not untrusted input
			if err != nil {
				t.Fatalf("read %s: %v", fp, err)
			}
			relativePath, err := filepath.Rel("..", fp)
			if err != nil {
				t.Fatalf("resolve response path %s: %v", fp, err)
			}
			findings, err := analyzeResponseFileDateDrift(
				relativePath,
				src,
				dateOnly,
			)
			if err != nil {
				t.Fatalf("parse %s: %v", fp, err)
			}
			for _, fnd := range findings {
				agg[driftKey(fnd.file, fnd.jsonName)]++
			}
		}
	}
	return agg
}

// ─── Reconciliation (pure) ───────────────────────────────────────────────────────────

func reconcileDateFormatDrift(found, allow map[string]int) []string {
	var violations []string
	for key, cnt := range found {
		allowed, ok := allow[key]
		switch {
		case !ok:
			violations = append(violations,
				"NEW openapi format:date ↔ response time.Time drift at "+key+" (count="+strconv.Itoa(cnt)+"). "+
					"Serve this field as a date-only string (In(time.Local).Format(\"2006-01-02\")) to match the "+
					"OpenAPI `format: date` declaration, or (if the wire format should be datetime) change the "+
					"OpenAPI declaration to `format: date-time`. If this is an accepted pre-existing drift, add it "+
					"to knownDateFormatDrifts with a follow-up note.")
		case cnt != allowed:
			violations = append(violations,
				"drift count changed at "+key+": found "+strconv.Itoa(cnt)+", allowlist pins "+strconv.Itoa(allowed)+
					". Re-review; update knownDateFormatDrifts only if the change is intended.")
		}
	}
	for key, allowed := range allow {
		if _, ok := found[key]; !ok {
			violations = append(violations,
				"stale allowlist entry "+key+" (pinned "+strconv.Itoa(allowed)+") no longer matches a real drift "+
					"— the drift was fixed or the field/name changed; delete the entry from knownDateFormatDrifts.")
		}
	}
	return violations
}

// ─── Gate tests ──────────────────────────────────────────────────────────────────────

// TestOpenAPIDateFormatDrift_MatchesAllowlist is the gate: every openapi format:date ↔ response
// time.Time drift must be on the pinned allowlist; no new drift, no stale entry. Floors guard
// against a vacuous pass if openapi parsing or the response glob silently breaks.
func TestOpenAPIDateFormatDrift_MatchesAllowlist(t *testing.T) {
	yamlSrc, err := os.ReadFile(openapiPath) //nolint:gosec // fixed docs path, not untrusted input
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	dateOnly, err := parseOpenAPIDateOnlyProps(yamlSrc)
	if err != nil {
		t.Fatalf("parse openapi: %v", err)
	}
	if len(dateOnly) < 8 {
		t.Fatalf("only %d date-only openapi props parsed; parser likely broke (expected the birth_date/date/"+
			"last_visit/... set). Would vacuously pass.", len(dateOnly))
	}
	// sanity: known date-only names must be present
	for _, must := range []string{"birth_date", "expiry_date", "date", "valid_until"} {
		if _, ok := dateOnly[must]; !ok {
			t.Fatalf("expected date-only openapi prop %q not parsed; parser regression", must)
		}
	}

	found := walkResponseDrifts(t, dateOnly)
	if len(found) < 10 {
		t.Fatalf("only %d response date-drift sites found; AST scan or glob likely broke (expected ~18 keys). "+
			"Would vacuously pass.", len(found))
	}

	for _, v := range reconcileDateFormatDrift(found, knownDateFormatDrifts) {
		t.Error(v)
	}
}

// TestOpenAPIDateFormatDrift_OpenAPIParser pins the openapi date-only parser on inline fixtures:
// a `type: string` + `format: date` leaf is date-only; `format: date-time` and structural keys are not.
func TestOpenAPIDateFormatDrift_OpenAPIParser(t *testing.T) {
	// go-reviewer HIGH-1 の反例を含む: query parameter の `schema: {type:string, format:date}` は
	// `properties:` 配下ではないため date-only に混入してはならない。また go-reviewer HIGH-2 の
	// key 順序反転（format が type より先）も、real YAML parse なので正しく検出されること。
	src := []byte("" +
		"paths:\n" +
		"  /pets:\n" +
		"    get:\n" +
		"      parameters:\n" +
		"        - name: date\n" +
		"          in: query\n" +
		"          schema:\n" +
		"            type: string\n" +
		"            format: date\n" +
		"components:\n" +
		"  schemas:\n" +
		"    Pet:\n" +
		"      properties:\n" +
		"        birth_date:\n" +
		"          type: string\n" +
		"          format: date\n" +
		"          nullable: true\n" +
		"        neutered_date:\n" + // key 順序反転（format が先）でも real parse なら検出
		"          format: date\n" +
		"          type: string\n" +
		"        created_at:\n" +
		"          type: string\n" +
		"          format: date-time\n" +
		"        name:\n" +
		"          type: string\n")
	got, err := parseOpenAPIDateOnlyProps(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, ok := got["birth_date"]; !ok {
		t.Error("birth_date (properties + type:string + format:date) must be date-only")
	}
	if _, ok := got["neutered_date"]; !ok {
		t.Error("neutered_date (format before type) must still be date-only — YAML parse is order-independent (HIGH-2)")
	}
	if _, ok := got["created_at"]; ok {
		t.Error("created_at (format:date-time) must NOT be date-only")
	}
	if _, ok := got["name"]; ok {
		t.Error("name (no format) must NOT be date-only")
	}
	if _, ok := got["schema"]; ok {
		t.Error("`schema` under a query parameter (not under properties:) must NOT be date-only (HIGH-1 regression)")
	}
	if _, ok := got["date"]; ok {
		t.Error("query parameter `name: date` must NOT be collected — only body-schema properties are (HIGH-1)")
	}
}

// TestOpenAPIDateFormatDrift_ResponseAnalyzer pins the response AST scan on inline fixtures:
// time.Time / *time.Time fields whose json name is date-only are detected; string fields and
// non-date-only names are not.
func TestOpenAPIDateFormatDrift_ResponseAnalyzer(t *testing.T) {
	dateOnly := map[string]struct{}{"birth_date": {}, "expiry_date": {}}
	cases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "*time.Time with date-only json name is a drift",
			src: "package p\nimport \"time\"\ntype R struct {\n" +
				"BirthDate *time.Time `json:\"birth_date,omitempty\"`\n}\n",
			want: 1,
		},
		{
			name: "time.Time (value) with date-only json name is a drift",
			src: "package p\nimport \"time\"\ntype R struct {\n" +
				"ExpiryDate time.Time `json:\"expiry_date\"`\n}\n",
			want: 1,
		},
		{
			name: "date-only served as *string is NOT a drift (correct fix)",
			src:  "package p\ntype R struct {\nExpiryDate *string `json:\"expiry_date,omitempty\"`\n}\n",
			want: 0,
		},
		{
			name: "time.Time with non-date-only json name is NOT flagged",
			src: "package p\nimport \"time\"\ntype R struct {\n" +
				"CreatedAt time.Time `json:\"created_at\"`\n}\n",
			want: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, err := analyzeResponseFileDateDrift("fixture_response.go", []byte(tc.src), dateOnly)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

// TestOpenAPIDateFormatDrift_Reconciler pins the gate's failure modes on synthetic inputs.
func TestOpenAPIDateFormatDrift_Reconciler(t *testing.T) {
	base := map[string]int{"pet_response.go|birth_date": 2, "inventory_response.go|expiry_date": 1}

	t.Run("clean baseline reports nothing", func(t *testing.T) {
		if v := reconcileDateFormatDrift(base, base); len(v) != 0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})
	t.Run("new drift fails", func(t *testing.T) {
		found := map[string]int{"pet_response.go|birth_date": 2, "inventory_response.go|expiry_date": 1, "new_response.go|date": 1}
		v := reconcileDateFormatDrift(found, base)
		if len(v) != 1 || !strings.Contains(v[0], "NEW openapi") {
			t.Fatalf("expected new-drift violation, got %v", v)
		}
	})
	t.Run("count change fails", func(t *testing.T) {
		found := map[string]int{"pet_response.go|birth_date": 3, "inventory_response.go|expiry_date": 1}
		v := reconcileDateFormatDrift(found, base)
		if len(v) != 1 || !strings.Contains(v[0], "count changed") {
			t.Fatalf("expected count-change violation, got %v", v)
		}
	})
	t.Run("stale entry fails (drift fixed)", func(t *testing.T) {
		found := map[string]int{"pet_response.go|birth_date": 2}
		v := reconcileDateFormatDrift(found, base)
		if len(v) != 1 || !strings.Contains(v[0], "stale allowlist entry") {
			t.Fatalf("expected stale violation, got %v", v)
		}
	})
}
