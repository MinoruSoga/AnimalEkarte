package csvimport

import (
	"context"
	"fmt"
)

const verifyCutoverPaymentGraphQuery = `
WITH payment_rows AS MATERIALIZED (
  SELECT payment.*, billing.clinic_id AS billing_clinic_id
  FROM payments payment
  LEFT JOIN billings billing ON billing.id = payment.billing_id
  WHERE payment.id >= $1 AND payment.id < $2
),
split_rows AS MATERIALIZED (
  SELECT
    split.*,
    payment.id AS payment_id,
    payment.paid_by AS payment_paid_by,
    payment.created_at AS payment_created_at,
    split_billing.clinic_id AS split_billing_clinic_id,
    split_staff.clinic_id AS split_staff_clinic_id
  FROM payment_splits split
  JOIN payment_rows payment ON payment.billing_id = split.billing_id
  LEFT JOIN billings split_billing ON split_billing.id = split.billing_id
  LEFT JOIN staffs split_staff ON split_staff.id = split.paid_by
),
split_summaries AS MATERIALIZED (
  SELECT
    split.payment_id,
    count(*) AS split_count,
    count(DISTINCT split.method) AS distinct_method_count,
    COALESCE(sum(split.amount), 0) AS split_amount,
    COALESCE(sum(split.received_amount) FILTER (WHERE split.method = 'cash'), 0) AS cash_received,
    COALESCE(sum(split.change_amount) FILTER (WHERE split.method = 'cash'), 0) AS cash_change,
    COALESCE(bool_or(split.method = 'cash'), false) AS has_cash,
    COALESCE(bool_and(COALESCE(
      split.clinic_id = $3
      AND split.split_billing_clinic_id = $3
      AND (
        (split.method = 'cash' AND split.payment_method_id = $4)
        OR (split.method = 'credit_card' AND split.payment_method_id = $5)
      )
      AND (split.paid_by IS NULL OR split.split_staff_clinic_id = $3)
      AND split.paid_by IS NOT DISTINCT FROM split.payment_paid_by
      AND split.created_at IS NOT NULL
      AND split.created_at IS NOT DISTINCT FROM split.payment_created_at
      AND split.amount > 0
      AND split.received_amount IS NOT NULL
      AND split.change_amount IS NOT NULL
      AND (
        (split.method = 'cash'
          AND split.received_amount >= split.amount
          AND split.change_amount = split.received_amount - split.amount)
        OR (split.method = 'credit_card'
          AND split.received_amount = 0
          AND split.change_amount = 0)
      ),
      false
    )), false) AS split_rows_valid
  FROM split_rows split
  GROUP BY split.payment_id
),
payment_violations AS (
  SELECT payment.id
  FROM payment_rows payment
  LEFT JOIN staffs payment_staff ON payment_staff.id = payment.paid_by
  LEFT JOIN split_summaries split_summary ON split_summary.payment_id = payment.id
  WHERE payment.billing_clinic_id IS DISTINCT FROM $3
    OR payment.subtotal IS NULL
    OR payment.subtotal < 0
    OR payment.tax_total IS NULL
    OR payment.tax_total < 0
    OR payment.total_amount IS NULL
    OR payment.total_amount < 0
    OR payment.billing_amount IS NULL
    OR payment.billing_amount <= 0
    OR payment.received_amount IS NULL
    OR payment.received_amount < 0
    OR payment.change_amount IS NULL
    OR payment.change_amount < 0
    OR payment.created_at IS NULL
    OR NOT COALESCE(
      (payment.method = 'cash' AND payment.payment_method_id = $4)
      OR (payment.method = 'credit_card' AND payment.payment_method_id = $5),
      false
    )
    OR (payment.paid_by IS NOT NULL AND payment_staff.clinic_id IS DISTINCT FROM $3)
    OR COALESCE(split_summary.split_count, 0) NOT BETWEEN 1 AND 2
    OR COALESCE(split_summary.distinct_method_count, 0) <> COALESCE(split_summary.split_count, 0)
    OR COALESCE(split_summary.split_amount, 0) <> payment.billing_amount
    OR COALESCE(split_summary.cash_received, 0) <> payment.received_amount
    OR COALESCE(split_summary.cash_change, 0) <> payment.change_amount
    OR payment.method IS DISTINCT FROM
      CASE WHEN COALESCE(split_summary.has_cash, false) THEN 'cash'::payment_method ELSE 'credit_card'::payment_method END
    OR NOT COALESCE(split_summary.split_rows_valid, false)
),
orphan_split_violations AS (
  SELECT split.id
  FROM payment_splits split
  LEFT JOIN payment_rows payment ON payment.billing_id = split.billing_id
  WHERE split.id >= $1
    AND split.id < $2
    AND payment.id IS NULL
)
SELECT count(*)
FROM (
  SELECT id FROM payment_violations
  UNION ALL
  SELECT id FROM orphan_split_violations
) violation`

