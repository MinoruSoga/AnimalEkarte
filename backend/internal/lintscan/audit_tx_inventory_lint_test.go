package lintscan

// audit_tx_inventory_lint_test.go — Clinical-result hard-delete audit-tx inventory gate
//
// Background (#211): patient clinical result rows (exam_results, checkup_field_results)
// are hard-deleted (no gorm.DeletedAt) during a replace operation. If the post-replace
// audit write executes OUTSIDE the delete transaction, a subsequent audit failure cannot
// roll back the already-committed deletion — the patient result is gone with no audit
// trail. This is a tamper-traceability gap in patient data.
//
// checkup_field_results was migrated to "audited-tx-internal" (fe04b460): ReplaceForCheckup
// uses dbOrTx(ctx, r.db).Transaction(...) so the delete participates in the caller's ambient
// tx; if the post-replace audit write fails, the caller's tx rollback also reverts the delete
// (fail-closed). The runtime proof is checkup_field_result_tx_atomicity_test.go.
//
// exam_results was migrated to "audited-tx-internal" (BE-refactor.md R1-2): ReplaceItemsByExamID
// now uses dbOrTx(ctx, r.db).Transaction(...) so the delete participates in the caller's ambient
// tx; examinationService.ReplaceItems wraps it in Transactor.WithTx and writes a post-replace
// AuditTxLogger.LogEntryTx audit entry (gated on deletedCount > 0) in the same tx — an audit
// failure rolls back the delete (fail-closed). The runtime proof is
// examination_repository_tx_atomicity_test.go.
//
// ─── Purpose ────────────────────────────────────────────────────────────────────────
//
// This gate enforces INVENTORY COVERAGE, not correctness:
//  - It lists every repository function that hard-deletes a clinical-result row.
//  - Each site must be on a curated allowlist with a status annotation.
//  - A NEW hard-delete site that is not on the allowlist FAILS CI → human review required.
//  - A stale allowlist entry (site renamed/removed) also FAILS CI.
//
// Static proof that "the delete is actually inside the ambient tx" is NOT attempted —
// that requires interprocedural taint analysis across layers, which go/ast cannot do
// reliably (#124 lesson). The ambient-tx property is proven at runtime by atomicity tests
// (checkup_field_result_tx_atomicity_test.go; exam_results must add one on migration).
//
// ─── Technique ──────────────────────────────────────────────────────────────────────
//
// Discovery is delegated to moduleInternalSource / legacyLintKey (both declared in
// preload_clinic_scope_lint_test.go; BE9-1), which source from the shared internal/lintscan
// package instead of a package-local go:embed. This scans every production .go file under the
// WHOLE module's internal/ tree, not only internal/repository — see
// preload_clinic_scope_lint_test.go's file header ("Discovery mechanism (BE9-1)") for the full
// key-format-compatibility rationale. AST walker detects .Delete(&model.X{}) calls where X is
// in clinicalResultHardDeleteModels. Key = (file, receiverType.Method, modelType) + exact
// occurrence count — same technique as preload site exceptions and master-FK lint.
//
// ─── AST detection blindspot ────────────────────────────────────────────────────────
//
// The matcher captures composite-literal form ONLY: .Delete(&model.X{}).
// The following forms are NOT detected and will silently bypass this gate:
//   (a) raw SQL: db.Exec("DELETE FROM exam_results ...")
//   (b) variable form: .Delete(&sliceVar) or .Delete(&itemVar)
//   (c) association clear: Association("Items").Unscoped().Clear()
// Both current sites use composite-literal form. When migrating exam_results, preserve
// this form (or extend the matcher) to ensure the gate continues to fire.
//
// More generally (BE9-1): a Go model type not registered in clinicalResultHardDeleteModels is
// invisible to this gate even in composite-literal form, and any hard-delete reachable only
// through a background job, cron, or worker path that never appears as a literal AST call shape
// inside a plain .go file this scanner visits is also invisible. The runtime atomicity tests
// (*_tx_atomicity_test.go) and the cross-tenant *_clinic_isolation_test.go-style guards remain
// required for these blind spots; this static inventory gate is a complement layered on top of
// them, not a substitute.
//
// ─── Intentional exclusions ─────────────────────────────────────────────────────────
//
// Models with gorm.DeletedAt (soft-delete) → excluded from the inventory:
//   - VitalRecord    (vital_records):   vital.go:32 DeletedAt gorm.DeletedAt gorm:"index"
//   - Examination    (exams):           examination_record.go:42 DeletedAt gorm.DeletedAt
// Models that are NOT clinical-result-value tables → excluded by charter:
//   - CarePlanItem   (care_plan_items): clinical care plan entries; no DeletedAt BUT
//                                        this gate covers result-value tables only.
//                                        Separate audit coverage tracked as follow-up.
//   - MedicalRecordImage: attachment metadata, not a structured result value.
//   - CareLog / StaffNote / DailyRecord: no delete path exists as of 2026-06-30.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// clinicalResultHardDeleteModels maps a Go model type name to its DB table (documentation-only).
// Only models WITHOUT gorm.DeletedAt are listed — these perform hard-deletes.
// Models with gorm.DeletedAt (e.g. VitalRecord) use soft-delete and are out of scope.
// NOTE: only the KEYS are checked programmatically; the values (table names) are annotations
// for human readers — a table rename would not be caught here.
var clinicalResultHardDeleteModels = map[string]string{
	"ExamResult":         "exam_results",          // no DeletedAt field → hard-delete
	"CheckupFieldResult": "checkup_field_results", // no DeletedAt field → hard-delete
}

