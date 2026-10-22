<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `exam_types`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:696](../../../backend/migrations/001_init.sql#L696) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:697](../../../backend/migrations/001_init.sql#L697) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:698](../../../backend/migrations/001_init.sql#L698) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:699](../../../backend/migrations/001_init.sql#L699) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:700](../../../backend/migrations/001_init.sql#L700) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:701](../../../backend/migrations/001_init.sql#L701) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:702](../../../backend/migrations/001_init.sql#L702) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:703](../../../backend/migrations/001_init.sql#L703) |
| `is_non_insurance` | `boolean` | NOT NULL | `false` | — | [001_init.sql:704](../../../backend/migrations/001_init.sql#L704) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:705](../../../backend/migrations/001_init.sql#L705) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:706](../../../backend/migrations/001_init.sql#L706) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:707](../../../backend/migrations/001_init.sql#L707) |
| `is_bill_discount_excluded` | `boolean` | NULL | — | — | [007_bill_discount_excluded_and_master_versions.sql:12](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L12) |
| `version` | `bigint` | NOT NULL | `1` | — | [007_bill_discount_excluded_and_master_versions.sql:37](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L37) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    exam_types {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        bigint parent_id FK
        integer sort_order
        boolean is_non_insurance
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        boolean is_bill_discount_excluded
        bigint version
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ exam_types : "clinic_id"
    exam_types o|..o{ exam_types : "parent_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `exam_types_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:697](../../../backend/migrations/001_init.sql#L697) |
| `exam_types_parent_id_fkey` | `parent_id` | [exam_types](exam_types.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:702](../../../backend/migrations/001_init.sql#L702) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `exam_types_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `exam_types_parent_id_fkey`: `REFERENCES exam_types(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `exam_types_pkey` | primary | `PRIMARY KEY` | [001_init.sql:696](../../../backend/migrations/001_init.sql#L696) |
| `uq_exam_types_id_clinic` | unique | `CONSTRAINT uq_exam_types_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:3912](../../../backend/migrations/001_init.sql#L3912) |
| `ck_exam_types_positive_version` | check | `CONSTRAINT ck_exam_types_positive_version CHECK (version > 0)` | [007_bill_discount_excluded_and_master_versions.sql:38](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L38) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_exam_types_clinic_id` | NO | — | `CREATE INDEX idx_exam_types_clinic_id ON exam_types(clinic_id)` | [001_init.sql:2220](../../../backend/migrations/001_init.sql#L2220) |
| `idx_exam_types_parent_id` | NO | — | `CREATE INDEX idx_exam_types_parent_id ON exam_types(parent_id)` | [001_init.sql:2221](../../../backend/migrations/001_init.sql#L2221) |
| `idx_exam_types_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_exam_types_clinic_sort        ON exam_types(clinic_id, sort_order)              WHERE deleted_at IS NULL` | [001_init.sql:2263](../../../backend/migrations/001_init.sql#L2263) |
| `idx_exam_types_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_exam_types_clinic_name ON exam_types(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2458](../../../backend/migrations/001_init.sql#L2458) |
| `idx_exam_types_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_exam_types_active ON exam_types(id) WHERE deleted_at IS NULL` | [001_init.sql:2478](../../../backend/migrations/001_init.sql#L2478) |