func verifyCutoverPaymentGraph(
	ctx context.Context,
	q cutoverQuerier,
	manifest CutoverManifest,
	seeds CutoverSeedIDs,
) error {
	var violations int64
	if err := q.QueryRow(
		ctx,
		verifyCutoverPaymentGraphQuery,
		manifest.IDBand.NonOwnerIDOffset,
		manifest.IDBand.EndExclusive,
		seeds.ClinicID,
		seeds.CashPaymentMethodID,
		seeds.CreditCardPaymentMethodID,
	).Scan(&violations); err != nil {
		return fmt.Errorf("verify payment graph")
	}
	if violations != 0 {
		return fmt.Errorf("payment graph contains %d row(s) without a valid payment parent or method binding", violations)
	}
	return nil
}

const paymentsBillingUniqueIndexQuery = `SELECT EXISTS (
  SELECT 1
  FROM pg_index index_definition
  JOIN pg_class target_table ON target_table.oid = index_definition.indrelid
  JOIN pg_namespace target_schema ON target_schema.oid = target_table.relnamespace
  WHERE target_schema.nspname = 'public'
    AND target_table.relname = 'payments'
    AND index_definition.indisunique = true
    AND index_definition.indisvalid = true
    AND index_definition.indisready = true
    AND index_definition.indpred IS NULL
    AND index_definition.indexprs IS NULL
    AND index_definition.indnkeyatts = 1
    AND index_definition.indnatts = 1
    AND (
      SELECT array_agg(attribute.attname::text ORDER BY key_column.ordinality)
      FROM unnest(index_definition.indkey) WITH ORDINALITY AS key_column(attnum, ordinality)
      JOIN pg_attribute attribute
        ON attribute.attrelid = target_table.oid
       AND attribute.attnum = key_column.attnum
    ) = ARRAY['billing_id']::text[]
)
`

const paymentMethodSystemKeyUniqueIndexQuery = `SELECT EXISTS (
  SELECT 1
  FROM pg_index index_definition
  JOIN pg_class target_table ON target_table.oid = index_definition.indrelid
  JOIN pg_namespace target_schema ON target_schema.oid = target_table.relnamespace
  WHERE target_schema.nspname = 'public'
    AND target_table.relname = 'payment_methods'
    AND index_definition.indisunique = true
    AND index_definition.indisvalid = true
    AND index_definition.indisready = true
    AND index_definition.indexprs IS NULL
    AND index_definition.indnkeyatts = 2
    AND index_definition.indnatts = 2
    AND (
      SELECT array_agg(attribute.attname::text ORDER BY key_column.ordinality)
      FROM unnest(index_definition.indkey) WITH ORDINALITY AS key_column(attnum, ordinality)
      JOIN pg_attribute attribute
        ON attribute.attrelid = target_table.oid
       AND attribute.attnum = key_column.attnum
    ) = ARRAY['clinic_id', 'system_key']::text[]
    AND regexp_replace(
      lower(pg_get_expr(index_definition.indpred, index_definition.indrelid)),
      '[[:space:]()]',
      '',
      'g'
    ) IN (
      'system_keyisnotnullanddeleted_atisnull',
      'deleted_atisnullandsystem_keyisnotnull'
    )
)
`

func validateCutoverUniqueIndexes(ctx context.Context, q cutoverQuerier) error {
	var paymentsBillingUnique bool
	if err := q.QueryRow(ctx, paymentsBillingUniqueIndexQuery).Scan(&paymentsBillingUnique); err != nil {
		return fmt.Errorf("inspect target unique index for payments.billing_id")
	}
	if !paymentsBillingUnique {
		return fmt.Errorf("target requires a valid unique index on payments.billing_id")
	}

	var paymentMethodSystemKeyUnique bool
	if err := q.QueryRow(ctx, paymentMethodSystemKeyUniqueIndexQuery).Scan(&paymentMethodSystemKeyUnique); err != nil {
		return fmt.Errorf("inspect target unique index for payment_methods system_key")
	}
	if !paymentMethodSystemKeyUnique {
		return fmt.Errorf("target requires a valid partial unique index on payment_methods(clinic_id, system_key)")
	}
	return nil
}
