package lintscan

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// domainImportAllowlist is the production-import allowlist for ARCH-A6 / ADR-006 /
// docs/architecture/be9-2a-boundary-map.md §5.
//
// Key = importer domain package (top-level under internal/).
// Value = domain packages that importer may import via a real Go import path.
//
// Rules encoded here:
//   - Only domain↔domain edges are gated. Cross-cutting packages (model, apperrors,
//     audit, persistence, sharedkernel, middleware, …) are free.
//   - Cycle-resolved reverse edges must stay as consumer-side interfaces (no Go import).
//     Examples: medicalrecord↛lstep, billing↛medicalrecord, reservation↛trimming,
//     owner↛billing, clinic↛auth.
//   - Adding a new edge requires updating this allowlist AND boundary map §5 / ADR-006
//     in the same PR (see boundary map §5.2).
//
// identitylink is included as the 14th target package (post-BE9 recensus).
var domainImportAllowlist = map[string]map[string]struct{}{
	"httpapi":       {},
	"clinic":        {"httpapi": {}},
	"inventory":     {"clinic": {}, "httpapi": {}},
	"manualarticle": {"httpapi": {}},
	"identitylink":  {"httpapi": {}},
	"owner":         {"clinic": {}, "httpapi": {}},
	"pet":           {"owner": {}, "clinic": {}, "httpapi": {}},
	"staff":         {"clinic": {}, "httpapi": {}},
	"auth":          {"staff": {}, "clinic": {}, "httpapi": {}},
	"reservation": {
		"owner": {}, "pet": {}, "staff": {}, "clinic": {}, "httpapi": {},
	},
	"trimming": {
		"reservation": {}, "pet": {}, "clinic": {}, "httpapi": {},
	},
	"billing": {
		"owner": {}, "reservation": {}, "trimming": {}, "inventory": {},
		"clinic": {}, "staff": {}, "httpapi": {},
	},
	"medicalrecord": {
		"owner": {}, "pet": {}, "staff": {}, "reservation": {}, "billing": {},
		"clinic": {}, "httpapi": {},
	},
	"lstep": {
		"owner": {}, "pet": {}, "staff": {}, "reservation": {}, "medicalrecord": {},
		"billing": {}, "clinic": {}, "httpapi": {},
	},
}

const backendInternalImportPrefix = "github.com/animal-ekarte/backend/internal/"

// TestDomainImportAllowlistLint fails when a production domain package imports another
// domain package outside domainImportAllowlist. Scoped: domain packages only; production
// .go files only (_test.go excluded by WalkInternalTree).
func TestDomainImportAllowlistLint(t *testing.T) {
	t.Run("allowlist_is_acyclic", func(t *testing.T) {
		if cycle := findAllowlistCycle(domainImportAllowlist); cycle != "" {
			t.Fatalf("domainImportAllowlist contains a cycle: %s", cycle)
		}
	})

	t.Run("allowlist_keys_match_domain_packages", func(t *testing.T) {
		for name := range domainPackages {
			if _, ok := domainImportAllowlist[name]; !ok {
				t.Errorf("domain package %q missing from domainImportAllowlist", name)
			}
		}
		for name := range domainImportAllowlist {
			if _, ok := domainPackages[name]; !ok {
				t.Errorf("domainImportAllowlist key %q is not in domainPackages", name)
			}
		}
	})

	t.Run("real_tree", func(t *testing.T) {
		files := WalkInternalTreeT(t)
		if len(files) < 500 {
			t.Fatalf("domain import lint discovered only %d production Go files; whole-tree coverage is not proven", len(files))
		}
		violations := findDomainImportViolations(files, domainImportAllowlist, domainPackages)
		if len(violations) > 0 {
			t.Fatalf("domain import allowlist violations:\n%s", strings.Join(violations, "\n"))
		}
	})

	t.Run("mutation_forbidden_edge_fails", func(t *testing.T) {
		files := map[string][]byte{
			"medicalrecord/bad.go": []byte(`package medicalrecord
import "github.com/animal-ekarte/backend/internal/lstep"
`),
		}
		violations := findDomainImportViolations(files, domainImportAllowlist, domainPackages)
		if len(violations) == 0 {
			t.Fatal("expected violation for medicalrecord -> lstep")
		}
		if !strings.Contains(violations[0], "medicalrecord") || !strings.Contains(violations[0], "lstep") {
			t.Fatalf("unexpected violation text: %v", violations)
		}
	})

	t.Run("mutation_allowed_edge_passes", func(t *testing.T) {
		files := map[string][]byte{
			"pet/ok.go": []byte(`package pet
import "github.com/animal-ekarte/backend/internal/owner"
`),
			"billing/ok.go": []byte(`package billing
import "github.com/animal-ekarte/backend/internal/reservation"
`),
		}
		violations := findDomainImportViolations(files, domainImportAllowlist, domainPackages)
		if len(violations) != 0 {
			t.Fatalf("expected 0 violations for allowed edges, got %v", violations)
		}
	})

	t.Run("mutation_non_domain_import_ignored", func(t *testing.T) {
		files := map[string][]byte{
			"owner/ok.go": []byte(`package owner
import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/apperrors"
)
`),
		}
		violations := findDomainImportViolations(files, domainImportAllowlist, domainPackages)
		if len(violations) != 0 {
			t.Fatalf("expected non-domain imports to be ignored, got %v", violations)
		}
	})

	t.Run("mutation_self_import_ignored", func(t *testing.T) {
		// Subpackage self-import of same domain top-level is not produced by normal
		// layout; ensure same-package path does not false-positive if path root matches.
		files := map[string][]byte{
			"owner/sub/x.go": []byte(`package sub
import "github.com/animal-ekarte/backend/internal/owner"
`),
		}
		// owner/sub is still under domain owner → importing owner is same top-level domain.
		violations := findDomainImportViolations(files, domainImportAllowlist, domainPackages)
		if len(violations) != 0 {
			t.Fatalf("same-domain import should be ignored, got %v", violations)
		}
	})
}

