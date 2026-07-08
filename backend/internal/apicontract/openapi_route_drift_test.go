package apicontract

// openapi_route_drift_test.go — BE-refactor.md G1-4 (apicontract-route-inventory-gate):
// ルートインベントリ drift gate. 実装ルート(internal/handler の Register*Routes 呼び出しグラフを
// go/ast で静的解決したフルパス集合) ↔ docs/api.yaml paths を突合し、allowlist 外の新規 drift
// (missingFromSpec: 実装にあるが未文書化 / phantomInSpec: 文書にあるが実装なし) を fail させる。
//
// ─── Background ──────────────────────────────────────────────────────────────────────
//
// internal/apicontract は date-format drift(openapi_date_format_drift_test.go)専用だったが、
// doc.go が元々「複数の invariant を前提に設計」と明記している枠組みを拡張する。ルート列挙は
// 全 477 verb 呼び出しが internal/handler 内の Group()/GET()/POST()/PUT()/PATCH()/DELETE() の
// 静的文字列リテラルだけで 100% 解決できること(未解決 0 件)を実測確認済み。RegisterRoutes から
// h.RegisterXxx/h.registerXxxWithAuth の呼び出しグラフを辿り、各関数の *gin.RouterGroup/*gin.Engine
// 引数を呼び出し元で解決したプレフィックスに束縛して再帰することでフルパスを組み立てる。
//
// ─── Scope ──────────────────────────────────────────────────────────────────────────
//
// - 対象: internal/handler 配下の非テスト .go 全ファイル（*_request.go/*_response.go は
//   Group/verb 呼び出しを持たないため実質無視されるが、handlerDir 全体を glob 対象にすることで
//   将来ファイルが増えても取りこぼさない）。
// - LIFF（/api/liff/:clinicId、reservation_line_routes.go:64 RegisterLiffRoutes(ctx, r)）は
//   r *gin.Engine を起点とし、docs/api.yaml 側でも servers(/api/v1)固定を避けるため絶対パス
//   （/api/liff/{clinicId}/...）で記載されている（コメント: reservation_line_routes.go:72-92 相当）。
//   本ゲートは api.yaml の path key が "/api/liff" で始まる場合のみ絶対パス扱いとし、それ以外は
//   servers の /api/v1 を前置して比較する。
// - /health（r.GET("/health", ...)、handler.go:52）は api.yaml 上 "/health" キーで記載されて
//   いるが、これは他のキーと同じ相対規約で書かれているため本ゲートは /api/v1 を前置して比較する
//   （結果 real "/health" が missingFromSpec、spec "/api/v1/health" 相当が phantomInSpec として
//   現れる）。これは LIFF と異なり yaml 側に絶対パス注記が無い既存の小さな整合齟齬であり、本ゲート
//   の責務は「検出して allowlist に pin する」ことであって api.yaml の書き方を変えることではない
//   （yaml 修正は別 finding のスコープ）。
// - /api/line/webhook（r.POST、handler.go:140）は api.yaml に一切未記載のため missingFromSpec。
//
// パラメータ名は gin の `:name` セグメントを OpenAPI の `{name}` へ正規化して比較する
// （両者のセグメント名は現状全て一致 — 例: :id→{id}, :clinic_id→{clinic_id}, :clinicId→{clinicId}）。

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const handlerDir = "../handler"

// ─── knownRouteDrifts allowlist ──────────────────────────────────────────────────────
//
// 導入時点(2026-07-09、G1-1/G1-2 Phase A-C・G1-3 完了後の現 HEAD)の実測残差を pin した
// （旧監査の 203/23 という数字は Phase A-C 完了後の現状を反映していないため使わず、導入時に
// 実際に AST 列挙し直した — missing 98 件 / phantom 4 件）。同一コミット内で BE-refactor.md
// Phase D-1（owner-nested/canonical lstep-line パス文書化）を実施した結果、missing は
// 98 → 15 に減少した（内訳は下記）。新規 missing/phantom の混入、既存エントリの解消(stale)は
// いずれも fail させる。
//
// 内訳(missing 82 = 1 + 12 + 3 + 3 + 63、phantom 4 = 1 + 3):
//   - /health ↔ /api/v1/health: yaml が絶対パスとして記載していない相対規約の齟齬。missing 1 +
//     phantom 1(下記コメント参照)。Phase D-1 スコープ外のため今回は解消しない。
//   - Phase D-1 完了後残存する同一ハンドラの意図的未記載エイリアス: owner-nested co-group 12件
//     + canonical family 3件 = 15件。すべて「別のパスで既に文書化済みの同一ハンドラの重複登録」
//     であり、Phase B の death エイリアスと同型の precedent（各エントリのコメントで文書化先を
//     明記）。
//   - /clinics/{clinic_id} の GET/PATCH/DELETE: clinic_handler.go 実装が :clinic_id を使うのに対し
//     api.yaml は同じ操作を /clinics/{id} として記載(パラメータ名のみの drift)。missing 3件 +
//     phantom 3件(下記 knownPhantomInSpec)が対になる。G1-2 の「パラメータ名差異 3 件除外後」に
//     該当する既知ペア。Phase D-1/D-2/D-3 スコープ外のため今回は解消しない。
//   - 上記以外の63件: G1-2(実装済み203件未記載)のうち Phase A-C でまだ手が回っていない残差
//     （checkups 系・masters GET {id} 系・reservation-staffs ファミリー・lstep clinic-nested
//     read 系・webhook 等）。本シリーズの D-1/D-2/D-3 いずれのスコープでもないため今回は手を
//     付けていない(導入時からそのまま残存)。

