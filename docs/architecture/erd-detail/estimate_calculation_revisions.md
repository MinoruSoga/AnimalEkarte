<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `estimate_calculation_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10956](../../../backend/migrations/001_init.sql#L10956) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10956](../../../backend/migrations/001_init.sql#L10956) |
| `estimate_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10957](../../../backend/migrations/001_init.sql#L10957) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10957](../../../backend/migrations/001_init.sql#L10957) |
| `rule_version` | `text` | NOT NULL | — | — | [001_init.sql:10957](../../../backend/migrations/001_init.sql#L10957) |
| `conditions` | `jsonb` | NOT NULL | — | — | [001_init.sql:10958](../../../backend/migrations/001_init.sql#L10958) |
| `insurance_confirmation` | `jsonb` | NOT NULL | — | — | [001_init.sql:10959](../../../backend/migrations/001_init.sql#L10959) |
| `tax_rate_totals` | `jsonb` | NOT NULL | — | — | [001_init.sql:10960](../../../backend/migrations/001_init.sql#L10960) |
| `line_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:10961](../../../backend/migrations/001_init.sql#L10961) |
| `subtotal` | `bigint` | NOT NULL | — | — | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `tax_total` | `bigint` | NOT NULL | — | — | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `total_amount` | `bigint` | NOT NULL | — | — | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10963](../../../backend/migrations/001_init.sql#L10963) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10963](../../../backend/migrations/001_init.sql#L10963) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10964](../../../backend/migrations/001_init.sql#L10964) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    estimate_calculation_revisions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint estimate_id FK,UK
        bigint revision UK
        text rule_version
        jsonb conditions
        jsonb insurance_confirmation
        jsonb tax_rate_totals
        jsonb line_values
        bigint subtotal
        bigint tax_total
        bigint total_amount
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    estimates {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ estimate_calculation_revisions : "clinic_id"
    staffs ||..o{ estimate_calculation_revisions : "recorded_by"
    estimates ||..o{ estimate_calculation_revisions : "estimate_id, clinic_id"
    operation_bindings ||..o{ estimate_calculation_revisions : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `estimate_calculation_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10956](../../../backend/migrations/001_init.sql#L10956) |
| `estimate_calculation_revisions_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10963](../../../backend/migrations/001_init.sql#L10963) |
| `estimate_calculation_revisions_estimate_id_clinic_id_fkey` | `estimate_id, clinic_id` | [estimates](estimates.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10966](../../../backend/migrations/001_init.sql#L10966) |
| `estimate_calculation_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10967](../../../backend/migrations/001_init.sql#L10967) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `estimate_calculation_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `estimate_calculation_revisions_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `estimate_calculation_revisions_estimate_id_clinic_id_fkey`: `FOREIGN KEY(estimate_id,clinic_id) REFERENCES estimates(id,clinic_id) ON DELETE RESTRICT`
- `estimate_calculation_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `estimate_calculation_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10956](../../../backend/migrations/001_init.sql#L10956) |
| `estimate_calculation_revisions_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10957](../../../backend/migrations/001_init.sql#L10957) |
| `estimate_calculation_revisions_conditions_check` | check | `CHECK(jsonb_typeof(conditions)='object')` | [001_init.sql:10958](../../../backend/migrations/001_init.sql#L10958) |
| `estimate_calculation_revisions_insurance_confirmation_check` | check | `CHECK(jsonb_typeof(insurance_confirmation)='object')` | [001_init.sql:10959](../../../backend/migrations/001_init.sql#L10959) |
| `estimate_calculation_revisions_tax_rate_totals_check` | check | `CHECK(jsonb_typeof(tax_rate_totals)='array')` | [001_init.sql:10960](../../../backend/migrations/001_init.sql#L10960) |
| `estimate_calculation_revisions_line_values_check` | check | `CHECK(jsonb_typeof(line_values)='array')` | [001_init.sql:10961](../../../backend/migrations/001_init.sql#L10961) |
| `estimate_calculation_revisions_subtotal_check` | check | `CHECK(subtotal>=0)` | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `estimate_calculation_revisions_tax_total_check` | check | `CHECK(tax_total>=0)` | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `estimate_calculation_revisions_total_amount_check` | check | `CHECK(total_amount>=0)` | [001_init.sql:10962](../../../backend/migrations/001_init.sql#L10962) |
| `estimate_calculation_revisions_id_clinic_id_estimate_id_key` | unique | `UNIQUE(id,clinic_id,estimate_id)` | [001_init.sql:10965](../../../backend/migrations/001_init.sql#L10965) |
| `estimate_calculation_revision_clinic_id_estimate_id_revisio_key` | unique | `UNIQUE(clinic_id,estimate_id,revision)` | [001_init.sql:10965](../../../backend/migrations/001_init.sql#L10965) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
