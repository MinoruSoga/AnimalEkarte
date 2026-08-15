package lintscan

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var acceptedTopLevelPackages = map[string]struct{}{
	"apicontract":   {},
	"apperrors":     {},
	"audit":         {},
	"auth":          {},
	"authjwt":       {},
	"billing":       {},
	"clinic":        {},
	"config":        {},
	"csvimport":     {},
	"dbconn":        {},
	"httpapi":       {},
	"identitylink":  {},
	"infra":         {},
	"inventory":     {},
	"lintscan":      {},
	"logger":        {},
	"lstep":         {},
	"manualarticle": {},
	"medicalrecord": {},
	"middleware":    {},
	"model":         {},
	"owner":         {},
	"persistence":   {},
	"pet":           {},
	"reservation":   {},
	"scheduler":     {},
	"seedbundle":    {},
	"sharedkernel":  {},
	"staff":         {},
	"testdb":        {},
	"textsearch":    {},
	"timeutil":      {},
	"trimming":      {},
}

var domainPackages = map[string]struct{}{
	"auth":          {},
	"billing":       {},
	"clinic":        {},
	"httpapi":       {},
	"identitylink":  {},
	"inventory":     {},
	"lstep":         {},
	"manualarticle": {},
	"medicalrecord": {},
	"owner":         {},
	"pet":           {},
	"reservation":   {},
	"staff":         {},
	"trimming":      {},
}

var retiredLayerPackages = map[string]struct{}{
	"handler":    {},
	"repository": {},
	"service":    {},
}

var layerPackageNames = map[string]struct{}{
	"handler":      {},
	"handlers":     {},
	"repositories": {},
	"repository":   {},
	"service":      {},
	"services":     {},
}

var bucketPackageNames = map[string]struct{}{
	"common":  {},
	"helpers": {},
	"shared":  {},
	"util":    {},
	"utils":   {},
}

type packageBoundaryFinding struct {
	check  string
	path   string
	detail string
}

func TestPackageBoundaryGate(t *testing.T) {
	t.Run("real_tree", func(t *testing.T) {
		backendRoot := realBackendRoot(t)
		findings, err := findPackageBoundaryViolations(backendRoot)
		requireCleanBoundaryTree(t, findings, err)
	})

	t.Run("accepted_and_bucket_sets_are_disjoint", func(t *testing.T) {
		requireSetSize(t, "accepted top-level packages", acceptedTopLevelPackages, 33)
		requireSetSize(t, "domain packages", domainPackages, 14)
		requireSetSize(t, "layer package names", layerPackageNames, 6)
		requireSetSize(t, "bucket package names", bucketPackageNames, 5)

		for name := range bucketPackageNames {
			if _, accepted := acceptedTopLevelPackages[name]; accepted {
				t.Errorf("bucket package %q is also accepted as a top-level package", name)
			}
		}
		for name := range domainPackages {
			if _, accepted := acceptedTopLevelPackages[name]; !accepted {
				t.Errorf("domain package %q is missing from the accepted top-level set", name)
			}
		}
	})

	t.Run("clean_tree", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		findings, err := findPackageBoundaryViolations(backendRoot)
		requireCleanBoundaryTree(t, findings, err)
	})

	t.Run("mutation_unapproved_top_level", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		writeBoundaryFixture(t, backendRoot, "internal/catchall/package_test.go", "package catchall\n")

		findings, err := findUnapprovedTopLevelPackages(filepath.Join(backendRoot, "internal"))
		requireBoundaryFinding(t, findings, err, "C1")
	})

	t.Run("mutation_retired_layer_resurrection", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		writeBoundaryFixture(t, backendRoot, "internal/service/resurrected_test.go", "package service\n")

		findings, err := findRetiredLayerViolations(backendRoot)
		requireBoundaryFinding(t, findings, err, "C2")
	})

	t.Run("mutation_retired_layer_import", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		writeBoundaryFixture(
			t,
			backendRoot,
			"cmd/api/main_test.go",
			"package main\n\nimport _ \"github.com/animal-ekarte/backend/internal/repository/example\"\n",
		)

		findings, err := findRetiredLayerViolations(backendRoot)
		requireBoundaryFinding(t, findings, err, "C2")
	})

	t.Run("mutation_domain_layer_subpackage", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		writeBoundaryFixture(t, backendRoot, "internal/owner/service/usecase_test.go", "package service\n")

		findings, err := findDomainLayerSubpackages(filepath.Join(backendRoot, "internal"))
		requireBoundaryFinding(t, findings, err, "C3")
	})

	t.Run("mutation_bucket_package", func(t *testing.T) {
		backendRoot := newSyntheticBackend(t)
		writeBoundaryFixture(t, backendRoot, "internal/lstep/util/helper_test.go", "package util\n")

		findings, err := findBucketPackages(filepath.Join(backendRoot, "internal"))
		requireBoundaryFinding(t, findings, err, "C4")
	})
}

func realBackendRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	root, err := FindModuleRoot(workingDirectory)
	if err != nil {
		t.Fatalf("find backend module root: %v", err)
	}
	return root
}

func newSyntheticBackend(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	writeBoundaryFixture(t, root, "internal/owner/owner.go", "package owner\n")
	writeBoundaryFixture(t, root, "internal/infra/line/client.go", "package line\n")
	writeBoundaryFixture(t, root, "internal/model/model_test.go", "package model\n")
	writeBoundaryFixture(t, root, "internal/lstep/lstep.go", "package lstep\n")
	writeBoundaryFixture(t, root, "internal/sharedkernel/kernel.go", "package sharedkernel\n")
	writeBoundaryFixture(
		t,
		root,
		"internal/auth/import_decoys.go",
		"package auth\n\n// github.com/animal-ekarte/backend/internal/service\n"+
			"const retiredPathExample = \"github.com/animal-ekarte/backend/internal/repository\"\n",
	)
	writeBoundaryFixture(t, root, "cmd/_archive/tool/main.go", "package main\n")
	return root
}

func writeBoundaryFixture(t *testing.T, backendRoot, relativePath, source string) {
	t.Helper()

	path := filepath.Join(backendRoot, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

func requireSetSize(t *testing.T, name string, set map[string]struct{}, expected int) {
	t.Helper()
	if len(set) != expected {
		t.Fatalf("%s has %d entries, want %d", name, len(set), expected)
	}
}

func requireCleanBoundaryTree(t *testing.T, findings []packageBoundaryFinding, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("check package boundaries: %v", err)
	}
	for _, finding := range findings {
		t.Errorf("%s %s: %s", finding.check, finding.path, finding.detail)
	}
}

func requireBoundaryFinding(
	t *testing.T,
	findings []packageBoundaryFinding,
	err error,
	expectedCheck string,
) {
	t.Helper()
	if err != nil {
		t.Fatalf("check package boundaries: %v", err)
	}
	for _, finding := range findings {
		if finding.check == expectedCheck {
			return
		}
	}
	t.Fatalf("expected a %s violation, got none", expectedCheck)
}

func findPackageBoundaryViolations(backendRoot string) ([]packageBoundaryFinding, error) {
	internalRoot := filepath.Join(backendRoot, "internal")
	checks := []struct {
		name string
		run  func() ([]packageBoundaryFinding, error)
	}{
		{name: "C1 top-level packages", run: func() ([]packageBoundaryFinding, error) {
			return findUnapprovedTopLevelPackages(internalRoot)
		}},
		{name: "C2 retired layers", run: func() ([]packageBoundaryFinding, error) {
			return findRetiredLayerViolations(backendRoot)
		}},
		{name: "C3 domain layer subpackages", run: func() ([]packageBoundaryFinding, error) {
			return findDomainLayerSubpackages(internalRoot)
		}},
		{name: "C4 bucket packages", run: func() ([]packageBoundaryFinding, error) {
			return findBucketPackages(internalRoot)
		}},
	}

	var findings []packageBoundaryFinding
	for _, check := range checks {
		checkFindings, err := check.run()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", check.name, err)
		}
		findings = append(findings, checkFindings...)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].check != findings[j].check {
			return findings[i].check < findings[j].check
		}
		if findings[i].path != findings[j].path {
			return findings[i].path < findings[j].path
		}
		return findings[i].detail < findings[j].detail
	})
	return findings, nil
}