// auditInventoryStatus records WHY a hard-delete site is on the allowlist.
// The gate does not verify this — it is the human review record.
type auditInventoryStatus string

const (
	// statusAuditedTxInternal: the delete uses dbOrTx(ctx, r.db).Transaction(...) —
	// ambient-tx-aware; if the caller's ambient tx rolls back, the delete is also reverted.
	// Proven at runtime by a *_tx_atomicity_test.go test.
	statusAuditedTxInternal auditInventoryStatus = "audited-tx-internal"

	// statusPendingMigration: the delete uses r.db.WithContext(ctx).Transaction(...) —
	// NOT ambient-tx-aware; delete commits before audit, so audit failure cannot roll it back.
	// Migration target: switch to dbOrTx + add a runtime atomicity test.
	statusPendingMigration auditInventoryStatus = "pending-migration"

	// statusDocumentedException: not a patient-result hard-delete in the clinical sense,
	// or validated by a documented alternative mechanism. Must justify in reason field.
	statusDocumentedException auditInventoryStatus = "documented-exception"
)

// auditInventoryEntry is one allowlist row.
//
//   - file:        base filename (e.g. "checkup_field_repository.go")
//   - function:    "ReceiverType.MethodName" for methods, "FuncName" for free functions
//   - modelType:   Go model struct name (must be in clinicalResultHardDeleteModels)
//   - occurrences: EXACT count of .Delete(&model.ModelType{}) calls in this function.
//     Count going UP → new unreviewed delete call, forced into review (cannot silently
//     inherit the waiver). Count going DOWN → stale entry must be removed.
//   - status / reason: human review record; gate does not verify correctness.
type auditInventoryEntry struct {
	file        string
	function    string
	modelType   string
	occurrences int
	status      auditInventoryStatus
	reason      string
}

