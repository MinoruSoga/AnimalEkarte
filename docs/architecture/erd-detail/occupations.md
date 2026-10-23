<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `occupations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:257](../../../backend/migrations/001_init.sql#L257) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:258](../../../backend/migrations/001_init.sql#L258) |
| `name` | `text` | NOT NULL | `''` | — | [001_init.sql:259](../../../backend/migrations/001_init.sql#L259) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:260](../../../backend/migrations/001_init.sql#L260) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [001_init.sql:261](../../../backend/migrations/001_init.sql#L261) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:262](../../../backend/migrations/001_init.sql#L262) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:263](../../../backend/migrations/001_init.sql#L263) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:264](../../../backend/migrations/001_init.sql#L264) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:265](../../../backend/migrations/001_init.sql#L265) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    occupations {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        text description
        integer sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ occupations : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `occupations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:258](../../../backend/migrations/001_init.sql#L258) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `occupations_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `occupations_pkey` | primary | `PRIMARY KEY` | [001_init.sql:257](../../../backend/migrations/001_init.sql#L257) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_occupations_clinic_id` | NO | — | `CREATE INDEX idx_occupations_clinic_id ON occupations(clinic_id)` | [001_init.sql:2230](../../../backend/migrations/001_init.sql#L2230) |
| `idx_occupations_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_occupations_clinic_sort       ON occupations(clinic_id, sort_order)             WHERE deleted_at IS NULL` | [001_init.sql:2285](../../../backend/migrations/001_init.sql#L2285) |
| `idx_occupations_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_occupations_clinic_name ON occupations(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2483](../../../backend/migrations/001_init.sql#L2483) |
| `idx_occupations_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_occupations_active ON occupations(id) WHERE deleted_at IS NULL` | [001_init.sql:2489](../../../backend/migrations/001_init.sql#L2489) |
