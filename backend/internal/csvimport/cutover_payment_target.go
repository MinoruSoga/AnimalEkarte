package csvimport

import (
	"context"
	"fmt"
)

const verifyCutoverPaymentGraphQuery = `
WITH payment_rows AS MATERIALIZED (
  SELECT
    payment.id,
    payment.clinic_id,
    payment.billing_id,
    payment.subtotal,
    payment.tax_total,
    payment.total_amount,
    payment.insurance_ratio,
    payment.insurance_amount,
    payment.discount_amount,
    payment.billing_amount,
    payment.received_amount,
    payment.change_amount,
    payment.method,
    payment.payment_method_id,
    payment.paid_by,
    payment.created_at,
    payment.deleted_at,
    billing.id AS parent_billing_id,
    billing.clinic_id AS billing_clinic_id,
    billing.total_amount AS billing_total_amount,
    billing.status AS billing_status,
    billing.completed_at AS billing_completed_at,
    billing.deleted_at AS billing_deleted_at
  FROM payments payment
  LEFT JOIN billings billing ON billing.id = payment.billing_id
  WHERE payment.id >= $1 AND payment.id < $2
),
split_rows AS MATERIALIZED (
  SELECT
    split.id,
    split.clinic_id,
    split.billing_id,
    split.method,
    split.payment_method_id,
    split.amount,
    split.received_amount,
    split.change_amount,
    split.paid_by,
    split.created_at,
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
      split.id >= $1
      AND split.id < $2
      AND split.clinic_id = $3
      AND split.billing_id >= $1
      AND split.billing_id < $2
      AND split.split_billing_clinic_id = $3
      AND (
        (split.method = 'cash' AND split.payment_method_id = $4)
        OR (split.method = 'credit_card' AND split.payment_method_id = $5)
      )
      AND (
        split.paid_by IS NULL
        OR (
          split.paid_by >= $1
          AND split.paid_by < $2
          AND split.split_staff_clinic_id = $3
        )
      )
      AND split.paid_by IS NOT DISTINCT FROM split.payment_paid_by
      AND split.created_at IS NOT NULL
      AND split.created_at IS NOT DISTINCT FROM split.payment_created_at
      AND split.amount <> 0
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
  WHERE payment.deleted_at IS NOT NULL
    OR payment.parent_billing_id IS NULL
    OR payment.parent_billing_id < $1
    OR payment.parent_billing_id >= $2
    OR payment.billing_deleted_at IS NOT NULL
    OR payment.clinic_id IS NULL
    OR payment.clinic_id IS DISTINCT FROM $3
    OR payment.clinic_id IS DISTINCT FROM payment.billing_clinic_id
    OR payment.billing_clinic_id IS DISTINCT FROM $3
    OR ($6::boolean AND payment.billing_total_amount IS DISTINCT FROM payment.total_amount)
    OR payment.billing_status IS DISTINCT FROM 'completed'
    OR payment.billing_completed_at IS DISTINCT FROM payment.created_at
    OR payment.subtotal IS NULL
    OR payment.tax_total IS NULL
    OR payment.total_amount IS NULL
    OR payment.insurance_ratio IS NULL
    OR payment.insurance_ratio < 0
    OR payment.insurance_ratio > 1
    OR payment.insurance_amount IS NULL
    OR payment.insurance_amount < 0
    OR payment.discount_amount IS NULL
    OR payment.billing_amount IS NULL
    OR payment.billing_amount = 0
    OR payment.received_amount IS NULL
    OR payment.change_amount IS NULL
    OR payment.created_at IS NULL
    OR NOT COALESCE(
      (payment.method = 'cash' AND payment.payment_method_id = $4)
      OR (payment.method = 'credit_card' AND payment.payment_method_id = $5),
      false
    )
    OR (
      payment.paid_by IS NOT NULL
      AND (
        payment.paid_by < $1
        OR payment.paid_by >= $2
        OR payment_staff.clinic_id IS DISTINCT FROM $3
      )
    )
    OR COALESCE(split_summary.split_count, 0) NOT BETWEEN 1 AND 2
    OR COALESCE(split_summary.distinct_method_count, 0) <> COALESCE(split_summary.split_count, 0)
    OR COALESCE(split_summary.split_amount, 0) <> payment.billing_amount
    OR COALESCE(split_summary.cash_received, 0) <> payment.received_amount
    OR COALESCE(split_summary.cash_change, 0) <> payment.change_amount
    OR payment.method IS DISTINCT FROM
      CASE WHEN COALESCE(split_summary.has_cash, false) THEN 'cash'::payment_method ELSE 'credit_card'::payment_method END
    OR NOT COALESCE(split_summary.split_rows_valid, false)
),
completed_billing_violations AS (
  SELECT billing.id
  FROM billings billing
  LEFT JOIN payment_rows payment ON payment.billing_id = billing.id
  WHERE billing.id >= $1
    AND billing.id < $2
    AND (
      billing.deleted_at IS NOT NULL
      OR billing.status IS NULL
      OR billing.status NOT IN ('waiting', 'completed', 'cancelled', 'pending')
      OR (
        billing.status = 'completed'
        AND (
          billing.completed_at IS NULL
          OR (
            payment.id IS NULL
            AND COALESCE(billing.total_amount, 0) <> 0
          )
        )
      )
    )
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
  SELECT id FROM completed_billing_violations
  UNION ALL
  SELECT id FROM orphan_split_violations
) violation`

