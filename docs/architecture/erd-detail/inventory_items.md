<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `inventory_items`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:672](../../../backend/migrations/001_init.sql#L672) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:673](../../../backend/migrations/001_init.sql#L673) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:674](../../../backend/migrations/001_init.sql#L674) |
| `category` | `inventory_category` | NOT NULL | — | — | [001_init.sql:675](../../../backend/migrations/001_init.sql#L675) |
| `quantity` | `integer` | NULL | `0` | — | [001_init.sql:676](../../../backend/migrations/001_init.sql#L676) |
| `unit` | `text` | NOT NULL | `''` | — | [001_init.sql:677](../../../backend/migrations/001_init.sql#L677) |
| `min_stock_level` | `integer` | NULL | `0` | — | [001_init.sql:678](../../../backend/migrations/001_init.sql#L678) |
| `location` | `text` | NOT NULL | `''` | — | [001_init.sql:679](../../../backend/migrations/001_init.sql#L679) |
| `expiry_date` | `date` | NULL | — | — | [001_init.sql:680](../../../backend/migrations/001_init.sql#L680) |
| `supplier` | `text` | NOT NULL | `''` | — | [001_init.sql:681](../../../backend/migrations/001_init.sql#L681) |
| `last_restocked` | `date` | NULL | — | — | [001_init.sql:682](../../../backend/migrations/001_init.sql#L682) |
| `status` | `inventory_status` | NULL | `'sufficient'` | — | [001_init.sql:683](../../../backend/migrations/001_init.sql#L683) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:684](../../../backend/migrations/001_init.sql#L684) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:685](../../../backend/migrations/001_init.sql#L685) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:686](../../../backend/migrations/001_init.sql#L686) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    inventory_items {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        inventory_category category
        integer quantity
        text unit
        integer min_stock_level
        text location
        date expiry_date
        text supplier
        date last_restocked
        inventory_status status
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ inventory_items : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `inventory_items_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:673](../../../backend/migrations/001_init.sql#L673) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `inventory_items_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `inventory_items_pkey` | primary | `PRIMARY KEY` | [001_init.sql:672](../../../backend/migrations/001_init.sql#L672) |
| `uq_inventory_items_id_clinic` | unique | `CONSTRAINT uq_inventory_items_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:690](../../../backend/migrations/001_init.sql#L690) |
| `chk_inventory_items_quantity_non_negative` | check | `CONSTRAINT chk_inventory_items_quantity_non_negative CHECK (quantity >= 0)` | [001_init.sql:4167](../../../backend/migrations/001_init.sql#L4167) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_inventory_items_clinic_id` | NO | — | `CREATE INDEX idx_inventory_items_clinic_id ON inventory_items(clinic_id)` | [001_init.sql:2219](../../../backend/migrations/001_init.sql#L2219) |
| `idx_inventory_items_category` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_inventory_items_category ON inventory_items(category) WHERE deleted_at IS NULL` | [001_init.sql:2433](../../../backend/migrations/001_init.sql#L2433) |
| `idx_inventory_items_name_fold_trgm` | NO | — | `CREATE INDEX idx_inventory_items_name_fold_trgm   ON inventory_items USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6726](../../../backend/migrations/001_init.sql#L6726) |
