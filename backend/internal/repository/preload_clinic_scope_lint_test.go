package repository

// P3.1 mechanical enforcement — clinic-scoped master Preloads must carry a clinic_id predicate.
//
// Background: a systemic cross-tenant read IDOR (b3638d5e) was fixed by hand: when a
// clinic-scoped "master/区分" is Preloaded by an FK value without a clinic_id predicate,
// a polluted FK (write-side validation gaps #124/#125, or pre-remediation data) makes the
// query return ANOTHER clinic's master name/price. Rule P3.1 (repository/CLAUDE.md) requires
// every such Preload to scope clinic_id. This test enforces that rule mechanically so the
// gap cannot regress — the read-side analogue of the runtime *_clinic_isolation_test.go guards.
//
// Technique is reused from model/audit_taxonomy_exhaustiveness_test.go: the package source is
// embedded with go:embed (so parsing is correct under -trimpath / embedded test binaries),
// parsed with go/parser + go/ast, and pinned with self-verification + good/bad fixtures so the
// check is never vacuously green.

import (
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

// repoSourceFS embeds every .go file in this package directory at compile time.
// The *.go glob also matches *_test.go (including this file); those are skipped at
// runtime in walkRepositoryPreloads so the lint never inspects its own fixtures or
// other test helpers. Embedding (vs parser.ParseDir on a relative path) keeps the
// lint correct under -trimpath — the reason the audit-taxonomy precedent uses embed.
//
//go:embed *.go
var repoSourceFS embed.FS

// clinicScopedMasterAssoc maps a GORM association name (the LAST dotted segment of a
// Preload path) to the clinic-scoped master model it resolves to. Every Preload of one
// of these MUST carry a clinic_id predicate (P3.1).
//
// Membership is curated from the P3.1 master list and confirmed against each model's
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
	"CheckupType":     "CheckupType",
	// #211: 健診パッケージのフィールド定義マスタ（checkup_type_fields, clinic_id 列あり）。
	// checkup_field_results.CheckupTypeField として clinic_id 述語付きで Preload される。
	"CheckupTypeField": "CheckupTypeField",
	"DiagnosisType":    "DiagnosisType",
	"DiagnosisName":    "DiagnosisName",
	"Names":            "DiagnosisName", // diagnosis_type.Names []DiagnosisName
	"Occupation":       "Occupation",
	// BE-refactor.md R3-2 (D9・P1): 以下3件は write 側 clinicScopedMasterFKField
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
	// 本マップへの事前登録はできない。P1 follow-up: これらに association フィールドを追加する際は
	// 同一コミットで本マップにも登録すること（read/write 両ゲートの双方向整合を維持する）。
}

// staffExemptAssoc: P3.1 Staff exception. Staff belongs to multiple clinics via
// staff_clinic_assignments (staffs.clinic_id is the primary only). Scoping a historical
// Doctor/Staff preload by staffs.clinic_id would wrongly hide shared/reassigned staff from
// past records; the leak is a staff NAME only, low severity, and unreachable in normal flow
// via write isolation (72e8887c). These names all resolve to the Staff model.
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
// with a P1 follow-up — NOT a blanket "writes are validated" waiver. The key is the exact
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

