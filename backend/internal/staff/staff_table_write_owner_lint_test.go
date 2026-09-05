package staff

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/lintscan"
)

var staffTableWriteSQL = regexp.MustCompile(`(?i)\b(update|insert\s+into|delete\s+from)\s+(?:"?[a-z_][a-z0-9_$]*"?\s*\.\s*)?"?(?:staffs|shift_entries)"?\b`)

var staffTableGormMutationMethods = map[string]struct{}{
	"Create":           {},
	"CreateInBatches":  {},
	"Delete":           {},
	"FirstOrCreate":    {},
	"Save":             {},
	"Update":           {},
	"UpdateColumn":     {},
	"UpdateColumns":    {},
	"Updates":          {},
	"UpdateScopedByID": {},
}

func TestStaffTableWriteOwnerScanner(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		source        string
		wantViolation bool
	}{
		{
			name: "gorm model update outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { db.Model(&model.Staff{}).Where("id = ?", 1).Update("name", "x") }
`,
			wantViolation: true,
		},
		{
			name: "gorm table create outside owner",
			path: "billing/repository.go",
			source: `package billing
func write(db DB) { db.Table("shift_entries").Create(&row) }
`,
			wantViolation: true,
		},
		{
			name: "gorm create via args outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { db.Create(&model.Staff{}) }
`,
			wantViolation: true,
		},
		{
			name: "count on staff model outside owner is allowed",
			path: "clinic/clinic_repository.go",
			source: `package clinic
import "github.com/animal-ekarte/backend/internal/model"
func read(db DB) { var n int64; db.Model(&model.Staff{}).Where("clinic_id = ?", 1).Count(&n) }
`,
			wantViolation: false,
		},
		{
			name: "owner package may update staffs",
			path: "staff/staff_repository.go",
			source: `package staff
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { db.Model(&model.Staff{}).Where("id = ?", 1).Update("name", "x") }
`,
			wantViolation: false,
		},
		{
			name: "seedlogin migrate upsert may write staffs",
			path: "seedlogin/apply.go",
			source: `package seedlogin
func write(tx Tx) { _, _ = tx.Exec("INSERT INTO staffs (id) VALUES (1)") }
`,
			wantViolation: false,
		},
		{
			name: "find and joins on shift entries outside owner are allowed",
			path: "reservation/reservation_schedule_repository.go",
			source: `package reservation
import "github.com/animal-ekarte/backend/internal/model"
func read(db DB) {
	var entries []model.ShiftEntry
	db.Model(&model.ShiftEntry{}).Joins("JOIN staffs ON staffs.id = shift_entries.staff_id").Find(&entries)
}
`,
			wantViolation: false,
		},
		{
			name: "consumer create with context first is not a GORM write",
			path: "reservation/reservation_staff_service.go",
			source: `package reservation
import (
	"context"
	"github.com/animal-ekarte/backend/internal/model"
)
func write(ctx context.Context, repo Repo, staff *model.Staff) {
	_ = repo.Create(ctx, staff, uint64(1))
}
`,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanStaffTableWriteOwner(tt.path, []byte(tt.source))
			hasViolation := len(got) > 0
			if hasViolation != tt.wantViolation {
				t.Fatalf("scanStaffTableWriteOwner() violations = %v, wantViolation %v", got, tt.wantViolation)
			}
		})
	}
}

func TestStaffTableWriteOwnerLint(t *testing.T) {
	files := lintscan.WalkInternalTreeT(t)
	if len(files) < 500 {
		t.Fatalf("staff table write-owner lint discovered only %d production Go files; whole-tree coverage is not proven", len(files))
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	violations := make([]string, 0)
	for _, path := range paths {
		// testdb is the importable test kernel (WalkInternalTree already drops *_test.go).
		if strings.HasPrefix(path, "testdb/") {
			continue
		}
		violations = append(violations, scanStaffTableWriteOwner(path, files[path])...)
	}
	if len(violations) > 0 {
		t.Fatalf("staffs/shift_entries write ownership violations:\n%s", strings.Join(violations, "\n"))
	}
}

func scanStaffTableWriteOwner(path string, src []byte) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse error: %v", path, err)}
	}

	modelAliases := staffModelImportAliases(file)
	staticStrings := collectStaffStaticStrings(file)
	ownedVars := collectStaffOwnedVariables(file, modelAliases)
	queryVars := collectStaffQueryVariables(file, modelAliases, ownedVars, staticStrings)
	isOwnerPath := strings.HasPrefix(path, "staff/") || strings.HasPrefix(path, "seedlogin/")
	seen := make(map[string]struct{})
	add := func(pos token.Pos, reason string) {
		seen[fmt.Sprintf("%s:%d: %s", path, fset.Position(pos).Line, reason)] = struct{}{}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok || isOwnerPath {
				return true
			}
			if isStaffTableGormWrite(sel, n, modelAliases, ownedVars, queryVars, staticStrings) {
				add(n.Pos(), "direct staffs/shift_entries persistence write outside internal/staff")
			}
			if sel.Sel.Name == "Exec" || sel.Sel.Name == "Raw" {
				for _, arg := range n.Args {
					literal, ok := staffStaticString(arg, staticStrings)
					if ok && staffTableWriteSQL.MatchString(literal) {
						add(n.Pos(), "raw SQL writes staffs/shift_entries outside internal/staff")
					}
				}
			}
		case *ast.BasicLit, *ast.BinaryExpr:
			if isOwnerPath {
				return true
			}
			literal, ok := staffStaticStringNode(n, staticStrings)
			if ok && staffTableWriteSQL.MatchString(literal) {
				add(n.Pos(), "raw SQL writes staffs/shift_entries outside internal/staff")
			}
		}
		return true
	})

	violations := make([]string, 0, len(seen))
	for violation := range seen {
		violations = append(violations, violation)
	}
	sort.Strings(violations)
	return violations
}

