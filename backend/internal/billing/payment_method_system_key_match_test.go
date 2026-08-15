package billing

// payment_method_system_key_match_test.go — TASK-ADR003 DB 境界ゲート。
//
// 本番 migration 006 をテスト DB に適用し、payments / payment_splits で
// method ⇔ payment_methods.system_key の不一致 INSERT/UPDATE が拒否され、
// 一致・payment_method_id NULL は許可されることを回帰固定する。
// Service 層の resolvePaymentMethodMasterID だけでは完了扱いにしない。

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// 006 は 2026-07-29 に 001_init.sql §9 へ原文アーカイブされた。独立ファイルは存在しない。
const paymentMethodSystemKeyMatchMigration = "006_payment_method_system_key_match.sql"

// readPaymentMethodSystemKeyMatchMigration は 001_init.sql §9 に統合された
// 006_payment_method_system_key_match.sql の SQL 本文を返す。
func readPaymentMethodSystemKeyMatchMigration(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("../../migrations/001_init.sql") //nolint:gocritic // B5b requires this relative path.
	require.NoError(t, err, "read 001_init.sql")
	initial := string(raw)

	sourceMarker := "-- Source file: " + paymentMethodSystemKeyMatchMigration
	const nextSourceMarker = "-- Source file: 007_owners_clinic_phone_unique.sql"
	start := strings.Index(initial, sourceMarker)
	require.GreaterOrEqual(t, start, 0, "001_init.sql must contain the archived %s", paymentMethodSystemKeyMatchMigration)

	endOffset := strings.Index(initial[start:], "\n"+nextSourceMarker)
	require.Greater(t, endOffset, 0, "archived %s must end at the 007 source marker", paymentMethodSystemKeyMatchMigration)
	block := initial[start : start+endOffset]

	shaOffset := strings.Index(block, "-- Source SHA-256:")
	require.GreaterOrEqual(t, shaOffset, 0, "archived %s must contain its SHA-256 header", paymentMethodSystemKeyMatchMigration)
	bodyOffset := strings.Index(block[shaOffset:], "\n")
	require.GreaterOrEqual(t, bodyOffset, 0, "archived migration metadata must end before its SQL body")

	return block[shaOffset+bodyOffset+1:]
}

var applyPaymentMethodSystemKeyMatchOnce sync.Once

func setupPaymentMethodSystemKeyMatchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.PaymentMethodMaster{},
		&model.Payment{},
		&model.PaymentSplit{},
		&model.Billing{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE payment_splits CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE payments CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE billings CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE payment_methods CASCADE").Error)

	applyPaymentMethodSystemKeyMatchOnce.Do(func() {
		applyPaymentMethodSystemKeyMatchMigration(t, db)
	})
	return db
}

func applyPaymentMethodSystemKeyMatchMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	raw := readPaymentMethodSystemKeyMatchMigration(t)

	for _, stmt := range splitPaymentMethodSystemKeyMatchSQL(raw) {
		require.NoError(t, db.Exec(stmt).Error, "apply migration stmt: %s", stmt)
	}
}

func splitPaymentMethodSystemKeyMatchSQL(sql string) []string {
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		if idx := strings.Index(line, "--"); idx >= 0 {
			// Keep $$ function bodies intact; only strip trailing line comments outside strings.
			// Migration 006 comments are full-line or trailing outside of $$ blocks.
			line = line[:idx]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// Function body uses $$ ... $$; split on ';' only outside dollar-quoted sections.
	text := b.String()
	var out []string
	var stmt strings.Builder
	inDollar := false
	for i := 0; i < len(text); i++ {
		if text[i] == '$' && i+1 < len(text) && text[i+1] == '$' {
			stmt.WriteString("$$")
			inDollar = !inDollar
			i++
			continue
		}
		if text[i] == ';' && !inDollar {
			s := strings.TrimSpace(stmt.String())
			if s != "" {
				out = append(out, s)
			}
			stmt.Reset()
			continue
		}
		stmt.WriteByte(text[i])
	}
	if s := strings.TrimSpace(stmt.String()); s != "" {
		out = append(out, s)
	}
	return out
}

func makeSystemKeyPaymentMethod(t *testing.T, db *gorm.DB, clinicID uint64, name, systemKey string) *model.PaymentMethodMaster {
	t.Helper()
	key := systemKey
	m := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         name,
		SystemKey:    &key,
		DisplayOrder: 1,
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makeSystemKeyBilling(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		TotalAmount:   1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

func assertSystemKeyMismatchError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	msg := err.Error()
	assert.True(t,
		strings.Contains(msg, "支払方法の不整合") ||
			strings.Contains(msg, "23514") ||
			strings.Contains(msg, "system_key"),
		"expected system_key match failure, got: %v", err)
}

func TestPaymentMethodSystemKeyMatch_PaymentSplits_InsertValid(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	cash := makeSystemKeyPaymentMethod(t, db, clinicID, "現金", "cash")
	billing := makeSystemKeyBilling(t, db, clinicID)

	split := model.PaymentSplit{
		ClinicID:        clinicID,
		BillingID:       billing.ID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: ptrUint64(cash.ID),
		Amount:          1000,
		ReceivedAmount:  1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&split).Error)
	assert.NotZero(t, split.ID)
}

func TestPaymentMethodSystemKeyMatch_PaymentSplits_InsertMismatchRejected(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	_ = makeSystemKeyPaymentMethod(t, db, clinicID, "現金", "cash")
	card := makeSystemKeyPaymentMethod(t, db, clinicID, "カード", "credit_card")
	billing := makeSystemKeyBilling(t, db, clinicID)

	err := db.WithContext(context.Background()).Create(&model.PaymentSplit{
		ClinicID:        clinicID,
		BillingID:       billing.ID,
		Method:          model.PaymentMethodCash, // mismatch: method cash, master credit_card
		PaymentMethodID: ptrUint64(card.ID),
		Amount:          1000,
		ReceivedAmount:  1000,
	}).Error
	assertSystemKeyMismatchError(t, err)
}

func TestPaymentMethodSystemKeyMatch_PaymentSplits_InsertNullPaymentMethodIDAllowed(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	billing := makeSystemKeyBilling(t, db, clinicID)

	split := model.PaymentSplit{
		ClinicID:        clinicID,
		BillingID:       billing.ID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: nil,
		Amount:          1000,
		ReceivedAmount:  1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&split).Error)
}

