package lintscan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type grandchildParentTarget struct {
	modelName          string
	childTable         string
	childFK            string
	parentTable        string
	parentPK           string
	parentClinicColumn string
	childClinicColumn  string
	scopeHelpers       []grandchildScopeHelper
}

type grandchildScopeHelper struct {
	qualifier string
	name      string
	tableArg  string
}

var grandchildParentTargets = []grandchildParentTarget{
	{
		modelName:          "DailyRecord",
		childTable:         "daily_records",
		childFK:            "hospitalization_id",
		parentTable:        "hospitalizations",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "CareLog",
		childTable:         "care_logs",
		childFK:            "daily_record_id",
		parentTable:        "daily_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "ExamResult",
		childTable:         "exam_results",
		childFK:            "exam_id",
		parentTable:        "exams",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
	},
	{
		modelName:          "BillingItem",
		childTable:         "billing_items",
		childFK:            "billing_id",
		parentTable:        "billings",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
	},
	{
		modelName:          "MedicalRecordImage",
		childTable:         "medical_record_images",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		scopeHelpers: []grandchildScopeHelper{
			{qualifier: "persistence", name: "MedicalRecordTenantScope", tableArg: "medical_record_images"},
		},
	},
	{
		modelName:          "MedicalRecordAddendum",
		childTable:         "medical_record_addenda",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "CheckupFieldResult",
		childTable:         "checkup_field_results",
		childFK:            "checkup_id",
		parentTable:        "checkups",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "Prescription",
		childTable:         "prescriptions",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
		scopeHelpers: []grandchildScopeHelper{
			{name: "prescriptionParentClinicScope"},
		},
	},
	{
		modelName:          "TreatmentPlan",
		childTable:         "treatment_plans",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
		scopeHelpers: []grandchildScopeHelper{
			{name: "treatmentPlanParentClinicScope"},
		},
	},
	{
		modelName:          "TreatmentPlan",
		childTable:         "treatment_plans",
		childFK:            "hospitalization_id",
		parentTable:        "hospitalizations",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
		scopeHelpers: []grandchildScopeHelper{
			{name: "treatmentPlanParentClinicScope"},
		},
	},
	{
		modelName:          "ShiftEntryBreak",
		childTable:         "shift_entry_breaks",
		childFK:            "shift_entry_id",
		parentTable:        "shift_entries",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
	},
	// SEC-SWEEP-02 residual surfaces (detectable; production repair is follow-up units).
	{
		modelName:          "AppointmentTrimmingDetail",
		childTable:         "appointment_trimming_details",
		childFK:            "appointment_id",
		parentTable:        "appointments",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "Billing",
		childTable:         "billings",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "Billing",
		childTable:         "billings",
		childFK:            "pet_id",
		parentTable:        "pets",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "Estimate",
		childTable:         "estimates",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "MedicalRecord",
		childTable:         "medical_records",
		childFK:            "appointment_id",
		parentTable:        "appointments",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "VitalRecord",
		childTable:         "vital_records",
		childFK:            "medical_record_id",
		parentTable:        "medical_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
	{
		modelName:          "VitalRecord",
		childTable:         "vital_records",
		childFK:            "daily_record_id",
		parentTable:        "daily_records",
		parentPK:           "id",
		parentClinicColumn: "clinic_id",
		childClinicColumn:  "clinic_id",
	},
}

// grandchildParentResidualSite is a site-level residual allowlist key.
// Line numbers are intentionally excluded — they churn on unrelated edits.
// Keys are (file, function, modelName, childTable); same model at a different
// site is never auto-exempted.
type grandchildParentResidualSite struct {
	file       string
	function   string
	modelName  string
	childTable string
}

// grandchildParentClinicCorrelationResidualSites pins the exact production
// uncorrelated reads that remain until SEC-SWEEP-02-*-B1 repair units. Count
// must match residual findings exactly; new sites fail closed and repaired
// sites must be removed from this list (stale entries fail closed).
// Measured after SEC-SWEEP-02-RES-B1: 0 residual findings remain.
// Last residual (reservation CountMedicalRecordsByReservationID) repaired; Estimate restored.
// Residual gate machinery (type/gate/bidirectional tests) is intentionally retained at count 0.
var grandchildParentClinicCorrelationResidualSites = []grandchildParentResidualSite{}

// grandchildParentClinicCorrelationResidualSiteCount pins the allowlist size.
// Mismatch with observed residual findings is a hard failure.
const grandchildParentClinicCorrelationResidualSiteCount = 0

func grandchildParentResidualSiteKey(file, function, modelName, childTable string) string {
	return file + "\x00" + function + "\x00" + modelName + "\x00" + childTable
}

func grandchildParentResidualSiteKeyFromFinding(f grandchildParentFinding) string {
	return grandchildParentResidualSiteKey(f.file, f.function, f.modelName, f.childTable)
}

func grandchildParentResidualSiteKeyFromSite(s grandchildParentResidualSite) string {
	return grandchildParentResidualSiteKey(s.file, s.function, s.modelName, s.childTable)
}

// grandchildParentResidualGateErrors returns human-readable gate failures for
// residual allowlist evaluation against the production pin. Empty means findings
// match the allowlist exactly: no unlisted residual/non-residual violations, no
// stale allowlist entries, and residual count equals the pinned allowlist size.
func grandchildParentResidualGateErrors(findings []grandchildParentFinding, allowlist []grandchildParentResidualSite) []string {
	return grandchildParentResidualGateErrorsWithCount(findings, allowlist, grandchildParentClinicCorrelationResidualSiteCount)
}

// grandchildParentResidualGateErrorsWithCount is the pin-parameterized gate used by
// bidirectional unit tests so stale/unlisted rejection remains exercisable when the
// production residual allowlist is empty (count 0).
func grandchildParentResidualGateErrorsWithCount(findings []grandchildParentFinding, allowlist []grandchildParentResidualSite, expectedCount int) []string {
	if len(allowlist) != expectedCount {
		return []string{
			"residual allowlist size " + strconv.Itoa(len(allowlist)) +
				" != pinned count " + strconv.Itoa(expectedCount),
		}
	}
	allow := make(map[string]grandchildParentResidualSite, len(allowlist))
	for _, site := range allowlist {
		key := grandchildParentResidualSiteKeyFromSite(site)
		if _, dup := allow[key]; dup {
			return []string{"duplicate residual allowlist site: " + formatResidualSite(site)}
		}
		allow[key] = site
	}

	observed := make(map[string]struct{}, len(findings))
	var errs []string
	for _, f := range findings {
		key := grandchildParentResidualSiteKeyFromFinding(f)
		if _, ok := allow[key]; ok {
			observed[key] = struct{}{}
			continue
		}
		errs = append(errs, "unlisted grandchild parent-clinic correlation violation: "+
			f.file+" func "+f.function+"() "+f.modelName+"/"+f.childTable+": "+f.detail)
	}
	for _, site := range allowlist {
		key := grandchildParentResidualSiteKeyFromSite(site)
		if _, ok := observed[key]; !ok {
			errs = append(errs, "stale residual allowlist entry (site repaired or moved; remove from allowlist): "+
				formatResidualSite(site))
		}
	}
	if len(observed) != expectedCount {
		errs = append(errs, "observed residual findings "+strconv.Itoa(len(observed))+
			" != pinned count "+strconv.Itoa(expectedCount))
	}
	return errs
}

func formatResidualSite(s grandchildParentResidualSite) string {
	return s.file + " func " + s.function + "() " + s.modelName + "/" + s.childTable
}

type grandchildParentFinding struct {
	file       string
	line       int
	function   string
	modelName  string
	childTable string
	detail     string
}

type grandchildParentStats struct {
	filesParsed int
	targetReads int
	// readsByKey tracks function|modelName for production canaries (shared with the
	// pet-facing RealSource gate after registry unification).
	readsByKey map[string]int
}

func analyzeFileGrandchildParentClinicCorrelation(filename string, src []byte) ([]grandchildParentFinding, grandchildParentStats, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, grandchildParentStats{}, err
	}

	base := baseFileName(filename)
	if strings.Contains(filename, "/") {
		base = filename
	}

	var findings []grandchildParentFinding
	stats := grandchildParentStats{filesParsed: 1, readsByKey: make(map[string]int)}

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		varTypes := grandchildLocalVarTypes(fd)
		funcKey := receiverMethodKey(fd)

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok || !grandchildIsReadTerminal(ce) {
				return true
			}

			calls := grandchildCallChain(ce)
			if len(calls) == 0 {
				return true
			}
			query := grandchildQueryInfo(calls)
			target, ok := grandchildTargetForQuery(calls, query, varTypes)
			if !ok {
				return true
			}
			if !grandchildQueryReferencesChildFK(query, target) && !grandchildHasAcceptedScopeHelper(query, target) {
				return true
			}

			stats.targetReads++
			stats.readsByKey[funcKey+"|"+target.modelName]++
			if grandchildHasAcceptedScopeHelper(query, target) || grandchildHasParentClinicCorrelation(query, target) {
				return true
			}

			findings = append(findings, grandchildParentFinding{
				file:       base,
				line:       fset.Position(ce.Pos()).Line,
				function:   funcKey,
				modelName:  target.modelName,
				childTable: target.childTable,
				detail: "direct grandchild read by " + target.childTable + "." + target.childFK +
					" requires parent " + target.parentTable + " clinic correlation",
			})
			return true
		})
	}
	return findings, stats, nil
}

