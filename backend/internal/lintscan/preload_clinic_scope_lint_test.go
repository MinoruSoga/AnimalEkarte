package lintscan

// Clinic-scope Preload rule (mechanical enforcement) — clinic-scoped master Preloads must
// carry a clinic_id predicate.
//
// Background: a systemic cross-tenant read IDOR (b3638d5e) was fixed by hand: when a
// clinic-scoped "master/区分" is Preloaded by an FK value without a clinic_id predicate,
// a polluted FK (write-side validation gaps #124/#125, or pre-remediation data) makes the
// query return ANOTHER clinic's master name/price. The clinic-scope Preload rule (repository/CLAUDE.md) requires
// every such Preload to scope clinic_id. This test enforces that rule mechanically so the
// gap cannot regress — the read-side analogue of the runtime *_clinic_isolation_test.go guards.
//
// ─── Discovery mechanism (BE9-1) ─────────────────────────────────────────────────────
//
// Source discovery is delegated to the shared internal/lintscan package
// (lintscan.WalkInternalTreeT), NOT to a package-local go:embed. Before BE9-1 this file declared
// `//go:embed *.go */*.go` and a package-level `repoSourceFS embed.FS`, which only reached this
// package's own root files plus exactly ONE level of domain subpackage (e.g.
// paymentmethod/repository.go) — it could not see 2+-level nesting, and it could not see any
// other internal/ package (service, model, handler, ...) at all. lintscan.WalkInternalTreeT walks
// the WHOLE module's internal/ tree to arbitrary depth, so preload / audit-tx / dbortx
// (this file, audit_tx_inventory_lint_test.go, dbortx_inventory_lint_test.go) now all scan every
// production .go file under internal/, not only internal/repository — a violation fixture placed
// anywhere under internal/ (any depth, any top-level package) is caught by all three gates.
//
// Key-format backward compatibility: lintscan keys are relative to internal/ (e.g.
// "repository/foo.go", "repository/shiftentry/repository.go", "service/vaccination_service.go").
// legacyLintKey (below) strips the "repository/" prefix before the key reaches
// analyzeFilePreloads, so a repository root file is still keyed "foo.go" and a repository
// subpackage file is still keyed "shiftentry/repository.go" — byte-for-byte the same shape the
// old repoSourceFS walk produced — which keeps every existing preloadClinicScopeSiteExceptions
// entry matching. Any OTHER top-level internal/ package keeps its full lintscan key unchanged
// (e.g. "service/vaccination_service.go"); those have no legacy site-exception entries, so a new,
// longer-but-stable key shape for anything newly discovered there is expected and fine.
//
// ─── Blind spots (static AST analysis; not a replacement for runtime guards) ────────
//
// This lint (like the other two BE9-1 lints in this package) parses Go source with go/ast and
// pattern-matches literal call shapes. It structurally cannot see:
//   - a clinic_id predicate hidden inside a raw SQL string literal (e.g. db.Exec("SELECT ...
//     WHERE vaccine_id = ?")) — raw SQL is never parsed as a Preload call, so a violation
//     reachable only through raw SQL is invisible to this gate;
//   - a bare scalar FK column that resolves to a clinic-scoped master but whose association
//     name is not registered in clinicScopedMasterAssoc — an unregistered name silently falls
//     outside this gate's scope until someone adds it;
//   - a violation reachable only through a background job, cron, worker, or any other code path
//     that is not represented as a literal .Preload(...) AST call shape inside a plain .go file
//     this scanner visits.
//
// The cross-tenant runtime tests (the *_clinic_isolation_test.go-style guards) remain required
// for exactly these blind spots. This static lint is a complement layered on top of those runtime
// guards, not a substitute for them.
//
// Technique origin: reused from model/audit_taxonomy_exhaustiveness_test.go — parsed with
// go/parser + go/ast, pinned with self-verification + good/bad fixtures so the check is never
// vacuously green.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// clinicScopedMasterAssoc maps a GORM association name (the LAST dotted segment of a
// Preload path) to the clinic-scoped master model it resolves to. Every Preload of one
// of these MUST carry a clinic_id predicate (the clinic-scope Preload rule).
//
// Membership is curated from the clinic-scope Preload rule's master list and confirmed against each model's
// clinic_id column — deliberately NOT "every table that has a clinic_id". Account and
// LineCustomer have clinic_id but are not configurable masters and are intentionally
// excluded. The generic alias names (Parent/Children/Group/Names/Course/Options) are
// listed with their resolved model so a future reuse of one of these names for a
// non-master association is reviewed (it would fail closed) rather than silently allowed.
var clinicScopedMasterAssoc = map[string]string{
	"Vaccine":         "Vaccine",
	"Medicine":        "Medicine",
	"Procedure":       "Procedure",
	"Consultation":    "Consultation",
	"ReservationType": "ReservationType",
	"Parent":          "ReservationType",      // reservation_type.parent_id (self-ref)
	"Children":        "ReservationType",      // reservation_type children (parent_id reverse)
	"Group":           "ReservationTypeGroup", // reservation_type.group_id
	"Course":          "TrimmingCourse",       // appointment_trimming_detail.course_id
	"Options":         "TrimmingOption",       // appointment_trimming_detail options
	"Cage":            "Cage",
	"Insurance":       "Insurance",
	"ExaminationType": "ExaminationType",
	// ExamTypeField is preloaded as association name "Items" on ExaminationType (see
	// clinicScopedMasterAssocByRootModel). Direct Preload("ExamTypeField") is also gated.
	"ExamTypeField": "ExamTypeField",
	"CheckupType":   "CheckupType",
	// #211: 健診パッケージのフィールド定義マスタ（checkup_type_fields, clinic_id 列あり）。
	// checkup_field_results.CheckupTypeField として clinic_id 述語付きで Preload される。
	"CheckupTypeField": "CheckupTypeField",
	"DiagnosisType":    "DiagnosisType",
	"Diagnosis2Type":   "DiagnosisType",
	"DiagnosisName":    "DiagnosisName",
	"Diagnosis2Name":   "DiagnosisName",
	"Names":            "DiagnosisName", // diagnosis_type.Names []DiagnosisName
	"Occupation":       "Occupation",
	// BE-refactor.md R3-2 (D9・pre-registration follow-up): 以下3件は write 側 clinicScopedMasterFKField
	// (master_fk_write_inventory_lint_test.go) には既に登録済みで、model 層にも
	// gorm:"foreignKey:...ID" association フィールドが既に存在するが、本ファイル作成時点では
	// repository 層でまだ一度も Preload されていない「未点火の地雷」だった。read 側に未登録のまま
	// 誰かが初めて Preload した瞬間、本ゲートの `!isMaster` 分岐を素通りして無述語のまま
	// 別クリニックのマスタ名が混入しうる。Preload が実際に書かれる前に事前登録することで、
	// 「新規 master 登録漏れ」を発生源で塞ぐ（read/write 両ゲートの双方向整合）。
	"ChiefComplaintType":  "ChiefComplaintType",  // inquiry.ChiefComplaintType (inquiry.go)
	"HospitalizationPlan": "HospitalizationPlan", // care_plan_item.HospitalizationPlan (hospitalization.go)
	"Inventory":           "InventoryItem",       // treatment.Inventory / medicine.Inventory (association名はInventoryItemでなくInventory)
	// 残り3件（MerchandiseItem/TrimmingCourseType/PaymentMethodMaster）は write 側
	// clinicScopedMasterFKField には登録済みだが、model 層に gorm:"foreignKey:...ID" association
	// フィールドがまだ存在しない（生の *ID フィールドのみ）。Preload 自体が構文的に不可能なため
	// 本マップへの事前登録はできない。TODO follow-up: これらに association フィールドを追加する際は
	// 同一コミットで本マップにも登録すること（read/write 両ゲートの双方向整合を維持する）。
}

