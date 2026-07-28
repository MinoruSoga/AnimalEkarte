package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- LstepCsvImportRepository モック ----

type mockLstepCsvImportRepo struct {
	createFn       func(ctx context.Context, imp *model.LstepCsvImport) error
	updateFn       func(ctx context.Context, imp *model.LstepCsvImport) error
	findByIDFn     func(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error)
	listByClinicFn func(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

type countingCSVReader struct {
	reader    *strings.Reader
	bytesRead int
}

func (r *countingCSVReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.bytesRead += n
	return n, err
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

// ---- CSV内の候補だけを照合する Owner lookup モック ----

type mockLstepImportOwnerLookup struct {
	findExistingLineUserIDsFn func(ctx context.Context, db *gorm.DB, clinicID uint64, lineUserIDs []string) (map[string]struct{}, error)
}

type mockLstepImportStaffRepo struct {
	findByIDFn func(ctx context.Context, clinicID, staffID uint64) (*model.Staff, error)
}

func (m *mockLstepImportStaffRepo) FindByID(ctx context.Context, clinicID, staffID uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, staffID)
	}
	return nil, errors.New("unexpected staff lookup")
}

func csvImportStaffRepoWithAccount(accountID uint64) *mockLstepImportStaffRepo {
	return &mockLstepImportStaffRepo{
		findByIDFn: func(_ context.Context, clinicID, staffID uint64) (*model.Staff, error) {
			return &model.Staff{ID: staffID, ClinicID: clinicID, AccountID: &accountID}, nil
		},
	}
}

func (m *mockLstepImportOwnerLookup) FindExistingLineUserIDs(
	ctx context.Context,
	db *gorm.DB,
	clinicID uint64,
	lineUserIDs []string,
) (map[string]struct{}, error) {
	if m.findExistingLineUserIDsFn != nil {
		return m.findExistingLineUserIDsFn(ctx, db, clinicID, lineUserIDs)
	}
	return map[string]struct{}{}, nil
}

// TestImportFriendAttributesCSV_FileTooLarge: 50MB+1 バイトの reader で WrapInvalidInput を返すことを確認。
// csvImportRepo.Create 到達前 (line 69-71) に return するため nil repo で safe。
func TestImportFriendAttributesCSV_FileTooLarge(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 50*1024*1024+1)
	svc := &lstepCsvImportService{
		db:            nil,
		csvImportRepo: nil,
		ownerLookup:   nil,
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
		ownerLookup:   nil,
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

func TestImportFriendAttributesCSV_RejectsTooManyRowsBeforeActorLookup(t *testing.T) {
	csv := "line_user_id\n" + strings.Repeat("U1\n", maxCSVDataRows+1)
	svc := &lstepCsvImportService{}

	_, err := svc.ImportFriendAttributesCSV(
		context.Background(), 100, "too-many-rows.csv", strings.NewReader(csv), 7,
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "row count exceeds")
}

func TestImportFriendAttributesCSV_RejectsOversizedCSVShapeBeforeActorLookup(t *testing.T) {
	headers := make([]string, maxCSVColumns+1)
	headers[0] = "line_user_id"
	for i := 1; i < len(headers); i++ {
		headers[i] = "extra"
	}

	tests := []struct {
		name       string
		csv        string
		wantDetail string
	}{
		{
			name:       "too many columns",
			csv:        strings.Join(headers, ",") + "\n",
			wantDetail: "column count exceeds",
		},
		{
			name:       "cell too large",
			csv:        "line_user_id\n" + strings.Repeat("x", maxCSVCellBytes+1) + "\n",
			wantDetail: "cell exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &lstepCsvImportService{}
			_, err := svc.ImportFriendAttributesCSV(
				context.Background(), 100, "oversized.csv", strings.NewReader(tt.csv), 7,
			)

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Contains(t, err.Error(), tt.wantDetail)
		})
	}
}

func TestPreflightCSVShape_RejectsCommaBombBeforeReadingWholeRecord(t *testing.T) {
	const bombBytes = 1024 * 1024
	reader := &countingCSVReader{reader: strings.NewReader(strings.Repeat(",", bombBytes))}

	err := preflightCSVShape(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "column count exceeds")
	assert.Less(t, reader.bytesRead, bombBytes/10, "preflight must stop before buffering the full record")
}

func TestCSVImportErrorCollector_CapsPersistedEntriesAndTracksTotal(t *testing.T) {
	collector := newCSVImportErrorCollector()
	for i := 0; i < maxCSVErrorLogEntries+5; i++ {
		collector.Add(csvImportErrorEntry{Row: i + 2, Reason: csvErrorReasonUnknownLineUserID})
	}

	assert.Equal(t, maxCSVErrorLogEntries+5, collector.total)
	assert.Len(t, collector.entries, maxCSVErrorLogEntries)
	encoded, err := json.Marshal(collector.entries)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), `"line_user_id":`)
}