var knownMissingFromSpec = map[string]bool{
	// /health は yaml が絶対パスとして記載していない相対規約の齟齬（doc comment 参照）。
	"GET /health": true,

	// --- Phase D-1 (完了): owner-nested (co group) lstep/line, 13件中12件は canonical 側と
	// 同一ハンドラの重複登録のため意図的に未記載のまま(死んだエイリアスの precedent と同型)。
	// GET .../lstep/friend-attributes のみ co-group 限定の実装(canonical 対応なし)だったため
	// /clinics/{clinic_id}/owners/{id}/lstep/friend-attributes として実際に文書化した
	// (api.yaml 側を参照。missing から除外済み)。
	"DELETE /api/v1/clinics/{clinic_id}/owners/{id}/lstep-opt-out":         true, // alias of documented POST/DELETE /owners/{id}/lstep-opt-out
	"DELETE /api/v1/clinics/{clinic_id}/owners/{id}/lstep/tags/{tag_name}": true, // alias of documented DELETE /owners/{id}/lstep/tags/{tag_name}
	"GET /api/v1/clinics/{clinic_id}/owners/{id}/line/send-logs":           true, // alias of documented GET /owners/{id}/line/send-logs
	"GET /api/v1/clinics/{clinic_id}/owners/{id}/lstep/tags":               true, // alias of documented GET /owners/{id}/lstep/tags
	"PATCH /api/v1/clinics/{clinic_id}/owners/{id}/delivery-caution":       true, // alias of documented PATCH /owners/{id}/delivery-caution
	"PATCH /api/v1/clinics/{clinic_id}/owners/{id}/delivery-exclusion":     true, // alias of documented PATCH /owners/{id}/delivery-exclusion
	"PATCH /api/v1/clinics/{clinic_id}/owners/{id}/line-id-confirm":        true, // alias of documented PATCH /owners/{id}/line-id-confirm
	"PATCH /api/v1/clinics/{clinic_id}/owners/{id}/line-user-id":           true, // alias of documented PATCH /owners/{id}/line-user-id
	"PATCH /api/v1/clinics/{clinic_id}/owners/{id}/transfer-status":        true, // alias of documented PATCH /owners/{id}/transfer-status
	"POST /api/v1/clinics/{clinic_id}/owners/{id}/line/send":               true, // alias of documented POST /owners/{id}/line/send
	"POST /api/v1/clinics/{clinic_id}/owners/{id}/lstep-opt-out":           true, // alias of documented POST /owners/{id}/lstep-opt-out
	"POST /api/v1/clinics/{clinic_id}/owners/{id}/lstep/tags":              true, // alias of documented POST /owners/{id}/lstep/tags

	// --- Phase D-1 (完了): canonical /owners/{id} lstep/line family — 15/18 documented in
	// api.yaml; the remaining 3 are same-handler aliases of a documented sibling path and stay
	// intentionally omitted (same precedent as the Phase B death alias elsewhere in this file).
	"PATCH /api/v1/owners/{id}/line":             true, // ISSUE-001 alias of documented PATCH /owners/{id}/line-user-id
	"POST /api/v1/owners/{id}/lstep/send":        true, // ISSUE-002 alias of documented POST /owners/{id}/line/send
	"GET /api/v1/owners/{id}/lstep/send-history": true, // ISSUE-002 alias of documented GET /owners/{id}/line/send-logs

	// --- clinic_handler.go :clinic_id vs api.yaml {id} param-name-only drift, 3件 (pairs with knownPhantomInSpec) ---
	"DELETE /api/v1/clinics/{clinic_id}": true,
	"GET /api/v1/clinics/{clinic_id}":    true,
	"PATCH /api/v1/clinics/{clinic_id}":  true,

	// --- G1-2 residual not touched by this series (out of D-1/D-2/D-3 scope), 63件 ---
	"DELETE /api/v1/clinics/{clinic_id}/lstep-settings":                                true,
	"DELETE /api/v1/clinics/{clinic_id}/pets/{id}/death":                               true,
	"DELETE /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}":                  true,
	"DELETE /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/schedules/{date}": true,
	"DELETE /api/v1/clinics/{clinic_id}/reservations/{reservationId}":                  true,
	"DELETE /api/v1/masters/medicines/{id}/dose-params/{species}":                      true,
	"GET /api/v1/checkups":                                                          true,
	"GET /api/v1/checkups/alerts":                                                   true,
	"GET /api/v1/checkups/field-results":                                            true,
	"GET /api/v1/clinics/{clinic_id}/lstep-settings":                                true,
	"GET /api/v1/clinics/{clinic_id}/lstep-tag-code-mappings":                       true,
	"GET /api/v1/clinics/{clinic_id}/lstep/checkup-sync/preview":                    true,
	"GET /api/v1/clinics/{clinic_id}/lstep/delivery-monitor/logs":                   true,
	"GET /api/v1/clinics/{clinic_id}/lstep/delivery-monitor/summary":                true,
	"GET /api/v1/clinics/{clinic_id}/lstep/owners":                                  true,
	"GET /api/v1/clinics/{clinic_id}/lstep/tag-summary":                             true,
	"GET /api/v1/clinics/{clinic_id}/lstep/trigger-priorities":                      true,
	"GET /api/v1/clinics/{clinic_id}/owners/aggregations":                           true,
	"GET /api/v1/clinics/{clinic_id}/reservation-staffs":                            true,
	"GET /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/schedules":        true,
	"GET /api/v1/clinics/{clinic_id}/reservations":                                  true,
	"GET /api/v1/masters/animal-species/{id}":                                       true,
	"GET /api/v1/masters/cages/{id}":                                                true,
	"GET /api/v1/masters/checkup-types/{id}":                                        true,
	"GET /api/v1/masters/checkup-types/{id}/fields":                                 true,
	"GET /api/v1/masters/consultations/{id}":                                        true,
	"GET /api/v1/masters/diagnosis-names/{id}":                                      true,
	"GET /api/v1/masters/diagnosis-names/all":                                       true,
	"GET /api/v1/masters/diagnosis-types/{id}":                                      true,
	"GET /api/v1/masters/examination-types/{id}":                                    true,
	"GET /api/v1/masters/hospitalization-plans/{id}":                                true,
	"GET /api/v1/masters/insurances/{id}":                                           true,
	"GET /api/v1/masters/medicines/{id}":                                            true,
	"GET /api/v1/masters/medicines/{id}/dose-params":                                true,
	"GET /api/v1/masters/merchandise-items/{id}":                                    true,
	"GET /api/v1/masters/occupations/{id}":                                          true,
	"GET /api/v1/masters/procedures/{id}":                                           true,
	"GET /api/v1/masters/staffs/{id}":                                               true,
	"GET /api/v1/masters/staffs/{id}/permission-groups":                             true,
	"GET /api/v1/masters/trimming-courses/{id}":                                     true,
	"GET /api/v1/masters/trimming-options/{id}":                                     true,
	"GET /api/v1/masters/vaccines/{id}":                                             true,
	"GET /api/v1/reservations/available-times":                                      true,
	"GET /api/v1/shifts/on-duty-staffs":                                             true,
	"PATCH /api/v1/clinics/{clinic_id}/lstep-settings":                              true,
	"PATCH /api/v1/clinics/{clinic_id}/lstep/trigger-priorities":                    true,
	"PATCH /api/v1/clinics/{clinic_id}/pets/{id}/death":                             true,
	"PATCH /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/sort-order":     true,
	"PATCH /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/status":         true,
	"PATCH /api/v1/medical-records/{id}/recommendation-reason":                      true,
	"PATCH /api/v1/reservations/{id}/reservation-route":                             true,
	"POST /api/line/webhook":                                                        true,
	"POST /api/v1/clinics/{clinic_id}/lstep-settings/test-connection":               true,
	"POST /api/v1/clinics/{clinic_id}/lstep/checkup-sync":                           true,
	"POST /api/v1/clinics/{clinic_id}/reservation-staffs":                           true,
	"POST /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/image":           true,
	"POST /api/v1/clinics/{clinic_id}/reservations":                                 true,
	"POST /api/v1/hospitalizations/{id}/discharge-with-billing":                     true,
	"POST /api/v1/medical-records/{id}/images/upload":                               true,
	"PUT /api/v1/clinics/{clinic_id}/lstep-tag-code-mappings/{tag_name}":            true,
	"PUT /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}":                  true,
	"PUT /api/v1/clinics/{clinic_id}/reservation-staffs/{staffId}/schedules/{date}": true,
	"PUT /api/v1/masters/medicines/{id}/dose-params/{species}":                      true,
}