// clinicScopedIntermediateAssoc lists sensitive clinic-owned associations that may appear before
// the final segment of a nested Preload path. GORM applies a nested Preload predicate only to the
// final association, so these prefixes must also be explicitly preloaded with clinic scope in the
// same query chain (for example TrimmingDetail before TrimmingDetail.Course).
var clinicScopedIntermediateAssoc = map[string]string{
	"TrimmingDetail": "AppointmentTrimmingDetail",
}

// clinicScopedMasterAssocByRootModel resolves association names that are reused across models
// (e.g. "Items") using the query root model (ExaminationType.Items → ExamTypeField, while
// Examination.Items / Billing.Items stay non-master). Keyed as last-segment → root model → master.
// This is the BUG-437 context-aware gate: end-name-only matching cannot distinguish same-name
// associations on different parents.
var clinicScopedMasterAssocByRootModel = map[string]map[string]string{
	"Items": {
		"ExaminationType": "ExamTypeField",
	},
}

// staffExemptAssoc: the clinic-scope Preload rule's Staff exception. Staff belongs to multiple clinics via
// staff_clinic_assignments (staffs.clinic_id is the primary only). Scoping a historical
// Doctor/Staff preload by staffs.clinic_id would wrongly hide shared/reassigned staff from
// past records; the leak is a staff NAME only, low severity, and unreachable in normal flow
// via write isolation (72e8887c). BUG-420 exception: vaccination uses a stricter runtime
// predicate (active doctor + current clinic assignment) and suppresses DoctorID when that
// predicate fails; TestVaccinationRepository_RelationPreloadsAreClinicScoped owns that contract.
// These names all resolve to the Staff model.
var staffExemptAssoc = map[string]struct{}{
	"Doctor":          {},
	"Staff":           {},
	"EnteredByStaff":  {},
	"PaidByStaff":     {},
	"ClosedByStaff":   {},
	"CreatedByStaff":  {},
	"RefundedByStaff": {},
	"CreatedStaff":    {},
}

// globalExemptAssoc: global masters with NO clinic_id column — a clinic_id predicate is
// impossible/meaningless.
var globalExemptAssoc = map[string]struct{}{
	"AnimalSpecies": {}, // global species master
	"ManualArticle": {}, // global help master
	"Account":       {}, // 1:1 auth record, no clinic_id, not a master
}

// preloadSiteException records a specific (file, assoc, predicate) site where a clinic-scoped
// master is intentionally Preloaded WITHOUT a clinic_id predicate because clinicID is not in
// scope and threading it is cross-layer-invasive. Each is a documented low-severity residual
// with a TODO follow-up — NOT a blanket "writes are validated" waiver. The key is the exact
// (base filename, assoc, raw predicate) so any normally-scoped preload added later is still
// enforced; only this precise unscoped pattern is tolerated.
//
//   - predicate holds the raw (unquoted) string predicate for a string-literal predicate, or
//     the sentinel "<closure>" / "<non-literal>" for closure / non-literal predicates.
//   - occurrences is the EXACT number of call sites this entry is expected to match. A mismatch
//     fails TestPreloadClinicScope_SiteExceptionsMatchExpectedOccurrences: a NEW unscoped
//     preload of the same pattern (count up) is forced into review instead of silently inheriting
//     the waiver, and a fixed/removed site (count down) forces deletion of the now-stale entry.
type preloadSiteException struct {
	file        string
	assoc       string
	predicate   string
	occurrences int
	reason      string
}

var preloadClinicScopeSiteExceptions []preloadSiteException

type preloadFinding struct {
	file   string
	line   int
	assoc  string
	detail string
}

type preloadStats struct {
	filesParsed    int
	masterPreloads int
	staffExempt    int
	globalExempt   int
	siteExempt     int
	// siteExemptByKey counts matched site exceptions per "file|assoc|predicate" key, so each
	// exception's occurrence count can be pinned (detects new unscoped occurrences + stale entries).
	siteExemptByKey map[string]int
}

func siteExceptionKey(file, assoc, predicate string) string {
	return file + "|" + assoc + "|" + predicate
}