func staffModelImportAliases(file *ast.File) map[string]struct{} {
	aliases := make(map[string]struct{})
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil || !strings.HasSuffix(importPath, "/internal/model") {
			continue
		}
		alias := "model"
		if spec.Name != nil {
			alias = spec.Name.Name
		} else if slash := strings.LastIndexByte(importPath, '/'); slash >= 0 {
			alias = importPath[slash+1:]
		}
		if alias != "_" && alias != "." {
			aliases[alias] = struct{}{}
		}
	}
	return aliases
}

func collectStaffOwnedVariables(file *ast.File, modelAliases map[string]struct{}) map[string]struct{} {
	vars := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			collectStaffOwnedFieldNames(vars, n.Recv, modelAliases)
			collectStaffOwnedFieldNames(vars, n.Type.Params, modelAliases)
		case *ast.FuncLit:
			collectStaffOwnedFieldNames(vars, n.Type.Params, modelAliases)
		case *ast.ValueSpec:
			if isStaffOwnedType(n.Type, modelAliases) {
				for _, name := range n.Names {
					vars[name.Name] = struct{}{}
				}
			}
			for i, value := range n.Values {
				if exprIsStaffOwnedValue(value, modelAliases, vars) && i < len(n.Names) {
					vars[n.Names[i].Name] = struct{}{}
				}
			}
		case *ast.AssignStmt:
			for i, value := range n.Rhs {
				if !exprIsStaffOwnedValue(value, modelAliases, vars) || i >= len(n.Lhs) {
					continue
				}
				if name, ok := n.Lhs[i].(*ast.Ident); ok {
					vars[name.Name] = struct{}{}
				}
			}
		}
		return true
	})
	return vars
}

func collectStaffOwnedFieldNames(vars map[string]struct{}, fields *ast.FieldList, modelAliases map[string]struct{}) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		if !isStaffOwnedType(field.Type, modelAliases) {
			continue
		}
		for _, name := range field.Names {
			vars[name.Name] = struct{}{}
		}
	}
}

func collectStaffQueryVariables(
	file *ast.File,
	modelAliases, ownedVars map[string]struct{},
	staticStrings map[string]string,
) map[string]struct{} {
	queryVars := make(map[string]struct{})
	changed := true
	for changed {
		changed = false
		ast.Inspect(file, func(node ast.Node) bool {
			var names []*ast.Ident
			var values []ast.Expr
			switch n := node.(type) {
			case *ast.ValueSpec:
				names = n.Names
				values = n.Values
			case *ast.AssignStmt:
				values = n.Rhs
				for _, lhs := range n.Lhs {
					name, ok := lhs.(*ast.Ident)
					if !ok {
						names = append(names, nil)
						continue
					}
					names = append(names, name)
				}
			default:
				return true
			}
			for i, value := range values {
				if i >= len(names) || names[i] == nil {
					continue
				}
				if !receiverTargetsStaffTables(value, modelAliases, ownedVars, queryVars, staticStrings) {
					continue
				}
				if _, exists := queryVars[names[i].Name]; exists {
					continue
				}
				queryVars[names[i].Name] = struct{}{}
				changed = true
			}
			return true
		})
	}
	return queryVars
}

func receiverTargetsStaffTables(
	expr ast.Expr,
	modelAliases, ownedVars, queryVars map[string]struct{},
	staticStrings map[string]string,
) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		_, ok := queryVars[n.Name]
		return ok
	case *ast.CallExpr:
		sel, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == "Model" && len(n.Args) > 0 && exprIsStaffOwnedValue(n.Args[0], modelAliases, ownedVars) {
			return true
		}
		if sel.Sel.Name == "Table" && len(n.Args) > 0 {
			if table, ok := staffStaticString(n.Args[0], staticStrings); ok && isStaffOwnedTable(table) {
				return true
			}
			if exprIsStaffOwnedTableName(n.Args[0], modelAliases, ownedVars) {
				return true
			}
		}
		return receiverTargetsStaffTables(sel.X, modelAliases, ownedVars, queryVars, staticStrings)
	case *ast.SelectorExpr:
		return receiverTargetsStaffTables(n.X, modelAliases, ownedVars, queryVars, staticStrings)
	case *ast.ParenExpr:
		return receiverTargetsStaffTables(n.X, modelAliases, ownedVars, queryVars, staticStrings)
	default:
		return false
	}
}

