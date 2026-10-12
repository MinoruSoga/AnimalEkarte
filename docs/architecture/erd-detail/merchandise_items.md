<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `merchandise_items`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1808](../../../backend/migrations/001_init.sql#L1808) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1809](../../../backend/migrations/001_init.sql#L1809) |
| `name` | `text` | NOT NULL | `''` | — | [001_init.sql:1810](../../../backend/migrations/001_init.sql#L1810) |
| `category` | `item_category` | NOT NULL | `'goods'` | — | [001_init.sql:1811](../../../backend/migrations/001_init.sql#L1811) |
| `unit_price` | `bigint` | NOT NULL | `0` | — | [001_init.sql:1812](../../../backend/migrations/001_init.sql#L1812) |
| `tax_type` | `tax_type` | NOT NULL | `'excluded'` | — | [001_init.sql:1813](../../../backend/migrations/001_init.sql#L1813) |
| `tax_rate` | `numeric` | NOT NULL | `0.10` | — | [001_init.sql:1814](../../../backend/migrations/001_init.sql#L1814) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1815](../../../backend/migrations/001_init.sql#L1815) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [001_init.sql:1816](../../../backend/migrations/001_init.sql#L1816) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1817](../../../backend/migrations/001_init.sql#L1817) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1818](../../../backend/migrations/001_init.sql#L1818) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1819](../../../backend/migrations/001_init.sql#L1819) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    merchandise_items {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        item_category category
        bigint unit_price
        tax_type tax_type
        numeric tax_rate
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ merchandise_items : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `merchandise_items_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1809](../../../backend/migrations/001_init.sql#L1809) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `merchandise_items_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `merchandise_items_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1808](../../../backend/migrations/001_init.sql#L1808) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_merchandise_items_clinic` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_merchandise_items_clinic ON merchandise_items(clinic_id) WHERE deleted_at IS NULL` | [001_init.sql:2248](../../../backend/migrations/001_init.sql#L2248) |
| `idx_merchandise_items_category` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_merchandise_items_category ON merchandise_items(clinic_id, category) WHERE deleted_at IS NULL` | [001_init.sql:2249](../../../backend/migrations/001_init.sql#L2249) |
| `idx_merchandise_items_sort` | NO | — | `CREATE INDEX idx_merchandise_items_sort ON merchandise_items(clinic_id, sort_order)` | [001_init.sql:2250](../../../backend/migrations/001_init.sql#L2250) |
| `idx_merchandise_items_clinic_name` | YES | `is_active = true AND deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_merchandise_items_clinic_name ON merchandise_items(clinic_id, name) WHERE is_active = true AND deleted_at IS NULL` | [001_init.sql:2466](../../../backend/migrations/001_init.sql#L2466) |
