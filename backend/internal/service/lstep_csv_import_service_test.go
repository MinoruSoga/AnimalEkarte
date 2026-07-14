package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- LstepCsvImportRepository モック ----

type mockLstepCsvImportRepo struct {
	createFn       func(ctx context.Context, imp *model.LstepCsvImport) error
	updateFn       func(ctx context.Context, imp *model.LstepCsvImport) error
	findByIDFn     func(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error)
	listByClinicFn func(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

func (m *mockLstepCsvImportRepo) Create(ctx context.Context, imp *model.LstepCsvImport) error {
	if m.createFn != nil {
		return m.createFn(ctx, imp)
	}
	return nil
}
func (m *mockLstepCsvImportRepo) Update(ctx context.Context, imp *model.LstepCsvImport) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, imp)
	}
	return nil
}
func (m *mockLstepCsvImportRepo) FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}
func (m *mockLstepCsvImportRepo) FindAllByClinicID(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, limit)
	}
	return nil, nil
}

// ---- OwnerRepository モック（ImportFriendAttributesCSV の FindAllWithLineUserID 用） ----

type mockLstepImportOwnerRepo struct {
	findAllWithLineUserIDFn func(ctx context.Context, clinicID uint64) ([]model.Owner, error)
}

func (m *mockLstepImportOwnerRepo) FindAll(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}
func (m *mockLstepImportOwnerRepo) FindByID(_ context.Context, _, _ uint64) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindByLineUserID(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error) {
	if m.findAllWithLineUserIDFn != nil {
		return m.findAllWithLineUserIDFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) FindAllWithLineUserIDCursor(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}
func (m *mockLstepImportOwnerRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockLstepImportOwnerRepo) UpdateLineUserID(_ context.Context, _, _ uint64, _ *string) error {
	return nil
}
func (m *mockLstepImportOwnerRepo) FindAllByLineUserID(_ context.Context, _ string) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLstepImportOwnerRepo) UpdateLineFollowedAt(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *mockLstepImportOwnerRepo) UpdateLineBlockedAt(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *mockLstepImportOwnerRepo) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockLstepImportOwnerRepo) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockLstepImportOwnerRepo) FindByIDs(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
	return nil, nil
}

// TestImportFriendAttributesCSV_FileTooLarge: 50MB+1 バイトの reader で WrapInvalidInput を返すことを確認。
// csvImportRepo.Create 到達前 (line 69-71) に return するため nil repo で safe。
func TestImportFriendAttributesCSV_FileTooLarge(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 50*1024*1024+1)
	svc := &lstepCsvImportService{
		db:            nil,
		csvImportRepo: nil,
		snapshotRepo:  nil,
		ownerRepo:     nil,
	}
	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "large.csv", bytes.NewReader(big), 1)
	if err == nil {
		t.Fatal("expected error for 50MB+ file, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}

// TestImportFriendAttributesCSV_EmptyFile: 0 バイトの reader で WrapInvalidInput を返すことを確認。
// CSV パース後の空チェック (line 86-88) で return するため nil repo で safe。
func TestImportFriendAttributesCSV_EmptyFile(t *testing.T) {
	svc := &lstepCsvImportService{
		db:            nil,
		csvImportRepo: nil,
		snapshotRepo:  nil,
		ownerRepo:     nil,
	}
	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "empty.csv", strings.NewReader(""), 1)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}

// TestImportFriendAttributesCSV_MismatchedFieldCount: ヘッダーとデータ行の列数が異なる場合、
// csv.Reader の ErrFieldCount により「failed to parse CSV」で InvalidInput を返すことを確認。
// repo 到達前 (line 82-85) に return するため nil repo で safe。
func TestImportFriendAttributesCSV_MismatchedFieldCount(t *testing.T) {
	svc := &lstepCsvImportService{}
	csv := "a,b\n1,2,3\n"

	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "bad.csv", strings.NewReader(csv), 1)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// TestImportFriendAttributesCSV_InvalidHeader_CreateFailedRecordError は、ヘッダー解析に失敗し
// 失敗レコード作成の Create 自体もエラーになった場合でも、csvImportRepo.Create のエラーはログのみで
// 呼び出し元には常にヘッダーエラー（InvalidInput）が返ることを確認する。
func TestImportFriendAttributesCSV_InvalidHeader_CreateFailedRecordError(t *testing.T) {
	createCalled := false
	repo := &mockLstepCsvImportRepo{
		createFn: func(_ context.Context, imp *model.LstepCsvImport) error {
			createCalled = true
			assert.Equal(t, csvImportStatusFailed, imp.Status)
			return errors.New("db error")
		},
	}
	svc := &lstepCsvImportService{csvImportRepo: repo}
	csv := "foo,bar\n1,2\n"

	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "bad-header.csv", strings.NewReader(csv), 1)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.True(t, createCalled)
}