// analyzeFilePreloads parses one Go source file and reports clinic-scoped master Preloads that
// lack a clinic_id predicate. It is a pure function over (filename, src) so the real discovered
// source and the inline fixtures exercise exactly the same logic.
func analyzeFilePreloads(filename string, src []byte) ([]preloadFinding, preloadStats, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, preloadStats{}, err
	}

	var findings []preloadFinding
	stats := preloadStats{filesParsed: 1}
	// Root files key by basename; 1-level subpackage files keep the relative path so e.g.
	// paymentmethod/repository.go and vaccine/repository.go do not share a site-exception
	// identity (BE-refactor.md §5-1-3; mirrors dbortx_inventory_lint_test.go's keyFile logic).
	base := baseFileName(filename)
	if strings.Contains(filename, "/") {
		base = filename
	}

	// Preload Pos → root model inferred from the Find/First chain that contains that Preload.
	// Used for ambiguous last-segment names (Items) where clinicScopedMasterAssoc alone is wrong.
	preloadRootModel := map[token.Pos]string{}
	localTypes := collectLocalVarTypes(f)
	ast.Inspect(f, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok || !isGormTerminalQuery(sel.Sel.Name) || len(ce.Args) == 0 {
			return true
		}
		root := rootModelTypeName(ce.Args[0], localTypes)
		if root == "" {
			return true
		}
		for _, preloadCE := range preloadCallsInReceiverChain(ce) {
			preloadRootModel[preloadCE.Pos()] = root
		}
		return true
	})

	ast.Inspect(f, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Preload" || len(ce.Args) == 0 {
			return true
		}
		assoc, ok := stringLitValue(ce.Args[0])
		if !ok {
			return true // non-literal association name; not used in real repository code
		}
		for _, prefix := range clinicScopedIntermediatePrefixes(assoc) {
			if preloadReceiverChainHasScopedAssociation(ce, prefix) {
				continue
			}
			findings = append(findings, preloadFinding{
				file:   base,
				line:   fset.Position(ce.Pos()).Line,
				assoc:  prefix,
				detail: "nested Preload requires a preceding clinic-scoped Preload for its intermediate association",
			})
		}
		key := lastAssocSegment(assoc)

		switch {
		case isInSet(staffExemptAssoc, key):
			stats.staffExempt++
			return true
		case isInSet(globalExemptAssoc, key):
			stats.globalExempt++
			return true
		}

		// For ambiguous last segments (Items), prefer the association parent on a nested
		// path (ExaminationType.Items → ExaminationType) over the query-root model
		// (Examination). Query root is used only for bare Preload("Items") chains.
		// Using query root alone would miss nested masters under non-parent Find destinations
		// (e.g. Preload("ExaminationType.Items").Find(&exams) with root Examination).
		parentForAmbiguous := ""
		if i := strings.LastIndex(assoc, "."); i > 0 {
			parentForAmbiguous = lastAssocSegment(assoc[:i])
		}
		if parentForAmbiguous == "" {
			parentForAmbiguous = preloadRootModel[ce.Pos()]
		}
		isMaster := false
		if _, ok := clinicScopedMasterAssoc[key]; ok {
			// Unambiguous end-name masters (Vaccine, Medicine, ...).
			// Ambiguous end-names registered only under ByRootModel must not match here.
			// Invariant: ByRootModel keys should not also be full end-name masters in
			// clinicScopedMasterAssoc (would accidentally weaken end-name gating).
			if _, ambiguous := clinicScopedMasterAssocByRootModel[key]; !ambiguous {
				isMaster = true
			}
		}
		if parents, ok := clinicScopedMasterAssocByRootModel[key]; ok {
			if parentForAmbiguous != "" {
				if _, ok := parents[parentForAmbiguous]; ok {
					isMaster = true
				}
			}
		}
		if !isMaster {
			return true // OTHER (Owner/Pet/non-master Items/Payments/...) — outside master scope
		}

		stats.masterPreloads++
		scoped, predStr := preloadHasClinicScope(ce)
		if scoped {
			return true
		}
		if isSiteExcepted(base, key, predStr) {
			stats.siteExempt++
			if stats.siteExemptByKey == nil {
				stats.siteExemptByKey = make(map[string]int)
			}
			stats.siteExemptByKey[siteExceptionKey(base, key, predStr)]++
			return true
		}
		findings = append(findings, preloadFinding{
			file:   base,
			line:   fset.Position(ce.Pos()).Line,
			assoc:  assoc,
			detail: preloadFailDetail(ce, predStr),
		})
		return true
	})

	return findings, stats, nil
}

func isGormTerminalQuery(name string) bool {
	switch name {
	case "Find", "First", "Take", "Last", "FindInBatches", "Scan":
		return true
	default:
		return false
	}
}

// preloadCallsInReceiverChain walks left from a terminal query call and returns every Preload
// CallExpr in that receiver chain (e.g. db.Preload(...).Order(...).Find(...)).
func preloadCallsInReceiverChain(ce *ast.CallExpr) []*ast.CallExpr {
	var out []*ast.CallExpr
	cur := ce
	for cur != nil {
		sel, ok := cur.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		if sel.Sel.Name == "Preload" {
			out = append(out, cur)
		}
		next, ok := sel.X.(*ast.CallExpr)
		if !ok {
			break
		}
		cur = next
	}
	return out
}

// collectLocalVarTypes maps local identifiers to their declared type expression string.
func collectLocalVarTypes(f *ast.File) map[string]string {
	out := make(map[string]string)
	ast.Inspect(f, func(n ast.Node) bool {
		switch d := n.(type) {
		case *ast.ValueSpec:
			if d.Type == nil {
				return true
			}
			typeStr := exprTypeString(d.Type)
			for _, name := range d.Names {
				if name != nil && name.Name != "_" {
					out[name.Name] = typeStr
				}
			}
		case *ast.AssignStmt:
			if d.Tok != token.DEFINE {
				return true
			}
			for i, lhs := range d.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || id.Name == "_" || i >= len(d.Rhs) {
					continue
				}
				if typeStr := exprTypeString(d.Rhs[i]); typeStr != "" {
					out[id.Name] = typeStr
				}
				// make([]model.T, 0) style
				if call, ok := d.Rhs[i].(*ast.CallExpr); ok {
					if fun, ok := call.Fun.(*ast.Ident); ok && fun.Name == "make" && len(call.Args) > 0 {
						out[id.Name] = exprTypeString(call.Args[0])
					}
				}
			}
		}
		return true
	})
	return out
}