type grandchildMethodCall struct {
	name string
	args []ast.Expr
}

type grandchildQuery struct {
	sqlParts     []string
	rawSQL       bool
	modelNames   map[string]struct{}
	tableNames   map[string]struct{}
	scopeHelpers map[string]struct{}
	readVars     map[string]struct{}
}

func grandchildCallChain(ce *ast.CallExpr) []grandchildMethodCall {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	var calls []grandchildMethodCall
	if parent, ok := sel.X.(*ast.CallExpr); ok {
		calls = append(calls, grandchildCallChain(parent)...)
	}
	calls = append(calls, grandchildMethodCall{name: sel.Sel.Name, args: ce.Args})
	return calls
}

func grandchildIsReadTerminal(ce *ast.CallExpr) bool {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Find", "First", "Take", "Scan", "Count":
		return true
	default:
		return false
	}
}

func grandchildQueryInfo(calls []grandchildMethodCall) grandchildQuery {
	query := grandchildQuery{
		modelNames:   map[string]struct{}{},
		tableNames:   map[string]struct{}{},
		scopeHelpers: map[string]struct{}{},
		readVars:     map[string]struct{}{},
	}
	for _, call := range calls {
		switch call.name {
		case "Model":
			if len(call.args) > 0 {
				if modelName := grandchildModelNameFromExpr(call.args[0]); modelName != "" {
					query.modelNames[modelName] = struct{}{}
				}
			}
		case "Table":
			if len(call.args) > 0 {
				if tableName, ok := stringLitValue(call.args[0]); ok {
					query.tableNames[grandchildFirstSQLIdentifier(tableName)] = struct{}{}
					query.sqlParts = append(query.sqlParts, tableName)
				}
			}
		case "Joins", "Where", "Raw":
			if call.name == "Raw" {
				query.rawSQL = true
			}
			if len(call.args) > 0 {
				if sql, ok := stringLitValue(call.args[0]); ok {
					query.sqlParts = append(query.sqlParts, sql)
				}
			}
		case "Scopes":
			for _, arg := range call.args {
				if key := grandchildScopeHelperKey(arg); key != "" {
					query.scopeHelpers[key] = struct{}{}
				}
			}
		case "Find", "First", "Take", "Scan", "Count":
			for _, arg := range call.args {
				if name := grandchildReadVarName(arg); name != "" {
					query.readVars[name] = struct{}{}
				}
			}
		}
	}
	return query
}

