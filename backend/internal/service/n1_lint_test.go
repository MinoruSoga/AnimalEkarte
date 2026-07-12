package service

// n1_lint_test.go — PERF-FOLLOWUP-07: N+1 クエリ静的検出 lint。
//
// Placement decision (deviates from the spec doc's literal path
// docs/tasks/open/PERF-FOLLOWUP-07-n1-lint-static-detection.md, which suggests
// internal/repository/n1_lint_test.go scanning the service/ directory): go:embed can only
// embed files inside its OWN package's directory subtree. The precedent
// repository/preload_clinic_scope_lint_test.go works because it lives IN package repository
// and embeds repository/*.go — a repository-package file cannot go:embed service/*.go. This
// lint instead lives in package service and embeds its own directory, which sidesteps the
// cross-package embed restriction while reusing the same technique (embed + go/parser +
// go/ast + curated allowlist + self-verification fixtures) as the precedent and as
// master_fk_write_inventory_lint_test.go (same package). It reuses that file's existing
// `serviceSourceFS` embed.FS (both files are package service; a second `//go:embed *.go`
// declaring a second variable would be redundant, not a conflict, but reuse is simpler).
//
// Detection targets (see spec doc):
//   Pattern 1: an *ast.RangeStmt body containing a call `<x>Repo.Find*(...)` — a repository
//              field (name ends "Repo") calling a Find-family method.
//   Pattern 2: an *ast.RangeStmt body containing a call `settingsSvc.Get*(...)` — the
//              settings-service field calling a Get-family method.
// Both are the shape of the PERF-1/PERF-2 (M1/M2) bugs fixed by hoisting the call above the
// loop: a per-clinic-cost lookup (mapping/threshold fetch) that was mistakenly re-run once
// per loop iteration instead of once before the loop.
//
// Scope: *ast.RangeStmt only (`for _, x := range xs` / `for i := range xs`). The
// PERF-FOLLOWUP-02 cursor-pagination loops added alongside this lint
// (`for { page, err := repo.FindXxxCursor(...); ...; break }` in
// lstep_health_tag_sync_batch.go's SyncHealthPreventionTagsForClinic and
// lstep_batch_dormant.go's DetectDormantOwners) are plain *ast.ForStmt nodes with a nil
// Cond/Post/Init — NOT *ast.RangeStmt — so they are structurally excluded by construction,
// with no allowlist entry needed. Verified by reading both functions and by the negative
// fixture TestN1Lint_Analyzer/"pagination for{} loop is not a RangeStmt, not flagged" below.
//
// KISS tradeoff (matches the precedent's stance): this is syntactic pattern matching on a
// single file's AST, not interprocedural/data-flow analysis. A repo/settings call reached
// only indirectly — e.g. a range loop invoking a closure or another service method which
// itself (in a DIFFERENT function) performs the per-iteration repo call — is NOT detected
// here; that shape is exactly how the batch functions call their WithMappings per-owner
// sync methods (SyncHealthcheckTagsWithMappings etc. call checkupRepo.FindByOwnerID once
// per owner, but the range loop and the repo call are lexically in different functions).
// Catching that would require call-graph analysis; the syntactic-only scope is a deliberate
// KISS choice, same as preload_clinic_scope_lint_test.go's substring/identifier heuristic.
//
// KNOWN REAL FINDINGS, TRACKED NOT SILENCED: this lint's first run against real source
// surfaced two genuine, previously-undiscovered N+1 candidates outside the M1/M2/FOLLOWUP-02
// scope of the task that added this gate (PERF-FOLLOWUP-07 was a test-only addition; fixing
// unrelated pre-existing N+1s was out of that task's scope):
//   - lstep_tag_sync_visit_ltv.go SyncLTVTopPercent(): tagCacheRepo.FindByOwner called once
//     per owner inside a loop over ALL clinic owners — same shape as the fixed M1/M2 bugs.
//   - liff_service_availability_slots.go buildStaffSlotInputs(): scheduleRepo.FindAllByDate
//     and scheduleRepo.FindAllBreaksByEntryID called once per staff inside a loop over all
//     clinic staff.
// Rather than either (a) leaving this gate permanently red — which teaches everyone to
// ignore a failing gate and blocks it from ever catching a NEW regression — or (b) silently
// allowlisting them with no trail, both are filed as a tracked follow-up
// (docs/tasks/open/PERF-FOLLOWUP-08-ltv-staffslots-n1.md) and allowlisted below with a
// direct reference to that doc. This mirrors the "documented site-exception" discipline
// preload_clinic_scope_lint_test.go already uses for known-but-deferred gaps: visible in the
// allowlist comment and in BE_todo.md, not swept under the rug. When PERF-FOLLOWUP-08 lands,
// delete both entries below and their occurrence pins in TestN1Lint_AllowlistEntriesAreLive.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"
)

