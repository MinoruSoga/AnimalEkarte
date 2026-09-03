package lstep

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	maxCSVSizeBytes           int64  = 50 * 1024 * 1024
	maxCSVUploadRequestBytes  int64  = maxCSVSizeBytes + 1024*1024
	maxCSVDataRows                   = 100_000
	maxCSVColumns                    = 64
	maxCSVCellBytes                  = 64 * 1024
	maxCSVErrorLogEntries            = 100
	csvSnapshotBatchSize             = 100
	csvTypeFriendAttribute    string = "friend_attribute"
	csvImportStatusProcessing string = "processing"
	csvImportStatusCompleted  string = "completed"
	csvImportStatusFailed     string = "failed"
)

// LstepCsvImportService は Lステップ CSV インポートサービス（FEAT-385）。
type LstepCsvImportService interface {
	// ImportFriendAttributesCSV は Lステップ友だち属性 CSV をインポートし、インポートレコードを返す。
	ImportFriendAttributesCSV(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByStaffID uint64) (*model.LstepCsvImport, error)
	// ListByClinic はクリニックスコープで最新順にインポート履歴一覧を返す。
	ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

type lstepCsvImportService struct {
	db            *gorm.DB
	csvImportRepo csvImportRepository
	ownerLookup   csvImportOwnerLookup
	staffRepo     csvImportStaffRepository
}

type csvImportRepository interface {
	Create(ctx context.Context, csvImport *model.LstepCsvImport) error
	Update(ctx context.Context, csvImport *model.LstepCsvImport) error
	FindAllByClinicID(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

type csvImportOwnerLookup interface {
	FindExistingLineUserIDs(ctx context.Context, db *gorm.DB, clinicID uint64, lineUserIDs []string) (map[string]struct{}, error)
}

type csvImportStaffRepository interface {
	FindByID(ctx context.Context, clinicID, staffID uint64) (*model.Staff, error)
}

// NewLstepCsvImportService は LstepCsvImportService を初期化して返す。
func NewLstepCsvImportService(
	db *gorm.DB,
	csvImportRepo csvImportRepository,
	ownerLookup csvImportOwnerLookup,
	staffRepo csvImportStaffRepository,
) LstepCsvImportService {
	return &lstepCsvImportService{
		db:            db,
		csvImportRepo: csvImportRepo,
		ownerLookup:   ownerLookup,
		staffRepo:     staffRepo,
	}
}

func (s *lstepCsvImportService) ImportFriendAttributesCSV(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByStaffID uint64) (*model.LstepCsvImport, error) {
	// 1-2. 50MiBまでを一時ファイルへspoolし、decode後のstreamを
	// encoding/csvより先に字句走査して列・セル・行上限を検証する。
	tempCSV, err := spoolCSVToTemp(fileReader)
	if err != nil {
		return nil, err
	}
	defer cleanupTempCSV(tempCSV)
	reader, header, err := decodePreflightAndReadCSVHeader(tempCSV)
	if err != nil {
		return nil, err
	}

	uploadedByAccountID, err := s.resolveCSVImportAccountID(ctx, clinicID, uploadedByStaffID)
	if err != nil {
		return nil, err
	}

	// 4. ヘッダー解析（失敗時は failed レコードを作成して終了）
	colIdx, err := resolveCsvHeaders(header)
	if err != nil {
		s.createFriendAttributeImport(ctx, clinicID, fileName, uploadedByAccountID, csvImportStatusFailed) //nolint:errcheck // failed-import row is best-effort; header error is returned below
		return nil, csvClientError(err)
	}

	// 5. processing レコード作成（メイン TX 外 — TX 失敗時も残る）
	imp, err := s.createFriendAttributeImport(ctx, clinicID, fileName, uploadedByAccountID, csvImportStatusProcessing)
	if err != nil {
		return nil, err
	}
	if s.db == nil || s.ownerLookup == nil {
		s.markImportFailed(ctx, imp)
		return nil, apperrors.WrapInternalServerError("CSV import owner lookup is not configured")
	}

	if err := s.commitFriendAttributeRows(ctx, reader, clinicID, imp, colIdx); err != nil {
		return nil, err
	}
	return imp, nil
}

func (s *lstepCsvImportService) ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	imports, err := s.csvImportRepo.FindAllByClinicID(ctx, clinicID, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list csv imports", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list csv imports")
	}
	return imports, nil
}

// parseFriendAttrCSVRow は 1 データ行からスナップショットを生成する。エラー時は errEntry を返す。
func (s *lstepCsvImportService) parseFriendAttrCSVRow(row []string, colIdx map[string]int, rowNum int, clinicID uint64, importID uuid.UUID, snapshotAt time.Time) (*model.LstepFriendAttributeSnapshot, *csvImportErrorEntry) {
	getCol := func(key string) string {
		idx, ok := colIdx[key]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	lineUserID := getCol("line_user_id")
	if lineUserID == "" {
		return nil, &csvImportErrorEntry{
			Row:    rowNum,
			Reason: csvErrorReasonMissingLineID,
			Detail: "line_user_id column is empty",
		}
	}

	snapshot := &model.LstepFriendAttributeSnapshot{
		ClinicID:        clinicID,
		LineUserID:      lineUserID,
		SnapshotTakenAt: snapshotAt,
		CsvImportID:     &importID,
	}

	if v := getCol("display_name"); v != "" {
		snapshot.DisplayName = &v
	}
	if v := getCol("traffic_source"); v != "" {
		snapshot.TrafficSource = &v
	}
	if v := getCol("block_status"); v != "" {
		snapshot.BlockStatus = &v
	}
	snapshot.RegisteredAt = parseCsvTime(getCol("registered_at"))
	snapshot.LastMessageAt = parseCsvTime(getCol("last_message_at"))

	if tags := splitMultiValue(getCol("tags")); len(tags) > 0 {
		b, _ := json.Marshal(tags)
		snapshot.Tags = datatypes.JSON(b)
	}
	if scenarios := splitMultiValue(getCol("scenarios")); len(scenarios) > 0 {
		b, _ := json.Marshal(scenarios)
		snapshot.Scenarios = datatypes.JSON(b)
	}

	return snapshot, nil
}

// updateCsvImportRecordTx は最終集計結果を呼び出し元のトランザクション内で永続化する。
// NOTE: GORM の Save() は imp.ID (主キー) がセット済みだと、事前に連結した .Where(...) 条件を
// 無視して主キーのみで UPDATE を発行する（clinic_id 境界がクロステナント書き込みに対して無力化
// される既知の落とし穴）。同じトランザクション内で完結させる必要があるため
// LstepCsvImportRepository.Update は使わず、Model+Where+Updates(map) で明示的に id と clinic_id
// 両方を WHERE 句に効かせる（P4: clinicScope on Update, MANDATORY）。
func updateCsvImportRecordTx(tx *gorm.DB, clinicID uint64, imp *model.LstepCsvImport) error {
	result := tx.Model(&model.LstepCsvImport{}).
		Where("id = ? AND clinic_id = ?", imp.ID, clinicID).
		Updates(map[string]any{
			"row_count":     imp.RowCount,
			"success_count": imp.SuccessCount,
			"error_count":   imp.ErrorCount,
			"status":        imp.Status,
			"error_log":     imp.ErrorLog,
			"imported_at":   imp.ImportedAt,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_csv_import", imp.ID.String())
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_csv_import", imp.ID.String())
	}
	return nil
}

// markImportFailed はインポートレコードを failed に更新する（TX 失敗後の補償処理）。
func (s *lstepCsvImportService) markImportFailed(ctx context.Context, imp *model.LstepCsvImport) {
	imp.Status = csvImportStatusFailed
	if err := s.csvImportRepo.Update(ctx, imp); err != nil {
		slog.ErrorContext(ctx, "failed to mark csv import as failed", "error", err, "import_id", imp.ID)
	}
}