func grandchildTargetForQuery(
	calls []grandchildMethodCall,
	query grandchildQuery,
	varTypes map[string]string,
) (grandchildParentTarget, bool) {
	for _, target := range grandchildParentTargets {
		if _, ok := query.modelNames[target.modelName]; ok &&
			(grandchildQueryReferencesChildFK(query, target) || grandchildHasAcceptedScopeHelper(query, target)) {
			return target, true
		}
		if _, ok := query.tableNames[target.childTable]; ok &&
			(grandchildQueryReferencesChildFK(query, target) || grandchildHasAcceptedScopeHelper(query, target)) {
			return target, true
		}
		for name := range query.readVars {
			if varTypes[name] == target.modelName &&
				(grandchildQueryReferencesChildFK(query, target) || grandchildHasAcceptedScopeHelper(query, target)) {
				return target, true
			}
		}
		normalized := grandchildNormalizedSQL(query.sqlParts...)
		if query.rawSQL &&
			strings.Contains(normalized, target.childTable) &&
			strings.Contains(normalized, target.childFK) &&
			grandchildRawDirectlySelectsTarget(query, target) {
			return target, true
		}
		for _, call := range calls {
			for _, arg := range call.args {
				if grandchildModelNameFromExpr(arg) == target.modelName &&
					(grandchildQueryReferencesChildFK(query, target) || grandchildHasAcceptedScopeHelper(query, target)) {
					return target, true
				}
			}
		}
	}
	return grandchildParentTarget{}, false
}

func grandchildRawDirectlySelectsTarget(query grandchildQuery, target grandchildParentTarget) bool {
	if _, ok := query.modelNames[target.modelName]; ok {
		return true
	}
	tokens := strings.Fields(strings.NewReplacer(",", " ", "(", " ", ")", " ").Replace(grandchildNormalizedSQL(query.sqlParts...)))
	lastSelect := -1
	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "select":
			lastSelect = i
		case "from":
			if i+1 >= len(tokens) || grandchildCleanSQLIdentifier(tokens[i+1]) != target.childTable || lastSelect < 0 {
				continue
			}
			aliases := map[string]struct{}{target.childTable: {}}
			if i+2 < len(tokens) {
				next := grandchildCleanSQLIdentifier(tokens[i+2])
				if next == "as" && i+3 < len(tokens) {
					aliases[grandchildCleanSQLIdentifier(tokens[i+3])] = struct{}{}
				} else if !grandchildSQLKeyword(next) {
					aliases[next] = struct{}{}
				}
			}
			for _, selected := range tokens[lastSelect+1 : i] {
				selected = strings.Trim(selected, " ,")
				if selected == "*" || selected == "id" {
					return true
				}
				for alias := range aliases {
					if selected == alias+".id" {
						return true
					}
				}
			}
		}
	}
	return false
}

func grandchildQueryReferencesChildFK(query grandchildQuery, target grandchildParentTarget) bool {
	normalized := grandchildNormalizedSQL(query.sqlParts...)
	compact := grandchildCompactSQL(query.sqlParts...)
	if strings.Contains(normalized, target.childFK) {
		return true
	}
	if strings.Contains(compact, target.childTable+"."+target.childFK) {
		return true
	}
	for readVar := range query.readVars {
		_ = readVar
	}
	return false
}

func grandchildHasParentClinicCorrelation(query grandchildQuery, target grandchildParentTarget) bool {
	sql := grandchildCompactSQL(query.sqlParts...)
	if sql == "" {
		return false
	}

	childAliases := grandchildAliasesForTable(query.sqlParts, target.childTable)
	parentAliases := grandchildAliasesForTable(query.sqlParts, target.parentTable)

	hasJoin := false
	for _, childAlias := range childAliases {
		for _, parentAlias := range parentAliases {
			left := parentAlias + "." + target.parentPK + "=" + childAlias + "." + target.childFK
			right := childAlias + "." + target.childFK + "=" + parentAlias + "." + target.parentPK
			if strings.Contains(sql, left) || strings.Contains(sql, right) {
				hasJoin = true
			}
		}
	}
	if !hasJoin {
		return false
	}

	for _, parentAlias := range parentAliases {
		if strings.Contains(sql, parentAlias+"."+target.parentClinicColumn+"=?") {
			return true
		}
		if target.childClinicColumn == "" {
			continue
		}
		for _, childAlias := range childAliases {
			left := parentAlias + "." + target.parentClinicColumn + "=" + childAlias + "." + target.childClinicColumn
			right := childAlias + "." + target.childClinicColumn + "=" + parentAlias + "." + target.parentClinicColumn
			if strings.Contains(sql, left) || strings.Contains(sql, right) {
				return true
			}
		}
	}
	return false
}

