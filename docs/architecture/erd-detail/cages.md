<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `cages`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:782](../../../backend/migrations/001_init.sql#L782) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:783](../../../backend/migrations/001_init.sql#L783) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:784](../../../backend/migrations/001_init.sql#L784) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:785](../../../backend/migrations/001_init.sql#L785) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:786](../../../backend/migrations/001_init.sql#L786) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:787](../../../backend/migrations/001_init.sql#L787) |
| `cage_type` | `cage_type` | NOT NULL | — | — | [001_init.sql:788](../../../backend/migrations/001_init.sql#L788) |
| `cage_size` | `cage_size` | NOT NULL | — | — | [001_init.sql:789](../../../backend/migrations/001_init.sql#L789) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:790](../../../backend/migrations/001_init.sql#L790) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:791](../../../backend/migrations/001_init.sql#L791) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:792](../../../backend/migrations/001_init.sql#L792) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:793](../../../backend/migrations/001_init.sql#L793) |
| `merged_into_cage_id` | `bigint` | NULL | — | FK | [025_cages_merged_into_cage_id.sql:16](../../../backend/migrations/025_cages_merged_into_cage_id.sql#L16) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    cages {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        cage_type cage_type
        cage_size cage_size
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        bigint merged_into_cage_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ cages : "clinic_id"
    cages o|..o{ cages : "merged_into_cage_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `cages_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:783](../../../backend/migrations/001_init.sql#L783) |
| `fk_cages_merged_into_cage` | `merged_into_cage_id, clinic_id` | [cages](cages.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [025_cages_merged_into_cage_id.sql:19](../../../backend/migrations/025_cages_merged_into_cage_id.sql#L19) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `cages_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_cages_merged_into_cage`: `CONSTRAINT fk_cages_merged_into_cage         FOREIGN KEY (merged_into_cage_id, clinic_id)         REFERENCES cages (id, clinic_id)         ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `cages_pkey` | primary | `PRIMARY KEY` | [001_init.sql:782](../../../backend/migrations/001_init.sql#L782) |
| `uq_cages_id_clinic` | unique | `CONSTRAINT uq_cages_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:797](../../../backend/migrations/001_init.sql#L797) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_cages_clinic_id` | NO | — | `CREATE INDEX idx_cages_clinic_id ON cages(clinic_id)` | [001_init.sql:2219](../../../backend/migrations/001_init.sql#L2219) |
| `idx_cages_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_cages_clinic_sort             ON cages(clinic_id, sort_order)                   WHERE deleted_at IS NULL` | [001_init.sql:2257](../../../backend/migrations/001_init.sql#L2257) |
| `idx_cages_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_cages_clinic_name ON cages(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2455](../../../backend/migrations/001_init.sql#L2455) |
| `idx_cages_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_cages_active ON cages(id) WHERE deleted_at IS NULL` | [001_init.sql:2474](../../../backend/migrations/001_init.sql#L2474) |
| `idx_cages_merged_into_cage_id` | NO | `merged_into_cage_id IS NOT NULL AND deleted_at IS NULL` | `CREATE INDEX idx_cages_merged_into_cage_id     ON cages (clinic_id, merged_into_cage_id)     WHERE merged_into_cage_id IS NOT NULL AND deleted_at IS NULL` | [025_cages_merged_into_cage_id.sql:27](../../../backend/migrations/025_cages_merged_into_cage_id.sql#L27) |
