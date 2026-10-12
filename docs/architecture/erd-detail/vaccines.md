<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `vaccines`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:721](../../../backend/migrations/001_init.sql#L721) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:722](../../../backend/migrations/001_init.sql#L722) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:723](../../../backend/migrations/001_init.sql#L723) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:724](../../../backend/migrations/001_init.sql#L724) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:725](../../../backend/migrations/001_init.sql#L725) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:726](../../../backend/migrations/001_init.sql#L726) |
| `species` | `vaccine_species` | NULL | — | — | [001_init.sql:727](../../../backend/migrations/001_init.sql#L727) |
| `interval` | `text` | NOT NULL | `''` | — | [001_init.sql:728](../../../backend/migrations/001_init.sql#L728) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:729](../../../backend/migrations/001_init.sql#L729) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:730](../../../backend/migrations/001_init.sql#L730) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:731](../../../backend/migrations/001_init.sql#L731) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:732](../../../backend/migrations/001_init.sql#L732) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:733](../../../backend/migrations/001_init.sql#L733) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:734](../../../backend/migrations/001_init.sql#L734) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    vaccines {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        bigint price
        boolean is_active
        text description
        vaccine_species species
        text interval
        bigint inventory_id FK
        bigint parent_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    inventory_items {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ vaccines : "clinic_id"
    inventory_items o|..o{ vaccines : "inventory_id"
    vaccines o|..o{ vaccines : "parent_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `vaccines_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:722](../../../backend/migrations/001_init.sql#L722) |
| `vaccines_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:729](../../../backend/migrations/001_init.sql#L729) |
| `vaccines_parent_id_fkey` | `parent_id` | [vaccines](vaccines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:730](../../../backend/migrations/001_init.sql#L730) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `vaccines_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `vaccines_inventory_id_fkey`: `REFERENCES inventory_items(id) ON DELETE SET NULL`
- `vaccines_parent_id_fkey`: `REFERENCES vaccines(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `vaccines_pkey` | primary | `PRIMARY KEY` | [001_init.sql:721](../../../backend/migrations/001_init.sql#L721) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_vaccines_clinic_id` | NO | — | `CREATE INDEX idx_vaccines_clinic_id ON vaccines(clinic_id)` | [001_init.sql:2214](../../../backend/migrations/001_init.sql#L2214) |
| `idx_vaccines_parent_id` | NO | — | `CREATE INDEX idx_vaccines_parent_id ON vaccines(parent_id)` | [001_init.sql:2215](../../../backend/migrations/001_init.sql#L2215) |
| `idx_vaccines_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_vaccines_clinic_sort          ON vaccines(clinic_id, sort_order)               WHERE deleted_at IS NULL` | [001_init.sql:2253](../../../backend/migrations/001_init.sql#L2253) |
| `idx_vaccines_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_vaccines_inventory_id ON vaccines(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2441](../../../backend/migrations/001_init.sql#L2441) |
| `idx_vaccines_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_vaccines_clinic_name ON vaccines(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2451](../../../backend/migrations/001_init.sql#L2451) |
| `idx_vaccines_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_vaccines_active ON vaccines(id) WHERE deleted_at IS NULL` | [001_init.sql:2471](../../../backend/migrations/001_init.sql#L2471) |
| `idx_vaccines_name_fold_trgm` | NO | — | `CREATE INDEX idx_vaccines_name_fold_trgm   ON vaccines USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:81](../../../backend/migrations/015_search_fold_indexes.sql#L81) |