// n1Finding is one loop-body Find/Get call site flagged as an N+1 candidate.
type n1Finding struct {
	file   string
	line   int
	fn     string // enclosing top-level function name
	callee string // e.g. "ownerRepo.FindByID" or "settingsSvc.GetHealthPreventionThresholds"
}

// n1AllowlistKey is "<enclosing func name>:<callee>", matching the granularity of the
// precedent's site-exception keying (function + call site, not just call site alone, so a
// legitimately-allowlisted call in one function does not silently exempt a same-named call
// newly introduced in another function).
func n1AllowlistKey(fn, callee string) string {
	return fn + ":" + callee
}

// n1Allowlist has two categories of entries. Read the reason comment on each before adding
// a new one — the two categories have very different bars:
//
//  1. Request-bound (genuinely not N+1): the loop bound is a single REQUEST's payload size (a
//     form the user submitted), not a clinic-wide DB result set, so it cannot scale with data
//     growth. Do NOT add a new entry in this category for a loop bound by a repository
//     Find*/FindAll result (clinic-wide data) — that is a REAL candidate and must be
//     investigated/fixed or reported, never silently placed here.
//  2. Tracked pre-existing debt (genuinely IS N+1, deliberately deferred): a real, data-bound
//     N+1 outside the scope of the task that discovered it, filed as its own follow-up task
//     and referenced by doc path so the waiver has a paper trail and an exit condition. Every
//     entry in this category MUST reference a docs/tasks/open/*.md follow-up; an entry with no
//     tracked follow-up is silent suppression, which this gate exists to prevent.
var n1Allowlist = map[string]bool{
	// --- Category 1: request-bound, not real N+1 ---
	// pets []CreatePetForOwnerInput is the request body of a single CreateWithPets call
	// (a human filling in one "add owner + pets" form) — bounded by form size, not table size.
	n1AllowlistKey("validateOwnerPetsInsuranceOwnership", "insuranceRepo.FindByID"): true,
	// input.TrimmingOptionIDs is the request body of a single reservation create — bounded by
	// how many trimming options one reservation can select (small, UI-constrained), not table size.
	n1AllowlistKey("ValidateAndCreate", "trimmingOptionRepo.FindByID"): true,

	// --- Category 2: tracked pre-existing debt (docs/tasks/open/PERF-FOLLOWUP-08-ltv-staffslots-n1.md) ---
	// SyncLTVTopPercent loops over ALL clinic owners (ownerRepo.FindAllWithLineUserID) and
	// re-fetches tagCacheRepo.FindByOwner per owner — same shape as the fixed M1/M2 bugs, but
	// discovered outside this task's scope. See PERF-FOLLOWUP-08 for the hoist plan.
	n1AllowlistKey("SyncLTVTopPercent", "tagCacheRepo.FindByOwner"): true,
	// buildStaffSlotInputs loops over ALL clinic staff and re-fetches schedule/breaks per
	// staff. Lower severity than the owner-scale finding above (staff count is small/bounded),
	// but still data-bound. See PERF-FOLLOWUP-08.
	n1AllowlistKey("buildStaffSlotInputs", "scheduleRepo.FindAllByDate"):          true,
	n1AllowlistKey("buildStaffSlotInputs", "scheduleRepo.FindAllBreaksByEntryID"): true,
}

// analyzeFileN1 parses one Go source file and reports RangeStmt-body Find*/Get* calls not
// covered by n1Allowlist, plus which allowlist keys were actually hit (for occurrence
// pinning — see TestN1Lint_AllowlistEntriesAreLive). Pure function over (filename, src) so the
// real embedded source and the inline self-verification fixtures exercise identical logic.
func analyzeFileN1(filename string, src []byte) ([]n1Finding, map[string]int, int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, nil, 0, err
	}

	var findings []n1Finding
	allowHits := make(map[string]int)
	rangeLoopsSeen := 0
	base := baseNameN1(filename)

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		fnName := fd.Name.Name

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			rs, ok := n.(*ast.RangeStmt)
			if !ok {
				return true
			}
			rangeLoopsSeen++

			ast.Inspect(rs.Body, func(inner ast.Node) bool {
				ce, ok := inner.(*ast.CallExpr)
				if !ok {
					return true
				}
				callee, matched := matchN1Call(ce)
				if !matched {
					return true
				}
				key := n1AllowlistKey(fnName, callee)
				if n1Allowlist[key] {
					allowHits[key]++
					return true
				}
				findings = append(findings, n1Finding{
					file:   base,
					line:   fset.Position(ce.Pos()).Line,
					fn:     fnName,
					callee: callee,
				})
				return true
			})

			// Do not let the outer Inspect separately re-visit this RangeStmt's descendants
			// as independent top-level RangeStmt nodes — the inner Inspect above already
			// covers any nested loop/closure within this body. Prevents double-reporting the
			// same call site if a RangeStmt is nested inside another RangeStmt.
			return false
		})
	}

	return findings, allowHits, rangeLoopsSeen, nil
}

