package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
)

const (
	testStaffID   = uint64(1001)
	testClinicID  = uint64(1)
	testGroupID   = uint64(10)
	testSecretRef = "sec-1"
	testPassword  = "UAT-Attach-Pass1"
)

func TestApply_AttachesAccountWithoutInsertingStaff(t *testing.T) {
	h := newHarness(t, testStaffID)
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	beforeCount := len(h.fake.staffs)
	result, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "applied", result.Status)
	assert.NotEmpty(t, result.Digest)
	assert.Equal(t, 1, result.StaffCount)
	assert.Equal(t, []uint64{testStaffID}, result.StaffIDs)
	assert.Equal(t, beforeCount, len(h.fake.staffs))
	assert.Zero(t, h.fake.staffsInsertCount)
	assert.Equal(t, 1, h.fake.createAccountCalls)

	staff := h.fake.staffs[testStaffID]
	require.NotNil(t, staff)
	require.NotNil(t, staff.AccountID)
	account := h.fake.accounts[*staff.AccountID]
	require.NotNil(t, account)
	assert.Equal(t, sampleEmail(testStaffID), account.Email)
	assert.True(t, account.IsActive)
	assert.False(t, account.IsSystemAdmin)
	assert.True(t, staff.IsActive)
}

func TestEmail_MustMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{name: "reject other@example.test", email: "other@example.test", wantErr: true},
		{name: "accept formatted email", email: "stg-staff-1001@example.test", wantErr: false},
		{name: "accept trim and lower", email: "  STG-STAFF-1001@EXAMPLE.TEST  ", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, testStaffID)
			rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, tt.email, []uint64{testGroupID}), sampleSecretsJSON())

			result, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Zero(t, h.fake.createAccountCalls)
				assert.Zero(t, h.fake.updateStaffCalls)
				assert.Zero(t, h.fake.assignCalls)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, 1, h.fake.createAccountCalls)
			account := h.fake.accounts[*h.fake.staffs[testStaffID].AccountID]
			require.NotNil(t, account)
			assert.Equal(t, sampleEmail(testStaffID), account.Email)
		})
	}
}

func TestPermissionGroups_EmptyOrUnknownFailBeforeCreateAccount(t *testing.T) {
	tests := []struct {
		name   string
		groups []uint64
	}{
		{name: "empty ids", groups: []uint64{}},
		{name: "unknown id", groups: []uint64{999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, testStaffID)
			rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), tt.groups), sampleSecretsJSON())

			result, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.Zero(t, h.fake.createAccountCalls)
			assert.Zero(t, h.fake.updateStaffCalls)
			assert.Zero(t, h.fake.assignCalls)
			assert.Zero(t, h.fake.staffsInsertCount)
		})
	}
}

func TestWriteJSON_ResultHasNoNameEmailPassword(t *testing.T) {
	t.Parallel()
	plantedPassword := "PlantedPassword1"
	plantedEmail := "stg-staff-1001@example.test"

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, attachResult{
		Status:     "applied",
		Digest:     "deadbeef",
		StaffCount: 1,
		StaffIDs:   []uint64{testStaffID},
	}))
	out := buf.String()
	lower := strings.ToLower(out)
	assert.Contains(t, out, `"status":"applied"`)
	assert.Contains(t, out, `"digest":"deadbeef"`)
	assert.Contains(t, out, `"staff_count":1`)
	assert.Contains(t, out, `"staff_ids"`)
	assert.NotContains(t, lower, "name")
	assert.NotContains(t, lower, "email")
	assert.NotContains(t, lower, "password")
	assert.NotContains(t, out, plantedPassword)
	assert.NotContains(t, out, plantedEmail)
}

func TestRun_ApplyRefusesNonLocalWithoutOverride(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte")
	t.Setenv("STG_UAT_STAFF_ATTACH_ALLOW_REMOTE", "")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := runDependencies{
		configureTimeZone: func() error { return nil },
		fromEnv: func() (dbconn.ConnParams, error) {
			return dbconn.ConnParams{
				Host: "example.invalid", Port: "5432", User: "u", Password: "p", SSLMode: "require",
			}, nil
		},
		openDB: func(string) (*gorm.DB, error) {
			t.Fatal("openDB must not be called when remote apply is refused")
			return nil, nil
		},
		repoRoots: func(string) ([]string, error) { return nil, nil },
		newAttacher: func(*gorm.DB, []string) *attacher {
			t.Fatal("attacher must not be built")
			return nil
		},
	}
	err := run(context.Background(), []string{
		"apply",
		"--roster=/tmp/r.json",
		"--secrets=/tmp/s.json",
	}, logger, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "STG_UAT_STAFF_ATTACH_ALLOW_REMOTE=YES_I_UNDERSTAND")
}