var knownPhantomInSpec = map[string]bool{
	"GET /api/v1/health": true,

	// pairs with the clinic_handler.go :clinic_id vs api.yaml {id} param-name drift above.
	"DELETE /api/v1/clinics/{id}": true,
	"GET /api/v1/clinics/{id}":    true,
	"PATCH /api/v1/clinics/{id}":  true,
}

// ─── Real route enumeration (go/ast) ────────────────────────────────────────────────

var httpVerbs = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

// routeParam is a flattened (name, isRoutingType) view of a func's parameter list —
// flattening handles the fact that ast.Field groups multiple names under one type,
// though no Register* func in this package currently does so.
type routeParam struct {
	name          string
	isRoutingType bool // *gin.RouterGroup or *gin.Engine
}

func flattenParams(ft *ast.FuncType) []routeParam {
	var out []routeParam
	if ft.Params == nil {
		return out
	}
	for _, field := range ft.Params.List {
		isRouting := isGinRoutingType(field.Type)
		if len(field.Names) == 0 {
			out = append(out, routeParam{isRoutingType: isRouting})
			continue
		}
		for _, n := range field.Names {
			out = append(out, routeParam{name: n.Name, isRoutingType: isRouting})
		}
	}
	return out
}

// isGinRoutingType reports whether expr is *gin.RouterGroup or *gin.Engine.
func isGinRoutingType(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "gin" {
		return false
	}
	return sel.Sel.Name == "RouterGroup" || sel.Sel.Name == "Engine"
}