func exprTypeString(e ast.Expr) string {
	if e == nil {
		return ""
	}
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + exprTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + exprTypeString(t.Elt)
	case *ast.CompositeLit:
		return exprTypeString(t.Type)
	case *ast.CallExpr:
		// make([]T, n) handled by caller; New-style not needed
		return ""
	}
	return ""
}

// rootModelTypeName extracts the GORM root model type name from a Find/First destination arg.
func rootModelTypeName(arg ast.Expr, localTypes map[string]string) string {
	switch a := arg.(type) {
	case *ast.UnaryExpr:
		if a.Op == token.AND {
			return rootModelTypeName(a.X, localTypes)
		}
	case *ast.Ident:
		if t, ok := localTypes[a.Name]; ok {
			return baseModelTypeName(t)
		}
	case *ast.CompositeLit:
		return baseModelTypeName(exprTypeString(a.Type))
	}
	return ""
}

func baseModelTypeName(typeStr string) string {
	typeStr = strings.TrimSpace(typeStr)
	for strings.HasPrefix(typeStr, "*") || strings.HasPrefix(typeStr, "[]") {
		typeStr = strings.TrimPrefix(typeStr, "*")
		typeStr = strings.TrimPrefix(typeStr, "[]")
	}
	if i := strings.LastIndex(typeStr, "."); i >= 0 {
		return typeStr[i+1:]
	}
	return typeStr
}

// preloadHasClinicScope reports whether a Preload call carries a clinic_id scope and returns
// the raw string predicate (for site-exception matching / diagnostics). For a closure-form
// predicate it inspects the closure body for a clinic_id literal or a clinicScope(...) call.
func preloadHasClinicScope(ce *ast.CallExpr) (hasScope bool, predicate string) {
	if len(ce.Args) < 2 {
		return false, "" // master Preload with no predicate at all
	}
	switch a := ce.Args[1].(type) {
	case *ast.BasicLit:
		if a.Kind != token.STRING {
			return false, ""
		}
		s, err := strconv.Unquote(a.Value)
		if err != nil {
			return false, ""
		}
		return strings.Contains(s, "clinic_id"), s
	case *ast.FuncLit:
		return funcLitHasClinicScope(a), "<closure>"
	default:
		return false, "<non-literal>"
	}
}

// funcLitHasClinicScope reports whether a closure-form Preload predicate scopes clinic_id,
// either via a string literal containing "clinic_id" (e.g. db.Where("clinic_id = ?", id)) or
// via a clinicScope(...) helper call (e.g. db.Scopes(clinicScope(id))) which does not contain
// the literal substring.
//
// KISS tradeoff (matches the string-predicate path and the clinic-scope Preload rule's documented stance): this is a
// substring/identifier heuristic, not data-flow analysis. A contrived closure with an incidental
// "clinic_id" string literal, or one that merely names the clinicScope identifier without applying
// it, would pass. Verifying that the clinic_id actually scopes the query would require type/SSA
// analysis (rejected as over-clever); such a contrivance is visible in code review and the closure
// form is used at exactly one (now-scoped) site today.
func funcLitHasClinicScope(fl *ast.FuncLit) bool {
	found := false
	ast.Inspect(fl, func(n ast.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				if s, err := strconv.Unquote(x.Value); err == nil && strings.Contains(s, "clinic_id") {
					found = true
				}
			}
		case *ast.Ident:
			if x.Name == "clinicScope" {
				found = true
			}
		}
		return !found
	})
	return found
}

func preloadFailDetail(ce *ast.CallExpr, predStr string) string {
	switch {
	case len(ce.Args) < 2:
		return "preloaded with NO predicate (clinic_id required)"
	case predStr == "<closure>":
		return "closure predicate without clinic_id / clinicScope"
	case predStr == "<non-literal>":
		return "non-literal predicate; cannot verify clinic_id (fail-closed)"
	default:
		return "predicate missing clinic_id: " + strconv.Quote(predStr)
	}
}

func isSiteExcepted(file, assoc, predicate string) bool {
	for _, ex := range preloadClinicScopeSiteExceptions {
		if ex.file == file && ex.assoc == assoc && ex.predicate == predicate {
			return true
		}
	}
	return false
}

