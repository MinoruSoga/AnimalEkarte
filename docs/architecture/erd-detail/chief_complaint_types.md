<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `chief_complaint_types`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1076](../../../backend/migrations/001_init.sql#L1076) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1077](../../../backend/migrations/001_init.sql#L1077) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:1078](../../../backend/migrations/001_init.sql#L1078) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:1079](../../../backend/migrations/001_init.sql#L1079) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1080](../../../backend/migrations/001_init.sql#L1080) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1081](../../../backend/migrations/001_init.sql#L1081) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1082](../../../backend/migrations/001_init.sql#L1082) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1083](../../../backend/migrations/001_init.sql#L1083) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1084](../../../backend/migrations/001_init.sql#L1084) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    chief_complaint_types {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        text description
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ chief_complaint_types : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `chief_complaint_types_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1077](../../../backend/migrations/001_init.sql#L1077) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `chief_complaint_types_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `chief_complaint_types_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1076](../../../backend/migrations/001_init.sql#L1076) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_chief_complaint_types_clinic_id` | NO | — | `CREATE INDEX idx_chief_complaint_types_clinic_id ON chief_complaint_types(clinic_id)` | [001_init.sql:2233](../../../backend/migrations/001_init.sql#L2233) |
| `idx_chief_complaints_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_chief_complaints_clinic_sort  ON chief_complaint_types(clinic_id, sort_order)   WHERE deleted_at IS NULL` | [001_init.sql:2259](../../../backend/migrations/001_init.sql#L2259) |
| `idx_chief_complaint_types_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_chief_complaint_types_clinic_name ON chief_complaint_types(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2464](../../../backend/migrations/001_init.sql#L2464) |
| `idx_chief_complaint_types_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_chief_complaint_types_active ON chief_complaint_types(id) WHERE deleted_at IS NULL` | [001_init.sql:2482](../../../backend/migrations/001_init.sql#L2482) |
