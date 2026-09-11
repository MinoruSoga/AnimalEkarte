package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
)

// This is a reviewed inventory, not a generated allowlist. New routes need an
// explicit row; never reconstruct expected entries from the live router.
//
//go:embed testdata/get_head_permissions.json
var getHEADInventoryJSON []byte

type getHEADInventoryEntry struct {
	Method          string `json:"method"`
	Path            string `json:"path"`
	Handler         string `json:"handler"`
	Class           string `json:"class"`
	Resource        string `json:"resource"`
	Action          string `json:"action"`
	Source          string `json:"source"`
	RouteExpression string `json:"routeExpression"`
	HandlerSource   string `json:"handlerSource"`
	Marker          string `json:"marker"`
	Authorization   string `json:"authorization"`
	Scope           string `json:"scope"`
	Regression      string `json:"regression"`
	Verification    string `json:"verification"`
}

func readGETHEADInventory(t *testing.T) []getHEADInventoryEntry {
	t.Helper()
	var entries []getHEADInventoryEntry
	decoder := json.NewDecoder(strings.NewReader(string(getHEADInventoryJSON)))
	decoder.DisallowUnknownFields()
	require.NoError(t, decoder.Decode(&entries))
	require.NotEmpty(t, entries)
	seen := make(map[string]bool)
	for _, entry := range entries {
		key := entry.Method + " " + entry.Path
		require.False(t, seen[key], "duplicate inventory row: %s", key)
		seen[key] = true
		require.Contains(t, []string{"GET", "HEAD"}, entry.Method, key)
		require.Contains(t, []string{"public", "liff", "self", "clinic-fixed", "cross-clinic", "shared-master"}, entry.Class, key)
		for field, value := range map[string]string{
			"handler": entry.Handler, "resource": entry.Resource, "action": entry.Action,
			"source": entry.Source, "handlerSource": entry.HandlerSource, "marker": entry.Marker,
			"authorization": entry.Authorization, "scope": entry.Scope,
			"regression": entry.Regression, "verification": entry.Verification,
		} {
			require.NotEmpty(t, value, "%s: missing %s", key, field)
		}
	}
	return entries
}

func registeredGETHEADRoutes(t *testing.T, storage string) gin.RoutesInfo {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Setenv("STORAGE_TYPE", storage)
	t.Setenv("SCHEDULER_INTERNAL_TOKEN", "test-scheduler-internal-token-32b!!")
	composition := newRuntimeComposition(runtimeCompositionDependencies{
		Config: &config.Config{JWTSecret: "test-secret-for-route-registration"},
	})
	router := gin.New()
	require.NoError(t, composition.registerRoutes(ctx, router, nil, storage == "s3"))
	return router.Routes()
}

func compareGETHEADInventory(routes gin.RoutesInfo, entries []getHEADInventoryEntry, storage string) error {
	expected := make(map[string]string, len(entries))
	for _, entry := range entries {
		if storage == "s3" && entry.Path == "/uploads/*filepath" {
			continue
		}
		expected[entry.Method+" "+entry.Path] = entry.Handler
	}
	for _, route := range routes {
		if route.Method != http.MethodGet && route.Method != http.MethodHead {
			continue
		}
		key := route.Method + " " + route.Path
		handler, found := expected[key]
		if !found {
			return fmt.Errorf("unreviewed GET/HEAD route: %s", key)
		}
		if handler != route.Handler {
			return fmt.Errorf("handler changed for %s: want %s, got %s", key, handler, route.Handler)
		}
		delete(expected, key)
	}
	if len(expected) != 0 {
		return fmt.Errorf("inventory contains removed GET/HEAD routes: %v", expected)
	}
	return nil
}

func TestGETHEADRoutesAreClassifiedForSelectedClinicGrant(t *testing.T) {
	entries := readGETHEADInventory(t)
	for _, storage := range []string{"", "s3"} {
		t.Run("storage="+storage, func(t *testing.T) {
			require.NoError(t, compareGETHEADInventory(registeredGETHEADRoutes(t, storage), entries, storage))
		})
	}
}

