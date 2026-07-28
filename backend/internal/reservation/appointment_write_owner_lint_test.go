package reservation

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

var appointmentWriteSQL = regexp.MustCompile(`(?i)\b(update|insert\s+into|delete\s+from)\s+(?:"?[a-z_][a-z0-9_$]*"?\s*\.\s*)?"?appointments"?\b`)

var gormMutationMethods = map[string]struct{}{
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

const backendInternalImportPrefix = "github.com/animal-ekarte/backend/internal/"

type appointmentPackageFacts struct {
	staticStrings                      map[string]string
	reservationFactories               map[string]struct{}
	reservationMethodFactories         map[string]map[string]struct{}
	functionReturnTypes                map[string]string
	staticStringFactories              map[string]string
	reservationTypeNames               map[string]struct{}
	mapStringAnyTypeNames              map[string]struct{}
	structFieldTypes                   map[string]map[string]string
	importedReservationFactories       map[string]map[string]struct{}
	importedReservationMethodFactories map[string]map[string]map[string]struct{}
	importedFunctionReturnTypes        map[string]map[string]string
}

// TestAppointmentWriteOwnerLint prevents a second appointments persistence implementation
// from appearing outside internal/reservation. Reads are intentionally allowed: the ownership
// rule is about writes, not about forcing every query through a single package.
func TestAppointmentWriteOwnerLint(t *testing.T) {
	files := lintscan.WalkInternalTreeT(t)
	if len(files) < 500 {
		t.Fatalf("appointment write-owner lint discovered only %d production Go files; whole-tree coverage is not proven", len(files))
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	packageFacts := collectAppointmentPackageFacts(t, files)

	violations := make([]string, 0, len(paths))
	for _, path := range paths {
		violations = append(violations, scanAppointmentWriteOwnerWithFacts(
			path,
			files[path],
			packageFacts[appointmentPackagePath(path)],
		)...)
	}
	if len(violations) > 0 {
		t.Fatalf("appointments write ownership violations:\n%s", strings.Join(violations, "\n"))
	}
}

func TestAppointmentWriteOwnerScanner(t *testing.T) {
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
func write(db DB) { db.Model(&model.Reservation{}).Where("id = ?", 1).Update("status", "completed") }
`,
			wantViolation: true,
		},
		{
			name: "gorm first or create outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { appointment := &model.Reservation{}; db.Model(&model.Reservation{}).FirstOrCreate(appointment) }
`,
			wantViolation: true,
		},
		{
			name: "gorm query variable update outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) {
	query := db.Model(&model.Reservation{}).Where("clinic_id = ?", 1)
	query.Update("status", "completed")
}
`,
			wantViolation: true,
		},
		{
			name: "typed value save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { var appt model.Reservation; db.Save(&appt) }
`,
			wantViolation: true,
		},
		{
			name: "typed function parameter save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func save(db DB, appt *model.Reservation) { db.Save(appt) }
`,
			wantViolation: true,
		},
		{
			name: "typed function parameter model update outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func save(db DB, appt *model.Reservation) { db.Model(appt).Update("status", "completed") }
`,
			wantViolation: true,
		},
		{
			name: "typed function parameter query update outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func save(db DB, appt *model.Reservation) {
	query := db.Model(appt).Where("clinic_id = ?", 1)
	query.Update("status", "completed")
}
`,
			wantViolation: true,
		},
		{
			name: "inferred function return save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func loadAppointment() *model.Reservation { return &model.Reservation{} }
func write(db DB) {
	appointment := loadAppointment()
	db.Save(appointment)
}
`,
			wantViolation: true,
		},
		{
			name: "receiver method return save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
type repository struct{}
func (repository) makeAppointment() *model.Reservation { return &model.Reservation{} }
func write(db DB, repo repository) { db.Save(repo.makeAppointment()) }
`,
			wantViolation: true,
		},
		{
			name: "receiver create method return save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
type repository struct{}
func (repository) Create() *model.Reservation { return &model.Reservation{} }
func write(db DB, repo repository) { db.Save(repo.Create()) }
`,
			wantViolation: true,
		},
		{
			name: "reservation slice create in batches outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB, appointments []model.Reservation) { db.CreateInBatches(&appointments, 100) }
`,
			wantViolation: true,
		},
		{
			name: "reservation array save outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB, appointments [2]*model.Reservation) { db.Save(&appointments) }
`,
			wantViolation: true,
		},
		{
			name: "raw SQL update outside owner",
			path: "lstep/repository.go",
			source: `package lstep
func write(db DB) { db.Exec("UPDATE appointments SET status = ? WHERE id = ?", "no_show", 1) }
`,
			wantViolation: true,
		},
		{
			name: "schema qualified raw SQL update outside owner",
			path: "lstep/repository.go",
			source: `package lstep
func write(db DB) { db.Exec("UPDATE public.appointments SET status = ? WHERE id = ?", "no_show", 1) }
`,
			wantViolation: true,
		},
		{
			name: "quoted schema qualified raw SQL update outside owner",
			path: "lstep/repository.go",
			source: `package lstep
func write(db DB) { db.Exec("UPDATE \"public\".\"appointments\" SET status = ? WHERE id = ?", "no_show", 1) }
`,
			wantViolation: true,
		},
		{
			name: "concatenated raw SQL update outside owner",
			path: "lstep/repository.go",
			source: `package lstep
const updatePrefix = "UPDATE "
const updateAppointment = updatePrefix + "appointments SET status = ? WHERE id = ?"
func write(db DB) { db.Exec(updateAppointment, "no_show", 1) }
`,
			wantViolation: true,
		},
		{
			name: "local concatenated raw SQL update outside owner",
			path: "lstep/repository.go",
			source: `package lstep
func write(db DB) {
	const updatePrefix = "UPDATE "
	const updateAppointment = updatePrefix + "appointments SET status = ? WHERE id = ?"
	db.Exec(updateAppointment, "no_show", 1)
}
`,
			wantViolation: true,
		},
		{
			name: "raw SQL assembled through helper function outside owner",
			path: "lstep/repository.go",
			source: `package lstep
func appointmentTable() string { return "appointments" }
func write(db DB) {
	query := "UPDATE " + appointmentTable() + " SET status = ? WHERE id = ?"
	db.Exec(query, "no_show", 1)
}
`,
			wantViolation: true,
		},
		{
			name: "gorm table alias update outside owner",
			path: "billing/repository.go",
			source: `package billing
func write(db DB) { db.Table("appointments AS a").Where("a.id = ?", 1).Update("status", "completed") }
`,
			wantViolation: true,
		},
		{
			name: "gorm schema qualified table update outside owner",
			path: "billing/repository.go",
			source: `package billing
func write(db DB) { db.Table("public.appointments").Where("id = ?", 1).Update("status", "completed") }
`,
			wantViolation: true,
		},
		{
			name: "gorm constant table alias update outside owner",
			path: "billing/repository.go",
			source: `package billing
const appointmentTable = "appointments AS a"
func write(db DB) { db.Table(appointmentTable).Where("a.id = ?", 1).Updates(map[string]any{"status": "completed"}) }
`,
			wantViolation: true,
		},
		{
			name: "gorm local constant table alias update outside owner",
			path: "billing/repository.go",
			source: `package billing
func write(db DB) {
	const appointmentTable = "appointments AS a"
	db.Table(appointmentTable).Where("a.id = ?", 1).Updates(map[string]any{"status": "completed"})
}
`,
			wantViolation: true,
		},
		{
			name: "gorm TableName method update outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { db.Table(model.Reservation{}.TableName()).Update("status", "completed") }
`,
			wantViolation: true,
		},
		{
			name: "gorm TableName variable update outside owner",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) {
	table := model.Reservation{}.TableName()
	db.Table(table).Update("status", "completed")
}
`,
			wantViolation: true,
		},
		{
			name: "generic reservation update interface outside owner",
			path: "medicalrecord/deps.go",
			source: `package medicalrecord
import (
	"context"
	"github.com/animal-ekarte/backend/internal/model"
)
type store interface {
	Update(context.Context, uint64, uint64, map[string]any) (*model.Reservation, error)
}
`,
			wantViolation: true,
		},
		{
			name: "generic appointment patch interface outside owner",
			path: "medicalrecord/deps.go",
			source: `package medicalrecord
import "context"
type appointmentStore interface {
	Patch(context.Context, uint64, uint64, map[string]any) error
}
`,
			wantViolation: true,
		},
		{
			name: "generic exported appointment patch method outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "context"
type appointmentStore struct{}
func (*appointmentStore) Patch(context.Context, uint64, uint64, map[string]any) error { return nil }
`,
			wantViolation: true,
		},
		{
			name: "named generic appointment patch interface outside owner",
			path: "medicalrecord/deps.go",
			source: `package medicalrecord
import "context"
type appointmentPatch map[string]any
type appointmentStore interface {
	Patch(context.Context, uint64, uint64, appointmentPatch) error
}
`,
			wantViolation: true,
		},
		{
			name: "legacy interface generic appointment patch interface outside owner",
			path: "medicalrecord/deps.go",
			source: `package medicalrecord
import "context"
type appointmentStore interface {
	Patch(context.Context, uint64, uint64, map[string]interface{}) error
}
`,
			wantViolation: true,
		},
		{
			name: "aliased generic appointment patch method outside owner",
			path: "medicalrecord/repository.go",
			source: `package medicalrecord
import "context"
type appointmentPatch = map[string]any
type appointmentStore struct{}
func (*appointmentStore) Patch(context.Context, uint64, uint64, appointmentPatch) error { return nil }
`,
			wantViolation: true,
		},
		{
			name: "generic exported reservation mutation outside interface",
			path: "reservation/repository.go",
			source: `package reservation
import (
	"context"
	"github.com/animal-ekarte/backend/internal/model"
)
func Mutate(context.Context, uint64, uint64, map[string]any) (*model.Reservation, error) { return nil, nil }
`,
			wantViolation: true,
		},
		{
			name: "owner must not export generic appointment update",
			path: "reservation/repository.go",
			source: `package reservation
import (
	"context"
	"github.com/animal-ekarte/backend/internal/model"
)
type ReservationCRUDRepository interface {
	Update(context.Context, uint64, uint64, map[string]any) (*model.Reservation, error)
}
`,
			wantViolation: true,
		},
		{
			name: "broad source interface dependency outside owner",
			path: "billing/service.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/reservation"
type service struct { reservations reservation.ReservationRepository }
`,
			wantViolation: true,
		},
		{
			name: "reservation read outside owner is allowed",
			path: "billing/repository.go",
			source: `package billing
import "github.com/animal-ekarte/backend/internal/model"
func read(db DB) { var n int64; db.Model(&model.Reservation{}).Where("clinic_id = ?", 1).Count(&n) }
`,
			wantViolation: false,
		},
		{
			name: "intent-specific consumer interface is allowed",
			path: "lstep/deps.go",
			source: `package lstep
import "context"
type batchReservationRepository interface {
	MarkNoShow(context.Context, uint64, uint64) (bool, error)
}
`,
			wantViolation: false,
		},
		{
			name: "owner package may persist reservations",
			path: "reservation/repository.go",
			source: `package reservation
import "github.com/animal-ekarte/backend/internal/model"
func write(db DB) { db.Model(&model.Reservation{}).Update("status", "completed") }
`,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanAppointmentWriteOwner(tt.path, []byte(tt.source))
			if tt.wantViolation {
				assertScannerViolation(t, got)
				return
			}
			if len(got) != 0 {
				t.Fatalf("unexpected violations: %v", got)
			}
		})
	}
}