func TestApply_IdempotentSameDigestNoop(t *testing.T) {
	h := newHarness(t, testStaffID)
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	first, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.NoError(t, err)
	require.Equal(t, "applied", first.Status)
	require.Equal(t, 1, h.fake.createAccountCalls)
	accountID := *h.fake.staffs[testStaffID].AccountID

	second, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, "noop", second.Status)
	assert.Equal(t, first.Digest, second.Digest)
	assert.Equal(t, 1, h.fake.createAccountCalls)
	assert.Equal(t, 1, len(h.fake.staffs))
	assert.Zero(t, h.fake.staffsInsertCount)
	assert.Equal(t, accountID, *h.fake.staffs[testStaffID].AccountID)
	assert.Equal(t, 1, len(h.fake.accounts))
}

func TestApply_HealsMissingDigestOnSameEmail(t *testing.T) {
	h := newHarness(t, testStaffID)
	accountID := uint64(50)
	h.fake.staffs[testStaffID].AccountID = &accountID
	h.fake.accounts[accountID] = &model.Account{
		ID:       accountID,
		Email:    sampleEmail(testStaffID),
		IsActive: true,
	}
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	first, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.NoError(t, err)
	require.Equal(t, "noop", first.Status)
	assert.Zero(t, h.fake.createAccountCalls)
	require.NotEmpty(t, h.fake.lastDigest[testStaffID])

	h.fake.validGroups[testClinicID][11] = struct{}{}
	changedPath, changedSecrets := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID, 11}), sampleSecretsJSON())
	second, err := h.attacher.Apply(context.Background(), changedPath, changedSecrets)
	require.Error(t, err)
	assert.Nil(t, second)
	assert.Zero(t, h.fake.createAccountCalls)
}

func TestApply_DifferentDigestConflict(t *testing.T) {
	h := newHarness(t, testStaffID)
	accountID := uint64(50)
	h.fake.staffs[testStaffID].AccountID = &accountID
	h.fake.accounts[accountID] = &model.Account{
		ID:       accountID,
		Email:    sampleEmail(testStaffID),
		IsActive: true,
	}
	h.fake.lastDigest[testStaffID] = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	result, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, h.fake.createAccountCalls)
	assert.Equal(t, 1, len(h.fake.accounts))
	assert.Equal(t, accountID, *h.fake.staffs[testStaffID].AccountID)
}

func TestApply_MissingStaffFails(t *testing.T) {
	h := newHarness(t, testStaffID)
	delete(h.fake.staffs, testStaffID)
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	result, err := h.attacher.Apply(context.Background(), rosterPath, secretsPath)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, h.fake.createAccountCalls)
	assert.Zero(t, h.fake.updateStaffCalls)
	assert.Zero(t, h.fake.assignCalls)
	assert.Zero(t, h.fake.staffsInsertCount)
}

func TestPreflight_WriteZero(t *testing.T) {
	h := newHarness(t, testStaffID)
	rosterPath, secretsPath := h.write(t, sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}), sampleSecretsJSON())

	result, err := h.attacher.Preflight(context.Background(), rosterPath, secretsPath)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "preflight", result.Status)
	assert.NotEmpty(t, result.Digest)
	assert.Equal(t, 1, result.StaffCount)
	assert.Equal(t, []uint64{testStaffID}, result.StaffIDs)
	assert.Zero(t, h.fake.createAccountCalls)
	assert.Zero(t, h.fake.updateStaffCalls)
	assert.Zero(t, h.fake.assignCalls)
	assert.Zero(t, h.fake.staffsInsertCount)
	assert.Empty(t, h.fake.lastDigest)
	assert.Nil(t, h.fake.staffs[testStaffID].AccountID)
}

func TestCommandSourceDoesNotCreateStaff(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("main.go")
	require.NoError(t, err)
	assert.NotContains(t, string(body), "CreateStaff")
}