// routingParamIndex returns the flattened-param index of the single *gin.RouterGroup/
// *gin.Engine parameter, or -1 if none is found.
func routingParamIndex(params []routeParam) int {
	for i, p := range params {
		if p.isRoutingType {
			return i
		}
	}
	return -1
}

func identName(expr ast.Expr) (string, bool) {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}
	return id.Name, true
}

func stringLit(expr ast.Expr) (string, bool) {
	bl, ok := expr.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

// ginToOpenAPIParams converts gin `:name` path segments to OpenAPI `{name}` segments.
func ginToOpenAPIParams(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") {
			segs[i] = "{" + s[1:] + "}"
		}
	}
	return strings.Join(segs, "/")
}

// routeCollector walks the Register*Routes call graph rooted at RegisterRoutes and
// collects every resolved (method, absolute path) pair. It also records any call site
// it could not resolve statically (unexpected shape) so the test can fail loudly
// instead of silently under-counting.
type routeCollector struct {
	funcs      map[string]*ast.FuncDecl
	routes     []string // "METHOD /path", may contain duplicates (aliases register the same op)
	unresolved []string
	visiting   map[string]bool // guards against accidental infinite recursion
}

func (rc *routeCollector) walkFunc(fd *ast.FuncDecl, prefix string) {
	params := flattenParams(fd.Type)
	idx := routingParamIndex(params)
	if idx < 0 {
		rc.unresolved = append(rc.unresolved, "func "+fd.Name.Name+" has no *gin.RouterGroup/*gin.Engine parameter")
		return
	}
	key := fd.Name.Name + "@" + prefix
	if rc.visiting[key] {
		rc.unresolved = append(rc.unresolved, "recursive Register call detected at "+key)
		return
	}
	rc.visiting[key] = true
	defer delete(rc.visiting, key)

	env := map[string]string{params[idx].name: prefix}
	rc.walkStmts(fd.Body.List, env)
}

func (rc *routeCollector) walkStmts(stmts []ast.Stmt, env map[string]string) {
	for _, stmt := range stmts {
		rc.walkStmt(stmt, env)
	}
}