func TestAppointmentWriteOwnerScannerPackageConstants(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string][]byte
		target string
		reason string
	}{
		{
			name: "raw SQL assembled in another file",
			files: map[string][]byte{
				"billing/constants.go": []byte(`package billing
const updatePrefix = "UPDATE "
const updateAppointment = updatePrefix + "appointments SET status = ? WHERE id = ?"
`),
				"billing/repository.go": []byte(`package billing
func write(db DB) { db.Exec(updateAppointment, "completed", 1) }
`),
			},
			target: "billing/repository.go",
			reason: "cross-file SQL constant must not bypass the write-owner gate",
		},
		{
			name: "table alias declared in another file",
			files: map[string][]byte{
				"billing/constants.go": []byte(`package billing
const appointmentTable = "appointments AS a"
`),
				"billing/repository.go": []byte(`package billing
func write(db DB) { db.Table(appointmentTable).Where("a.id = ?", 1).Update("status", "completed") }
`),
			},
			target: "billing/repository.go",
			reason: "cross-file table constant must not bypass the write-owner gate",
		},
		{
			name: "reservation factory declared in another file",
			files: map[string][]byte{
				"billing/load.go": []byte(`package billing
import "github.com/animal-ekarte/backend/internal/model"
func loadAppointment() *model.Reservation { return &model.Reservation{} }
`),
				"billing/repository.go": []byte(`package billing
func write(db DB) { appointment := loadAppointment(); db.Save(appointment) }
`),
			},
			target: "billing/repository.go",
			reason: "cross-file reservation-returning helper must not bypass the write-owner gate",
		},
		{
			name: "reservation method factory declared in another file",
			files: map[string][]byte{
				"billing/load.go": []byte(`package billing
import "github.com/animal-ekarte/backend/internal/model"
type repository struct{}
func (repository) makeAppointment() *model.Reservation { return &model.Reservation{} }
`),
				"billing/repository.go": []byte(`package billing
func write(db DB, repo repository) { db.Save(repo.makeAppointment()) }
`),
			},
			target: "billing/repository.go",
			reason: "cross-file reservation-returning method must not bypass the write-owner gate",
		},
		{
			name: "arbitrarily named package factory declared in another package",
			files: map[string][]byte{
				"appointmenthelper/factory.go": []byte(`package appointmenthelper
import "github.com/animal-ekarte/backend/internal/model"
func Build() *model.Reservation { return &model.Reservation{} }
`),
				"medicalrecord/repository.go": []byte(`package medicalrecord
import helper "github.com/animal-ekarte/backend/internal/appointmenthelper"
func write(db DB) { db.Save(helper.Build()) }
`),
			},
			target: "medicalrecord/repository.go",
			reason: "an imported reservation-returning helper must be identified by its declared type, not its function name",
		},
		{
			name: "receiver factory declared in another package",
			files: map[string][]byte{
				"appointmenthelper/factory.go": []byte(`package appointmenthelper
import "github.com/animal-ekarte/backend/internal/model"
type Builder struct{}
func (Builder) Create() *model.Reservation { return &model.Reservation{} }
`),
				"medicalrecord/repository.go": []byte(`package medicalrecord
import helper "github.com/animal-ekarte/backend/internal/appointmenthelper"
func write(db DB) { db.Save(helper.Builder{}.Create()) }
`),
			},
			target: "medicalrecord/repository.go",
			reason: "an imported receiver method returning a reservation must not bypass the write-owner gate",
		},
		{
			name: "inferred receiver factory value declared in another package",
			files: map[string][]byte{
				"appointmenthelper/factory.go": []byte(`package appointmenthelper
import "github.com/animal-ekarte/backend/internal/model"
type Builder struct{}
func (Builder) Create() *model.Reservation { return &model.Reservation{} }
`),
				"medicalrecord/repository.go": []byte(`package medicalrecord
import helper "github.com/animal-ekarte/backend/internal/appointmenthelper"
func write(db DB) {
	var builder = helper.Builder{}
	db.Save(builder.Create())
}
`),
			},
			target: "medicalrecord/repository.go",
			reason: "an inferred imported receiver value must retain its factory method type",
		},
		{
			name: "constructor receiver factory declared in another package",
			files: map[string][]byte{
				"appointmenthelper/factory.go": []byte(`package appointmenthelper
import "github.com/animal-ekarte/backend/internal/model"
type Builder struct{}
func NewBuilder() *Builder { return &Builder{} }
func (Builder) Create() *model.Reservation { return &model.Reservation{} }
`),
				"medicalrecord/repository.go": []byte(`package medicalrecord
import helper "github.com/animal-ekarte/backend/internal/appointmenthelper"
func write(db DB) { db.Save(helper.NewBuilder().Create()) }
`),
			},
			target: "medicalrecord/repository.go",
			reason: "an imported constructor return type must lead to its reservation factory method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := collectAppointmentPackageFacts(t, tt.files)
			got := scanAppointmentWriteOwnerWithFacts(
				tt.target,
				tt.files[tt.target],
				facts[appointmentPackagePath(tt.target)],
			)
			if len(got) == 0 {
				t.Fatal(tt.reason)
			}
		})
	}
}