func TestParseAttachDigestReceipt(t *testing.T) {
	t.Parallel()
	got, err := parseAttachDigestReceipt([]byte(`{"digest":"abc"}`))
	require.NoError(t, err)
	assert.Equal(t, "abc", got)

	empty, err := parseAttachDigestReceipt(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestRun_PreflightStdoutHasNoSecrets(t *testing.T) {
	filesDir := t.TempDir()
	repoRoot := t.TempDir()
	fake := newFakeAttachRepo()
	seedUnattachedStaff(fake, testStaffID, testClinicID)
	fake.staffs[testStaffID].Name = "合成"

	rosterPath := writeJSONFile(t, filesDir, "roster.json", sampleRosterJSON(testStaffID, sampleEmail(testStaffID), []uint64{testGroupID}))
	secretsPath := writeJSONFile(t, filesDir, "secrets.json", sampleSecretsJSON())

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
		newAttacher: func(_ *gorm.DB, roots []string) *attacher {
			return newStaffAttacher(fake, roots)
		},
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStdout := os.Stdout
	os.Stdout = w
	runErr := run(context.Background(), []string{
		"preflight",
		"--roster=" + rosterPath,
		"--secrets=" + secretsPath,
		"--repo-root=" + repoRoot,
	}, logger, deps)
	require.NoError(t, w.Close())
	os.Stdout = oldStdout
	require.NoError(t, runErr)

	var stdoutBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(r)
	_ = r.Close()
	out := stdoutBuf.String()

	var got attachResult
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "preflight", got.Status)
	assert.NotEmpty(t, got.Digest)
	assert.Equal(t, 1, got.StaffCount)
	assert.Equal(t, []uint64{testStaffID}, got.StaffIDs)
	assert.Contains(t, out, `"digest"`)
	assert.Contains(t, out, `"staff_count"`)
	assert.Contains(t, out, `"staff_ids"`)
	assert.NotContains(t, out, testPassword)
	assert.NotContains(t, out, sampleEmail(testStaffID))
	assert.NotContains(t, out, "合成")
	assert.NotContains(t, strings.ToLower(out), "password")
	assert.NotContains(t, strings.ToLower(out), "email")

	logs := logBuf.String()
	assert.NotContains(t, logs, testPassword)
	assert.NotContains(t, logs, sampleEmail(testStaffID))
	assert.NotContains(t, logs, "合成")
}

type fakeAttachRepo struct {
	staffs             map[uint64]*model.Staff
	accounts           map[uint64]*model.Account
	nextAccountID      uint64
	createAccountCalls int
	updateStaffCalls   int
	assignCalls        int
	staffsInsertCount  int
	validGroups        map[uint64]map[uint64]struct{}
	lastDigest         map[uint64]string
	assignments        map[string]struct{}
}

var _ attachRepository = (*fakeAttachRepo)(nil)

func newFakeAttachRepo() *fakeAttachRepo {
	return &fakeAttachRepo{
		staffs:        make(map[uint64]*model.Staff),
		accounts:      make(map[uint64]*model.Account),
		nextAccountID: 1,
		validGroups:   make(map[uint64]map[uint64]struct{}),
		lastDigest:    make(map[uint64]string),
		assignments:   make(map[string]struct{}),
	}
}

func seedUnattachedStaff(fake *fakeAttachRepo, staffID, clinicID uint64) {
	fake.staffs[staffID] = &model.Staff{
		ID:       staffID,
		ClinicID: clinicID,
		Name:     fmt.Sprintf("synthetic-%d", staffID),
		IsActive: false,
	}
	if fake.validGroups[clinicID] == nil {
		fake.validGroups[clinicID] = make(map[uint64]struct{})
	}
	fake.validGroups[clinicID][testGroupID] = struct{}{}
}

func (f *fakeAttachRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (f *fakeAttachRepo) FindStaffByID(_ context.Context, id uint64) (*model.Staff, error) {
	staff, ok := f.staffs[id]
	if !ok {
		return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	cp := *staff
	if staff.AccountID != nil {
		accountID := *staff.AccountID
		cp.AccountID = &accountID
	}
	return &cp, nil
}

func (f *fakeAttachRepo) FindAccountByID(_ context.Context, accountID uint64) (*model.Account, error) {
	account, ok := f.accounts[accountID]
	if !ok {
		return nil, apperrors.WrapNotFound("account", fmt.Sprintf("%d", accountID))
	}
	cp := *account
	return &cp, nil
}

func (f *fakeAttachRepo) PermissionGroupsBelongToClinic(_ context.Context, clinicID uint64, groupIDs []uint64) error {
	groups, ok := f.validGroups[clinicID]
	if !ok {
		return apperrors.WrapInvalidInput("permission_group_ids contains invalid permission group")
	}
	for _, id := range groupIDs {
		if _, exists := groups[id]; !exists {
			return apperrors.WrapInvalidInput("permission_group_ids contains invalid permission group")
		}
	}
	return nil
}

func (f *fakeAttachRepo) CreateAccount(_ context.Context, account *model.Account) error {
	f.createAccountCalls++
	if f.nextAccountID == 0 {
		f.nextAccountID = 1
	}
	account.ID = f.nextAccountID
	f.nextAccountID++
	cp := *account
	f.accounts[account.ID] = &cp
	return nil
}

func (f *fakeAttachRepo) UpdateStaffAccount(_ context.Context, staffID, clinicID, accountID uint64, setActive bool) error {
	f.updateStaffCalls++
	staff, ok := f.staffs[staffID]
	if !ok || staff.AccountID != nil || staff.ClinicID != clinicID {
		return apperrors.WrapConflict("staff account attach did not update exactly one row")
	}
	id := accountID
	staff.AccountID = &id
	if setActive {
		staff.IsActive = true
	}
	return nil
}

func (f *fakeAttachRepo) EnsureClinicAssignment(_ context.Context, staffID, clinicID uint64) error {
	key := fmt.Sprintf("%d:%d", staffID, clinicID)
	f.assignments[key] = struct{}{}
	return nil
}

func (f *fakeAttachRepo) AssignPermissionGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	f.assignCalls++
	_ = staffID
	return f.PermissionGroupsBelongToClinic(ctx, clinicID, groupIDs)
}

func (f *fakeAttachRepo) LastAttachDigest(_ context.Context, staffID uint64) (string, error) {
	return f.lastDigest[staffID], nil
}

func (f *fakeAttachRepo) SaveAttachDigest(_ context.Context, staffID uint64, digest string) error {
	f.lastDigest[staffID] = digest
	return nil
}

type attachHarness struct {
	filesDir string
	repoRoot string
	fake     *fakeAttachRepo
	attacher *attacher
}

func newHarness(t *testing.T, staffID uint64) *attachHarness {
	t.Helper()
	filesDir := t.TempDir()
	repoRoot := t.TempDir()
	fake := newFakeAttachRepo()
	seedUnattachedStaff(fake, staffID, testClinicID)
	return &attachHarness{
		filesDir: filesDir,
		repoRoot: repoRoot,
		fake:     fake,
		attacher: newStaffAttacher(fake, []string{repoRoot}),
	}
}

func (h *attachHarness) write(t *testing.T, roster, secrets any) (string, string) {
	t.Helper()
	return writeJSONFile(t, h.filesDir, "roster.json", roster), writeJSONFile(t, h.filesDir, "secrets.json", secrets)
}

func writeJSONFile(t *testing.T, dir, name string, v any) string {
	t.Helper()
	body, err := json.Marshal(v)
	require.NoError(t, err)
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, body, 0o600))
	require.NoError(t, os.Chmod(path, 0o600))
	return path
}

func sampleEmail(staffID uint64) string {
	return fmt.Sprintf("stg-staff-%d@example.test", staffID)
}

func sampleRosterJSON(staffID uint64, email string, groups []uint64) map[string]any {
	return map[string]any{
		"schema_version": "stg-uat-staff-attach-v1",
		"staff": []map[string]any{
			{
				"staff_id":             staffID,
				"clinic_id":            testClinicID,
				"email":                email,
				"secret_ref":           testSecretRef,
				"permission_group_ids": groups,
				"set_active":           true,
			},
		},
	}
}

func sampleSecretsJSON() map[string]any {
	return map[string]any{
		"secrets": []map[string]any{
			{"secret_ref": testSecretRef, "password": testPassword},
		},
	}
}