func verifyCutoverPaymentGraph(
	ctx context.Context,
	q cutoverQuerier,
	manifest *CutoverManifest,
	seeds CutoverSeedIDs,
	provenance CutoverProvenanceContract,
) error {
	requireExactPaymentSnapshot := provenance.Mode != CutoverProvenanceLocalRehearsal
	var violations int64
	if err := q.QueryRow(
		ctx,
		verifyCutoverPaymentGraphQuery,
		manifest.IDBand.NonOwnerIDOffset,
		manifest.IDBand.EndExclusive,
		seeds.ClinicID,
		seeds.CashPaymentMethodID,
		seeds.CreditCardPaymentMethodID,
		requireExactPaymentSnapshot,
	).Scan(&violations); err != nil {
		return fmt.Errorf("verify payment graph: %w", err)
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

const paymentSplitsBillingIndexQuery = `SELECT EXISTS (
  SELECT 1
  FROM pg_index index_definition
  JOIN pg_class target_table ON target_table.oid = index_definition.indrelid
  JOIN pg_namespace target_schema ON target_schema.oid = target_table.relnamespace
  WHERE target_schema.nspname = 'public'
    AND target_table.relname = 'payment_splits'
    AND index_definition.indisvalid = true
    AND index_definition.indisready = true
    AND index_definition.indpred IS NULL
    AND index_definition.indexprs IS NULL
    AND index_definition.indnkeyatts >= 1
    AND (
      SELECT attribute.attname::text
      FROM unnest(index_definition.indkey) WITH ORDINALITY AS key_column(attnum, ordinality)
      JOIN pg_attribute attribute
        ON attribute.attrelid = target_table.oid
       AND attribute.attnum = key_column.attnum
      WHERE key_column.ordinality = 1
    ) = 'billing_id'
)
`

func validateCutoverUniqueIndexes(ctx context.Context, q cutoverQuerier) error {
	var paymentsBillingUnique bool
	if err := q.QueryRow(ctx, paymentsBillingUniqueIndexQuery).Scan(&paymentsBillingUnique); err != nil {
		return fmt.Errorf("inspect target unique index for payments.billing_id: %w", err)
	}
	if !paymentsBillingUnique {
		return fmt.Errorf("target requires a valid unique index on payments.billing_id")
	}

	var paymentMethodSystemKeyUnique bool
	if err := q.QueryRow(ctx, paymentMethodSystemKeyUniqueIndexQuery).Scan(&paymentMethodSystemKeyUnique); err != nil {
		return fmt.Errorf("inspect target unique index for payment_methods system_key: %w", err)
	}
	if !paymentMethodSystemKeyUnique {
		return fmt.Errorf("target requires a valid partial unique index on payment_methods(clinic_id, system_key)")
	}

	var paymentSplitsBillingIndexed bool
	if err := q.QueryRow(ctx, paymentSplitsBillingIndexQuery).Scan(&paymentSplitsBillingIndexed); err != nil {
		return fmt.Errorf("inspect target index for payment_splits.billing_id: %w", err)
	}
	if !paymentSplitsBillingIndexed {
		return fmt.Errorf("target requires a valid index starting with payment_splits.billing_id")
	}
	return nil
}