func assertScannerViolation(t *testing.T, violations []string) {
	t.Helper()
	if len(violations) == 0 {
		t.Fatal("scanner accepted a known write-owner violation")
	}
}

func scanAppointmentWriteOwner(path string, src []byte) []string {
	return scanAppointmentWriteOwnerWithFacts(path, src, appointmentPackageFacts{})
}

func scanAppointmentWriteOwnerWithFacts(path string, src []byte, packageFacts appointmentPackageFacts) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse error: %v", path, err)}
	}

	modelAliases, sourceAliases := appointmentImportAliases(file)
	reservationTypeNames := cloneStringSet(packageFacts.reservationTypeNames)
	mapStringAnyTypeNames := cloneStringSet(packageFacts.mapStringAnyTypeNames)
	collectAppointmentNamedTypeFactsInto(
		[]*ast.File{file},
		reservationTypeNames,
		mapStringAnyTypeNames,
	)
	staticStrings := make(map[string]string, len(packageFacts.staticStrings))
	for name, value := range packageFacts.staticStrings {
		staticStrings[name] = value
	}
	for name, value := range collectStaticStrings(file) {
		staticStrings[name] = value
	}
	reservationFactories := make(map[string]struct{}, len(packageFacts.reservationFactories))
	for name := range packageFacts.reservationFactories {
		reservationFactories[name] = struct{}{}
	}
	reservationMethodFactories := cloneNestedStringSet(packageFacts.reservationMethodFactories)
	functionReturnTypes := cloneStringMap(packageFacts.functionReturnTypes)
	collectReservationFactoriesInto(
		file,
		modelAliases,
		reservationTypeNames,
		reservationFactories,
		reservationMethodFactories,
	)
	collectFunctionReturnTypesInto(file, functionReturnTypes)
	structFieldTypes := cloneNestedStringMap(packageFacts.structFieldTypes)
	collectStructFieldTypesInto(file, structFieldTypes)
	for alias, importPath := range appointmentImportPaths(file) {
		for name := range packageFacts.importedReservationFactories[importPath] {
			reservationFactories[alias+"."+name] = struct{}{}
		}
		for receiverType, methods := range packageFacts.importedReservationMethodFactories[importPath] {
			for method := range methods {
				addNestedStringSetValue(reservationMethodFactories, alias+"."+receiverType, method)
			}
		}
		for function, returnType := range packageFacts.importedFunctionReturnTypes[importPath] {
			functionReturnTypes[alias+"."+function] = alias + "." + returnType
		}
	}
	collectReservationFactoryCallPositionsInto(
		file,
		reservationMethodFactories,
		structFieldTypes,
		functionReturnTypes,
		reservationFactories,
	)
	staticStringFactories := make(map[string]string, len(packageFacts.staticStringFactories))
	for name, value := range packageFacts.staticStringFactories {
		staticStringFactories[name] = value
	}
	collectStaticStringFactoriesInto([]*ast.File{file}, staticStrings, staticStringFactories)
	reservationVars := collectReservationVariables(file, modelAliases, reservationTypeNames, reservationFactories)
	collectAppointmentTableVariablesInto(
		file,
		modelAliases,
		reservationTypeNames,
		reservationVars,
		reservationFactories,
		staticStrings,
		staticStringFactories,
	)
	appointmentQueryVars := collectAppointmentQueryVariables(file, modelAliases, reservationTypeNames, reservationVars, reservationFactories, staticStrings, staticStringFactories)
	isOwnerPath := strings.HasPrefix(path, "reservation/")
	seen := make(map[string]struct{})
	add := func(pos token.Pos, reason string) {
		line := fset.Position(pos).Line
		key := fmt.Sprintf("%s:%d: %s", path, line, reason)
		seen[key] = struct{}{}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !isOwnerPath {
				if _, isMutation := gormMutationMethods[sel.Sel.Name]; isMutation &&
					(receiverTargetsAppointments(sel.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories, appointmentQueryVars, staticStrings, staticStringFactories) ||
						callArgsContainReservation(n.Args, modelAliases, reservationTypeNames, reservationVars, reservationFactories)) {
					add(n.Pos(), "direct appointments persistence write outside internal/reservation")
				}
			}
			if !isOwnerPath && (sel.Sel.Name == "Exec" || sel.Sel.Name == "Raw") {
				for _, arg := range n.Args {
					literal, ok := staticStringWithFactories(arg, staticStrings, staticStringFactories)
					if ok && appointmentWriteSQL.MatchString(literal) {
						add(n.Pos(), "raw SQL writes appointments outside internal/reservation")
					}
				}
			}
		case *ast.TypeSpec:
			iface, ok := n.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}
			for _, method := range iface.Methods.List {
				if len(method.Names) != 1 || !ast.IsExported(method.Names[0].Name) {
					continue
				}
				fn, ok := method.Type.(*ast.FuncType)
				if !ok || !fieldListContainsMapStringAny(fn.Params, mapStringAnyTypeNames) {
					continue
				}
				if appointmentCapabilityName(n.Name.Name) ||
					appointmentCapabilityName(method.Names[0].Name) ||
					funcTypeReturnsReservation(fn, modelAliases, reservationTypeNames) {
					add(method.Pos(), "generic appointment mutation map capability is exported")
				}
			}
		case *ast.FuncDecl:
			if ast.IsExported(n.Name.Name) &&
				fieldListContainsMapStringAny(n.Type.Params, mapStringAnyTypeNames) &&
				(funcTypeReturnsReservation(n.Type, modelAliases, reservationTypeNames) ||
					appointmentCapabilityName(n.Name.Name) ||
					fieldListTargetsAppointmentCapability(n.Recv)) {
				add(n.Pos(), "generic exported appointment mutation capability")
			}
		case *ast.SelectorExpr:
			alias, ok := n.X.(*ast.Ident)
			if !isOwnerPath && ok && isBroadReservationSourceType(n.Sel.Name) {
				if _, sourcePackage := sourceAliases[alias.Name]; sourcePackage && !isReservationFacadePath(path) {
					add(n.Pos(), "broad ReservationRepository dependency; declare an intent-specific consumer interface")
				}
			}
		case *ast.BasicLit, *ast.BinaryExpr:
			if !isOwnerPath {
				literal, ok := staticStringNodeWithFactories(n, staticStrings, staticStringFactories)
				if ok && appointmentWriteSQL.MatchString(literal) {
					add(n.Pos(), "raw SQL writes appointments outside internal/reservation")
				}
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

func funcTypeReturnsReservation(
	fn *ast.FuncType,
	modelAliases, reservationTypeNames map[string]struct{},
) bool {
	if fn.Results == nil {
		return false
	}
	for _, result := range fn.Results.List {
		if isReservationType(result.Type, modelAliases, reservationTypeNames) {
			return true
		}
	}
	return false
}

func appointmentImportAliases(file *ast.File) (modelAliases, sourceAliases map[string]struct{}) {
	modelAliases = make(map[string]struct{})
	sourceAliases = make(map[string]struct{})
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		alias := ""
		if spec.Name != nil {
			alias = spec.Name.Name
		} else if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
			alias = path[slash+1:]
		} else {
			alias = path
		}
		if strings.HasSuffix(path, "/internal/model") {
			modelAliases[alias] = struct{}{}
		}
		if strings.HasSuffix(path, "/internal/reservation") || strings.HasSuffix(path, "/internal/repository") {
			sourceAliases[alias] = struct{}{}
		}
	}
	return modelAliases, sourceAliases
}