// matchN1Call reports whether ce is a Pattern-1 (`<x>Repo.Find*`) or Pattern-2
// (`settingsSvc.Get*`) call, returning "<field>.<Method>" for allowlist/diagnostic use.
func matchN1Call(ce *ast.CallExpr) (callee string, matched bool) {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	recv, ok := sel.X.(*ast.SelectorExpr)
	if !ok {
		return "", false // not a `s.xxxRepo.Method(...)` / `s.settingsSvc.Method(...)` shape
	}

	switch {
	case strings.HasSuffix(recv.Sel.Name, "Repo") && strings.HasPrefix(sel.Sel.Name, "Find"):
		return recv.Sel.Name + "." + sel.Sel.Name, true
	case recv.Sel.Name == "settingsSvc" && strings.HasPrefix(sel.Sel.Name, "Get"):
		return recv.Sel.Name + "." + sel.Sel.Name, true
	default:
		return "", false
	}
}

func baseNameN1(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// walkServiceN1 runs the analyzer over every non-test .go file embedded from this package
// directory (reusing master_fk_write_inventory_lint_test.go's serviceSourceFS) and aggregates
// findings + allowlist hit counts + the total RangeStmt count (a floor guard against a
// vacuously-green embed/AST break).
func walkServiceN1(t *testing.T) ([]n1Finding, map[string]int, int) {
	t.Helper()

	names, err := fs.Glob(serviceSourceFS, "*.go")
	if err != nil {
		t.Fatalf("glob embedded service source: %v", err)
	}

	var all []n1Finding
	agg := make(map[string]int)
	totalRangeLoops := 0
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := serviceSourceFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", name, err)
		}
		findings, allowHits, rangeLoops, err := analyzeFileN1(name, src)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		all = append(all, findings...)
		for k, v := range allowHits {
			agg[k] += v
		}
		totalRangeLoops += rangeLoops
	}
	return all, agg, totalRangeLoops
}

// TestN1Lint_RealServiceSourceHasNoUnresolvedLoopBodyFindOrGet is the gate: no RangeStmt body
// in package service may call a `<x>Repo.Find*` or `settingsSvc.Get*` method unless the exact
// (function, callee) is on n1Allowlist with a documented reason. The floor guard prevents a
// vacuous green if the embed glob or AST matching silently breaks.
//
// See the "KNOWN REAL FINDINGS, TRACKED NOT SILENCED" comment at the top of this file: two
// pre-existing N+1s discovered by this gate's first run are allowlisted with a reference to
// docs/tasks/open/PERF-FOLLOWUP-08-ltv-staffslots-n1.md, not fixed here — fixing them was out
// of scope for the task (PERF-FOLLOWUP-07) that added this test-only gate.
func TestN1Lint_RealServiceSourceHasNoUnresolvedLoopBodyFindOrGet(t *testing.T) {
	findings, _, rangeLoops := walkServiceN1(t)

	if rangeLoops < 200 {
		t.Fatalf("only %d RangeStmt loops parsed across package service; embed glob or AST walk likely broken (would vacuously pass)", rangeLoops)
	}

	for _, f := range findings {
		t.Errorf("N+1 candidate: %s:%d func %s(): %q called inside a for-range loop body. "+
			"Hoist the call above the loop (see PERF-1/PERF-2), or if it is a genuine "+
			"request-bound (not data-bound) per-iteration lookup, add a documented entry to n1Allowlist.",
			f.file, f.line, f.fn, f.callee)
	}
}

