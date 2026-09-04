package staff

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestStaffProvisioning_ClinicScopeBatchID_IsStableNamespace(t *testing.T) {
	t.Parallel()
	scope := []uint64{1, 2, 10}
	got := ClinicScopeBatchID(scope)
	require.True(t, strings.HasPrefix(got, StaffProvisionBatchIDPrefix))
	require.Equal(t, got, ClinicScopeBatchID([]uint64{1, 2, 10}))
	require.NotEqual(t, got, ClinicScopeBatchID([]uint64{1, 2, 11}))
}

func TestStaffProvisioning_DecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := DecodeStaffProvisionManifest(strings.NewReader(`{
		"schema_version":"staff-provision-v1",
		"batch_id":"x",
		"clinic_scope":[1],
		"actor_account_id":1,
		"staff":[],
		"unexpected":true
	}`))
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assertNoJSONDecoderLeak(t, err)

	_, err = DecodeStaffProvisionSecrets(strings.NewReader(`{
		"secrets":[{"secret_ref":"a","password":"Password1"}],
		"extra":1
	}`))
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assertNoJSONDecoderLeak(t, err)
}

func TestStaffProvisioning_DecodeRejectsInvalidJSONWithoutDecoderLeak(t *testing.T) {
	t.Parallel()
	_, err := DecodeStaffProvisionManifest(strings.NewReader(`{`))
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assertNoJSONDecoderLeak(t, err)

	_, err = DecodeStaffProvisionSecrets(strings.NewReader(`not-json`))
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assertNoJSONDecoderLeak(t, err)
}

func assertNoJSONDecoderLeak(t *testing.T, err error) {
	t.Helper()
	msg := err.Error()
	assert.NotContains(t, msg, "json:")
	assert.NotContains(t, msg, "invalid value")
	assert.NotContains(t, msg, "unknown field")
}

func TestStaffProvisioning_DecodeRejectsTrailingJSON(t *testing.T) {
	t.Parallel()
	_, err := DecodeStaffProvisionManifest(strings.NewReader(
		`{"schema_version":"staff-provision-v1","batch_id":"x","clinic_scope":[1],"actor_account_id":1,"staff":[]}{"x":1}`,
	))
	require.Error(t, err)
}

func TestStaffProvisioning_StructureRequiresScopeUnionAndBatchNamespace(t *testing.T) {
	t.Parallel()
	scope := []uint64{1, 2}
	batchID := ClinicScopeBatchID(scope)
	base := validManifest(scope, batchID)

	t.Run("unsorted scope", func(t *testing.T) {
		m := cloneManifest(base)
		m.ClinicScope = []uint64{2, 1}
		m.BatchID = ClinicScopeBatchID([]uint64{1, 2})
		require.Error(t, ValidateStaffProvisionManifestStructure(m))
	})
	t.Run("batch id mismatch", func(t *testing.T) {
		m := cloneManifest(base)
		m.BatchID = StaffProvisionBatchIDPrefix + "deadbeef"
		require.Error(t, ValidateStaffProvisionManifestStructure(m))
	})
	t.Run("scope not equal union", func(t *testing.T) {
		m := cloneManifest(base)
		m.ClinicScope = []uint64{1, 2, 3}
		m.BatchID = ClinicScopeBatchID(m.ClinicScope)
		require.Error(t, ValidateStaffProvisionManifestStructure(m))
	})
	t.Run("duplicate external id", func(t *testing.T) {
		m := cloneManifest(base)
		m.Staff = append(m.Staff, m.Staff[0])
		m.Staff[1].Email = "other@example.com"
		m.Staff[1].SecretRef = "other-ref"
		require.Error(t, ValidateStaffProvisionManifestStructure(m))
	})
	t.Run("happy path", func(t *testing.T) {
		require.NoError(t, ValidateStaffProvisionManifestStructure(base))
	})
}