func appointmentImportPaths(file *ast.File) map[string]string {
	paths := make(map[string]string)
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		alias := ""
		if spec.Name != nil {
			alias = spec.Name.Name
		} else if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
			alias = path[slash+1:]
		} else {
			alias = path
		}
		if alias != "_" && alias != "." {
			paths[alias] = path
		}
	}
	return paths
}

func collectReservationVariables(
	file *ast.File,
	modelAliases, reservationTypeNames, reservationFactories map[string]struct{},
) map[string]struct{} {
	vars := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			collectReservationFieldNames(vars, n.Recv, modelAliases, reservationTypeNames)
			collectReservationFieldNames(vars, n.Type.Params, modelAliases, reservationTypeNames)
		case *ast.FuncLit:
			collectReservationFieldNames(vars, n.Type.Params, modelAliases, reservationTypeNames)
		case *ast.ValueSpec:
			if isReservationType(n.Type, modelAliases, reservationTypeNames) {
				for _, name := range n.Names {
					vars[name.Name] = struct{}{}
				}
			}
			for i, value := range n.Values {
				if exprIsReservationValue(value, modelAliases, reservationTypeNames, vars, reservationFactories) && i < len(n.Names) {
					vars[n.Names[i].Name] = struct{}{}
				}
			}
		case *ast.AssignStmt:
			for i, value := range n.Rhs {
				if !exprIsReservationValue(value, modelAliases, reservationTypeNames, vars, reservationFactories) || i >= len(n.Lhs) {
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

func collectReservationFieldNames(
	vars map[string]struct{},
	fields *ast.FieldList,
	modelAliases, reservationTypeNames map[string]struct{},
) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		if !isReservationType(field.Type, modelAliases, reservationTypeNames) {
			continue
		}
		for _, name := range field.Names {
			vars[name.Name] = struct{}{}
		}
	}
}

func collectAppointmentQueryVariables(
	file *ast.File,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories map[string]struct{},
	staticStrings, staticStringFactories map[string]string,
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
				if receiverTargetsAppointments(value, modelAliases, reservationTypeNames, reservationVars, reservationFactories, queryVars, staticStrings, staticStringFactories) ||
					exprIsReservationValue(value, modelAliases, reservationTypeNames, reservationVars, reservationFactories) {
					if _, exists := queryVars[names[i].Name]; !exists {
						queryVars[names[i].Name] = struct{}{}
						changed = true
					}
				}
			}
			return true
		})
	}
	return queryVars
}

