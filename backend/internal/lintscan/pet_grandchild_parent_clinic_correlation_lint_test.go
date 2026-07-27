package lintscan

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

type petGrandchildTarget struct {
	model        string
	table        string
	parentTable  string
	fkColumn     string
	requiredJoin []string
}

var petGrandchildTargets = []petGrandchildTarget{
	{
		model:       "DailyRecord",
		table:       "daily_records",
		parentTable: "hospitalizations",
		fkColumn:    "hospitalization_id",
		requiredJoin: []string{
			"hospitalizations.id = daily_records.hospitalization_id",
			"hospitalizations.clinic_id = daily_records.clinic_id",
		},
	},
	{
		model:       "CareLog",
		table:       "care_logs",
		parentTable: "daily_records",
		fkColumn:    "daily_record_id",
		requiredJoin: []string{
			"daily_records.id = care_logs.daily_record_id",
			"daily_records.clinic_id = care_logs.clinic_id",
		},
	},
	{
		model:       "ExamResult",
		table:       "exam_results",
		parentTable: "exams",
		fkColumn:    "exam_id",
		requiredJoin: []string{
			"exam_results.exam_id = exams.id",
			"exams.clinic_id = ?",
		},
	},
	{
		model:       "BillingItem",
		table:       "billing_items",
		parentTable: "billings",
		fkColumn:    "billing_id",
		requiredJoin: []string{
			"billings.id = billing_items.billing_id",
			"billings.clinic_id",
		},
	},
	{
		model:       "MedicalRecordImage",
		table:       "medical_record_images",
		parentTable: "medical_records",
		fkColumn:    "medical_record_id",
		requiredJoin: []string{
			"medical_records.id = medical_record_images.medical_record_id",
			"medical_records.clinic_id",
		},
	},
	{
		model:       "MedicalRecordAddendum",
		table:       "medical_record_addenda",
		parentTable: "medical_records",
		fkColumn:    "medical_record_id",
		requiredJoin: []string{
			"medical_records.id = medical_record_addenda.medical_record_id",
			"medical_records.clinic_id = medical_record_addenda.clinic_id",
		},
	},
}

type petGrandchildFinding struct {
	file     string
	line     int
	function string
	model    string
	table    string
	detail   string
}

type petGrandchildStats struct {
	filesParsed int
	readsByKey  map[string]int
}

func analyzeFilePetGrandchildParentClinicCorrelation(filename string, src []byte) ([]petGrandchildFinding, petGrandchildStats, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, petGrandchildStats{}, err
	}

	base := baseFileName(filename)
	if strings.Contains(filename, "/") {
		base = filename
	}
	stats := petGrandchildStats{filesParsed: 1, readsByKey: make(map[string]int)}
	var findings []petGrandchildFinding

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		varTypes := collectLocalModelVars(fd.Body)
		funcKey := receiverMethodKey(fd)
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok || !isReadTerminal(ce) {
				return true
			}
			literals := normalizedCallChainLiterals(ce)
			models := readModelsFromCall(ce, varTypes)
			for _, target := range petGrandchildTargets {
				if !readTargetsModel(target, models) || !containsAny(literals, target.fkColumn) {
					continue
				}
				stats.readsByKey[funcKey+"|"+target.model]++
				if hasRequiredJoin(literals, target.requiredJoin) {
					continue
				}
				findings = append(findings, petGrandchildFinding{
					file:     base,
					line:     fset.Position(ce.Pos()).Line,
					function: funcKey,
					model:    target.model,
					table:    target.table,
					detail:   fmt.Sprintf("read filters by %s but lacks parent clinic correlation through %s", target.fkColumn, target.parentTable),
				})
			}
			return true
		})
	}
	return findings, stats, nil
}

func collectLocalModelVars(body *ast.BlockStmt) map[string]string {
	vars := make(map[string]string)
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ValueSpec:
			modelName := modelNameFromExpr(node.Type)
			if modelName == "" {
				return true
			}
			for _, name := range node.Names {
				vars[name.Name] = modelName
			}
		case *ast.AssignStmt:
			for i, lhs := range node.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || i >= len(node.Rhs) {
					continue
				}
				if modelName := modelNameFromExpr(node.Rhs[i]); modelName != "" {
					vars[ident.Name] = modelName
				}
			}
		}
		return true
	})
	return vars
}

func modelNameFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.ArrayType:
		return modelNameFromExpr(e.Elt)
	case *ast.StarExpr:
		return modelNameFromExpr(e.X)
	case *ast.CompositeLit:
		return modelNameFromExpr(e.Type)
	case *ast.CallExpr:
		if fn, ok := e.Fun.(*ast.Ident); ok && fn.Name == "make" && len(e.Args) > 0 {
			return modelNameFromExpr(e.Args[0])
		}
	case *ast.SelectorExpr:
		if pkg, ok := e.X.(*ast.Ident); ok && pkg.Name == "model" {
			return e.Sel.Name
		}
	}
	return ""
}

func isReadTerminal(ce *ast.CallExpr) bool {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Find", "First", "Count", "Scan":
		return true
	default:
		return false
	}
}

func normalizedCallChainLiterals(ce *ast.CallExpr) []string {
	var literals []string
	ast.Inspect(ce, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := stringLitValue(arg); ok {
				literals = append(literals, normalizeSQLFragment(lit))
			}
		}
		return true
	})
	return literals
}