func findDomainImportViolations(
	files map[string][]byte,
	allowlist map[string]map[string]struct{},
	domains map[string]struct{},
) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var violations []string
	fset := token.NewFileSet()
	for _, relPath := range paths {
		importer, ok := domainTopLevel(relPath, domains)
		if !ok {
			continue
		}
		src := files[relPath]
		file, err := parser.ParseFile(fset, relPath, src, parser.ImportsOnly)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: parse imports: %v", relPath, err))
			continue
		}
		allowed := allowlist[importer]
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				violations = append(violations, fmt.Sprintf("%s: unquote import: %v", relPath, err))
				continue
			}
			dep, ok := domainImportTarget(importPath, domains)
			if !ok || dep == importer {
				continue
			}
			if _, ok := allowed[dep]; ok {
				continue
			}
			violations = append(violations, fmt.Sprintf(
				"%s: domain %q imports %q (not in allowlist; update boundary map §5 + domainImportAllowlist or use a consumer-side interface)",
				relPath, importer, dep,
			))
		}
	}
	sort.Strings(violations)
	return violations
}

func domainTopLevel(relPath string, domains map[string]struct{}) (string, bool) {
	slash := filepath.ToSlash(relPath)
	parts := strings.Split(slash, "/")
	if len(parts) < 1 {
		return "", false
	}
	top := parts[0]
	if _, ok := domains[top]; !ok {
		return "", false
	}
	return top, true
}

func domainImportTarget(importPath string, domains map[string]struct{}) (string, bool) {
	if !strings.HasPrefix(importPath, backendInternalImportPrefix) {
		return "", false
	}
	rest := strings.TrimPrefix(importPath, backendInternalImportPrefix)
	if rest == "" {
		return "", false
	}
	// Only the first path segment is the top-level package (subpackages share domain).
	seg := rest
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		seg = rest[:i]
	}
	if _, ok := domains[seg]; !ok {
		return "", false
	}
	return seg, true
}

// findAllowlistCycle returns a human-readable cycle path or "" if acyclic (Kahn).
func findAllowlistCycle(allowlist map[string]map[string]struct{}) string {
	nodes := make([]string, 0, len(allowlist))
	indegree := make(map[string]int, len(allowlist))
	for n := range allowlist {
		nodes = append(nodes, n)
		if _, ok := indegree[n]; !ok {
			indegree[n] = 0
		}
	}
	sort.Strings(nodes)
	for from, tos := range allowlist {
		_ = from
		for to := range tos {
			if _, ok := allowlist[to]; !ok {
				// Edge to unknown domain — treat as invalid cycle source for safety.
				return fmt.Sprintf("%s -> %s (target not in allowlist keys)", from, to)
			}
			indegree[to]++
		}
	}
	queue := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if indegree[n] == 0 {
			queue = append(queue, n)
		}
	}
	seen := 0
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		seen++
		for to := range allowlist[n] {
			indegree[to]--
			if indegree[to] == 0 {
				queue = append(queue, to)
			}
		}
	}
	if seen == len(nodes) {
		return ""
	}
	remaining := make([]string, 0)
	for _, n := range nodes {
		if indegree[n] > 0 {
			remaining = append(remaining, n)
		}
	}
	sort.Strings(remaining)
	return "nodes still have indegree: " + strings.Join(remaining, ", ")
}