func collectAppointmentTableVariablesInto(
	file *ast.File,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories map[string]struct{},
	staticStrings, staticStringFactories map[string]string,
) {
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
				targetsAppointments := exprIsReservationTableName(value, modelAliases, reservationTypeNames, reservationVars, reservationFactories)
				if !targetsAppointments {
					table, ok := staticStringWithFactories(value, staticStrings, staticStringFactories)
					targetsAppointments = ok && isAppointmentsTable(table)
				}
				if !targetsAppointments || isAppointmentsTable(staticStrings[names[i].Name]) {
					continue
				}
				staticStrings[names[i].Name] = "appointments"
				changed = true
			}
			return true
		})
	}
}

func receiverTargetsAppointments(
	expr ast.Expr,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories, queryVars map[string]struct{},
	staticStrings, staticStringFactories map[string]string,
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
		if sel.Sel.Name == "Model" && len(n.Args) > 0 && exprIsReservationValue(n.Args[0], modelAliases, reservationTypeNames, reservationVars, reservationFactories) {
			return true
		}
		if sel.Sel.Name == "Table" && len(n.Args) > 0 {
			if table, ok := staticStringWithFactories(n.Args[0], staticStrings, staticStringFactories); ok && isAppointmentsTable(table) {
				return true
			}
			if exprIsReservationTableName(n.Args[0], modelAliases, reservationTypeNames, reservationVars, reservationFactories) {
				return true
			}
		}
		return receiverTargetsAppointments(sel.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories, queryVars, staticStrings, staticStringFactories)
	case *ast.SelectorExpr:
		return receiverTargetsAppointments(n.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories, queryVars, staticStrings, staticStringFactories)
	case *ast.ParenExpr:
		return receiverTargetsAppointments(n.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories, queryVars, staticStrings, staticStringFactories)
	default:
		return false
	}
}

func isAppointmentsTable(table string) bool {
	fields := strings.Fields(strings.TrimSpace(strings.ToLower(table)))
	if len(fields) == 0 {
		return false
	}
	tableName := fields[0]
	if dot := strings.LastIndexByte(tableName, '.'); dot >= 0 {
		tableName = tableName[dot+1:]
	}
	return strings.Trim(tableName, `"`) == "appointments"
}

func exprIsReservationTableName(
	expr ast.Expr,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories map[string]struct{},
) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "TableName" {
		return false
	}
	return exprIsReservationValue(selector.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories)
}

func isBroadReservationSourceType(name string) bool {
	switch name {
	case "ReservationCRUDRepository", "ReservationIntentRepository", "ReservationRepository", "ReservationStore":
		return true
	default:
		return false
	}
}

func callArgsContainReservation(
	args []ast.Expr,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories map[string]struct{},
) bool {
	for _, arg := range args {
		if exprIsReservationValue(arg, modelAliases, reservationTypeNames, reservationVars, reservationFactories) {
			return true
		}
	}
	return false
}

func exprIsReservationValue(
	expr ast.Expr,
	modelAliases, reservationTypeNames, reservationVars, reservationFactories map[string]struct{},
) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		_, ok := reservationVars[n.Name]
		return ok
	case *ast.UnaryExpr:
		return exprIsReservationValue(n.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories)
	case *ast.ParenExpr:
		return exprIsReservationValue(n.X, modelAliases, reservationTypeNames, reservationVars, reservationFactories)
	case *ast.CompositeLit:
		return isReservationType(n.Type, modelAliases, reservationTypeNames)
	case *ast.CallExpr:
		switch fn := n.Fun.(type) {
		case *ast.Ident:
			if fn.Name == "new" && len(n.Args) == 1 && isReservationType(n.Args[0], modelAliases, reservationTypeNames) {
				return true
			}
			_, ok := reservationFactories[fn.Name]
			return ok
		case *ast.SelectorExpr:
			if packageAlias, ok := fn.X.(*ast.Ident); ok {
				if _, found := reservationFactories[packageAlias.Name+"."+fn.Sel.Name]; found {
					return true
				}
			}
			_, ok := reservationFactories[reservationFactoryCallKey(n.Pos())]
			return ok
		default:
			return false
		}
	default:
		return false
	}
}