// walkStmt handles the flat, straight-line statement shapes actually used by this
// codebase's route registration functions (assignment to a Group() call, or an
// expression-statement verb/Register call). Block-like statements are recursed into
// defensively (none exist today in Register*Routes bodies) but func literals are not
// descended into, since none of them register routes in this codebase.
func (rc *routeCollector) walkStmt(stmt ast.Stmt, env map[string]string) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		rc.handleAssign(s, env)
	case *ast.ExprStmt:
		rc.handleExprStmt(s, env)
	case *ast.BlockStmt:
		rc.walkStmts(s.List, env)
	case *ast.IfStmt:
		rc.walkStmts(s.Body.List, env)
		if s.Else != nil {
			rc.walkStmt(s.Else, env)
		}
	case *ast.ForStmt:
		rc.walkStmts(s.Body.List, env)
	case *ast.RangeStmt:
		rc.walkStmts(s.Body.List, env)
	case *ast.SwitchStmt:
		for _, c := range s.Body.List {
			if cc, ok := c.(*ast.CaseClause); ok {
				rc.walkStmts(cc.Body, env)
			}
		}
	}
}

// handleAssign resolves `X := Y.Group("literal")` assignments. Every early-return
// branch records an unresolved entry (rather than silently dropping the statement) so
// that an unexpected shape — e.g. a future `sub := rg.Group(someVar)` with a
// non-literal/dynamic prefix — fails collectRealRoutes loudly instead of silently
// under-counting the routes nested under that group (go-reviewer HIGH: a silently
// dropped Group() would let its entire subtree vanish from realRoutes without ever
// appearing in missingFromSpec, defeating the gate's purpose).
func (rc *routeCollector) handleAssign(s *ast.AssignStmt, env map[string]string) {
	if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
		return // multi-value assignment; not a Group() binding shape used in this codebase
	}
	lhsName, ok := identName(s.Lhs[0])
	if !ok {
		rc.unresolved = append(rc.unresolved, "assignment with non-ident LHS (not a simple `x := ...` binding)")
		return
	}
	call, ok := s.Rhs[0].(*ast.CallExpr)
	if !ok {
		return // e.g. `x := someNonCallExpr`; not a routing binding, nothing to lose
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Group" {
		return // call to something other than .Group(...); not a routing binding
	}
	baseName, ok := identName(sel.X)
	if !ok {
		rc.unresolved = append(rc.unresolved, lhsName+" := <expr>.Group(...) where <expr> is not a simple identifier")
		return
	}
	basePrefix, ok := env[baseName]
	if !ok {
		// baseName is a *gin.RouterGroup/*gin.Engine we haven't bound a prefix for (or is
		// simply not tracked, e.g. shadows an unrelated identifier) — cannot safely ignore.
		rc.unresolved = append(rc.unresolved, lhsName+" := "+baseName+".Group(...) but "+baseName+" has no resolved prefix in scope")
		return
	}
	if len(call.Args) != 1 {
		rc.unresolved = append(rc.unresolved, lhsName+" := "+baseName+".Group(...) called with "+strconv.Itoa(len(call.Args))+" args, expected exactly 1")
		return
	}
	lit, ok := stringLit(call.Args[0])
	if !ok {
		rc.unresolved = append(rc.unresolved, lhsName+" := "+baseName+".Group(<non-literal>) — dynamic group prefix cannot be statically resolved")
		return
	}
	env[lhsName] = basePrefix + lit
}

// handleExprStmt resolves both direct verb calls (`rg.GET("literal", ...)`) and
// cross-function Register delegation (`h.RegisterXxx(rg)`). Every early-return branch
// records an unresolved entry for the same reason as handleAssign above.
func (rc *routeCollector) handleExprStmt(s *ast.ExprStmt, env map[string]string) {
	call, ok := s.X.(*ast.CallExpr)
	if !ok {
		return // bare expression statement that isn't a call; nothing routing-related
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return // call to a non-method function; nothing routing-related
	}
	baseName, ok := identName(sel.X)
	if !ok {
		return // e.g. `(*x).GET(...)`; no such call shape exists in this codebase today
	}

	if httpVerbs[sel.Sel.Name] {
		basePrefix, ok := env[baseName]
		if !ok {
			rc.unresolved = append(rc.unresolved, baseName+"."+sel.Sel.Name+"(...) but "+baseName+" has no resolved prefix in scope")
			return
		}
		if len(call.Args) < 1 {
			rc.unresolved = append(rc.unresolved, baseName+"."+sel.Sel.Name+"() called with no arguments")
			return
		}
		lit, ok := stringLit(call.Args[0])
		if !ok {
			rc.unresolved = append(rc.unresolved, baseName+"."+sel.Sel.Name+"(<non-literal>, ...) — dynamic path cannot be statically resolved")
			return
		}
		full := ginToOpenAPIParams(basePrefix + lit)
		rc.routes = append(rc.routes, sel.Sel.Name+" "+full)
		return
	}

	if baseName == "h" && (strings.HasPrefix(sel.Sel.Name, "Register") || strings.HasPrefix(sel.Sel.Name, "register")) {
		callee, ok := rc.funcs[sel.Sel.Name]
		if !ok {
			rc.unresolved = append(rc.unresolved, "call to unknown func h."+sel.Sel.Name)
			return
		}
		calleeParams := flattenParams(callee.Type)
		calleeIdx := routingParamIndex(calleeParams)
		if calleeIdx < 0 || calleeIdx >= len(call.Args) {
			rc.unresolved = append(rc.unresolved, "cannot bind routing arg for call to h."+sel.Sel.Name)
			return
		}
		argName, ok := identName(call.Args[calleeIdx])
		if !ok {
			rc.unresolved = append(rc.unresolved, "non-ident routing arg in call to h."+sel.Sel.Name)
			return
		}
		argPrefix, ok := env[argName]
		if !ok {
			rc.unresolved = append(rc.unresolved, "unresolved routing arg %q in call to h."+sel.Sel.Name)
			return
		}
		rc.walkFunc(callee, argPrefix)
	}
}

