package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
)

func TestParseOptions_RequiresCommandAndAbsoluteFlags(t *testing.T) {
	t.Parallel()
	_, err := parseOptions(nil)
	require.Error(t, err)

	_, err = parseOptions([]string{"explode"})
	require.Error(t, err)

	_, err = parseOptions([]string{"preflight"})
	require.Error(t, err)

	opt, err := parseOptions([]string{
		"apply",
		"--manifest=/tmp/m.json",
		"--secrets=/tmp/s.json",
		"--repo-root=/tmp/repo",
	})
	require.NoError(t, err)
	assert.Equal(t, "apply", opt.command)
	assert.Equal(t, "/tmp/m.json", opt.manifestPath)
	assert.Equal(t, "/tmp/s.json", opt.secretsPath)
	assert.Equal(t, "/tmp/repo", opt.repoRoot)
}

func TestSanitizeError_RedactsEmailAndPassword(t *testing.T) {
	t.Parallel()
	assert.Equal(t,
		"INVALID_INPUT: input path mode must be 0600",
		sanitizeError(apperrors.WrapInvalidInput("input path mode must be 0600")),
	)
	assert.Equal(t,
		"staff provision failed (details redacted)",
		sanitizeError(errors.New("failed for user@example.com")),
	)
	assert.Equal(t,
		"staff provision failed (details redacted)",
		sanitizeError(errors.New("bad password value")),
	)
}

func TestDefaultRepoRoots_ExplicitAndEnv(t *testing.T) {
	t.Setenv("STAFF_PROVISION_REPO_ROOT", "/abs/env-root")
	roots, err := defaultRepoRoots("/abs/explicit")
	require.NoError(t, err)
	assert.Contains(t, roots, "/abs/explicit")
	assert.Contains(t, roots, "/abs/env-root")

	_, err = defaultRepoRoots("relative")
	require.Error(t, err)
}

func TestRun_PreflightHappyPathLogsDigestOnly(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()
	scope := []uint64{1}
	batchID := staff.ClinicScopeBatchID(scope)
	manifest := staff.StaffProvisionManifest{
		SchemaVersion:  staff.StaffProvisionSchemaVersion,
		BatchID:        batchID,
		ClinicScope:    scope,
		ActorAccountID: 1,
		Staff: []staff.StaffProvisionStaffEntry{{
			ExternalStaffID:    "ext-1",
			Name:               "合成",
			Email:              "syn@example.test",
			MainClinicID:       1,
			ClinicIDs:          []uint64{1},
			StaffType:          "doctor",
			IsActive:           true,
			ReservationVisible: true,
			SecretRef:          "sec-1",
		}},
	}
	secrets := staff.StaffProvisionSecretsFile{
		Secrets: []staff.StaffProvisionSecretEntry{
			{SecretRef: "sec-1", Password: "Password1"},
		},
	}
	mb, err := json.Marshal(manifest)
	require.NoError(t, err)
	sb, err := json.Marshal(secrets)
	require.NoError(t, err)
	mp := filepath.Join(dir, "manifest.json")
	sp := filepath.Join(dir, "secrets.json")
	require.NoError(t, os.WriteFile(mp, mb, 0o600))
	require.NoError(t, os.WriteFile(sp, sb, 0o600))

	mockRepo := &cliMockRepo{
		account: &model.Account{ID: 1, IsActive: true, IsSystemAdmin: true},
		clinics: map[uint64]bool{1: true},
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	t.Setenv("DB_NAME", "animalekarte_test")

	deps := runDependencies{
		configureTimeZone: func() error { return nil },
		fromEnv: func() (dbconn.ConnParams, error) {
			return dbconn.ConnParams{
				Host: "localhost", Port: "5432", User: "u", Password: "p", SSLMode: "disable",
			}, nil
		},
		openDB: func(string) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		},
		repoRoots: func(string) ([]string, error) {
			return []string{repoRoot}, nil
		},
		newProvisioner: func(_ *gorm.DB, repoRoots []string) *staff.StaffProvisioner {
			return staff.NewStaffProvisioner(mockRepo, repoRoots)
		},
	}

	// Capture stdout JSON by temporarily replacing os.Stdout is heavy; run() writes
	// to os.Stdout. Instead assert logs and error-free completion via a pipe.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStdout := os.Stdout
	os.Stdout = w
	err = run(context.Background(), []string{
		"preflight",
		"--manifest=" + mp,
		"--secrets=" + sp,
		"--repo-root=" + repoRoot,
	}, logger, deps)
	require.NoError(t, w.Close())
	os.Stdout = oldStdout
	require.NoError(t, err)

	var stdoutBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(r)
	_ = r.Close()
	out := stdoutBuf.String()
	assert.Contains(t, out, `"batch_id"`)
	assert.Contains(t, out, `"digest"`)
	assert.NotContains(t, out, "Password1")
	assert.NotContains(t, out, "syn@example.test")
	assert.NotContains(t, out, "合成")

	logs := logBuf.String()
	assert.Contains(t, logs, "preflight PASS")
	assert.NotContains(t, logs, "Password1")
	assert.NotContains(t, logs, "syn@example.test")
}