// clinicalResultAuditTxAllowlist records every enumerated clinical-result hard-delete
// site with a reviewed status. Sorted by file for review diffs.
//
//	audited-tx-internal = delete wraps in dbOrTx(ctx, r.db).Transaction(...)
//	                      (ambient-tx-aware). Runtime proof: *_tx_atomicity_test.go.
//	pending-migration   = delete wraps in r.db.WithContext(ctx).Transaction(...)
//	                      (NOT ambient-tx-aware). Migration target.
var clinicalResultAuditTxAllowlist = []auditInventoryEntry{
	{
		file:        "medicalrecord/checkup_field_repository.go",
		function:    "checkupFieldResultRepository.ReplaceForCheckup",
		modelType:   "CheckupFieldResult",
		occurrences: 1,
		status:      statusAuditedTxInternal,
		reason: "medicalrecord/checkup_field_repository.go: ReplaceForCheckup uses dbOrTx(ctx, r.db).Transaction(...) — " +
			"ambient-tx-aware. The delete participates in the caller's Transactor.WithTx ambient tx; " +
			"if the post-replace audit write (auditRepository.CreateTx) fails and the caller rolls back, " +
			"the delete is also reverted (fail-closed #211). " +
			"Runtime proof: checkup_field_result_tx_atomicity_test.go (fe04b460).",
	},
	{
		file:        "medicalrecord/examination_repository.go",
		function:    "examinationRepository.ReplaceItemsByExamID",
		modelType:   "ExamResult",
		occurrences: 1,
		status:      statusAuditedTxInternal,
		reason: "internal/medicalrecord/examination_repository.go (BE9-2D ⑦, moved): ReplaceItemsByExamID uses dbOrTx(ctx, r.db).Transaction(...) — " +
			"ambient-tx-aware (BE-refactor.md R1-2). The delete participates in the caller's " +
			"Transactor.WithTx ambient tx; examinationService.ReplaceItems writes a post-replace " +
			"AuditTxLogger.LogEntryTx entry (gated on deletedCount > 0) in the same tx, and if it fails " +
			"and the caller rolls back, the delete is also reverted (fail-closed). " +
			"Runtime proof: examination_repository_tx_atomicity_test.go.",
	},
	{
		file:        "medicalrecord/examination_revision_workflow_repository.go",
		function:    "restoreMutableExaminationFromRevision",
		modelType:   "ExamResult",
		occurrences: 1,
		status:      statusAuditedTxInternal,
		reason: "Free function restoreMutableExaminationFromRevision hard-deletes ExamResult once (examination_revision_workflow_repository.go:486-497) " +
			"via the injected ambient tx parameter (tx *gorm.DB). Both callers AppendWorkingRevisionFromOfficial " +
			"(lines 21-24, 66-77) and AppendOfficialRevisionFromWorking (lines 160-163, 206-216) fail-closed when " +
			"persistence.TxFromContext(ctx) is nil, so the delete never runs outside an ambient transaction. " +
			"Service layer examinationService.Unconfirm (examination_revision_service.go:129-184) and " +
			"reconfirmRevisionTx (lines 250-310, invoked from examination_service.go:342 inside WithTx) write " +
			"parent-mutation audit via logParentMutationWithReasonTx / logParentMutationTx → AuditTxLogger.LogEntryTx " +
			"in the same ambient tx after the restore; audit failure returns error and rolls back the delete " +
			"(fail-closed). Official revision items remain append-only; this delete only clears the mutable " +
			"working projection before restore. Runtime rollback proof: examination_revision_workflow_safety_test.go.",
	},
}

// ─── Analyzer (pure over (filename, src), same as preload lint) ─────────────────────

type auditInventoryFinding struct {
	file      string
	line      int
	function  string
	modelType string
}

// analyzeFileForClinicalResultDeletes parses one Go source file and reports every
// .Delete(&model.X{}) call whose X is in clinicalResultHardDeleteModels.
// It is a pure function so fixtures and the real discovered source exercise the same logic.
func analyzeFileForClinicalResultDeletes(filename string, src []byte) ([]auditInventoryFinding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, err
	}

	var findings []auditInventoryFinding
	// Root files key by basename; 1-level subpackage files keep the relative path so e.g.
	// paymentmethod/repository.go and vaccine/repository.go do not share an allowlist identity
	// (BE-refactor.md §5-1-3; mirrors dbortx_inventory_lint_test.go's keyFile logic).
	base := baseFileName(filename) // helper from preload_clinic_scope_lint_test.go
	if strings.Contains(filename, "/") {
		base = filename
	}

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		funcKey := receiverMethodKey(fd)

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := ce.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Delete" || len(ce.Args) == 0 {
				return true
			}
			modelType := clinicalResultModelFromArg(ce.Args[0])
			if modelType == "" {
				return true
			}
			findings = append(findings, auditInventoryFinding{
				file:      base,
				line:      fset.Position(ce.Pos()).Line,
				function:  funcKey,
				modelType: modelType,
			})
			return true
		})
	}
	return findings, nil
}