func isReservationType(
	expr ast.Expr,
	modelAliases, reservationTypeNames map[string]struct{},
) bool {
	switch n := expr.(type) {
	case *ast.StarExpr:
		return isReservationType(n.X, modelAliases, reservationTypeNames)
	case *ast.ArrayType:
		return isReservationType(n.Elt, modelAliases, reservationTypeNames)
	case *ast.Ellipsis:
		return isReservationType(n.Elt, modelAliases, reservationTypeNames)
	case *ast.ParenExpr:
		return isReservationType(n.X, modelAliases, reservationTypeNames)
	case *ast.Ident:
		_, ok := reservationTypeNames[n.Name]
		return ok
	case *ast.SelectorExpr:
		alias, ok := n.X.(*ast.Ident)
		if !ok || n.Sel.Name != "Reservation" {
			return false
		}
		_, ok = modelAliases[alias.Name]
		return ok
	default:
		return false
	}
}

func collectReservationFactoriesInto(
	file *ast.File,
	modelAliases, reservationTypeNames map[string]struct{},
	factories map[string]struct{},
	methodFactories map[string]map[string]struct{},
) {
	for _, decl := range file.Decls {
		switch typedDecl := decl.(type) {
		case *ast.FuncDecl:
			if !funcTypeReturnsReservation(typedDecl.Type, modelAliases, reservationTypeNames) {
				continue
			}
			if typedDecl.Recv == nil {
				factories[typedDecl.Name.Name] = struct{}{}
				continue
			}
			receiverType := firstFieldTypeName(typedDecl.Recv)
			addNestedStringSetValue(methodFactories, receiverType, typedDecl.Name.Name)
		case *ast.GenDecl:
			if typedDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range typedDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				iface, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				for _, method := range iface.Methods.List {
					fn, ok := method.Type.(*ast.FuncType)
					if !ok || !funcTypeReturnsReservation(fn, modelAliases, reservationTypeNames) {
						continue
					}
					for _, name := range method.Names {
						addNestedStringSetValue(methodFactories, typeSpec.Name.Name, name.Name)
					}
				}
			}
		}
	}
}

func reservationFactoryCallKey(pos token.Pos) string {
	return fmt.Sprintf("call@%d", pos)
}

func collectFunctionReturnTypesInto(file *ast.File, returnTypes map[string]string) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
			continue
		}
		if returnType := baseTypeName(fn.Type.Results.List[0].Type); returnType != "" {
			returnTypes[fn.Name.Name] = returnType
		}
	}
}

func collectReservationFactoryCallPositionsInto(
	file *ast.File,
	methodFactories map[string]map[string]struct{},
	structFieldTypes map[string]map[string]string,
	functionReturnTypes map[string]string,
	factories map[string]struct{},
) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		valueTypes := make(map[string]string)
		collectFieldValueTypesInto(fn.Recv, valueTypes)
		collectFieldValueTypesInto(fn.Type.Params, valueTypes)
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.ValueSpec:
				typeName := baseTypeName(n.Type)
				for i, name := range n.Names {
					inferredType := typeName
					if inferredType == "" && i < len(n.Values) {
						inferredType = expressionStaticType(n.Values[i], valueTypes, structFieldTypes, functionReturnTypes)
					}
					if inferredType != "" {
						valueTypes[name.Name] = inferredType
					}
				}
			case *ast.AssignStmt:
				for i, value := range n.Rhs {
					if i >= len(n.Lhs) {
						continue
					}
					name, ok := n.Lhs[i].(*ast.Ident)
					if !ok {
						continue
					}
					if typeName := expressionStaticType(value, valueTypes, structFieldTypes, functionReturnTypes); typeName != "" {
						valueTypes[name.Name] = typeName
					}
				}
			case *ast.CallExpr:
				selector, ok := n.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				receiverType := expressionStaticType(selector.X, valueTypes, structFieldTypes, functionReturnTypes)
				if _, ok := methodFactories[receiverType][selector.Sel.Name]; ok {
					factories[reservationFactoryCallKey(n.Pos())] = struct{}{}
				}
			}
			return true
		})
	}
}

func collectStructFieldTypesInto(file *ast.File, structFieldTypes map[string]map[string]string) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				fieldType := baseTypeName(field.Type)
				for _, name := range field.Names {
					addNestedStringMapValue(structFieldTypes, typeSpec.Name.Name, name.Name, fieldType)
				}
			}
		}
	}
}

func collectFieldValueTypesInto(fields *ast.FieldList, values map[string]string) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		typeName := baseTypeName(field.Type)
		for _, name := range field.Names {
			values[name.Name] = typeName
		}
	}
}

func expressionStaticType(
	expr ast.Expr,
	valueTypes map[string]string,
	structFieldTypes map[string]map[string]string,
	functionReturnTypes map[string]string,
) string {
	switch n := expr.(type) {
	case *ast.Ident:
		return valueTypes[n.Name]
	case *ast.ParenExpr:
		return expressionStaticType(n.X, valueTypes, structFieldTypes, functionReturnTypes)
	case *ast.UnaryExpr:
		return expressionStaticType(n.X, valueTypes, structFieldTypes, functionReturnTypes)
	case *ast.CompositeLit:
		return baseTypeName(n.Type)
	case *ast.SelectorExpr:
		receiverType := expressionStaticType(n.X, valueTypes, structFieldTypes, functionReturnTypes)
		return structFieldTypes[receiverType][n.Sel.Name]
	case *ast.CallExpr:
		switch called := n.Fun.(type) {
		case *ast.Ident:
			if called.Name == "new" && len(n.Args) == 1 {
				return baseTypeName(n.Args[0])
			}
			return functionReturnTypes[called.Name]
		case *ast.SelectorExpr:
			if packageAlias, ok := called.X.(*ast.Ident); ok {
				return functionReturnTypes[packageAlias.Name+"."+called.Sel.Name]
			}
		}
	}
	return ""
}