func TestGETHEADInventoryRejectsRouteDrift(t *testing.T) {
	entries := []getHEADInventoryEntry{{Method: "GET", Path: "/api/v1/owners", Handler: "owner.List"}}
	for _, tc := range []struct {
		name   string
		routes gin.RoutesInfo
	}{
		{"known-prefix-addition", gin.RoutesInfo{{Method: "GET", Path: "/api/v1/owners", Handler: "owner.List"}, {Method: "GET", Path: "/api/v1/owners/new-unreviewed-route", Handler: "owner.New"}}},
		{"removal", nil},
		{"method-change", gin.RoutesInfo{{Method: "HEAD", Path: "/api/v1/owners", Handler: "owner.List"}}},
		{"handler-change", gin.RoutesInfo{{Method: "GET", Path: "/api/v1/owners", Handler: "owner.UnscopedList"}}},
		{"write-method-change", gin.RoutesInfo{{Method: "POST", Path: "/api/v1/owners", Handler: "owner.List"}}},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Error(t, compareGETHEADInventory(tc.routes, entries, "")) })
	}
	require.NoError(t, compareGETHEADInventory(gin.RoutesInfo{{Method: "GET", Path: "/api/v1/owners", Handler: "owner.List"}}, entries, ""))
}

// Source references prevent stale/missing evidence links; they do not prove
// middleware execution or returned-data isolation. Those require the referenced
// handler tests and DB/API scenarios, tracked separately in Verification.
func TestGETHEADInventoryEvidenceReferences(t *testing.T) {
	for _, entry := range readGETHEADInventory(t) {
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			registration, err := os.ReadFile(filepath.Join("../..", entry.Source))
			require.NoError(t, err)
			if entry.RouteExpression != "" {
				require.Contains(t, compactGETHEADSource(string(registration)), compactGETHEADSource(entry.RouteExpression), "route registration/permission changed; review inventory")
			}
			handler, err := os.ReadFile(filepath.Join("../..", entry.HandlerSource))
			require.NoError(t, err)
			if _, method, found := strings.Cut(entry.Handler, ")."); found && strings.HasSuffix(entry.Handler, "-fm") {
				method = strings.TrimSuffix(method, "-fm")
				fset := token.NewFileSet()
				parsed, err := parser.ParseFile(fset, entry.HandlerSource, handler, 0)
				require.NoError(t, err)
				var declaration *ast.FuncDecl
				for _, decl := range parsed.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == method {
						declaration = fn
						break
					}
				}
				require.NotNil(t, declaration, "handler source reference disappeared")
				body := handler[fset.Position(declaration.Pos()).Offset:fset.Position(declaration.End()).Offset]
				require.Contains(t, string(body), entry.Marker, "handler authorization/scope evidence disappeared")
			} else {
				require.Contains(t, string(handler), entry.Marker)
			}
			file, symbol, ok := strings.Cut(entry.Regression, "#")
			require.True(t, ok, "expected file#TestSymbol reference")
			parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join("../..", file), nil, 0)
			require.NoError(t, err)
			found := false
			for _, decl := range parsed.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == symbol {
					found = true
				}
			}
			require.True(t, found, "regression reference missing: %s", entry.Regression)
		})
	}
}

func TestGETHEADInventoryVerificationMentionsExecutedTests(t *testing.T) {
	for _, entry := range readGETHEADInventory(t) {
		key := entry.Method + " " + entry.Path
		switch entry.Class {
		case "clinic-fixed":
			require.Contains(t, entry.Verification, "TestGETHEADClinicFixedSelectedClinicWithoutGrantIsForbidden", key)
			require.Contains(t, entry.Verification, "TestGETHEADClinicFixedSelectedClinicWithGrantAllowsHandler", key)
		case "cross-clinic":
			require.Contains(t, entry.Verification, "TestGETHEADCrossClinicAllowingDoesNotGrantSelectedClinicB", key)
		case "public", "liff", "self":
			require.Contains(t, entry.Verification, "contract-exclusion-from-selected-clinic-grant", key)
		case "shared-master":
			require.Contains(t, entry.Verification, "shared-master-contract", key)
		}
	}
}

func compactGETHEADSource(source string) string {
	// Registration expressions can be split across lines with trailing commas.
	return strings.TrimSuffix(strings.ReplaceAll(strings.Join(strings.Fields(source), ""), ",)", ")"), ";")
}
