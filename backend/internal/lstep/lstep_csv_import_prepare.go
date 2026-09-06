package lstep

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

func decodePreflightAndReadCSVHeader(tempCSV *os.File) (*csv.Reader, []string, error) {
	preflightReader, err := newDecodedCSVReader(tempCSV)
	if err != nil {
		return nil, nil, err
	}
	if err := preflightCSVShape(preflightReader); err != nil {
		return nil, nil, csvClientError(err)
	}
	decoded, err := newDecodedCSVReader(tempCSV)
	if err != nil {
		return nil, nil, err
	}

	// 3. ヘッダーだけを先に読み、データ行は後段のTX内で1行ずつ処理する。
	reader := csv.NewReader(decoded)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil, apperrors.WrapInvalidInput("CSV file is empty")
	}
	if err != nil {
		return nil, nil, csvClientError(err)
	}
	if err := validateDecodedCSVRecord(header); err != nil {
		return nil, nil, csvClientError(err)
	}
	return reader, header, nil
}

func (s *lstepCsvImportService) resolveCSVImportAccountID(
	ctx context.Context,
	clinicID, uploadedByStaffID uint64,
) (uint64, error) {
	if s.staffRepo == nil {
		return 0, apperrors.WrapInternalServerError("CSV import actor repository is not configured")
	}
	staff, err := s.staffRepo.FindByID(ctx, clinicID, uploadedByStaffID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to resolve CSV import actor")
	}
	if staff == nil || staff.AccountID == nil {
		return 0, apperrors.WrapInternalServerError("CSV import actor is not linked to an account")
	}
	return *staff.AccountID, nil
}

func (s *lstepCsvImportService) createFriendAttributeImport(
	ctx context.Context,
	clinicID uint64,
	fileName string,
	uploadedByAccountID uint64,
	status string,
) (*model.LstepCsvImport, error) {
	imp := &model.LstepCsvImport{
		ClinicID:         clinicID,
		CsvType:          csvTypeFriendAttribute,
		FileName:         fileName,
		UploadedByUserID: uploadedByAccountID,
		Status:           status,
	}
	if err := s.csvImportRepo.Create(ctx, imp); err != nil {
		if status == csvImportStatusFailed {
			slog.ErrorContext(ctx, "failed to create failed csv import record", "error", err, "clinic_id", clinicID)
			return nil, nil
		}
		return nil, apperrors.Wrap(err, "failed to create csv import record")
	}
	return imp, nil
}

func (s *lstepCsvImportService) commitFriendAttributeRows(
	ctx context.Context,
	reader *csv.Reader,
	clinicID uint64,
	imp *model.LstepCsvImport,
	colIdx map[string]int,
) error {
	// 6-8. データ行を逐次処理し、CSV内のLINE User IDだけを100件単位で照合・保存する。
	now := time.Now().In(config.JST)
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result, processErr := s.processFriendAttributeRows(
			ctx, tx, reader, clinicID, imp.ID, colIdx, now,
		)
		if processErr != nil {
			return processErr
		}
		imp.RowCount = result.rowCount
		imp.SuccessCount = result.successCount
		imp.ErrorCount = result.errors.total
		imp.Status = csvImportStatusCompleted
		imp.ImportedAt = &now
		if len(result.errors.entries) > 0 {
			errLog, _ := json.Marshal(result.errors.entries)
			imp.ErrorLog = datatypes.JSON(errLog)
		}
		return updateCsvImportRecordTx(tx, clinicID, imp)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if txErr != nil {
		s.markImportFailed(ctx, imp)
		if apperrors.IsInvalidInput(txErr) {
			return fmt.Errorf("failed to process CSV rows: %w", txErr)
		}
		return apperrors.Wrap(txErr, "failed to save csv import results")
	}
	return nil
}
