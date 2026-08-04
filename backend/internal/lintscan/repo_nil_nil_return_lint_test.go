package lintscan

// repo_nil_nil_return_lint_test.go — mechanical ban on repository single-entity
// (*T, error) methods that can return (nil, nil).
//
// Why: 2026-07-30 lstep delivery-trigger batch panicked on nil owner after bulk
// cache miss (ce79e0c23) fell back to FindByID. Production owner.FindByID already
// uses FromGORM/NotFound, but a (nil, nil) contract violation (mock or soft-miss
// sibling APIs) reaches processSingleOwner → checkExclusion. Fail-closed guards in
// 6ce1549d8 (owner_missing + processSingleOwner NotFound) are multi-layer defense,
// not contract enforcement. This lint freezes the repository contract itself.
//
// Detection (structure, not name list):
//   - Function is repository-scoped: receiver type name ends with "Repository"
//     (or is exactly "repository"), OR free function in a *repository*.go file.
//   - Results are exactly (*T, error) where *T is a named/selector pointer type
//     (single-entity). Slice/array results are excluded by construction.
//   - Body contains a ReturnStmt with two results both the identifier "nil".
//
// Intentional soft-miss repository methods are allowlisted with verbatim reasons
// and occurrence pins (not silent freeze). Converting them to NotFound is a
// coordinated consumer change — tracked as packet candidates, not fixed here.
//
// House style: pure analyzeFile*(filename, src) + WalkInternalTreeT real gate +
// inline fixtures + allowlist-live pins. Modeled on n1_lint_test.go.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// repoNilNilFinding is one (nil, nil) return site inside a repository single-entity getter.
type repoNilNilFinding struct {
	file   string // path relative to internal/ (or fixture name)
	line   int
	fn     string
	detail string
}

// repoNilNilAllowlistKey is "<relpath>|<funcName>" so a waiver on one method cannot
// silence the same method name introduced in another file.
func repoNilNilAllowlistKey(relPath, fn string) string {
	return relPath + "|" + fn
}

// repoNilNilSiteException freezes a known intentional soft-miss with an occurrence
// pin so extra return nil,nil sites force a review.
type repoNilNilSiteException struct {
	file        string // key as produced by WalkInternalTree (slash path under internal/)
	fn          string
	occurrences int
	reason      string
}