func isStaffOwnedTable(table string) bool {
	fields := strings.Fields(strings.TrimSpace(strings.ToLower(table)))
	if len(fields) == 0 {
		return false
	}
	tableName := fields[0]
	if dot := strings.LastIndexByte(tableName, '.'); dot >= 0 {
		tableName = tableName[dot+1:]
	}
	switch strings.Trim(tableName, `"`) {
	case "staffs", "shift_entries":
		return true
	default:
		return false
	}
}

func exprIsStaffOwnedTableName(expr ast.Expr, modelAliases, ownedVars map[string]struct{}) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "TableName" {
		return false
	}
	return exprIsStaffOwnedValue(selector.X, modelAliases, ownedVars)
}

func isStaffTableGormWrite(
	sel *ast.SelectorExpr,
	call *ast.CallExpr,
	modelAliases, ownedVars, queryVars map[string]struct{},
	staticStrings map[string]string,
) bool {
	_, isMutation := staffTableGormMutationMethods[sel.Sel.Name]
	if !isMutation {
		return false
	}
	if receiverTargetsStaffTables(sel.X, modelAliases, ownedVars, queryVars, staticStrings) {
		return true
	}
	if sel.Sel.Name == "UpdateScopedByID" {
		return callArgsContainStaffOwned(call.Args, modelAliases, ownedVars)
	}
	return len(call.Args) > 0 && exprIsStaffOwnedValue(call.Args[0], modelAliases, ownedVars)
}

func callArgsContainStaffOwned(args []ast.Expr, modelAliases, ownedVars map[string]struct{}) bool {
	for _, arg := range args {
		if exprIsStaffOwnedValue(arg, modelAliases, ownedVars) {
			return true
		}
	}
	return false
}

func exprIsStaffOwnedValue(expr ast.Expr, modelAliases, ownedVars map[string]struct{}) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		_, ok := ownedVars[n.Name]
		return ok
	case *ast.UnaryExpr:
		return exprIsStaffOwnedValue(n.X, modelAliases, ownedVars)
	case *ast.ParenExpr:
		return exprIsStaffOwnedValue(n.X, modelAliases, ownedVars)
	case *ast.CompositeLit:
		return isStaffOwnedType(n.Type, modelAliases)
	case *ast.CallExpr:
		ident, ok := n.Fun.(*ast.Ident)
		return ok && ident.Name == "new" && len(n.Args) == 1 && isStaffOwnedType(n.Args[0], modelAliases)
	default:
		return false
	}
}

func isStaffOwnedType(expr ast.Expr, modelAliases map[string]struct{}) bool {
	switch n := expr.(type) {
	case *ast.StarExpr:
		return isStaffOwnedType(n.X, modelAliases)
	case *ast.ArrayType:
		return isStaffOwnedType(n.Elt, modelAliases)
	case *ast.Ellipsis:
		return isStaffOwnedType(n.Elt, modelAliases)
	case *ast.ParenExpr:
		return isStaffOwnedType(n.X, modelAliases)
	case *ast.Ident:
		return isStaffOwnedTypeName(n.Name)
	case *ast.SelectorExpr:
		alias, ok := n.X.(*ast.Ident)
		if !ok || !isStaffOwnedTypeName(n.Sel.Name) {
			return false
		}
		_, ok = modelAliases[alias.Name]
		return ok
	default:
		return false
	}
}

func isStaffOwnedTypeName(name string) bool {
	return name == "Staff" || name == "ShiftEntry"
}

func collectStaffStaticStrings(file *ast.File) map[string]string {
	values := make(map[string]string)
	changed := true
	for changed {
		changed = false
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range valueSpec.Names {
					if i >= len(valueSpec.Values) {
						continue
					}
					value, ok := staffStaticString(valueSpec.Values[i], values)
					if !ok || values[name.Name] == value {
						continue
					}
					values[name.Name] = value
					changed = true
				}
			}
		}
	}
	return values
}

func staffStaticStringNode(node ast.Node, constants map[string]string) (string, bool) {
	expr, ok := node.(ast.Expr)
	if !ok {
		return "", false
	}
	return staffStaticString(expr, constants)
}

func staffStaticString(expr ast.Expr, constants map[string]string) (string, bool) {
	switch n := expr.(type) {
	case *ast.BasicLit:
		if n.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(n.Value)
		if err != nil {
			return "", false
		}
		return value, true
	case *ast.Ident:
		value, ok := constants[n.Name]
		return value, ok
	case *ast.ParenExpr:
		return staffStaticString(n.X, constants)
	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			return "", false
		}
		left, leftOK := staffStaticString(n.X, constants)
		right, rightOK := staffStaticString(n.Y, constants)
		if !leftOK || !rightOK {
			return "", false
		}
		return left + right, true
	default:
		return "", false
	}
}