func grandchildHasAcceptedScopeHelper(query grandchildQuery, target grandchildParentTarget) bool {
	for _, helper := range target.scopeHelpers {
		key := helper.qualifier + "." + helper.name + "|" + helper.tableArg
		if _, ok := query.scopeHelpers[key]; ok {
			return true
		}
	}
	return false
}

func grandchildScopeHelperKey(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return "." + ident.Name + "|"
	}
	ce, ok := expr.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	qualifier, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	tableArg := ""
	if len(ce.Args) > 0 {
		if s, ok := stringLitValue(ce.Args[0]); ok {
			tableArg = s
		}
	}
	return qualifier.Name + "." + sel.Sel.Name + "|" + tableArg
}

func grandchildLocalVarTypes(fd *ast.FuncDecl) map[string]string {
	varTypes := map[string]string{}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			modelName := grandchildModelNameFromExpr(x.Type)
			if modelName == "" {
				return true
			}
			for _, name := range x.Names {
				varTypes[name.Name] = modelName
			}
		case *ast.AssignStmt:
			for i, rhs := range x.Rhs {
				modelName := grandchildModelNameFromExpr(rhs)
				if modelName == "" || i >= len(x.Lhs) {
					continue
				}
				if ident, ok := x.Lhs[i].(*ast.Ident); ok {
					varTypes[ident.Name] = modelName
				}
			}
		}
		return true
	})
	return varTypes
}

func grandchildModelNameFromExpr(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.UnaryExpr:
		return grandchildModelNameFromExpr(x.X)
	case *ast.CompositeLit:
		return grandchildModelNameFromExpr(x.Type)
	case *ast.ArrayType:
		return grandchildModelNameFromExpr(x.Elt)
	case *ast.StarExpr:
		return grandchildModelNameFromExpr(x.X)
	case *ast.SelectorExpr:
		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "model" {
			return x.Sel.Name
		}
	case *ast.CallExpr:
		if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "make" && len(x.Args) > 0 {
			return grandchildModelNameFromExpr(x.Args[0])
		}
	}
	return ""
}

func grandchildReadVarName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			return grandchildReadVarName(x.X)
		}
	case *ast.Ident:
		return x.Name
	}
	return ""
}

func grandchildAliasesForTable(sqlParts []string, table string) []string {
	aliases := map[string]struct{}{table: {}}
	tokens := strings.Fields(strings.NewReplacer(",", " ", "(", " ", ")", " ").Replace(grandchildNormalizedSQL(sqlParts...)))
	for i := 0; i < len(tokens)-1; i++ {
		if tokens[i] != "from" && tokens[i] != "join" && tokens[i] != "update" {
			continue
		}
		if grandchildCleanSQLIdentifier(tokens[i+1]) != table {
			continue
		}
		if i+2 >= len(tokens) {
			continue
		}
		next := grandchildCleanSQLIdentifier(tokens[i+2])
		if next == "as" && i+3 < len(tokens) {
			aliases[grandchildCleanSQLIdentifier(tokens[i+3])] = struct{}{}
			continue
		}
		if !grandchildSQLKeyword(next) {
			aliases[next] = struct{}{}
		}
	}
	return sortedStringSet(aliases)
}

func grandchildNormalizedSQL(parts ...string) string {
	joined := strings.ToLower(strings.Join(parts, " "))
	replacer := strings.NewReplacer(
		"\n", " ",
		"\t", " ",
		"\r", " ",
		"`", "",
		"\"", "",
	)
	joined = replacer.Replace(joined)
	return strings.Join(strings.Fields(joined), " ")
}

func grandchildCompactSQL(parts ...string) string {
	sql := grandchildNormalizedSQL(parts...)
	replacer := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "", "\"", "", "`", "")
	return replacer.Replace(sql)
}

func grandchildFirstSQLIdentifier(s string) string {
	fields := strings.Fields(strings.ToLower(s))
	if len(fields) == 0 {
		return ""
	}
	return grandchildCleanSQLIdentifier(fields[0])
}

func grandchildCleanSQLIdentifier(s string) string {
	s = strings.Trim(s, " ,()")
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return s
}

func grandchildSQLKeyword(s string) bool {
	switch s {
	case "", "on", "where", "and", "or", "left", "right", "inner", "outer", "join", "group", "order", "limit", "for":
		return true
	default:
		return false
	}
}