// TestN1Lint_Analyzer pins the detection behaviour on inline fixtures: the M1/M2 regression
// shape must be detected (self-verification / "never vacuously green"), and known-good /
// out-of-scope / pagination-loop shapes must not be flagged.
func TestN1Lint_Analyzer(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "repo Find call inside range loop body is flagged (M1/M2 regression shape)",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64, owners []Owner) {
	for _, o := range owners {
		s.settingsRepo.FindByClinicID(ctx, clinicID)
	}
}`,
			want: 1,
		},
		{
			name: "settingsSvc Get call inside range loop body is flagged (PERF-1/PERF-2 regression shape)",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64, owners []Owner) {
	for _, o := range owners {
		s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	}
}`,
			want: 1,
		},
		{
			name: "hoisted repo Find call above the loop is not flagged (the actual M1/M2 fix)",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64, owners []Owner) {
	thresholds, _ := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	for _, o := range owners {
		_ = thresholds
		_ = o
	}
}`,
			want: 0,
		},
		{
			name: "pagination for{} loop is not a RangeStmt, not flagged (FOLLOWUP-02 shape)",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64) {
	page := uint64(0)
	for {
		batch, err := s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, page, 500)
		if err != nil || len(batch) == 0 {
			break
		}
		page = batch[len(batch)-1].ID
	}
}`,
			want: 0,
		},
		{
			name: "call through a closure field is out of scope (syntactic-only, matches SyncHealthPreventionTagsForClinic shape)",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64, owners []Owner) {
	for i := range owners {
		fn := func() error { return s.checkupRepo.FindByOwnerID(ctx, clinicID, owners[i].ID) }
		_ = fn()
	}
}`,
			want: 1, // the inner Inspect DOES see it lexically inside the range body (documented scope)
		},
		{
			name: "non-repo, non-settingsSvc method call is ignored",
			src: `package p
func (s *xService) f(ctx context.Context, owners []Owner) {
	for _, o := range owners {
		s.auditSvc.LogSomething(ctx, o.ID)
	}
}`,
			want: 0,
		},
		{
			name: "repo method not in Find family is ignored",
			src: `package p
func (s *xService) f(ctx context.Context, clinicID uint64, owners []Owner) {
	for _, o := range owners {
		s.tagCacheRepo.UpsertTag(ctx, clinicID, o.ID, "tag", "manual", "")
	}
}`,
			want: 0,
		},
		{
			name: "allowlisted request-bound ownership validation loop is not flagged",
			src: `package p
func (s *xService) validateOwnerPetsInsuranceOwnership(ctx context.Context, clinicID uint64, pets []CreatePetForOwnerInput) error {
	for i := range pets {
		s.insuranceRepo.FindByID(ctx, clinicID, *pets[i].InsuranceID)
	}
	return nil
}`,
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, _, _, err := analyzeFileN1("fixture.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

// TestN1Lint_AllowlistEntriesAreLive proves every allowlist entry is exercised against the
// REAL source (not dead code / stale after a rename) and pins each to its exact expected
// occurrence count, the same bidirectional-exhaustiveness discipline as
// preloadClinicScopeSiteExceptions in preload_clinic_scope_lint_test.go: a new unallowlisted
// occurrence of the same (func, callee) pair fails closed instead of silently inheriting the
// waiver, and a removed/renamed site forces deleting the now-stale entry.
func TestN1Lint_AllowlistEntriesAreLive(t *testing.T) {
	_, allowHits, _ := walkServiceN1(t)

	wantOccurrences := map[string]int{
		n1AllowlistKey("validateOwnerPetsInsuranceOwnership", "insuranceRepo.FindByID"): 1,
		n1AllowlistKey("ValidateAndCreate", "trimmingOptionRepo.FindByID"):              1,
		n1AllowlistKey("SyncLTVTopPercent", "tagCacheRepo.FindByOwner"):                 1,
		n1AllowlistKey("buildStaffSlotInputs", "scheduleRepo.FindAllByDate"):            1,
		n1AllowlistKey("buildStaffSlotInputs", "scheduleRepo.FindAllBreaksByEntryID"):   1,
	}
	if len(n1Allowlist) != len(wantOccurrences) {
		t.Fatalf("n1Allowlist has %d entries but this test pins %d; keep both lists in sync", len(n1Allowlist), len(wantOccurrences))
	}
	for key, want := range wantOccurrences {
		got := allowHits[key]
		if got != want {
			t.Errorf("allowlist entry %q matched %d call site(s), expected %d. "+
				"If a new occurrence was added, verify it is still request-bound (not data-bound) "+
				"before inheriting the waiver; if the site was fixed/removed, delete the stale entry.",
				key, got, want)
		}
	}
}