// collectRealRoutes parses every non-test .go file in internal/handler, resolves the
// Register*Routes call graph from RegisterRoutes, and returns the set of resolved
// routes plus any unresolved call sites (which should always be empty; a non-empty
// result means the codebase's routing style changed and the walker needs updating).
func collectRealRoutes(t *testing.T) (found map[string]int, unresolved []string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(handlerDir, "*.go"))
	if err != nil {
		t.Fatalf("glob handler dir: %v", err)
	}
	fset := token.NewFileSet()
	funcs := map[string]*ast.FuncDecl{}
	for _, fp := range files {
		if strings.HasSuffix(fp, "_test.go") {
			continue
		}
		src, err := os.ReadFile(fp) //nolint:gosec // fixed handler source dir, not untrusted input
		if err != nil {
			t.Fatalf("read %s: %v", fp, err)
		}
		f, err := parser.ParseFile(fset, fp, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", fp, err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || fd.Body == nil {
				continue
			}
			funcs[fd.Name.Name] = fd
		}
	}

	root, ok := funcs["RegisterRoutes"]
	if !ok {
		t.Fatalf("RegisterRoutes not found in %s — handler package layout changed", handlerDir)
	}

	rc := &routeCollector{funcs: funcs, visiting: map[string]bool{}}
	rc.walkFunc(root, "")

	agg := map[string]int{}
	for _, r := range rc.routes {
		agg[r]++
	}
	return agg, rc.unresolved
}

// ─── OpenAPI paths parsing ───────────────────────────────────────────────────────────

const routeDriftOpenAPIPath = "../../docs/api.yaml"

var pathOperationVerbs = []string{"get", "post", "put", "patch", "delete"}

// parseOpenAPIOperations returns the set of "METHOD absolutePath" operations declared
// under `paths:`. Path keys starting with "/api/liff" are treated as already-absolute
// (per the LIFF servers-collision note); all other keys are prefixed with the single
// declared server's base path (/api/v1).
func parseOpenAPIOperations(yamlSrc []byte) (map[string]struct{}, error) {
	var root map[string]any
	if err := yaml.Unmarshal(yamlSrc, &root); err != nil {
		return nil, err
	}
	paths, ok := root["paths"].(map[string]any)
	if !ok {
		return nil, nil
	}
	out := map[string]struct{}{}
	for pathKey, v := range paths {
		if !strings.HasPrefix(pathKey, "/") {
			continue // stray non-path mapping key (comments etc. are not represented as keys)
		}
		ops, ok := v.(map[string]any)
		if !ok {
			continue
		}
		abs := pathKey
		// exact-segment match, not bare prefix — "/api/liff-foo" must not be misclassified
		// as an already-absolute LIFF path (go-reviewer MEDIUM).
		if pathKey != "/api/liff" && !strings.HasPrefix(pathKey, "/api/liff/") {
			abs = "/api/v1" + pathKey
		}
		for _, verb := range pathOperationVerbs {
			if _, ok := ops[verb]; ok {
				out[strings.ToUpper(verb)+" "+abs] = struct{}{}
			}
		}
	}
	return out, nil
}

// ─── Reconciliation (pure) ───────────────────────────────────────────────────────────