func TestGrandchildParentClinicCorrelation_Analyzer(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{
			name: "daily_records gorm read by hospitalization_id missing parent hospitalization correlation",
			body: `db.Model(&model.DailyRecord{}).Scopes(persistence.ClinicScope(clinicID)).Where("hospitalization_id = ?", hospitalizationID).Find(&records)`,
			want: 1,
		},
		{
			name: "care_logs gorm read by daily_record_id missing parent daily_record correlation",
			body: `db.Model(&model.CareLog{}).Scopes(persistence.ClinicScope(clinicID)).Where("daily_record_id = ?", dailyRecordID).Find(&logs)`,
			want: 1,
		},
		{
			name: "exam_results gorm read by exam_id missing parent exam correlation",
			body: `db.Model(&model.ExamResult{}).Where("exam_results.exam_id = ?", examID).Find(&items)`,
			want: 1,
		},
		{
			name: "billing_items gorm read by billing_id missing parent billing correlation",
			body: `db.Model(&model.BillingItem{}).Where("billing_items.billing_id = ?", billingID).Find(&items)`,
			want: 1,
		},
		{
			name: "medical_record_images gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.MedicalRecordImage{}).Where("medical_record_images.medical_record_id = ?", medicalRecordID).Find(&images)`,
			want: 1,
		},
		{
			name: "medical_record_addenda gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.MedicalRecordAddendum{}).Scopes(persistence.ClinicScope(clinicID)).Where("medical_record_id = ?", medicalRecordID).Find(&addenda)`,
			want: 1,
		},
		{
			name: "checkup_field_results gorm read by checkup_id missing parent checkup correlation",
			body: `db.Model(&model.CheckupFieldResult{}).Scopes(persistence.ClinicScope(clinicID)).Where("checkup_field_results.checkup_id = ?", checkupID).Find(&results)`,
			want: 1,
		},
		{
			name: "prescriptions gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.Prescription{}).Scopes(persistence.ClinicScope(clinicID)).Where("prescriptions.medical_record_id = ?", medicalRecordID).Find(&prescriptions)`,
			want: 1,
		},
		{
			name: "treatment_plans gorm read by hospitalization_id missing parent hospitalization correlation",
			body: `db.Model(&model.TreatmentPlan{}).Scopes(persistence.ClinicScope(clinicID)).Where("treatment_plans.hospitalization_id = ?", hospitalizationID).Find(&plans)`,
			want: 1,
		},
		// SEC-SWEEP-02 residual model classes: uncorrelated reads must be detected.
		{
			name: "appointment_trimming_details gorm read by appointment_id missing parent appointment correlation",
			body: `db.Model(&model.AppointmentTrimmingDetail{}).Scopes(persistence.ClinicScope(clinicID)).Where("appointment_id = ?", appointmentID).First(&detail)`,
			want: 1,
		},
		{
			name: "billings gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.Billing{}).Scopes(persistence.ClinicScope(clinicID)).Where("medical_record_id = ?", medicalRecordID).Find(&billings)`,
			want: 1,
		},
		{
			name: "estimates gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.Estimate{}).Scopes(persistence.ClinicScope(clinicID)).Where("medical_record_id = ?", medicalRecordID).Find(&estimates)`,
			want: 1,
		},
		{
			name: "medical_records gorm read by appointment_id missing parent appointment correlation",
			body: `db.Model(&model.MedicalRecord{}).Where("clinic_id = ? AND appointment_id = ?", clinicID, appointmentID).First(&record)`,
			want: 1,
		},
		{
			name: "vital_records gorm read by medical_record_id missing parent medical_record correlation",
			body: `db.Model(&model.VitalRecord{}).Scopes(persistence.ClinicScope(clinicID)).Where("vital_records.medical_record_id = ?", medicalRecordID).Find(&vitals)`,
			want: 1,
		},
		{
			name: "exam_results raw SQL missing parent exam correlation",
			body: "db.Raw(`SELECT * FROM exam_results WHERE exam_id = ?`, examID).Scan(&items)",
			want: 1,
		},
		{
			name: "billing_items raw SQL alias missing parent billing correlation",
			body: "db.Raw(`SELECT * FROM billing_items bi WHERE bi.billing_id = ?`, billingID).Scan(&items)",
			want: 1,
		},
		{
			name: "care_logs raw SQL joins parent id only without clinic correlation",
			body: "db.Raw(`SELECT * FROM care_logs cl JOIN daily_records dr ON dr.id = cl.daily_record_id WHERE cl.daily_record_id = ?`, dailyRecordID).Scan(&logs)",
			want: 1,
		},
		{
			name: "prescriptions raw SQL select id alias missing parent medical_record correlation",
			body: "db.Raw(`SELECT p.id FROM prescriptions p WHERE p.medical_record_id = ?`, medicalRecordID).Scan(&ids)",
			want: 1,
		},
		// Pattern coverage for residual classes: raw SQL, alias, qualified identifier, stale join.
		{
			name: "appointment_trimming_details raw SQL alias missing parent appointment correlation",
			body: "db.Raw(`SELECT * FROM appointment_trimming_details atd WHERE atd.appointment_id = ?`, appointmentID).Scan(&details)",
			want: 1,
		},
		{
			name: "estimates raw SQL qualified identifier missing parent medical_record correlation",
			body: "db.Raw(`SELECT * FROM estimates WHERE estimates.medical_record_id = ?`, medicalRecordID).Scan(&estimates)",
			want: 1,
		},
		{
			name: "vital_records raw SQL stale parent id join without clinic correlation",
			body: "db.Raw(`SELECT * FROM vital_records vr JOIN medical_records mr ON mr.id = vr.medical_record_id WHERE vr.medical_record_id = ?`, medicalRecordID).Scan(&vitals)",
			want: 1,
		},
		{
			name: "billings raw SQL select id alias missing parent medical_record correlation",
			body: "db.Raw(`SELECT b.id FROM billings b WHERE b.medical_record_id = ?`, medicalRecordID).Scan(&ids)",
			want: 1,
		},
		{
			name: "medical_records raw SQL missing parent appointment correlation",
			body: "db.Raw(`SELECT * FROM medical_records WHERE appointment_id = ?`, appointmentID).Scan(&records)",
			want: 1,
		},
		{
			name: "daily_records gorm read with hospitalization clinic correlation",
			body: `db.Model(&model.DailyRecord{}).Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").Where("daily_records.hospitalization_id = ?", hospitalizationID).Find(&records)`,
			want: 0,
		},
		{
			name: "care_logs gorm read with daily_record clinic correlation",
			body: `db.Model(&model.CareLog{}).Joins("JOIN daily_records ON daily_records.id = care_logs.daily_record_id AND daily_records.clinic_id = care_logs.clinic_id").Where("care_logs.daily_record_id = ?", dailyRecordID).Find(&logs)`,
			want: 0,
		},
		{
			name: "exam_results gorm read with parent exam clinic predicate",
			body: `db.Model(&model.ExamResult{}).Joins("JOIN exams ON exams.id = exam_results.exam_id").Where("exams.clinic_id = ? AND exam_results.exam_id = ?", clinicID, examID).Find(&items)`,
			want: 0,
		},
		{
			name: "billing_items gorm read with parent billing clinic predicate",
			body: `db.Model(&model.BillingItem{}).Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ?", clinicID).Where("billing_items.billing_id = ?", billingID).Find(&items)`,
			want: 0,
		},
		{
			name: "medical_record_images accepted medical record tenant scope",
			body: `db.Model(&model.MedicalRecordImage{}).Scopes(persistence.MedicalRecordTenantScope("medical_record_images", clinicID)).Where("medical_record_images.id = ?", id).First(&image)`,
			want: 0,
		},
		{
			name: "medical_record_addenda raw SQL with parent medical_record clinic correlation",
			body: "db.Raw(`SELECT * FROM medical_record_addenda mra JOIN medical_records mr ON mr.id = mra.medical_record_id AND mr.clinic_id = mra.clinic_id WHERE mra.medical_record_id = ?`, medicalRecordID).Scan(&addenda)",
			want: 0,
		},
		{
			name: "checkup_field_results gorm read with parent checkup clinic correlation",
			body: `db.Model(&model.CheckupFieldResult{}).Where("EXISTS (SELECT 1 FROM checkups WHERE checkups.id = checkup_field_results.checkup_id AND checkups.clinic_id = checkup_field_results.clinic_id)").Where("checkup_field_results.checkup_id = ?", checkupID).Find(&results)`,
			want: 0,
		},
		{
			name: "prescriptions accepted parent clinic helper",
			body: `db.Model(&model.Prescription{}).Scopes(prescriptionParentClinicScope).Where("prescriptions.medical_record_id = ?", medicalRecordID).Find(&prescriptions)`,
			want: 0,
		},
		{
			name: "treatment_plans accepted parent clinic helper for medical record",
			body: `db.Model(&model.TreatmentPlan{}).Scopes(treatmentPlanParentClinicScope).Where("treatment_plans.medical_record_id = ?", medicalRecordID).Find(&plans)`,
			want: 0,
		},
		{
			name: "treatment_plans raw SQL with parent hospitalization clinic correlation",
			body: "db.Raw(`SELECT * FROM treatment_plans tp JOIN hospitalizations h ON h.id = tp.hospitalization_id AND h.clinic_id = tp.clinic_id WHERE tp.hospitalization_id = ?`, hospitalizationID).Scan(&plans)",
			want: 0,
		},
		{
			name: "appointment_trimming_details gorm read with parent appointment clinic correlation",
			body: `db.Model(&model.AppointmentTrimmingDetail{}).Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id").Where("appointment_trimming_details.appointment_id = ?", appointmentID).First(&detail)`,
			want: 0,
		},
		{
			name: "billings gorm read with parent medical_record clinic correlation",
			body: `db.Model(&model.Billing{}).Joins("JOIN medical_records ON medical_records.id = billings.medical_record_id AND medical_records.clinic_id = billings.clinic_id").Where("billings.medical_record_id = ?", medicalRecordID).Find(&billings)`,
			want: 0,
		},
		{
			name: "estimates gorm read with parent medical_record clinic correlation",
			body: `db.Model(&model.Estimate{}).Joins("JOIN medical_records ON medical_records.id = estimates.medical_record_id AND medical_records.clinic_id = estimates.clinic_id").Where("estimates.medical_record_id = ?", medicalRecordID).Find(&estimates)`,
			want: 0,
		},
		{
			name: "medical_records gorm read with parent appointment clinic correlation",
			body: `db.Model(&model.MedicalRecord{}).Joins("JOIN appointments ON appointments.id = medical_records.appointment_id AND appointments.clinic_id = medical_records.clinic_id").Where("medical_records.appointment_id = ?", appointmentID).First(&record)`,
			want: 0,
		},
		{
			name: "vital_records gorm read with parent medical_record clinic correlation",
			body: `db.Model(&model.VitalRecord{}).Joins("JOIN medical_records ON medical_records.id = vital_records.medical_record_id AND medical_records.clinic_id = vital_records.clinic_id").Where("vital_records.medical_record_id = ?", medicalRecordID).Find(&vitals)`,
			want: 0,
		},
		{
			name: "pet lifecycle predicates are not required",
			body: `db.Model(&model.BillingItem{}).Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ?", clinicID).Where("billing_items.billing_id = ?", billingID).Find(&items)`,
			want: 0,
		},
		{
			name: "unrelated direct pet child is out of grandchild lint scope",
			body: `db.Model(&model.PetChronicCondition{}).Where("clinic_id = ? AND pet_id = ?", clinicID, petID).Find(&records)`,
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package p\nfunc f() { _ = " + tc.body + " }\n"
			findings, _, err := analyzeFileGrandchildParentClinicCorrelation("fixture.go", []byte(src))
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.want, findings)
			}
		})
	}
}