// clinicalResultModelFromArg returns the model type name if expr is &model.X{} where
// X ∈ clinicalResultHardDeleteModels, otherwise "".
func clinicalResultModelFromArg(e ast.Expr) string {
	ue, ok := e.(*ast.UnaryExpr)
	if !ok || ue.Op != token.AND {
		return ""
	}
	cl, ok := ue.X.(*ast.CompositeLit)
	if !ok {
		return ""
	}
	se, ok := cl.Type.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	pkg, ok := se.X.(*ast.Ident)
	if !ok || pkg.Name != "model" {
		return ""
	}
	if _, isClinical := clinicalResultHardDeleteModels[se.Sel.Name]; !isClinical {
		return ""
	}
	return se.Sel.Name
}

// receiverMethodKey returns "ReceiverType.MethodName" for methods, "FuncName" for free functions.
func receiverMethodKey(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recvType := fd.Recv.List[0].Type
	var recvName string
	switch t := recvType.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			recvName = id.Name
		}
	case *ast.Ident:
		recvName = t.Name
	}
	if recvName == "" {
		return fd.Name.Name
	}
	return recvName + "." + fd.Name.Name
}

// auditInventoryKey returns the unique lookup key for a finding/entry.
func auditInventoryKey(file, function, modelType string) string {
	return file + "|" + function + "|" + modelType
}

// walkRepositoryForClinicalResultDeletes runs the analyzer over every production .go file
// discovered module-wide (moduleInternalSource; BE9-1) — internal/repository and every other
// internal/ package. Filenames fed to analyzeFileForClinicalResultDeletes are
// legacyLintKey-normalized so allowlist identity for internal/repository/** files is unchanged
// from the pre-BE9-1 repoSourceFS walk.
func walkRepositoryForClinicalResultDeletes(t *testing.T) []auditInventoryFinding {
	t.Helper()
	tree := moduleInternalSource(t)
	var all []auditInventoryFinding
	for rawKey, src := range tree {
		key := legacyLintKey(rawKey)
		// analyzeFileForClinicalResultDeletes derives its own key from the filename argument:
		// basename for root package files, relative path for subpackage files
		// (BE-refactor.md §5-1-3).
		findings, err := analyzeFileForClinicalResultDeletes(key, src)
		if err != nil {
			t.Fatalf("parse %s: %v", key, err)
		}
		all = append(all, findings...)
	}
	return all
}

// aggregateClinicalResultFindings groups findings by key (file|function|modelType) → count.
func aggregateClinicalResultFindings(findings []auditInventoryFinding) map[string]int {
	agg := map[string]int{}
	for _, f := range findings {
		agg[auditInventoryKey(f.file, f.function, f.modelType)]++
	}
	return agg
}

// ─── Reconciliation (pure, same bidirectional approach as master-FK lint) ──────────