func stringLitValue(e ast.Expr) (string, bool) {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

func clinicScopedIntermediatePrefixes(assoc string) []string {
	parts := strings.Split(assoc, ".")
	prefixes := make([]string, 0, len(parts)-1)
	for i := 0; i < len(parts)-1; i++ {
		if _, ok := clinicScopedIntermediateAssoc[parts[i]]; !ok {
			continue
		}
		prefixes = append(prefixes, strings.Join(parts[:i+1], "."))
	}
	return prefixes
}

func preloadReceiverChainHasScopedAssociation(ce *ast.CallExpr, association string) bool {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	found := false
	ast.Inspect(sel.X, func(n ast.Node) bool {
		if found {
			return false
		}
		candidate, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		candidateSelector, ok := candidate.Fun.(*ast.SelectorExpr)
		if !ok || candidateSelector.Sel.Name != "Preload" || len(candidate.Args) == 0 {
			return true
		}
		candidateAssociation, ok := stringLitValue(candidate.Args[0])
		if ok && candidateAssociation == association {
			found, _ = preloadHasClinicScope(candidate)
		}
		return !found
	})
	return found
}

// lastAssocSegment returns the final segment of a (possibly nested) Preload path. The clinic_id
// predicate of a nested Preload applies to its LAST association (e.g. Preload("Pets.Insurance", …)
// scopes Insurance), so master classification keys on that segment. Sensitive intermediate
// associations are checked separately by clinicScopedIntermediateAssoc.
func lastAssocSegment(assoc string) string {
	if i := strings.LastIndex(assoc, "."); i >= 0 {
		return assoc[i+1:]
	}
	return assoc
}

func baseFileName(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func isInSet(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}

// moduleInternalSource returns every production .go file across the WHOLE module's internal/
// tree (internal/repository, internal/service, internal/model, ...) via
// lintscan.WalkInternalTreeT(t). Keys are lintscan's raw path relative to internal/ (e.g.
// "repository/foo.go", "repository/shiftentry/repository.go", "service/vaccination_service.go").
// preload / audit-tx / dbortx (this file, audit_tx_inventory_lint_test.go,
// dbortx_inventory_lint_test.go) all share this single discovery step (BE9-1) so a violation
// planted anywhere under internal/ — not only internal/repository — is caught by all three gates.
func moduleInternalSource(t *testing.T) map[string][]byte {
	t.Helper()
	return WalkInternalTreeT(t)
}

// legacyLintKey converts a lintscan raw key (relative to internal/) into the key shape the
// preload / audit-tx / dbortx lints have used since before BE9-1's module-wide migration, so
// every existing site-exception / allowlist entry keeps matching byte-for-byte:
//
//	"repository/foo.go"                      -> "foo.go"                      (root file)
//	"repository/shiftentry/repository.go" -> "shiftentry/repository.go" (1-level subpkg)
//	"repository/a/b/c.go"                    -> "a/b/c.go"                    (N-level nested)
//
// Any OTHER top-level internal/ package (service/, model/, ...) keeps its full lintscan key
// unchanged: those have no legacy allowlist/site-exception entries, so a new, longer-but-stable
// key shape for anything newly discovered there is expected and fine — only files that originate
// under internal/repository/** need this exact historical shape preserved.
func legacyLintKey(key string) string {
	const repositoryPrefix = "repository/"
	if strings.HasPrefix(key, repositoryPrefix) {
		return strings.TrimPrefix(key, repositoryPrefix)
	}
	return key
}

// assertDiscoversFileFromDifferentTopLevelPackage proves tree (the shared module-wide discovery
// set every one of preload/audit-tx/dbortx's walkers now sources from via moduleInternalSource)
// reaches at least one file OUTSIDE internal/repository — proving real, live module-wide reach
// (BE9-1), not just internal/repository reach.
func assertDiscoversFileFromDifferentTopLevelPackage(t *testing.T, tree map[string][]byte) {
	t.Helper()
	for key := range tree {
		if !strings.HasPrefix(key, "repository/") {
			return
		}
	}
	t.Fatal("module-wide discovery set contains no file outside internal/repository; " +
		"module-wide reach (BE9-1) is not proven — lintscan.WalkInternalTreeT may have narrowed " +
		"back to a single top-level package")
}

// assertLintscanReachesTwoOrMoreNestingLevels proves the shared lintscan discovery mechanism
// underlying moduleInternalSource is not artificially limited to 1-level subpackages. This uses a
// synthetic t.TempDir() tree rather than the real repository because, as of this writing, no real
// internal/ package nests 2+ levels deep (the deepest today is exactly 1-level, e.g.
// repository/shiftentry/repository.go) — the CAPABILITY, not today's real depth, is what BE9-1
// requires proven. See lintscan.WalkInternalTree's own doc comment: "walking arbitrarily deep is
// what makes it agnostic to however many subpackage levels exist today or are added later".
func assertLintscanReachesTwoOrMoreNestingLevels(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	internalDir := filepath.Join(tmp, "internal")
	deep := filepath.Join(internalDir, "pkg", "sub", "deeper", "file.go")
	if err := os.MkdirAll(filepath.Dir(deep), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(deep), err)
	}
	if err := os.WriteFile(deep, []byte("package deeper\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", deep, err)
	}

	tree, err := WalkInternalTree(internalDir)
	if err != nil {
		t.Fatalf("lintscan.WalkInternalTree(%s): %v", internalDir, err)
	}
	const want = "pkg/sub/deeper/file.go"
	if _, ok := tree[want]; !ok {
		t.Fatalf("lintscan.WalkInternalTree did not discover a 2+-level nested file %q; got %v", want, tree)
	}
}

// walkRepositoryPreloads runs the analyzer over every production .go file discovered
// module-wide (moduleInternalSource) and aggregates findings + stats. Filenames fed to
// analyzeFilePreloads are legacyLintKey-normalized so site-exception identity for
// internal/repository/** files is unchanged from the pre-BE9-1 repoSourceFS walk.
func walkRepositoryPreloads(t *testing.T) ([]preloadFinding, preloadStats) {
	t.Helper()

	tree := moduleInternalSource(t)

	var all []preloadFinding
	agg := preloadStats{siteExemptByKey: make(map[string]int)}
	for rawKey, src := range tree {
		key := legacyLintKey(rawKey)
		findings, stats, err := analyzeFilePreloads(key, src)
		if err != nil {
			t.Fatalf("parse %s: %v", key, err)
		}
		all = append(all, findings...)
		agg.filesParsed += stats.filesParsed
		agg.masterPreloads += stats.masterPreloads
		agg.staffExempt += stats.staffExempt
		agg.globalExempt += stats.globalExempt
		agg.siteExempt += stats.siteExempt
		for k, v := range stats.siteExemptByKey {
			agg.siteExemptByKey[k] += v
		}
	}
	return all, agg
}

// TestPreloadClinicScope_RealRepositorySourceHasNoUnscopedMasterPreload is the gate: no
// clinic-scoped master Preload in this package may omit a clinic_id predicate (unless covered
// by a documented site exception). The floor guards prevent a vacuous green if discovery or the
// AST matching silently breaks.
func TestPreloadClinicScope_RealRepositorySourceHasNoUnscopedMasterPreload(t *testing.T) {
	findings, stats := walkRepositoryPreloads(t)

	if stats.filesParsed < 40 {
		t.Fatalf("only %d repository source files parsed; discovery likely broken (would vacuously pass)", stats.filesParsed)
	}
	// Floor, not an exact pin: master preloads grow as the schema grows. A broken AST matcher
	// would drop this to ~0, far below the floor. Update the floor only if it ever exceeds the
	// real count.
	if stats.masterPreloads < 40 {
		t.Fatalf("only %d clinic-scoped master preloads detected; Preload AST matching likely broken", stats.masterPreloads)
	}

	for _, f := range findings {
		t.Errorf("clinic-scope Preload violation: %s:%d Preload(%q): %s", f.file, f.line, f.assoc, f.detail)
	}
	if len(findings) > 0 {
		t.Errorf("%d clinic-scoped master Preload(s) missing clinic_id predicate. "+
			"Add a clinic_id predicate, or (if clinicID is genuinely unavailable) add a documented "+
			"entry to preloadClinicScopeSiteExceptions.", len(findings))
	}
}

// TestPreloadClinicScope_ExemptionsAreLive proves the allowlists are exercised, not dead code —
// the analogue of the precedent's negative-path self-verification.
func TestPreloadClinicScope_ExemptionsAreLive(t *testing.T) {
	_, stats := walkRepositoryPreloads(t)

	if stats.staffExempt == 0 {
		t.Error("no staff-exempt Preloads encountered; staffExemptAssoc may be dead code or assoc names drifted")
	}
	if stats.globalExempt == 0 {
		t.Error("no global-exempt Preloads encountered; globalExemptAssoc may be dead code or assoc names drifted")
	}
	if len(preloadClinicScopeSiteExceptions) > 0 && stats.siteExempt == 0 {
		t.Error("no site exceptions matched; preloadClinicScopeSiteExceptions is stale (the exempted site was fixed or moved)")
	}
}

// TestPreloadClinicScope_Analyzer pins the detection behaviour on inline fixtures: known-bad
// patterns (incl. the b3638d5e regression `Preload("Vaccine", "deleted_at IS NULL")`) must be
// detected, and known-good / exempt / out-of-scope patterns must not. This both drives the
// analyzer (RED→GREEN) and freezes the old failure mode against reintroduction.
func TestPreloadClinicScope_Analyzer(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"master missing clinic_id (b3638d5e regression)", `db.Preload("Vaccine", "deleted_at IS NULL")`, 1},
		{"master with clinic_id =", `db.Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 0},
		{"master with clinic_id IN (#86)", `db.Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", ids)`, 0},
		{"master with no predicate", `db.Preload("Vaccine")`, 1},
		{"nested master last segment missing", `db.Preload("ReservationType.Group", "deleted_at IS NULL")`, 1},
		{"nested master last segment scoped", `db.Preload("ReservationType.Group", "clinic_id IN ? AND deleted_at IS NULL", ids)`, 0},
		{"nested trimming detail missing intermediate scope", `db.Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 1},
		{"nested trimming detail and master scoped", `db.Preload("TrimmingDetail", "clinic_id = ?", clinicID).Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 0},
		{"closure missing clinic_id", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Where("deleted_at IS NULL") })`, 1},
		{"closure with clinic_id literal", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID) })`, 0},
		{"closure with clinicScope helper", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Scopes(clinicScope(clinicID)) })`, 0},
		{"staff alias exempt", `db.Preload("Doctor", "deleted_at IS NULL")`, 0},
		{"global master exempt", `db.Preload("AnimalSpecies")`, 0},
		// Bare Items without root model context is not master-scoped (Billing/Examination Items).
		{"non-master Items ignored without root model", `db.Preload("Items", "deleted_at IS NULL")`, 0},
		{"non-literal predicate fails closed", `db.Preload("Vaccine", cond)`, 1},
		// BE-refactor.md R3-2 (D9・pre-registration follow-up): 事前登録した「未点火の地雷」3件が実際に効くことの証明。
		{"pre-registered ChiefComplaintType missing clinic_id", `db.Preload("ChiefComplaintType", "deleted_at IS NULL")`, 1},
		{"pre-registered ChiefComplaintType scoped", `db.Preload("ChiefComplaintType", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 0},
		{"pre-registered HospitalizationPlan missing clinic_id", `db.Preload("HospitalizationPlan", "deleted_at IS NULL")`, 1},
		{"pre-registered Inventory missing clinic_id", `db.Preload("Inventory", "deleted_at IS NULL")`, 1},
		{"pre-registered Inventory scoped", `db.Preload("Inventory", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package p\nfunc f() { _ = " + tc.body + " }\n"
			findings, _, err := analyzeFilePreloads("fixture.go", []byte(src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

// BUG-437: Preload("Items") is context-aware — ExaminationType.Items (ExamTypeField) is gated;
// Examination.Items / bare Items without root model are not.
func TestPreloadClinicScope_ItemsContextAware(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "ExaminationType Items missing clinic_id is RED",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ Find(interface{}, ...interface{}) error } }, clinicID uint64) {
	exTypes := make([]model.ExaminationType, 0)
	_ = db.Preload("Items", "deleted_at IS NULL").Find(&exTypes)
}
`,
			want: 1,
		},
		{
			name: "ExaminationType Items string-scoped is GREEN",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ Find(interface{}, ...interface{}) error } }, clinicID uint64) {
	exTypes := make([]model.ExaminationType, 0)
	_ = db.Preload("Items", "clinic_id = ?", clinicID).Find(&exTypes)
}
`,
			want: 0,
		},
		{
			name: "ExaminationType Items closure-scoped is GREEN",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ First(interface{}, ...interface{}) error } }, clinicID uint64) {
	var exType model.ExaminationType
	_ = db.Preload("Items", func(itemsDB interface{ Where(string, ...interface{}) interface{} }) interface{} {
		return itemsDB.Where("clinic_id = ?", clinicID)
	}).First(&exType)
}
`,
			want: 0,
		},
		{
			name: "Examination Items (non-master) is GREEN even unscoped",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ Find(interface{}, ...interface{}) error } }) {
	var exams []*model.Examination
	_ = db.Preload("Items").Find(&exams)
}
`,
			want: 0,
		},
		{
			name: "nested ExaminationType.Items unscoped is RED",
			src: `package p