func TestGrandchildParentClinicCorrelation_AnalyzerDetectsViolationUnderNestedPathFilename(t *testing.T) {
	src := []byte(`package p
func f() {
	_ = db.Model(&model.ExamResult{}).Where("exam_results.exam_id = ?", examID).Find(&items)
}
`)
	filenames := []string{
		"repository/x.go",
		"medicalrecord/x.go",
		"medicalrecord/two/deep/x.go",
	}
	for _, fn := range filenames {
		t.Run(fn, func(t *testing.T) {
			findings, _, err := analyzeFileGrandchildParentClinicCorrelation(fn, src)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("got %d findings for filename %q, want 1: %+v", len(findings), fn, findings)
			}
		})
	}
}

func TestGrandchildParentClinicCorrelation_AcceptsRegisteredScopeHelpers(t *testing.T) {
	src := []byte(`package p
func f() {
	_ = db.Model(&model.MedicalRecordImage{}).
		Scopes(persistence.MedicalRecordTenantScope("medical_record_images", clinicID)).
		Where("medical_record_images.id = ?", id).
		First(&image)
}
`)
	findings, _, err := analyzeFileGrandchildParentClinicCorrelation("fixture.go", src)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("registered scope helper should be accepted, got findings: %+v", findings)
	}
}

func TestGrandchildParentClinicCorrelation_RegistryCoversMinimumSecSweep02Targets(t *testing.T) {
	// Pin registry *rows* (modelName/parentTable/childFK), not just unique model
	// names, so dual-FK rows cannot be deleted silently.
	wantRows := map[string]struct{}{
		"DailyRecord|hospitalizations|hospitalization_id":         {},
		"CareLog|daily_records|daily_record_id":                   {},
		"ExamResult|exams|exam_id":                                {},
		"BillingItem|billings|billing_id":                         {},
		"MedicalRecordImage|medical_records|medical_record_id":    {},
		"MedicalRecordAddendum|medical_records|medical_record_id": {},
		"CheckupFieldResult|checkups|checkup_id":                  {},
		"Prescription|medical_records|medical_record_id":          {},
		"TreatmentPlan|medical_records|medical_record_id":         {},
		"TreatmentPlan|hospitalizations|hospitalization_id":       {},
		"AppointmentTrimmingDetail|appointments|appointment_id":   {},
		"Billing|medical_records|medical_record_id":               {},
		"Billing|pets|pet_id":                                     {},
		"Estimate|medical_records|medical_record_id":              {},
		"MedicalRecord|appointments|appointment_id":               {},
		"VitalRecord|medical_records|medical_record_id":           {},
		"VitalRecord|daily_records|daily_record_id":               {},
		"ShiftEntryBreak|shift_entries|shift_entry_id":            {},
	}
	gotRows := map[string]struct{}{}
	gotModels := map[string]struct{}{}
	for _, target := range grandchildParentTargets {
		row := target.modelName + "|" + target.parentTable + "|" + target.childFK
		gotRows[row] = struct{}{}
		gotModels[target.modelName] = struct{}{}
	}
	for row := range wantRows {
		if _, ok := gotRows[row]; !ok {
			t.Errorf("grandchild parent correlation registry missing SEC-SWEEP-02 row %s", row)
		}
	}
	if len(wantRows) != 18 {
		t.Fatalf("wantRows pin size %d, want 18 registry rows", len(wantRows))
	}
	if len(gotRows) < len(wantRows) {
		t.Errorf("grandchild parent correlation registry has %d rows, want at least %d: %v",
			len(gotRows), len(wantRows), sortedStringSet(gotRows))
	}
	// Unique model floor retained as a secondary pin (14 models after S1).
	if len(gotModels) < 15 {
		t.Errorf("grandchild parent correlation registry has %d unique models, want at least 15: %v",
			len(gotModels), sortedStringSet(gotModels))
	}
}