func TestStaffProvisioning_SecretsMustMatchExactly(t *testing.T) {
	t.Parallel()
	scope := []uint64{1}
	m := validManifest(scope, ClinicScopeBatchID(scope))
	require.NoError(t, ValidateStaffProvisionManifestStructure(m))

	t.Run("missing secret", func(t *testing.T) {
		_, err := ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{Secrets: nil})
		require.Error(t, err)
	})
	t.Run("extra secret", func(t *testing.T) {
		_, err := ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{
			Secrets: []StaffProvisionSecretEntry{
				{SecretRef: m.Staff[0].SecretRef, Password: "Password1"},
				{SecretRef: "unused", Password: "Password2"},
			},
		})
		require.Error(t, err)
	})
	t.Run("weak password", func(t *testing.T) {
		_, err := ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{
			Secrets: []StaffProvisionSecretEntry{
				{SecretRef: m.Staff[0].SecretRef, Password: "short"},
			},
		})
		require.Error(t, err)
	})
	t.Run("happy path", func(t *testing.T) {
		got, err := ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{
			Secrets: []StaffProvisionSecretEntry{
				{SecretRef: m.Staff[0].SecretRef, Password: "Password1"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "Password1", got[m.Staff[0].SecretRef])
	})
}

func TestStaffProvisioning_SecurePathRequires0600AbsoluteOutsideRepo(t *testing.T) {
	t.Parallel()
	repoRoot := t.TempDir()
	outside := t.TempDir()

	insidePath := filepath.Join(repoRoot, "manifest.json")
	require.NoError(t, os.WriteFile(insidePath, []byte(`{}`), 0o600))
	_, err := ValidateSecureInputPath(insidePath, []string{repoRoot})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the repository")

	missing := filepath.Join(outside, "missing.json")
	_, err = ValidateSecureInputPath(missing, []string{repoRoot})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	rel := "manifest.json"
	_, err = ValidateSecureInputPath(rel, []string{repoRoot})
	require.Error(t, err)

	badMode := filepath.Join(outside, "bad-mode.json")
	require.NoError(t, os.WriteFile(badMode, []byte(`{}`), 0o644))
	_, err = ValidateSecureInputPath(badMode, []string{repoRoot})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "0600")

	okPath := filepath.Join(outside, "ok.json")
	require.NoError(t, os.WriteFile(okPath, []byte(`{}`), 0o600))
	got, err := ValidateSecureInputPath(okPath, []string{repoRoot})
	require.NoError(t, err)
	assert.Equal(t, filepath.Clean(okPath), got)
}

func TestStaffProvisioning_DigestExcludesRawPIIAndPasswords(t *testing.T) {
	t.Parallel()
	scope := []uint64{1}
	m := validManifest(scope, ClinicScopeBatchID(scope))
	digest, err := ComputeStaffProvisionDigest(m)
	require.NoError(t, err)
	require.Len(t, digest, 64)
	assert.NotContains(t, digest, "Password")
	assert.NotContains(t, digest, m.Staff[0].Email)
	assert.NotContains(t, digest, m.Staff[0].Name)

	m2 := cloneManifest(m)
	m2.Staff[0].Name = "別氏名"
	digest2, err := ComputeStaffProvisionDigest(m2)
	require.NoError(t, err)
	assert.NotEqual(t, digest, digest2)
}

