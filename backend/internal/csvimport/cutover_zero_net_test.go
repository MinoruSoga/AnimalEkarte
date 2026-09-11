package csvimport

import (
	"strings"
	"testing"
)

func TestPreflightCutoverBundleZeroNetSignedSplits(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "opposite nonzero tenders", true: "zero tenders rejected"}[invalid], func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				pc, sc := CutoverTableSpecs()[13].Columns, CutoverTableSpecs()[14].Columns
				cash, card := "-100", "100"
				if invalid {
					cash, card = "0", "0"
				}
				f.rows["payments"][0][columnIndex(pc, "billing_amount")] = "0"
				f.rows["payments"][0][columnIndex(pc, "received_amount")] = cash
				f.rows["payment_splits"][0][columnIndex(sc, "amount")] = cash
				f.rows["payment_splits"][0][columnIndex(sc, "received_amount")] = cash
				row := append([]string(nil), f.rows["payment_splits"][0]...)
				row[columnIndex(sc, "id")] = "1000002"
				row[columnIndex(sc, "method")] = "credit_card"
				row[columnIndex(sc, "payment_method_id")] = "{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
				row[columnIndex(sc, "amount")] = card
				row[columnIndex(sc, "received_amount")] = "0"
				f.rows["payment_splits"] = append(f.rows["payment_splits"], row)
				f.manifest.Tables[14].RowCount++
			})
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest, ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1",
			})
			if invalid && (err == nil || !strings.Contains(err.Error(), "amount must not be zero")) {
				t.Fatalf("want zero split rejection, got %v", err)
			}
			if !invalid && err != nil {
				t.Fatalf("valid signed zero-net payment rejected: %v", err)
			}
		})
	}
}

func TestPaymentTargetAllowsOnlyGraphValidatedZeroNet(t *testing.T) {
	if strings.Contains(verifyCutoverPaymentGraphQuery, "OR payment.billing_amount = 0") {
		t.Fatal("zero-net parent must be validated through its nonzero split graph")
	}
	for _, guard := range []string{
		"AND split.amount <> 0",
		"COALESCE(split_summary.split_count, 0) NOT BETWEEN 1 AND 2",
		"COALESCE(split_summary.distinct_method_count, 0) <> COALESCE(split_summary.split_count, 0)",
		"COALESCE(split_summary.split_amount, 0) <> payment.billing_amount",
		"OR NOT COALESCE(split_summary.split_rows_valid, false)",
	} {
		if !strings.Contains(verifyCutoverPaymentGraphQuery, guard) {
			t.Fatalf("missing zero-net safety guard: %s", guard)
		}
	}
}