func TestGrandchildParentClinicCorrelation_ResidualAllowlistRejectsUnlistedSite(t *testing.T) {
	// Synthetic residual-model finding at a site not on the allowlist must FAIL the gate.
	// This proves new uncorrelated reads on residual models are no longer silently logged.
	base := make([]grandchildParentFinding, 0, len(grandchildParentClinicCorrelationResidualSites)+1)
	for _, site := range grandchildParentClinicCorrelationResidualSites {
		base = append(base, grandchildParentFinding{
			file:       site.file,
			function:   site.function,
			modelName:  site.modelName,
			childTable: site.childTable,
			detail:     "synthetic residual site",
		})
	}
	unlisted := grandchildParentFinding{
		file:       "billing/unlisted_residual_site.go",
		function:   "unlistedRepository.FindByMedicalRecordID",
		modelName:  "Billing",
		childTable: "billings",
		detail:     "direct grandchild read by billings.medical_record_id requires parent medical_records clinic correlation",
	}
	errs := grandchildParentResidualGateErrors(append(base, unlisted), grandchildParentClinicCorrelationResidualSites)
	if len(errs) == 0 {
		t.Fatal("expected residual gate to reject unlisted residual-model site; gate did not fail")
	}
	joined := strings.Join(errs, "\n")
	if !strings.Contains(joined, "unlisted") || !strings.Contains(joined, "billing/unlisted_residual_site.go") {
		t.Fatalf("expected unlisted-site rejection error, got: %v", errs)
	}
}