// reconcileClinicalResultDeletes is the pure bidirectional check, separated so its
// three failure modes can be frozen by fixture tests (not only transiently during RED).
// Returns human-readable violations (empty = clean).
func reconcileClinicalResultDeletes(found map[string]int, allowlist []auditInventoryEntry) []string {
	var violations []string

	allow := map[string]auditInventoryEntry{}
	for _, e := range allowlist {
		key := auditInventoryKey(e.file, e.function, e.modelType)
		isDup := false
		if _, exists := allow[key]; exists {
			violations = append(violations, "duplicate allowlist entry "+key)
			isDup = true
		}
		if e.status != statusAuditedTxInternal && e.status != statusPendingMigration && e.status != statusDocumentedException {
			violations = append(violations, "allowlist entry "+key+" has invalid status "+string(e.status))
		}
		if !isDup {
			allow[key] = e
		}
	}

	// (i) every detected site must be allowlisted with matching occurrence count.
	for key, cnt := range found {
		e, ok := allow[key]
		if !ok {
			violations = append(violations,
				"clinical-result hard-delete site "+key+" (count="+strconv.Itoa(cnt)+") is NOT on the "+
					"allowlist. Add an entry with a status (audited-tx-internal/pending-migration/"+
					"documented-exception) + reason. If this is a new site, also add a runtime atomicity "+
					"test analogous to checkup_field_result_tx_atomicity_test.go.")
			continue
		}
		if cnt != e.occurrences {
			violations = append(violations,
				"clinical-result hard-delete site "+key+" has "+strconv.Itoa(cnt)+" Delete call(s), "+
					"allowlist pins "+strconv.Itoa(e.occurrences)+". A Delete call was added/removed "+
					"within this function — re-review and update the occurrences field.")
		}
	}
	// (ii) no stale allowlist entry.
	for key, e := range allow {
		if _, ok := found[key]; !ok {
			violations = append(violations,
				"allowlist entry {"+e.file+", "+e.function+", "+e.modelType+"} no longer matches any "+
					"clinical-result Delete call (function renamed/removed, or model changed). "+
					"Delete the stale entry.")
		}
	}
	return violations
}

