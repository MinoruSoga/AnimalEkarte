package lstep

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type csvImportErrorCollector struct {
	entries []csvImportErrorEntry
	total   int
}

func newCSVImportErrorCollector() *csvImportErrorCollector {
	return &csvImportErrorCollector{entries: make([]csvImportErrorEntry, 0, maxCSVErrorLogEntries)}
}

func (c *csvImportErrorCollector) Add(entry csvImportErrorEntry) {
	c.total++
	if len(c.entries) < maxCSVErrorLogEntries {
		c.entries = append(c.entries, entry)
	}
}

type csvImportProcessingResult struct {
	rowCount     int
	successCount int
	errors       *csvImportErrorCollector
}

type pendingCSVSnapshot struct {
	row      int
	snapshot *model.LstepFriendAttributeSnapshot
}

func (s *lstepCsvImportService) processFriendAttributeRows(
	ctx context.Context,
	tx *gorm.DB,
	reader *csv.Reader,
	clinicID uint64,
	importID uuid.UUID,
	colIdx map[string]int,
	now time.Time,
) (*csvImportProcessingResult, error) {
	result := &csvImportProcessingResult{errors: newCSVImportErrorCollector()}
	pendingBatch := make([]pendingCSVSnapshot, 0, csvSnapshotBatchSize)
	flushPending := func() error {
		if len(pendingBatch) == 0 {
			return nil
		}
		lineUserIDs := make([]string, 0, len(pendingBatch))
		for _, pending := range pendingBatch {
			lineUserIDs = append(lineUserIDs, pending.snapshot.LineUserID)
		}
		existing, err := s.ownerLookup.FindExistingLineUserIDs(ctx, tx, clinicID, lineUserIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to find owners")
		}
		snapshotBatch := make([]*model.LstepFriendAttributeSnapshot, 0, len(pendingBatch))
		for _, pending := range pendingBatch {
			if _, ok := existing[pending.snapshot.LineUserID]; !ok {
				result.errors.Add(csvImportErrorEntry{
					Row:    pending.row,
					Reason: csvErrorReasonUnknownLineUserID,
					Detail: "no matching owner found for line_user_id",
				})
				continue
			}
			snapshotBatch = append(snapshotBatch, pending.snapshot)
			result.successCount++
		}
		if len(snapshotBatch) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
				CreateInBatches(snapshotBatch, csvSnapshotBatchSize).Error; err != nil {
				return apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "bulk_create")
			}
		}
		pendingBatch = make([]pendingCSVSnapshot, 0, csvSnapshotBatchSize)
		return nil
	}

	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, apperrors.WrapInvalidInput("failed to parse CSV: " + err.Error())
		}
		if result.rowCount >= maxCSVDataRows {
			return nil, apperrors.WrapInvalidInput("CSV row count exceeds limit")
		}
		if err := validateDecodedCSVRecord(row); err != nil {
			return nil, apperrors.WrapInvalidInput(err.Error())
		}

		result.rowCount++
		rowNum := result.rowCount + 1 // ヘッダー行 = 1
		snapshot, errEntry := s.parseFriendAttrCSVRow(row, colIdx, rowNum, clinicID, importID, now)
		if errEntry != nil {
			result.errors.Add(*errEntry)
			continue
		}
		pendingBatch = append(pendingBatch, pendingCSVSnapshot{row: rowNum, snapshot: snapshot})
		if len(pendingBatch) == csvSnapshotBatchSize {
			if err := flushPending(); err != nil {
				return nil, err
			}
		}
	}
	if err := flushPending(); err != nil {
		return nil, err
	}
	return result, nil
}