func TestGrandchildParentClinicCorrelation_ResidualAllowlistRejectsStaleEntry(t *testing.T) {
	// Allowlist entry with no corresponding finding (site repaired) must FAIL the gate.
	// This forces follow-up repair units to shrink the allowlist when they fix a site.
	// Uses a synthetic fixture allowlist, independent of the live (now-empty)
	// grandchildParentClinicCorrelationResidualSites — this test proves gate
	// behavior, not current production residual state, and must not panic or
	// become a no-op once SEC-SWEEP-02's real residual count reaches 0.
	syntheticAllowlist := []grandchildParentResidualSite{
		{file: "billing/synthetic_repaired_site.go", function: "syntheticRepository.RepairedRead", modelName: "Billing", childTable: "billings"},
		{file: "billing/synthetic_pending_site.go", function: "syntheticRepository.PendingRead", modelName: "Estimate", childTable: "estimates"},
	}
	findings := []grandchildParentFinding{
		{
			file:       syntheticAllowlist[1].file,
			function:   syntheticAllowlist[1].function,
			modelName:  syntheticAllowlist[1].modelName,
			childTable: syntheticAllowlist[1].childTable,
			detail:     "synthetic residual site",
		},
	}
	// expectedCount must match the synthetic allowlist length (not the production pin),
	// otherwise the size pin fails before stale-entry detection can run at residual count 0.
	errs := grandchildParentResidualGateErrorsWithCount(findings, syntheticAllowlist, len(syntheticAllowlist))
	if len(errs) == 0 {
		t.Fatal("expected residual gate to reject stale allowlist entry; gate did not fail")
	}
	joined := strings.Join(errs, "\n")
	if !strings.Contains(joined, "stale residual allowlist entry") {
		t.Fatalf("expected stale-entry rejection error, got: %v", errs)
	}
	if !strings.Contains(joined, syntheticAllowlist[0].file) {
		t.Fatalf("expected stale entry to name repaired site %q, got: %v",
			syntheticAllowlist[0].file, errs)
	}
}

func TestGrandchildParentClinicCorrelation_ResidualAllowlistExactMatchPasses(t *testing.T) {
	findings := make([]grandchildParentFinding, 0, len(grandchildParentClinicCorrelationResidualSites))
	for _, site := range grandchildParentClinicCorrelationResidualSites {
		findings = append(findings, grandchildParentFinding{
			file:       site.file,
			function:   site.function,
			modelName:  site.modelName,
			childTable: site.childTable,
			detail:     "synthetic residual site",
		})
	}
	errs := grandchildParentResidualGateErrors(findings, grandchildParentClinicCorrelationResidualSites)
	if len(errs) != 0 {
		t.Fatalf("exact allowlist match should pass, got: %v", errs)
	}
}

func TestGrandchildParentClinicCorrelation_DiscoveryReachesModuleWideAndNestedPackages(t *testing.T) {
	tree := moduleInternalSource(t)

	foundNestedPackageCanary := false
	nestedProductionFiles := 0
	for name := range tree {
		if name == "infra/smtp/sender.go" {
			foundNestedPackageCanary = true
		}
		if strings.Count(name, "/") >= 2 {
			nestedProductionFiles++
		}
	}
	if !foundNestedPackageCanary {
		t.Fatal("module-wide discovery does not include infra/smtp/sender.go; lintscan walk may have narrowed")
	}
	if nestedProductionFiles < 10 {
		t.Fatalf("only %d nested production source files discovered, want >= 10; lintscan walk may have narrowed", nestedProductionFiles)
	}

	assertDiscoversFileFromDifferentTopLevelPackage(t, tree)
	assertLintscanReachesTwoOrMoreNestingLevels(t)
}

func TestGrandchildParentClinicCorrelation_RealSourceHasNoUncorrelatedGrandchildReads(t *testing.T) {
	findings, stats := walkGrandchildParentClinicCorrelation(t)

	if stats.filesParsed < 40 {
		t.Fatalf("only %d internal source files parsed; discovery likely broken", stats.filesParsed)
	}
	if stats.targetReads < 9 {
		t.Fatalf("only %d grandchild target reads detected; AST matching likely broken", stats.targetReads)
	}
	// Site-level residual allowlist with exact count pin. New residual-model
	// violations and stale (repaired) allowlist entries both fail closed.
	if len(grandchildParentClinicCorrelationResidualSites) != grandchildParentClinicCorrelationResidualSiteCount {
		t.Fatalf("residual allowlist has %d sites, want pinned count %d",
			len(grandchildParentClinicCorrelationResidualSites), grandchildParentClinicCorrelationResidualSiteCount)
	}
	for _, finding := range findings {
		key := grandchildParentResidualSiteKeyFromFinding(finding)
		matched := false
		for _, site := range grandchildParentClinicCorrelationResidualSites {
			if grandchildParentResidualSiteKeyFromSite(site) == key {
				matched = true
				break
			}
		}
		if matched {
			t.Logf("residual grandchild parent-clinic finding (follow-up repair unit): %s:%d func %s() %s/%s: %s",
				finding.file, finding.line, finding.function, finding.modelName, finding.childTable, finding.detail)
		}
	}
	for _, errMsg := range grandchildParentResidualGateErrors(findings, grandchildParentClinicCorrelationResidualSites) {
		t.Error(errMsg)
	}
}

func walkGrandchildParentClinicCorrelation(t *testing.T) ([]grandchildParentFinding, grandchildParentStats) {
	t.Helper()
	tree := moduleInternalSource(t)

	names := make([]string, 0, len(tree))
	for name := range tree {
		names = append(names, name)
	}
	sort.Strings(names)

	var all []grandchildParentFinding
	agg := grandchildParentStats{readsByKey: make(map[string]int)}
	for _, rawKey := range names {
		key := legacyLintKey(rawKey)
		findings, stats, err := analyzeFileGrandchildParentClinicCorrelation(key, tree[rawKey])
		if err != nil {
			t.Fatalf("parse %s: %v", key, err)
		}
		all = append(all, findings...)
		agg.filesParsed += stats.filesParsed
		agg.targetReads += stats.targetReads
		for statKey, count := range stats.readsByKey {
			agg.readsByKey[statKey] += count
		}
	}
	return all, agg
}

func sortedStringSet(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