// TestClinicalResultAuditTxInventory_AnalyzerDetectsViolationUnderNestedPathFilename proves the
// analyzer itself (not just discovery reach — see
// TestClinicalResultAuditTxInventory_DiscoveryReachesModuleWideAndNestedPackages) flags a
// non-allowlisted clinical-result hard-delete identically regardless of the filename/location
// shape the source is presented under: a repository-root-shaped filename, a DIFFERENT top-level
// internal/ package filename, and a 2+-level nested filename must all report the SAME violation
// (BE9-1). BE8-0 follow-up: the prior walk-smoke test never fed a violation through
// analyzeFileForClinicalResultDeletes + reconcileClinicalResultDeletes together (see
// BE-refactor.md BE8-0 follow-up note).
func TestClinicalResultAuditTxInventory_AnalyzerDetectsViolationUnderNestedPathFilename(t *testing.T) {
	src := "package p\nfunc (r *repository) Upsert() { db.Delete(&model.ExamResult{}) }\n"
	filenames := []string{
		"repository/x.go",          // root-shaped
		"service/x.go",             // a different top-level internal/ package
		"repository/two/deep/x.go", // 2+-level nested
	}
	for _, fn := range filenames {
		t.Run(fn, func(t *testing.T) {
			findings, err := analyzeFileForClinicalResultDeletes(fn, []byte(src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("got %d findings for filename %q, want 1: %+v", len(findings), fn, findings)
			}
			agg := aggregateClinicalResultFindings(findings)
			// Empty allowlist: the site is unreviewed, so it must be flagged (not silently accepted).
			violations := reconcileClinicalResultDeletes(agg, nil)
			if len(violations) == 0 {
				t.Fatalf("an unreviewed clinical-result hard-delete under filename %q was not flagged against an empty allowlist", fn)
			}
		})
	}
}

// TestClinicalResultAuditTxInventory_AllowlistDoesNotCrossSubpackageBasenameCollision proves an
// allowlist entry scoped to one 1-level domain subpackage's file cannot silently cover the same
// hard-delete pattern in a DIFFERENT subpackage sharing the same basename (all 14 domain
// subpackages currently have a "repository.go" — BE-refactor.md §5-1-3). The entry below
// simulates one that was (or would be) registered under the bare basename a pre-fix
// baseFileName(name) computes for ANY nested file — e.g. it was reviewed and intended for
// paymentmethod/repository.go only. A different subpackage (vaccine/) with an identical
// unreviewed ExamResult hard-delete must NOT inherit that review.
func TestClinicalResultAuditTxInventory_AllowlistDoesNotCrossSubpackageBasenameCollision(t *testing.T) {
	allowlist := []auditInventoryEntry{
		{
			file:        "repository.go",
			function:    "repository.Upsert",
			modelType:   "ExamResult",
			occurrences: 1,
			status:      statusAuditedTxInternal,
			reason:      "BE8-0 collision-safety fixture: an allowlist entry intentionally scoped to one specific domain subpackage file only",
		},
	}
	src := "package p\nfunc (r *repository) Upsert() { db.Delete(&model.ExamResult{}) }\n"
	findings, err := analyzeFileForClinicalResultDeletes("vaccine/repository.go", []byte(src))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	agg := aggregateClinicalResultFindings(findings)
	violations := reconcileClinicalResultDeletes(agg, allowlist)
	if len(violations) == 0 {
		t.Fatal("vaccine/repository.go's ExamResult hard-delete was silently matched by an allowlist entry " +
			"keyed under the bare basename \"repository.go\" — this is the cross-subpackage basename collision " +
			"BE-refactor.md §5-1-3 warns about: paymentmethod/repository.go and vaccine/repository.go must not " +
			"share an allowlist identity")
	}
}

// ─── Gate tests ─────────────────────────────────────────────────────────────────────

// TestClinicalResultAuditTxInventory_AllowlistMatchesRealSource is the gate: every
// clinical-result hard-delete site in this repository package must be on the allowlist
// with a matching occurrence count, and every allowlist entry must still match a real site.
// The floor prevents a vacuous green if discovery or the AST matching silently breaks.
func TestClinicalResultAuditTxInventory_AllowlistMatchesRealSource(t *testing.T) {
	findings := walkRepositoryForClinicalResultDeletes(t)

	// Floor: at least 2 sites (ExamResult + CheckupFieldResult). A broken AST matcher
	// or broken discovery would drop this to 0 — far below the floor.
	if len(findings) < 2 {
		t.Fatalf("only %d clinical-result hard-delete site(s) found; AST matching or "+
			"discovery likely broken (would vacuously pass). Expected ≥2 (ExamResult + CheckupFieldResult).",
			len(findings))
	}

	agg := aggregateClinicalResultFindings(findings)
	for _, v := range reconcileClinicalResultDeletes(agg, clinicalResultAuditTxAllowlist) {
		t.Error(v)
	}
}

// TestClinicalResultAuditTxInventory_StatusesAreLive proves the audited-tx-internal status is
// exercised by a real entry. Both known clinical-result hard-delete sites (checkup_field_results
// fe04b460, exam_results BE-refactor.md R1-2) are migrated as of this test — there is currently
// no 'pending-migration' example. statusPendingMigration itself stays in the taxonomy (see the
// const block above and TestClinicalResultAuditTxInventory_GateDetectsViolations fixtures) for
// the next clinical-result hard-delete site that is added but not yet made ambient-tx-aware.
func TestClinicalResultAuditTxInventory_StatusesAreLive(t *testing.T) {
	counts := map[auditInventoryStatus]int{}
	for _, e := range clinicalResultAuditTxAllowlist {
		counts[e.status]++
	}
	if counts[statusAuditedTxInternal] == 0 {
		t.Error("no 'audited-tx-internal' entries; checkup_field_results (fe04b460) / exam_results " +
			"(R1-2) migrations may have drifted or been removed — verify before deleting")
	}
	if counts[statusAuditedTxInternal] != len(clinicalResultAuditTxAllowlist) {
		t.Error("a non-audited-tx-internal entry exists; if this is a genuine new pending-migration " +
			"site that is expected, this assertion should be relaxed with a comment explaining why")
	}
}

// TestClinicalResultAuditTxInventory_DiscoveryReachesModuleWideAndNestedPackages pins that the
// module-wide discovery set (moduleInternalSource, backed by lintscan.WalkInternalTreeT; BE9-1)
// walkRepositoryForClinicalResultDeletes iterates over reaches: (a) a real 2+-level production
// package, (b) at least one file from a DIFFERENT top-level internal/ package, and (c) arbitrary
// deeper nesting (scanner capability, proven via a synthetic tree — see
// assertLintscanReachesTwoOrMoreNestingLevels). Also smoke-exercises the actual production walk
// and confirms every reached nested package file's raw source is independently parseable
// via analyzeFileForClinicalResultDeletes.
//
// Renamed + strengthened from the pre-BE9-1
// TestClinicalResultAuditTxInventory_WalksAllEmbeddedFilesIncludingSubpackages, which only pinned
// the go:embed glob's 1-level reach within internal/repository. NOTE (audit gap, recorded
// 2026-07-17, still true post-BE9-1): clinicalResultAuditTxAllowlist currently has zero
// subpackage entries, so this test cannot assert that walkRepositoryForClinicalResultDeletes's
// own internals would still visit a subpackage file if a future edit added a silent path-based
// filter inside that function — it can only prove the discovery reaches subpackages (and other
// packages, and depth) and their source is parseable today.
func TestClinicalResultAuditTxInventory_DiscoveryReachesModuleWideAndNestedPackages(t *testing.T) {
	tree := moduleInternalSource(t)
	if _, ok := tree["infra/smtp/sender.go"]; !ok {
		t.Fatal("module-wide discovery does not include infra/smtp/sender.go; " +
			"walkRepositoryForClinicalResultDeletes may have narrowed and would silently drop nested production packages")
	}
	// Smoke-exercise the actual production walk against the current discovered set — proves it
	// does not t.Fatal/panic while nested and top-level target-package files are present.
	_ = walkRepositoryForClinicalResultDeletes(t)
	for rawKey, src := range tree {
		if strings.Count(rawKey, "/") < 2 {
			continue
		}
		if _, err := analyzeFileForClinicalResultDeletes(legacyLintKey(rawKey), src); err != nil {
			t.Fatalf("analyzeFileForClinicalResultDeletes failed to parse module-wide nested package file %s: %v — "+
				"this lint's walk would silently skip it", rawKey, err)
		}
	}

	assertDiscoversFileFromDifferentTopLevelPackage(t, tree)
	assertLintscanReachesTwoOrMoreNestingLevels(t)
}

// TestClinicalResultAuditTxInventory_Analyzer pins detection on inline fixtures (SC3 non-vacuous
// proof): known clinical-result Delete calls must be detected, and non-clinical/soft-delete
// models must not. This both drives the analyzer (RED→GREEN) and freezes all failure modes.
func TestClinicalResultAuditTxInventory_Analyzer(t *testing.T) {
	cases := []struct {
		name      string
		src       string
		wantCount int
	}{
		{
			name: "ExamResult hard-delete detected",
			src: `package p
func (r *examinationRepository) ReplaceItemsByExamID() {
	_ = tx.Where("exam_id = ?", 1).Delete(&model.ExamResult{}).Error
}`,
			wantCount: 1,
		},
		{
			name: "CheckupFieldResult hard-delete detected (chained assignment form)",
			src: `package p
func (r *checkupFieldResultRepository) ReplaceForCheckup() {
	del := tx.Where("checkup_id = ? AND clinic_id = ?", 1, 2).Delete(&model.CheckupFieldResult{})
	_ = del.Error
}`,
			wantCount: 1,
		},
		{
			name: "two ExamResult deletes in one function counted separately (occurrence pin)",
			src: `package p
func (r *examinationRepository) DoubleDelete() {
	tx.Where("exam_id = ?", 1).Delete(&model.ExamResult{})
	tx.Where("exam_id = ?", 2).Delete(&model.ExamResult{})
}`,
			wantCount: 2,
		},
		{
			name: "Vaccine (soft-delete model with gorm.DeletedAt) not detected",
			src: `package p
func (r *vaccineRepository) Delete() {
	db.Scopes(clinicScope(1)).Where("id = ?", 5).Delete(&model.Vaccine{})
}`,
			wantCount: 0,
		},
		{
			name: "VitalRecord (soft-delete) not detected",
			src: `package p
func (r *vitalRepository) Delete() {
	db.Where("id = ?", 1).Delete(&model.VitalRecord{})
}`,
			wantCount: 0,
		},
		{
			name: "non-model.X form (no model. qualifier) not detected",
			src: `package p
func (r *someRepo) BrokenDelete() {
	db.Where("id = ?", 1).Delete(&ExamResult{})
}`,
			wantCount: 0,
		},
		{
			name: "wrong package qualifier not detected",
			src: `package p
func (r *someRepo) WrongPkg() {
	db.Where("id = ?", 1).Delete(&other.ExamResult{})
}`,
			wantCount: 0,
		},
		{
			name: "free function (no receiver) is also captured",
			src: `package p
func purgeExamResults() {
	db.Where("exam_id = ?", 1).Delete(&model.ExamResult{})
}`,
			wantCount: 1,
		},
		{
			// M-1 (go-reviewer): addr-of a variable — not a composite literal → not detected.
			// Proves clinicalResultModelFromArg correctly rejects non-CompositeLit arguments.
			name: "addr-of variable (non-literal) not detected",
			src: `package p
func (r *examinationRepository) BulkDelete(result model.ExamResult) {
	db.Where("id = ?", 1).Delete(&result)
}`,
			wantCount: 0,
		},
		{
			// M-2 (go-reviewer): value receiver (non-pointer) exercised so the
			// *ast.Ident branch in receiverMethodKey is covered by tests.
			name: "value receiver (non-pointer) is also captured",
			src: `package p
func (r examinationRepository) CleanupExamResults() {
	db.Where("exam_id = ?", 1).Delete(&model.ExamResult{})
}`,
			wantCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, err := analyzeFileForClinicalResultDeletes("fixture.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.wantCount {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.wantCount, findings)
			}
		})
	}
}

