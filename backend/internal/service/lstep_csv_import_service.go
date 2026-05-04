package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	maxCSVSizeBytes           int64  = 50 * 1024 * 1024
	csvTypeFriendAttribute    string = "friend_attribute"
	csvImportStatusProcessing string = "processing"
	csvImportStatusCompleted  string = "completed"
	csvImportStatusFailed     string = "failed"
)

// LstepCsvImportService は Lステップ CSV インポートサービス（FEAT-385）。
type LstepCsvImportService interface {
	// ImportFriendAttributesCSV は Lステップ友だち属性 CSV をインポートし、インポート ID を返す。
	ImportFriendAttributesCSV(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByUserID *uint64) (uuid.UUID, error)
	// GetByID はクリニックスコープでインポート履歴を返す。
	GetByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error)
	// ListByClinic はクリニックスコープで最新順にインポート履歴一覧を返す。
	ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

type lstepCsvImportService struct {
	db            *gorm.DB
	csvImportRepo repository.LstepCsvImportRepository
	snapshotRepo  repository.LstepFriendAttributeSnapshotRepository
	ownerRepo     repository.OwnerRepository
}

// NewLstepCsvImportService は LstepCsvImportService を初期化して返す。
func NewLstepCsvImportService(
	db *gorm.DB,
	csvImportRepo repository.LstepCsvImportRepository,
	snapshotRepo repository.LstepFriendAttributeSnapshotRepository,
	ownerRepo repository.OwnerRepository,
) LstepCsvImportService {
	return &lstepCsvImportService{
		db:            db,
		csvImportRepo: csvImportRepo,
		snapshotRepo:  snapshotRepo,
		ownerRepo:     ownerRepo,
	}
}

func (s *lstepCsvImportService) ImportFriendAttributesCSV(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByUserID *uint64) (uuid.UUID, error) {
	// 1. サイズ制限付き読み込み
	limited := io.LimitReader(fileReader, maxCSVSizeBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return uuid.Nil, apperrors.WrapInvalidInput("failed to read CSV file")
	}
	if int64(len(raw)) > maxCSVSizeBytes {
		return uuid.Nil, apperrors.WrapInvalidInput("CSV file exceeds 50MB limit")
	}

	// 2. 文字コード判定・UTF-8 変換
	decoded, err := decodeCsvBytes(raw)
	if err != nil {
		return uuid.Nil, apperrors.WrapInvalidInput("failed to decode CSV encoding")
	}

	// 3. CSV パース
	reader := csv.NewReader(strings.NewReader(decoded))
	reader.LazyQuotes = true
	allRecords, err := reader.ReadAll()
	if err != nil {
		return uuid.Nil, apperrors.WrapInvalidInput("failed to parse CSV: " + err.Error())
	}
	if len(allRecords) == 0 {
		return uuid.Nil, apperrors.WrapInvalidInput("CSV file is empty")
	}

	// 4. ヘッダー解析（失敗時は failed レコードを作成して終了）
	colIdx, err := resolveCsvHeaders(allRecords[0])
	if err != nil {
		imp := &model.LstepCsvImport{
			ClinicID:         clinicID,
			CsvType:          csvTypeFriendAttribute,
			FileName:         fileName,
			UploadedByUserID: uploadedByUserID,
			Status:           csvImportStatusFailed,
		}
		if createErr := s.csvImportRepo.Create(ctx, imp); createErr != nil {
			slog.ErrorContext(ctx, "failed to create failed csv import record", "error", createErr, "clinic_id", clinicID)
		}
		return uuid.Nil, apperrors.WrapInvalidInput(err.Error())
	}

	// 5. processing レコード作成（メイン TX 外 — TX 失敗時も残る）
	imp := &model.LstepCsvImport{
		ClinicID:         clinicID,
		CsvType:          csvTypeFriendAttribute,
		FileName:         fileName,
		UploadedByUserID: uploadedByUserID,
		Status:           csvImportStatusProcessing,
	}
	if err := s.csvImportRepo.Create(ctx, imp); err != nil {
		slog.ErrorContext(ctx, "failed to create csv import record", "error", err, "clinic_id", clinicID)
		return uuid.Nil, apperrors.Wrap(err, "failed to create csv import record")
	}

	// 6. 飼主 line_user_id → owner_id マップ構築（TX 外・読み取り専用）
	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owners with line_user_id", "error", err, "clinic_id", clinicID)
		s.markImportFailed(ctx, imp)
		return uuid.Nil, apperrors.Wrap(err, "failed to find owners")
	}
	ownerIDByLineUserID := make(map[string]uint64, len(owners))
	for i := range owners {
		o := &owners[i]
		if o.LineUserID != nil && *o.LineUserID != "" {
			ownerIDByLineUserID[*o.LineUserID] = o.ID
		}
	}

	// 7. データ行ループ — スナップショット収集
	now := time.Now().In(jst)
	dataRows := allRecords[1:]
	snapshots := make([]*model.LstepFriendAttributeSnapshot, 0, len(dataRows))
	var errEntries []csvImportErrorEntry

	for i, row := range dataRows {
		rowNum := i + 2 // ヘッダー行 = 1 なので +2
		snapshot, errEntry := s.parseFriendAttrCSVRow(row, colIdx, rowNum, clinicID, imp.ID, now)
		if errEntry != nil {
			errEntries = append(errEntries, *errEntry)
			continue
		}
		if _, ok := ownerIDByLineUserID[snapshot.LineUserID]; !ok {
			errEntries = append(errEntries, csvImportErrorEntry{
				Row:        rowNum,
				LineUserID: snapshot.LineUserID,
				Reason:     csvErrorReasonUnknownLineUserID,
				Detail:     "no matching owner found for line_user_id",
			})
			continue
		}
		snapshots = append(snapshots, snapshot)
	}

	// 最終集計をレコードに反映
	imp.RowCount = len(dataRows)
	imp.SuccessCount = len(snapshots)
	imp.ErrorCount = len(errEntries)
	imp.Status = csvImportStatusCompleted
	imp.ImportedAt = &now
	if len(errEntries) > 0 {
		errLog, _ := json.Marshal(errEntries)
		imp.ErrorLog = datatypes.JSON(errLog)
	}

	// 8. トランザクション: BulkCreate + インポートレコード更新
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(snapshots) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
				CreateInBatches(snapshots, 100).Error; err != nil {
				return apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "bulk_create")
			}
		}
		if err := tx.Where("clinic_id = ?", clinicID).Save(imp).Error; err != nil {
			return apperrors.FromGORM(err, "lstep_csv_import", imp.ID.String())
		}
		return nil
	})
	if txErr != nil {
		slog.ErrorContext(ctx, "failed to commit csv import transaction", "error", txErr, "import_id", imp.ID)
		s.markImportFailed(ctx, imp)
		return uuid.Nil, apperrors.Wrap(txErr, "failed to save csv import results")
	}

	return imp.ID, nil
}

func (s *lstepCsvImportService) GetByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LstepCsvImport, error) {
	imp, err := s.csvImportRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find csv import")
	}
	return imp, nil
}

func (s *lstepCsvImportService) ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	imports, err := s.csvImportRepo.ListByClinic(ctx, clinicID, limit)
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

// markImportFailed はインポートレコードを failed に更新する（TX 失敗後の補償処理）。
func (s *lstepCsvImportService) markImportFailed(ctx context.Context, imp *model.LstepCsvImport) {
	imp.Status = csvImportStatusFailed
	if err := s.csvImportRepo.Update(ctx, imp); err != nil {
		slog.ErrorContext(ctx, "failed to mark csv import as failed", "error", err, "import_id", imp.ID)
	}
}
