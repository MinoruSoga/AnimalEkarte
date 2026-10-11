<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medicines`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:741](../../../backend/migrations/001_init.sql#L741) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:742](../../../backend/migrations/001_init.sql#L742) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:743](../../../backend/migrations/001_init.sql#L743) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:744](../../../backend/migrations/001_init.sql#L744) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:745](../../../backend/migrations/001_init.sql#L745) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:746](../../../backend/migrations/001_init.sql#L746) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:747](../../../backend/migrations/001_init.sql#L747) |
| `dosage_form` | `dosage_form` | NULL | — | — | [001_init.sql:748](../../../backend/migrations/001_init.sql#L748) |
| `medicine_unit` | `medicine_unit` | NULL | — | — | [001_init.sql:749](../../../backend/migrations/001_init.sql#L749) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:750](../../../backend/migrations/001_init.sql#L750) |
| `default_quantity` | `numeric(10,2)` | NULL | `1` | — | [001_init.sql:3453](../../../backend/migrations/001_init.sql#L3453) |
| `tax_type` | `tax_type` | NOT NULL | `'excluded'` | — | [001_init.sql:752](../../../backend/migrations/001_init.sql#L752) |
| `tax_rate` | `numeric` | NOT NULL | `0.10` | — | [001_init.sql:753](../../../backend/migrations/001_init.sql#L753) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:754](../../../backend/migrations/001_init.sql#L754) |
| `is_non_insurance` | `boolean` | NOT NULL | `false` | — | [001_init.sql:755](../../../backend/migrations/001_init.sql#L755) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:756](../../../backend/migrations/001_init.sql#L756) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:757](../../../backend/migrations/001_init.sql#L757) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:758](../../../backend/migrations/001_init.sql#L758) |
| `calculation_type` | `medicine_calculation_type` | NOT NULL | `'none'` | — | [001_init.sql:3449](../../../backend/migrations/001_init.sql#L3449) |
| `strength` | `numeric(10,4)` | NULL | — | — | [001_init.sql:3450](../../../backend/migrations/001_init.sql#L3450) |
| `frequency_per_day` | `integer` | NULL | — | — | [001_init.sql:3451](../../../backend/migrations/001_init.sql#L3451) |
| `default_duration_days` | `integer` | NULL | — | — | [001_init.sql:3452](../../../backend/migrations/001_init.sql#L3452) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medicines {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        bigint parent_id FK
        dosage_form dosage_form
        medicine_unit medicine_unit
        bigint inventory_id FK
        numeric_10_2 default_quantity
        tax_type tax_type
        numeric tax_rate
        integer sort_order
        boolean is_non_insurance
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        medicine_calculation_type calculation_type
        numeric_10_4 strength
        integer frequency_per_day
        integer default_duration_days
    }
    clinics {
        BIGSERIAL id PK
    }
    inventory_items {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ medicines : "clinic_id"
    medicines o|..o{ medicines : "parent_id"
    inventory_items o|..o{ medicines : "inventory_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medicines_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:742](../../../backend/migrations/001_init.sql#L742) |
| `medicines_parent_id_fkey` | `parent_id` | [medicines](medicines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:747](../../../backend/migrations/001_init.sql#L747) |
| `medicines_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:750](../../../backend/migrations/001_init.sql#L750) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medicines_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medicines_parent_id_fkey`: `REFERENCES medicines(id) ON DELETE SET NULL`
- `medicines_inventory_id_fkey`: `REFERENCES inventory_items(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medicines_pkey` | primary | `PRIMARY KEY` | [001_init.sql:741](../../../backend/migrations/001_init.sql#L741) |
| `ck_medicines_per_weight_strength` | check | `CONSTRAINT ck_medicines_per_weight_strength     CHECK (calculation_type = 'none' OR strength IS NOT NULL)` | [001_init.sql:3457](../../../backend/migrations/001_init.sql#L3457) |
| `ck_medicines_strength_positive` | check | `CONSTRAINT ck_medicines_strength_positive     CHECK (strength IS NULL OR strength > 0)` | [001_init.sql:3462](../../../backend/migrations/001_init.sql#L3462) |
| `ck_medicines_frequency_positive` | check | `CONSTRAINT ck_medicines_frequency_positive     CHECK (frequency_per_day IS NULL OR frequency_per_day > 0)` | [001_init.sql:3466](../../../backend/migrations/001_init.sql#L3466) |
| `ck_medicines_duration_positive` | check | `CONSTRAINT ck_medicines_duration_positive     CHECK (default_duration_days IS NULL OR default_duration_days > 0)` | [001_init.sql:3470](../../../backend/migrations/001_init.sql#L3470) |
| `uq_medicines_id_clinic` | unique | `CONSTRAINT uq_medicines_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:3476](../../../backend/migrations/001_init.sql#L3476) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_medicines_clinic_id` | NO | — | `CREATE INDEX idx_medicines_clinic_id ON medicines(clinic_id)` | [001_init.sql:2216](../../../backend/migrations/001_init.sql#L2216) |
| `idx_medicines_parent_id` | NO | — | `CREATE INDEX idx_medicines_parent_id ON medicines(parent_id)` | [001_init.sql:2217](../../../backend/migrations/001_init.sql#L2217) |
| `idx_medicines_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_clinic_sort         ON medicines(clinic_id, sort_order)               WHERE deleted_at IS NULL` | [001_init.sql:2254](../../../backend/migrations/001_init.sql#L2254) |
| `idx_medicines_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_medicines_inventory_id ON medicines(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2442](../../../backend/migrations/001_init.sql#L2442) |
| `idx_medicines_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2452](../../../backend/migrations/001_init.sql#L2452) |
| `idx_medicines_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_active ON medicines(id) WHERE deleted_at IS NULL` | [001_init.sql:2472](../../../backend/migrations/001_init.sql#L2472) |
| `idx_medicines_name_fold_trgm` | NO | — | `CREATE INDEX idx_medicines_name_fold_trgm   ON medicines USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:72](../../../backend/migrations/015_search_fold_indexes.sql#L72) |
