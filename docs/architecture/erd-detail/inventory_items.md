<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `inventory_items`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, FK, UK | [001_init.sql:686](../../../backend/migrations/001_init.sql#L686) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:687](../../../backend/migrations/001_init.sql#L687) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:688](../../../backend/migrations/001_init.sql#L688) |
| `category` | `inventory_category` | NOT NULL | — | — | [001_init.sql:689](../../../backend/migrations/001_init.sql#L689) |
| `quantity` | `bigint` | NULL | `0` | — | [001_init.sql:10235](../../../backend/migrations/001_init.sql#L10235) |
| `unit` | `text` | NOT NULL | `''` | — | [001_init.sql:691](../../../backend/migrations/001_init.sql#L691) |
| `min_stock_level` | `integer` | NULL | `0` | — | [001_init.sql:692](../../../backend/migrations/001_init.sql#L692) |
| `location` | `text` | NOT NULL | `''` | — | [001_init.sql:693](../../../backend/migrations/001_init.sql#L693) |
| `expiry_date` | `date` | NULL | — | — | [001_init.sql:694](../../../backend/migrations/001_init.sql#L694) |
| `supplier` | `text` | NOT NULL | `''` | — | [001_init.sql:695](../../../backend/migrations/001_init.sql#L695) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:698](../../../backend/migrations/001_init.sql#L698) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:699](../../../backend/migrations/001_init.sql#L699) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:700](../../../backend/migrations/001_init.sql#L700) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:9746](../../../backend/migrations/001_init.sql#L9746) |
| `quantity_base_id` | `bigint` | NULL | — | FK | [001_init.sql:9763](../../../backend/migrations/001_init.sql#L9763) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    inventory_items {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
        text name
        inventory_category category
        bigint quantity
        text unit
        integer min_stock_level
        text location
        date expiry_date
        text supplier
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        bigint version
        bigint quantity_base_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    inventory_quantity_bases {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint inventory_item_id FK,UK
    }
    clinics ||..o{ inventory_items : "clinic_id"
    inventory_quantity_bases o|..o| inventory_items : "quantity_base_id, clinic_id, id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `inventory_items_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:687](../../../backend/migrations/001_init.sql#L687) |
| `fk_inventory_quantity_base` | `quantity_base_id, clinic_id, id` | [inventory_quantity_bases](inventory_quantity_bases.md) | `id, clinic_id, inventory_item_id` | `ON DELETE RESTRICT` | [001_init.sql:9764](../../../backend/migrations/001_init.sql#L9764) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `inventory_items_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_inventory_quantity_base`: `CONSTRAINT fk_inventory_quantity_base  FOREIGN KEY (quantity_base_id,clinic_id,id) REFERENCES inventory_quantity_bases(id,clinic_id,inventory_item_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `inventory_items_pkey` | primary | `PRIMARY KEY` | [001_init.sql:686](../../../backend/migrations/001_init.sql#L686) |
| `uq_inventory_items_id_clinic` | unique | `CONSTRAINT uq_inventory_items_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:704](../../../backend/migrations/001_init.sql#L704) |
| `chk_inventory_items_quantity_non_negative` | check | `CONSTRAINT chk_inventory_items_quantity_non_negative CHECK (quantity >= 0)` | [001_init.sql:4181](../../../backend/migrations/001_init.sql#L4181) |
| `inventory_items_version_check` | check | `CHECK (version > 0)` | [001_init.sql:9746](../../../backend/migrations/001_init.sql#L9746) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_inventory_items_clinic_id` | NO | — | `CREATE INDEX idx_inventory_items_clinic_id ON inventory_items(clinic_id)` | [001_init.sql:2233](../../../backend/migrations/001_init.sql#L2233) |
| `idx_inventory_items_category` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_inventory_items_category ON inventory_items(category) WHERE deleted_at IS NULL` | [001_init.sql:2447](../../../backend/migrations/001_init.sql#L2447) |
| `idx_inventory_items_name_fold_trgm` | NO | — | `CREATE INDEX idx_inventory_items_name_fold_trgm   ON inventory_items USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6740](../../../backend/migrations/001_init.sql#L6740) |