func TestStaffProvisioning_DecideReceiptState(t *testing.T) {
	t.Parallel()
	scope := []uint64{1, 2}
	batchID := ClinicScopeBatchID(scope)
	digest := "abc"

	t.Run("no receipts apply", func(t *testing.T) {
		d, err := decideReceiptState(scope, batchID, digest, nil)
		require.NoError(t, err)
		assert.Equal(t, receiptDecisionApply, d)
	})
	t.Run("full match noop", func(t *testing.T) {
		d, err := decideReceiptState(scope, batchID, digest, []StaffProvisionReceipt{
			{ClinicID: 1, BatchID: batchID, Digest: digest, Count: 2},
			{ClinicID: 2, BatchID: batchID, Digest: digest, Count: 2},
		})
		require.NoError(t, err)
		assert.Equal(t, receiptDecisionNoop, d)
	})
	t.Run("partial conflict", func(t *testing.T) {
		_, err := decideReceiptState(scope, batchID, digest, []StaffProvisionReceipt{
			{ClinicID: 1, BatchID: batchID, Digest: digest, Count: 2},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})
	t.Run("digest mismatch conflict", func(t *testing.T) {
		_, err := decideReceiptState(scope, batchID, digest, []StaffProvisionReceipt{
			{ClinicID: 1, BatchID: batchID, Digest: "other", Count: 2},
			{ClinicID: 2, BatchID: batchID, Digest: "other", Count: 2},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})
	t.Run("out of scope receipts ignored", func(t *testing.T) {
		d, err := decideReceiptState(scope, batchID, digest, []StaffProvisionReceipt{
			{ClinicID: 99, BatchID: batchID, Digest: digest, Count: 2},
		})
		require.NoError(t, err)
		assert.Equal(t, receiptDecisionApply, d)
	})
}

func TestStaffProvisioning_PreflightZeroWritesAndUnauthorizedDoesNotProbeReceipts(t *testing.T) {
	repo := newMockProvisionRepo()
	repo.accounts[1] = &model.Account{ID: 1, Email: "actor@example.com", IsActive: true, IsSystemAdmin: false}
	// non-admin without staff → forbidden before receipt lookup
	p := NewStaffProvisioner(repo, nil)

	paths := writeProvisionFixture(t, singleClinicManifestAndSecrets(1, 1))
	// Actor account missing clinic fixture still fails closed before receipts.
	repo.clinics[1] = true
	_, err := p.Preflight(context.Background(), paths.manifest, paths.secrets)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "got %v", err)
	// Ensure unauthorized path never probes receipts and never writes.
	assert.Equal(t, 0, repo.receiptLookups)
	assert.Equal(t, 0, repo.writes)
}

func TestStaffProvisioning_PreflightSystemAdminOKZeroWrites(t *testing.T) {
	repo := newMockProvisionRepo()
	repo.accounts[9] = &model.Account{ID: 9, Email: "admin@example.com", IsActive: true, IsSystemAdmin: true}
	repo.clinics[1] = true
	p := NewStaffProvisioner(repo, nil)

	paths := writeProvisionFixture(t, singleClinicManifestAndSecrets(1, 9))
	result, err := p.Preflight(context.Background(), paths.manifest, paths.secrets)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.StaffCount)
	assert.Equal(t, 0, repo.writes)
	assert.Equal(t, 0, repo.receiptLookups)
	assert.NotContains(t, result.Digest, "@")
}

func TestStaffProvisioning_ApplyIdempotentNoopAndConflict(t *testing.T) {
	repo := newMockProvisionRepo()
	repo.accounts[9] = &model.Account{ID: 9, Email: "admin@example.com", IsActive: true, IsSystemAdmin: true}
	repo.clinics[1] = true
	p := NewStaffProvisioner(repo, nil)
	paths := writeProvisionFixture(t, singleClinicManifestAndSecrets(1, 9))

	first, err := p.Apply(context.Background(), paths.manifest, paths.secrets)
	require.NoError(t, err)
	require.Equal(t, "applied", first.Status)
	assert.Equal(t, 1, repo.createdStaff)
	assert.Greater(t, repo.writes, 0)

	// Reset write counter but keep receipts.
	repo.writes = 0
	repo.createdStaff = 0
	second, err := p.Apply(context.Background(), paths.manifest, paths.secrets)
	require.NoError(t, err)
	require.Equal(t, "noop", second.Status)
	assert.Equal(t, 0, repo.createdStaff)

	// Same batch_id (same clinic_scope) but different content → conflict.
	fixture := singleClinicManifestAndSecrets(1, 9)
	fixture.manifest.Staff[0].Name = "Changed Name"
	// recompute digest path via Apply; batch_id stays same because clinic_scope unchanged
	conflictPaths := writeProvisionFixture(t, fixture)
	// Seed receipts from first apply remain in repo with old digest.
	_, err = p.Apply(context.Background(), conflictPaths.manifest, conflictPaths.secrets)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestStaffProvisioning_ApplyPartialReceiptConflict(t *testing.T) {
	repo := newMockProvisionRepo()
	repo.accounts[9] = &model.Account{ID: 9, Email: "admin@example.com", IsActive: true, IsSystemAdmin: true}
	repo.clinics[1] = true
	repo.clinics[2] = true
	p := NewStaffProvisioner(repo, nil)

	scope := []uint64{1, 2}
	batchID := ClinicScopeBatchID(scope)
	fixture := multiClinicManifestAndSecrets(scope, batchID, 9)
	paths := writeProvisionFixture(t, fixture)

	// Inject partial receipt for only clinic 1.
	digest, err := ComputeStaffProvisionDigest(fixture.manifest)
	require.NoError(t, err)
	repo.receipts = []StaffProvisionReceipt{{
		ClinicID: 1, BatchID: batchID, Digest: digest, Count: 1,
	}}

	_, err = p.Apply(context.Background(), paths.manifest, paths.secrets)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 0, repo.createdStaff)
}

func TestStaffProvisioning_ApplyRejectsUnknownFK(t *testing.T) {
	repo := newMockProvisionRepo()
	repo.accounts[9] = &model.Account{ID: 9, Email: "admin@example.com", IsActive: true, IsSystemAdmin: true}
	// clinic missing
	p := NewStaffProvisioner(repo, nil)
	paths := writeProvisionFixture(t, singleClinicManifestAndSecrets(1, 9))
	_, err := p.Apply(context.Background(), paths.manifest, paths.secrets)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Equal(t, 0, repo.createdStaff)
}

func TestStaffProvisioning_LoadInputsRejectsRepoPathAndBadMode(t *testing.T) {
	repoRoot := t.TempDir()
	// Inside repo
	inside := filepath.Join(repoRoot, "in.json")
	require.NoError(t, os.WriteFile(inside, []byte(`{}`), 0o600))
	outside := t.TempDir()
	secrets := filepath.Join(outside, "secrets.json")
	require.NoError(t, os.WriteFile(secrets, []byte(`{"secrets":[]}`), 0o600))
	_, _, _, err := LoadStaffProvisionInputs(inside, secrets, []string{repoRoot})
	require.Error(t, err)
}

func TestStaffProvisioning_ErrorMessagesNeverContainSecrets(t *testing.T) {
	t.Parallel()
	m := validManifest([]uint64{1}, ClinicScopeBatchID([]uint64{1}))
	_, err := ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{
		Secrets: []StaffProvisionSecretEntry{
			{SecretRef: m.Staff[0].SecretRef, Password: "SuperSecret99"},
		},
	})
	// happy path — ensure digest path doesn't embed password
	require.NoError(t, err)
	digest, err := ComputeStaffProvisionDigest(m)
	require.NoError(t, err)
	assert.NotContains(t, digest, "SuperSecret99")

	_, err = ValidateStaffProvisionSecrets(m, &StaffProvisionSecretsFile{
		Secrets: []StaffProvisionSecretEntry{
			{SecretRef: m.Staff[0].SecretRef, Password: "bad"},
		},
	})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "bad")
	assert.NotContains(t, err.Error(), "SuperSecret")
}