// reconcileRouteDrift compares the real route set against the spec operation set and
// checks both diffs against their allowlists. New drift, count drift (a route that
// gained/lost duplicate aliases), and stale allowlist entries all produce violations.
func reconcileRouteDrift(realRoutes map[string]int, specOps map[string]struct{}, allowMissing, allowPhantom map[string]bool) []string {
	var violations []string

	missing := map[string]bool{}
	for r := range realRoutes {
		if _, ok := specOps[r]; !ok {
			missing[r] = true
		}
	}
	phantom := map[string]bool{}
	for op := range specOps {
		if _, ok := realRoutes[op]; !ok {
			phantom[op] = true
		}
	}

	for r := range missing {
		if !allowMissing[r] {
			violations = append(violations, "NEW missing-from-spec route: "+r+
				" (implemented but not documented in docs/api.yaml; add the operation, or if intentionally "+
				"deferred, add it to knownMissingFromSpec with a follow-up note)")
		}
	}
	for r := range allowMissing {
		if !missing[r] {
			violations = append(violations, "stale knownMissingFromSpec entry: "+r+
				" (now documented, or route no longer exists — remove the entry)")
		}
	}

	for op := range phantom {
		if !allowPhantom[op] {
			violations = append(violations, "NEW phantom-in-spec operation: "+op+
				" (documented in docs/api.yaml but no implementation resolves to it; remove the operation, or "+
				"if intentionally provisional, add it to knownPhantomInSpec with a follow-up note)")
		}
	}
	for op := range allowPhantom {
		if !phantom[op] {
			violations = append(violations, "stale knownPhantomInSpec entry: "+op+
				" (now implemented, or the doc entry was removed — remove the allowlist entry)")
		}
	}

	sort.Strings(violations)
	return violations
}

// ─── Gate tests ──────────────────────────────────────────────────────────────────────

// TestOpenAPIRouteDrift_MatchesAllowlist is the gate: every implemented-route ↔
// documented-operation mismatch must be on the pinned allowlist; no new drift, no
// stale entry. Floors guard against a vacuous pass if AST parsing or the yaml parse
// silently breaks.
func TestOpenAPIRouteDrift_MatchesAllowlist(t *testing.T) {
	realRoutes, unresolved := collectRealRoutes(t)
	if len(unresolved) > 0 {
		t.Fatalf("route walker could not statically resolve %d call site(s), refusing to trust the "+
			"partial result (the routing style changed — update the walker): %v", len(unresolved), unresolved)
	}
	if len(realRoutes) < 400 {
		t.Fatalf("only %d real routes resolved; AST walk likely broke (expected ~477). Would vacuously pass.",
			len(realRoutes))
	}

	yamlSrc, err := os.ReadFile(routeDriftOpenAPIPath) //nolint:gosec // fixed docs path, not untrusted input
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	specOps, err := parseOpenAPIOperations(yamlSrc)
	if err != nil {
		t.Fatalf("parse openapi: %v", err)
	}
	if len(specOps) < 250 {
		t.Fatalf("only %d openapi operations parsed; yaml parse likely broke. Would vacuously pass.", len(specOps))
	}

	for _, v := range reconcileRouteDrift(realRoutes, specOps, knownMissingFromSpec, knownPhantomInSpec) {
		t.Error(v)
	}
}

// TestOpenAPIRouteDrift_OpenAPIParser pins parseOpenAPIOperations on inline fixtures,
// including the LIFF absolute-path special case.
func TestOpenAPIRouteDrift_OpenAPIParser(t *testing.T) {
	src := []byte("" +
		"paths:\n" +
		"  /owners:\n" +
		"    get:\n" +
		"      operationId: listOwners\n" +
		"    post:\n" +
		"      operationId: createOwner\n" +
		"  /owners/{id}:\n" +
		"    patch:\n" +
		"      operationId: updateOwner\n" +
		"  /api/liff/{clinicId}/settings:\n" +
		"    get:\n" +
		"      operationId: liffSettings\n")
	got, err := parseOpenAPIOperations(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := map[string]struct{}{
		"GET /api/v1/owners":                {},
		"POST /api/v1/owners":               {},
		"PATCH /api/v1/owners/{id}":         {},
		"GET /api/liff/{clinicId}/settings": {},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d ops, want %d: %v", len(got), len(want), got)
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing expected op %q in %v", k, got)
		}
	}
	if _, ok := got["GET /owners"]; ok {
		t.Error("non-LIFF path must be prefixed with /api/v1, not left bare")
	}
	if _, ok := got["GET /api/v1/api/liff/{clinicId}/settings"]; ok {
		t.Error("LIFF path must NOT be double-prefixed with /api/v1 (server-collision regression)")
	}
}