func findUnapprovedTopLevelPackages(internalRoot string) ([]packageBoundaryFinding, error) {
	unapproved := make(map[string]struct{})
	err := walkBoundaryGoFiles(internalRoot, func(path string) error {
		relativePath, err := filepath.Rel(internalRoot, path)
		if err != nil {
			return fmt.Errorf("make %s relative to %s: %w", path, internalRoot, err)
		}
		parts := strings.Split(filepath.ToSlash(relativePath), "/")
		if len(parts) < 2 {
			return nil
		}
		topLevelPackage := parts[0]
		if _, accepted := acceptedTopLevelPackages[topLevelPackage]; !accepted {
			unapproved[topLevelPackage] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	findings := make([]packageBoundaryFinding, 0, len(unapproved))
	for name := range unapproved {
		findings = append(findings, packageBoundaryFinding{
			check:  "C1",
			path:   "internal/" + name,
			detail: "top-level package contains Go files but is not accepted by ADR-006",
		})
	}
	return findings, nil
}

func findRetiredLayerViolations(backendRoot string) ([]packageBoundaryFinding, error) {
	var findings []packageBoundaryFinding
	err := walkBoundaryGoFiles(backendRoot, func(path string) error {
		relativePath, err := filepath.Rel(backendRoot, path)
		if err != nil {
			return fmt.Errorf("make %s relative to %s: %w", path, backendRoot, err)
		}
		slashPath := filepath.ToSlash(relativePath)
		parts := strings.Split(slashPath, "/")
		if len(parts) >= 3 && parts[0] == "internal" {
			if _, retired := retiredLayerPackages[parts[1]]; retired {
				findings = append(findings, packageBoundaryFinding{
					check:  "C2",
					path:   slashPath,
					detail: "retired layer contains a Go file",
				})
			}
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse imports from %s: %w", path, err)
		}
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return fmt.Errorf("unquote import %s in %s: %w", importSpec.Path.Value, path, err)
			}
			if retiredLayer, found := importedRetiredLayer(importPath); found {
				findings = append(findings, packageBoundaryFinding{
					check:  "C2",
					path:   slashPath,
					detail: "imports retired internal/" + retiredLayer + " package: " + importPath,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func findDomainLayerSubpackages(internalRoot string) ([]packageBoundaryFinding, error) {
	matches := make(map[string]struct{})
	err := walkBoundaryGoFiles(internalRoot, func(path string) error {
		directory := filepath.Dir(path)
		relativeDirectory, err := filepath.Rel(internalRoot, directory)
		if err != nil {
			return fmt.Errorf("make %s relative to %s: %w", directory, internalRoot, err)
		}
		parts := strings.Split(filepath.ToSlash(relativeDirectory), "/")
		if len(parts) < 2 {
			return nil
		}
		if _, domain := domainPackages[parts[0]]; !domain {
			return nil
		}
		if _, layerName := layerPackageNames[parts[len(parts)-1]]; layerName {
			matches[filepath.ToSlash(relativeDirectory)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	findings := make([]packageBoundaryFinding, 0, len(matches))
	for path := range matches {
		findings = append(findings, packageBoundaryFinding{
			check:  "C3",
			path:   "internal/" + path,
			detail: "domain package contains a layer-named subpackage with Go files",
		})
	}
	return findings, nil
}

func findBucketPackages(internalRoot string) ([]packageBoundaryFinding, error) {
	matches := make(map[string]struct{})
	err := walkBoundaryGoFiles(internalRoot, func(path string) error {
		directory := filepath.Dir(path)
		if _, bucketName := bucketPackageNames[filepath.Base(directory)]; !bucketName {
			return nil
		}
		relativeDirectory, err := filepath.Rel(internalRoot, directory)
		if err != nil {
			return fmt.Errorf("make %s relative to %s: %w", directory, internalRoot, err)
		}
		matches[filepath.ToSlash(relativeDirectory)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	findings := make([]packageBoundaryFinding, 0, len(matches))
	for path := range matches {
		findings = append(findings, packageBoundaryFinding{
			check:  "C4",
			path:   "internal/" + path,
			detail: "bucket-named package contains Go files",
		})
	}
	return findings, nil
}

func importedRetiredLayer(importPath string) (string, bool) {
	const moduleInternalPrefix = "github.com/animal-ekarte/backend/internal/"
	for layer := range retiredLayerPackages {
		retiredImport := moduleInternalPrefix + layer
		if importPath == retiredImport || strings.HasPrefix(importPath, retiredImport+"/") {
			return layer, true
		}
	}
	return "", false
}

func walkBoundaryGoFiles(root string, visit func(path string) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && isPrunedBoundaryDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		return visit(path)
	})
}

func isPrunedBoundaryDirectory(name string) bool {
	return name == "testdata" || name == "vendor"
}