func TestRun_ApplyRefusesNonLocalWithoutOverride(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte")
	t.Setenv("STAFF_PROVISION_ALLOW_REMOTE", "")
	logger := slog.New(slog.NewTextHandler(&ioDiscard{}, nil))
	deps := runDependencies{
		configureTimeZone: func() error { return nil },
		fromEnv: func() (dbconn.ConnParams, error) {
			return dbconn.ConnParams{
				Host: "db.example.internal", Port: "5432", User: "u", Password: "p", SSLMode: "require",
			}, nil
		},
		openDB: func(string) (*gorm.DB, error) {
			t.Fatal("openDB must not be called when remote apply is refused")
			return nil, nil
		},
		repoRoots: func(string) ([]string, error) { return nil, nil },
		newProvisioner: func(*gorm.DB, []string) *staff.StaffProvisioner {
			t.Fatal("provisioner must not be built")
			return nil
		},
	}
	err := run(context.Background(), []string{
		"apply",
		"--manifest=/tmp/m.json",
		"--secrets=/tmp/s.json",
	}, logger, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-local")
}

func TestWriteJSON_EmitsDigestOnlySurface(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, &staff.StaffProvisionApplyResult{
		Status:      "applied",
		BatchID:     "staff-provision:abc",
		Digest:      "deadbeef",
		StaffCount:  2,
		ClinicScope: []uint64{1, 2},
	}))
	out := buf.String()
	assert.Contains(t, out, `"status":"applied"`)
	assert.Contains(t, out, `"digest":"deadbeef"`)
	assert.NotContains(t, out, "password")
	assert.NotContains(t, out, "email")
}

func TestFindGoModRoot_FindsAncestor(t *testing.T) {
	t.Parallel()
	// This test file lives under backend/cmd/staff-provision; module root is backend/.
	wd, err := os.Getwd()
	require.NoError(t, err)
	root := findGoModRoot(wd)
	require.NotEmpty(t, root)
	_, err = os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, err)
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

// cliMockRepo satisfies staff.StaffProvisioningRepository for CLI wiring tests.
type cliMockRepo struct {
	account *model.Account
	clinics map[uint64]bool
}

func (m *cliMockRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
func (m *cliMockRepo) AcquireBatchLock(context.Context, string) error { return nil }
func (m *cliMockRepo) FindReceiptsInScope(context.Context, []uint64, string) ([]staff.StaffProvisionReceipt, error) {
	return nil, nil
}
func (m *cliMockRepo) FindAccountByID(_ context.Context, accountID uint64) (*model.Account, error) {
	if m.account == nil || m.account.ID != accountID {
		return nil, apperrors.WrapNotFound("account", "x")
	}
	cp := *m.account
	return &cp, nil
}
func (m *cliMockRepo) FindStaffByAccountID(context.Context, uint64) (*model.Staff, error) {
	return nil, apperrors.WrapNotFound("staff", "x")
}
func (m *cliMockRepo) EmailExists(context.Context, string) (bool, error) { return false, nil }
func (m *cliMockRepo) ClinicExists(_ context.Context, clinicID uint64) (bool, error) {
	return m.clinics[clinicID], nil
}
func (m *cliMockRepo) OccupationBelongsToClinic(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (m *cliMockRepo) PermissionGroupsBelongToClinic(context.Context, uint64, []uint64) error {
	return nil
}
func (m *cliMockRepo) StaffAssignedToClinic(context.Context, uint64, uint64) (bool, error) {
	return false, nil
}
func (m *cliMockRepo) HasMasterStaffCreate(context.Context, uint64, uint64) (bool, error) {
	return false, nil
}
func (m *cliMockRepo) CreateAccount(context.Context, *model.Account) error { return nil }
func (m *cliMockRepo) CreateStaff(context.Context, *model.Staff) error     { return nil }
func (m *cliMockRepo) CreateAssignment(context.Context, *model.StaffClinicAssignment) error {
	return nil
}
func (m *cliMockRepo) AssignPermissionGroups(context.Context, uint64, uint64, []uint64) error {
	return nil
}
func (m *cliMockRepo) LockOccupationForShare(context.Context, uint64, uint64) error { return nil }
func (m *cliMockRepo) WriteAudit(context.Context, *model.AuditLog) error            { return nil }

var _ staff.StaffProvisioningRepository = (*cliMockRepo)(nil)

func TestUniqueStringsAndSanitizeCodePath(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a", "b"}, uniqueStrings([]string{"a", "a", "b", ""}))
	assert.True(t, strings.HasPrefix(sanitizeError(apperrors.WrapConflict("x")), "CONFLICT:"))
}