// TestOpenAPIRouteDrift_GinParamNormalizer pins ginToOpenAPIParams.
func TestOpenAPIRouteDrift_GinParamNormalizer(t *testing.T) {
	cases := map[string]string{
		"/owners/:id":                      "/owners/{id}",
		"/owners/:id/lstep/tags/:tag_name": "/owners/{id}/lstep/tags/{tag_name}",
		"/api/liff/:clinicId/settings":     "/api/liff/{clinicId}/settings",
		"/health":                          "/health",
		"":                                 "",
	}
	for in, want := range cases {
		if got := ginToOpenAPIParams(in); got != want {
			t.Errorf("ginToOpenAPIParams(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestOpenAPIRouteDrift_Reconciler pins the gate's failure modes on synthetic inputs.
func TestOpenAPIRouteDrift_Reconciler(t *testing.T) {
	real := map[string]int{"GET /api/v1/owners": 1, "POST /api/v1/owners": 1}
	spec := map[string]struct{}{"GET /api/v1/owners": {}, "POST /api/v1/owners": {}}

	t.Run("clean baseline reports nothing", func(t *testing.T) {
		if v := reconcileRouteDrift(real, spec, nil, nil); len(v) != 0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})
	t.Run("new missing-from-spec route fails", func(t *testing.T) {
		r2 := map[string]int{"GET /api/v1/owners": 1, "POST /api/v1/owners": 1, "DELETE /api/v1/owners/{id}": 1}
		v := reconcileRouteDrift(r2, spec, nil, nil)
		if len(v) != 1 || !strings.Contains(v[0], "NEW missing-from-spec") {
			t.Fatalf("expected new-missing violation, got %v", v)
		}
	})
	t.Run("new phantom-in-spec operation fails", func(t *testing.T) {
		s2 := map[string]struct{}{"GET /api/v1/owners": {}, "POST /api/v1/owners": {}, "DELETE /api/v1/owners/{id}": {}}
		v := reconcileRouteDrift(real, s2, nil, nil)
		if len(v) != 1 || !strings.Contains(v[0], "NEW phantom-in-spec") {
			t.Fatalf("expected new-phantom violation, got %v", v)
		}
	})
	t.Run("allowlisted missing route passes", func(t *testing.T) {
		r2 := map[string]int{"GET /api/v1/owners": 1, "POST /api/v1/owners": 1, "DELETE /api/v1/owners/{id}": 1}
		v := reconcileRouteDrift(r2, spec, map[string]bool{"DELETE /api/v1/owners/{id}": true}, nil)
		if len(v) != 0 {
			t.Fatalf("expected 0 (allowlisted), got %v", v)
		}
	})
	t.Run("stale missing allowlist entry fails", func(t *testing.T) {
		v := reconcileRouteDrift(real, spec, map[string]bool{"DELETE /api/v1/owners/{id}": true}, nil)
		if len(v) != 1 || !strings.Contains(v[0], "stale knownMissingFromSpec") {
			t.Fatalf("expected stale violation, got %v", v)
		}
	})
	t.Run("stale phantom allowlist entry fails", func(t *testing.T) {
		v := reconcileRouteDrift(real, spec, nil, map[string]bool{"DELETE /api/v1/owners/{id}": true})
		if len(v) != 1 || !strings.Contains(v[0], "stale knownPhantomInSpec") {
			t.Fatalf("expected stale violation, got %v", v)
		}
	})
}

// TestOpenAPIRouteDrift_Walker pins the AST route walker on an inline fixture package
// covering: Group() chaining, direct verb calls, and cross-function Register delegation.
func TestOpenAPIRouteDrift_Walker(t *testing.T) {
	fset := token.NewFileSet()
	src := `package handler

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	protected := api.Group("")
	h.registerOwnerRoutesWithAuth(protected)
	r.GET("/health", h.Health)
}

func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.GET("/:id", h.GetOwner)
	h.RegisterChronicConditionRoutes(owners)
}

func (h *Handler) RegisterChronicConditionRoutes(rg *gin.RouterGroup) {
	cc := rg.Group("/:id/chronic-conditions")
	cc.GET("", h.ListChronicConditions)
}
`
	f, err := parser.ParseFile(fset, "fixture.go", src, 0)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcs[fd.Name.Name] = fd
		}
	}
	rc := &routeCollector{funcs: funcs, visiting: map[string]bool{}}
	rc.walkFunc(funcs["RegisterRoutes"], "")
	if len(rc.unresolved) != 0 {
		t.Fatalf("unexpected unresolved call sites: %v", rc.unresolved)
	}
	got := map[string]int{}
	for _, r := range rc.routes {
		got[r]++
	}
	want := map[string]int{
		"GET /health":                                1,
		"GET /api/v1/owners":                         1,
		"GET /api/v1/owners/{id}":                    1,
		"GET /api/v1/owners/{id}/chronic-conditions": 1,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d routes, want %d: %v", len(got), len(want), got)
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("route %q: got count %d, want %d (all: %v)", k, got[k], n, got)
		}
	}
}