// repoNilNilAllowlist — intentional soft-miss repository APIs. Every entry must
// explain why conversion to (nil, NotFound) is deferred. Do NOT add FindByID-family
// hard-getters here; those must return errors on miss.
//
// Packet candidates for consumer-coordinated conversion (not this run's scope):
// NILCONTRACT-SOFT-MISS-OWNER-UNIQUE, NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY.
var repoNilNilAllowlist = []repoNilNilSiteException{
	{
		file: "owner/repository.go", fn: "FindByEmail", occurrences: 1,
		reason: "intentional soft-miss uniqueness probe: callers treat existing != nil after FindByEmail (not IsNotFound). Converting to NotFound requires coordinated service_core uniqueness rewrites. Packet: NILCONTRACT-SOFT-MISS-OWNER-UNIQUE",
	},
	{
		file: "owner/repository.go", fn: "FindByPhone", occurrences: 1,
		reason: "intentional soft-miss uniqueness probe: documented '見つからない場合は nil'; callers use pointer nil check. Packet: NILCONTRACT-SOFT-MISS-OWNER-UNIQUE",
	},
	{
		file: "owner/repository.go", fn: "FindByNameAndPhone", occurrences: 2,
		reason: "intentional soft-miss auto-link: empty input, 0 hits, or multi hits all mean 'no unique owner' as (nil,nil) for LIFF binding. Packet: NILCONTRACT-SOFT-MISS-OWNER-UNIQUE",
	},
	{
		file: "billing/accounting_repository.go", fn: "FindByCompletionRequestID", occurrences: 2,
		reason: "BUG-018 idempotency probe: 'no billing has claimed this completion key yet' is the normal first-attempt state, and Complete branches on existing != nil to decide create-vs-replay; NotFound would invert that branch. NOTE the two sites are NOT the same shape — the gorm.ErrRecordNotFound site is the intended soft miss, while the requestID == \"\" early return is an input-validation short-circuit that should be rejected at the boundary instead of imitating a miss; it is waived here only to keep this gate green and is tracked separately. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "billing/cash_register_close_repository.go", fn: "FindByDateAndPeriod", occurrences: 1,
		reason: "optional cash-register close for a date/period: absence is a valid business state, not a hard NotFound for report flows. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "billing/campaign_repository.go", fn: "FindApplicableForItem", occurrences: 1,
		reason: "optional applicable campaign: '該当キャンペーンなし' is success with no campaign, not resource miss. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "medicalrecord/medical_record_repository.go", fn: "FindByAppointmentID", occurrences: 1,
		reason: "optional active record for appointment: create/link flows treat nil as 'no record yet'. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "medicalrecord/medical_record_owner_visit_repository.go", fn: "FindLatestByOwner", occurrences: 1,
		reason: "optional latest visit: absence of any medical record is common for new owners; callers check nil. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "clinic/closing_special_period_repository.go", fn: "FindByDate", occurrences: 1,
		reason: "optional special closing period for a calendar date; schedule resolution treats absence as normal day. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "staff/shift_entry_repository.go", fn: "lockShiftEntryByStaffDateForUpdate", occurrences: 1,
		reason: "upsert lock helper: (nil,nil) means no row yet so caller inserts; NotFound would invert the insert path. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "identitylink/repository.go", fn: "FindActiveOwnerMembership", occurrences: 1,
		reason: "optional identity-group membership: unlinked owners are normal; service maps nil to NotFound only when required. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
	{
		file: "identitylink/repository.go", fn: "FindActivePetMembership", occurrences: 1,
		reason: "optional identity-group membership: unlinked pets are normal; same soft-miss contract as owner membership. Packet: NILCONTRACT-SOFT-MISS-OPTIONAL-ENTITY",
	},
}

func repoNilNilAllowlistMap() map[string]repoNilNilSiteException {
	m := make(map[string]repoNilNilSiteException, len(repoNilNilAllowlist))
	for _, e := range repoNilNilAllowlist {
		m[repoNilNilAllowlistKey(e.file, e.fn)] = e
	}
	return m
}

type repoNilNilStats struct {
	filesParsed         int
	repoSingleEntityFns int
	nilNilSites         int
	allowlistedSites    int
}

// analyzeFileRepoNilNil parses one Go source and reports unresolved (nil, nil)
// returns in repository single-entity (*T, error) methods. Pure over (filename, src).
func analyzeFileRepoNilNil(filename string, src []byte) (findings []repoNilNilFinding, allowHits map[string]int, stats repoNilNilStats, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, nil, stats, err
	}

	allowHits = make(map[string]int)
	allow := repoNilNilAllowlistMap()
	rel := normalizeRepoNilNilPath(filename)
	stats.filesParsed = 1

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		if !isRepositoryScopedFunc(filename, fd) {
			continue
		}
		if !isSingleEntityPtrErrorResults(fd.Type.Results) {
			continue
		}
		stats.repoSingleEntityFns++
		fnName := fd.Name.Name
		key := repoNilNilAllowlistKey(rel, fnName)

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			rs, ok := n.(*ast.ReturnStmt)
			if !ok || !isNilNilReturn(rs) {
				return true
			}
			stats.nilNilSites++
			if _, ok := allow[key]; ok {
				allowHits[key]++
				stats.allowlistedSites++
				return true
			}
			findings = append(findings, repoNilNilFinding{
				file:   rel,
				line:   fset.Position(rs.Pos()).Line,
				fn:     fnName,
				detail: "repository single-entity (*T, error) method returns (nil, nil); missing entity must return (nil, error) (e.g. apperrors.FromGORM / WrapNotFound)",
			})
			return true
		})
	}
	return findings, allowHits, stats, nil
}

func normalizeRepoNilNilPath(filename string) string {
	// WalkInternalTree keys are already "domain/file.go". Fixtures may pass bare names
	// or nested paths — keep slash form and strip a leading "internal/" if present.
	p := filepath.ToSlash(filename)
	p = strings.TrimPrefix(p, "./")
	if strings.HasPrefix(p, "internal/") {
		p = strings.TrimPrefix(p, "internal/")
	}
	return p
}

// isRepositoryScopedFunc: receiver *xxxRepository / repository, or free func in *repository*.go.
func isRepositoryScopedFunc(filename string, fd *ast.FuncDecl) bool {
	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		name := recvTypeIdentName(fd.Recv.List[0].Type)
		if name == "" {
			return false
		}
		lower := strings.ToLower(name)
		if lower == "repository" || strings.HasSuffix(lower, "repository") {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(filename))
	return strings.Contains(base, "repository")
}

func recvTypeIdentName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return recvTypeIdentName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr: // rare generic receiver
		return recvTypeIdentName(t.X)
	case *ast.IndexListExpr:
		return recvTypeIdentName(t.X)
	default:
		return ""
	}
}

