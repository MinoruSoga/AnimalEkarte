package reservation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const updateForReservationName = "UpdateForReservation"

func TestReservationStaffWriteOwnerLint(t *testing.T) {
	files := reservationStaffWriteOwnerSources(t)
	if _, ok := files["staff/staff_repository.go"]; !ok {
		t.Fatal("staff/staff_repository.go was not discovered")
	}
	if _, ok := files["reservation/reservation_staff_repository.go"]; !ok {
		t.Fatal("reservation/reservation_staff_repository.go was not discovered")
	}

	violations := make([]string, 0)
	foundStaffRepo := false
	foundStaffsWriter := false
	for _, path := range sortedKeys(files) {
		hits, typed := scanUpdateForReservation(path, files[path])
		violations = append(violations, hits...)
		if path == "staff/staff_repository.go" && typed["StaffRepository"] {
			foundStaffRepo = true
		}
		if path == "reservation/reservation_staff_repository.go" && typed["staffsWriter"] {
			foundStaffsWriter = true
		}
	}
	if len(violations) > 0 {
		t.Fatalf("reservation staffs write-owner violations:\n%s", strings.Join(violations, "\n"))
	}
	if !foundStaffRepo {
		t.Fatal("exported staff.StaffRepository.UpdateForReservation must accept ReservationStaffUpdate")
	}
	if !foundStaffsWriter {
		t.Fatal("staffsWriter.UpdateForReservation must accept ReservationStaffUpdate")
	}
}

func TestReservationStaffWriteOwnerScanner(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		source        string
		wantViolation bool
	}{
		{
			name: "map typed UpdateForReservation is a violation",
			path: "staff/staff_repository.go",
			source: `package staff
func (r *staffRepository) UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return nil
}
`,
			wantViolation: true,
		},
		{
			name: "StaffRepository interface map typed UpdateForReservation is a violation",
			path: "staff/staff_repository.go",
			source: `package staff
type StaffRepository interface {
	UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error
}
`,
			wantViolation: true,
		},
		{
			name: "staffsWriter map typed UpdateForReservation is a violation",
			path: "reservation/reservation_staff_repository.go",
			source: `package reservation
type staffsWriter interface {
	UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error
}
`,
			wantViolation: true,
		},
		{
			name: "typed ReservationStaffUpdate is not a violation",
			path: "staff/staff_repository.go",
			source: `package staff
type ReservationStaffUpdate struct{}
func (r *staffRepository) UpdateForReservation(ctx context.Context, clinicID, id uint64, cmd ReservationStaffUpdate) error {
	return nil
}
`,
			wantViolation: false,
		},
		{
			name: "selector typed ReservationStaffUpdate is not a violation",
			path: "reservation/reservation_staff_repository.go",
			source: `package reservation
type staffsWriter interface {
	UpdateForReservation(ctx context.Context, clinicID, id uint64, cmd staff.ReservationStaffUpdate) error
}
`,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := scanUpdateForReservation(tt.path, tt.source)
			hasViolation := len(got) > 0
			if hasViolation != tt.wantViolation {
				t.Fatalf("scanUpdateForReservation() violations = %v, wantViolation %v", got, tt.wantViolation)
			}
		})
	}
}

func reservationStaffWriteOwnerSources(t *testing.T) map[string]string {
	t.Helper()
	out := make(map[string]string)

	staffDir := filepath.Join("..", "staff")
	staffEntries, err := os.ReadDir(staffDir)
	if err != nil {
		t.Fatalf("read staff dir: %v", err)
	}
	for _, entry := range staffEntries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(staffDir, name)
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		out[filepath.ToSlash(filepath.Join("staff", name))] = string(body)
	}

	reservationEntries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read reservation dir: %v", err)
	}
	for _, entry := range reservationEntries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "reservation_staff") ||
			!strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		body, readErr := os.ReadFile(name)
		if readErr != nil {
			t.Fatalf("read %s: %v", name, readErr)
		}
		out[filepath.ToSlash(filepath.Join("reservation", name))] = string(body)
	}
	return out
}

func scanUpdateForReservation(path, source string) (violations []string, typedInterfaces map[string]bool) {
	typedInterfaces = make(map[string]bool)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse error: %v", path, err)}, typedInterfaces
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			if n.Name == nil || n.Name.Name != updateForReservationName {
				return true
			}
			if funcTypeHasMapStringAnyParam(n.Type) {
				pos := fset.Position(n.Pos())
				violations = append(violations, fmt.Sprintf("%s:%d: UpdateForReservation accepts map[string]any", path, pos.Line))
			}
		case *ast.TypeSpec:
			iface, ok := n.Type.(*ast.InterfaceType)
			if !ok || iface.Methods == nil {
				return true
			}
			for _, method := range iface.Methods.List {
				if len(method.Names) != 1 || method.Names[0].Name != updateForReservationName {
					continue
				}
				fn, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				if funcTypeHasMapStringAnyParam(fn) {
					pos := fset.Position(method.Pos())
					violations = append(violations, fmt.Sprintf("%s:%d: %s.UpdateForReservation accepts map[string]any", path, pos.Line, n.Name.Name))
					continue
				}
				if funcTypeHasReservationStaffUpdateParam(fn) {
					typedInterfaces[n.Name.Name] = true
				}
			}
		}
		return true
	})
	return violations, typedInterfaces
}

func funcTypeHasMapStringAnyParam(fn *ast.FuncType) bool {
	if fn == nil || fn.Params == nil {
		return false
	}
	for _, field := range fn.Params.List {
		if exprIsMapStringAny(field.Type) {
			return true
		}
	}
	return false
}

func funcTypeHasReservationStaffUpdateParam(fn *ast.FuncType) bool {
	if fn == nil || fn.Params == nil {
		return false
	}
	for _, field := range fn.Params.List {
		if exprIsReservationStaffUpdate(field.Type) {
			return true
		}
	}
	return false
}

func exprIsMapStringAny(expr ast.Expr) bool {
	m, ok := expr.(*ast.MapType)
	if !ok {
		return false
	}
	key, ok := m.Key.(*ast.Ident)
	if !ok || key.Name != "string" {
		return false
	}
	switch value := m.Value.(type) {
	case *ast.Ident:
		return value.Name == "any" || value.Name == "interface"
	case *ast.InterfaceType:
		return value.Methods != nil && len(value.Methods.List) == 0
	default:
		return false
	}
}

func exprIsReservationStaffUpdate(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name == "ReservationStaffUpdate"
	case *ast.SelectorExpr:
		return n.Sel != nil && n.Sel.Name == "ReservationStaffUpdate"
	case *ast.StarExpr:
		return exprIsReservationStaffUpdate(n.X)
	case *ast.ParenExpr:
		return exprIsReservationStaffUpdate(n.X)
	default:
		return false
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