func firstFieldTypeName(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	return baseTypeName(fields.List[0].Type)
}

func baseTypeName(expr ast.Expr) string {
	switch n := expr.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return n.Name
	case *ast.StarExpr:
		return baseTypeName(n.X)
	case *ast.ParenExpr:
		return baseTypeName(n.X)
	case *ast.IndexExpr:
		return baseTypeName(n.X)
	case *ast.IndexListExpr:
		return baseTypeName(n.X)
	case *ast.SelectorExpr:
		if alias, ok := n.X.(*ast.Ident); ok {
			return alias.Name + "." + n.Sel.Name
		}
	}
	return ""
}

func appointmentCapabilityName(name string) bool {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "appointment") {
		return true
	}
	switch lower {
	case "reservationrepository", "reservationcrudrepository", "reservationstore",
		"reservationwriter", "reservationupdater", "reservationmutator":
		return true
	default:
		return false
	}
}

func fieldListTargetsAppointmentCapability(fields *ast.FieldList) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if typeExprTargetsAppointmentCapability(field.Type) {
			return true
		}
	}
	return false
}

func typeExprTargetsAppointmentCapability(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		return appointmentCapabilityName(n.Name)
	case *ast.StarExpr:
		return typeExprTargetsAppointmentCapability(n.X)
	case *ast.IndexExpr:
		return typeExprTargetsAppointmentCapability(n.X)
	case *ast.IndexListExpr:
		return typeExprTargetsAppointmentCapability(n.X)
	case *ast.SelectorExpr:
		return appointmentCapabilityName(n.Sel.Name)
	default:
		return false
	}
}

func collectStaticStringFactoriesInto(
	files []*ast.File,
	constants map[string]string,
	factories map[string]string,
) {
	changed := true
	for changed {
		changed = false
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || fn.Body == nil || fn.Type.Params.NumFields() != 0 || len(fn.Body.List) != 1 {
					continue
				}
				ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
				if !ok || len(ret.Results) != 1 {
					continue
				}
				value, ok := staticStringWithFactories(ret.Results[0], constants, factories)
				if !ok || factories[fn.Name.Name] == value {
					continue
				}
				factories[fn.Name.Name] = value
				changed = true
			}
		}
	}
}

func fieldListContainsMapStringAny(
	fields *ast.FieldList,
	mapStringAnyTypeNames map[string]struct{},
) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if isMapStringAnyType(field.Type, mapStringAnyTypeNames) {
			return true
		}
	}
	return false
}

func isMapStringAnyType(expr ast.Expr, namedTypes map[string]struct{}) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		_, ok := namedTypes[n.Name]
		return ok
	case *ast.StarExpr:
		return isMapStringAnyType(n.X, namedTypes)
	case *ast.ParenExpr:
		return isMapStringAnyType(n.X, namedTypes)
	case *ast.MapType:
		key, keyOK := n.Key.(*ast.Ident)
		if !keyOK || key.Name != "string" {
			return false
		}
		switch value := n.Value.(type) {
		case *ast.Ident:
			return value.Name == "any"
		case *ast.InterfaceType:
			return value.Methods != nil && len(value.Methods.List) == 0
		default:
			return false
		}
	default:
		return false
	}
}

func collectAppointmentNamedTypeFactsInto(
	files []*ast.File,
	reservationTypeNames, mapStringAnyTypeNames map[string]struct{},
) {
	changed := true
	for changed {
		changed = false
		for _, file := range files {
			modelAliases, _ := appointmentImportAliases(file)
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if isReservationType(typeSpec.Type, modelAliases, reservationTypeNames) {
						if _, exists := reservationTypeNames[typeSpec.Name.Name]; !exists {
							reservationTypeNames[typeSpec.Name.Name] = struct{}{}
							changed = true
						}
					}
					if isMapStringAnyType(typeSpec.Type, mapStringAnyTypeNames) {
						if _, exists := mapStringAnyTypeNames[typeSpec.Name.Name]; !exists {
							mapStringAnyTypeNames[typeSpec.Name.Name] = struct{}{}
							changed = true
						}
					}
				}
			}
		}
	}
}

func cloneStringSet(source map[string]struct{}) map[string]struct{} {
	clone := make(map[string]struct{}, len(source))
	for value := range source {
		clone[value] = struct{}{}
	}
	return clone
}

func cloneStringMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneNestedStringSet(source map[string]map[string]struct{}) map[string]map[string]struct{} {
	clone := make(map[string]map[string]struct{}, len(source))
	for key, values := range source {
		clone[key] = cloneStringSet(values)
	}
	return clone
}

func addNestedStringSetValue(target map[string]map[string]struct{}, key, value string) {
	if key == "" || value == "" {
		return
	}
	if target[key] == nil {
		target[key] = make(map[string]struct{})
	}
	target[key][value] = struct{}{}
}

func cloneNestedStringMap(source map[string]map[string]string) map[string]map[string]string {
	clone := make(map[string]map[string]string, len(source))
	for key, values := range source {
		clone[key] = make(map[string]string, len(values))
		for nestedKey, value := range values {
			clone[key][nestedKey] = value
		}
	}
	return clone
}

func addNestedStringMapValue(target map[string]map[string]string, key, nestedKey, value string) {
	if key == "" || nestedKey == "" || value == "" {
		return
	}
	if target[key] == nil {
		target[key] = make(map[string]string)
	}
	target[key][nestedKey] = value
}

