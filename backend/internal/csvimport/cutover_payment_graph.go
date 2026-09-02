package csvimport

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
)

func loadCutoverBillingFacts(sourceDir string, spec CutoverTableSpec, table CutoverManifestTable) (map[int64]cutoverBillingFact, error) {
	billings := make(map[int64]cutoverBillingFact)
	path := filepath.Join(sourceDir, table.File)
	err := streamCutoverCSV(path, spec, table.SHA256, func(row []string, indexes map[string]int, line int64) error {
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
		status := row[indexes["status"]]
		completedAt := row[indexes["completed_at"]]
		switch status {
		case "waiting", "completed", "cancelled", "pending":
		default:
			return fmt.Errorf("table billings column status row %d: billing status is invalid", line)
		}
		if status != "completed" && completedAt != "" {
			return fmt.Errorf("table billings column completed_at row %d: non-completed billing must not have a completion timestamp", line)
		}
		if err := validateCompletedBillingTimestamp(status, completedAt, line); err != nil {
			return err
		}
		billings[billingID] = cutoverBillingFact{
			totalAmount: totalAmount,
			status:      status,
			completedAt: sha256.Sum256([]byte(completedAt)),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return billings, nil
}

func loadCutoverPaymentParents(
	sourceDir string,
	spec CutoverTableSpec,
	table CutoverManifestTable,
	billings map[int64]cutoverBillingFact,
	provenance CutoverProvenanceContract,
) (map[int64]cutoverPaymentParent, error) {
	parents := make(map[int64]cutoverPaymentParent)
	path := filepath.Join(sourceDir, table.File)
	err := streamCutoverCSV(path, spec, table.SHA256, func(row []string, indexes map[string]int, line int64) error {
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
		if totalAmount != billing.totalAmount && provenance.Mode != CutoverProvenanceLocalRehearsal {
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
	})
	if err != nil {
		return nil, err
	}
	return parents, nil
}

func accumulateCutoverPaymentSplits(
	sourceDir string,
	spec CutoverTableSpec,
	table CutoverManifestTable,
	parents map[int64]cutoverPaymentParent,
) error {
	path := filepath.Join(sourceDir, table.File)
	return streamCutoverCSV(path, spec, table.SHA256, func(row []string, indexes map[string]int, line int64) error {
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
	})
}

func reconcileCutoverPaymentGraph(billings map[int64]cutoverBillingFact, parents map[int64]cutoverPaymentParent) error {
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
		if billing.status != "completed" {
			continue
		}
		if _, ok := parents[billingID]; ok {
			continue
		}
		if billing.totalAmount == 0 {
			continue
		}
		return fmt.Errorf("table billings: completed billing is missing its payment graph")
	}
	return nil
}
