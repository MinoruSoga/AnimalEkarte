<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_options`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1008](../../../backend/migrations/001_init.sql#L1008) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1009](../../../backend/migrations/001_init.sql#L1009) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:1010](../../../backend/migrations/001_init.sql#L1010) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:1011](../../../backend/migrations/001_init.sql#L1011) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1012](../../../backend/migrations/001_init.sql#L1012) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:1013](../../../backend/migrations/001_init.sql#L1013) |
| `duration` | `integer` | NULL | — | — | [001_init.sql:1014](../../../backend/migrations/001_init.sql#L1014) |
| `is_combinable` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1015](../../../backend/migrations/001_init.sql#L1015) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1016](../../../backend/migrations/001_init.sql#L1016) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1017](../../../backend/migrations/001_init.sql#L1017) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1018](../../../backend/migrations/001_init.sql#L1018) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1019](../../../backend/migrations/001_init.sql#L1019) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_options {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        bigint price
        boolean is_active
        text description
        integer duration
        boolean is_combinable
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ trimming_options : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_options_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1009](../../../backend/migrations/001_init.sql#L1009) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_options_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_options_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1008](../../../backend/migrations/001_init.sql#L1008) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_trimming_options_clinic_id` | NO | — | `CREATE INDEX idx_trimming_options_clinic_id ON trimming_options(clinic_id)` | [001_init.sql:2227](../../../backend/migrations/001_init.sql#L2227) |
| `idx_trimming_options_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_trimming_options_clinic_sort  ON trimming_options(clinic_id, sort_order)        WHERE deleted_at IS NULL` | [001_init.sql:2263](../../../backend/migrations/001_init.sql#L2263) |
| `idx_trimming_options_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_trimming_options_clinic_name ON trimming_options(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2459](../../../backend/migrations/001_init.sql#L2459) |
| `idx_trimming_options_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_trimming_options_active ON trimming_options(id) WHERE deleted_at IS NULL` | [001_init.sql:2479](../../../backend/migrations/001_init.sql#L2479) |