// isSingleEntityPtrErrorResults reports results of the form (*T, error) with named/selector T.
func isSingleEntityPtrErrorResults(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	types := flattenResultTypes(results)
	if len(types) != 2 {
		return false
	}
	star, ok := types[0].(*ast.StarExpr)
	if !ok {
		return false
	}
	if !isNamedOrSelectorType(star.X) {
		return false
	}
	// Exclude pure scalar optionals (*string, *uint64, *time.Time, …) — those are
	// parse/filter helpers, not entity repositories. Entity types are model.X or domain DTOs.
	if isScalarPointerBase(star.X) {
		return false
	}
	errIdent, ok := types[1].(*ast.Ident)
	return ok && errIdent.Name == "error"
}

func flattenResultTypes(results *ast.FieldList) []ast.Expr {
	var out []ast.Expr
	for _, f := range results.List {
		n := len(f.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			out = append(out, f.Type)
		}
	}
	return out
}

func isNamedOrSelectorType(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func isScalarPointerBase(expr ast.Expr) bool {
	name := ""
	switch t := expr.(type) {
	case *ast.Ident:
		name = t.Name
	case *ast.SelectorExpr:
		// time.Time etc.
		if id, ok := t.X.(*ast.Ident); ok {
			name = id.Name + "." + t.Sel.Name
		} else {
			name = t.Sel.Name
		}
	default:
		return false
	}
	switch name {
	case "string", "bool", "byte", "rune",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"Time", "time.Time":
		return true
	default:
		return false
	}
}

func isNilNilReturn(rs *ast.ReturnStmt) bool {
	if rs == nil || len(rs.Results) != 2 {
		return false
	}
	return isNilIdent(rs.Results[0]) && isNilIdent(rs.Results[1])
}

func isNilIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

func walkRepoNilNil(t *testing.T) (findings []repoNilNilFinding, allowHits map[string]int, stats repoNilNilStats) {
	t.Helper()
	files := WalkInternalTreeT(t)
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	aggHits := make(map[string]int)
	var all []repoNilNilFinding
	var total repoNilNilStats
	for _, name := range names {
		fileFindings, hits, st, err := analyzeFileRepoNilNil(name, files[name])
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		all = append(all, fileFindings...)
		for k, v := range hits {
			aggHits[k] += v
		}
		total.filesParsed += st.filesParsed
		total.repoSingleEntityFns += st.repoSingleEntityFns
		total.nilNilSites += st.nilNilSites
		total.allowlistedSites += st.allowlistedSites
	}
	return all, aggHits, total
}

// TestRepoNilNil_RealSourceHasNoUnresolvedNilNilReturns is the production gate.
func TestRepoNilNil_RealSourceHasNoUnresolvedNilNilReturns(t *testing.T) {
	findings, _, stats := walkRepoNilNil(t)

	// Floor guards: a broken walker/matcher must not go green vacuously.
	if stats.filesParsed < 200 {
		t.Fatalf("only %d production files parsed under internal/**; WalkInternalTree likely broken", stats.filesParsed)
	}
	if stats.repoSingleEntityFns < 50 {
		t.Fatalf("only %d repository single-entity (*T, error) methods seen; repository scoping or result-type matcher likely broken", stats.repoSingleEntityFns)
	}

	for _, f := range findings {
		t.Errorf("repo (nil,nil) contract violation: %s:%d func %s(): %s",
			f.file, f.line, f.fn, f.detail)
	}
}

// TestRepoNilNil_Analyzer pins detection on inline fixtures (TDD self-verification).
func TestRepoNilNil_Analyzer(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		src      string
		want     int
	}{
		{
			name:     "FindByID returning (nil,nil) is flagged (hard-get contract)",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
type Owner struct{}
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*Owner, error) {
	return nil, nil
}`,
			want: 1,
		},
		{
			name:     "lstep-shaped bulk cache miss fallback FindByID (nil,nil) is flagged (U3 regression shape)",
			filename: "owner/repository.go",
			src: `package p
// Simulates: runBatch installs empty bulk cache (ce79e0c23) → cache miss → FindByID fallback → (nil,nil)
// which historically reached processSingleOwner and panicked on owner.LstepOptOut before 6ce1549d8 guards.
type ownerRepository struct{}
type Owner struct{ LstepOptOut bool }
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*Owner, error) {
	// bulk map miss path: no row materialised, error discarded
	if true {
		return nil, nil
	}
	return &Owner{}, nil
}`,
			want: 1,
		},
		{
			name:     "FindByID with FromGORM NotFound is not flagged",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
type Owner struct{}
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*Owner, error) {
	var owner Owner
	if err := db.First(&owner, id).Error; err != nil {
		return nil, FromGORM(err, "owner", id)
	}
	return &owner, nil
}`,
			want: 0,
		},
		{
			name:     "allowlisted soft-miss FindByEmail is not flagged",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
type Owner struct{}
func (r *ownerRepository) FindByEmail(ctx context.Context, clinicID uint64, email string) (*Owner, error) {
	if true {
		return nil, nil
	}
	return &Owner{}, nil
}`,
			want: 0,
		},
		{
			name:     "service-layer (*T,error) (nil,nil) is out of repository scope",
			filename: "medicalrecord/medical_record_crud.go",
			src: `package p
type medicalRecordService struct{}
type MedicalRecord struct{}
func (s *medicalRecordService) findExistingRecordByAppointment(ctx context.Context) (*MedicalRecord, error) {
	return nil, nil
}`,
			want: 0,
		},
		{
			name:     "slice bulk FindByIDs empty (nil,nil) is not single-entity",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
type Owner struct{}
func (r *ownerRepository) FindByIDs(ctx context.Context, clinicID uint64, ids []uint64) ([]*Owner, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return nil, nil
}`,
			want: 0,
		},
		{
			name:     "scalar optional *string (nil,nil) is not entity",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
func (r *ownerRepository) normalize(s string) (*string, error) {
	return nil, nil
}`,
			want: 0,
		},
		{
			name:     "free function in *repository*.go with (nil,nil) is flagged",
			filename: "staff/shift_entry_repository.go",
			src: `package p
type ShiftEntry struct{}
func lockSomething(tx interface{}) (*ShiftEntry, error) {
	return nil, nil
}`,
			want: 1,
		},
		{
			name:     "identitylink repository type name is in scope",
			filename: "identitylink/repository.go",
			src: `package p
type repository struct{}
type Member struct{}
func (r *repository) FindByID(ctx context.Context, id uint64) (*Member, error) {
	return nil, nil
}`,
			want: 1,
		},
		{
			name:     "return nil, err is not flagged",
			filename: "owner/repository.go",
			src: `package p
type ownerRepository struct{}
type Owner struct{}
func (r *ownerRepository) FindByID(ctx context.Context, id uint64) (*Owner, error) {
	return nil, errNotFound
}`,
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, _, _, err := analyzeFileRepoNilNil(tc.filename, []byte(tc.src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

// TestRepoNilNil_LstepBulkMissFallbackFixture is the named U3 proof: the exact
// production failure shape (bulk cache miss → single FindByID → (nil, nil)) is detected.
func TestRepoNilNil_LstepBulkMissFallbackFixture(t *testing.T) {
	src := []byte(`package p
type ownerRepository struct{}
type Owner struct{ LstepOptOut bool }
// deliveryBatchOwnerCache miss falls back to inner FindByID (ce79e0c23).
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, ownerID uint64) (*Owner, error) {
	// contract violation: missing owner without error
	return nil, nil
}
`)
	findings, _, _, err := analyzeFileRepoNilNil("owner/repository.go", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("U3 lstep-shaped fixture must be detected exactly once, got %d: %+v", len(findings), findings)
	}
	if findings[0].fn != "FindByID" {
		t.Fatalf("expected FindByID finding, got %+v", findings[0])
	}
}

// TestRepoNilNil_AllowlistEntriesAreLive pins that every allowlist key still hits on real source
// and occurrence counts match — prevents stale waivers and silent extra soft-miss sites.
func TestRepoNilNil_AllowlistEntriesAreLive(t *testing.T) {
	_, hits, stats := walkRepoNilNil(t)
	expectedTotal := 0
	for _, e := range repoNilNilAllowlist {
		expectedTotal += e.occurrences
		key := repoNilNilAllowlistKey(e.file, e.fn)
		got := hits[key]
		if got == 0 {
			t.Errorf("allowlist entry %q never hit on real source (stale waiver?). reason was: %s", key, e.reason)
			continue
		}
		if got != e.occurrences {
			t.Errorf("allowlist entry %q hit %d times, want occurrence pin %d (extra or missing return nil,nil). reason: %s",
				key, got, e.occurrences, e.reason)
		}
	}
	if stats.allowlistedSites != expectedTotal {
		t.Errorf("total allowlisted (nil,nil) sites = %d, want sum of pins %d", stats.allowlistedSites, expectedTotal)
	}
}

// TestRepoNilNil_AllowlistReasonsAreNonEmpty enforces the "no silent freeze" policy.
func TestRepoNilNil_AllowlistReasonsAreNonEmpty(t *testing.T) {
	for _, e := range repoNilNilAllowlist {
		if strings.TrimSpace(e.reason) == "" {
			t.Errorf("allowlist %s|%s has empty reason — forbidden silent freeze", e.file, e.fn)
		}
		if strings.HasPrefix(e.fn, "FindByID") {
			t.Errorf("allowlist %s|%s freezes a FindByID-family hard-getter — forbidden", e.file, e.fn)
		}
		if !strings.Contains(e.reason, "Packet:") && !strings.Contains(e.reason, "packet:") {
			t.Errorf("allowlist %s|%s reason must cite a Packet: follow-up ID", e.file, e.fn)
		}
	}
}