func stringLiteral(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func appointmentPackagePath(filePath string) string {
	if index := strings.LastIndex(filePath, "/"); index >= 0 {
		return filePath[:index]
	}
	return ""
}

func collectAppointmentPackageFacts(t testing.TB, sources map[string][]byte) map[string]appointmentPackageFacts {
	t.Helper()
	parsedByPackage := make(map[string][]*ast.File)
	for filePath, source := range sources {
		file, err := parser.ParseFile(token.NewFileSet(), filePath, source, 0)
		if err != nil {
			t.Fatalf("parse %s while collecting appointment write-owner constants: %v", filePath, err)
		}
		packagePath := appointmentPackagePath(filePath)
		parsedByPackage[packagePath] = append(parsedByPackage[packagePath], file)
	}

	result := make(map[string]appointmentPackageFacts, len(parsedByPackage))
	exportedReservationFactories := make(map[string]map[string]struct{}, len(parsedByPackage))
	exportedReservationMethodFactories := make(map[string]map[string]map[string]struct{}, len(parsedByPackage))
	exportedFunctionReturnTypes := make(map[string]map[string]string, len(parsedByPackage))
	for packagePath, files := range parsedByPackage {
		values := make(map[string]string)
		changed := true
		for changed {
			changed = false
			for _, file := range files {
				if collectStaticStringsInto(file, values) {
					changed = true
				}
			}
		}
		reservationTypeNames := make(map[string]struct{})
		mapStringAnyTypeNames := make(map[string]struct{})
		collectAppointmentNamedTypeFactsInto(files, reservationTypeNames, mapStringAnyTypeNames)
		reservationFactories := make(map[string]struct{})
		reservationMethodFactories := make(map[string]map[string]struct{})
		functionReturnTypes := make(map[string]string)
		structFieldTypes := make(map[string]map[string]string)
		declaredTypeNames := make(map[string]struct{})
		exportedFactories := make(map[string]struct{})
		for _, file := range files {
			modelAliases, _ := appointmentImportAliases(file)
			collectReservationFactoriesInto(
				file,
				modelAliases,
				reservationTypeNames,
				reservationFactories,
				reservationMethodFactories,
			)
			collectFunctionReturnTypesInto(file, functionReturnTypes)
			collectStructFieldTypesInto(file, structFieldTypes)
			for _, decl := range file.Decls {
				if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
					for _, spec := range gen.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							declaredTypeNames[typeSpec.Name.Name] = struct{}{}
						}
					}
				}
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || !ast.IsExported(fn.Name.Name) ||
					!funcTypeReturnsReservation(fn.Type, modelAliases, reservationTypeNames) {
					continue
				}
				exportedFactories[fn.Name.Name] = struct{}{}
			}
		}
		importPath := backendInternalImportPrefix + packagePath
		exportedReservationFactories[importPath] = exportedFactories
		exportedMethods := make(map[string]map[string]struct{})
		for receiverType, methods := range reservationMethodFactories {
			if !ast.IsExported(receiverType) {
				continue
			}
			for method := range methods {
				if ast.IsExported(method) {
					addNestedStringSetValue(exportedMethods, receiverType, method)
				}
			}
		}
		exportedReservationMethodFactories[importPath] = exportedMethods
		exportedReturns := make(map[string]string)
		for function, returnType := range functionReturnTypes {
			if _, declaredHere := declaredTypeNames[returnType]; declaredHere && ast.IsExported(function) && ast.IsExported(returnType) {
				exportedReturns[function] = returnType
			}
		}
		exportedFunctionReturnTypes[importPath] = exportedReturns
		staticStringFactories := make(map[string]string)
		collectStaticStringFactoriesInto(files, values, staticStringFactories)
		result[packagePath] = appointmentPackageFacts{
			staticStrings:              values,
			reservationFactories:       reservationFactories,
			reservationMethodFactories: reservationMethodFactories,
			functionReturnTypes:        functionReturnTypes,
			staticStringFactories:      staticStringFactories,
			reservationTypeNames:       reservationTypeNames,
			mapStringAnyTypeNames:      mapStringAnyTypeNames,
			structFieldTypes:           structFieldTypes,
		}
	}
	for packagePath, facts := range result {
		facts.importedReservationFactories = exportedReservationFactories
		facts.importedReservationMethodFactories = exportedReservationMethodFactories
		facts.importedFunctionReturnTypes = exportedFunctionReturnTypes
		result[packagePath] = facts
	}
	return result
}

func collectStaticStrings(file *ast.File) map[string]string {
	values := make(map[string]string)
	changed := true
	for changed {
		changed = collectStaticStringsInto(file, values)
	}
	return values
}

func collectStaticStringsInto(file *ast.File, values map[string]string) bool {
	changed := false
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
				value, ok := staticString(valueSpec.Values[i], values)
				if !ok || values[name.Name] == value {
					continue
				}
				values[name.Name] = value
				changed = true
			}
		}
	}
	return changed
}

func staticStringNodeWithFactories(
	node ast.Node,
	constants, factories map[string]string,
) (string, bool) {
	expr, ok := node.(ast.Expr)
	if !ok {
		return "", false
	}
	return staticStringWithFactories(expr, constants, factories)
}

func staticString(expr ast.Expr, constants map[string]string) (string, bool) {
	return staticStringWithFactories(expr, constants, nil)
}

func staticStringWithFactories(
	expr ast.Expr,
	constants, factories map[string]string,
) (string, bool) {
	switch n := expr.(type) {
	case *ast.BasicLit:
		return stringLiteral(n)
	case *ast.Ident:
		if value, ok := constants[n.Name]; ok {
			return value, true
		}
		if n.Obj == nil || n.Obj.Kind != ast.Con {
			return "", false
		}
		valueSpec, ok := n.Obj.Decl.(*ast.ValueSpec)
		if !ok {
			return "", false
		}
		for i, name := range valueSpec.Names {
			if name.Name != n.Name || i >= len(valueSpec.Values) {
				continue
			}
			return staticStringWithFactories(valueSpec.Values[i], constants, factories)
		}
		return "", false
	case *ast.ParenExpr:
		return staticStringWithFactories(n.X, constants, factories)
	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			return "", false
		}
		left, leftOK := staticStringWithFactories(n.X, constants, factories)
		right, rightOK := staticStringWithFactories(n.Y, constants, factories)
		if !leftOK || !rightOK {
			return "", false
		}
		return left + right, true
	case *ast.CallExpr:
		if len(n.Args) != 0 {
			return "", false
		}
		name, ok := n.Fun.(*ast.Ident)
		if !ok {
			return "", false
		}
		value, ok := factories[name.Name]
		return value, ok
	default:
		return "", false
	}
}

func isReservationFacadePath(path string) bool {
	return path == "repository/reservation_repository.go" || path == "repository/repositories.go"
}