// TestClinicalResultAuditTxInventory_GateDetectsViolations freezes the gate's failure
// modes: (a) new unlisted site FAILS, (b) stale allowlist entry FAILS, (c) occurrence
// count mismatch FAILS, (d) duplicate key FAILS. The clean baseline must report zero.
func TestClinicalResultAuditTxInventory_GateDetectsViolations(t *testing.T) {
	baseAllowlist := []auditInventoryEntry{
		{
			file: "examination_repository.go", function: "examinationRepository.ReplaceItemsByExamID",
			modelType: "ExamResult", occurrences: 1, status: statusPendingMigration, reason: "fixture",
		},
	}
	baseFound := map[string]int{
		auditInventoryKey("examination_repository.go", "examinationRepository.ReplaceItemsByExamID", "ExamResult"): 1,
	}

	t.Run("clean baseline reports no violation", func(t *testing.T) {
		got := reconcileClinicalResultDeletes(baseFound, baseAllowlist)
		if len(got) != 0 {
			t.Fatalf("expected 0 violations, got %v", got)
		}
	})

	t.Run("unlisted site fails (core regression guard)", func(t *testing.T) {
		extra := map[string]int{
			auditInventoryKey("examination_repository.go", "examinationRepository.ReplaceItemsByExamID", "ExamResult"): 1,
			auditInventoryKey("new_repository.go", "newRepository.DeleteAllResults", "ExamResult"):                     1,
		}
		got := reconcileClinicalResultDeletes(extra, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "NOT on the allowlist") {
			t.Fatalf("expected unlisted site to be flagged, got %v", got)
		}
	})

	t.Run("stale allowlist entry fails", func(t *testing.T) {
		got := reconcileClinicalResultDeletes(map[string]int{}, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "stale") {
			t.Fatalf("expected stale entry to be flagged, got %v", got)
		}
	})

	t.Run("occurrence count mismatch fails", func(t *testing.T) {
		mismatch := map[string]int{
			auditInventoryKey("examination_repository.go", "examinationRepository.ReplaceItemsByExamID", "ExamResult"): 2,
		}
		got := reconcileClinicalResultDeletes(mismatch, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "added/removed") {
			t.Fatalf("expected occurrence mismatch to be flagged, got %v", got)
		}
	})

	t.Run("duplicate allowlist key fails", func(t *testing.T) {
		dup := append([]auditInventoryEntry(nil), baseAllowlist...)
		dup = append(dup, baseAllowlist[0])
		got := reconcileClinicalResultDeletes(baseFound, dup)
		if len(got) != 1 || !strings.Contains(got[0], "duplicate") {
			t.Fatalf("expected duplicate key to be flagged, got %v", got)
		}
	})
}
