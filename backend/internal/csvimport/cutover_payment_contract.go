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
	"path/filepath"
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

func validateCutoverPaymentGraph(sourceDir string, manifest *CutoverManifest) error {
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

	billings := make(map[int64]cutoverBillingFact)
	billingsPath := filepath.Join(sourceDir, billingsTable.File)
	if err := streamCutoverCSV(billingsPath, billingsSpec, billingsTable.SHA256, func(row []string, indexes map[string]int, line int64) error {
		billingID, err := parsePaymentGraphInt("billings", "id", row[indexes["id"]], line)
		if err != nil {
			return err
		}
		if _, duplicate := billings[billingID]; duplicate {
			return fmt.Errorf("table billings column id row %d: duplicate billing id", line)
		}
		totalAmount, err := parsePaymentGraphInt("billings", "total_amount", row[indexes["total_amount"]], line)
		if err != nil {
			return err
		}
		// AE-MIG-NEG-1: Jouto refunds/red-slips keep their sign. Zero is still
		// invalid here because a billing row with a payment graph must carry an amount.
		status := row[indexes["status"]]
		completedAt := row[indexes["completed_at"]]
		switch status {
		case "waiting", "completed", "cancelled", "pending":
		default:
			return fmt.Errorf("table billings column status row %d: billing status is invalid", line)
		}
		if status == "completed" {
			if err := validatePaymentGraphTimestamp("billings", "completed_at", completedAt, line); err != nil {
				return err
			}
		} else if completedAt != "" {
			return fmt.Errorf("table billings column completed_at row %d: non-completed billing must not have a completion timestamp", line)
		}
		billings[billingID] = cutoverBillingFact{
			totalAmount: totalAmount,
			status:      status,
			completedAt: sha256.Sum256([]byte(completedAt)),
		}
		return nil
	}); err != nil {
		return err
	}

	parents := make(map[int64]cutoverPaymentParent)
	paymentsPath := filepath.Join(sourceDir, paymentsTable.File)
	if err := streamCutoverCSV(paymentsPath, paymentsSpec, paymentsTable.SHA256, func(row []string, indexes map[string]int, line int64) error {
		billingID, err := parsePaymentGraphInt("payments", "billing_id", row[indexes["billing_id"]], line)
		if err != nil {
			return err
		}
		billing, ok := billings[billingID]
		if !ok {
			return fmt.Errorf("table payments column billing_id row %d: billing parent is missing", line)
		}
		if billing.status != "completed" {
			return fmt.Errorf("table payments column billing_id row %d: billing status must be completed", line)
		}
		if _, duplicate := parents[billingID]; duplicate {
			return fmt.Errorf("table payments column billing_id row %d: duplicate billing_id", line)
		}
		billingAmount, err := parsePaymentGraphInt("payments", "billing_amount", row[indexes["billing_amount"]], line)
		if err != nil {
			return err
		}
		receivedAmount, err := parsePaymentGraphInt("payments", "received_amount", row[indexes["received_amount"]], line)
		if err != nil {
			return err
		}
		changeAmount, err := parsePaymentGraphInt("payments", "change_amount", row[indexes["change_amount"]], line)
		if err != nil {
			return err
		}
		if billingAmount == 0 {
			return fmt.Errorf("table payments row %d: payment amounts violate the cutover contract", line)
		}
		var totalAmount int64
		for _, column := range []string{"subtotal", "tax_total", "total_amount", "discount_amount"} {
			amount, err := parsePaymentGraphInt("payments", column, row[indexes[column]], line)
			if err != nil {
				return err
			}
			if column == "total_amount" {
				totalAmount = amount
			}
		}
		if totalAmount != billing.totalAmount {
			return fmt.Errorf("table payments column total_amount row %d: payment snapshot does not match billing", line)
		}
		if _, err := parsePaymentGraphRatio(row[indexes["insurance_ratio"]], line); err != nil {
			return err
		}
		insuranceAmount, err := parsePaymentGraphInt("payments", "insurance_amount", row[indexes["insurance_amount"]], line)
		if err != nil {
			return err
		}
		if insuranceAmount < 0 {
			return fmt.Errorf("table payments column insurance_amount row %d: amount must not be negative", line)
		}
		createdAt := row[indexes["created_at"]]
		if err := validatePaymentGraphTimestamp("payments", "created_at", createdAt, line); err != nil {
			return err
		}
		if billing.completedAt != sha256.Sum256([]byte(createdAt)) {
			return fmt.Errorf("table payments column created_at row %d: completion timestamp does not match billing", line)
		}
		parents[billingID] = cutoverPaymentParent{
			billingAmount:  billingAmount,
			receivedAmount: receivedAmount,
			changeAmount:   changeAmount,
			method:         row[indexes["method"]],
			paidBy:         sha256.Sum256([]byte(row[indexes["paid_by"]])),
			createdAt:      sha256.Sum256([]byte(createdAt)),
		}
		return nil
	}); err != nil {
		return err
	}

	splitsPath := filepath.Join(sourceDir, splitsTable.File)
	if err := streamCutoverCSV(splitsPath, splitsSpec, splitsTable.SHA256, func(row []string, indexes map[string]int, line int64) error {
		billingID, err := parsePaymentGraphInt("payment_splits", "billing_id", row[indexes["billing_id"]], line)
		if err != nil {
			return err
		}
		parent, ok := parents[billingID]
		if !ok {
			return fmt.Errorf("table payment_splits column billing_id row %d: payment parent is missing", line)
		}
		if parent.splitCount >= 2 {
			return fmt.Errorf("table payment_splits row %d: payment has more than two splits", line)
		}
		method := row[indexes["method"]]
		if method == "cash" && parent.hasCash || method == "credit_card" && parent.hasCreditCard {
			return fmt.Errorf("table payment_splits column method row %d: duplicate method for payment", line)
		}
		amount, err := parsePaymentGraphInt("payment_splits", "amount", row[indexes["amount"]], line)
		if err != nil {
			return err
		}
		receivedAmount, err := parsePaymentGraphInt("payment_splits", "received_amount", row[indexes["received_amount"]], line)
		if err != nil {
			return err
		}
		changeAmount, err := parsePaymentGraphInt("payment_splits", "change_amount", row[indexes["change_amount"]], line)
		if err != nil {
			return err
		}
		if amount == 0 {
			return fmt.Errorf("table payment_splits column amount row %d: amount must not be zero", line)
		}
		switch method {
		case "cash":
			if receivedAmount < amount || changeAmount != receivedAmount-amount {
				return fmt.Errorf("table payment_splits row %d: cash arithmetic is invalid", line)
			}
			parent.hasCash = true
			parent.cashReceived, err = addPaymentGraphAmount(parent.cashReceived, receivedAmount)
			if err != nil {
				return fmt.Errorf("table payment_splits row %d: cash amount overflow", line)
			}
			parent.cashChange, err = addPaymentGraphAmount(parent.cashChange, changeAmount)
			if err != nil {
				return fmt.Errorf("table payment_splits row %d: cash change overflow", line)
			}
		case "credit_card":
			if receivedAmount != 0 || changeAmount != 0 {
				return fmt.Errorf("table payment_splits row %d: credit-card arithmetic is invalid", line)
			}
			parent.hasCreditCard = true
		default:
			return fmt.Errorf("table payment_splits column method row %d: unsupported payment method", line)
		}
		if parent.paidBy != sha256.Sum256([]byte(row[indexes["paid_by"]])) ||
			parent.createdAt != sha256.Sum256([]byte(row[indexes["created_at"]])) {
			return fmt.Errorf("table payment_splits row %d: payment parent metadata does not match", line)
		}
		parent.splitAmount, err = addPaymentGraphAmount(parent.splitAmount, amount)
		if err != nil {
			return fmt.Errorf("table payment_splits row %d: split amount overflow", line)
		}
		parent.splitCount++
		parents[billingID] = parent
		return nil
	}); err != nil {
		return err
	}

	for billingID := range parents {
		parent := parents[billingID]
		if parent.splitCount < 1 || parent.splitAmount != parent.billingAmount {
			return fmt.Errorf("table payments: split set does not match payment billing amount")
		}
		wantMethod := "credit_card"
		if parent.hasCash {
			wantMethod = "cash"
		}
		if parent.method != wantMethod ||
			parent.receivedAmount != parent.cashReceived ||
			parent.changeAmount != parent.cashChange {
			return fmt.Errorf("table payments: split set does not match payment summary")
		}
	}
	for billingID, billing := range billings {
		if billing.status == "completed" {
			if _, ok := parents[billingID]; !ok {
				if billing.totalAmount != 0 {
					return fmt.Errorf("table billings: completed billing is missing its payment graph")
				}
			}
		}
	}
	return nil
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