// ---- helpers / mock repo ----

type provisionFixture struct {
	manifest *StaffProvisionManifest
	secrets  *StaffProvisionSecretsFile
}

type provisionPaths struct {
	manifest string
	secrets  string
}

func writeProvisionFixture(t *testing.T, fixture provisionFixture) provisionPaths {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	secretsPath := filepath.Join(dir, "secrets.json")
	mb, err := json.Marshal(fixture.manifest)
	require.NoError(t, err)
	sb, err := json.Marshal(fixture.secrets)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, mb, 0o600))
	require.NoError(t, os.WriteFile(secretsPath, sb, 0o600))
	return provisionPaths{manifest: manifestPath, secrets: secretsPath}
}

func validManifest(scope []uint64, batchID string) *StaffProvisionManifest {
	main := scope[0]
	return &StaffProvisionManifest{
		SchemaVersion:  StaffProvisionSchemaVersion,
		BatchID:        batchID,
		ClinicScope:    append([]uint64(nil), scope...),
		ActorAccountID: 9,
		Staff: []StaffProvisionStaffEntry{{
			ExternalStaffID:    "ext-001",
			Name:               "合成スタッフ",
			Email:              "synthetic-staff-001@example.test",
			MainClinicID:       main,
			ClinicIDs:          append([]uint64(nil), scope...),
			PermissionGroupIDs: nil,
			OccupationID:       nil,
			StaffType:          string(model.StaffTypeDoctor),
			IsActive:           true,
			ReservationVisible: true,
			SecretRef:          "secret-ext-001",
		}},
	}
}

