package clinic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClinicRepositoryConsumerUpdateRejectsMap(t *testing.T) {
	src, err := os.ReadFile("ports.go")
	require.NoError(t, err)

	violations := clinicRepositoryMapUpdateViolations(string(src))
	if len(violations) > 0 {
		t.Fatalf("ClinicRepository still exposes a generic map update:\n%s", violations[0])
	}
}

func TestClinicRepositoryMapUpdateScanner(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		wantViolate bool
	}{
		{
			name: "old map Update on ClinicRepository is a violation",
			src: `package clinic
type ClinicRepository interface {
	Update(ctx context.Context, id uint64, fields map[string]any) error
}
`,
			wantViolate: true,
		},
		{
			name: "typed UpdateClinic is not a violation",
			src: `package clinic
type ClinicRepository interface {
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) error
}
`,
			wantViolate: false,
		},
		{
			name: "other interfaces may still take maps",
			src: `package clinic
type ClinicRepository interface {
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) error
}
type CompanyRepository interface {
	Update(ctx context.Context, fields map[string]any) error
}
`,
			wantViolate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := clinicRepositoryMapUpdateViolations(tt.src)
			if tt.wantViolate && len(violations) == 0 {
				t.Fatal("expected a map-update violation")
			}
			if !tt.wantViolate && len(violations) > 0 {
				t.Fatalf("unexpected violations: %v", violations)
			}
		})
	}
}

func clinicRepositoryMapUpdateViolations(src string) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "ports.go", src, 0)
	if err != nil {
		return []string{"parse error: " + err.Error()}
	}

	var violations []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name == nil || spec.Name.Name != "ClinicRepository" {
			return true
		}
		iface, ok := spec.Type.(*ast.InterfaceType)
		if !ok || iface.Methods == nil {
			return false
		}
		for _, method := range iface.Methods.List {
			fn, ok := method.Type.(*ast.FuncType)
			if !ok || fn.Params == nil || len(method.Names) == 0 {
				continue
			}
			for _, param := range fn.Params.List {
				if isMapStringAny(param.Type) {
					violations = append(violations, method.Names[0].Name+" accepts map[string]any")
				}
			}
		}
		return false
	})
	return violations
}

func isMapStringAny(expr ast.Expr) bool {
	m, ok := expr.(*ast.MapType)
	if !ok {
		return false
	}
	key, ok := m.Key.(*ast.Ident)
	if !ok || key.Name != "string" {
		return false
	}
	val, ok := m.Value.(*ast.Ident)
	return ok && val.Name == "any"
}