func readModelsFromCall(ce *ast.CallExpr, varTypes map[string]string) map[string]struct{} {
	models := make(map[string]struct{})
	ast.Inspect(ce, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Model" && len(call.Args) > 0 {
			if modelName := modelNameFromExpr(call.Args[0]); modelName != "" {
				models[modelName] = struct{}{}
			}
		}
		for _, arg := range call.Args {
			if ident := identFromPossiblyAddressed(arg); ident != "" {
				if modelName := varTypes[ident]; modelName != "" {
					models[modelName] = struct{}{}
				}
			}
		}
		return true
	})
	return models
}

func identFromPossiblyAddressed(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return identFromPossiblyAddressed(e.X)
		}
	}
	return ""
}

func readTargetsModel(target petGrandchildTarget, models map[string]struct{}) bool {
	if _, ok := models[target.model]; ok {
		return true
	}
	return false
}

func hasRequiredJoin(literals, required []string) bool {
	for _, req := range required {
		if !containsAny(literals, req) {
			return false
		}
	}
	return true
}

func containsAny(literals []string, needle string) bool {
	normalizedNeedle := normalizeSQLFragment(needle)
	for _, lit := range literals {
		if strings.Contains(lit, normalizedNeedle) {
			return true
		}
	}
	return false
}

func normalizeSQLFragment(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func TestPetGrandchildParentClinicCorrelation_AnalyzerFixtures(t *testing.T) {
	fixtures := []struct {
		name     string
		src      string
		wantFind int
	}{
		{
			name: "daily record read missing hospitalization clinic correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 1,
		},
		{
			name: "addendum read missing medical record clinic correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.medical_record_id = ?", 1, 2).Find(&records)
}`,
			wantFind: 1,
		},
		{
			name: "daily record read with parent correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "addendum read with parent correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id").Where("medical_record_addenda.medical_record_id = ?", 2).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "historical read does not require pet lifecycle predicates",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "unrelated table ignored",
			src: `package fixture
func f(db DB) {
	db.Where("pets.id = ?", 1).Find(&pets)
}`,
			wantFind: 0,
		},
	}
	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			findings, _, err := analyzeFilePetGrandchildParentClinicCorrelation("repository.go", []byte(tt.src))
			if err != nil {
				t.Fatalf("analyze fixture: %v", err)
			}
			if len(findings) != tt.wantFind {
				t.Fatalf("findings = %d, want %d: %#v", len(findings), tt.wantFind, findings)
			}
		})
	}
}

func TestPetGrandchildParentClinicCorrelation_LocationAgnostic(t *testing.T) {
	src := []byte(`package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Where("medical_record_addenda.medical_record_id = ?", 2).Find(&records)
}`)
	for _, file := range []string{"repository.go", "medicalrecord/repository.go", "medicalrecord/sub/repository.go"} {
		t.Run(file, func(t *testing.T) {
			findings, _, err := analyzeFilePetGrandchildParentClinicCorrelation(file, src)
			if err != nil {
				t.Fatalf("analyze fixture: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("findings = %d, want 1", len(findings))
			}
			if findings[0].file != file && file != "repository.go" {
				t.Fatalf("finding file = %q, want %q", findings[0].file, file)
			}
		})
	}
}

func TestPetGrandchildParentClinicCorrelation_RealSourceHasNoUncorrelatedReads(t *testing.T) {
	tree := moduleInternalSource(t)
	if len(tree) < 40 {
		t.Fatalf("only %d production files discovered; lint would be vacuous", len(tree))
	}

	totalStats := petGrandchildStats{readsByKey: make(map[string]int)}
	var findings []petGrandchildFinding
	for key, src := range tree {
		fileFindings, stats, err := analyzeFilePetGrandchildParentClinicCorrelation(legacyLintKey(key), src)
		if err != nil {
			t.Fatalf("analyze %s: %v", key, err)
		}
		findings = append(findings, fileFindings...)
		totalStats.filesParsed += stats.filesParsed
		for statKey, count := range stats.readsByKey {
			totalStats.readsByKey[statKey] += count
		}
	}

	for _, required := range []string{
		"dailyRecordRepository.FindByHospitalizationID|DailyRecord",
		"dailyRecordRepository.FindByHospitalizationIDAndDate|DailyRecord",
		"medicalRecordAddendumRepository.FindByID|MedicalRecordAddendum",
		"medicalRecordAddendumRepository.FindByMedicalRecordID|MedicalRecordAddendum",
		"medicalRecordImageRepository.FindByMedicalRecordID|MedicalRecordImage",
		"billingItemRepository.FindByID|BillingItem",
		"examinationRepository.FindAllItemsByExamID|ExamResult",
	} {
		if totalStats.readsByKey[required] == 0 {
			t.Fatalf("required read site %q was not detected; lint would be vacuous", required)
		}
	}

	if len(findings) == 0 {
		return
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].file != findings[j].file {
			return findings[i].file < findings[j].file
		}
		return findings[i].line < findings[j].line
	})
	for _, finding := range findings {
		t.Errorf("%s:%d %s %s(%s): %s", finding.file, finding.line, finding.function, finding.model, finding.table, finding.detail)
	}
}