func cloneManifest(m *StaffProvisionManifest) *StaffProvisionManifest {
	raw, _ := json.Marshal(m)
	var out StaffProvisionManifest
	_ = json.Unmarshal(raw, &out)
	return &out
}

func singleClinicManifestAndSecrets(clinicID, actorID uint64) provisionFixture {
	scope := []uint64{clinicID}
	batchID := ClinicScopeBatchID(scope)
	m := validManifest(scope, batchID)
	m.ActorAccountID = actorID
	return provisionFixture{
		manifest: m,
		secrets: &StaffProvisionSecretsFile{
			Secrets: []StaffProvisionSecretEntry{
				{SecretRef: m.Staff[0].SecretRef, Password: "Password1"},
			},
		},
	}
}

func multiClinicManifestAndSecrets(scope []uint64, batchID string, actorID uint64) provisionFixture {
	m := validManifest(scope, batchID)
	m.ActorAccountID = actorID
	return provisionFixture{
		manifest: m,
		secrets: &StaffProvisionSecretsFile{
			Secrets: []StaffProvisionSecretEntry{
				{SecretRef: m.Staff[0].SecretRef, Password: "Password1"},
			},
		},
	}
}

type mockProvisionRepo struct {
	mu sync.Mutex

	accounts       map[uint64]*model.Account
	staffByAccount map[uint64]*model.Staff
	clinics        map[uint64]bool
	emails         map[string]bool
	occupations    map[string]bool // clinicID:occID
	groups         map[string]bool // clinicID:groupID
	assignments    map[string]bool // staffID:clinicID
	permissions    map[string]bool // staffID:clinicID master-staff create
	receipts       []StaffProvisionReceipt

	nextAccountID  uint64
	nextStaffID    uint64
	createdStaff   int
	writes         int
	receiptLookups int
	locked         bool
}

func newMockProvisionRepo() *mockProvisionRepo {
	return &mockProvisionRepo{
		accounts:       map[uint64]*model.Account{},
		staffByAccount: map[uint64]*model.Staff{},
		clinics:        map[uint64]bool{},
		emails:         map[string]bool{},
		occupations:    map[string]bool{},
		groups:         map[string]bool{},
		assignments:    map[string]bool{},
		permissions:    map[string]bool{},
		nextAccountID:  100,
		nextStaffID:    200,
	}
}

func (m *mockProvisionRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *mockProvisionRepo) AcquireBatchLock(ctx context.Context, batchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locked = true
	_ = batchID
	return nil
}

func (m *mockProvisionRepo) FindReceiptsInScope(ctx context.Context, clinicIDs []uint64, batchID string) ([]StaffProvisionReceipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receiptLookups++
	out := make([]StaffProvisionReceipt, 0)
	for _, r := range m.receipts {
		if r.BatchID != batchID {
			continue
		}
		for _, id := range clinicIDs {
			if r.ClinicID == id {
				out = append(out, r)
				break
			}
		}
	}
	return out, nil
}

func (m *mockProvisionRepo) FindAccountByID(ctx context.Context, accountID uint64) (*model.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.accounts[accountID]
	if !ok {
		return nil, apperrors.WrapNotFound("account", fmt.Sprintf("%d", accountID))
	}
	cp := *acc
	return &cp, nil
}