func TestPaymentMethodSystemKeyMatch_PaymentSplits_UpdateMismatchRejected(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	cash := makeSystemKeyPaymentMethod(t, db, clinicID, "現金", "cash")
	card := makeSystemKeyPaymentMethod(t, db, clinicID, "カード", "credit_card")
	billing := makeSystemKeyBilling(t, db, clinicID)

	split := model.PaymentSplit{
		ClinicID:        clinicID,
		BillingID:       billing.ID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: ptrUint64(cash.ID),
		Amount:          1000,
		ReceivedAmount:  1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&split).Error)

	err := db.WithContext(context.Background()).
		Model(&model.PaymentSplit{}).
		Where("id = ?", split.ID).
		Updates(map[string]any{
			"method":            model.PaymentMethodCash,
			"payment_method_id": card.ID,
		}).Error
	assertSystemKeyMismatchError(t, err)
}

func TestPaymentMethodSystemKeyMatch_Payments_InsertValid(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	cash := makeSystemKeyPaymentMethod(t, db, clinicID, "現金", "cash")
	billing := makeSystemKeyBilling(t, db, clinicID)

	payment := model.Payment{
		BillingID:       billing.ID,
		ClinicID:        clinicID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: ptrUint64(cash.ID),
		TotalAmount:     1000,
		BillingAmount:   1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&payment).Error)
	assert.NotZero(t, payment.ID)
}

func TestPaymentMethodSystemKeyMatch_Payments_InsertMismatchRejected(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	card := makeSystemKeyPaymentMethod(t, db, clinicID, "カード", "credit_card")
	billing := makeSystemKeyBilling(t, db, clinicID)

	err := db.WithContext(context.Background()).Create(&model.Payment{
		BillingID:       billing.ID,
		ClinicID:        clinicID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: ptrUint64(card.ID),
		TotalAmount:     1000,
		BillingAmount:   1000,
	}).Error
	assertSystemKeyMismatchError(t, err)
}

func TestPaymentMethodSystemKeyMatch_Payments_InsertNullPaymentMethodIDAllowed(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	billing := makeSystemKeyBilling(t, db, clinicID)

	payment := model.Payment{
		BillingID:       billing.ID,
		ClinicID:        clinicID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: nil,
		TotalAmount:     1000,
		BillingAmount:   1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&payment).Error)
}

func TestPaymentMethodSystemKeyMatch_Payments_UpdateMismatchRejected(t *testing.T) {
	db := setupPaymentMethodSystemKeyMatchTestDB(t)
	const clinicID = uint64(1)
	cash := makeSystemKeyPaymentMethod(t, db, clinicID, "現金", "cash")
	card := makeSystemKeyPaymentMethod(t, db, clinicID, "カード", "credit_card")
	billing := makeSystemKeyBilling(t, db, clinicID)

	payment := model.Payment{
		BillingID:       billing.ID,
		ClinicID:        clinicID,
		Method:          model.PaymentMethodCash,
		PaymentMethodID: ptrUint64(cash.ID),
		TotalAmount:     1000,
		BillingAmount:   1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&payment).Error)

	err := db.WithContext(context.Background()).
		Model(&model.Payment{}).
		Where("id = ?", payment.ID).
		Updates(map[string]any{
			"method":            model.PaymentMethodCash,
			"payment_method_id": card.ID,
		}).Error
	assertSystemKeyMismatchError(t, err)
}

func TestPaymentMethodSystemKeyMatchMigration_StaticContents(t *testing.T) {
	ddl := readPaymentMethodSystemKeyMatchMigration(t)
	normalized := strings.Join(strings.Fields(ddl), " ")

	assert.Contains(t, normalized, "app_private.enforce_payment_method_system_key_match")
	assert.Contains(t, normalized, "BEFORE INSERT OR UPDATE ON payment_splits")
	assert.Contains(t, normalized, "BEFORE INSERT OR UPDATE ON payments")
	assert.Contains(t, normalized, "pm.system_key IS NOT DISTINCT FROM NEW.method::text")
	assert.NotContains(t, strings.ToUpper(ddl), "ON DELETE CASCADE")
}