var preloadClinicScopeSiteExceptions = []preloadSiteException{
	{
		file:        "staff_repository.go",
		assoc:       "Occupation",
		predicate:   "deleted_at IS NULL",
		occurrences: 1, // FindByID(id) only; FindAll is clinic_id-scoped
		reason: "FindByID(id) is a cross-clinic identity/auth lookup with no clinicID param; " +
			"the Occupation NAME leak is low severity; threading clinicID would break the " +
			"identity-lookup contract used by auth/service callers. P1: scope via a dedicated " +
			"call or drop the preload.",
	},
	{
		file:        "reservation_staff_repository.go",
		assoc:       "ReservationType",
		predicate:   "deleted_at IS NULL",
		occurrences: 2, // covers FindAllExcludedReservationTypes + ...ByStaffIDs (2 call sites)
		reason: "FindAllExcludedReservationTypes(ByStaffIDs) key by staff_id with no clinicID " +
			"param; threading it cascades through the service+handler layers. Writes validate " +
			"reservation_type clinic ownership (UpdateExcludedReservationTypes). Residual: " +
			"cross-clinic type NAME via past pollution, low severity. P1: plumb clinicID.",
	},
}

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
// lack a clinic_id predicate. It is a pure function over (filename, src) so the real embedded
// source and the inline fixtures exercise exactly the same logic.
func analyzeFilePreloads(filename string, src []byte) ([]preloadFinding, preloadStats, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, preloadStats{}, err
	}

	var findings []preloadFinding
	stats := preloadStats{filesParsed: 1}
	base := baseFileName(filename)

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
		key := lastAssocSegment(assoc)

		switch {
		case isInSet(staffExemptAssoc, key):
			stats.staffExempt++
			return true
		case isInSet(globalExemptAssoc, key):
			stats.globalExempt++
			return true
		}
		if _, isMaster := clinicScopedMasterAssoc[key]; !isMaster {
			return true // OTHER (Owner/Pet/Items/Payments/...) — outside P3.1 master scope
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
// KISS tradeoff (matches the string-predicate path and P3.1's documented stance): this is a
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

// lastAssocSegment returns the final segment of a (possibly nested) Preload path. The clinic_id
// predicate of a nested Preload applies to its LAST association (e.g. Preload("Pets.Insurance", …)
// scopes Insurance), so classification keys on that segment. Intermediate segments are NOT
// validated — no intermediate segment resolves to a clinic-scoped master in current code; a future
// Preload("Master.Child", …) whose Child is a non-master would not be checked.
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

// walkRepositoryPreloads runs the analyzer over every non-test .go file embedded from this
// package directory and aggregates findings + stats.
func walkRepositoryPreloads(t *testing.T) ([]preloadFinding, preloadStats) {
	t.Helper()

	names, err := fs.Glob(repoSourceFS, "*.go")
	if err != nil {
		t.Fatalf("glob embedded repository source: %v", err)
	}

	var all []preloadFinding
	agg := preloadStats{siteExemptByKey: make(map[string]int)}
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := repoSourceFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", name, err)
		}
		findings, stats, err := analyzeFilePreloads(name, src)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
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
// by a documented site exception). The floor guards prevent a vacuous green if the embed glob
// or AST matching silently breaks.
func TestPreloadClinicScope_RealRepositorySourceHasNoUnscopedMasterPreload(t *testing.T) {
	findings, stats := walkRepositoryPreloads(t)

	if stats.filesParsed < 40 {
		t.Fatalf("only %d repository source files parsed; embed glob likely broken (would vacuously pass)", stats.filesParsed)
	}
	// Floor, not an exact pin: master preloads grow as the schema grows. A broken AST matcher
	// would drop this to ~0, far below the floor. Update the floor only if it ever exceeds the
	// real count.
	if stats.masterPreloads < 40 {
		t.Fatalf("only %d clinic-scoped master preloads detected; Preload AST matching likely broken", stats.masterPreloads)
	}

	for _, f := range findings {
		t.Errorf("P3.1 violation: %s:%d Preload(%q): %s", f.file, f.line, f.assoc, f.detail)
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
	if stats.siteExempt == 0 {
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
		{"nested trimming master scoped", `db.Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID)`, 0},
		{"closure missing clinic_id", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Where("deleted_at IS NULL") })`, 1},
		{"closure with clinic_id literal", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID) })`, 0},
		{"closure with clinicScope helper", `db.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Scopes(clinicScope(clinicID)) })`, 0},
		{"staff alias exempt", `db.Preload("Doctor", "deleted_at IS NULL")`, 0},
		{"global master exempt", `db.Preload("AnimalSpecies")`, 0},
		{"non-master ignored", `db.Preload("Items", "deleted_at IS NULL")`, 0},
		{"non-literal predicate fails closed", `db.Preload("Vaccine", cond)`, 1},
		// BE-refactor.md R3-2 (D9・P1): 事前登録した「未点火の地雷」3件が実際に効くことの証明。
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

// TestPreloadClinicScope_SiteExceptionMatchingIsExact proves a site exception does not leak to
// the normal scoped form: the same (file, assoc) with a clinic_id predicate is NOT exempted,
// and an exception only matches its exact unscoped predicate string.
func TestPreloadClinicScope_SiteExceptionMatchingIsExact(t *testing.T) {
	if !isSiteExcepted("staff_repository.go", "Occupation", "deleted_at IS NULL") {
		t.Error("expected staff_repository.go Occupation deleted_at-only preload to be a known exception")
	}
	if isSiteExcepted("staff_repository.go", "Occupation", "clinic_id = ? AND deleted_at IS NULL") {
		t.Error("a clinic_id-scoped Occupation preload must NOT be treated as an exception")
	}
	if isSiteExcepted("vaccine_repository.go", "Vaccine", "deleted_at IS NULL") {
		t.Error("unrelated file/assoc must not match a site exception")
	}
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