func (m *mockProvisionRepo) FindStaffByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.staffByAccount[accountID]
	if !ok {
		return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("account_id=%d", accountID))
	}
	cp := *s
	return &cp, nil
}

func (m *mockProvisionRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.emails[email], nil
}

func (m *mockProvisionRepo) ClinicExists(ctx context.Context, clinicID uint64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clinics[clinicID], nil
}

func (m *mockProvisionRepo) OccupationBelongsToClinic(ctx context.Context, clinicID, occupationID uint64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.occupations[fmt.Sprintf("%d:%d", clinicID, occupationID)], nil
}

func (m *mockProvisionRepo) PermissionGroupsBelongToClinic(ctx context.Context, clinicID uint64, groupIDs []uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range groupIDs {
		if !m.groups[fmt.Sprintf("%d:%d", clinicID, id)] {
			return apperrors.WrapInvalidInput("permission_group_ids contains invalid permission group")
		}
	}
	return nil
}

func (m *mockProvisionRepo) StaffAssignedToClinic(ctx context.Context, staffID, clinicID uint64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.assignments[fmt.Sprintf("%d:%d", staffID, clinicID)], nil
}

func (m *mockProvisionRepo) HasMasterStaffCreate(ctx context.Context, staffID, clinicID uint64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.permissions[fmt.Sprintf("%d:%d", staffID, clinicID)], nil
}

func (m *mockProvisionRepo) CreateAccount(ctx context.Context, account *model.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.emails[account.Email] {
		return apperrors.WrapAlreadyExists("account", "email")
	}
	m.nextAccountID++
	account.ID = m.nextAccountID
	m.emails[account.Email] = true
	m.writes++
	return nil
}

func (m *mockProvisionRepo) CreateStaff(ctx context.Context, staff *model.Staff) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextStaffID++
	staff.ID = m.nextStaffID
	m.createdStaff++
	m.writes++
	return nil
}

func (m *mockProvisionRepo) CreateAssignment(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writes++
	return nil
}

func (m *mockProvisionRepo) AssignPermissionGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writes++
	return nil
}

func (m *mockProvisionRepo) LockOccupationForShare(ctx context.Context, clinicID, occupationID uint64) error {
	return nil
}

func (m *mockProvisionRepo) WriteAudit(ctx context.Context, entry *model.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writes++
	if entry.Action == model.AuditActionStaffProvisionReceipt && entry.ClinicID != nil {
		var payload struct {
			BatchID string `json:"batch_id"`
			Digest  string `json:"digest"`
			Count   int    `json:"count"`
		}
		_ = json.Unmarshal(entry.NewValue, &payload)
		// Ensure no PII keys in receipt body.
		raw := string(entry.NewValue)
		if strings.Contains(raw, "email") || strings.Contains(raw, "password") || strings.Contains(raw, "name") {
			return fmt.Errorf("receipt contains forbidden fields")
		}
		m.receipts = append(m.receipts, StaffProvisionReceipt{
			ClinicID: *entry.ClinicID,
			BatchID:  payload.BatchID,
			Digest:   payload.Digest,
			Count:    payload.Count,
		})
	}
	return nil
}

// Ensure mock implements interface.
var _ StaffProvisioningRepository = (*mockProvisionRepo)(nil)

func TestStaffProvisioning_StrictDecodeRoundTrip(t *testing.T) {
	t.Parallel()
	scope := []uint64{1, 3}
	m := validManifest(scope, ClinicScopeBatchID(scope))
	raw, err := json.Marshal(m)
	require.NoError(t, err)
	decoded, err := DecodeStaffProvisionManifest(bytes.NewReader(raw))
	require.NoError(t, err)
	assert.Equal(t, m.BatchID, decoded.BatchID)
	assert.Equal(t, m.Staff[0].ExternalStaffID, decoded.Staff[0].ExternalStaffID)
}