func f(db interface{ Preload(string, ...interface{}) error }) {
	_ = db.Preload("ExaminationType.Items", "deleted_at IS NULL")
}
`,
			want: 1,
		},
		{
			name: "nested ExaminationType.Items on Examination Find root is still RED",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ Find(interface{}, ...interface{}) error } }) {
	var exams []*model.Examination
	_ = db.Preload("ExaminationType.Items", "deleted_at IS NULL").Find(&exams)
}
`,
			want: 1,
		},
		{
			name: "nested ExaminationType.Items scoped on Examination Find is GREEN",
			src: `package p
import "model"
func f(db interface{ Preload(string, ...interface{}) interface{ Find(interface{}, ...interface{}) error } }, clinicID uint64) {
	var exams []*model.Examination
	_ = db.Preload("ExaminationType.Items", "clinic_id = ?", clinicID).Find(&exams)
}
`,
			want: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, _, err := analyzeFilePreloads("fixture.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

// TestPreloadClinicScope_AnalyzerDetectsViolationUnderNestedPathFilename proves the analyzer
// itself (not just discovery reach — see
// TestPreloadClinicScope_DiscoveryReachesModuleWideAndNestedPackages) flags a known-bad Preload
// pattern identically regardless of the filename/location shape the source is presented under: a
// repository-root-shaped filename, a DIFFERENT top-level internal/ package filename, and a
// 2+-level nested filename must all report the SAME violation (BE9-1: module-wide, depth-agnostic
// scanning). BE8-0 follow-up: the prior 3 embed-reachability tests never fed a violation through
// the analyzer at all (see BE-refactor.md BE8-0 follow-up note).
func TestPreloadClinicScope_AnalyzerDetectsViolationUnderNestedPathFilename(t *testing.T) {
	src := "package p\nfunc f() { _ = db.Preload(\"Vaccine\", \"deleted_at IS NULL\") }\n"
	filenames := []string{
		"repository/x.go",          // root-shaped
		"service/x.go",             // a different top-level internal/ package
		"repository/two/deep/x.go", // 2+-level nested
	}
	for _, fn := range filenames {
		t.Run(fn, func(t *testing.T) {
			findings, _, err := analyzeFilePreloads(fn, []byte(src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("got %d findings for filename %q, want 1: %+v", len(findings), fn, findings)
			}
		})
	}
}

// TestPreloadClinicScope_ExceptionDoesNotCrossSubpackageBasenameCollision proves a site
// exception scoped to one 1-level domain subpackage's file cannot silently exempt the same
// violation pattern in a DIFFERENT subpackage that happens to share the same basename (all 14
// domain subpackages currently have a "repository.go" — BE-refactor.md §5-1-3). The exception
// below simulates one that was (or would be) registered under the bare basename a pre-fix
// baseFileName(filename) computes for ANY nested file — e.g. it was intended for
// paymentmethod/repository.go. A different subpackage (vaccine/) with an identical unscoped
// Vaccine Preload must NOT inherit that waiver.
func TestPreloadClinicScope_ExceptionDoesNotCrossSubpackageBasenameCollision(t *testing.T) {
	original := preloadClinicScopeSiteExceptions
	t.Cleanup(func() { preloadClinicScopeSiteExceptions = original })
	preloadClinicScopeSiteExceptions = []preloadSiteException{
		{
			file:        "repository.go",
			assoc:       "Vaccine",
			predicate:   "deleted_at IS NULL",
			occurrences: 1,
			reason:      "BE8-0 collision-safety fixture: an exception intentionally scoped to one specific domain subpackage file only",
		},
	}

	src := "package p\nfunc f() { _ = db.Preload(\"Vaccine\", \"deleted_at IS NULL\") }\n"
	findings, _, err := analyzeFilePreloads("vaccine/repository.go", []byte(src))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("vaccine/repository.go's unscoped Vaccine Preload was silently exempted by a site exception " +
			"keyed under the bare basename \"repository.go\" — this is the cross-subpackage basename collision " +
			"BE-refactor.md §5-1-3 warns about: paymentmethod/repository.go and vaccine/repository.go must not " +
			"share a site-exception identity")
	}
}

// TestPreloadClinicScope_SiteExceptionMatchingIsExact proves a site exception does not leak to
// the normal scoped form: the same (file, assoc) with a clinic_id predicate is NOT exempted,
// and an exception only matches its exact unscoped predicate string.
func TestPreloadClinicScope_SiteExceptionMatchingIsExact(t *testing.T) {
	original := preloadClinicScopeSiteExceptions
	t.Cleanup(func() { preloadClinicScopeSiteExceptions = original })
	preloadClinicScopeSiteExceptions = []preloadSiteException{
		{
			file:        "synthetic/repository.go",
			assoc:       "Occupation",
			predicate:   "deleted_at IS NULL",
			occurrences: 1,
			reason:      "exact-match analyzer fixture; production waivers remain empty",
		},
	}

	if !isSiteExcepted("synthetic/repository.go", "Occupation", "deleted_at IS NULL") {
		t.Error("expected the synthetic Occupation deleted_at-only preload to match exactly")
	}
	if isSiteExcepted("synthetic/repository.go", "Occupation", "clinic_id = ? AND deleted_at IS NULL") {
		t.Error("a clinic_id-scoped Occupation preload must NOT be treated as an exception")
	}
	if isSiteExcepted("vaccine_repository.go", "Vaccine", "deleted_at IS NULL") {
		t.Error("unrelated file/assoc must not match a site exception")
	}
}

// TestPreloadClinicScope_DiscoveryReachesModuleWideAndNestedPackages pins that moduleInternalSource
// (backed by lintscan.WalkInternalTreeT) actually reaches: (a) a real 2+-level production
// package, (b) at least one file from a DIFFERENT top-level internal/ package (proving real
// module-wide reach, not just internal/repository reach), and (c) 2+-level nesting (a scanner
// CAPABILITY proof via a synthetic tree — see assertLintscanReachesTwoOrMoreNestingLevels for why
// the real repository cannot exercise this today). preload_clinic_scope / audit_tx_inventory /
// dbortx_inventory all share this discovery step via moduleInternalSource — if it narrowed back
// to a single top-level package or a fixed depth, a violation outside internal/repository (or
// 2+ levels deep) would silently fall off all three clinical-safety gates without any other test
// catching it (BE9-1; historical precedent: BE-refactor.md BE8-0 §5-1, for the original 1-level
// embed-glob regression).
//
// Renamed + strengthened from the pre-BE9-1 TestRepoSourceEmbed_ReachesOneLevelSubpackages, which
// only pinned the go:embed glob's 1-level reach within this package.
func TestPreloadClinicScope_DiscoveryReachesModuleWideAndNestedPackages(t *testing.T) {
	tree := moduleInternalSource(t)

	foundNestedPackageCanary := false
	nestedProductionFiles := 0
	for n := range tree {
		if n == "infra/smtp/sender.go" {
			foundNestedPackageCanary = true
		}
		if strings.Count(n, "/") >= 2 {
			nestedProductionFiles++
		}
	}
	if !foundNestedPackageCanary {
		t.Fatal("module-wide discovery does not include infra/smtp/sender.go; " +
			"lintscan's walk may have narrowed and would silently drop nested production packages " +
			"from preload/audit-tx/dbortx clinical-safety coverage without any other test catching it")
	}
	// Floor, not an exact pin: a broken/narrowed discovery walk would drop this near 0.
	const nestedFloor = 10
	if nestedProductionFiles < nestedFloor {
		t.Fatalf("only %d nested production source files discovered, want >= %d; "+
			"lintscan's walk may have narrowed", nestedProductionFiles, nestedFloor)
	}

	assertDiscoversFileFromDifferentTopLevelPackage(t, tree)
	assertLintscanReachesTwoOrMoreNestingLevels(t)
}

// TestPreloadClinicScope_SiteExceptionsMatchExpectedOccurrences pins each site exception to its
// exact number of matched call sites. This closes two gaps a coarse (file,assoc,predicate) key
// would otherwise leave open:
//   - a NEW unscoped Preload of the same pattern added to the same file would silently inherit
//     the waiver (count goes UP → fail → forces scoping the new site, not exempting it);
//   - a stale exception left behind after its site was fixed/removed becomes a standing bypass
//     (count goes DOWN → fail → forces deleting the dead entry).
func TestPreloadClinicScope_SiteExceptionsMatchExpectedOccurrences(t *testing.T) {
	_, stats := walkRepositoryPreloads(t)

	for _, ex := range preloadClinicScopeSiteExceptions {
		got := stats.siteExemptByKey[siteExceptionKey(ex.file, ex.assoc, ex.predicate)]
		if got != ex.occurrences {
			t.Errorf("site exception {%s, %q, %q} matched %d call site(s), expected %d. "+
				"If a new unscoped preload was added, scope it with clinic_id instead of inheriting "+
				"this waiver; if the exempted site was fixed/removed, delete the now-stale exception.",
				ex.file, ex.assoc, ex.predicate, got, ex.occurrences)
		}
	}
}

// TestPreloadClinicScope_EndToEndDiscoveryIsLocationAgnostic is the BE9-1 end-to-end proof: using
// a synthetic module-shaped t.TempDir() tree (go.mod + internal/), the SAME minimal bad
// (clinic_id-less) Preload source snippet is planted at 3 different depths/locations — a
// root-level file, a 1-level subpackage file, and a 2+-level nested subpackage file — then
// lintscan.WalkInternalTree (the pure discovery function) is run against that temp internal/ dir
// and every discovered file is fed through the REAL, unmodified analyzeFilePreloads. All 3 must
// report the identical violation. This proves the full discovery+judgment pipeline is
// location-agnostic end-to-end, not just the pure analyzer in isolation (see
// TestPreloadClinicScope_AnalyzerDetectsViolationUnderNestedPathFilename for that narrower proof).
func TestPreloadClinicScope_EndToEndDiscoveryIsLocationAgnostic(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/synthetic\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	internalDir := filepath.Join(tmp, "internal")

	badSrc := []byte("package p\nfunc f() { _ = db.Preload(\"Vaccine\", \"deleted_at IS NULL\") }\n")
	locations := []string{
		filepath.Join(internalDir, "repository", "root_violation.go"),                           // root-level
		filepath.Join(internalDir, "repository", "sub", "one_violation.go"),                     // 1-level subpackage
		filepath.Join(internalDir, "repository", "sub", "deeper", "nested", "two_violation.go"), // 2+-level nested
	}
	for _, loc := range locations {
		if err := os.MkdirAll(filepath.Dir(loc), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(loc), err)
		}
		if err := os.WriteFile(loc, badSrc, 0o644); err != nil {
			t.Fatalf("write %s: %v", loc, err)
		}
	}

	tree, err := WalkInternalTree(internalDir)
	if err != nil {
		t.Fatalf("lintscan.WalkInternalTree(%s): %v", internalDir, err)
	}
	if len(tree) != len(locations) {
		t.Fatalf("got %d discovered files, want %d: %v", len(tree), len(locations), tree)
	}

	total := 0
	for key, src := range tree {
		findings, _, err := analyzeFilePreloads(key, src)
		if err != nil {
			t.Fatalf("parse %s: %v", key, err)
		}
		if len(findings) != 1 {
			t.Errorf("location %s: got %d findings, want 1: %+v", key, len(findings), findings)
		}
		total += len(findings)
	}
	if total != len(locations) {
		t.Fatalf("expected %d total violations across %d planted locations, got %d", len(locations), len(locations), total)
	}
}