// TestImportFriendAttributesCSV_ProcessingRecordCreateError は、ヘッダー解析成功後の
// processing レコード作成 (csvImportRepo.Create) が失敗した場合、ラップされたエラーを返すことを確認する。
func TestImportFriendAttributesCSV_ProcessingRecordCreateError(t *testing.T) {
	repo := &mockLstepCsvImportRepo{
		createFn: func(_ context.Context, _ *model.LstepCsvImport) error {
			return errors.New("db error")
		},
	}
	svc := &lstepCsvImportService{csvImportRepo: repo}
	csv := "line_user_id\nU1\n"

	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "valid.csv", strings.NewReader(csv), 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create csv import record")
}

// TestImportFriendAttributesCSV_FindOwnersError は、飼主一覧取得 (ownerRepo.FindAllWithLineUserID) が
// 失敗した場合、markImportFailed が呼ばれ（Status=failed で Update）、エラーを返すことを確認する。
func TestImportFriendAttributesCSV_FindOwnersError(t *testing.T) {
	var markedFailed *model.LstepCsvImport
	repo := &mockLstepCsvImportRepo{
		createFn: func(_ context.Context, _ *model.LstepCsvImport) error {
			return nil
		},
		updateFn: func(_ context.Context, imp *model.LstepCsvImport) error {
			markedFailed = imp
			return nil
		},
	}
	ownerRepo := &mockLstepImportOwnerRepo{
		findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepCsvImportService{csvImportRepo: repo, ownerRepo: ownerRepo}
	csv := "line_user_id\nU1\n"

	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "valid.csv", strings.NewReader(csv), 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find owners")
	require.NotNil(t, markedFailed)
	assert.Equal(t, csvImportStatusFailed, markedFailed.Status)
}


// ---- ListByClinic ----

func TestLstepCsvImportService_ListByClinic(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockLstepCsvImportRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "returns list on success",
			repo: &mockLstepCsvImportRepo{
				listByClinicFn: func(_ context.Context, _ uint64, _ int) ([]*model.LstepCsvImport, error) {
					return []*model.LstepCsvImport{{}, {}}, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "propagates repository error",
			repo: &mockLstepCsvImportRepo{
				listByClinicFn: func(_ context.Context, _ uint64, _ int) ([]*model.LstepCsvImport, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepCsvImportService(nil, tt.repo, nil, nil)

			got, err := svc.ListByClinic(context.Background(), 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// ---- markImportFailed ----

func TestLstepCsvImportService_MarkImportFailed(t *testing.T) {
	t.Run("sets status to failed and calls Update", func(t *testing.T) {
		var updated *model.LstepCsvImport
		repo := &mockLstepCsvImportRepo{
			updateFn: func(_ context.Context, imp *model.LstepCsvImport) error {
				updated = imp
				return nil
			},
		}
		svc := &lstepCsvImportService{csvImportRepo: repo}
		imp := &model.LstepCsvImport{ID: uuid.New(), Status: csvImportStatusProcessing}

		svc.markImportFailed(context.Background(), imp)

		assert.Equal(t, csvImportStatusFailed, imp.Status)
		require.NotNil(t, updated)
		assert.Equal(t, csvImportStatusFailed, updated.Status)
	})

	t.Run("logs but does not panic when update fails", func(t *testing.T) {
		repo := &mockLstepCsvImportRepo{
			updateFn: func(_ context.Context, _ *model.LstepCsvImport) error {
				return errors.New("db error")
			},
		}
		svc := &lstepCsvImportService{csvImportRepo: repo}
		imp := &model.LstepCsvImport{ID: uuid.New(), Status: csvImportStatusProcessing}

		assert.NotPanics(t, func() {
			svc.markImportFailed(context.Background(), imp)
		})
		assert.Equal(t, csvImportStatusFailed, imp.Status)
	})
}
