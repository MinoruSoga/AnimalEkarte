package csvimport

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const maxCutoverPaymentRows = int64(1_000_000)

type cutoverPaymentParent struct {
	billingAmount  int64
	receivedAmount int64
	changeAmount   int64
	method         string
	paidBy         [sha256.Size]byte
	createdAt      [sha256.Size]byte
	splitCount     int64
	splitAmount    int64
	cashReceived   int64
	cashChange     int64
	hasCash        bool
	hasCreditCard  bool
}

type cutoverBillingFact struct {
	totalAmount int64
	status      string
	completedAt [sha256.Size]byte
}

func validateCutoverPaymentGraph(sourceDir string, manifest *CutoverManifest, provenance CutoverProvenanceContract) error {
	paymentsSpec, paymentsTable, err := cutoverPaymentContractPart(manifest, "payments")
	if err != nil {
		return err
	}
	splitsSpec, splitsTable, err := cutoverPaymentContractPart(manifest, "payment_splits")
	if err != nil {
		return err
	}
	if paymentsTable.RowCount > maxCutoverPaymentRows {
		return fmt.Errorf("table payments: row count exceeds the cutover payment limit")
	}
	if splitsTable.RowCount > paymentsTable.RowCount*2 {
		return fmt.Errorf("table payment_splits: row count exceeds two rows per payment")
	}
	billingsSpec, billingsTable, err := cutoverPaymentContractPart(manifest, "billings")
	if err != nil {
		return err
	}
	billings, err := loadCutoverBillingFacts(sourceDir, billingsSpec, billingsTable)
	if err != nil {
		return err
	}
	parents, err := loadCutoverPaymentParents(
		sourceDir,
		paymentsSpec,
		paymentsTable,
		billings,
		relaxesCutoverPaymentSnapshot(*manifest, provenance.Mode),
	)
	if err != nil {
		return err
	}
	if err := accumulateCutoverPaymentSplits(sourceDir, splitsSpec, splitsTable, parents); err != nil {
		return err
	}
	return reconcileCutoverPaymentGraph(billings, parents)
}

func cutoverPaymentContractPart(manifest *CutoverManifest, tableName string) (CutoverTableSpec, CutoverManifestTable, error) {
	var spec CutoverTableSpec
	for _, candidate := range CutoverTableSpecs() {
		if candidate.Name == tableName {
			spec = candidate
			break
		}
	}
	for _, table := range manifest.Tables {
		if table.Table == tableName {
			return spec, table, nil
		}
	}
	return CutoverTableSpec{}, CutoverManifestTable{}, fmt.Errorf("manifest is missing payment contract table %s", tableName)
}

func streamCutoverCSV(
	path string,
	spec CutoverTableSpec,
	expectedSHA256 string,
	visit func([]string, map[string]int, int64) error,
) error {
	file, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return fmt.Errorf("table %s: open payment CSV: %w", spec.Name, err)
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(file, maxCutoverCSVBytes+1), hash)))
	reader.FieldsPerRecord = len(spec.Columns)
	header, err := reader.Read()
	if err != nil || !reflect.DeepEqual(header, spec.Columns) {
		return fmt.Errorf("table %s: payment CSV header changed after preflight", spec.Name)
	}
	indexes := make(map[string]int, len(header))
	for i, column := range header {
		indexes[column] = i
	}
	var line int64 = 2
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("table %s row %d: read payment CSV", spec.Name, line)
		}
		if err := visit(row, indexes, line); err != nil {
			return err
		}
		line++
	}
	if actual := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(actual, expectedSHA256) {
		return fmt.Errorf("table %s: payment CSV sha256 changed after preflight", spec.Name)
	}
	return nil
}

func parsePaymentGraphInt(table, column, value string, line int64) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("table %s column %s row %d: value must be an integer", table, column, line)
	}
	return parsed, nil
}

func parsePaymentGraphRatio(value string, line int64) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < 0 || parsed > 1 {
		return 0, fmt.Errorf("table payments column insurance_ratio row %d: value must be between 0 and 1", line)
	}
	return parsed, nil
}

func validateCompletedBillingTimestamp(status, completedAt string, line int64) error {
	if status != "completed" {
		return nil
	}
	return validatePaymentGraphTimestamp("billings", "completed_at", completedAt, line)
}

func validatePaymentGraphTimestamp(table, column, value string, line int64) error {
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return fmt.Errorf("table %s column %s row %d: value must be an RFC3339 timestamp", table, column, line)
	}
	return nil
}

func addPaymentGraphAmount(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, fmt.Errorf("amount overflow")
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, fmt.Errorf("amount underflow")
	}
	return left + right, nil
}