func TestImportFriendAttributesCSV_StoresResolvedAccountID(t *testing.T) {
	const clinicID = uint64(100)
	const staffID = uint64(7)
	const accountID = uint64(9007)
	var created *model.LstepCsvImport
	repo := &mockLstepCsvImportRepo{
		createFn: func(_ context.Context, imp *model.LstepCsvImport) error {
			created = imp
			return errors.New("stop after capture")
		},
	}
	svc := &lstepCsvImportService{
		csvImportRepo: repo,
		staffRepo:     csvImportStaffRepoWithAccount(accountID),
	}

	_, err := svc.ImportFriendAttributesCSV(
		context.Background(), clinicID, "valid.csv", strings.NewReader("line_user_id\nU1\n"), staffID,
	)

	require.Error(t, err)
	require.NotNil(t, created)
	assert.Equal(t, accountID, created.UploadedByUserID)
}

func TestImportFriendAttributesCSV_RejectsMissingActorDependency(t *testing.T) {
	svc := &lstepCsvImportService{}
	var err error

	assert.NotPanics(t, func() {
		_, err = svc.ImportFriendAttributesCSV(
			context.Background(), 100, "valid.csv", strings.NewReader("line_user_id\nU1\n"), 7,
		)
	})
	require.Error(t, err)
}

func TestImportFriendAttributesCSV_RejectsStaffWithoutAccount(t *testing.T) {
	staffRepo := &mockLstepImportStaffRepo{
		findByIDFn: func(_ context.Context, clinicID, staffID uint64) (*model.Staff, error) {
			return &model.Staff{ID: staffID, ClinicID: clinicID}, nil
		},
	}
	svc := &lstepCsvImportService{staffRepo: staffRepo}

	_, err := svc.ImportFriendAttributesCSV(
		context.Background(), 100, "valid.csv", strings.NewReader("line_user_id\nU1\n"), 7,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not linked to an account")
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
	svc := &lstepCsvImportService{csvImportRepo: repo, staffRepo: csvImportStaffRepoWithAccount(1)}
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
	svc := &lstepCsvImportService{csvImportRepo: repo, staffRepo: csvImportStaffRepoWithAccount(1)}
	csv := "line_user_id\nU1\n"

	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "valid.csv", strings.NewReader(csv), 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create csv import record")
}

// TestImportFriendAttributesCSV_FindOwnersError は、CSV内の飼主照合が
// 失敗した場合、markImportFailed が呼ばれ（Status=failed で Update）、エラーを返すことを確認する。
func TestImportFriendAttributesCSV_FindOwnersError(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
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
	ownerLookup := &mockLstepImportOwnerLookup{
		findExistingLineUserIDsFn: func(_ context.Context, _ *gorm.DB, _ uint64, _ []string) (map[string]struct{}, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepCsvImportService{
		db:            db,
		csvImportRepo: repo,
		ownerLookup:   ownerLookup,
		staffRepo:     csvImportStaffRepoWithAccount(1),
	}
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
